package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/auth-service/internal/service"
	"github.com/kart-io/k8s-agent/auth-service/pkg/types"
)

// PermissionHandler handles permission management HTTP requests
type PermissionHandler struct {
	permissionService *service.PermissionService
}

// NewPermissionHandler creates a new permission handler
func NewPermissionHandler(permissionService *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
	}
}

// List retrieves all permissions with optional filtering
// GET /api/v1/permissions
func (h *PermissionHandler) List(c *gin.Context) {
	typeFilter := c.Query("type")
	statusFilter := c.Query("status")

	permissions, err := h.permissionService.List(typeFilter, statusFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": permissions,
	})
}

// GetTree retrieves permission tree
// GET /api/v1/permissions/tree
func (h *PermissionHandler) GetTree(c *gin.Context) {
	tree, err := h.permissionService.GetTree()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tree": tree,
	})
}

// GetByID retrieves a permission by ID
// GET /api/v1/permissions/:id
func (h *PermissionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": "Permission ID is required",
		})
		return
	}

	permission, err := h.permissionService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"code":    404,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, permission)
}

// Create creates a new permission
// POST /api/v1/permissions
func (h *PermissionHandler) Create(c *gin.Context) {
	var req types.PermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": err.Error(),
		})
		return
	}

	permission, err := h.permissionService.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, permission)
}

// Update updates a permission
// PUT /api/v1/permissions/:id
func (h *PermissionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": "Permission ID is required",
		})
		return
	}

	var req types.PermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": err.Error(),
		})
		return
	}

	if err := h.permissionService.Update(id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Permission updated successfully",
	})
}

// Delete deletes a permission
// DELETE /api/v1/permissions/:id
func (h *PermissionHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": "Permission ID is required",
		})
		return
	}

	if err := h.permissionService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Permission deleted successfully",
	})
}
