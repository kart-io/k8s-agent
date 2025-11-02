// Package middleware provides common HTTP middleware for the k8s-agent project.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/common/idempotent"
)

// IdempotentConfig configures the idempotency middleware.
type IdempotentConfig struct {
	// Handler is the idempotency handler
	Handler *idempotent.Handler

	// HeaderName is the header to check for idempotency key (default: "X-Idempotent-Key")
	HeaderName string

	// PathBlacklist defines paths that require idempotency checks
	// Format: "METHOD /path/pattern"
	// Example: "POST /api/v1/workflows", "POST /api/v1/tasks"
	PathBlacklist map[string]bool

	// SkipFunc is an optional function to skip idempotency checks
	// Return true to skip the check for the given request
	SkipFunc func(*gin.Context) bool
}

// DefaultPathBlacklist returns the default paths requiring idempotency checks.
// These are typically create/update operations that should not be duplicated.
func DefaultPathBlacklist() map[string]bool {
	return map[string]bool{
		// Orchestrator Service
		"POST /api/v1/workflows":             true,
		"POST /api/v1/strategies":            true,
		"POST /api/v1/workflows/:id/execute": true,

		// Agent Manager Service
		"POST /api/v1/commands": true,
		"POST /api/v1/events":   true,
		"POST /api/v1/agents":   true,
		"POST /api/v1/clusters": true,

		// Reasoning Service
		"POST /api/v1/analyze/root-cause": true,
		"POST /api/v1/recommendations":    true,
	}
}

// Idempotent returns a middleware that enforces idempotency for configured paths.
//
// The middleware checks for an idempotency key in the request header and uses
// the provided handler to ensure the operation is only processed once.
//
// Usage:
//
//	handler := idempotent.NewHandler(store, 24*time.Hour, 5*time.Minute)
//	router.Use(middleware.Idempotent(middleware.IdempotentConfig{
//	    Handler: handler,
//	    PathBlacklist: middleware.DefaultPathBlacklist(),
//	}))
func Idempotent(config IdempotentConfig) gin.HandlerFunc {
	// Apply defaults
	if config.HeaderName == "" {
		config.HeaderName = "X-Idempotent-Key"
	}
	if config.PathBlacklist == nil {
		config.PathBlacklist = DefaultPathBlacklist()
	}

	return func(c *gin.Context) {
		// Skip if handler not configured
		if config.Handler == nil {
			c.Next()
			return
		}

		// Skip if custom skip function returns true
		if config.SkipFunc != nil && config.SkipFunc(c) {
			c.Next()
			return
		}

		// Check if path requires idempotency
		route := c.Request.Method + " " + c.FullPath()
		if !config.PathBlacklist[route] {
			c.Next()
			return
		}

		// Get idempotency key from header
		key := c.GetHeader(config.HeaderName)
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Missing " + config.HeaderName + " header for idempotent operation",
				"code":  "MISSING_IDEMPOTENT_KEY",
			})
			c.Abort()
			return
		}

		// Check idempotency
		record, err := config.Handler.Check(c.Request.Context(), key)
		if err == nil && record != nil {
			// Operation exists
			switch record.Status {
			case idempotent.StatusCompleted:
				// Return cached response
				c.Header("X-Idempotent-Replayed", "true")
				c.Data(http.StatusOK, "application/json", record.Response)
				c.Abort()
				return

			case idempotent.StatusFailed:
				// Previous attempt failed
				c.JSON(http.StatusConflict, gin.H{
					"error": "Previous operation with this key failed: " + record.Error,
					"code":  "IDEMPOTENT_OPERATION_FAILED",
				})
				c.Abort()
				return

			case idempotent.StatusProcessing:
				// Duplicate request
				c.JSON(http.StatusConflict, gin.H{
					"error": "Duplicate request detected - operation is currently processing",
					"code":  "DUPLICATE_REQUEST",
				})
				c.Abort()
				return
			}
		}

		// Store key in context for handler to use
		c.Set("idempotent_key", key)

		c.Next()
	}
}

// GetIdempotentKey retrieves the idempotency key from the Gin context.
// This is useful in handlers that need to manually manage idempotency.
func GetIdempotentKey(c *gin.Context) string {
	if key, exists := c.Get("idempotent_key"); exists {
		if keyStr, ok := key.(string); ok {
			return keyStr
		}
	}
	return ""
}
