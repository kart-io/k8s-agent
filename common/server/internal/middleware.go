// Package internal provides shared internal utilities for server implementations.
package internal

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/common/utils"
	"github.com/kart-io/logger/core"
)

// Context keys for trace and request IDs
type contextKey string

const (
	traceIDKey   contextKey = "trace_id"
	requestIDKey contextKey = "request_id"
)

// GetTraceID retrieves trace ID from context
func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		return id
	}
	return ""
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// WithTraceID adds trace ID to context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetOrCreateTraceID gets existing trace ID or creates a new one
func GetOrCreateTraceID(ctx context.Context) (context.Context, string) {
	traceID := GetTraceID(ctx)
	if traceID == "" {
		traceID = utils.GenerateID()
		ctx = WithTraceID(ctx, traceID)
	}
	return ctx, traceID
}

// GetOrCreateRequestID gets existing request ID or creates a new one
func GetOrCreateRequestID(ctx context.Context) (context.Context, string) {
	requestID := GetRequestID(ctx)
	if requestID == "" {
		requestID = utils.GenerateID()
		ctx = WithRequestID(ctx, requestID)
	}
	return ctx, requestID
}

// Middleware 定义框架无关的中间件类型
// 中间件接收一个 HTTP handler 并返回一个新的 handler
type Middleware func(http.Handler) http.Handler

// LoggerMiddleware creates a framework-agnostic logging middleware.
func LoggerMiddleware(logger core.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ctx := r.Context()

			// Extract trace and request IDs from context
			traceID := GetTraceID(ctx)
			requestID := GetRequestID(ctx)

			// Create response wrapper to capture status code
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call next handler
			next.ServeHTTP(rw, r)

			// Log request
			logger.Infow("HTTP request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.statusCode,
				"duration", time.Since(start).String(),
				"trace_id", traceID,
				"request_id", requestID,
				"client_ip", r.RemoteAddr,
			)
		})
	}
}

// RecoveryMiddleware creates a framework-agnostic panic recovery middleware.
func RecoveryMiddleware(logger core.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					ctx := r.Context()
					traceID := GetTraceID(ctx)
					requestID := GetRequestID(ctx)

					logger.Errorw("Panic recovered",
						"error", err,
						"path", r.URL.Path,
						"method", r.Method,
						"trace_id", traceID,
						"request_id", requestID,
					)

					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware creates a framework-agnostic CORS middleware.
func CORSMiddleware(allowOrigins []string, allowMethods []string, allowHeaders []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Simple CORS - allow all origins if not specified
			if len(allowOrigins) == 0 || utils.Contains(allowOrigins, origin) || utils.Contains(allowOrigins, "*") {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				if origin == "" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				}
			}

			// Set allowed methods
			if len(allowMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", utils.Join(allowMethods, ", "))
			} else {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			}

			// Set allowed headers
			if len(allowHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", utils.Join(allowHeaders, ", "))
			} else {
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, X-Trace-ID")
			}

			w.Header().Set("Access-Control-Max-Age", "3600")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TraceIDMiddleware creates a framework-agnostic trace ID middleware.
func TraceIDMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get or create trace ID
			ctx := r.Context()
			ctx, traceID := GetOrCreateTraceID(ctx)

			// Update request with new context
			r = r.WithContext(ctx)

			// Set response header
			w.Header().Set("X-Trace-ID", traceID)

			next.ServeHTTP(w, r)
		})
	}
}

// RequestIDMiddleware creates a framework-agnostic request ID middleware.
func RequestIDMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get request ID from header or create new one
			ctx := r.Context()
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				ctx, requestID = GetOrCreateRequestID(ctx)
			} else {
				ctx = WithRequestID(ctx, requestID)
			}

			// Update request with new context
			r = r.WithContext(ctx)

			// Set response header
			w.Header().Set("X-Request-ID", requestID)

			next.ServeHTTP(w, r)
		})
	}
}

// ConvertGinMiddleware converts a Gin middleware to framework-agnostic middleware.
// This is useful for gradually migrating existing Gin middleware.
// Note: This is a simple adapter and may not work for all Gin middleware.
func ConvertGinMiddleware(ginMW gin.HandlerFunc) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For now, just call the next handler
			// Full Gin context emulation is complex and out of scope
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
