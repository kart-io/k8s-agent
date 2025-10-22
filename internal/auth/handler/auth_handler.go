package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kart-io/k8s-agent/internal/auth/forced-logout/session"
	"github.com/kart-io/k8s-agent/internal/auth/types"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/k8s-agent/common/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	db              *gorm.DB
	jwtSecret       []byte
	jwtExpiresHours int
	sessionService  *session.Service
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(db *gorm.DB, jwtSecret string, jwtExpiresHours int, sessionService *session.Service) *AuthHandler {
	return &AuthHandler{
		db:              db,
		jwtSecret:       []byte(jwtSecret),
		jwtExpiresHours: jwtExpiresHours,
		sessionService:  sessionService,
	}
}

// LoginHandler handles user login requests with session tracking
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var loginReq types.LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		response.BadRequest(c, "Invalid request format", err)
		return
	}

	// Retrieve user from database
	var user types.User
	if err := h.db.Where("username = ?", loginReq.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Unauthorized(c, "Invalid username or password", nil)
			return
		}
		response.InternalError(c, "Database error", err)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
		response.Unauthorized(c, "Invalid username or password", nil)
		return
	}

	// Check if user is active
	if user.Status != 1 {
		response.Forbidden(c, "User account is disabled", nil)
		return
	}

	// Generate unique JTI for session tracking
	jti := uuid.New().String()
	expiresAt := time.Now().Add(time.Duration(h.jwtExpiresHours) * time.Hour)

	// Create JWT token with JTI claim
	claims := jwt.MapClaims{
		"jti":      jti,
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		response.InternalError(c, "Failed to generate token", err)
		return
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
	loginResponse := types.LoginResponse{
		Token:     tokenString,
		JTI:       jti,
		ExpiresAt: expiresAt,
		User:      userInfo,
	}

	response.Success(c, loginResponse)
}

// LogoutHandler handles user logout requests
// In the future, this can be enhanced to revoke the session
func (h *AuthHandler) LogoutHandler(c *gin.Context) {
	// Extract JTI from token if available
	// For now, just return success
	// TODO: Implement session revocation in Phase 4
	response.SuccessWithMessage(c, "Logged out successfully", nil)
}

// GetCurrentUserHandler returns the current authenticated user's information
func (h *AuthHandler) GetCurrentUserHandler(c *gin.Context) {
	// Extract user ID from JWT claims (set by JWT middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated", nil)
		return
	}

	var user types.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		response.NotFound(c, "User not found", err)
		return
	}

	// Retrieve user roles
	var roles []types.Role
	h.db.Table("roles").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", user.ID).
		Find(&roles)

	userInfo := &types.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		RealName: user.RealName,
		Avatar:   user.Avatar,
		Roles:    roles,
	}

	response.Success(c, userInfo)
}

// GetAccessCodesHandler returns an array of permission codes for the authenticated user
// This is used by the frontend to determine what features/pages the user can access
func (h *AuthHandler) GetAccessCodesHandler(c *gin.Context) {
	// Extract user ID from JWT claims (set by JWT middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated", nil)
		return
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
		response.InternalError(c, "Failed to retrieve access codes", err)
		return
	}

	// Return empty array if no permissions found (instead of null)
	if codes == nil {
		codes = []string{}
	}

	response.Success(c, codes)
}
