package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/auth-service/internal/service"
	"github.com/kart-io/k8s-agent/auth-service/pkg/types"
)

// RoleHandler handles role management HTTP requests
type RoleHandler struct {
	roleService *service.RoleService
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// List retrieves all roles
// GET /api/v1/roles
func (h *RoleHandler) List(c *gin.Context) {
	roles, err := h.roleService.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": roles,
	})
}

// GetByID retrieves a role by ID
// GET /api/v1/roles/:id
func (h *RoleHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": "Role ID is required",
		})
		return
	}

	role, err := h.roleService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"code":    404,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, role)
}

// Create creates a new role
// POST /api/v1/roles
func (h *RoleHandler) Create(c *gin.Context) {
	var req types.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": err.Error(),
		})
		return
	}

	role, err := h.roleService.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// Update updates a role
// PUT /api/v1/roles/:id
func (h *RoleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": "Role ID is required",
		})
		return
	}

	var req types.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": err.Error(),
		})
		return
	}

	if err := h.roleService.Update(id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Role updated successfully",
	})
}

// Delete deletes a role
// DELETE /api/v1/roles/:id
func (h *RoleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": "Role ID is required",
		})
		return
	}

	if err := h.roleService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Role deleted successfully",
	})
}

// AssignPermissions assigns permissions to a role
// POST /api/v1/roles/:id/permissions
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": "Role ID is required",
		})
		return
	}

	var req struct {
		PermissionIDs []string `json:"permission_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": err.Error(),
		})
		return
	}

	if err := h.roleService.AssignPermissions(id, req.PermissionIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Permissions assigned successfully",
	})
}

// GetPermissions retrieves permissions for a role
// GET /api/v1/roles/:id/permissions
func (h *RoleHandler) GetPermissions(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": "Role ID is required",
		})
		return
	}

	permissions, err := h.roleService.GetPermissions(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": permissions,
	})
}
