package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func parseResponse(t *testing.T, w *httptest.ResponseRecorder) APIResponse {
	var resp APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "Failed to parse response body")
	return resp
}

// TestSuccess tests the Success function
func TestSuccess(t *testing.T) {
	c, w := setupTestContext()
	testData := map[string]string{"key": "value"}

	Success(c, testData)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.NotNil(t, resp.Data)
	assert.Empty(t, resp.Error)
}

// TestSuccessWithNilData tests Success with nil data
func TestSuccessWithNilData(t *testing.T) {
	c, w := setupTestContext()

	Success(c, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
}

// TestSuccessWithMessage tests the SuccessWithMessage function
func TestSuccessWithMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		data    interface{}
	}{
		{
			name:    "custom message with data",
			message: "operation completed",
			data:    map[string]string{"status": "done"},
		},
		{
			name:    "custom message with nil data",
			message: "deleted successfully",
			data:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupTestContext()

			SuccessWithMessage(c, tt.message, tt.data)

			assert.Equal(t, http.StatusOK, w.Code)
			resp := parseResponse(t, w)
			assert.Equal(t, 0, resp.Code)
			assert.Equal(t, tt.message, resp.Message)
		})
	}
}

// TestError tests the Error function
func TestError(t *testing.T) {
	tests := []struct {
		name       string
		httpStatus int
		code       int
		message    string
		err        error
		wantError  string
	}{
		{
			name:       "error with error object",
			httpStatus: http.StatusBadRequest,
			code:       400,
			message:    "invalid input",
			err:        errors.New("validation failed"),
			wantError:  "validation failed",
		},
		{
			name:       "error without error object",
			httpStatus: http.StatusInternalServerError,
			code:       500,
			message:    "server error",
			err:        nil,
			wantError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupTestContext()

			Error(c, tt.httpStatus, tt.code, tt.message, tt.err)

			assert.Equal(t, tt.httpStatus, w.Code)
			resp := parseResponse(t, w)
			assert.Equal(t, tt.code, resp.Code)
			assert.Equal(t, tt.message, resp.Message)
			assert.Equal(t, tt.wantError, resp.Error)
		})
	}
}

// TestBadRequest tests the BadRequest function
func TestBadRequest(t *testing.T) {
	c, w := setupTestContext()
	err := errors.New("invalid parameter")

	BadRequest(c, "bad request", err)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "bad request", resp.Message)
	assert.Equal(t, "invalid parameter", resp.Error)
}

// TestUnauthorized tests the Unauthorized function
func TestUnauthorized(t *testing.T) {
	c, w := setupTestContext()
	err := errors.New("token expired")

	Unauthorized(c, "unauthorized access", err)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "unauthorized access", resp.Message)
	assert.Equal(t, "token expired", resp.Error)
}

// TestForbidden tests the Forbidden function
func TestForbidden(t *testing.T) {
	c, w := setupTestContext()
	err := errors.New("insufficient permissions")

	Forbidden(c, "access forbidden", err)

	assert.Equal(t, http.StatusForbidden, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, 403, resp.Code)
	assert.Equal(t, "access forbidden", resp.Message)
	assert.Equal(t, "insufficient permissions", resp.Error)
}

// TestNotFound tests the NotFound function
func TestNotFound(t *testing.T) {
	c, w := setupTestContext()
	err := errors.New("resource not found")

	NotFound(c, "not found", err)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, 404, resp.Code)
	assert.Equal(t, "not found", resp.Message)
	assert.Equal(t, "resource not found", resp.Error)
}

// TestConflict tests the Conflict function
func TestConflict(t *testing.T) {
	c, w := setupTestContext()
	err := errors.New("duplicate entry")

	Conflict(c, "conflict detected", err)

	assert.Equal(t, http.StatusConflict, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, 409, resp.Code)
	assert.Equal(t, "conflict detected", resp.Message)
	assert.Equal(t, "duplicate entry", resp.Error)
}

