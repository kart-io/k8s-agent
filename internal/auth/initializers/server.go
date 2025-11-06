package initializers

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/cmd/auth/app/options"
	"github.com/kart-io/k8s-agent/internal/auth/handler"
	"github.com/kart-io/k8s-agent/internal/auth/middleware"
	"github.com/kart-io/k8s-agent/internal/auth/routes"
	"github.com/kart-io/k8s-agent/internal/auth/service"
	"github.com/kart-io/k8s-agent/internal/auth/storage"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// HTTPServerInitializer wraps the common HTTP server initializer.
// 使用标准化的 common/server 框架替代手动服务器管理
type HTTPServerInitializer struct {
	*pkginitializers.HTTPServerInitializer
	cfg              *options.ServerOptions
	logger           core.Logger
	dbInit           *DatabaseInitializer
	redisInit        *RedisInitializer
	sessionInit      *SessionServiceInitializer
	auditInit        *AuditServiceInitializer
	notificationInit *NotificationServiceInitializer
	forcedLogoutInit *ForcedLogoutServiceInitializer
	emailInit        *EmailClientInitializer
}

// NewHTTPServerInitializer creates a new HTTP server initializer using common/server framework.
func NewHTTPServerInitializer(
	cfg *options.ServerOptions,
	logger core.Logger,
	dbInit *DatabaseInitializer,
	redisInit *RedisInitializer,
	sessionInit *SessionServiceInitializer,
	auditInit *AuditServiceInitializer,
	notificationInit *NotificationServiceInitializer,
	forcedLogoutInit *ForcedLogoutServiceInitializer,
	emailInit *EmailClientInitializer,
) *HTTPServerInitializer {
	h := &HTTPServerInitializer{
		cfg:              cfg,
		logger:           logger,
		dbInit:           dbInit,
		redisInit:        redisInit,
		sessionInit:      sessionInit,
		auditInit:        auditInit,
		notificationInit: notificationInit,
		forcedLogoutInit: forcedLogoutInit,
		emailInit:        emailInit,
	}

	// Create standard HTTP server config
	serverConfig := &pkginitializers.HTTPServerConfig{
		Name:     "auth-http-server",
		Priority: bootstrap.PriorityHTTP,
		Config:   cfg.Server,
		// CORS and RateLimit are handled in the route setup
		RouteSetup: h.setupRoutes, // Method defined below
	}

	// Create standard HTTP server initializer
	h.HTTPServerInitializer = pkginitializers.NewHTTPServerInitializer(serverConfig, logger)
	return h
}

