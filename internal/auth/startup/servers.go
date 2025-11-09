// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package startup

import (
	"context"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	commonserver "github.com/kart-io/k8s-agent/common/server"
	authgrpc "github.com/kart-io/k8s-agent/internal/auth/grpc"
	"github.com/kart-io/k8s-agent/internal/auth/handler"
	"github.com/kart-io/k8s-agent/internal/auth/middleware"
	"github.com/kart-io/k8s-agent/internal/auth/routes"
	authv1 "github.com/kart-io/k8s-agent/pkg/api/auth/v1"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// GRPCServerInitializer initializes gRPC server.
type GRPCServerInitializer struct {
	opts         *commonapp.StandardOptions
	logger       core.Logger
	coreServices *CoreServicesInitializer
	sessionInit  *SessionServiceInitializer

	standardInit *pkginitializers.GRPCServerInitializer
}

// NewGRPCServerInitializer creates a new gRPC server initializer.
func NewGRPCServerInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	coreServices *CoreServicesInitializer,
	sessionInit *SessionServiceInitializer,
) *GRPCServerInitializer {
	return &GRPCServerInitializer{
		opts:         opts,
		logger:       logger,
		coreServices: coreServices,
		sessionInit:  sessionInit,
	}
}

// Name returns the initializer name.
func (g *GRPCServerInitializer) Name() string {
	return "auth-grpc-server"
}

// Priority returns initialization priority.
func (g *GRPCServerInitializer) Priority() int {
	return 900
}

// Initialize creates and starts the gRPC server.
func (g *GRPCServerInitializer) Initialize(ctx context.Context) error {
	if !g.opts.GRPC.Enable {
		g.logger.Infow("gRPC server is disabled, skipping initialization")
		return nil
	}

	g.logger.Infow("Initializing gRPC server",
		"host", g.opts.GRPC.Host,
		"port", g.opts.GRPC.Port,
	)

	services := g.coreServices.Services()

	// Create gRPC service servers using app's service instances
	authGRPC := authgrpc.NewAuthServiceServer(services.Auth, g.logger)
	userGRPC := authgrpc.NewUserServiceServer(services.User, g.logger)
	roleGRPC := authgrpc.NewRoleServiceServer(services.Role, g.logger)
	permissionGRPC := authgrpc.NewPermissionServiceServer(services.Permission, g.logger)
	sessionGRPC := authgrpc.NewSessionServiceServer(g.sessionInit.Service(), g.logger)

	serverConfig := &pkginitializers.GRPCServerConfig{
		Name:     g.Name(),
		Priority: g.Priority(),
		Config:   g.opts.GRPC,
		ServiceRegister: func(s *grpc.Server) error {
			authv1.RegisterAuthServiceServer(s, authGRPC)
			authv1.RegisterUserServiceServer(s, userGRPC)
			authv1.RegisterRoleServiceServer(s, roleGRPC)
			authv1.RegisterPermissionServiceServer(s, permissionGRPC)
			authv1.RegisterSessionServiceServer(s, sessionGRPC)

			g.logger.Infow("All gRPC services registered successfully",
				"services", []string{"Auth", "User", "Role", "Permission", "Session"},
			)
			return nil
		},
	}

	g.standardInit = pkginitializers.NewGRPCServerInitializer(serverConfig, g.logger)
	return g.standardInit.Initialize(ctx)
}

// GetServer returns the underlying gRPC server.
func (g *GRPCServerInitializer) GetServer() commonserver.Server {
	if g.standardInit == nil {
		return nil
	}
	return g.standardInit.GetServer()
}

// Close shuts down the gRPC server.
func (g *GRPCServerInitializer) Close(ctx context.Context) error {
	if g.standardInit == nil {
		return nil
	}
	return g.standardInit.Close(ctx)
}

// HTTPServerInitializer initializes HTTP server.
type HTTPServerInitializer struct {
	opts             *commonapp.StandardOptions
	logger           core.Logger
	infra            *InfrastructureInitializers
	coreServices     *CoreServicesInitializer
	sessionInit      *SessionServiceInitializer
	auditInit        *AuditServiceInitializer
	notificationInit *NotificationServiceInitializer
	forcedLogoutInit *ForcedLogoutServiceInitializer

	standardInit *pkginitializers.HTTPServerInitializer
}

// NewHTTPServerInitializer creates a new HTTP server initializer.
func NewHTTPServerInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	infra *InfrastructureInitializers,
	coreServices *CoreServicesInitializer,
	sessionInit *SessionServiceInitializer,
	auditInit *AuditServiceInitializer,
	notificationInit *NotificationServiceInitializer,
	forcedLogoutInit *ForcedLogoutServiceInitializer,
) *HTTPServerInitializer {
	return &HTTPServerInitializer{
		opts:             opts,
		logger:           logger,
		infra:            infra,
		coreServices:     coreServices,
		sessionInit:      sessionInit,
		auditInit:        auditInit,
		notificationInit: notificationInit,
		forcedLogoutInit: forcedLogoutInit,
	}
}

