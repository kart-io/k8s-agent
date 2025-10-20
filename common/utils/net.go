package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// ExtractIPAddress extracts the client IP address from the Gin context
// It checks X-Forwarded-For, X-Real-IP headers and falls back to RemoteAddr
func ExtractIPAddress(c *gin.Context) string {
	// Check X-Forwarded-For header (may contain multiple IPs)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain (original client IP)
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header (single IP)
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to ClientIP (which handles RemoteAddr)
	ip := c.ClientIP()
	if ip == "" {
		ip = "Unknown"
	}

	return ip
}
