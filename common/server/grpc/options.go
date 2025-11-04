// Copyright 2025 Kart Project. All rights reserved.
// gRPC server implementation using Options pattern
// 参考 OneX 项目设计

package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

// GRPCOptionsServer 使用 Options 配置模式的 gRPC 服务器
// 参考 OneX 项目设计，实现 Server 接口
type GRPCOptionsServer struct {
	srv          *grpc.Server
	lis          net.Listener
	log          core.Logger
	serviceName  string
	healthServer *health.Server
}

// NewGRPCOptionsServer 创建使用 Options 配置的 gRPC 服务器
// 参考 OneX 项目的 NewGRPCServer 设计
//
// 参数:
//   - grpcOpts: gRPC 服务器配置选项
//   - tlsOpts: TLS 配置选项 (可选，为 nil 则不使用 TLS)
//   - registerFunc: 服务注册函数，用于注册 gRPC 服务
//   - log: 日志记录器
//
// 使用示例:
//
//	// 创建 gRPC 服务器
//	grpcOpts := options.NewGRPCOptions()
//	grpcSrv, err := server.NewGRPCOptionsServer(
//	    grpcOpts,
//	    nil, // 不使用 TLS
//	    func(srv grpc.ServiceRegistrar) {
//	        mypb.RegisterMyServiceServer(srv, myServiceImpl)
//	    },
//	    log,
//	)
//
//	// 启动服务器
//	ctx, cancel := context.WithCancel(context.Background())
//	if err := server.Serve(ctx, grpcSrv, log); err != nil {
//	    log.Fatalw("Server failed", "err", err)
//	}
func NewGRPCOptionsServer(
	grpcOpts *options.GRPCOptions,
	tlsOpts *options.TLSOptions,
	registerFunc func(grpc.ServiceRegistrar),
	log core.Logger,
) (*GRPCOptionsServer, error) {
	// 构建监听地址
	addr := fmt.Sprintf("%s:%d", grpcOpts.Host, grpcOpts.Port)

	// 创建监听器
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorw("Failed to create gRPC listener",
			"addr", addr,
			"err", err,
		)
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// 构建 gRPC 服务器选项
	var serverOptions []grpc.ServerOption

	// 消息大小限制
	if grpcOpts.MaxRecvMsgSize > 0 {
		serverOptions = append(serverOptions, grpc.MaxRecvMsgSize(grpcOpts.MaxRecvMsgSize))
	}
	if grpcOpts.MaxSendMsgSize > 0 {
		serverOptions = append(serverOptions, grpc.MaxSendMsgSize(grpcOpts.MaxSendMsgSize))
	}

	// KeepAlive 配置
	kaParams := keepalive.ServerParameters{
		Time:    grpcOpts.KeepAliveTime,
		Timeout: grpcOpts.KeepAliveTimeout,
	}
	if grpcOpts.MaxConnectionIdle > 0 {
		kaParams.MaxConnectionIdle = grpcOpts.MaxConnectionIdle
	}
	if grpcOpts.MaxConnectionAge > 0 {
		kaParams.MaxConnectionAge = grpcOpts.MaxConnectionAge
	}
	if grpcOpts.MaxConnectionAgeGrace > 0 {
		kaParams.MaxConnectionAgeGrace = grpcOpts.MaxConnectionAgeGrace
	}
	serverOptions = append(serverOptions,
		grpc.KeepaliveParams(kaParams),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             grpcOpts.KeepAliveTime,
			PermitWithoutStream: true,
		}),
	)

	// TLS 配置
	if tlsOpts != nil && tlsOpts.UseTLS {
		tlsConfig, err := buildGRPCTLSConfig(tlsOpts, log)
		if err != nil {
			_ = lis.Close()
			return nil, fmt.Errorf("failed to build TLS config: %w", err)
		}

		serverOptions = append(serverOptions, grpc.Creds(credentials.NewTLS(tlsConfig)))
		log.Infow("gRPC server TLS enabled",
			"cert", tlsOpts.Cert,
			"key", tlsOpts.Key,
		)
	}

	// 创建 gRPC 服务器
	grpcsrv := grpc.NewServer(serverOptions...)

	// 注册用户服务
	if registerFunc != nil {
		registerFunc(grpcsrv)
	}

	// 健康检查服务
	var healthServer *health.Server
	if grpcOpts.EnableHealthCheck {
		healthServer = health.NewServer()
		grpc_health_v1.RegisterHealthServer(grpcsrv, healthServer)
	}

	// 反射服务
	if grpcOpts.EnableReflection {
		reflection.Register(grpcsrv)
	}

	serviceName := "GRPCService" // 默认服务名

	// 设置健康状态
	if healthServer != nil {
		healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
	}

	return &GRPCOptionsServer{
		srv:          grpcsrv,
		lis:          lis,
		log:          log.With("component", "grpc-options-server", "addr", lis.Addr().String()),
		serviceName:  serviceName,
		healthServer: healthServer,
	}, nil
}

// RunOrDie 启动 gRPC 服务器并在出错时记录致命错误
// 实现 Server 接口
func (s *GRPCOptionsServer) RunOrDie() {
	s.log.Infow("Starting gRPC server",
		"addr", s.lis.Addr().String(),
		"service", s.serviceName,
	)

	if err := s.srv.Serve(s.lis); err != nil {
		s.log.Fatalw("gRPC server failed", "err", err)
	}
}

// GracefulStop 优雅地关闭 gRPC 服务器
// 实现 Server 接口
func (s *GRPCOptionsServer) GracefulStop(ctx context.Context) {
	s.log.Infow("Gracefully stopping gRPC server")

	// 设置健康检查状态为 NOT_SERVING
	if s.healthServer != nil {
		s.healthServer.SetServingStatus(s.serviceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	}

	// gRPC 的 GracefulStop 不支持 context，但会等待所有 RPC 完成
	done := make(chan struct{})
	go func() {
		s.srv.GracefulStop()
		close(done)
	}()

	// 等待优雅关闭完成或超时
	select {
	case <-done:
		s.log.Infow("gRPC server stopped gracefully")
	case <-ctx.Done():
		s.log.Warnw("gRPC server graceful shutdown timeout, forcing stop")
		s.srv.Stop() // 强制停止
	}
}

// GetGRPCServer 返回底层的 *grpc.Server (用于高级配置)
func (s *GRPCOptionsServer) GetGRPCServer() *grpc.Server {
	return s.srv
}

// GetListener 返回网络监听器
func (s *GRPCOptionsServer) GetListener() net.Listener {
	return s.lis
}

// Addr 返回服务器监听地址
func (s *GRPCOptionsServer) Addr() string {
	return s.lis.Addr().String()
}

// buildGRPCTLSConfig 构建 gRPC TLS 配置
// 参考 OneX 的 MustTLSConfig 实现
func buildGRPCTLSConfig(tlsOpts *options.TLSOptions, log core.Logger) (*tls.Config, error) {
	// 加载服务器证书和密钥
	cert, err := tls.LoadX509KeyPair(tlsOpts.Cert, tlsOpts.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tlsOpts.MinVersion,
	}

	// 如果指定了客户端 CA 文件，启用 mTLS
	if tlsOpts.CaCert != "" {
		caCert, err := os.ReadFile(tlsOpts.CaCert)
		if err != nil {
			return nil, fmt.Errorf("failed to read client CA file: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse client CA certificate")
		}

		tlsConfig.ClientCAs = caCertPool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert

		log.Infow("mTLS enabled for gRPC server", "client_ca", tlsOpts.CaCert)
	}

	return tlsConfig, nil
}
