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

// AppError is the application error structure following OneX ErrorX pattern.
// It provides structured error handling with:
// - Error codes for HTTP status mapping
// - Business reason codes for logic routing
// - User-facing messages with formatting
// - Structured metadata for debugging
// - Request ID tracking for distributed tracing
//
// This is backward compatible with the previous AppError design while adding
// OneX capabilities for better observability and error handling.
type AppError struct {
	Code     ErrorCode         // HTTP status code (backward compatible)
	Reason   string            // Business reason code (NEW - OneX pattern)
	Message  string            // User-facing message
	Metadata map[string]string // Structured metadata (replaces Details interface{})
	Err      error             // Underlying error for wrapping

	// Deprecated: Use Metadata instead for structured data
	// Kept for backward compatibility during migration
	Details interface{}
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

// WithDetails adds detailed information.
// Deprecated: Use KV() or WithMetadata() for structured metadata.
// Kept for backward compatibility.
func (e *AppError) WithDetails(details interface{}) *AppError {
	e.Details = details
	return e
}

// ============ OneX ErrorX Chainable Methods ============

// WithReason sets the business reason code.
// The reason code is separate from HTTP status codes and used for business logic routing.
//
// Example:
//
//	err := errors.ErrValidationFailed.WithReason("InvalidEmail")
func (e *AppError) WithReason(reason string) *AppError {
	e.Reason = reason
	return e
}

// KV adds key-value pairs to the error metadata.
// This is the primary method for adding structured debugging context.
//
// Example:
//
//	err := errors.ErrNotFound.
//	    KV("resource", "agent", "agent_id", agentID, "cluster", clusterID)
func (e *AppError) KV(kvs ...string) *AppError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	for i := 0; i < len(kvs); i += 2 {
		if i+1 < len(kvs) {
			e.Metadata[kvs[i]] = kvs[i+1]
		}
	}
	return e
}

// WithRequestID adds the request ID to metadata for distributed tracing.
// This is a convenience wrapper for KV("X-Request-ID", requestID).
//
// Example:
//
//	err := errors.ErrInternalError.WithRequestID(requestID)
func (e *AppError) WithRequestID(requestID string) *AppError {
	return e.KV("X-Request-ID", requestID)
}

// WithTraceID adds the trace ID to metadata for distributed tracing.
// This is a convenience wrapper for KV("X-Trace-ID", traceID).
//
// Example:
//
//	err := errors.ErrInternalError.
//	    WithRequestID(requestID).
//	    WithTraceID(traceID)
func (e *AppError) WithTraceID(traceID string) *AppError {
	return e.KV("X-Trace-ID", traceID)
}

// WithMetadata replaces the entire metadata map.
// Use this when you have a pre-built metadata map.
//
// Example:
//
//	metadata := map[string]string{"resource": "pod", "namespace": "default"}
//	err := errors.ErrNotFound.WithMetadata(metadata)
func (e *AppError) WithMetadata(md map[string]string) *AppError {
	e.Metadata = md
	return e
}

// WithMessage updates the error message with formatting.
// Use this to provide more specific context while keeping the error code.
//
// Example:
//
//	err := errors.ErrNotFound.
//	    WithMessage("Agent '%s' not found in cluster '%s'", agentID, clusterID)
func (e *AppError) WithMessage(format string, args ...interface{}) *AppError {
	e.Message = fmt.Sprintf(format, args...)
	return e
}

// GetCode extracts error code from any error.
func GetCode(err error) ErrorCode {
	if err == nil {
		return CodeSuccess
	}

	if e, ok := err.(*AppError); ok {
		return e.Code
	}

	return CodeUnknown
}

// ============ OneX ErrorX Error Conversion Methods ============

// FromError converts any error to AppError following OneX pattern.
// This is the universal error converter that handles:
// 1. nil errors → nil
// 2. Already AppError → return as-is
// 3. AppError in error chain → extract it
// 4. Other errors → wrap as CodeInternalError
//
// Example:
//
//	// Convert any error to AppError
//	appErr := errors.FromError(someError)
//	if appErr != nil {
//	    appErr.WithRequestID(requestID).KV("operation", "create_user")
//	}
func FromError(err error) *AppError {
	if err == nil {
		return nil
	}

	// Already an AppError?
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	// Try to extract AppError from error chain
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	// Fallback: wrap as internal error
	return &AppError{
		Code:    CodeInternalError,
		Message: err.Error(),
		Err:     err,
	}
}

