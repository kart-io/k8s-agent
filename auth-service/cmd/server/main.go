package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/auth-service/internal/config"
	"github.com/kart-io/k8s-agent/auth-service/internal/handler"
	"github.com/kart-io/k8s-agent/auth-service/internal/middleware"
	"github.com/kart-io/k8s-agent/auth-service/internal/routes"
	"github.com/kart-io/k8s-agent/auth-service/internal/storage"
	forcedlogout "github.com/kart-io/k8s-agent/auth-service/pkg/forced-logout"
	"github.com/kart-io/k8s-agent/auth-service/pkg/forced-logout/audit"
	"github.com/kart-io/k8s-agent/auth-service/pkg/forced-logout/notification"
	"github.com/kart-io/k8s-agent/auth-service/pkg/forced-logout/session"
	notifyhubconfig "github.com/kart-io/notifyhub/pkg/config"
	"github.com/kart-io/notifyhub/pkg/notifyhub"
	"github.com/kart-io/notifyhub/pkg/utils/logger"
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

	// Initialize NotifyHub client for email notifications
	notifyHubLogger := logger.New().LogMode(logger.Info)

	var notifyHubClient notifyhub.Client
	if cfg.Email.Enabled {
		// Configure email platform using NotifyHub
		emailConfig := notifyhubconfig.EmailConfig{
			Host:     cfg.Email.SMTPHost,
			Port:     cfg.Email.SMTPPort,
			Username: cfg.Email.SMTPUser,
			Password: cfg.Email.SMTPPassword,
			From:     cfg.Email.FromAddress,
			UseTLS:   true,
			Timeout:  30 * time.Second,
		}

		notifyHubClient, err = notifyhub.NewClientFromOptions(
			notifyhubconfig.WithEmail(emailConfig),
			notifyhubconfig.WithLogger(notifyHubLogger),
			notifyhubconfig.WithTimeout(30*time.Second),
		)
		if err != nil {
			log.Fatalf("Failed to initialize NotifyHub client: %v", err)
		}
		log.Printf("✅ NotifyHub client initialized with email platform (%s:%d)\n", cfg.Email.SMTPHost, cfg.Email.SMTPPort)
	} else {
		// Create a test client with no-op configuration
		notifyHubClient, err = notifyhub.NewClientFromOptions(
			notifyhubconfig.WithTestDefaults(),
			notifyhubconfig.WithLogger(notifyHubLogger),
		)
		if err != nil {
			log.Fatalf("Failed to initialize test NotifyHub client: %v", err)
		}
		log.Println("⚠️  Email notifications disabled - using test mode")
	}

	// Initialize notification template engine
	templateEngine, err := notification.NewTemplateEngine(cfg.Email.TemplateDir)
	if err != nil {
		log.Fatalf("Failed to initialize template engine: %v", err)
	}

	// Initialize notification service with NotifyHub
	notificationService := notification.NewService(notificationRepo, notifyHubClient, templateEngine)

	// Determine login URL (use environment variable or default)
	loginURL := "http://localhost:8090/login"

	// Initialize forced logout service
	forcedLogoutService := forcedlogout.NewService(
		sessionService,
		auditService,
		notificationService,
		loginURL,
	)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(db, cfg.JWT.Secret, cfg.JWT.ExpiresHours, sessionService)
	sessionHandler := handler.NewSessionHandler(sessionService)
	forcedLogoutHandler := handler.NewForcedLogoutHandler(forcedLogoutService)
	auditHandler := handler.NewAuditHandler(auditService)

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
	fmt.Println("NotifyHub Integration: ✅ Enabled")
	if cfg.Email.Enabled {
		fmt.Printf("Email Platform: %s:%d (from: %s)\n", cfg.Email.SMTPHost, cfg.Email.SMTPPort, cfg.Email.FromAddress)
	} else {
		fmt.Println("Email Platform: Disabled (test mode)")
	}
	fmt.Println("\nRegistered Routes:")
	fmt.Println("GET    /health                              - Health check")
	fmt.Println("POST   /api/v1/auth/login                   - User login")
	fmt.Println("POST   /api/v1/auth/logout                  - User logout")
	fmt.Println("GET    /api/v1/auth/me                      - Get current user")
	for _, route := range forcedLogoutRoutes.PrintRegisteredRoutes() {
		fmt.Println(route)
	}
	fmt.Println("============================")

	// Start server
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