// TestInternalError tests the InternalError function
func TestInternalError(t *testing.T) {
	c, w := setupTestContext()
	err := errors.New("database connection failed")

	InternalError(c, "internal server error", err)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, 500, resp.Code)
	assert.Equal(t, "internal server error", resp.Message)
	assert.Equal(t, "database connection failed", resp.Error)
}

// TestServiceUnavailable tests the ServiceUnavailable function
func TestServiceUnavailable(t *testing.T) {
	c, w := setupTestContext()
	err := errors.New("service overloaded")

	ServiceUnavailable(c, "service unavailable", err)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, 503, resp.Code)
	assert.Equal(t, "service unavailable", resp.Message)
	assert.Equal(t, "service overloaded", resp.Error)
}

// TestValidationError tests the ValidationError function
func TestValidationError(t *testing.T) {
	c, w := setupTestContext()

	ValidationError(c, "validation failed")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, CodeValidationError, resp.Code)
	assert.Equal(t, "validation failed", resp.Message)
	assert.Empty(t, resp.Error)
}

// TestAuthenticationError tests the AuthenticationError function
func TestAuthenticationError(t *testing.T) {
	c, w := setupTestContext()

	AuthenticationError(c, "authentication failed")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, CodeAuthenticationErr, resp.Code)
	assert.Equal(t, "authentication failed", resp.Message)
	assert.Empty(t, resp.Error)
}

// TestPermissionDenied tests the PermissionDenied function
func TestPermissionDenied(t *testing.T) {
	c, w := setupTestContext()

	PermissionDenied(c, "permission denied")

	assert.Equal(t, http.StatusForbidden, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, CodePermissionDenied, resp.Code)
	assert.Equal(t, "permission denied", resp.Message)
	assert.Empty(t, resp.Error)
}

// TestDatabaseError tests the DatabaseError function
func TestDatabaseError(t *testing.T) {
	c, w := setupTestContext()

	DatabaseError(c, "database error")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, CodeDatabaseError, resp.Code)
	assert.Equal(t, "database error", resp.Message)
	assert.Empty(t, resp.Error)
}

// TestCreated tests the Created function
func TestCreated(t *testing.T) {
	c, w := setupTestContext()
	testData := map[string]string{"id": "123", "name": "test"}

	Created(c, testData)

	assert.Equal(t, http.StatusCreated, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, CodeSuccess, resp.Code)
	assert.Equal(t, "created", resp.Message)
	assert.NotNil(t, resp.Data)
}

// TestSuccessList tests the SuccessList function
func TestSuccessList(t *testing.T) {
	tests := []struct {
		name  string
		items interface{}
		total int64
	}{
		{
			name:  "list with items",
			items: []string{"item1", "item2", "item3"},
			total: 3,
		},
		{
			name:  "empty list",
			items: []string{},
			total: 0,
		},
		{
			name:  "large count",
			items: []int{1, 2, 3},
			total: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupTestContext()

			SuccessList(c, tt.items, tt.total)

			assert.Equal(t, http.StatusOK, w.Code)
			resp := parseResponse(t, w)
			assert.Equal(t, 0, resp.Code)
			assert.Equal(t, "success", resp.Message)

			// Parse the Data field as ListResponse
			dataBytes, err := json.Marshal(resp.Data)
			require.NoError(t, err)
			var listResp ListResponse
			err = json.Unmarshal(dataBytes, &listResp)
			require.NoError(t, err)
			assert.Equal(t, tt.total, listResp.Total)
		})
	}
}

