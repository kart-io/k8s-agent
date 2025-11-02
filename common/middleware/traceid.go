// Package middleware provides HTTP middleware for Gin framework.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kart-io/k8s-agent/common/contextx"
)

const (
	// TraceIDHeader is the HTTP header name for trace ID.
	TraceIDHeader = "X-Trace-ID"
)

// TraceID is a Gin middleware that adds distributed tracing support.
// It extracts or generates a trace ID for each request and propagates it through:
// 1. Request context (for downstream processing)
// 2. Response headers (for client visibility)
// 3. Logger context (for correlated logging)
//
// Usage:
//
//	router := gin.New()
//	router.Use(middleware.TraceID())
//
// The trace ID can be accessed in handlers via:
//
//	traceID := contextx.GetTraceID(c.Request.Context())
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get trace ID from request header
		traceID := c.Request.Header.Get(TraceIDHeader)

		// Generate new trace ID if not provided
		if traceID == "" {
			traceID = uuid.New().String()
			c.Request.Header.Set(TraceIDHeader, traceID)
		}

		// Add trace ID to response headers for client visibility
		c.Writer.Header().Set(TraceIDHeader, traceID)

		// Inject trace ID into context for downstream processing
		ctx := contextx.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		// Continue to next handler
		c.Next()
	}
}

// TraceIDConfig provides configuration options for TraceID middleware.
type TraceIDConfig struct {
	// HeaderName is the HTTP header name for trace ID (default: "X-Trace-ID").
	HeaderName string

	// Generator is a custom trace ID generator function.
	// If nil, uses uuid.New().String() by default.
	Generator func() string

	// SkipPaths is a list of paths that should skip trace ID generation.
	// Useful for health check endpoints that don't need tracing.
	SkipPaths []string
}

// TraceIDWithConfig creates a TraceID middleware with custom configuration.
//
// Example:
//
//	router.Use(middleware.TraceIDWithConfig(middleware.TraceIDConfig{
//	    HeaderName: "X-Request-ID",
//	    SkipPaths:  []string{"/health", "/metrics"},
//	    Generator: func() string {
//	        return fmt.Sprintf("trace-%d", time.Now().UnixNano())
//	    },
//	}))
func TraceIDWithConfig(config TraceIDConfig) gin.HandlerFunc {
	// Set defaults
	if config.HeaderName == "" {
		config.HeaderName = TraceIDHeader
	}
	if config.Generator == nil {
		config.Generator = func() string {
			return uuid.New().String()
		}
	}

	// Build skip paths map for O(1) lookup
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		// Skip trace ID generation for specified paths
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Try to get trace ID from request header
		traceID := c.Request.Header.Get(config.HeaderName)

		// Generate new trace ID if not provided
		if traceID == "" {
			traceID = config.Generator()
			c.Request.Header.Set(config.HeaderName, traceID)
		}

		// Add trace ID to response headers
		c.Writer.Header().Set(config.HeaderName, traceID)

		// Inject trace ID into context
		ctx := contextx.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		// Continue to next handler
		c.Next()
	}
}
