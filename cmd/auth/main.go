package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/internal/auth/config"
	"github.com/kart-io/k8s-agent/internal/auth/email"
	"github.com/kart-io/k8s-agent/internal/auth/handler"
	"github.com/kart-io/k8s-agent/internal/auth/middleware"
	"github.com/kart-io/k8s-agent/internal/auth/routes"
	"github.com/kart-io/k8s-agent/internal/auth/service"
	"github.com/kart-io/k8s-agent/internal/auth/storage"
	forcedlogout "github.com/kart-io/k8s-agent/internal/auth/forced-logout"
	"github.com/kart-io/k8s-agent/internal/auth/forced-logout/audit"
	"github.com/kart-io/k8s-agent/internal/auth/forced-logout/notification"
	"github.com/kart-io/k8s-agent/internal/auth/forced-logout/session"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Parse command-line flags
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to configuration file (defaults to ./configs/config.yaml)")
	flag.StringVar(&configPath, "c", "", "Path to configuration file (shorthand)")
	flag.Parse()

	// Load configuration from config file
	var cfg *config.Config
	var err error

	if configPath != "" {
		cfg, err = config.LoadFromPath(configPath)
		if err != nil {
			log.Fatalf("Failed to load configuration from %s: %v", configPath, err)
		}
		log.Printf("Loaded configuration from: %s", configPath)
	} else {
		cfg, err = config.Load()
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
	}

	// Initialize database connection using internal/storage
	dbConn, err := storage.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	db := dbConn.DB

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	// Initialize repositories
	sessionRepo := session.NewRedisRepository(redisClient, cfg.JWT.ExpiresHours)
	auditRepo := audit.NewPostgresRepository(db)
	notificationRepo := notification.NewPostgresRepository(db)

	// Initialize services
	sessionService := session.NewService(sessionRepo)
	auditService := audit.NewService(auditRepo)

	// Initialize email client
	var emailClient email.Client
	if cfg.Email.Enabled {
		// Create email client with SMTP config
		emailConfig := &email.Config{
			Host:     cfg.Email.SMTPHost,
			Port:     cfg.Email.SMTPPort,
			Username: cfg.Email.SMTPUser,
			Password: cfg.Email.SMTPPassword,
			From:     cfg.Email.FromAddress,
			UseTLS:   true,
			Timeout:  30 * time.Second,
		}

		var err error
		emailClient, err = email.NewClient(emailConfig)
		if err != nil {
			log.Fatalf("Failed to initialize email client: %v", err)
		}
		log.Printf("✅ Email client initialized (%s:%d)\n", cfg.Email.SMTPHost, cfg.Email.SMTPPort)
	} else {
		// Create a no-op client
		emailClient, _ = email.NewClient(nil)
		log.Println("⚠️  Email notifications disabled - using no-op mode")
	}

	// Initialize notification template engine
	templateEngine, err := notification.NewTemplateEngine(cfg.Email.TemplateDir)
	if err != nil {
		log.Fatalf("Failed to initialize template engine: %v", err)
	}

	// Initialize notification service
	notificationService := notification.NewService(notificationRepo, emailClient, templateEngine)

	// Determine login URL (use environment variable or default)
	loginURL := "http://localhost:8090/login"

	// Initialize forced logout service
	forcedLogoutService := forcedlogout.NewService(
		sessionService,
		auditService,
		notificationService,
		loginURL,
	)

	// Initialize services
	userService := service.NewUserService(dbConn)
	roleService := service.NewRoleService(dbConn)
	permissionService := service.NewPermissionService(dbConn)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(db, cfg.JWT.Secret, cfg.JWT.ExpiresHours, sessionService)
	sessionHandler := handler.NewSessionHandler(sessionService)
	forcedLogoutHandler := handler.NewForcedLogoutHandler(forcedLogoutService)
	auditHandler := handler.NewAuditHandler(auditService)
	userHandler := handler.NewUserHandler(userService)
	roleHandler := handler.NewRoleHandler(roleService)
	permissionHandler := handler.NewPermissionHandler(permissionService)

	// Initialize middleware
	jwtMiddleware := middleware.NewJWTMiddleware(cfg.JWT.Secret, sessionService)
	authMiddleware := middleware.NewForcedLogoutAuthMiddleware(db)
	rateLimiter := middleware.NewRateLimiter(redisClient)

	// Initialize Gin router
	router := gin.Default()
	router.SetTrustedProxies(nil) // Don't trust any proxies in development

	// Or for production with specific proxies:
	// router.SetTrustedProxies([]string{"127.0.0.1", "10.0.0.0/8"})

	// Set Gin mode based on config
	gin.SetMode(cfg.Server.Mode)

	// Register health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":              "healthy",
			"service":             "auth-service",
			"version":             "1.0.0",
			"email_notifications": cfg.Email.Enabled,
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
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Println("\n=== Auth Service Started ===")
	fmt.Println("Listening on:", serverAddr)
	fmt.Println("Email Integration: ✅ Enabled")
	if cfg.Email.Enabled {
		fmt.Printf("Email Platform: %s:%d (from: %s)\n", cfg.Email.SMTPHost, cfg.Email.SMTPPort, cfg.Email.FromAddress)
	} else {
		fmt.Println("Email Platform: Disabled (no-op mode)")
	}
	fmt.Println("\nRegistered Routes:")
	fmt.Println("GET    /health                              - Health check")
	fmt.Println("\nAuthentication:")
	fmt.Println("POST   /api/v1/auth/login                   - User login")
	fmt.Println("POST   /api/v1/auth/logout                  - User logout")
	fmt.Println("GET    /api/v1/auth/me                      - Get current user")
	fmt.Println("GET    /api/v1/auth/codes                   - Get access permission codes")
	fmt.Println("\nUser Management:")
	fmt.Println("GET    /api/v1/users                        - List users")
	fmt.Println("GET    /api/v1/users/:id                    - Get user by ID")
	fmt.Println("POST   /api/v1/users                        - Create user")
	fmt.Println("PUT    /api/v1/users/:id                    - Update user")
	fmt.Println("DELETE /api/v1/users/:id                    - Delete user")
	fmt.Println("POST   /api/v1/users/:id/roles              - Assign roles to user")
	fmt.Println("\nRole Management:")
	fmt.Println("GET    /api/v1/roles                        - List roles")
	fmt.Println("GET    /api/v1/roles/:id                    - Get role by ID")
	fmt.Println("POST   /api/v1/roles                        - Create role")
	fmt.Println("PUT    /api/v1/roles/:id                    - Update role")
	fmt.Println("DELETE /api/v1/roles/:id                    - Delete role")
	fmt.Println("POST   /api/v1/roles/:id/permissions        - Assign permissions to role")
	fmt.Println("GET    /api/v1/roles/:id/permissions        - Get role permissions")
	fmt.Println("\nPermission Management:")
	fmt.Println("GET    /api/v1/permissions                  - List permissions")
	fmt.Println("GET    /api/v1/permissions/tree             - Get permission tree")
	fmt.Println("GET    /api/v1/permissions/:id              - Get permission by ID")
	fmt.Println("POST   /api/v1/permissions                  - Create permission")
	fmt.Println("PUT    /api/v1/permissions/:id              - Update permission")
	fmt.Println("DELETE /api/v1/permissions/:id              - Delete permission")
	fmt.Println("\nForced Logout & Session Management:")
	for _, route := range forcedLogoutRoutes.PrintRegisteredRoutes() {
		fmt.Println(route)
	}
	fmt.Println("============================")

	// Start server
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
