package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kart-io/logger/core"
)

// RequestLogger logs HTTP request details.
func RequestLogger(log core.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Extract initial request details
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(startTime)
		statusCode := c.Writer.Status()

		// Get user ID if authenticated
		userID, exists := c.Get("user_id")
		if !exists {
			userID = "-"
		}

		// Get auth type if exists
		authType, exists := c.Get("auth_type")
		if !exists {
			authType = "-"
		}

		// Log with different levels based on status code
		switch {
		case statusCode >= 500:
			log.Errorw("Server error",
				"method", method,
				"path", path,
				"status", statusCode,
				"latency_ms", latency.Milliseconds(),
				"client_ip", clientIP,
				"user_id", userID,
				"auth_type", authType,
			)
		case statusCode >= 400:
			log.Warnw("Client error",
				"method", method,
				"path", path,
				"status", statusCode,
				"latency_ms", latency.Milliseconds(),
				"client_ip", clientIP,
				"user_id", userID,
				"auth_type", authType,
			)
		case statusCode >= 300:
			log.Debugw("Redirect",
				"method", method,
				"path", path,
				"status", statusCode,
				"latency_ms", latency.Milliseconds(),
				"client_ip", clientIP,
			)
		default:
			log.Infow("Request processed",
				"method", method,
				"path", path,
				"status", statusCode,
				"latency_ms", latency.Milliseconds(),
				"client_ip", clientIP,
				"user_id", userID,
				"auth_type", authType,
			)
		}
	}
}
