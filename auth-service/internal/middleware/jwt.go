package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/auth-service/internal/storage"
	"github.com/kart-io/k8s-agent/auth-service/pkg/jwt"
)

// JWTAuth creates JWT authentication middleware
func JWTAuth(secret string, redis *storage.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"code":    401,
				"details": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Check Bearer format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"code":    401,
				"details": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Check if token is blacklisted
		ctx := context.Background()
		blacklisted, err := redis.IsTokenBlacklisted(ctx, tokenString)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"code":    500,
				"details": "Failed to check token blacklist",
			})
			c.Abort()
			return
		}

		if blacklisted {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"code":    401,
				"details": "Token has been revoked",
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwt.ValidateToken(tokenString, secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"code":    401,
				"details": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("token", tokenString)

		c.Next()
	}
}
