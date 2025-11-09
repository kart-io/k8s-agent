package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/internal/auth/service"
	"github.com/kart-io/k8s-agent/internal/auth/types"
)

// UserHandler handles user management HTTP requests.
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// List retrieves list of users with pagination
// GET /api/v1/users.
// Refactored to use WithQueryParams decorator
func (h *UserHandler) List(c *gin.Context) {
	handler := WithQueryParams(h.listLogic)
	handler(c)
}

// listLogic contains the core business logic for listing users
func (h *UserHandler) listLogic(c *gin.Context, params *types.PaginationParams) (*types.PaginatedResponse, error) {
	// Parse status filter from query (not part of PaginationParams)
	var statusFilter *int
	if statusStr := c.Query("status"); statusStr != "" {
		// Status filter logic handled in service
		statusVal := 1 // active status
		statusFilter = &statusVal
	}

	return h.userService.List(*params, statusFilter)
}

// GetByID retrieves a user by ID
// GET /api/v1/users/:id.
func (h *UserHandler) GetByID(c *gin.Context) {
	handler := WithNoRequest(h.getByIDLogic)
	handler(c)
}

// getByIDLogic contains the core business logic for getting user by ID
func (h *UserHandler) getByIDLogic(c *gin.Context) (*types.User, error) {
	id := c.Param("id")
	if id == "" {
		return nil, &ValidationError{Message: "User ID is required"}
	}

	return h.userService.GetByID(id)
}

// Create creates a new user
// POST /api/v1/users.
// Refactored to use WithJSONRequestCreated decorator
func (h *UserHandler) Create(c *gin.Context) {
	handler := WithJSONRequestCreated(h.createLogic)
	handler(c)
}

// createLogic contains the core business logic for creating a user
func (h *UserHandler) createLogic(c *gin.Context, req *types.UserCreateRequest) (*types.User, error) {
	return h.userService.Create(req)
}

// Update updates user information
// PUT /api/v1/users/:id.
// Refactored to use WithJSONRequestNoResponse decorator
func (h *UserHandler) Update(c *gin.Context) {
	handler := WithJSONRequestNoResponse(h.updateLogic)
	handler(c)
}

// updateLogic contains the core business logic for updating a user
func (h *UserHandler) updateLogic(c *gin.Context, req *types.UserUpdateRequest) error {
	id := c.Param("id")
	if id == "" {
		return &ValidationError{Message: "User ID is required"}
	}

	return h.userService.Update(id, req)
}

// Delete soft deletes a user
// DELETE /api/v1/users/:id.
func (h *UserHandler) Delete(c *gin.Context) {
	handler := WithNoRequest(h.deleteLogic)
	handler(c)
}

// deleteLogic contains the core business logic for deleting a user
func (h *UserHandler) deleteLogic(c *gin.Context) (*SuccessMessage, error) {
	id := c.Param("id")
	if id == "" {
		return nil, &ValidationError{Message: "User ID is required"}
	}

	if err := h.userService.Delete(id); err != nil {
		return nil, err
	}

	return &SuccessMessage{Message: "User deleted successfully"}, nil
}

// AssignRoles assigns roles to a user
// POST /api/v1/users/:id/roles.
func (h *UserHandler) AssignRoles(c *gin.Context) {
	handler := WithJSONRequestNoResponse(h.assignRolesLogic)
	handler(c)
}

// assignRolesLogic contains the core business logic for assigning roles
func (h *UserHandler) assignRolesLogic(c *gin.Context, req *struct {
	RoleIDs []string `json:"role_ids" binding:"required"`
},
) error {
	id := c.Param("id")
	if id == "" {
		return &ValidationError{Message: "User ID is required"}
	}

	return h.userService.AssignRoles(id, req.RoleIDs)
}

// SuccessMessage is a simple response type for operations without data
type SuccessMessage struct {
	Message string `json:"message"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
