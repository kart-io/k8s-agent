// Copyright 2025 Kart Project. All rights reserved.
// Gin HTTP server implementation

package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/logger/core"

	"github.com/kart-io/k8s-agent/common/middleware"
	"github.com/kart-io/k8s-agent/common/options"
)

// GinServerOption Gin 服务器配置选项函数
// 用于函数式配置，简化场景使用
type GinServerOption func(*options.ServerOptions)

// WithGinHost 设置服务器地址
func WithGinHost(host string) GinServerOption {
	return func(o *options.ServerOptions) {
		o.Host = host
	}
}

// WithGinPort 设置端口
func WithGinPort(port int) GinServerOption {
	return func(o *options.ServerOptions) {
		o.Port = port
	}
}

// WithGinMode 设置运行模式 (debug, release, test)
func WithGinMode(mode string) GinServerOption {
	return func(o *options.ServerOptions) {
		o.Mode = mode
	}
}

// WithGinReadTimeout 设置读超时
func WithGinReadTimeout(d ...time.Duration) GinServerOption {
	return func(o *options.ServerOptions) {
		if len(d) > 0 {
			o.ReadTimeout = d[0]
		}
	}
}

// WithGinWriteTimeout 设置写超时
func WithGinWriteTimeout(d ...time.Duration) GinServerOption {
	return func(o *options.ServerOptions) {
		if len(d) > 0 {
			o.WriteTimeout = d[0]
		}
	}
}

// WithGinIdleTimeout 设置空闲超时
func WithGinIdleTimeout(d ...time.Duration) GinServerOption {
	return func(o *options.ServerOptions) {
		if len(d) > 0 {
			o.IdleTimeout = d[0]
		}
	}
}

// GinServer Gin HTTP 服务器
// 实现 Server 接口，可与 Serve() 和 MultiServe() 函数配合使用
type GinServer struct {
	Engine *gin.Engine
	server *http.Server
	logger core.Logger
}

// NewGinServer 创建 Gin 服务器（使用函数式选项）
//
// 使用示例:
//
//	// 方式1: 使用函数式选项（简单场景）
//	ginSrv := server.NewGinServer(log,
//	    server.WithGinPort(8080),
//	    server.WithGinMode("release"),
//	)
//
//	// 方式2: 使用配置对象（推荐用于配置文件）
//	opts := options.NewServerOptions()
//	opts.Port = 8080
//	opts.Mode = "release"
//	ginSrv := server.NewGinServerFromConfig(log, opts)
//
//	// 注册路由
//	ginSrv.Engine.GET("/health", func(c *gin.Context) {
//	    c.JSON(200, gin.H{"status": "ok"})
//	})
//
//	// 启动服务器
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	if err := server.Serve(ctx, ginSrv, log); err != nil {
//	    log.Fatalw("Server failed", "err", err)
//	}
func NewGinServer(log core.Logger, opts ...GinServerOption) *GinServer {
	// 应用默认配置
	serverOpts := options.NewServerOptions()

	// 应用用户配置
	for _, opt := range opts {
		opt(serverOpts)
	}

	return newGinServerFromOptions(log, serverOpts)
}

// NewGinServerFromConfig 从配置对象创建 Gin 服务器
// 推荐用于从配置文件加载配置的场景
func NewGinServerFromConfig(log core.Logger, opts *options.ServerOptions) *GinServer {
	return newGinServerFromOptions(log, opts)
}

// newGinServerFromOptions 内部函数：从 ServerOptions 创建服务器
func newGinServerFromOptions(log core.Logger, opts *options.ServerOptions) *GinServer {
	// 设置 Gin 模式
	gin.SetMode(opts.Mode)

	engine := gin.New()

	// 默认中间件
	engine.Use(middleware.Recovery())
	engine.Use(middleware.RequestID())
	engine.Use(middleware.RequestLogger())
	engine.Use(middleware.CORS())

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", opts.Host, opts.Port),
		Handler:      engine,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		IdleTimeout:  opts.IdleTimeout,
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
