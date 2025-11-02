// Copyright 2025 Kart Project. All rights reserved.
// Standard gRPC server implementation with functional options
//
// This file provides a traditional gRPC server that can be configured
// using functional options or from GRPCOptions config. It includes
// support for TLS, interceptors, health checks, and reflection.

package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

// StandardGRPCOption 函数式配置选项（现在修改 options.GRPCOptions）
type StandardGRPCOption func(*options.GRPCOptions)

// WithGRPCHost 设置服务器地址
func WithGRPCHost(host string) StandardGRPCOption {
	return func(o *options.GRPCOptions) {
		o.Host = host
	}
}

// WithGRPCPort 设置端口
func WithGRPCPort(port int) StandardGRPCOption {
	return func(o *options.GRPCOptions) {
		o.Port = port
	}
}

// WithGRPCMaxRecvMsgSize 设置最大接收消息大小 (字节)
func WithGRPCMaxRecvMsgSize(size int) StandardGRPCOption {
	return func(o *options.GRPCOptions) {
		o.MaxRecvMsgSize = size
	}
}

// WithGRPCMaxSendMsgSize 设置最大发送消息大小 (字节)
func WithGRPCMaxSendMsgSize(size int) StandardGRPCOption {
	return func(o *options.GRPCOptions) {
		o.MaxSendMsgSize = size
	}
}

// WithGRPCConnectionTimeout 设置连接超时时间
func WithGRPCConnectionTimeout(d ...time.Duration) StandardGRPCOption {
	return func(o *options.GRPCOptions) {
		if len(d) > 0 {
			o.ConnectionTimeout = d[0]
		}
	}
}

// WithGRPCKeepAliveTime 设置 KeepAlive 时间间隔
func WithGRPCKeepAliveTime(d ...time.Duration) StandardGRPCOption {
	return func(o *options.GRPCOptions) {
		if len(d) > 0 {
			o.KeepAliveTime = d[0]
		}
	}
}

// WithGRPCKeepAliveTimeout 设置 KeepAlive 超时时间
func WithGRPCKeepAliveTimeout(d ...time.Duration) StandardGRPCOption {
	return func(o *options.GRPCOptions) {
		if len(d) > 0 {
			o.KeepAliveTimeout = d[0]
		}
	}
}

// WithGRPCMaxConnectionIdle 设置最大连接空闲时间
func WithGRPCMaxConnectionIdle(d ...time.Duration) StandardGRPCOption {
	return func(o *options.GRPCOptions) {
		if len(d) > 0 {
			o.MaxConnectionIdle = d[0]
		}
	}
}

// WithGRPCMaxConnectionAge 设置最大连接存活时间
func WithGRPCMaxConnectionAge(d ...time.Duration) StandardGRPCOption {
	return func(o *options.GRPCOptions) {
		if len(d) > 0 {
			o.MaxConnectionAge = d[0]
		}
	}
}

// WithGRPCMaxConnectionAgeGrace 设置连接关闭宽限时间
func WithGRPCMaxConnectionAgeGrace(d ...time.Duration) StandardGRPCOption {
	return func(o *options.GRPCOptions) {
		if len(d) > 0 {
			o.MaxConnectionAgeGrace = d[0]
		}
	}
}

// WithGRPCReflection 启用/禁用 gRPC 反射服务
func WithGRPCReflection(enabled bool) StandardGRPCOption {
	return func(o *options.GRPCOptions) {
		o.EnableReflection = enabled
	}
}

// WithGRPCHealthCheck 启用/禁用健康检查服务
func WithGRPCHealthCheck(enabled bool) StandardGRPCOption {
	return func(o *options.GRPCOptions) {
		o.EnableHealthCheck = enabled
	}
}

// grpcTLSConfig 内部TLS配置（不暴露到options）
type grpcTLSConfig struct {
	tlsConfig          *tls.Config
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
}