// Is checks if an error matches the given AppError by Code and optionally Reason.
// This enables error matching beyond simple equality checks.
//
// Matching rules:
// - If target.Reason is empty, match by Code only
// - If target.Reason is set, match by both Code and Reason
//
// Example:
//
//	// Match by code only
//	if errors.Is(err, errors.ErrNotFound) {
//	    // Handle not found
//	}
//
//	// Match by code and reason
//	notFoundAgent := errors.ErrNotFound.WithReason("AgentNotFound")
//	if errors.Is(err, notFoundAgent) {
//	    // Handle agent-specific not found
//	}
func Is(err error, target *AppError) bool {
	if err == nil || target == nil {
		return false
	}

	appErr := FromError(err)
	if appErr == nil {
		return false
	}

	// Match by code
	if appErr.Code != target.Code {
		return false
	}

	// If target has a reason, also match by reason
	if target.Reason != "" && appErr.Reason != target.Reason {
		return false
	}

	return true
}

// IsCode 检查错误码是否匹配
func IsCode(err error, code ErrorCode) bool {
	return GetCode(err) == code
}

// Predefined errors - HTTP related
var (
	ErrBadRequest         = New(CodeBadRequest, "Invalid request parameters")
	ErrUnauthorized       = New(CodeUnauthorized, "Unauthorized")
	ErrForbidden          = New(CodeForbidden, "Forbidden")
	ErrNotFound           = New(CodeNotFound, "Resource not found")
	ErrConflict           = New(CodeConflict, "Resource conflict")
	ErrValidationFailed   = New(CodeValidationFailed, "Validation failed")
	ErrInternalError      = New(CodeInternalError, "Internal server error")
	ErrServiceUnavailable = New(CodeServiceUnavailable, "Service unavailable")
	ErrTimeout            = New(CodeTimeout, "Request timeout")
)

// Predefined errors - General
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

// NewK8sAPIError creates a Kubernetes API error.
func NewK8sAPIError(err error) *AppError {
	return Wrap(CodeK8sAPIError, "Kubernetes API error", err)
}

// ============ OneX ErrorX HTTP Response Integration ============

// HTTPStatus returns the HTTP status code for this error.
// This maps ErrorCode to standard HTTP status codes.
func (e *AppError) HTTPStatus() int {
	// ErrorCode already uses HTTP status codes (400-599)
	// For custom codes (1000+), map to appropriate HTTP status
	if e.Code >= 400 && e.Code < 600 {
		return int(e.Code)
	}

	// Map custom codes to HTTP status
	switch e.Code {
	case CodeSuccess:
		return 200
	case CodeInvalidParam, CodeValidationFailed:
		return 400
	case CodeAlreadyExists, CodeResourceExists:
		return 409
	case CodeNotFound, CodeResourceNotFound, CodeClusterNotFound,
		CodeNamespaceNotFound, CodePodNotFound, CodeDeploymentNotFound,
		CodeServiceNotFound, CodeAgentNotFound, CodeWorkflowNotFound:
		return 404
	default:
		return 500 // Internal server error for unknown codes
	}
}

// ToMap converts AppError to a map for JSON serialization.
// This is useful for API responses following OneX pattern.
//
// Example:
//
//	err := errors.ErrNotFound.
//	    WithMessage("Agent not found").
//	    WithRequestID(requestID).
//	    KV("agent_id", agentID)
//	c.JSON(err.HTTPStatus(), err.ToMap())
func (e *AppError) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"code":    int(e.Code),
		"message": e.Message,
	}

	// Add reason if set
	if e.Reason != "" {
		result["reason"] = e.Reason
	}

	// Add metadata if present
	if len(e.Metadata) > 0 {
		result["metadata"] = e.Metadata
	}

	// Add request ID if present in metadata
	if requestID, ok := e.Metadata["X-Request-ID"]; ok {
		result["request_id"] = requestID
	}

	// Add trace ID if present in metadata
	if traceID, ok := e.Metadata["X-Trace-ID"]; ok {
		result["trace_id"] = traceID
	}

	return result
}
