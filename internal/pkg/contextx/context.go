// Package contextx provides enhanced context handling utilities for request tracing,
// user context propagation, and timeout management.
package contextx

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Context key types to prevent collisions
type contextKey string

const (
	// RequestID key for request ID
	RequestIDKey contextKey = "request_id"

	// UserID key for user ID
	UserIDKey contextKey = "user_id"

	// Username key for username
	UsernameKey contextKey = "username"

	// TraceID key for distributed tracing
	TraceIDKey contextKey = "trace_id"

	// SpanID key for distributed tracing span
	SpanIDKey contextKey = "span_id"

	// TenantID key for multi-tenancy
	TenantIDKey contextKey = "tenant_id"

	// ClientIP key for client IP address
	ClientIPKey contextKey = "client_ip"

	// UserAgent key for user agent
	UserAgentKey contextKey = "user_agent"

	// RealIP key for real IP address
	RealIPKey contextKey = "real_ip"

	// SessionID key for session ID
	SessionIDKey contextKey = "session_id"
)

// WithRequestID adds request ID to context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID retrieves request ID from context.
func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(RequestIDKey); v != nil {
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
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID retrieves user ID from context.
func GetUserID(ctx context.Context) string {
	if v := ctx.Value(UserIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// WithUsername adds username to context.
func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, UsernameKey, username)
}

// GetUsername retrieves username from context.
func GetUsername(ctx context.Context) string {
	if v := ctx.Value(UsernameKey); v != nil {
		if name, ok := v.(string); ok {
			return name
		}
	}
	return ""
}

// WithTraceID adds trace ID to context.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// GetTraceID retrieves trace ID from context.
func GetTraceID(ctx context.Context) string {
	if v := ctx.Value(TraceIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// WithSpanID adds span ID to context.
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, SpanIDKey, spanID)
}

// GetSpanID retrieves span ID from context.
func GetSpanID(ctx context.Context) string {
	if v := ctx.Value(SpanIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// WithTenantID adds tenant ID to context.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// GetTenantID retrieves tenant ID from context.
func GetTenantID(ctx context.Context) string {
	if v := ctx.Value(TenantIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// WithClientIP adds client IP to context.
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ClientIPKey, ip)
}

// GetClientIP retrieves client IP from context.
func GetClientIP(ctx context.Context) string {
	if v := ctx.Value(ClientIPKey); v != nil {
		if ip, ok := v.(string); ok {
			return ip
		}
	}
	return ""
}

// WithUserAgent adds user agent to context.
func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, UserAgentKey, userAgent)
}

// GetUserAgent retrieves user agent from context.
func GetUserAgent(ctx context.Context) string {
	if v := ctx.Value(UserAgentKey); v != nil {
		if ua, ok := v.(string); ok {
			return ua
		}
	}
	return ""
}

// WithRealIP adds real IP to context.
func WithRealIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, RealIPKey, ip)
}

// GetRealIP retrieves real IP from context.
func GetRealIP(ctx context.Context) string {
	if v := ctx.Value(RealIPKey); v != nil {
		if ip, ok := v.(string); ok {
			return ip
		}
	}
	return ""
}

// WithSessionID adds session ID to context.
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, SessionIDKey, sessionID)
}

// GetSessionID retrieves session ID from context.
func GetSessionID(ctx context.Context) string {
	if v := ctx.Value(SessionIDKey); v != nil {
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
