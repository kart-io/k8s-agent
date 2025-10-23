package errors

import (
	"errors"
	"fmt"
)

// ErrorCode 错误码类型
type ErrorCode int

const (
	// 成功
	CodeSuccess ErrorCode = 0

	// HTTP 客户端错误 (400-499)
	CodeBadRequest       ErrorCode = 400
	CodeUnauthorized     ErrorCode = 401
	CodeForbidden        ErrorCode = 403
	CodeNotFound         ErrorCode = 404
	CodeConflict         ErrorCode = 409
	CodeValidationFailed ErrorCode = 422
	CodeTooManyRequests  ErrorCode = 429

	// HTTP 服务器错误 (500-599)
	CodeInternalError      ErrorCode = 500
	CodeNotImplemented     ErrorCode = 501
	CodeServiceUnavailable ErrorCode = 503
	CodeTimeout            ErrorCode = 504

	// 通用错误 (1000-1999)
	CodeUnknown         ErrorCode = 1000
	CodeInvalidParam    ErrorCode = 1002
	CodeAlreadyExists   ErrorCode = 1004
	CodeOperationFailed ErrorCode = 1007

	// K8s相关错误 (1100-1199)
	CodeClusterNotFound    ErrorCode = 1101
	CodeClusterUnreachable ErrorCode = 1102
	CodeResourceNotFound   ErrorCode = 1103
	CodeResourceExists     ErrorCode = 1104
	CodeInvalidKubeconfig  ErrorCode = 1105
	CodeK8sAPIError        ErrorCode = 1106
	CodeNamespaceNotFound  ErrorCode = 1107
	CodePodNotFound        ErrorCode = 1108
	CodeDeploymentNotFound ErrorCode = 1109
	CodeServiceNotFound    ErrorCode = 1110

	// Agent Manager 错误 (2000-2999)
	CodeAgentNotFound ErrorCode = 2001
	CodeAgentOffline  ErrorCode = 2002
	CodeCommandFailed ErrorCode = 2004

	// Orchestrator 错误 (3000-3999)
	CodeWorkflowNotFound  ErrorCode = 3001
	CodeWorkflowFailed    ErrorCode = 3002
	CodeStrategyNotFound  ErrorCode = 3003
	CodeExecutionFailed   ErrorCode = 3004
	CodeWorkflowCancelled ErrorCode = 3005

	// Reasoning 错误 (4000-4999)
	CodeAnalysisFailed    ErrorCode = 4001
	CodeLLMError          ErrorCode = 4002
	CodeKnowledgeNotFound ErrorCode = 4003
	CodeInvalidEvidence   ErrorCode = 4004
)

// AppError 应用错误结构
// 兼容旧的 AppError 和新的 Error 设计
type AppError struct {
	Code    ErrorCode
	Message string
	Details interface{} // 详细错误信息
	Err     error       // 底层错误
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
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Newf 创建格式化的新错误
func Newf(code ErrorCode, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap 包装已有错误
func Wrap(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Wrapf 包装已有错误,使用格式化消息
func Wrapf(code ErrorCode, err error, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
	}
}

// WithDetails 添加详细信息
func (e *AppError) WithDetails(details interface{}) *AppError {
	e.Details = details
	return e
}

// GetCode 从 error 中提取错误码
func GetCode(err error) ErrorCode {
	if err == nil {
		return CodeSuccess
	}

	if e, ok := err.(*AppError); ok {
		return e.Code
	}

	return CodeUnknown
}

// IsCode 检查错误码是否匹配
func IsCode(err error, code ErrorCode) bool {
	return GetCode(err) == code
}

// 预定义错误 - HTTP 相关
var (
	ErrBadRequest         = New(CodeBadRequest, "Invalid request parameters")
	ErrUnauthorized       = New(CodeUnauthorized, "Unauthorized")
	ErrForbidden          = New(CodeForbidden, "Forbidden")
	ErrNotFound           = New(CodeNotFound, "Resource not found")
	ErrConflict           = New(CodeConflict, "Resource conflict")
	ErrInternalError      = New(CodeInternalError, "Internal server error")
	ErrServiceUnavailable = New(CodeServiceUnavailable, "Service unavailable")
	ErrTimeout            = New(CodeTimeout, "Request timeout")
)

// 预定义错误 - 通用
var (
	ErrInvalidParam  = New(CodeInvalidParam, "Invalid parameter")
	ErrAlreadyExists = New(CodeAlreadyExists, "Resource already exists")
)

// 预定义错误 - K8s 相关
var (
	ErrClusterNotFound    = New(CodeClusterNotFound, "Cluster not found")
	ErrClusterUnreachable = New(CodeClusterUnreachable, "Cluster unreachable")
	ErrResourceNotFound   = New(CodeResourceNotFound, "Kubernetes resource not found")
	ErrResourceExists     = New(CodeResourceExists, "Kubernetes resource already exists")
	ErrInvalidKubeconfig  = New(CodeInvalidKubeconfig, "Invalid kubeconfig")
	ErrK8sAPIError        = New(CodeK8sAPIError, "Kubernetes API error")
	ErrNamespaceNotFound  = New(CodeNamespaceNotFound, "Namespace not found")
	ErrPodNotFound        = New(CodePodNotFound, "Pod not found")
	ErrDeploymentNotFound = New(CodeDeploymentNotFound, "Deployment not found")
	ErrServiceNotFound    = New(CodeServiceNotFound, "Service not found")
)

// 预定义错误 - Agent Manager
var (
	ErrAgentNotFound = New(CodeAgentNotFound, "Agent not found")
	ErrAgentOffline  = New(CodeAgentOffline, "Agent offline")
)

// 预定义错误 - Orchestrator
var (
	ErrWorkflowNotFound = New(CodeWorkflowNotFound, "Workflow not found")
	ErrAnalysisFailed   = New(CodeAnalysisFailed, "Analysis failed")
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
			appErr.Code == CodeServiceNotFound ||
			appErr.Code == CodeAgentNotFound ||
			appErr.Code == CodeWorkflowNotFound
	}
	return false
}

// IsConflict 判断是否为资源冲突错误
func IsConflict(err error) bool {
	appErr := GetAppError(err)
	if appErr != nil {
		return appErr.Code == CodeConflict ||
			appErr.Code == CodeResourceExists ||
			appErr.Code == CodeAlreadyExists
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
