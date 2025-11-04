package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/internal/auth/service"
	"github.com/kart-io/k8s-agent/internal/auth/types"
)

// APIKeyHandler handles API key management HTTP requests.
type APIKeyHandler struct {
	apikeyService *service.APIKeyService
}

// NewAPIKeyHandler creates a new API key handler.
func NewAPIKeyHandler(apikeyService *service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{
		apikeyService: apikeyService,
	}
}

// List retrieves user's API keys (secrets are masked)
// GET /api/v1/api-keys.
func (h *APIKeyHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"code":    401,
			"details": "User ID not found in context",
		})
		return
	}

	keys, err := h.apikeyService.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": keys,
	})
}

// Create creates a new API key
// POST /api/v1/api-keys.
func (h *APIKeyHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"code":    401,
			"details": "User ID not found in context",
		})
		return
	}

	var req types.APIKeyCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": err.Error(),
		})
		return
	}

	keyWithSecret, err := h.apikeyService.Create(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          keyWithSecret.ID,
		"name":        keyWithSecret.Name,
		"key":         keyWithSecret.Key,
		"secret":      keyWithSecret.SecretPlain, // Shown only once
		"description": keyWithSecret.Description,
		"expires_at":  keyWithSecret.ExpiresAt,
		"created_at":  keyWithSecret.CreatedAt,
		"warning":     "Save the secret now. You won't be able to see it again!",
	})
}

// Delete deletes an API key
// DELETE /api/v1/api-keys/:id.
func (h *APIKeyHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": "API key ID is required",
		})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"code":    401,
			"details": "User ID not found in context",
		})
		return
	}

	if err := h.apikeyService.Delete(id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API key deleted successfully",
	})
}
