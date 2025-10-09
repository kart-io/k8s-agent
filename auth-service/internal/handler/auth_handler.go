package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/auth-service/internal/service"
	"github.com/kart-io/k8s-agent/auth-service/pkg/types"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login handles user login
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req types.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": err.Error(),
		})
		return
	}

	resp, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"code":    401,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Logout handles user logout
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetString("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": "Token not found in context",
		})
		return
	}

	if err := h.authService.Logout(token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// Me retrieves current user information
// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"code":    401,
			"details": "User ID not found in context",
		})
		return
	}

	user, err := h.authService.GetCurrentUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"code":    404,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Menus retrieves user menu tree
// GET /api/v1/auth/menus
func (h *AuthHandler) Menus(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"code":    401,
			"details": "User ID not found in context",
		})
		return
	}

	menus, err := h.authService.GetUserMenus(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"menus": menus,
	})
}

// CheckPermission checks if user has permission (public endpoint for frontend)
// POST /api/v1/auth/check
func (h *AuthHandler) CheckPermission(c *gin.Context) {
	var req types.PermissionCheck
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": err.Error(),
		})
		return
	}

	// TODO: Implement permission check logic
	// For now, return success
	c.JSON(http.StatusOK, gin.H{
		"has_permission": true,
	})
}
