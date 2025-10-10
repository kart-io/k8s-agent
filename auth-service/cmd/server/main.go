package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/auth-service/internal/handler"
	"github.com/kart-io/k8s-agent/auth-service/internal/middleware"
	"github.com/kart-io/k8s-agent/auth-service/internal/routes"
	"github.com/kart-io/k8s-agent/auth-service/pkg/forced-logout/audit"
	"github.com/kart-io/k8s-agent/auth-service/pkg/forced-logout/notification"
	"github.com/kart-io/k8s-agent/auth-service/pkg/forced-logout/session"
	"github.com/kart-io/notifyhub/pkg/config"
	"github.com/kart-io/notifyhub/pkg/notifyhub"
	"github.com/kart-io/notifyhub/pkg/utils/logger"
	forcedlogout "github.com/kart-io/k8s-agent/auth-service/pkg/forced-logout"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration (simplified - in production, use proper config management)
	cfg := loadConfig()

	// Initialize database connection
	db, err := initDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	// Initialize repositories
	sessionRepo := session.NewRedisRepository(redisClient, cfg.JWTExpiresHours)
	auditRepo := audit.NewPostgresRepository(db)
	notificationRepo := notification.NewPostgresRepository(db)

	// Initialize services
	sessionService := session.NewService(sessionRepo)
	auditService := audit.NewService(auditRepo)

	// Initialize NotifyHub client for email notifications
	notifyHubLogger := logger.New().LogMode(logger.Info)

	var notifyHubClient notifyhub.Client
	if cfg.EmailEnabled {
		// Configure email platform using NotifyHub
		emailConfig := config.EmailConfig{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUser,
			Password: cfg.SMTPPassword,
			From:     cfg.EmailFromAddress,
			UseTLS:   true,
			Timeout:  30 * time.Second,
		}

		notifyHubClient, err = notifyhub.NewClientFromOptions(
			config.WithEmail(emailConfig),
			config.WithLogger(notifyHubLogger),
			config.WithTimeout(30*time.Second),
		)
		if err != nil {
			log.Fatalf("Failed to initialize NotifyHub client: %v", err)
		}
		log.Printf("✅ NotifyHub client initialized with email platform (%s:%d)\n", cfg.SMTPHost, cfg.SMTPPort)
	} else {
		// Create a test client with no-op configuration
		notifyHubClient, err = notifyhub.NewClientFromOptions(
			config.WithTestDefaults(),
			config.WithLogger(notifyHubLogger),
		)
		if err != nil {
			log.Fatalf("Failed to initialize test NotifyHub client: %v", err)
		}
		log.Println("⚠️  Email notifications disabled - using test mode")
	}

	// Initialize notification template engine
	templateEngine, err := notification.NewTemplateEngine(cfg.EmailTemplateDir)
	if err != nil {
		log.Fatalf("Failed to initialize template engine: %v", err)
	}

	// Initialize notification service with NotifyHub
	notificationService := notification.NewService(notificationRepo, notifyHubClient, templateEngine)

	// Initialize forced logout service
	forcedLogoutService := forcedlogout.NewService(
		sessionService,
		auditService,
		notificationService,
		cfg.LoginURL,
	)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(db, cfg.JWTSecret, cfg.JWTExpiresHours, sessionService)
	sessionHandler := handler.NewSessionHandler(sessionService)
	forcedLogoutHandler := handler.NewForcedLogoutHandler(forcedLogoutService)
	auditHandler := handler.NewAuditHandler(auditService)

	// Initialize middleware
	jwtMiddleware := middleware.NewJWTMiddleware(cfg.JWTSecret, sessionService)
	authMiddleware := middleware.NewForcedLogoutAuthMiddleware(db)
	rateLimiter := middleware.NewRateLimiter(redisClient)

	// Initialize Gin router
	router := gin.Default()

	// Register health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":              "healthy",
			"service":             "auth-service",
			"version":             "1.0.0",
			"email_notifications": cfg.EmailEnabled,
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
	fmt.Println("\n=== Auth Service Started ===")
	fmt.Println("Listening on:", cfg.ServerAddr)
	fmt.Println("NotifyHub Integration: ✅ Enabled")
	if cfg.EmailEnabled {
		fmt.Printf("Email Platform: %s:%d (from: %s)\n", cfg.SMTPHost, cfg.SMTPPort, cfg.EmailFromAddress)
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
	if err := router.Run(cfg.ServerAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Config holds application configuration
type Config struct {
	ServerAddr       string
	DatabaseURL      string
	RedisAddr        string
	RedisPassword    string
	JWTSecret        string
	JWTExpiresHours  int
	EmailEnabled     bool
	SMTPHost         string
	SMTPPort         int
	SMTPUser         string
	SMTPPassword     string
	EmailFromAddress string
	EmailFromName    string
	EmailTemplateDir string
	LoginURL         string
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	return Config{
		ServerAddr:       getEnv("SERVER_ADDR", ":8090"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/k8s_agent_auth?sslmode=disable"),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		JWTSecret:        getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTExpiresHours:  24,
		EmailEnabled:     getEnv("EMAIL_ENABLED", "false") == "true",
		SMTPHost:         getEnv("SMTP_HOST", "smtp.example.com"),
		SMTPPort:         587,
		SMTPUser:         getEnv("SMTP_USER", ""),
		SMTPPassword:     getEnv("SMTP_PASSWORD", ""),
		EmailFromAddress: getEnv("EMAIL_FROM_ADDRESS", "noreply@k8s-agent.com"),
		EmailFromName:    getEnv("EMAIL_FROM_NAME", "K8s Agent Security"),
		EmailTemplateDir: getEnv("EMAIL_TEMPLATE_DIR", "templates/email"),
		LoginURL:         getEnv("LOGIN_URL", "http://localhost:8090/login"),
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// initDatabase initializes database connection
func initDatabase(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}
