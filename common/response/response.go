package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse 统一的 API 响应格式
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 返回成功响应（自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// Error 返回错误响应
func Error(c *gin.Context, httpStatus int, code int, message string, err error) {
	response := APIResponse{
		Code:    code,
		Message: message,
	}
	if err != nil {
		response.Error = err.Error()
	}
	c.JSON(httpStatus, response)
}

// BadRequest 返回 400 错误
func BadRequest(c *gin.Context, message string, err error) {
	Error(c, http.StatusBadRequest, 400, message, err)
}

// Unauthorized 返回 401 错误
func Unauthorized(c *gin.Context, message string, err error) {
	Error(c, http.StatusUnauthorized, 401, message, err)
}

// Forbidden 返回 403 错误
func Forbidden(c *gin.Context, message string, err error) {
	Error(c, http.StatusForbidden, 403, message, err)
}

// NotFound 返回 404 错误
func NotFound(c *gin.Context, message string, err error) {
	Error(c, http.StatusNotFound, 404, message, err)
}

// Conflict 返回 409 错误
func Conflict(c *gin.Context, message string, err error) {
	Error(c, http.StatusConflict, 409, message, err)
}

// InternalError 返回 500 错误
func InternalError(c *gin.Context, message string, err error) {
	Error(c, http.StatusInternalServerError, 500, message, err)
}

// ServiceUnavailable 返回 503 错误
func ServiceUnavailable(c *gin.Context, message string, err error) {
	Error(c, http.StatusServiceUnavailable, 503, message, err)
}

// Auth service specific error codes
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

// ValidationError sends a 400 validation error
func ValidationError(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, CodeValidationError, message, nil)
}

// AuthenticationError sends a 401 authentication error
func AuthenticationError(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, CodeAuthenticationErr, message, nil)
}

// PermissionDenied sends a 403 permission denied error
func PermissionDenied(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, CodePermissionDenied, message, nil)
}

// DatabaseError sends a 500 database error
func DatabaseError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, CodeDatabaseError, message, nil)
}

// Created sends a created response (201)
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Code:    CodeSuccess,
		Message: "created",
		Data:    data,
	})
}

// ListResponse 列表响应结构
type ListResponse struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
}

// SuccessList 返回列表成功响应
func SuccessList(c *gin.Context, items interface{}, total int64) {
	Success(c, ListResponse{
		Items: items,
		Total: total,
	})
}

// PaginatedResponse represents paginated response
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// Paginated sends a paginated response
func Paginated(c *gin.Context, data interface{}, total int64, page int, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    CodeSuccess,
		Message: "success",
		Data: PaginatedResponse{
			Items:      data,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}
