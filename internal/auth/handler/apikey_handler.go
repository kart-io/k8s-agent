package handler

import (
	"errors"

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
	handler := WithNoRequest(h.listLogic)
	handler(c)
}

// listLogic contains the core business logic for listing API keys
func (h *APIKeyHandler) listLogic(c *gin.Context) (*map[string]interface{}, error) {
	userID := c.GetString("user_id")
	if userID == "" {
		return nil, errors.New("user ID not found in context")
	}

	keys, err := h.apikeyService.List(userID)
	if err != nil {
		return nil, err
	}

	return &map[string]interface{}{
		"items": keys,
	}, nil
}

// Create creates a new API key
// POST /api/v1/api-keys.
func (h *APIKeyHandler) Create(c *gin.Context) {
	handler := WithJSONRequestCreated(h.createLogic)
	handler(c)
}

// createLogic contains the core business logic for creating an API key
func (h *APIKeyHandler) createLogic(c *gin.Context, req *types.APIKeyCreateRequest) (*map[string]interface{}, error) {
	userID := c.GetString("user_id")
	if userID == "" {
		return nil, errors.New("user ID not found in context")
	}

	keyWithSecret, err := h.apikeyService.Create(userID, req)
	if err != nil {
		return nil, err
	}

	return &map[string]interface{}{
		"id":          keyWithSecret.ID,
		"name":        keyWithSecret.Name,
		"key":         keyWithSecret.Key,
		"secret":      keyWithSecret.SecretPlain, // Shown only once
		"description": keyWithSecret.Description,
		"expires_at":  keyWithSecret.ExpiresAt,
		"created_at":  keyWithSecret.CreatedAt,
		"warning":     "Save the secret now. You won't be able to see it again!",
	}, nil
}

// Delete deletes an API key
// DELETE /api/v1/api-keys/:id.
func (h *APIKeyHandler) Delete(c *gin.Context) {
	handler := WithURIParamsNoResponse(h.deleteLogic)
	handler(c)
}

// deleteLogic contains the core business logic for deleting an API key
func (h *APIKeyHandler) deleteLogic(c *gin.Context, params *struct {
	ID string `uri:"id" binding:"required"`
}) error {
	userID := c.GetString("user_id")
	if userID == "" {
		return errors.New("user ID not found in context")
	}

	return h.apikeyService.Delete(params.ID, userID)
}
