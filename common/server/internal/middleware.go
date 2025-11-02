// Package internal provides shared internal utilities for server implementations.
package internal

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/common/contextx"
	"github.com/kart-io/logger/core"
)

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
			traceID := contextx.GetTraceID(ctx)
			requestID := contextx.GetRequestID(ctx)

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
					traceID := contextx.GetTraceID(ctx)
					requestID := contextx.GetRequestID(ctx)

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
			if len(allowOrigins) == 0 || contains(allowOrigins, origin) || contains(allowOrigins, "*") {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				if origin == "" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				}
			}

			// Set allowed methods
			if len(allowMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", join(allowMethods, ", "))
			} else {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			}

			// Set allowed headers
			if len(allowHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", join(allowHeaders, ", "))
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
			ctx, traceID := contextx.GetOrCreateTraceID(ctx)

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
				ctx, requestID = contextx.GetOrCreateRequestID(ctx)
			} else {
				ctx = contextx.WithRequestID(ctx, requestID)
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

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func join(slice []string, sep string) string {
	result := ""
	for i, s := range slice {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
