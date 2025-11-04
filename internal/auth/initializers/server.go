package initializers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/internal/auth/config"
	"github.com/kart-io/k8s-agent/internal/auth/handler"
	"github.com/kart-io/k8s-agent/internal/auth/middleware"
	"github.com/kart-io/k8s-agent/internal/auth/routes"
	"github.com/kart-io/k8s-agent/internal/auth/service"
	"github.com/kart-io/k8s-agent/internal/auth/storage"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
)

// HTTPServerInitializer HTTP 服务器初始化器
type HTTPServerInitializer struct {
	cfg              *config.Config
	logger           core.Logger
	dbInit           *DatabaseInitializer
	redisInit        *RedisInitializer
	sessionInit      *SessionServiceInitializer
	auditInit        *AuditServiceInitializer
	notificationInit *NotificationServiceInitializer
	forcedLogoutInit *ForcedLogoutServiceInitializer
	emailInit        *EmailClientInitializer
	server           *http.Server
	errChan          chan error // 服务器错误通道
}

// NewHTTPServerInitializer 创建 HTTP 服务器初始化器
func NewHTTPServerInitializer(
	cfg *config.Config,
	logger core.Logger,
	dbInit *DatabaseInitializer,
	redisInit *RedisInitializer,
	sessionInit *SessionServiceInitializer,
	auditInit *AuditServiceInitializer,
	notificationInit *NotificationServiceInitializer,
	forcedLogoutInit *ForcedLogoutServiceInitializer,
	emailInit *EmailClientInitializer,
) *HTTPServerInitializer {
	return &HTTPServerInitializer{
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
}

// Name 返回初始化器名称
func (h *HTTPServerInitializer) Name() string {
	return "http-server"
}

// Priority 返回初始化优先级
func (h *HTTPServerInitializer) Priority() int {
	return bootstrap.PriorityHTTP
}

// Initialize 执行初始化
func (h *HTTPServerInitializer) Initialize(ctx context.Context) error {
	h.logger.Infow("Initializing HTTP server",
		"host", h.cfg.Server.Host,
		"port", h.cfg.Server.Port,
	)

	// 创建 Gin 引擎
	gin.SetMode(h.cfg.Server.Mode)
	router := gin.Default()
	router.SetTrustedProxies(nil)

	// 创建 PostgresDB 包装器 (用于兼容现有handlers)
	dbConn := &storage.PostgresDB{DB: h.dbInit.DB()}

	// 初始化 services (user, role, permission services)
	userService := service.NewUserService(dbConn)
	roleService := service.NewRoleService(dbConn)
	permissionService := service.NewPermissionService(dbConn)

	// 初始化 handlers
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

	// 初始化 middleware
	jwtMiddleware := middleware.NewJWTMiddleware(h.cfg.JWT.Secret, h.sessionInit.Service())
	authMiddleware := middleware.NewForcedLogoutAuthMiddleware(h.dbInit.DB())
	rateLimiter := middleware.NewRateLimiter(h.redisInit.Client())

	// Register health check endpoint
	router.GET("/health", func(c *gin.Context) {
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
	v1 := router.Group("/api/v1/auth")
	{
		v1.POST("/login", authHandler.LoginHandler)
		v1.POST("/logout", jwtMiddleware.JWTAuth(), authHandler.LogoutHandler)
		v1.GET("/me", jwtMiddleware.JWTAuth(), authHandler.GetCurrentUserHandler)
		v1.GET("/codes", jwtMiddleware.JWTAuth(), authHandler.GetAccessCodesHandler)
	}

	// Register user management routes
	userRoutes := router.Group("/api/v1/users")
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
	roleRoutes := router.Group("/api/v1/roles")
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
	permissionRoutes := router.Group("/api/v1/permissions")
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
	forcedLogoutRoutes.RegisterRoutes(router)

	// Print registered routes
	h.logger.Infow("Auth Service Routes Registered",
		"health", "GET /health",
		"auth_login", "POST /api/v1/auth/login",
		"auth_logout", "POST /api/v1/auth/logout",
		"users", "GET/POST/PUT/DELETE /api/v1/users",
		"roles", "GET/POST/PUT/DELETE /api/v1/roles",
		"permissions", "GET/POST/PUT/DELETE /api/v1/permissions",
	)

	// 创建 HTTP 服务器
	addr := fmt.Sprintf("%s:%d", h.cfg.Server.Host, h.cfg.Server.Port)
	h.server = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// 创建错误通道用于捕获服务器启动错误
	h.errChan = make(chan error, 1)

	// 启动服务器 (在 goroutine 中)
	go func() {
		h.logger.Infow("Starting HTTP server", "addr", addr)
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.logger.Errorw("HTTP server fatal error", "error", err)
			h.errChan <- err
		}
	}()

	// 启动错误监听器
	go h.monitorServerErrors()

	h.logger.Infow("HTTP server initialized successfully")
	return nil
}

// monitorServerErrors 监听服务器致命错误
func (h *HTTPServerInitializer) monitorServerErrors() {
	for err := range h.errChan {
		if err != nil {
			h.logger.Fatalw("HTTP server encountered fatal error, shutting down",
				"error", err,
				"component", "http-server",
			)
		}
	}
}

// Close 关闭服务器
func (h *HTTPServerInitializer) Close(ctx context.Context) error {
	if h.server != nil {
		h.logger.Infow("Shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := h.server.Shutdown(shutdownCtx); err != nil {
			return err
		}
	}
	// 关闭错误通道
	if h.errChan != nil {
		close(h.errChan)
	}
	return nil
}
