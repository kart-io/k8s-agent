// Copyright 2025 Kart Project. All rights reserved.
// Gin HTTP server implementation

package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/middleware"
	"github.com/kart-io/logger/core"
)

// GinServer Gin HTTP 服务器
// 实现 Server 接口，通过 Bootstrap 系统管理生命周期
type GinServer struct {
	Engine *gin.Engine
	server *http.Server
	logger core.Logger
}

// NewGinServerFromFullConfig 从完整配置创建 Gin 服务器
// 这是创建 GinServer 的唯一入口函数
//
// 使用示例:
//
//	config := server.NewGinServerOptions(serverOpts).
//	    WithCORS(corsOpts).
//	    WithJWT(jwtOpts).
//	    WithRateLimit(rateLimitOpts)
//	ginSrv := server.NewGinServerFromFullConfig(log, config)
//
// 然后通过 HTTPServerInitializer 注册到 Bootstrap:
//
//	httpInit := initializers.NewHTTPServerInitializer(config, logger)
//	bootstrap.Register(httpInit)
func NewGinServerFromFullConfig(log core.Logger, config *GinServerOptions) *GinServer {
	if config == nil {
		log.Fatalw("GinServerOptions is required")
	}

	if config.Server == nil {
		log.Fatalw("ServerOptions is required")
	}

	// 设置 Gin 模式
	gin.SetMode(config.Server.Mode)

	engine := gin.New()

	// 应用基础中间件
	if config.EnableRecovery {
		engine.Use(middleware.Recovery())
	}

	if config.EnableRequestID {
		engine.Use(middleware.RequestID())
	}

	if config.EnableLogger {
		engine.Use(middleware.RequestLogger())
	}

	// 应用 CORS 中间件
	if config.CORS != nil && config.CORS.Enabled {
		corsConfig := ToCORSConfig(config.CORS)
		engine.Use(middleware.CORSWithConfig(corsConfig))
		log.Infow("CORS middleware enabled", "allow_origins", config.CORS.AllowOrigins)
	}

	// 注意：JWT 中间件需要在特定路由上使用，不在这里全局应用
	// 服务可以通过 GetEngine() 获取 engine 后手动添加
	if config.JWT != nil {
		log.Infow("JWT configuration provided", "expires_hours", config.JWT.ExpiresHours)
		// JWT 中间件将由调用者在需要的路由上应用
	}

	// Rate limit 中间件
	if config.RateLimit != nil && config.RateLimit.Enable {
		// 使用 IP 作为限流键
		engine.Use(middleware.RateLimitByIP(config.RateLimit.Rate, config.RateLimit.Burst))
		log.Infow("Rate limit middleware enabled",
			"rate", config.RateLimit.Rate,
			"burst", config.RateLimit.Burst)
	}

	httpServer := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port),
		Handler:        engine,
		ReadTimeout:    config.Server.ReadTimeout,
		WriteTimeout:   config.Server.WriteTimeout,
		IdleTimeout:    config.Server.IdleTimeout,
		MaxHeaderBytes: config.Server.MaxHeaderBytes, // 支持新的 MaxHeaderBytes 字段
	}

	return &GinServer{
		Engine: engine,
		server: httpServer,
		logger: log.With("component", "gin-server", "addr", httpServer.Addr),
	}
}

// RunOrDie 启动服务器（实现 Server 接口）
// 如果启动失败会记录致命错误并退出
func (s *GinServer) RunOrDie() {
	s.logger.Infow("Starting Gin HTTP server", "addr", s.server.Addr)

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Fatalw("Gin server failed", "err", err)
	}
}

// GracefulStop 优雅关停服务器（实现 Server 接口）
func (s *GinServer) GracefulStop(ctx context.Context) {
	s.logger.Infow("Gracefully stopping Gin server")

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Errorw("Gin server forced to shutdown", "err", err)
		return
	}

	s.logger.Infow("Gin server stopped gracefully")
}

// GetEngine 返回 Gin Engine 实例（用于注册路由）
func (s *GinServer) GetEngine() *gin.Engine {
	return s.Engine
}

// GetHTTPServer 返回底层的 *http.Server
func (s *GinServer) GetHTTPServer() *http.Server {
	return s.server
}

// Addr 返回服务器监听地址
func (s *GinServer) Addr() string {
	return s.server.Addr
}

// GetJWTMiddleware creates a JWT middleware from JWT config
// Returns nil if no JWT config is provided
func GetJWTMiddleware(config *GinServerOptions) *middleware.JWTMiddleware {
	if config == nil || config.JWT == nil {
		return nil
	}

	jwtConfig := ToJWTConfig(config.JWT)
	return middleware.NewJWTMiddleware(jwtConfig)
}