// WithGRPCTLS 设置 TLS 配置
func WithGRPCTLS(tlsConfig *tls.Config) func(*grpcTLSConfig) {
	return func(c *grpcTLSConfig) {
		c.tlsConfig = tlsConfig
	}
}

// WithGRPCUnaryInterceptor 添加一元拦截器
func WithGRPCUnaryInterceptor(interceptor grpc.UnaryServerInterceptor) func(*grpcTLSConfig) {
	return func(c *grpcTLSConfig) {
		c.unaryInterceptors = append(c.unaryInterceptors, interceptor)
	}
}

// WithGRPCStreamInterceptor 添加流拦截器
func WithGRPCStreamInterceptor(interceptor grpc.StreamServerInterceptor) func(*grpcTLSConfig) {
	return func(c *grpcTLSConfig) {
		c.streamInterceptors = append(c.streamInterceptors, interceptor)
	}
}

// StandardGRPCServer 标准 gRPC 服务器（函数式配置）
// 实现 Server 接口，可与 Serve() 和 MultiServe() 函数配合使用
type StandardGRPCServer struct {
	server       *grpc.Server
	listener     net.Listener
	healthServer *health.Server
	logger       core.Logger
	options      *options.GRPCOptions
	tlsConfig    *grpcTLSConfig
}

// NewStandardGRPCServer 创建标准 gRPC 服务器（函数式配置）
//
// 使用示例:
//
//	// 方式1: 使用函数式选项（简单场景）
//	grpcSrv, err := server.NewStandardGRPCServer(log,
//	    server.WithGRPCPort(9090),
//	    server.WithGRPCReflection(true),
//	    server.WithGRPCHealthCheck(true),
//	)
//
//	// 方式2: 使用配置对象（推荐用于配置文件）
//	opts := options.NewGRPCOptions()
//	opts.Port = 9090
//	grpcSrv, err := server.NewStandardGRPCServerFromConfig(log, opts)
//
//	// 注册服务
//	mypb.RegisterMyServiceServer(grpcSrv.GetServer(), myImpl)
//
//	// 启动服务器
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	if err := server.Serve(ctx, grpcSrv, log); err != nil {
//	    log.Fatalw("Server failed", "err", err)
//	}
func NewStandardGRPCServer(log core.Logger, opts ...StandardGRPCOption) (*StandardGRPCServer, error) {
	// 应用默认配置
	grpcOpts := options.NewGRPCOptions()

	// 应用用户配置
	for _, opt := range opts {
		opt(grpcOpts)
	}

	return newStandardGRPCServerFromOptions(log, grpcOpts, &grpcTLSConfig{})
}

// NewStandardGRPCServerFromConfig 从 GRPCOptions 配置创建服务器
// 这种方式使用配置文件而非函数式选项
func NewStandardGRPCServerFromConfig(log core.Logger, cfg *options.GRPCOptions) (*StandardGRPCServer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("gRPC config is nil")
	}

	// 添加默认拦截器（使用 LoggingOptions 配置）
	tlsCfg := &grpcTLSConfig{}

	interceptorCfg := &InterceptorConfig{
		Logger:        log,
		LoggingConfig: cfg.Logging, // 使用 gRPC 专用的日志配置
	}

	// 添加日志拦截器
	tlsCfg.unaryInterceptors = append(tlsCfg.unaryInterceptors, LoggingUnaryInterceptor(interceptorCfg))
	tlsCfg.streamInterceptors = append(tlsCfg.streamInterceptors, LoggingStreamInterceptor(interceptorCfg))

	// 添加恢复拦截器
	tlsCfg.unaryInterceptors = append(tlsCfg.unaryInterceptors, RecoveryUnaryInterceptor(log))
	tlsCfg.streamInterceptors = append(tlsCfg.streamInterceptors, RecoveryStreamInterceptor(log))

	// 添加请求ID拦截器
	tlsCfg.unaryInterceptors = append(tlsCfg.unaryInterceptors, RequestIDUnaryInterceptor())
	tlsCfg.streamInterceptors = append(tlsCfg.streamInterceptors, RequestIDStreamInterceptor())

	return newStandardGRPCServerFromOptions(log, cfg, tlsCfg)
}

