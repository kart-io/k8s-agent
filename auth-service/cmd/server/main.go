package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/auth-service/internal/config"
	"github.com/kart-io/k8s-agent/auth-service/internal/handler"
	"github.com/kart-io/k8s-agent/auth-service/internal/middleware"
	"github.com/kart-io/k8s-agent/auth-service/internal/service"
	"github.com/kart-io/k8s-agent/auth-service/internal/storage"
	"github.com/kart-io/k8s-agent/auth-service/pkg/logger"
	"github.com/kart-io/k8s-agent/auth-service/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(&cfg.Logging); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	log := logger.GetLogger()
	log.Infow("Starting auth-service...")

	// Connect to PostgreSQL
	db, err := storage.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Info("Database connected successfully")

	// Connect to Redis
	redis, err := storage.NewRedisClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()
	log.Info("Redis connected successfully")

	// NOTE: Database migrations are now handled automatically by GORM AutoMigrate
	// in postgres.go if cfg.Database.AutoMigrate is true

	// Initialize services
	authService := service.NewAuthService(db, redis, cfg)
	userService := service.NewUserService(db)
	roleService := service.NewRoleService(db)
	permissionService := service.NewPermissionService(db)
	apikeyService := service.NewAPIKeyService(db)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	roleHandler := handler.NewRoleHandler(roleService)
	permissionHandler := handler.NewPermissionHandler(permissionService)
	apikeyHandler := handler.NewAPIKeyHandler(apikeyService)

	// Setup Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// Apply global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.RequestLogger())
	router.Use(metrics.PrometheusMiddleware())

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Public auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/check", authHandler.CheckPermission)
		}

		// Protected auth routes (require JWT)
		authProtected := v1.Group("/auth")
		authProtected.Use(middleware.JWTAuth(cfg.JWT.Secret, redis))
		{
			authProtected.POST("/logout", authHandler.Logout)
			authProtected.GET("/me", authHandler.Me)
			authProtected.GET("/menus", authHandler.Menus)
		}

		// User management routes (require JWT + permissions)
		users := v1.Group("/users")
		users.Use(middleware.JWTAuth(cfg.JWT.Secret, redis))
		{
			users.GET("", middleware.RequirePermission(db, "user:list"), userHandler.List)
			users.GET("/:id", middleware.RequirePermission(db, "user:list"), userHandler.GetByID)
			users.POST("", middleware.RequirePermission(db, "user:create"), userHandler.Create)
			users.PUT("/:id", middleware.RequirePermission(db, "user:update"), userHandler.Update)
			users.DELETE("/:id", middleware.RequirePermission(db, "user:delete"), userHandler.Delete)
			users.POST("/:id/roles", middleware.RequirePermission(db, "user:update"), userHandler.AssignRoles)
		}

		// Role management routes (require JWT + permissions)
		roles := v1.Group("/roles")
		roles.Use(middleware.JWTAuth(cfg.JWT.Secret, redis))
		{
			roles.GET("", middleware.RequirePermission(db, "role:list"), roleHandler.List)
			roles.GET("/:id", middleware.RequirePermission(db, "role:view"), roleHandler.GetByID)
			roles.POST("", middleware.RequirePermission(db, "role:create"), roleHandler.Create)
			roles.PUT("/:id", middleware.RequirePermission(db, "role:update"), roleHandler.Update)
			roles.DELETE("/:id", middleware.RequirePermission(db, "role:delete"), roleHandler.Delete)
			roles.POST("/:id/permissions", middleware.RequirePermission(db, "role:assign"), roleHandler.AssignPermissions)
			roles.GET("/:id/permissions", middleware.RequirePermission(db, "role:view"), roleHandler.GetPermissions)
		}

		// Permission management routes (require JWT + permissions)
		permissions := v1.Group("/permissions")
		permissions.Use(middleware.JWTAuth(cfg.JWT.Secret, redis))
		{
			permissions.GET("", middleware.RequirePermission(db, "permission:list"), permissionHandler.List)
			permissions.GET("/tree", middleware.RequirePermission(db, "permission:list"), permissionHandler.GetTree)
			permissions.GET("/:id", middleware.RequirePermission(db, "permission:view"), permissionHandler.GetByID)
			permissions.POST("", middleware.RequirePermission(db, "permission:create"), permissionHandler.Create)
			permissions.PUT("/:id", middleware.RequirePermission(db, "permission:update"), permissionHandler.Update)
			permissions.DELETE("/:id", middleware.RequirePermission(db, "permission:delete"), permissionHandler.Delete)
		}

		// API Key management routes (require JWT + permissions)
		apikeys := v1.Group("/api-keys")
		apikeys.Use(middleware.JWTAuth(cfg.JWT.Secret, redis))
		{
			apikeys.GET("", middleware.RequirePermission(db, "apikey:list"), apikeyHandler.List)
			apikeys.POST("", middleware.RequirePermission(db, "apikey:create"), apikeyHandler.Create)
			apikeys.DELETE("/:id", middleware.RequirePermission(db, "apikey:delete"), apikeyHandler.Delete)
		}
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		// Check database connection
		dbStatus := "ok"
		if err := db.Health(); err != nil {
			dbStatus = "error"
			log.Warnf("Database health check failed: %v", err)
		}

		// Check Redis connection
		redisStatus := "ok"
		if err := redis.Client.Ping(context.Background()).Err(); err != nil {
			redisStatus = "error"
			log.Warnf("Redis health check failed: %v", err)
		}

		// Overall status
		overallStatus := "ok"
		if dbStatus != "ok" || redisStatus != "ok" {
			overallStatus = "degraded"
		}

		statusCode := http.StatusOK
		if overallStatus == "degraded" {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, gin.H{
			"status": overallStatus,
			"time":   time.Now().Format(time.RFC3339),
			"checks": gin.H{
				"database": dbStatus,
				"redis":    redisStatus,
			},
		})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Infof("Server started on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited")
}
