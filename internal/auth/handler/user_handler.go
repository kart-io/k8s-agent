package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/internal/auth/service"
	"github.com/kart-io/k8s-agent/internal/auth/types"
)

// UserHandler handles user management HTTP requests
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// List retrieves list of users with pagination
// GET /api/v1/users
func (h *UserHandler) List(c *gin.Context) {
	// Parse pagination parameters
	var params types.PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": err.Error(),
		})
		return
	}

	// Parse status filter
	var statusFilter *int
	if statusStr := c.Query("status"); statusStr != "" {
		status, err := strconv.Atoi(statusStr)
		if err == nil {
			statusFilter = &status
		}
	}

	resp, err := h.userService.List(params, statusFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetByID retrieves a user by ID
// GET /api/v1/users/:id
func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": "User ID is required",
		})
		return
	}

	user, err := h.userService.GetByID(id)
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

// Create creates a new user
// POST /api/v1/users
func (h *UserHandler) Create(c *gin.Context) {
	var req types.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": err.Error(),
		})
		return
	}

	user, err := h.userService.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Update updates user information
// PUT /api/v1/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": "User ID is required",
		})
		return
	}

	var req types.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": err.Error(),
		})
		return
	}

	if err := h.userService.Update(id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
	})
}

// Delete soft deletes a user
// DELETE /api/v1/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": "User ID is required",
		})
		return
	}

	if err := h.userService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// AssignRoles assigns roles to a user
// POST /api/v1/users/:id/roles
func (h *UserHandler) AssignRoles(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": "User ID is required",
		})
		return
	}

	var req struct {
		RoleIDs []string `json:"role_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"code":    400,
			"details": err.Error(),
		})
		return
	}

	if err := h.userService.AssignRoles(id, req.RoleIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"code":    500,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Roles assigned successfully",
	})
}
