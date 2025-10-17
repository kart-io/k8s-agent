package errors

import (
	"errors"
	"fmt"
)

// 错误码常量
const (
	// 成功
	CodeSuccess = 0

	// 客户端错误 (4xx)
	CodeBadRequest          = 400
	CodeUnauthorized        = 401
	CodeForbidden           = 403
	CodeNotFound            = 404
	CodeConflict            = 409
	CodeValidationFailed    = 422
	CodeTooManyRequests     = 429

	// 服务器错误 (5xx)
	CodeInternalError       = 500
	CodeNotImplemented      = 501
	CodeServiceUnavailable  = 503
	CodeTimeout             = 504

	// 业务错误 (1000+)
	CodeClusterNotFound     = 1001
	CodeClusterUnreachable  = 1002
	CodeResourceNotFound    = 1003
	CodeResourceExists      = 1004
	CodeInvalidKubeconfig   = 1005
	CodeK8sAPIError         = 1006
	CodeNamespaceNotFound   = 1007
	CodePodNotFound         = 1008
	CodeDeploymentNotFound  = 1009
	CodeServiceNotFound     = 1010
)

// AppError 应用错误结构
type AppError struct {
	Code    int
	Message string
	Err     error
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 支持 errors.Unwrap
func (e *AppError) Unwrap() error {
	return e.Err
}

// New 创建新的应用错误
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap 包装已有错误
func Wrap(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// 预定义错误
var (
	ErrBadRequest          = New(CodeBadRequest, "Invalid request parameters")
	ErrUnauthorized        = New(CodeUnauthorized, "Unauthorized")
	ErrForbidden           = New(CodeForbidden, "Forbidden")
	ErrNotFound            = New(CodeNotFound, "Resource not found")
	ErrConflict            = New(CodeConflict, "Resource conflict")
	ErrInternalError       = New(CodeInternalError, "Internal server error")
	ErrServiceUnavailable  = New(CodeServiceUnavailable, "Service unavailable")
	ErrClusterNotFound     = New(CodeClusterNotFound, "Cluster not found")
	ErrClusterUnreachable  = New(CodeClusterUnreachable, "Cluster unreachable")
	ErrResourceNotFound    = New(CodeResourceNotFound, "Kubernetes resource not found")
	ErrResourceExists      = New(CodeResourceExists, "Kubernetes resource already exists")
	ErrInvalidKubeconfig   = New(CodeInvalidKubeconfig, "Invalid kubeconfig")
	ErrK8sAPIError         = New(CodeK8sAPIError, "Kubernetes API error")
	ErrNamespaceNotFound   = New(CodeNamespaceNotFound, "Namespace not found")
	ErrPodNotFound         = New(CodePodNotFound, "Pod not found")
	ErrDeploymentNotFound  = New(CodeDeploymentNotFound, "Deployment not found")
	ErrServiceNotFound     = New(CodeServiceNotFound, "Service not found")
)

// IsAppError 判断是否为应用错误
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError 获取应用错误
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

// IsNotFound 判断是否为资源不存在错误
func IsNotFound(err error) bool {
	appErr := GetAppError(err)
	if appErr != nil {
		return appErr.Code == CodeNotFound ||
			appErr.Code == CodeResourceNotFound ||
			appErr.Code == CodeClusterNotFound ||
			appErr.Code == CodeNamespaceNotFound ||
			appErr.Code == CodePodNotFound ||
			appErr.Code == CodeDeploymentNotFound ||
			appErr.Code == CodeServiceNotFound
	}
	return false
}

// IsConflict 判断是否为资源冲突错误
func IsConflict(err error) bool {
	appErr := GetAppError(err)
	if appErr != nil {
		return appErr.Code == CodeConflict || appErr.Code == CodeResourceExists
	}
	return false
}

// NewValidationError 创建验证错误
func NewValidationError(err error) *AppError {
	return Wrap(CodeValidationFailed, "Validation failed", err)
}

// NewDatabaseError 创建数据库错误
func NewDatabaseError(err error) *AppError {
	return Wrap(CodeInternalError, "Database error", err)
}

// NewK8sAPIError 创建 Kubernetes API 错误
func NewK8sAPIError(err error) *AppError {
	return Wrap(CodeK8sAPIError, "Kubernetes API error", err)
}

