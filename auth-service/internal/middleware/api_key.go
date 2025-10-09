package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/auth-service/internal/service"
)

// APIKeyAuth creates API key authentication middleware
// Alternative to JWT for service accounts
func APIKeyAuth(apikeyService *service.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract key and secret from headers
		key := c.GetHeader("X-API-Key")
		secret := c.GetHeader("X-API-Secret")

		if key == "" || secret == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"code":    401,
				"details": "API key and secret are required",
			})
			c.Abort()
			return
		}

		// Validate API key
		apiKey, err := apikeyService.Validate(key, secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"code":    401,
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Set user context from API key
		c.Set("user_id", apiKey.UserID)
		c.Set("auth_type", "apikey")

		c.Next()
	}
}