// newStandardGRPCServerFromOptions 从选项创建 gRPC 服务器（内部方法）
func newStandardGRPCServerFromOptions(log core.Logger, opts *options.GRPCOptions, tlsCfg *grpcTLSConfig) (*StandardGRPCServer, error) {
	// 创建监听器
	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// 构建 gRPC 服务器选项
	serverOpts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(opts.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(opts.MaxSendMsgSize),
		grpc.ConnectionTimeout(opts.ConnectionTimeout),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:              opts.KeepAliveTime,
			Timeout:           opts.KeepAliveTimeout,
			MaxConnectionIdle: opts.MaxConnectionIdle,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             opts.KeepAliveTime / 2,
			PermitWithoutStream: true,
		}),
	}

	// 添加 TLS 支持
	if tlsCfg.tlsConfig != nil {
		serverOpts = append(serverOpts, grpc.Creds(credentials.NewTLS(tlsCfg.tlsConfig)))
	}

	// 添加拦截器
	if len(tlsCfg.unaryInterceptors) > 0 {
		serverOpts = append(serverOpts, grpc.ChainUnaryInterceptor(tlsCfg.unaryInterceptors...))
	}
	if len(tlsCfg.streamInterceptors) > 0 {
		serverOpts = append(serverOpts, grpc.ChainStreamInterceptor(tlsCfg.streamInterceptors...))
	}

	// 创建 gRPC 服务器
	grpcServer := grpc.NewServer(serverOpts...)

	// 创建健康检查服务
	var healthServer *health.Server
	if opts.EnableHealthCheck {
		healthServer = health.NewServer()
		grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	}

	// 启用反射服务
	if opts.EnableReflection {
		reflection.Register(grpcServer)
	}

	return &StandardGRPCServer{
		server:       grpcServer,
		listener:     listener,
		healthServer: healthServer,
		logger:       log.With("component", "grpc-standard-server", "addr", addr),
		options:      opts,
		tlsConfig:    tlsCfg,
	}, nil
}

// RunOrDie 启动服务器（实现 Server 接口）
// 如果启动失败会记录致命错误并退出
func (s *StandardGRPCServer) RunOrDie() {
	addr := s.listener.Addr().String()
	s.logger.Infow("Starting standard gRPC server", "addr", addr)

	if err := s.server.Serve(s.listener); err != nil {
		s.logger.Fatalw("gRPC server failed", "err", err)
	}
}

// GracefulStop 优雅关停服务器（实现 Server 接口）
func (s *StandardGRPCServer) GracefulStop(ctx context.Context) {
	s.logger.Infow("Gracefully stopping gRPC server")

	// 标记所有服务为 NOT_SERVING
	if s.healthServer != nil {
		s.healthServer.Shutdown()
	}

	// 使用 channel 等待 GracefulStop 完成
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	// 等待优雅关闭完成或超时
	select {
	case <-done:
		s.logger.Infow("gRPC server stopped gracefully")
	case <-ctx.Done():
		s.logger.Warnw("gRPC server shutdown timeout, forcing stop")
		s.server.Stop()
	}
}

// GetServer 返回底层的 gRPC 服务器实例（用于注册服务）
func (s *StandardGRPCServer) GetServer() *grpc.Server {
	return s.server
}

// GetHealthServer 返回健康检查服务器（用于设置服务状态）
func (s *StandardGRPCServer) GetHealthServer() *health.Server {
	return s.healthServer
}

// Addr 返回服务器监听地址
func (s *StandardGRPCServer) Addr() string {
	return s.listener.Addr().String()
}
