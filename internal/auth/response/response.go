package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/internal/auth/types"
)

// Error codes.
const (
	CodeSuccess           = 0
	CodeBadRequest        = 400
	CodeUnauthorized      = 401
	CodeForbidden         = 403
	CodeNotFound          = 404
	CodeConflict          = 409
	CodeInternalError     = 500
	CodeDatabaseError     = 5001
	CodeValidationError   = 4001
	CodeAuthenticationErr = 4011
	CodePermissionDenied  = 4031
)

// Success sends a successful response.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    CodeSuccess,
		"message": "success",
		"data":    data,
	})
}

// Created sends a created response (201).
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{
		"code":    CodeSuccess,
		"message": "created",
		"data":    data,
	})
}

// Error sends an error response.
func Error(c *gin.Context, httpStatus int, code int, message string, details string) {
	detailsMap := make(map[string]interface{})
	if details != "" {
		detailsMap["description"] = details
	}
	detailsMap["code"] = code

	c.JSON(httpStatus, types.ErrorResponse{
		Error:   message,
		Message: message,
		Details: detailsMap,
	})
}

// BadRequest sends a 400 bad request error.
func BadRequest(c *gin.Context, details string) {
	Error(c, http.StatusBadRequest, CodeBadRequest, "Bad Request", details)
}

// ValidationError sends a 400 validation error.
func ValidationError(c *gin.Context, details string) {
	Error(c, http.StatusBadRequest, CodeValidationError, "Validation Error", details)
}

// Unauthorized sends a 401 unauthorized error.
func Unauthorized(c *gin.Context, details string) {
	Error(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthorized", details)
}

// AuthenticationError sends a 401 authentication error.
func AuthenticationError(c *gin.Context, details string) {
	Error(c, http.StatusUnauthorized, CodeAuthenticationErr, "Authentication Failed", details)
}

// Forbidden sends a 403 forbidden error.
func Forbidden(c *gin.Context, details string) {
	Error(c, http.StatusForbidden, CodeForbidden, "Forbidden", details)
}

// PermissionDenied sends a 403 permission denied error.
func PermissionDenied(c *gin.Context, details string) {
	Error(c, http.StatusForbidden, CodePermissionDenied, "Permission Denied", details)
}

// NotFound sends a 404 not found error.
func NotFound(c *gin.Context, details string) {
	Error(c, http.StatusNotFound, CodeNotFound, "Not Found", details)
}

// Conflict sends a 409 conflict error.
func Conflict(c *gin.Context, details string) {
	Error(c, http.StatusConflict, CodeConflict, "Conflict", details)
}

// InternalError sends a 500 internal server error.
func InternalError(c *gin.Context, details string) {
	Error(c, http.StatusInternalServerError, CodeInternalError, "Internal Server Error", details)
}

// DatabaseError sends a 500 database error.
func DatabaseError(c *gin.Context, details string) {
	Error(c, http.StatusInternalServerError, CodeDatabaseError, "Database Error", details)
}

// Paginated sends a paginated response.
func Paginated(c *gin.Context, data interface{}, total int64, page int, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    CodeSuccess,
		"message": "success",
		"data": types.PaginatedResponse{
			Items:      data,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}
