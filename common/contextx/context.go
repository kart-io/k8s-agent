// Package contextx provides enhanced context handling utilities for request tracing,
// user context propagation, and timeout management.
//
// This package uses unexported struct types as context keys (following OneX best practices)
// to ensure type safety and prevent key collisions. Each context value has a dedicated
// type-safe getter and setter.
package contextx

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Context key types using unexported structs for type safety.
// This prevents key collisions and provides compile-time type checking.
// Following OneX pattern: use unexported struct{} as keys instead of string/int.
type (
	requestIDKey  struct{}
	userIDKey     struct{}
	usernameKey   struct{}
	traceIDKey    struct{}
	spanIDKey     struct{}
	tenantIDKey   struct{}
	clientIPKey   struct{}
	userAgentKey  struct{}
	realIPKey     struct{}
	sessionIDKey  struct{}
	agentIDKey    struct{}
	clusterIDKey  struct{}
	workflowIDKey struct{}
	taskIDKey     struct{}
	eventIDKey    struct{}
	commandIDKey  struct{}
)

// WithRequestID adds request ID to context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// GetRequestID retrieves request ID from context.
func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(requestIDKey{}); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// GetOrCreateRequestID gets existing request ID or creates a new one.
func GetOrCreateRequestID(ctx context.Context) (context.Context, string) {
	requestID := GetRequestID(ctx)
	if requestID == "" {
		requestID = uuid.New().String()
		ctx = WithRequestID(ctx, requestID)
	}
	return ctx, requestID
}

// WithUserID adds user ID to context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// GetUserID retrieves user ID from context.
func GetUserID(ctx context.Context) string {
	if v := ctx.Value(userIDKey{}); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// WithUsername adds username to context.
func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameKey{}, username)
}

// GetUsername retrieves username from context.
func GetUsername(ctx context.Context) string {
	if v := ctx.Value(usernameKey{}); v != nil {
		if name, ok := v.(string); ok {
			return name
		}
	}
	return ""
}

// WithTraceID adds trace ID to context.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

// GetTraceID retrieves trace ID from context.
func GetTraceID(ctx context.Context) string {
	if v := ctx.Value(traceIDKey{}); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// GetOrCreateTraceID gets existing trace ID or creates a new one.
func GetOrCreateTraceID(ctx context.Context) (context.Context, string) {
	traceID := GetTraceID(ctx)
	if traceID == "" {
		traceID = uuid.New().String()
		ctx = WithTraceID(ctx, traceID)
	}
	return ctx, traceID
}

// WithSpanID adds span ID to context.
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, spanIDKey{}, spanID)
}

// GetSpanID retrieves span ID from context.
func GetSpanID(ctx context.Context) string {
	if v := ctx.Value(spanIDKey{}); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// WithTenantID adds tenant ID to context.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey{}, tenantID)
}

// GetTenantID retrieves tenant ID from context.
func GetTenantID(ctx context.Context) string {
	if v := ctx.Value(tenantIDKey{}); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// WithClientIP adds client IP to context.
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey{}, ip)
}

// GetClientIP retrieves client IP from context.
func GetClientIP(ctx context.Context) string {
	if v := ctx.Value(clientIPKey{}); v != nil {
		if ip, ok := v.(string); ok {
			return ip
		}
	}
	return ""
}

// WithUserAgent adds user agent to context.
func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, userAgentKey{}, userAgent)
}

// GetUserAgent retrieves user agent from context.
func GetUserAgent(ctx context.Context) string {
	if v := ctx.Value(userAgentKey{}); v != nil {
		if ua, ok := v.(string); ok {
			return ua
		}
	}
	return ""
}

// WithRealIP adds real IP to context.
func WithRealIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, realIPKey{}, ip)
}

// GetRealIP retrieves real IP from context.
func GetRealIP(ctx context.Context) string {
	if v := ctx.Value(realIPKey{}); v != nil {
		if ip, ok := v.(string); ok {
			return ip
		}
	}
	return ""
}

// WithSessionID adds session ID to context.
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey{}, sessionID)
}

