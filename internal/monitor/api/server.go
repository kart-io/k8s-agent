package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/kart-io/k8s-agent/internal/monitor/handler"
	"github.com/kart-io/k8s-agent/internal/monitor/middleware"
	"github.com/kart-io/logger/core"
)

type Server struct {
	router      *gin.Engine
	httpServer  *http.Server
	log         core.Logger
	jwtSecret   string
	metricsPort int
}

type ServerConfig struct {
	Port         int
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	JWTSecret    string
	MetricsPort  int
}

func NewServer(config *ServerConfig, metricsHandler *handler.MetricsHandler, logger core.Logger) *Server {
	gin.SetMode(config.Mode)
	router := gin.New()

	// 全局中间件
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())

	s := &Server{
		router:      router,
		log:         logger,
		jwtSecret:   config.JWTSecret,
		metricsPort: config.MetricsPort,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", config.Port),
			Handler:      router,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
		},
	}

	s.setupRoutes(metricsHandler)
	return s
}

func (s *Server) setupRoutes(metricsHandler *handler.MetricsHandler) {
	// 健康检查
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now(),
		})
	})

	// API v1
	v1 := s.router.Group("/api/v1")
	{
		// 认证保护的路由
		authRoutes := v1.Group("")
		authRoutes.Use(middleware.AuthMiddleware(s.jwtSecret))
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
}

// Start 启动服务器
func (s *Server) Start() error {
	// 启动 Prometheus metrics 服务器
	if s.metricsPort > 0 {
		go s.startMetricsServer()
	}

	s.log.Infow("Starting monitor service", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

// Shutdown 优雅关闭
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
}

// startMetricsServer 启动 Prometheus 指标服务器
func (s *Server) startMetricsServer() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.metricsPort),
		Handler: mux,
	}

	s.log.Infow("Starting Prometheus metrics server", "port", s.metricsPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.log.Errorw("Metrics server failed", "error", err)
	}
}
