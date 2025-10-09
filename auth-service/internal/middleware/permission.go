package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/auth-service/internal/model"
	"github.com/kart-io/k8s-agent/auth-service/internal/storage"
)

// RequirePermission checks if user has specific permission
func RequirePermission(db *storage.PostgresDB, permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"code":    401,
				"details": "User ID not found in context",
			})
			c.Abort()
			return
		}

		// Check if user has permission
		hasPermission, err := checkUserPermission(db, userID, permissionCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"code":    500,
				"details": "Failed to check permission",
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"code":    403,
				"details": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole checks if user has specific role
func RequireRole(db *storage.PostgresDB, roleCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"code":    401,
				"details": "User ID not found in context",
			})
			c.Abort()
			return
		}

		// Check if user has role
		hasRole, err := checkUserRole(db, userID, roleCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"code":    500,
				"details": "Failed to check role",
			})
			c.Abort()
			return
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"code":    403,
				"details": "Insufficient role",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkUserPermission verifies if user has a specific permission
func checkUserPermission(db *storage.PostgresDB, userID, permissionCode string) (bool, error) {
	var count int64
	err := db.DB.Model(&model.Permission{}).
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ? AND permissions.code = ? AND permissions.status = ?", userID, permissionCode, 1).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// checkUserRole verifies if user has a specific role
func checkUserRole(db *storage.PostgresDB, userID, roleCode string) (bool, error) {
	var count int64
	err := db.DB.Model(&model.Role{}).
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.code = ? AND roles.status = ?", userID, roleCode, 1).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
