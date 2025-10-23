package middleware

import (
	"github.com/gin-gonic/gin"
	commonmiddleware "github.com/kart-io/k8s-agent/common/middleware"
	"github.com/kart-io/k8s-agent/internal/auth/forced-logout/session"
)

// JWTMiddleware wraps the common JWT middleware with session service
type JWTMiddleware struct {
	*commonmiddleware.JWTMiddleware
}

// NewJWTMiddleware creates a new JWT middleware that uses the common JWT middleware
// and integrates with the session service for forced logout checking
func NewJWTMiddleware(jwtSecret string, sessionService *session.Service) *JWTMiddleware {
	config := &commonmiddleware.JWTConfig{
		Secret:           []byte(jwtSecret),
		SessionValidator: sessionService, // session.Service implements SessionValidator interface
	}

	return &JWTMiddleware{
		JWTMiddleware: commonmiddleware.NewJWTMiddleware(config),
	}
}

// JWTAuth is an alias for Auth() to maintain backward compatibility
func (m *JWTMiddleware) JWTAuth() gin.HandlerFunc {
	return m.Auth()
}

// OptionalJWTAuth is an alias for OptionalAuth() to maintain backward compatibility
func (m *JWTMiddleware) OptionalJWTAuth() gin.HandlerFunc {
	return m.OptionalAuth()
}
