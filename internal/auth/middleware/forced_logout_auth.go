package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ForcedLogoutAuthMiddleware provides authorization middleware for forced logout operations
type ForcedLogoutAuthMiddleware struct {
	db *gorm.DB
}

// NewForcedLogoutAuthMiddleware creates a new authorization middleware
func NewForcedLogoutAuthMiddleware(db *gorm.DB) *ForcedLogoutAuthMiddleware {
	return &ForcedLogoutAuthMiddleware{
		db: db,
	}
}

// RequireSessionAdmin middleware checks if the user has session-admin role or higher
// This middleware should be applied after JWT authentication middleware
func (m *ForcedLogoutAuthMiddleware) RequireSessionAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract user ID from context (set by JWT middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "User not authenticated",
			})
			c.Abort()
			return
		}

		// Extract username from context (for logging)
		username, _ := c.Get("username")

		// Get user's roles from database
		roles, err := m.getUserRoles(userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to retrieve user roles",
			})
			c.Abort()
			return
		}

		// Check if user has required role
		if !m.hasRequiredRole(roles) {
			// Log authorization failure
			c.Header("X-Authorization-Failed", "insufficient_permissions")

			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Insufficient permissions. This operation requires 'session-admin' or 'superadmin' role.",
				"details": map[string]interface{}{
					"required_roles": []string{"session-admin", "superadmin"},
					"user_roles":     roles,
				},
			})
			c.Abort()
			return
		}

		// Authorization successful - store actor info in context for audit trail
		c.Set("actor_id", userID)
		c.Set("actor_username", username)
		c.Set("actor_roles", roles)

		// Log successful authorization (for audit purposes)
		c.Header("X-Authorization-Success", "true")

		c.Next()
	}
}

// getUserRoles retrieves all role codes for a user
func (m *ForcedLogoutAuthMiddleware) getUserRoles(userID string) ([]string, error) {
	var roles []struct {
		Code string
	}

	// Query user roles through user_roles join
	err := m.db.Table("roles").
		Select("roles.code").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.status = 1", userID).
		Scan(&roles).Error
	if err != nil {
		return nil, err
	}

	// Extract role codes
	roleCodes := make([]string, len(roles))
	for i, role := range roles {
		roleCodes[i] = role.Code
	}

	return roleCodes, nil
}

// hasRequiredRole checks if user has session-admin or superadmin role
// Per Q1 clarification: superadmin can also perform forced logout operations
func (m *ForcedLogoutAuthMiddleware) hasRequiredRole(roles []string) bool {
	for _, role := range roles {
		roleCode := strings.ToLower(role)
		if roleCode == "session-admin" || roleCode == "superadmin" || roleCode == "super_admin" {
			return true
		}
	}
	return false
}

// ExtractActorInfo extracts actor information from context
// This is a helper function for handlers to get actor details for audit logging
func ExtractActorInfo(c *gin.Context) (actorID, actorUsername, actorIPAddress string) {
	if id, exists := c.Get("actor_id"); exists {
		actorID = id.(string)
	}
	if username, exists := c.Get("actor_username"); exists {
		actorUsername = username.(string)
	}

	// Extract IP address
	actorIPAddress = c.ClientIP()
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		actorIPAddress = strings.Split(xff, ",")[0]
	}

	return
}
