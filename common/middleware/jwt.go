package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kart-io/k8s-agent/common/response"
)

// SessionValidator interface for session validation
// Implement this interface if you want to add session revocation checking
type SessionValidator interface {
	ValidateSession(ctx context.Context, jti string) (bool, error)
}

// JWTConfig holds configuration for JWT middleware
type JWTConfig struct {
	Secret           []byte
	SessionValidator SessionValidator // Optional: for session revocation checking
}

// JWTMiddleware provides JWT authentication middleware
type JWTMiddleware struct {
	config *JWTConfig
}

// NewJWTMiddleware creates a new JWT middleware with config
func NewJWTMiddleware(config *JWTConfig) *JWTMiddleware {
	return &JWTMiddleware{
		config: config,
	}
}

// Auth middleware validates JWT token and optionally checks session revocation
func (m *JWTMiddleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Missing authorization header", nil)
			c.Abort()
			return
		}

		// Check Bearer scheme
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid authorization header format. Expected: Bearer <token>", nil)
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
			return m.config.Secret, nil
		})

		if err != nil {
			response.Unauthorized(c, "Invalid or expired token", err)
			c.Abort()
			return
		}

		if !token.Valid {
			response.Unauthorized(c, "Token is not valid", nil)
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Unauthorized(c, "Invalid token claims", nil)
			c.Abort()
			return
		}

		// Extract user information from claims
		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			response.Unauthorized(c, "Invalid user_id in token", nil)
			c.Abort()
			return
		}

		username, _ := claims["username"].(string)

		// Extract JTI (JWT ID) for session tracking
		jti, hasJTI := claims["jti"].(string)

		// If token has JTI and session validator is configured, check revocation status
		if hasJTI && jti != "" && m.config.SessionValidator != nil {
			// Create context with timeout for session validation
			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()

			// Check if session is revoked
			isValid, err := m.config.SessionValidator.ValidateSession(ctx, jti)
			if err != nil {
				// SECURITY FIX: On validation error, reject the request to be safe
				// This prevents potential security bypasses when the session store is unreachable
				response.Unauthorized(c, "Session validation failed. Please try again.", err)
				c.Abort()
				return
			} else if !isValid {
				// Session has been revoked (forced logout)
				c.Header("X-Session-Terminated", "forced-logout")
				response.Unauthorized(c, "Your session has been terminated by an administrator. Please log in again.", nil)
				c.Abort()
				return
			}
		} else if hasJTI && m.config.SessionValidator == nil {
			// Token has JTI but no validator configured (backward compatibility)
			c.Header("X-Session-Tracking", "no-validator")
		} else if !hasJTI {
			// Token doesn't have JTI (backward compatibility with old tokens)
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

// OptionalAuth middleware validates JWT if present, but doesn't require it
// Useful for endpoints that work for both authenticated and anonymous users
func (m *JWTMiddleware) OptionalAuth() gin.HandlerFunc {
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
			return m.config.Secret, nil
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

// GenerateToken generates a JWT token with given claims
func (m *JWTMiddleware) GenerateToken(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.config.Secret)
}

// RefreshToken generates a new token with the same claims but extended expiration
func (m *JWTMiddleware) RefreshToken(oldToken string, expiresHours int) (string, error) {
	// Parse old token without validation (to allow expired tokens)
	token, err := jwt.Parse(oldToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return m.config.Secret, nil
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
	return m.GenerateToken(newClaims)
}
