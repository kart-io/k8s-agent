package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/logger/core"

	"github.com/kart-io/k8s-agent/common/middleware"
)

// GinServerOptions Gin 服务器配置选项
type GinServerOptions struct {
	host         string
	port         int
	mode         string
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
}

// GinServerOption Gin 服务器配置选项函数
type GinServerOption func(*GinServerOptions)

// WithGinHost 设置服务器地址
func WithGinHost(host string) GinServerOption {
	return func(o *GinServerOptions) {
		o.host = host
	}
}

// WithGinPort 设置端口
func WithGinPort(port int) GinServerOption {
	return func(o *GinServerOptions) {
		o.port = port
	}
}

// WithGinMode 设置运行模式 (debug, release, test)
func WithGinMode(mode string) GinServerOption {
	return func(o *GinServerOptions) {
		o.mode = mode
	}
}

// WithGinReadTimeout 设置读超时
func WithGinReadTimeout(d time.Duration) GinServerOption {
	return func(o *GinServerOptions) {
		o.readTimeout = d
	}
}

// WithGinWriteTimeout 设置写超时
func WithGinWriteTimeout(d time.Duration) GinServerOption {
	return func(o *GinServerOptions) {
		o.writeTimeout = d
	}
}

// WithGinIdleTimeout 设置空闲超时
func WithGinIdleTimeout(d time.Duration) GinServerOption {
	return func(o *GinServerOptions) {
		o.idleTimeout = d
	}
}

// defaultGinServerOptions 返回默认服务器配置
func defaultGinServerOptions() *GinServerOptions {
	return &GinServerOptions{
		host:         "0.0.0.0",
		port:         8080,
		mode:         "release",
		readTimeout:  10 * time.Second,
		writeTimeout: 10 * time.Second,
		idleTimeout:  60 * time.Second,
	}
}

// GinServer Gin 服务器
type GinServer struct {
	Engine *gin.Engine
	Server *http.Server
	logger core.Logger
}

// NewGinServer 创建 Gin 服务器
func NewGinServer(log core.Logger, opts ...GinServerOption) *GinServer {
	// 应用默认配置
	options := defaultGinServerOptions()

	// 应用用户配置
	for _, opt := range opts {
		opt(options)
	}

	// 设置 Gin 模式
	gin.SetMode(options.mode)

	engine := gin.New()

	// 默认中间件
	engine.Use(middleware.Recovery())
	engine.Use(middleware.RequestID())
	engine.Use(middleware.RequestLogger())
	engine.Use(middleware.CORS())

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", options.host, options.port),
		Handler:      engine,
		ReadTimeout:  options.readTimeout,
		WriteTimeout: options.writeTimeout,
		IdleTimeout:  options.idleTimeout,
	}

	return &GinServer{
		Engine: engine,
		Server: server,
		logger: log.With("component", "http-server"),
	}
}

// Run 启动服务器
func (s *GinServer) Run() error {
	s.logger.Infow("Starting HTTP server", "addr", s.Server.Addr)
	if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

// RunTLS 启动 HTTPS 服务器
func (s *GinServer) RunTLS(certFile string, keyFile string) error {
	s.logger.Infow("Starting HTTPS server", "addr", s.Server.Addr)
	if err := s.Server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start TLS server: %w", err)
	}
	return nil
}

// Shutdown 优雅关闭
func (s *GinServer) Shutdown(ctx context.Context) error {
	s.logger.Infow("Shutting down HTTP server")
	if err := s.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}
	s.logger.Infow("HTTP server stopped")
	return nil
}
