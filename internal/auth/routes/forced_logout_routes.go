package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/internal/auth/handler"
	"github.com/kart-io/k8s-agent/internal/auth/middleware"
)

// ForcedLogoutRoutes registers all forced logout related routes.
type ForcedLogoutRoutes struct {
	sessionHandler      *handler.SessionHandler
	forcedLogoutHandler *handler.ForcedLogoutHandler
	auditHandler        *handler.AuditHandler
	jwtMiddleware       *middleware.JWTMiddleware
	authMiddleware      *middleware.ForcedLogoutAuthMiddleware
	rateLimiter         *middleware.RateLimiter
}

// NewForcedLogoutRoutes creates a new forced logout routes instance.
func NewForcedLogoutRoutes(
	sessionHandler *handler.SessionHandler,
	forcedLogoutHandler *handler.ForcedLogoutHandler,
	auditHandler *handler.AuditHandler,
	jwtMiddleware *middleware.JWTMiddleware,
	authMiddleware *middleware.ForcedLogoutAuthMiddleware,
	rateLimiter *middleware.RateLimiter,
) *ForcedLogoutRoutes {
	return &ForcedLogoutRoutes{
		sessionHandler:      sessionHandler,
		forcedLogoutHandler: forcedLogoutHandler,
		auditHandler:        auditHandler,
		jwtMiddleware:       jwtMiddleware,
		authMiddleware:      authMiddleware,
		rateLimiter:         rateLimiter,
	}
}

// RegisterRoutes registers all forced logout routes under /api/v1.
func (r *ForcedLogoutRoutes) RegisterRoutes(router *gin.Engine) {
	// Create API v1 group
	v1 := router.Group("/api/v1")

	// Rate limit configuration for forced logout operations
	rateLimitConfig := middleware.DefaultForcedLogoutRateLimit()

	// Middleware chain: JWT → SessionAdmin → RateLimit → Handler
	// This ensures proper authentication, authorization, and rate limiting
	authChain := []gin.HandlerFunc{
		r.jwtMiddleware.JWTAuth(),
		r.authMiddleware.RequireSessionAdmin(),
		r.rateLimiter.RateLimit(rateLimitConfig),
	}

	// Session Management Routes
	sessionsGroup := v1.Group("/sessions")
	sessionsGroup.Use(authChain...)
	{
		// GET /api/v1/sessions/users/:userId - List user sessions
		sessionsGroup.GET("/users/:userId", r.sessionHandler.ListUserSessions)
	}

	// Forced Logout Routes
	forcedLogoutGroup := v1.Group("/forced-logout")
	forcedLogoutGroup.Use(authChain...)
	{
		// POST /api/v1/forced-logout/session/:jti - Force logout single session
		forcedLogoutGroup.POST("/session/:jti", r.forcedLogoutHandler.ForceLogoutSession)

		// POST /api/v1/forced-logout/user/:userId - Force logout all user sessions
		forcedLogoutGroup.POST("/user/:userId", r.forcedLogoutHandler.ForceLogoutUser)

		// POST /api/v1/forced-logout/sessions - Bulk force logout
		forcedLogoutGroup.POST("/sessions", r.forcedLogoutHandler.BulkForceLogout)
	}

	// Audit Routes
	auditGroup := v1.Group("/audit")
	auditGroup.Use(authChain...)
	{
		// GET /api/v1/audit/forced-logout - Query audit events
		auditGroup.GET("/forced-logout", r.auditHandler.ListAuditEvents)

		// GET /api/v1/audit/forced-logout/export - Export audit logs
		auditGroup.GET("/forced-logout/export", r.auditHandler.ExportAuditEvents)
	}
}

// PrintRegisteredRoutes logs all registered forced logout routes.
func (r *ForcedLogoutRoutes) PrintRegisteredRoutes() []string {
	routes := []string{
		"GET    /api/v1/sessions/users/:userId          - List user active sessions",
		"POST   /api/v1/forced-logout/session/:jti      - Force logout single session",
		"POST   /api/v1/forced-logout/user/:userId      - Force logout all user sessions",
		"POST   /api/v1/forced-logout/sessions          - Bulk force logout multiple sessions",
		"GET    /api/v1/audit/forced-logout             - Query forced logout audit events",
		"GET    /api/v1/audit/forced-logout/export      - Export audit logs (JSON/CSV)",
	}
	return routes
}