// setupRoutes configures all routes for the auth service.
// This is called by the common HTTP server initializer during Initialize.
func (h *HTTPServerInitializer) setupRoutes(engine *gin.Engine) error {
	h.logger.Infow("Setting up auth service routes")

	// Create MySQLDB wrapper (for compatibility with existing handlers)
	dbConn := &storage.MySQLDB{DB: h.dbInit.DB()}

	// Initialize services (user, role, permission services)
	userService := service.NewUserService(dbConn)
	roleService := service.NewRoleService(dbConn)
	permissionService := service.NewPermissionService(dbConn)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(
		h.dbInit.DB(),
		h.cfg.JWT.Secret,
		h.cfg.JWT.ExpiresHours,
		h.sessionInit.Service(),
	)
	sessionHandler := handler.NewSessionHandler(h.sessionInit.Service())
	forcedLogoutHandler := handler.NewForcedLogoutHandler(h.forcedLogoutInit.Service())
	auditHandler := handler.NewAuditHandler(h.auditInit.Service())
	userHandler := handler.NewUserHandler(userService)
	roleHandler := handler.NewRoleHandler(roleService)
	permissionHandler := handler.NewPermissionHandler(permissionService)

	// Initialize middleware
	jwtMiddleware := middleware.NewJWTMiddleware(h.cfg.JWT.Secret, h.sessionInit.Service())
	authMiddleware := middleware.NewForcedLogoutAuthMiddleware(h.dbInit.DB())
	rateLimiter := middleware.NewRateLimiter(h.redisInit.Client())

	// Register health check endpoint
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":              "healthy",
			"service":             "auth-service",
			"version":             "1.0.0",
			"email_notifications": h.cfg.Email.Enabled,
			"database":            h.dbInit.DB() != nil,
			"redis":               h.redisInit.Client() != nil,
		})
	})

	// Register authentication routes
	v1 := engine.Group("/api/v1/auth")
	{
		v1.POST("/login", authHandler.LoginHandler)
		v1.POST("/logout", jwtMiddleware.JWTAuth(), authHandler.LogoutHandler)
		v1.POST("/refresh", authHandler.RefreshTokenHandler) // No auth required - uses refresh token
		v1.GET("/me", jwtMiddleware.JWTAuth(), authHandler.GetCurrentUserHandler)
		v1.GET("/codes", jwtMiddleware.JWTAuth(), authHandler.GetAccessCodesHandler)
	}

	// Register user management routes
	userRoutes := engine.Group("/api/v1/users")
	userRoutes.Use(jwtMiddleware.JWTAuth())
	{
		userRoutes.GET("", userHandler.List)
		userRoutes.GET("/:id", userHandler.GetByID)
		userRoutes.POST("", userHandler.Create)
		userRoutes.PUT("/:id", userHandler.Update)
		userRoutes.DELETE("/:id", userHandler.Delete)
		userRoutes.POST("/:id/roles", userHandler.AssignRoles)
	}

	// Register role management routes
	roleRoutes := engine.Group("/api/v1/roles")
	roleRoutes.Use(jwtMiddleware.JWTAuth())
	{
		roleRoutes.GET("", roleHandler.List)
		roleRoutes.GET("/:id", roleHandler.GetByID)
		roleRoutes.POST("", roleHandler.Create)
		roleRoutes.PUT("/:id", roleHandler.Update)
		roleRoutes.DELETE("/:id", roleHandler.Delete)
		roleRoutes.POST("/:id/permissions", roleHandler.AssignPermissions)
		roleRoutes.GET("/:id/permissions", roleHandler.GetPermissions)
	}

	// Register permission management routes
	permissionRoutes := engine.Group("/api/v1/permissions")
	permissionRoutes.Use(jwtMiddleware.JWTAuth())
	{
		permissionRoutes.GET("", permissionHandler.List)
		permissionRoutes.GET("/tree", permissionHandler.GetTree)
		permissionRoutes.GET("/:id", permissionHandler.GetByID)
		permissionRoutes.POST("", permissionHandler.Create)
		permissionRoutes.PUT("/:id", permissionHandler.Update)
		permissionRoutes.DELETE("/:id", permissionHandler.Delete)
	}

	// Register forced logout routes
	forcedLogoutRoutes := routes.NewForcedLogoutRoutes(
		sessionHandler,
		forcedLogoutHandler,
		auditHandler,
		jwtMiddleware,
		authMiddleware,
		rateLimiter,
	)
	forcedLogoutRoutes.RegisterRoutes(engine)

	// Log registered routes
	h.logger.Infow("Auth Service Routes Registered",
		"health", "GET /health",
		"auth_login", "POST /api/v1/auth/login",
		"auth_logout", "POST /api/v1/auth/logout",
		"auth_refresh", "POST /api/v1/auth/refresh",
		"users", "GET/POST/PUT/DELETE /api/v1/users",
		"roles", "GET/POST/PUT/DELETE /api/v1/roles",
		"permissions", "GET/POST/PUT/DELETE /api/v1/permissions",
		"sessions", "Forced logout routes registered",
	)

	return nil
}

// Initialize overrides to add custom initialization if needed.
// The actual server initialization is handled by the embedded HTTPServerInitializer.
func (h *HTTPServerInitializer) Initialize(ctx context.Context) error {
	// The embedded HTTPServerInitializer will handle server creation and route setup
	return h.HTTPServerInitializer.Initialize(ctx)
}

// Close is handled by the embedded HTTPServerInitializer.
// The server lifecycle is managed by the bootstrap framework.
