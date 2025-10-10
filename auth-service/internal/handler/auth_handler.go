package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kart-io/k8s-agent/auth-service/pkg/forced-logout/session"
	"github.com/kart-io/k8s-agent/auth-service/pkg/types"
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Retrieve user from database
	var user types.User
	if err := h.db.Where("username = ?", loginReq.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid username or password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error",
		})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid username or password",
		})
		return
	}

	// Check if user is active
	if user.Status != 1 {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "User account is disabled",
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	// Extract IP address from request
	ipAddress := h.extractIPAddress(c)

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
	response := types.LoginResponse{
		Token:     tokenString,
		JTI:       jti,
		ExpiresAt: expiresAt,
		User:      userInfo,
	}

	c.JSON(http.StatusOK, response)
}

// extractIPAddress extracts the client IP address from the request
// Handles X-Forwarded-For, X-Real-IP headers and direct connection
func (h *AuthHandler) extractIPAddress(c *gin.Context) string {
	// Check X-Forwarded-For header (may contain multiple IPs)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain (original client IP)
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header (single IP)
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := c.ClientIP()
	if ip == "" {
		ip = "Unknown"
	}

	return ip
}

// LogoutHandler handles user logout requests
// In the future, this can be enhanced to revoke the session
func (h *AuthHandler) LogoutHandler(c *gin.Context) {
	// Extract JTI from token if available
	// For now, just return success
	// TODO: Implement session revocation in Phase 4
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// GetCurrentUserHandler returns the current authenticated user's information
func (h *AuthHandler) GetCurrentUserHandler(c *gin.Context) {
	// Extract user ID from JWT claims (set by JWT middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var user types.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
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

	c.JSON(http.StatusOK, userInfo)
}