// TestPaginated tests the Paginated function
func TestPaginated(t *testing.T) {
	tests := []struct {
		name           string
		data           interface{}
		total          int64
		page           int
		pageSize       int
		expectedPages  int
	}{
		{
			name:          "first page",
			data:          []int{1, 2, 3, 4, 5},
			total:         50,
			page:          1,
			pageSize:      10,
			expectedPages: 5,
		},
		{
			name:          "last page with remainder",
			data:          []int{1, 2, 3},
			total:         23,
			page:          3,
			pageSize:      10,
			expectedPages: 3,
		},
		{
			name:          "exact division",
			data:          []int{1, 2, 3, 4, 5},
			total:         100,
			page:          2,
			pageSize:      20,
			expectedPages: 5,
		},
		{
			name:          "single item per page",
			data:          []int{1},
			total:         5,
			page:          3,
			pageSize:      1,
			expectedPages: 5,
		},
		{
			name:          "zero total",
			data:          []int{},
			total:         0,
			page:          1,
			pageSize:      10,
			expectedPages: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupTestContext()

			Paginated(c, tt.data, tt.total, tt.page, tt.pageSize)

			assert.Equal(t, http.StatusOK, w.Code)
			resp := parseResponse(t, w)
			assert.Equal(t, CodeSuccess, resp.Code)
			assert.Equal(t, "success", resp.Message)

			// Parse the Data field as PaginatedResponse
			dataBytes, err := json.Marshal(resp.Data)
			require.NoError(t, err)
			var pageResp PaginatedResponse
			err = json.Unmarshal(dataBytes, &pageResp)
			require.NoError(t, err)

			assert.Equal(t, tt.total, pageResp.Total)
			assert.Equal(t, tt.page, pageResp.Page)
			assert.Equal(t, tt.pageSize, pageResp.PageSize)
			assert.Equal(t, tt.expectedPages, pageResp.TotalPages)
		})
	}
}

// TestErrorCodeConstants tests that error code constants have expected values
func TestErrorCodeConstants(t *testing.T) {
	assert.Equal(t, 0, CodeSuccess)
	assert.Equal(t, 400, CodeBadRequest)
	assert.Equal(t, 401, CodeUnauthorized)
	assert.Equal(t, 403, CodeForbidden)
	assert.Equal(t, 404, CodeNotFound)
	assert.Equal(t, 409, CodeConflict)
	assert.Equal(t, 500, CodeInternalError)
	assert.Equal(t, 5001, CodeDatabaseError)
	assert.Equal(t, 4001, CodeValidationError)
	assert.Equal(t, 4011, CodeAuthenticationErr)
	assert.Equal(t, 4031, CodePermissionDenied)
}

// TestAPIResponseJSONSerialization tests JSON serialization of APIResponse
func TestAPIResponseJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		response APIResponse
		wantJSON string
	}{
		{
			name: "success response",
			response: APIResponse{
				Code:    0,
				Message: "success",
				Data:    map[string]string{"key": "value"},
			},
			wantJSON: `{"code":0,"message":"success","data":{"key":"value"}}`,
		},
		{
			name: "error response",
			response: APIResponse{
				Code:    400,
				Message: "bad request",
				Error:   "validation failed",
			},
			wantJSON: `{"code":400,"message":"bad request","error":"validation failed"}`,
		},
		{
			name: "response with nil data",
			response: APIResponse{
				Code:    0,
				Message: "success",
				Data:    nil,
			},
			wantJSON: `{"code":0,"message":"success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.response)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(jsonData))
		})
	}
}

// TestListResponseStructure tests the ListResponse structure
func TestListResponseStructure(t *testing.T) {
	items := []string{"item1", "item2"}
	listResp := ListResponse{
		Items: items,
		Total: 2,
	}

	assert.Equal(t, items, listResp.Items)
	assert.Equal(t, int64(2), listResp.Total)
}

// TestPaginatedResponseStructure tests the PaginatedResponse structure
func TestPaginatedResponseStructure(t *testing.T) {
	items := []int{1, 2, 3}
	pageResp := PaginatedResponse{
		Items:      items,
		Total:      100,
		Page:       2,
		PageSize:   10,
		TotalPages: 10,
	}

	assert.Equal(t, items, pageResp.Items)
	assert.Equal(t, int64(100), pageResp.Total)
	assert.Equal(t, 2, pageResp.Page)
	assert.Equal(t, 10, pageResp.PageSize)
	assert.Equal(t, 10, pageResp.TotalPages)
}
