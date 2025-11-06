package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/internal/auth/service"
	"github.com/kart-io/k8s-agent/internal/auth/types"
)

// PermissionHandler handles permission management HTTP requests.
type PermissionHandler struct {
	permissionService *service.PermissionService
}

// NewPermissionHandler creates a new permission handler.
func NewPermissionHandler(permissionService *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
	}
}

// List retrieves all permissions with optional filtering
// GET /api/v1/permissions.
func (h *PermissionHandler) List(c *gin.Context) {
	handler := WithQueryParams(h.listLogic)
	handler(c)
}

// listLogic contains the core business logic for listing permissions
func (h *PermissionHandler) listLogic(c *gin.Context, params *struct {
	Type   string `form:"type"`
	Status string `form:"status"`
}) (*map[string]interface{}, error) {
	permissions, err := h.permissionService.List(params.Type, params.Status)
	if err != nil {
		return nil, err
	}

	return &map[string]interface{}{
		"items": permissions,
	}, nil
}

// GetTree retrieves permission tree
// GET /api/v1/permissions/tree.
func (h *PermissionHandler) GetTree(c *gin.Context) {
	handler := WithNoRequest(h.getTreeLogic)
	handler(c)
}

// getTreeLogic contains the core business logic for getting permission tree
func (h *PermissionHandler) getTreeLogic(c *gin.Context) (*map[string]interface{}, error) {
	tree, err := h.permissionService.GetTree()
	if err != nil {
		return nil, err
	}

	return &map[string]interface{}{
		"tree": tree,
	}, nil
}

// GetByID retrieves a permission by ID
// GET /api/v1/permissions/:id.
func (h *PermissionHandler) GetByID(c *gin.Context) {
	handler := WithURIParams(h.getByIDLogic)
	handler(c)
}

// getByIDLogic contains the core business logic for getting permission by ID
func (h *PermissionHandler) getByIDLogic(c *gin.Context, params *struct {
	ID string `uri:"id" binding:"required"`
}) (*types.Permission, error) {
	return h.permissionService.GetByID(params.ID)
}

// Create creates a new permission
// POST /api/v1/permissions.
func (h *PermissionHandler) Create(c *gin.Context) {
	handler := WithJSONRequestCreated(h.createLogic)
	handler(c)
}

// createLogic contains the core business logic for creating a permission
func (h *PermissionHandler) createLogic(c *gin.Context, req *types.PermissionRequest) (*types.Permission, error) {
	return h.permissionService.Create(req)
}

// Update updates a permission
// PUT /api/v1/permissions/:id.
func (h *PermissionHandler) Update(c *gin.Context) {
	handler := WithURIAndJSONRequest(h.updateLogic)
	handler(c)
}

// updateLogic contains the core business logic for updating a permission
func (h *PermissionHandler) updateLogic(c *gin.Context, req *struct {
	ID   string `uri:"id" binding:"required"`
	Body types.PermissionRequest
}) error {
	return h.permissionService.Update(req.ID, &req.Body)
}

// Delete deletes a permission
// DELETE /api/v1/permissions/:id.
func (h *PermissionHandler) Delete(c *gin.Context) {
	handler := WithURIParamsNoResponse(h.deleteLogic)
	handler(c)
}

// deleteLogic contains the core business logic for deleting a permission
func (h *PermissionHandler) deleteLogic(c *gin.Context, params *struct {
	ID string `uri:"id" binding:"required"`
}) error {
	return h.permissionService.Delete(params.ID)
}
