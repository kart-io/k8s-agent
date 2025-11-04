// Package contextutil provides unified context management utilities.
// This package consolidates trace ID, request ID, and other context values
// management that was previously duplicated across different packages.
package contextutil

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Context key types using unexported structs for type safety.
type (
	requestIDKey struct{}
	traceIDKey   struct{}
	userIDKey    struct{}
	usernameKey  struct{}
	tenantIDKey  struct{}
	sessionIDKey struct{}
)

// TraceID operations

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

func GetTraceID(ctx context.Context) string {
	if v := ctx.Value(traceIDKey{}); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

func GetOrCreateTraceID(ctx context.Context) (context.Context, string) {
	traceID := GetTraceID(ctx)
	if traceID == "" {
		traceID = uuid.New().String()
		ctx = WithTraceID(ctx, traceID)
	}
	return ctx, traceID
}

// RequestID operations

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(requestIDKey{}); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

func GetOrCreateRequestID(ctx context.Context) (context.Context, string) {
	requestID := GetRequestID(ctx)
	if requestID == "" {
		requestID = uuid.New().String()
		ctx = WithRequestID(ctx, requestID)
	}
	return ctx, requestID
}

// UserID operations

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

func GetUserID(ctx context.Context) string {
	if v := ctx.Value(userIDKey{}); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// Username operations

func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameKey{}, username)
}

func GetUsername(ctx context.Context) string {
	if v := ctx.Value(usernameKey{}); v != nil {
		if name, ok := v.(string); ok {
			return name
		}
	}
	return ""
}

// TenantID operations

func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey{}, tenantID)
}

func GetTenantID(ctx context.Context) string {
	if v := ctx.Value(tenantIDKey{}); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// SessionID operations

func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey{}, sessionID)
}

func GetSessionID(ctx context.Context) string {
	if v := ctx.Value(sessionIDKey{}); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// WithTimeout is a convenience wrapper for context.WithTimeout
func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, timeout)
}

// WithDeadline is a convenience wrapper for context.WithDeadline
func WithDeadline(parent context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(parent, deadline)
}
