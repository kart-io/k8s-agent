package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kart/k8s-agent/auth-service/pkg/forced-logout/session"
)

// JWTMiddleware provides JWT authentication with session revocation checking
type JWTMiddleware struct {
	jwtSecret      []byte
	sessionService *session.Service
}

// NewJWTMiddleware creates a new JWT middleware
func NewJWTMiddleware(jwtSecret string, sessionService *session.Service) *JWTMiddleware {
	return &JWTMiddleware{
		jwtSecret:      []byte(jwtSecret),
		sessionService: sessionService,
	}
}

// JWTAuth middleware validates JWT token and checks session revocation status
func (m *JWTMiddleware) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Missing authorization header",
			})
			c.Abort()
			return
		}

		// Check Bearer scheme
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid authorization header format. Expected: Bearer <token>",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return m.jwtSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid or expired token",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Token is not valid",
			})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid token claims",
			})
			c.Abort()
			return
		}

		// Extract user information from claims
		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid user_id in token",
			})
			c.Abort()
			return
		}

		username, _ := claims["username"].(string)

		// Extract JTI (JWT ID) for session tracking
		jti, hasJTI := claims["jti"].(string)

		// If token has JTI, check revocation status
		if hasJTI && jti != "" && m.sessionService != nil {
			// Create context with timeout for session validation
			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()

			// Check if session is revoked
			isValid, err := m.sessionService.ValidateSession(ctx, jti)
			if err != nil {
				// Log error but don't block request
				// In production, use proper structured logging
				c.Header("X-Session-Validation-Error", "true")
			} else if !isValid {
				// Session has been revoked (forced logout)
				c.Header("X-Session-Terminated", "forced-logout")
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "session_terminated",
					"message": "Your session has been terminated by an administrator. Please log in again.",
					"details": map[string]interface{}{
						"reason":         "forced_logout",
						"requires_login": true,
					},
				})
				c.Abort()
				return
			}
		} else {
			// Token doesn't have JTI (backward compatibility with old tokens)
			// Allow the request but mark it for monitoring
			c.Header("X-Session-Tracking", "legacy-token")
		}

		// Store user information in context for handlers
		c.Set("user_id", userID)
		c.Set("username", username)
		if hasJTI {
			c.Set("jti", jti)
		}

		c.Next()
	}
}

// OptionalJWTAuth middleware validates JWT if present, but doesn't require it
// Useful for endpoints that work for both authenticated and anonymous users
func (m *JWTMiddleware) OptionalJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token, continue without authentication
			c.Next()
			return
		}

		// Token present, validate it
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format, continue without authentication
			c.Next()
			return
		}

		tokenString := parts[1]

		// Parse token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return m.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// Valid token, extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok {
			if userID, ok := claims["user_id"].(string); ok {
				c.Set("user_id", userID)
			}
			if username, ok := claims["username"].(string); ok {
				c.Set("username", username)
			}
			if jti, ok := claims["jti"].(string); ok {
				c.Set("jti", jti)
			}
		}

		c.Next()
	}
}

// RefreshToken generates a new token with the same claims but extended expiration
func (m *JWTMiddleware) RefreshToken(oldToken string, expiresHours int) (string, error) {
	// Parse old token without validation (to allow expired tokens)
	token, err := jwt.Parse(oldToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return m.jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	// Extract old claims
	oldClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", jwt.ErrInvalidType
	}

	// Create new claims with extended expiration
	newClaims := jwt.MapClaims{
		"user_id":  oldClaims["user_id"],
		"username": oldClaims["username"],
		"jti":      oldClaims["jti"], // Keep same JTI for session tracking
		"exp":      time.Now().Add(time.Duration(expiresHours) * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	// Generate new token
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	return newToken.SignedString(m.jwtSecret)
}
