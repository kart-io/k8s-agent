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

	"github.com/kart-io/k8s-agent/common/config"
	"github.com/kart-io/logger/core"
)

// GRPCServerOptions gRPC 服务器配置选项
type GRPCServerOptions struct {
	host                  string
	port                  int
	maxRecvMsgSize        int
	maxSendMsgSize        int
	connectionTimeout     time.Duration
	keepAliveTime         time.Duration
	keepAliveTimeout      time.Duration
	maxConnectionIdle     time.Duration
	maxConnectionAge      time.Duration
	maxConnectionAgeGrace time.Duration
	enableReflection      bool
	enableHealthCheck     bool
	tlsConfig             *tls.Config
	unaryInterceptors     []grpc.UnaryServerInterceptor
	streamInterceptors    []grpc.StreamServerInterceptor
}

// GRPCServerOption gRPC 服务器配置选项函数
type GRPCServerOption func(*GRPCServerOptions)

// WithGRPCHost 设置服务器地址
func WithGRPCHost(host string) GRPCServerOption {
	return func(o *GRPCServerOptions) {
		o.host = host
	}
}

// WithGRPCPort 设置端口
func WithGRPCPort(port int) GRPCServerOption {
	return func(o *GRPCServerOptions) {
		o.port = port
	}
}

// WithGRPCMaxRecvMsgSize 设置最大接收消息大小 (字节)
func WithGRPCMaxRecvMsgSize(size int) GRPCServerOption {
	return func(o *GRPCServerOptions) {
		o.maxRecvMsgSize = size
	}
}

// WithGRPCMaxSendMsgSize 设置最大发送消息大小 (字节)
func WithGRPCMaxSendMsgSize(size int) GRPCServerOption {
	return func(o *GRPCServerOptions) {
		o.maxSendMsgSize = size
	}
}

// WithGRPCConnectionTimeout 设置连接超时时间
func WithGRPCConnectionTimeout(d time.Duration) GRPCServerOption {
	return func(o *GRPCServerOptions) {
		o.connectionTimeout = d
	}
}

// WithGRPCKeepAliveTime 设置 KeepAlive 时间间隔
func WithGRPCKeepAliveTime(d time.Duration) GRPCServerOption {
	return func(o *GRPCServerOptions) {
		o.keepAliveTime = d
	}
}

// WithGRPCKeepAliveTimeout 设置 KeepAlive 超时时间
func WithGRPCKeepAliveTimeout(d time.Duration) GRPCServerOption {
	return func(o *GRPCServerOptions) {
		o.keepAliveTimeout = d
	}
}

// WithGRPCMaxConnectionIdle 设置最大连接空闲时间
func WithGRPCMaxConnectionIdle(d time.Duration) GRPCServerOption {
	return func(o *GRPCServerOptions) {
		o.maxConnectionIdle = d
	}
}

// WithGRPCMaxConnectionAge 设置最大连接存活时间
func WithGRPCMaxConnectionAge(d time.Duration) GRPCServerOption {
	return func(o *GRPCServerOptions) {
		o.maxConnectionAge = d
	}
}

// WithGRPCMaxConnectionAgeGrace 设置连接关闭宽限时间
func WithGRPCMaxConnectionAgeGrace(d time.Duration) GRPCServerOption {
	return func(o *GRPCServerOptions) {
		o.maxConnectionAgeGrace = d
	}
}

// WithGRPCReflection 启用/禁用 gRPC 反射服务
func WithGRPCReflection(enabled bool) GRPCServerOption {
	return func(o *GRPCServerOptions) {
		o.enableReflection = enabled
	}
}

// WithGRPCHealthCheck 启用/禁用健康检查服务
func WithGRPCHealthCheck(enabled bool) GRPCServerOption {
	return func(o *GRPCServerOptions) {
		o.enableHealthCheck = enabled
	}
}

// WithGRPCTLS 设置 TLS 配置
func WithGRPCTLS(tlsConfig *tls.Config) GRPCServerOption {
	return func(o *GRPCServerOptions) {
		o.tlsConfig = tlsConfig
	}
}

// WithGRPCUnaryInterceptor 添加一元拦截器
func WithGRPCUnaryInterceptor(interceptor grpc.UnaryServerInterceptor) GRPCServerOption {
	return func(o *GRPCServerOptions) {
		o.unaryInterceptors = append(o.unaryInterceptors, interceptor)
	}
}

// WithGRPCStreamInterceptor 添加流拦截器
func WithGRPCStreamInterceptor(interceptor grpc.StreamServerInterceptor) GRPCServerOption {
	return func(o *GRPCServerOptions) {
		o.streamInterceptors = append(o.streamInterceptors, interceptor)
	}
}

// defaultGRPCServerOptions 返回默认服务器配置
func defaultGRPCServerOptions() *GRPCServerOptions {
	return &GRPCServerOptions{
		host:                  "0.0.0.0",
		port:                  9090,
		maxRecvMsgSize:        4 * 1024 * 1024,  // 4MB
		maxSendMsgSize:        4 * 1024 * 1024,  // 4MB
		connectionTimeout:     120 * time.Second,
		keepAliveTime:         30 * time.Second,
		keepAliveTimeout:      10 * time.Second,
		maxConnectionIdle:     5 * time.Minute,
		maxConnectionAge:      30 * time.Minute,
		maxConnectionAgeGrace: 5 * time.Second,
		enableReflection:      true,
		enableHealthCheck:     true,
		unaryInterceptors:     []grpc.UnaryServerInterceptor{},
		streamInterceptors:    []grpc.StreamServerInterceptor{},
	}
}

