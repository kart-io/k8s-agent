// Copyright 2025 Kart Project. All rights reserved.
// Kratos HTTP server implementation

package server

import (
	"context"
	"fmt"
	"time"

	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/kart-io/logger/core"

	"github.com/kart-io/k8s-agent/common/options"
)

// KratosServerOption Kratos 服务器配置选项函数
type KratosServerOption func(*options.ServerOptions)

// WithKratosHost 设置服务器地址
func WithKratosHost(host string) KratosServerOption {
	return func(o *options.ServerOptions) {
		o.Host = host
	}
}

// WithKratosPort 设置端口
func WithKratosPort(port int) KratosServerOption {
	return func(o *options.ServerOptions) {
		o.Port = port
	}
}

// WithKratosReadTimeout 设置读超时
func WithKratosReadTimeout(d ...time.Duration) KratosServerOption {
	return func(o *options.ServerOptions) {
		if len(d) > 0 {
			o.ReadTimeout = d[0]
		}
	}
}

// WithKratosWriteTimeout 设置写超时
func WithKratosWriteTimeout(d ...time.Duration) KratosServerOption {
	return func(o *options.ServerOptions) {
		if len(d) > 0 {
			o.WriteTimeout = d[0]
		}
	}
}

// WithKratosIdleTimeout 设置空闲超时
func WithKratosIdleTimeout(d ...time.Duration) KratosServerOption {
	return func(o *options.ServerOptions) {
		if len(d) > 0 {
			o.IdleTimeout = d[0]
		}
	}
}

// KratosServer Kratos HTTP 服务器
// 实现 Server 接口，可与 Serve() 和 MultiServe() 函数配合使用
type KratosServer struct {
	server *kratoshttp.Server
	logger core.Logger
	opts   *options.ServerOptions
}

// NewKratosServer 创建 Kratos HTTP 服务器（使用函数式选项）
//
// 使用示例:
//
//	// 方式1: 使用函数式选项（简单场景）
//	kratosSrv := server.NewKratosServer(log,
//	    server.WithKratosPort(8080),
//	    server.WithKratosReadTimeout(30*time.Second),
//	)
//
//	// 方式2: 使用配置对象（推荐用于配置文件）
//	opts := options.NewServerOptions()
//	opts.Port = 8080
//	kratosSrv := server.NewKratosServerFromConfig(log, opts)
//
//	// 注册路由
//	kratosSrv.GetServer().Route("/").GET("/health", func(ctx kratoshttp.Context) error {
//	    return ctx.JSON(200, map[string]string{"status": "ok"})
//	})
//
//	// 启动服务器
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	if err := server.Serve(ctx, kratosSrv, log); err != nil {
//	    log.Fatalw("Server failed", "err", err)
//	}
func NewKratosServer(log core.Logger, opts ...KratosServerOption) *KratosServer {
	// 应用默认配置
	serverOpts := options.NewServerOptions()

	// 应用用户配置
	for _, opt := range opts {
		opt(serverOpts)
	}

	return newKratosServerFromOptions(log, serverOpts)
}

// NewKratosServerFromConfig 从配置对象创建 Kratos 服务器
// 推荐用于从配置文件加载配置的场景
func NewKratosServerFromConfig(log core.Logger, opts *options.ServerOptions) *KratosServer {
	return newKratosServerFromOptions(log, opts)
}

// newKratosServerFromOptions 内部函数：从 ServerOptions 创建服务器
func newKratosServerFromOptions(log core.Logger, opts *options.ServerOptions) *KratosServer {
	// 创建 Kratos HTTP 服务器
	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	kratosServer := kratoshttp.NewServer(
		kratoshttp.Address(addr),
		kratoshttp.Timeout(opts.ReadTimeout),
	)

	return &KratosServer{
		server: kratosServer,
		logger: log.With("component", "kratos-server", "addr", addr),
		opts:   opts,
	}
}

// RunOrDie 启动服务器（实现 Server 接口）
// 如果启动失败会记录致命错误并退出
func (s *KratosServer) RunOrDie() {
	s.logger.Infow("Starting Kratos HTTP server", "addr", s.Addr())

	// Kratos 的 Start 方法会阻塞直到服务器停止
	if err := s.server.Start(context.Background()); err != nil {
		s.logger.Fatalw("Kratos server failed", "err", err)
	}
}

// GracefulStop 优雅关停服务器（实现 Server 接口）
func (s *KratosServer) GracefulStop(ctx context.Context) {
	s.logger.Infow("Gracefully stopping Kratos server")

	if err := s.server.Stop(ctx); err != nil {
		s.logger.Errorw("Kratos server forced to shutdown", "err", err)
		return
	}

	s.logger.Infow("Kratos server stopped gracefully")
}

// GetServer 返回底层 Kratos Server 实例（用于注册路由）
func (s *KratosServer) GetServer() *kratoshttp.Server {
	return s.server
}

// Addr 返回服务器监听地址
func (s *KratosServer) Addr() string {
	return fmt.Sprintf("%s:%d", s.opts.Host, s.opts.Port)
}
