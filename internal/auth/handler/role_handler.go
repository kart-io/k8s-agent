package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/internal/auth/service"
	"github.com/kart-io/k8s-agent/internal/auth/types"
)

// RoleHandler handles role management HTTP requests.
type RoleHandler struct {
	roleService *service.RoleService
}

// NewRoleHandler creates a new role handler.
func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// List retrieves all roles
// GET /api/v1/roles.
func (h *RoleHandler) List(c *gin.Context) {
	handler := WithNoRequest(h.listLogic)
	handler(c)
}

// listLogic contains the core business logic for listing roles
func (h *RoleHandler) listLogic(c *gin.Context) (*map[string]interface{}, error) {
	roles, err := h.roleService.List()
	if err != nil {
		return nil, err
	}

	return &map[string]interface{}{
		"items": roles,
	}, nil
}

// GetByID retrieves a role by ID
// GET /api/v1/roles/:id.
func (h *RoleHandler) GetByID(c *gin.Context) {
	handler := WithURIParams(h.getByIDLogic)
	handler(c)
}

// getByIDLogic contains the core business logic for getting role by ID
func (h *RoleHandler) getByIDLogic(c *gin.Context, params *struct {
	ID string `uri:"id" binding:"required"`
},
) (*types.Role, error) {
	return h.roleService.GetByID(params.ID)
}

// Create creates a new role
// POST /api/v1/roles.
func (h *RoleHandler) Create(c *gin.Context) {
	handler := WithJSONRequestCreated(h.createLogic)
	handler(c)
}

// createLogic contains the core business logic for creating a role
func (h *RoleHandler) createLogic(c *gin.Context, req *types.RoleRequest) (*types.Role, error) {
	return h.roleService.Create(req)
}

// Update updates a role
// PUT /api/v1/roles/:id.
func (h *RoleHandler) Update(c *gin.Context) {
	handler := WithURIAndJSONRequest(h.updateLogic)
	handler(c)
}

// updateLogic contains the core business logic for updating a role
func (h *RoleHandler) updateLogic(c *gin.Context, req *struct {
	ID   string `uri:"id" binding:"required"`
	Body types.RoleRequest
},
) error {
	return h.roleService.Update(req.ID, &req.Body)
}

// Delete deletes a role
// DELETE /api/v1/roles/:id.
func (h *RoleHandler) Delete(c *gin.Context) {
	handler := WithURIParamsNoResponse(h.deleteLogic)
	handler(c)
}

// deleteLogic contains the core business logic for deleting a role
func (h *RoleHandler) deleteLogic(c *gin.Context, params *struct {
	ID string `uri:"id" binding:"required"`
},
) error {
	return h.roleService.Delete(params.ID)
}

// AssignPermissions assigns permissions to a role
// POST /api/v1/roles/:id/permissions.
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	handler := WithURIAndJSONRequestNoResponse(h.assignPermissionsLogic)
	handler(c)
}

// assignPermissionsLogic contains the core business logic for assigning permissions
func (h *RoleHandler) assignPermissionsLogic(c *gin.Context, req *struct {
	ID            string   `uri:"id" binding:"required"`
	PermissionIDs []string `json:"permission_ids" binding:"required"`
},
) error {
	return h.roleService.AssignPermissions(req.ID, req.PermissionIDs)
}

// GetPermissions retrieves permissions for a role
// GET /api/v1/roles/:id/permissions.
func (h *RoleHandler) GetPermissions(c *gin.Context) {
	handler := WithURIParams(h.getPermissionsLogic)
	handler(c)
}

// getPermissionsLogic contains the core business logic for getting role permissions
func (h *RoleHandler) getPermissionsLogic(c *gin.Context, params *struct {
	ID string `uri:"id" binding:"required"`
},
) (*map[string]interface{}, error) {
	permissions, err := h.roleService.GetPermissions(params.ID)
	if err != nil {
		return nil, err
	}

	return &map[string]interface{}{
		"permissions": permissions,
	}, nil
}