// GRPCServer gRPC 服务器
type GRPCServer struct {
	server       *grpc.Server
	listener     net.Listener
	healthServer *health.Server
	logger       core.Logger
	options      *GRPCServerOptions
}

// NewGRPCServer 创建 gRPC 服务器
func NewGRPCServer(log core.Logger, opts ...GRPCServerOption) (*GRPCServer, error) {
	// 应用默认配置
	options := defaultGRPCServerOptions()

	// 应用用户配置
	for _, opt := range opts {
		opt(options)
	}

	return newGRPCServerFromOptions(log, options)
}

// NewGRPCServerFromConfig 从配置创建 gRPC 服务器
func NewGRPCServerFromConfig(log core.Logger, cfg *config.GRPCOptions) (*GRPCServer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("gRPC config is nil")
	}

	options := &GRPCServerOptions{
		host:                  cfg.Host,
		port:                  cfg.Port,
		maxRecvMsgSize:        cfg.MaxRecvMsgSize,
		maxSendMsgSize:        cfg.MaxSendMsgSize,
		connectionTimeout:     cfg.ConnectionTimeout,
		keepAliveTime:         cfg.KeepAliveTime,
		keepAliveTimeout:      cfg.KeepAliveTimeout,
		maxConnectionIdle:     cfg.MaxConnectionIdle,
		maxConnectionAge:      cfg.MaxConnectionAge,
		maxConnectionAgeGrace: cfg.MaxConnectionAgeGrace,
		enableReflection:      cfg.EnableReflection,
		enableHealthCheck:     cfg.EnableHealthCheck,
		unaryInterceptors:     []grpc.UnaryServerInterceptor{},
		streamInterceptors:    []grpc.StreamServerInterceptor{},
	}

	// 添加默认拦截器（使用 LoggingOptions 配置）
	interceptorCfg := &InterceptorConfig{
		Logger:        log,
		LoggingConfig: cfg.Logging, // 使用 gRPC 专用的日志配置
	}

	// 添加日志拦截器
	options.unaryInterceptors = append(options.unaryInterceptors, LoggingUnaryInterceptor(interceptorCfg))
	options.streamInterceptors = append(options.streamInterceptors, LoggingStreamInterceptor(interceptorCfg))

	// 添加恢复拦截器
	options.unaryInterceptors = append(options.unaryInterceptors, RecoveryUnaryInterceptor(log))
	options.streamInterceptors = append(options.streamInterceptors, RecoveryStreamInterceptor(log))

	// 添加请求ID拦截器
	options.unaryInterceptors = append(options.unaryInterceptors, RequestIDUnaryInterceptor())
	options.streamInterceptors = append(options.streamInterceptors, RequestIDStreamInterceptor())

	return newGRPCServerFromOptions(log, options)
}

// newGRPCServerFromOptions 从选项创建 gRPC 服务器（内部方法）
func newGRPCServerFromOptions(log core.Logger, options *GRPCServerOptions) (*GRPCServer, error) {

	// 创建监听器
	addr := fmt.Sprintf("%s:%d", options.host, options.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// 构建 gRPC 服务器选项
	serverOpts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(options.maxRecvMsgSize),
		grpc.MaxSendMsgSize(options.maxSendMsgSize),
		grpc.ConnectionTimeout(options.connectionTimeout),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:              options.keepAliveTime,
			Timeout:           options.keepAliveTimeout,
			MaxConnectionIdle: options.maxConnectionIdle,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             options.keepAliveTime / 2,
			PermitWithoutStream: true,
		}),
	}

	// 添加 TLS 支持
	if options.tlsConfig != nil {
		serverOpts = append(serverOpts, grpc.Creds(credentials.NewTLS(options.tlsConfig)))
	}

	// 添加拦截器
	if len(options.unaryInterceptors) > 0 {
		serverOpts = append(serverOpts, grpc.ChainUnaryInterceptor(options.unaryInterceptors...))
	}
	if len(options.streamInterceptors) > 0 {
		serverOpts = append(serverOpts, grpc.ChainStreamInterceptor(options.streamInterceptors...))
	}

	// 创建 gRPC 服务器
	server := grpc.NewServer(serverOpts...)

	// 创建健康检查服务
	var healthServer *health.Server
	if options.enableHealthCheck {
		healthServer = health.NewServer()
		grpc_health_v1.RegisterHealthServer(server, healthServer)
	}

	// 启用反射服务
	if options.enableReflection {
		reflection.Register(server)
	}

	return &GRPCServer{
		server:       server,
		listener:     listener,
		healthServer: healthServer,
		logger:       log.With("component", "grpc-server"),
		options:      options,
	}, nil
}

// Server 返回底层的 gRPC 服务器实例，用于注册服务
func (s *GRPCServer) Server() *grpc.Server {
	return s.server
}

// HealthServer 返回健康检查服务器，用于设置服务状态
func (s *GRPCServer) HealthServer() *health.Server {
	return s.healthServer
}

// Run 启动服务器（阻塞）
func (s *GRPCServer) Run() error {
	addr := s.listener.Addr().String()
	s.logger.Infow("Starting gRPC server", "addr", addr)

	if err := s.server.Serve(s.listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}

// Shutdown 优雅关闭服务器
func (s *GRPCServer) Shutdown(ctx context.Context) error {
	s.logger.Infow("Shutting down gRPC server")

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
		return nil
	case <-ctx.Done():
		s.logger.Warnw("gRPC server shutdown timeout, forcing stop")
		s.server.Stop()
		return ctx.Err()
	}
}

// Stop 强制停止服务器
func (s *GRPCServer) Stop() {
	s.logger.Warnw("Forcing gRPC server to stop")
	s.server.Stop()
	s.logger.Infow("gRPC server stopped")
}
