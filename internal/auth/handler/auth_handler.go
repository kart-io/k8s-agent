package handler

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	jwt2 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/kart-io/k8s-agent/common/utils"
	"github.com/kart-io/k8s-agent/internal/auth/forced-logout/session"
	"github.com/kart-io/k8s-agent/internal/auth/jwt"
	"github.com/kart-io/k8s-agent/internal/auth/types"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	db              *gorm.DB
	jwtSecret       []byte
	jwtExpiresHours int
	sessionService  *session.Service
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(db *gorm.DB, jwtSecret string, jwtExpiresHours int, sessionService *session.Service) *AuthHandler {
	return &AuthHandler{
		db:              db,
		jwtSecret:       []byte(jwtSecret),
		jwtExpiresHours: jwtExpiresHours,
		sessionService:  sessionService,
	}
}

// LoginHandler handles user login requests with session tracking.
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	handler := WithJSONRequest(h.loginLogic)
	handler(c)
}

// loginLogic contains the core business logic for user login
func (h *AuthHandler) loginLogic(c *gin.Context, loginReq *types.LoginRequest) (*types.LoginResponse, error) {
	// Retrieve user from database
	var user types.User
	if err := h.db.Where("username = ?", loginReq.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Check if user is active
	if user.Status != 1 {
		return nil, errors.New("user account is disabled")
	}

	// Generate unique JTI for session tracking
	jti := uuid.New().String()
	expiresAt := time.Now().Add(time.Duration(h.jwtExpiresHours) * time.Hour)

	// Create JWT token with JTI claim
	claims := jwt2.MapClaims{
		"jti":      jti,
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt2.NewWithClaims(jwt2.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		return nil, err
	}

	// Extract IP address from request
	ipAddress := utils.ExtractIPAddress(c)

	// Extract User-Agent from headers
	userAgent := c.GetHeader("User-Agent")
	if userAgent == "" {
		userAgent = "Unknown"
	}

	// Create session info for tracking
	sessionInfo := &types.SessionInfo{
		JTI:            jti,
		UserID:         user.ID,
		Username:       user.Username,
		Email:          user.Email,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		Location:       "Unknown", // TODO: Implement GeoIP lookup
		LoginAt:        time.Now(),
		LastActivityAt: time.Now(),
		ExpiresAt:      expiresAt,
	}

	// Store session in Redis (non-blocking)
	// Failed storage is logged but doesn't block login
	if err := h.sessionService.CreateSession(c.Request.Context(), sessionInfo); err != nil {
		// Log the error but don't fail the login
		// In production, use proper structured logging
		c.Writer.Header().Add("X-Session-Store-Warning", "Session tracking temporarily unavailable")
	}

	// Retrieve user roles for response
	var roles []types.Role
	h.db.Table("roles").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", user.ID).
		Find(&roles)

	// Build user info for response
	userInfo := &types.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		RealName: user.RealName,
		Avatar:   user.Avatar,
		Roles:    roles,
	}

	// Return login response with JTI
	return &types.LoginResponse{
		Token:     tokenString,
		JTI:       jti,
		ExpiresAt: expiresAt,
		User:      userInfo,
	}, nil
}

// LogoutHandler handles user logout requests
// In the future, this can be enhanced to revoke the session.
func (h *AuthHandler) LogoutHandler(c *gin.Context) {
	handler := WithJSONRequestNoResponse(h.logoutLogic)
	handler(c)
}

// logoutLogic contains the core business logic for logout
func (h *AuthHandler) logoutLogic(c *gin.Context, _ *interface{}) error {
	// Extract JTI from token if available
	// For now, just return success
	// TODO: Implement session revocation in Phase 4
	return nil
}

// RefreshTokenHandler handles token refresh requests using refresh tokens.
// Implements token rotation for enhanced security.
func (h *AuthHandler) RefreshTokenHandler(c *gin.Context) {
	handler := WithJSONRequest(h.refreshTokenLogic)
	handler(c)
}

// refreshTokenLogic contains the core business logic for token refresh
func (h *AuthHandler) refreshTokenLogic(c *gin.Context, req *types.RefreshTokenRequest) (*types.RefreshTokenResponse, error) {
	// Validate refresh token
	claims, err := jwt.ValidateRefreshToken(req.RefreshToken, string(h.jwtSecret))
	if err != nil {
		return nil, err
	}

	// Check if refresh token is blacklisted
	isBlacklisted, err := h.sessionService.IsRefreshTokenBlacklisted(c.Request.Context(), claims.ID)
	if err != nil {
		return nil, err
	}
	if isBlacklisted {
		return nil, errors.New("refresh token has been revoked")
	}

	// Verify user still exists and is active
	var user types.User
	if err := h.db.Where("id = ? AND status = ?", claims.UserID, 1).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found or disabled")
		}
		return nil, err
	}

	// Generate new token pair (token rotation)
	jti := uuid.New().String()
	expiresAt := time.Now().Add(time.Duration(h.jwtExpiresHours) * time.Hour)

	// Create new access token
	newClaims := jwt2.MapClaims{
		"jti":      jti,
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt2.NewWithClaims(jwt2.SigningMethodHS256, newClaims)
	accessToken, err := token.SignedString(h.jwtSecret)
	if err != nil {
		return nil, err
	}

	// Generate new refresh token (7 days)
	refreshJTI := uuid.New().String()
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour)

	refreshClaims := jwt2.MapClaims{
		"jti":        refreshJTI,
		"user_id":    user.ID,
		"username":   user.Username,
		"token_type": "refresh",
		"exp":        refreshExpiresAt.Unix(),
		"iat":        time.Now().Unix(),
	}

	refreshToken := jwt2.NewWithClaims(jwt2.SigningMethodHS256, refreshClaims)
	newRefreshToken, err := refreshToken.SignedString(h.jwtSecret)
	if err != nil {
		return nil, err
	}

	// Store new refresh token in Redis
	ctx := c.Request.Context()
	refreshTTL := time.Until(refreshExpiresAt)
	if err := h.sessionService.StoreRefreshToken(ctx, refreshJTI, user.ID, refreshTTL); err != nil {
		return nil, err
	}

	// Revoke old refresh token (token rotation)
	oldTokenTTL := time.Until(claims.ExpiresAt.Time)
	if oldTokenTTL > 0 {
		// Blacklist old refresh token to prevent reuse
		_ = h.sessionService.BlacklistRefreshToken(ctx, claims.ID, oldTokenTTL)
	}

	// Calculate expires_in
	expiresIn := int(time.Until(expiresAt).Seconds())

	return &types.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		ExpiresIn:    expiresIn,
	}, nil
}

