package initializers

import (
        commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"context"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	
	"github.com/kart-io/k8s-agent/internal/monitor/handler"
	monitormiddleware "github.com/kart-io/k8s-agent/internal/monitor/middleware"
	"github.com/kart-io/k8s-agent/internal/monitor/service"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// HTTPServerInitializer wraps the common HTTP server initializer.
type HTTPServerInitializer struct {
	*pkginitializers.HTTPServerInitializer
	cfg           *commonapp.StandardOptions
	logger        core.Logger
	dbInit        *DatabaseInitializer
	redisInit     *RedisInitializer
	monitorSvc    *service.MonitorService
}

// NewHTTPServerInitializer creates a new HTTP server initializer.
func NewHTTPServerInitializer(
	cfg *commonapp.StandardOptions,
	logger core.Logger,
	dbInit *DatabaseInitializer,
	redisInit *RedisInitializer,
) *HTTPServerInitializer {
	h := &HTTPServerInitializer{
		cfg:       cfg,
		logger:    logger,
		dbInit:    dbInit,
		redisInit: redisInit,
	}

	// Create standard HTTP server config
	serverConfig := &pkginitializers.HTTPServerConfig{
		Name:       "monitor-http-server",
		Priority:   bootstrap.PriorityHTTP,
		Config:     cfg.Server,
		RouteSetup: h.setupRoutes,
	}

	// Create standard HTTP server initializer
	h.HTTPServerInitializer = pkginitializers.NewHTTPServerInitializer(serverConfig, logger)
	return h
}

// Initialize initializes the HTTP server and creates monitor service.
func (h *HTTPServerInitializer) Initialize(ctx context.Context) error {
	// Initialize monitor service
	h.monitorSvc = service.NewMonitorService(
		h.dbInit.Storage(),
		h.redisInit.Storage(),
		h.logger,
	)

	// Initialize HTTP server (creates server and calls setupRoutes)
	return h.HTTPServerInitializer.Initialize(ctx)
}

// setupRoutes configures all routes for the monitor service.
func (h *HTTPServerInitializer) setupRoutes(engine *gin.Engine) error {
	h.logger.Infow("Setting up monitor service routes")

	// Add monitor-specific middleware
	engine.Use(monitormiddleware.Logger(h.logger))
	engine.Use(monitormiddleware.CORS())

	// Initialize metrics handler
	metricsHandler := handler.NewMetricsHandler(h.monitorSvc, h.logger)

	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := engine.Group("/api/v1")
	{
		// 认证保护的路由 (简化版 - 实际应该使用 JWT middleware)
		authRoutes := v1.Group("")
		// authRoutes.Use(middleware.AuthMiddleware(h.cfg.JWT.Secret))
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
	if h.cfg.Prometheus.Enabled {
		engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
		h.logger.Infow("Prometheus metrics enabled",
			"endpoint", "/metrics",
		)
	}

	h.logger.Infow("Monitor service routes registered",
		"health", "GET /health",
		"metrics_summary", "GET /api/v1/metrics/summary",
		"metrics_agents", "GET /api/v1/metrics/agents",
		"metrics_trends", "GET /api/v1/metrics/trends",
		"prometheus_metrics", h.cfg.Prometheus.Enabled,
	)

	return nil
}