// Name returns the initializer name.
func (h *HTTPServerInitializer) Name() string {
	return "auth-http-server"
}

// Priority returns initialization priority.
func (h *HTTPServerInitializer) Priority() int {
	return 1000
}

// Initialize creates and starts the HTTP server.
func (h *HTTPServerInitializer) Initialize(ctx context.Context) error {
	h.logger.Infow("Initializing HTTP server",
		"host", h.opts.Server.Host,
		"port", h.opts.Server.Port,
	)

	// Configure HTTP server with route setup
	serverConfig := &pkginitializers.HTTPServerConfig{
		Name:     h.Name(),
		Priority: h.Priority(),
		Config:   h.opts.Server,
		RouteSetup: func(engine *gin.Engine) error {
			return h.setupRoutes(engine)
		},
	}

	h.standardInit = pkginitializers.NewHTTPServerInitializer(serverConfig, h.logger)
	return h.standardInit.Initialize(ctx)
}

// setupRoutes configures all HTTP routes.
func (h *HTTPServerInitializer) setupRoutes(engine *gin.Engine) error {
	h.logger.Infow("Setting up auth service routes")

	services := h.coreServices.Services()

	// Inject session service into auth service
	services.Auth.SetSessionService(h.sessionInit.Service())

	// Initialize handlers
	authHandler := handler.NewAuthHandler(services.Auth)
	sessionHandler := handler.NewSessionHandler(h.sessionInit.Service())
	forcedLogoutHandler := handler.NewForcedLogoutHandler(h.forcedLogoutInit.Service())
	auditHandler := handler.NewAuditHandler(h.auditInit.Service())
	userHandler := handler.NewUserHandler(services.User)
	roleHandler := handler.NewRoleHandler(services.Role)
	permissionHandler := handler.NewPermissionHandler(services.Permission)
	apikeyHandler := handler.NewAPIKeyHandler(services.APIKey)

	// Initialize middleware
	jwtMiddleware := middleware.NewJWTMiddleware(h.opts.JWT.Secret, h.sessionInit.Service())
	authMiddleware := middleware.NewForcedLogoutAuthMiddleware(h.infra.Database.DB())
	rateLimiter := middleware.NewRateLimiter(h.infra.Redis.Client())

	// Register health check endpoint
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":              "healthy",
			"service":             "auth-service",
			"version":             "1.0.0",
			"email_notifications": h.opts.Email.Enabled,
			"database":            h.infra.Database.DB() != nil,
			"redis":               h.infra.Redis.Client() != nil,
		})
	})

	// Register authentication routes
	v1 := engine.Group("/api/v1/auth")
	{
		v1.POST("/login", authHandler.LoginHandler)
		v1.POST("/logout", jwtMiddleware.JWTAuth(), authHandler.LogoutHandler)
		v1.POST("/refresh", authHandler.RefreshTokenHandler)
		v1.GET("/me", jwtMiddleware.JWTAuth(), authHandler.GetCurrentUserHandler)
		v1.GET("/codes", jwtMiddleware.JWTAuth(), authHandler.GetAccessCodesHandler)
		v1.GET("/menus", jwtMiddleware.JWTAuth(), authHandler.GetMenusHandler)
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

	// Register API Key management routes
	apikeyRoutes := engine.Group("/api/v1/api-keys")
	apikeyRoutes.Use(jwtMiddleware.JWTAuth())
	{
		apikeyRoutes.GET("", apikeyHandler.List)
		apikeyRoutes.POST("", apikeyHandler.Create)
		apikeyRoutes.DELETE("/:id", apikeyHandler.Delete)
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

	h.logger.Infow("Auth Service Routes Registered",
		"health", "GET /health",
		"auth_login", "POST /api/v1/auth/login",
		"auth_logout", "POST /api/v1/auth/logout",
		"auth_refresh", "POST /api/v1/auth/refresh",
		"auth_menus", "GET /api/v1/auth/menus",
		"users", "GET/POST/PUT/DELETE /api/v1/users",
		"roles", "GET/POST/PUT/DELETE /api/v1/roles",
		"permissions", "GET/POST/PUT/DELETE /api/v1/permissions",
		"api_keys", "GET/POST/DELETE /api/v1/api-keys",
		"sessions", "Forced logout routes registered",
	)

	return nil
}

// GetServer returns the underlying HTTP server.
func (h *HTTPServerInitializer) GetServer() commonserver.Server {
	if h.standardInit == nil {
		return nil
	}
	return h.standardInit.GetServer()
}

// Close shuts down the HTTP server.
func (h *HTTPServerInitializer) Close(ctx context.Context) error {
	if h.standardInit == nil {
		return nil
	}
	return h.standardInit.Close(ctx)
}