// GetCurrentUserHandler returns the current authenticated user's information.
// Refactored to use WithNoRequest decorator pattern
func (h *AuthHandler) GetCurrentUserHandler(c *gin.Context) {
	handler := WithNoRequest(h.getCurrentUserLogic)
	handler(c)
}

// getCurrentUserLogic contains the core business logic for retrieving current user
func (h *AuthHandler) getCurrentUserLogic(c *gin.Context) (*types.UserInfo, error) {
	// Extract user ID from JWT claims (set by JWT middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, errors.New("user not authenticated")
	}

	var user types.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	// Retrieve user roles
	var roles []types.Role
	h.db.Table("roles").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", user.ID).
		Find(&roles)

	return &types.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		RealName: user.RealName,
		Avatar:   user.Avatar,
		Roles:    roles,
	}, nil
}

// GetAccessCodesHandler returns an array of permission codes for the authenticated user
// This is used by the frontend to determine what features/pages the user can access.
// Refactored to use WithNoRequest decorator pattern
func (h *AuthHandler) GetAccessCodesHandler(c *gin.Context) {
	handler := WithNoRequest(h.getAccessCodesLogic)
	handler(c)
}

// getAccessCodesLogic contains the core business logic for retrieving access codes
func (h *AuthHandler) getAccessCodesLogic(c *gin.Context) (*[]string, error) {
	// Extract user ID from JWT claims (set by JWT middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, errors.New("user not authenticated")
	}

	// Query permission codes for the user through their roles
	// Join: user_roles -> role_permissions -> permissions
	var codes []string
	err := h.db.Table("permissions").
		Select("DISTINCT permissions.code").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ? AND permissions.status = 1", userID).
		Pluck("code", &codes).Error
	if err != nil {
		return nil, err
	}

	// Return empty array if no permissions found (instead of null)
	if codes == nil {
		codes = []string{}
	}

	return &codes, nil
}
