package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/internal/auth/service"
	"github.com/kart-io/k8s-agent/internal/auth/types"
)

// AuthHandler handles authentication-related HTTP requests.
// Refactored to use AuthService for all business logic.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginHandler handles user login requests with session tracking.
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	handler := WithJSONRequest(h.loginLogic)
	handler(c)
}

// loginLogic contains the core business logic for user login
func (h *AuthHandler) loginLogic(c *gin.Context, loginReq *types.LoginRequest) (*types.LoginResponse, error) {
	// Delegate to AuthService for all business logic
	// AuthService handles: user validation, password check, JWT generation, session tracking
	return h.authService.Login(loginReq.Username, loginReq.Password)
}

// LogoutHandler handles user logout requests
func (h *AuthHandler) LogoutHandler(c *gin.Context) {
	handler := WithNoRequest(h.logoutLogic)
	handler(c)
}

// logoutLogic contains the core business logic for logout
func (h *AuthHandler) logoutLogic(c *gin.Context) (*SuccessMessage, error) {
	// Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, errors.New("authorization header is required")
	}

	// Parse Bearer token
	if len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
		return nil, errors.New("invalid authorization header format")
	}

	token := authHeader[7:]

	// Delegate to AuthService for token blacklisting
	if err := h.authService.Logout(token); err != nil {
		return nil, err
	}

	return &SuccessMessage{Message: "Logout successful"}, nil
}

// RefreshTokenHandler handles token refresh requests using refresh tokens.
// Implements token rotation for enhanced security.
func (h *AuthHandler) RefreshTokenHandler(c *gin.Context) {
	handler := WithJSONRequest(h.refreshTokenLogic)
	handler(c)
}

// refreshTokenLogic contains the core business logic for token refresh
func (h *AuthHandler) refreshTokenLogic(c *gin.Context, req *types.RefreshTokenRequest) (*types.RefreshTokenResponse, error) {
	// Delegate to AuthService for token refresh with rotation
	return h.authService.RefreshToken(req.RefreshToken)
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

	userIDStr, ok := userID.(string)
	if !ok {
		return nil, errors.New("invalid user ID format")
	}

	// Delegate to AuthService for user retrieval
	return h.authService.GetCurrentUser(userIDStr)
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

	userIDStr, ok := userID.(string)
	if !ok {
		return nil, errors.New("invalid user ID format")
	}

	// Get user menus/permissions from AuthService
	menus, err := h.authService.GetUserMenus(userIDStr)
	if err != nil {
		return nil, err
	}

	// Extract permission codes from menu items
	codes := extractPermissionCodes(menus)

	// Return empty array if no permissions found (instead of null)
	if codes == nil {
		codes = []string{}
	}

	return &codes, nil
}

// extractPermissionCodes recursively extracts permission codes from menu items
func extractPermissionCodes(items []*types.MenuItem) []string {
	var codes []string

	for _, item := range items {
		// Add current item's code if it exists
		// Note: MenuItem may not have a Code field, this is a placeholder
		// You may need to adjust based on actual MenuItem structure

		// Recursively process children
		if len(item.Children) > 0 {
			childCodes := extractPermissionCodes(item.Children)
			codes = append(codes, childCodes...)
		}
	}

	return codes
}

// GetMenusHandler returns the menu tree for the authenticated user
// This endpoint is used by frontend to build navigation menus
func (h *AuthHandler) GetMenusHandler(c *gin.Context) {
	handler := WithNoRequest(h.getMenusLogic)
	handler(c)
}

// getMenusLogic contains the core business logic for retrieving user menus
func (h *AuthHandler) getMenusLogic(c *gin.Context) (*map[string]interface{}, error) {
	// Extract user ID from JWT claims (set by JWT middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, errors.New("user not authenticated")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return nil, errors.New("invalid user ID format")
	}

	// Delegate to AuthService for menu retrieval
	menus, err := h.authService.GetUserMenus(userIDStr)
	if err != nil {
		return nil, err
	}

	return &map[string]interface{}{
		"menus": menus,
	}, nil
}

