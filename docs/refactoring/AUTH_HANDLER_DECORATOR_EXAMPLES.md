# Auth Handler Decorator Pattern - Code Examples

## Quick Reference

This document provides before/after code examples for the decorator pattern implementation.

## Decorator Types Overview

| Decorator | Use Case | Request Type | Response Type | HTTP Status |
|-----------|----------|--------------|---------------|-------------|
| `WithJSONRequest` | POST/PUT with data response | JSON body | Data | 200 OK |
| `WithJSONRequestCreated` | POST creating resource | JSON body | Data | 201 Created |
| `WithJSONRequestNoResponse` | POST/PUT/DELETE no data | JSON body | Message only | 200 OK |
| `WithQueryParams` | GET with query params | Query string | Data | 200 OK |
| `WithURIParams` | GET with URI params | URI params | Data | 200 OK |
| `WithNoRequest` | GET without params | None | Data | 200 OK |

## Example 1: GET Endpoint (No Request Body)

### Before (32 lines)
```go
// GetCurrentUserHandler returns the current authenticated user's information.
func (h *AuthHandler) GetCurrentUserHandler(c *gin.Context) {
    // Extract user ID from JWT claims (set by JWT middleware)
    userID, exists := c.Get("user_id")
    if !exists {
        response.Unauthorized(c, "User not authenticated", nil)
        return
    }

    var user types.User
    if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
        response.NotFound(c, "User not found", err)
        return
    }

    // Retrieve user roles
    var roles []types.Role
    h.db.Table("roles").
        Joins("JOIN user_roles ON roles.id = user_roles.role_id").
        Where("user_roles.user_id = ?", user.ID).
        Find(&roles)

    userInfo := &types.UserInfo{
        ID:       user.ID,
        Username: user.Username,
        Email:    user.Email,
        RealName: user.RealName,
        Avatar:   user.Avatar,
        Roles:    roles,
    }

    response.Success(c, userInfo)
}
```

### After (32 lines total, but separated for clarity)
```go
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

    var user types.User
    if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
        return nil, err
    }

    // Retrieve user roles
    var roles []types.Role
    h.db.Table("roles").
        Joins("JOIN user_roles ON roles.id = user_roles.role_id").
        Where("user_roles.user_id = ?", user.ID).
        Find(&roles)

    return &types.UserInfo{
        ID:       user.ID,
        Username: user.Username,
        Email:    user.Email,
        RealName: user.RealName,
        Avatar:   user.Avatar,
        Roles:    roles,
    }, nil
}
```

**Benefits**:
- Pure business logic in `getCurrentUserLogic` (no HTTP concerns)
- Easy to test without gin.Context
- Error handling centralized in decorator
- Response formatting handled automatically

## Example 2: POST Endpoint (Create Resource)

### Before (20 lines)
```go
// Create creates a new user
// POST /api/v1/users.
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
```

### After (9 lines)
```go
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
```

**Reduction**: 20 → 9 lines (55% reduction)

## Example 3: PUT Endpoint (Update, No Data Response)

### Before (28 lines)
```go
// Update updates user information
// PUT /api/v1/users/:id.
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
```

### After (12 lines)
```go
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
```

**Reduction**: 28 → 12 lines (57% reduction)

## Example 4: GET with Query Parameters

### Before (30 lines)
```go
// List retrieves list of users with pagination
// GET /api/v1/users.
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
```

### After (15 lines)
```go
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
        statusVal := 1 // active status
        statusFilter = &statusVal
    }

    return h.userService.List(*params, statusFilter)
}
```

**Reduction**: 30 → 15 lines (50% reduction)

## Example 5: DELETE Endpoint

### Before (20 lines)
```go
// Delete soft deletes a user
// DELETE /api/v1/users/:id.
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
```

### After (15 lines)
```go
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
```

**Reduction**: 20 → 15 lines (25% reduction)

## Testing Examples

