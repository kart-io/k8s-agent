package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kart-io/logger/core"
)

// Context keys for trace and request IDs
type contextKey string

const (
	traceIDKey   contextKey = "trace_id"
	requestIDKey contextKey = "request_id"
)

// getTraceID retrieves trace ID from context
func getTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		return id
	}
	return ""
}

// getRequestID retrieves request ID from context
func getRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// RequestLogger is a Gin middleware for structured request logging.
// It logs request details including trace ID and request ID from context.
//
// This middleware should be used after TraceID() and RequestID() middlewares
// to ensure trace and request IDs are available in the context.
//
// Usage:
//
//	router := gin.New()
//	router := gin.New()
//	router.Use(middleware.TraceID())
//	router.Use(middleware.RequestID())
//	router.Use(middleware.RequestLoggerWithLogger(logger))
func RequestLoggerWithLogger(logger core.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate request duration
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Extract trace and request IDs from context
		ctx := c.Request.Context()
		traceID := getTraceID(ctx)
		requestID := getRequestID(ctx)

		// Build structured log fields
		logFields := []interface{}{
			"method", method,
			"path", path,
			"query", query,
			"status", statusCode,
			"latency_ms", latency.Milliseconds(),
			"client_ip", clientIP,
			"user_agent", c.Request.UserAgent(),
		}

		// Add trace ID if available
		if traceID != "" {
			logFields = append(logFields, "trace_id", traceID)
		}

		// Add request ID if available
		if requestID != "" {
			logFields = append(logFields, "request_id", requestID)
		}

		// Add error message if present
		if errorMessage != "" {
			logFields = append(logFields, "error", errorMessage)
		}

		// Log based on status code
		if statusCode >= 400 {
			// 4xx and 5xx errors
			logger.Errorw("Request error", logFields...)
		} else {
			logger.Infow("Request completed", logFields...)
		}
	}
}

// RequestLogger returns a request logging middleware with a noop logger.
// Deprecated: Use RequestLoggerWithLogger instead.
func RequestLogger() gin.HandlerFunc {
	return RequestLoggerWithLogger(core.NewNoOpLogger(nil))
}

// RequestID is a Gin middleware that adds request ID support.
// It extracts or generates a request ID for each request and propagates it through:
// 1. Request context (for downstream processing)
// 2. Response headers (for client visibility)
// 3. Gin context (for backward compatibility)
//
// This follows OneX best practices by injecting into context.Context
// instead of just Gin's context.
//
// Usage:
//
//	router := gin.New()
//	router.Use(middleware.RequestID())
//
// The request ID can be accessed in handlers via:
//
//	requestID := contextutil.GetRequestID(c.Request.Context())
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get request ID from header
		requestID := c.Request.Header.Get("X-Request-ID")

		// Generate new request ID if not provided
		if requestID == "" {
			requestID = uuid.New().String()
			c.Request.Header.Set("X-Request-ID", requestID)
		}

		// Add request ID to response headers for client visibility
		c.Writer.Header().Set("X-Request-ID", requestID)

		// Inject into context.Context
		ctx := context.WithValue(c.Request.Context(), requestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		// Continue to next handler
		c.Next()
	}
}