// GetSessionID retrieves session ID from context.
func GetSessionID(ctx context.Context) string {
	if v := ctx.Value(sessionIDKey{}); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// ContextInfo contains extracted context information.
type ContextInfo struct {
	RequestID string
	UserID    string
	Username  string
	TraceID   string
	SpanID    string
	TenantID  string
	ClientIP  string
	UserAgent string
	RealIP    string
	SessionID string
}

// ExtractInfo extracts all context information.
func ExtractInfo(ctx context.Context) *ContextInfo {
	return &ContextInfo{
		RequestID: GetRequestID(ctx),
		UserID:    GetUserID(ctx),
		Username:  GetUsername(ctx),
		TraceID:   GetTraceID(ctx),
		SpanID:    GetSpanID(ctx),
		TenantID:  GetTenantID(ctx),
		ClientIP:  GetClientIP(ctx),
		UserAgent: GetUserAgent(ctx),
		RealIP:    GetRealIP(ctx),
		SessionID: GetSessionID(ctx),
	}
}

// ApplyInfo applies context information to a new context.
func ApplyInfo(ctx context.Context, info *ContextInfo) context.Context {
	if info.RequestID != "" {
		ctx = WithRequestID(ctx, info.RequestID)
	}
	if info.UserID != "" {
		ctx = WithUserID(ctx, info.UserID)
	}
	if info.Username != "" {
		ctx = WithUsername(ctx, info.Username)
	}
	if info.TraceID != "" {
		ctx = WithTraceID(ctx, info.TraceID)
	}
	if info.SpanID != "" {
		ctx = WithSpanID(ctx, info.SpanID)
	}
	if info.TenantID != "" {
		ctx = WithTenantID(ctx, info.TenantID)
	}
	if info.ClientIP != "" {
		ctx = WithClientIP(ctx, info.ClientIP)
	}
	if info.UserAgent != "" {
		ctx = WithUserAgent(ctx, info.UserAgent)
	}
	if info.RealIP != "" {
		ctx = WithRealIP(ctx, info.RealIP)
	}
	if info.SessionID != "" {
		ctx = WithSessionID(ctx, info.SessionID)
	}
	return ctx
}

// NewContext creates a new context with timeout and values from parent.
func NewContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	info := ExtractInfo(parent)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	ctx = ApplyInfo(ctx, info)
	return ctx, cancel
}

// CopyContext copies context values to a new background context.
func CopyContext(parent context.Context) context.Context {
	info := ExtractInfo(parent)
	return ApplyInfo(context.Background(), info)
}

// ============ k8s-agent Specific Context Functions ============

// WithAgentID adds agent ID to context.
func WithAgentID(ctx context.Context, agentID string) context.Context {
	return context.WithValue(ctx, agentIDKey{}, agentID)
}

// GetAgentID retrieves agent ID from context.
func GetAgentID(ctx context.Context) string {
	if v := ctx.Value(agentIDKey{}); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// WithClusterID adds cluster ID to context.
func WithClusterID(ctx context.Context, clusterID string) context.Context {
	return context.WithValue(ctx, clusterIDKey{}, clusterID)
}

// GetClusterID retrieves cluster ID from context.
func GetClusterID(ctx context.Context) string {
	if v := ctx.Value(clusterIDKey{}); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// WithWorkflowID adds workflow ID to context.
func WithWorkflowID(ctx context.Context, workflowID string) context.Context {
	return context.WithValue(ctx, workflowIDKey{}, workflowID)
}

// GetWorkflowID retrieves workflow ID from context.
func GetWorkflowID(ctx context.Context) string {
	if v := ctx.Value(workflowIDKey{}); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// WithTaskID adds task ID to context.
func WithTaskID(ctx context.Context, taskID string) context.Context {
	return context.WithValue(ctx, taskIDKey{}, taskID)
}

// GetTaskID retrieves task ID from context.
func GetTaskID(ctx context.Context) string {
	if v := ctx.Value(taskIDKey{}); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// WithEventID adds event ID to context.
func WithEventID(ctx context.Context, eventID string) context.Context {
	return context.WithValue(ctx, eventIDKey{}, eventID)
}

// GetEventID retrieves event ID from context.
func GetEventID(ctx context.Context) string {
	if v := ctx.Value(eventIDKey{}); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// WithCommandID adds command ID to context.
func WithCommandID(ctx context.Context, commandID string) context.Context {
	return context.WithValue(ctx, commandIDKey{}, commandID)
}

// GetCommandID retrieves command ID from context.
func GetCommandID(ctx context.Context) string {
	if v := ctx.Value(commandIDKey{}); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// ExtractK8sAgentInfo extracts k8s-agent specific information for logging.
// This is useful for structured logging with all relevant context.
func ExtractK8sAgentInfo(ctx context.Context) map[string]string {
	info := make(map[string]string)

	if v := GetRequestID(ctx); v != "" {
		info["request_id"] = v
	}
	if v := GetUserID(ctx); v != "" {
		info["user_id"] = v
	}
	if v := GetTraceID(ctx); v != "" {
		info["trace_id"] = v
	}
	if v := GetAgentID(ctx); v != "" {
		info["agent_id"] = v
	}
	if v := GetClusterID(ctx); v != "" {
		info["cluster_id"] = v
	}
	if v := GetWorkflowID(ctx); v != "" {
		info["workflow_id"] = v
	}
	if v := GetTaskID(ctx); v != "" {
		info["task_id"] = v
	}
	if v := GetEventID(ctx); v != "" {
		info["event_id"] = v
	}
	if v := GetCommandID(ctx); v != "" {
		info["command_id"] = v
	}

	return info
}
