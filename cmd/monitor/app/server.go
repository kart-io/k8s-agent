package app

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/kart-io/k8s-agent/cmd/monitor/app/options"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	commonserver "github.com/kart-io/k8s-agent/common/server"
	httpserver "github.com/kart-io/k8s-agent/common/server/http"
	"github.com/kart-io/k8s-agent/internal/monitor/handler"
	monitormiddleware "github.com/kart-io/k8s-agent/internal/monitor/middleware"
	"github.com/kart-io/k8s-agent/internal/monitor/service"
	"github.com/kart-io/k8s-agent/internal/monitor/storage"
	"github.com/kart-io/logger/core"
)

// MonitorService represents the monitor service using common/server.
type MonitorService struct {
	opts           *options.Options
	log            core.Logger
	pgStorage      *storage.PostgresStorage
	redisStorage   *storage.RedisStorage
	monitorService *service.MonitorService
	server         commonserver.Server
}

// NewServer creates a new monitor service (使用 common/server).
func NewServer(opts *options.Options, log core.Logger) (*MonitorService, error) {
	srv := &MonitorService{
		opts: opts,
		log:  log,
	}

	if err := srv.initialize(); err != nil {
		return nil, err
	}

	return srv, nil
}

// initialize initializes all server components.
func (s *MonitorService) initialize() error {
	var err error

	// Initialize PostgreSQL storage
	dbOpts := &commonoptions.DatabaseOptions{
		Host:         s.opts.Database.Host,
		Port:         s.opts.Database.Port,
		User:         s.opts.Database.User,
		Password:     s.opts.Database.Password,
		Database:     s.opts.Database.DBName,
		MaxOpenConns: s.opts.Database.MaxOpenConns,
		MaxIdleConns: s.opts.Database.MaxIdleConns,
	}
	s.pgStorage, err = storage.NewPostgresStorage(dbOpts, s.log)
	if err != nil {
		return fmt.Errorf("failed to initialize PostgreSQL storage: %w", err)
	}

	// Initialize Redis storage
	s.redisStorage, err = storage.NewRedisStorage(&storage.RedisConfig{
		Host:     s.opts.Redis.Host,
		Port:     s.opts.Redis.Port,
		Password: s.opts.Redis.Password,
		DB:       s.opts.Redis.DB,
		PoolSize: s.opts.Redis.PoolSize,
	}, s.log)
	if err != nil {
		if closeErr := s.pgStorage.Close(); closeErr != nil {
			s.log.Errorw("Failed to close PostgreSQL storage", "error", closeErr)
		}
		return fmt.Errorf("failed to initialize Redis storage: %w", err)
	}

	// Initialize monitor service
	s.monitorService = service.NewMonitorService(s.pgStorage, s.redisStorage, s.log)

	// Initialize metrics handler
	metricsHandler := handler.NewMetricsHandler(s.monitorService, s.log)

	// Parse timeout durations
	readTimeout, _ := time.ParseDuration(s.opts.Server.ReadTimeout)
	writeTimeout, _ := time.ParseDuration(s.opts.Server.WriteTimeout)

	// Create Gin server config using common/server
	ginConfig := httpserver.NewGinServerConfig(&commonoptions.ServerOptions{
		Host:         "",
		Port:         s.opts.Server.Port,
		Mode:         s.opts.Server.Mode,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	})

	// Create Gin server
	ginServer := httpserver.NewGinServerFromFullConfig(s.log, ginConfig)
	engine := ginServer.GetEngine()

	// Add monitor-specific middleware
	engine.Use(gin.Recovery())
	engine.Use(monitormiddleware.Logger(s.log))
	engine.Use(monitormiddleware.CORS())

	// Setup routes
	s.setupRoutes(engine, metricsHandler)

	s.server = ginServer

	return nil
}

// setupRoutes sets up all API routes.
func (s *MonitorService) setupRoutes(engine *gin.Engine, metricsHandler *handler.MetricsHandler) {
	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := engine.Group("/api/v1")
	{
		// 认证保护的路由 (简化版 - 实际应该使用 JWT middleware)
		authRoutes := v1.Group("")
		// authRoutes.Use(middleware.AuthMiddleware(s.opts.JWT.Secret))
		{
			// 监控指标
			metrics := authRoutes.Group("/metrics")
			{
				metrics.GET("/summary", metricsHandler.GetSummary)
				metrics.GET("/agents", metricsHandler.GetAgentMetrics)
				metrics.GET("/trends", metricsHandler.GetTrends)
			}
		}
	}

	// Prometheus metrics endpoint (if enabled)
	if s.opts.Prometheus.Enabled {
		engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
		s.log.Infow("Prometheus metrics enabled",
			"endpoint", "/metrics",
		)
	}

	s.log.Infow("Monitor API routes configured")
}

// Run starts the monitor service using common/server.Serve().
func (s *MonitorService) Run(ctx context.Context) error {
	s.log.Infow("Starting Monitor Service",
		"port", s.opts.Server.Port,
		"mode", s.opts.Server.Mode,
	)

	// 使用 common/server 的标准 Serve 方法
	// 它会自动处理信号和优雅关停
	return commonserver.Serve(ctx, s.server, s.log)
}

// GetServer returns the server instance (实现 ServerProvider 接口).
func (s *MonitorService) GetServer() commonserver.Server {
	return s.server
}

// Cleanup cleans up resources.
func (s *MonitorService) Cleanup() error {
	var lastErr error

	if s.redisStorage != nil {
		if err := s.redisStorage.Close(); err != nil {
			s.log.Errorw("Failed to close Redis storage", "error", err)
			lastErr = err
		}
	}

	if s.pgStorage != nil {
		if err := s.pgStorage.Close(); err != nil {
			s.log.Errorw("Failed to close PostgreSQL storage", "error", err)
			lastErr = err
		}
	}

	s.log.Info("Monitor service resources cleaned up")
	return lastErr
}
