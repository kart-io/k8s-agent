package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	commonoptions "github.com/kart-io/k8s-agent/common/options"
	commonserver "github.com/kart-io/k8s-agent/common/server"
	httpserver "github.com/kart-io/k8s-agent/common/server/http"
	"github.com/kart-io/k8s-agent/internal/monitor/handler"
	"github.com/kart-io/k8s-agent/internal/monitor/middleware"
	"github.com/kart-io/logger/core"
)

type Server struct {
	ginServer     commonserver.Server // 使用 common/server 的 Server 接口
	log           core.Logger
	jwtSecret     string
	metricsPort   int
	metricsServer *http.Server // Prometheus metrics 单独的服务器
}

type ServerConfig struct {
	Port         int
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	JWTSecret    string
	MetricsPort  int
}

// NewServer 创建新的服务器实例，使用 common/server 框架
func NewServer(config *ServerConfig, metricsHandler *handler.MetricsHandler, logger core.Logger) *Server {
	// 创建 common/server 的配置
	serverOpts := &commonoptions.ServerOptions{
		Host:         "",
		Port:         config.Port,
		Mode:         config.Mode,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// 创建 Gin 服务器配置
	ginConfig := httpserver.NewGinServerConfig(serverOpts)

	// 创建 Gin 服务器
	ginServer := httpserver.NewGinServerFromFullConfig(logger, ginConfig)
	engine := ginServer.GetEngine()

	// 设置额外的中间件（除了 common/server 已经添加的）
	engine.Use(gin.Recovery())
	engine.Use(middleware.Logger(logger))
	engine.Use(middleware.CORS())

	s := &Server{
		ginServer:   ginServer,
		log:         logger,
		jwtSecret:   config.JWTSecret,
		metricsPort: config.MetricsPort,
	}

	// 设置路由
	s.setupRoutes(engine, metricsHandler)

	return s
}

func (s *Server) setupRoutes(engine *gin.Engine, metricsHandler *handler.MetricsHandler) {
	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now(),
		})
	})

	// API v1
	v1 := engine.Group("/api/v1")
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

	if s.log != nil {
		s.log.Infow("Monitor service routes configured",
			"jwt_enabled", s.jwtSecret != "",
			"metrics_port", s.metricsPort,
		)
	}
}

// Start 启动服务器 - 使用 common/server 框架
// 注意：这个方法现在由 initializer 或 bootstrap 调用
func (s *Server) Start() error {
	// 启动 Prometheus metrics 服务器
	if s.metricsPort > 0 {
		go s.startMetricsServer()
	}

	// 现在由 common/server 框架处理
	// 这个方法主要是为了向后兼容
	return s.Run(context.Background())
}

// Run 运行服务器 - 使用 common/server 框架
func (s *Server) Run(ctx context.Context) error {
	// 启动 Prometheus metrics 服务器
	if s.metricsPort > 0 {
		go s.startMetricsServer()
	}

	if s.log != nil {
		s.log.Infow("Starting monitor service with common/server framework")
	}

	// 使用 common/server 的 Serve 方法来管理生命周期
	return commonserver.Serve(ctx, s.ginServer, s.log)
}

// startMetricsServer 启动 Prometheus metrics 服务器（独立端口）
func (s *Server) startMetricsServer() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	s.metricsServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.metricsPort),
		Handler: mux,
	}

	s.log.Infow("Starting Prometheus metrics server", "port", s.metricsPort)
	if err := s.metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.log.Errorw("Prometheus metrics server failed", "error", err)
	}
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown(ctx context.Context) error {
	if s.log != nil {
		s.log.Info("Monitor service shutting down")
	}

	// 关闭 Prometheus metrics 服务器
	if s.metricsServer != nil {
		if err := s.metricsServer.Shutdown(ctx); err != nil {
			s.log.Errorw("Failed to shutdown metrics server", "error", err)
		}
	}

	// common/server 的 Serve 方法会自动处理主服务器的优雅关闭
	// 这里主要是为了处理额外的 metrics 服务器
	return nil
}

// GetServer 返回底层的 common/server.Server 实例
func (s *Server) GetServer() commonserver.Server {
	return s.ginServer
}

// GetEngine 返回 Gin Engine 实例（用于测试或其他需要）
func (s *Server) GetEngine() *gin.Engine {
	if ginSrv, ok := s.ginServer.(*httpserver.GinServer); ok {
		return ginSrv.GetEngine()
	}
	return nil
}