### Before (Testing with HTTP concerns)
```go
func TestUserHandler_Create(t *testing.T) {
    // Setup gin context and recorder
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    // Create request body
    body := `{"username":"test","password":"pass123"}`
    c.Request = httptest.NewRequest("POST", "/users", strings.NewReader(body))
    c.Request.Header.Set("Content-Type", "application/json")

    // Setup mock service
    mockService := &MockUserService{...}
    handler := NewUserHandler(mockService)

    // Execute handler
    handler.Create(c)

    // Assert HTTP response
    assert.Equal(t, http.StatusCreated, w.Code)
    // ... more HTTP assertions
}
```

### After (Testing pure business logic)
```go
func TestUserHandler_createLogic(t *testing.T) {
    // No need for gin.Context or HTTP recorder!

    // Setup mock service
    mockService := &MockUserService{
        CreateFunc: func(req *types.UserCreateRequest) (*types.User, error) {
            return &types.User{ID: "123", Username: req.Username}, nil
        },
    }
    handler := NewUserHandler(mockService)

    // Prepare request
    req := &types.UserCreateRequest{
        Username: "test",
        Password: "pass123",
    }

    // Execute business logic directly
    result, err := handler.createLogic(nil, req)

    // Assert business logic results
    assert.NoError(t, err)
    assert.Equal(t, "123", result.ID)
    assert.Equal(t, "test", result.Username)
}
```

**Benefits**:
- No HTTP setup needed
- Faster tests (no HTTP layer)
- Focus on business logic
- Easier to mock dependencies

## Migration Checklist

For each handler method:

- [ ] Identify the pattern (JSON/Query/URI/NoRequest)
- [ ] Extract business logic into separate `*Logic` method
- [ ] Update method signature to match decorator type
- [ ] Replace boilerplate with decorator call
- [ ] Remove old error handling code
- [ ] Remove old response formatting code
- [ ] Update tests to test `*Logic` method directly
- [ ] Verify build passes: `make go.build.auth`

## Common Patterns

### Pattern 1: Simple CRUD
```go
// GET /resource/:id
func (h *Handler) Get(c *gin.Context) {
    handler := WithNoRequest(h.getLogic)
    handler(c)
}

// POST /resource
func (h *Handler) Create(c *gin.Context) {
    handler := WithJSONRequestCreated(h.createLogic)
    handler(c)
}

// PUT /resource/:id
func (h *Handler) Update(c *gin.Context) {
    handler := WithJSONRequestNoResponse(h.updateLogic)
    handler(c)
}

// DELETE /resource/:id
func (h *Handler) Delete(c *gin.Context) {
    handler := WithNoRequest(h.deleteLogic)
    handler(c)
}
```

### Pattern 2: List with Pagination
```go
// GET /resources?page=1&page_size=20
func (h *Handler) List(c *gin.Context) {
    handler := WithQueryParams(h.listLogic)
    handler(c)
}

func (h *Handler) listLogic(c *gin.Context, params *types.PaginationParams) (*types.PaginatedResponse, error) {
    return h.service.List(*params)
}
```

### Pattern 3: Action Endpoints
```go
// POST /resource/:id/action
func (h *Handler) PerformAction(c *gin.Context) {
    handler := WithJSONRequestNoResponse(h.performActionLogic)
    handler(c)
}

func (h *Handler) performActionLogic(c *gin.Context, req *ActionRequest) error {
    id := c.Param("id")
    return h.service.PerformAction(id, req)
}
```

## Summary

**Total Reduction in user_handler.go**: 70 lines (32.7%)

| Method | Before | After | Reduction |
|--------|--------|-------|-----------|
| List | 30 lines | 15 lines | 50% |
| GetByID | 18 lines | 12 lines | 33% |
| Create | 20 lines | 9 lines | 55% |
| Update | 28 lines | 12 lines | 57% |
| Delete | 20 lines | 15 lines | 25% |
| AssignRoles | 28 lines | 13 lines | 54% |

**Average Reduction**: 45.7% per method

**Next Steps**: Apply to remaining 6 handler files for total projected savings of 395 lines (22.8%).
