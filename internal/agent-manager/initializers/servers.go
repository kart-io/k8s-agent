package initializers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kart-io/k8s-agent/cmd/agent-manager/app/options"
	"github.com/kart-io/k8s-agent/common/bootstrap"
	"github.com/kart-io/k8s-agent/internal/agent-manager/api"
	agentgrpc "github.com/kart-io/k8s-agent/internal/agent-manager/grpc"
	"github.com/kart-io/k8s-agent/pkg/types"
	"github.com/kart-io/logger/core"
)

// HTTPServerInitializer HTTP API 服务器初始化器
type HTTPServerInitializer struct {
	opts       *options.ServerOptions
	logger     core.Logger
	registry   *RegistryInitializer
	dispatcher *DispatcherInitializer
	dbInit     *DatabaseInitializer
	redisInit  *RedisInitializer
	natsInit   *NATSInitializer
	apiServer  *api.Server
	eventProc  interface{} // 临时存储 event processor
	errChan    chan error  // 服务器错误通道
}

// NewHTTPServerInitializer 创建 HTTP 服务器初始化器
func NewHTTPServerInitializer(
	opts *options.ServerOptions,
	logger core.Logger,
	registry *RegistryInitializer,
	dispatcher *DispatcherInitializer,
	dbInit *DatabaseInitializer,
	redisInit *RedisInitializer,
	natsInit *NATSInitializer,
) *HTTPServerInitializer {
	return &HTTPServerInitializer{
		opts:       opts,
		logger:     logger,
		registry:   registry,
		dispatcher: dispatcher,
		dbInit:     dbInit,
		redisInit:  redisInit,
		natsInit:   natsInit,
	}
}

// Name 返回初始化器名称
func (h *HTTPServerInitializer) Name() string {
	return "http-server"
}

// Priority 返回初始化优先级
func (h *HTTPServerInitializer) Priority() int {
	return bootstrap.PriorityHTTP
}

// Initialize 执行初始化
func (h *HTTPServerInitializer) Initialize(ctx context.Context) error {
	h.logger.Infow("Initializing HTTP API server",
		"host", h.opts.Server.Host,
		"port", h.opts.Server.Port,
	)

	// 从 NATS 初始化器获取 event processor
	eventProc := h.natsInit.EventProcessor()

	// 创建 API 服务器
	h.apiServer = api.NewServer(
		types.ServerConfig{
			Host:         h.opts.Server.Host,
			Port:         h.opts.Server.Port,
			ReadTimeout:  h.opts.Server.ReadTimeout,
			WriteTimeout: h.opts.Server.WriteTimeout,
			GracefulStop: h.opts.Server.GracefulStop,
		},
		h.registry.Registry(),
		eventProc,
		h.dispatcher.Dispatcher(),
		h.dbInit.Store(),
		h.redisInit.Store(),
		h.logger,
	)

	// 创建错误通道用于捕获服务器启动错误
	h.errChan = make(chan error, 1)

	// 在后台启动 HTTP 服务器
	go func() {
		if err := h.apiServer.Start(); err != nil && err != http.ErrServerClosed {
			h.logger.Errorw("HTTP server fatal error", "error", err)
			h.errChan <- err
		}
	}()

	// 启动错误监听器
	go h.monitorServerErrors()

	h.logger.Infow("HTTP API server initialized successfully",
		"address", fmt.Sprintf("%s:%d", h.opts.Server.Host, h.opts.Server.Port),
	)
	return nil
}

// monitorServerErrors 监听服务器致命错误
func (h *HTTPServerInitializer) monitorServerErrors() {
	for err := range h.errChan {
		if err != nil {
			h.logger.Fatalw("HTTP server encountered fatal error, shutting down",
				"error", err,
				"component", "http-server",
			)
		}
	}
}

// Close 关闭 HTTP 服务器
func (h *HTTPServerInitializer) Close(ctx context.Context) error {
	if h.apiServer != nil {
		h.logger.Infow("Stopping HTTP API server")
		if err := h.apiServer.Stop(); err != nil {
			return err
		}
	}
	// 关闭错误通道
	if h.errChan != nil {
		close(h.errChan)
	}
	return nil
}

// GRPCServerInitializer gRPC 服务器初始化器
type GRPCServerInitializer struct {
	opts       *options.ServerOptions
	logger     core.Logger
	registry   *RegistryInitializer
	dispatcher *DispatcherInitializer
	dbInit     *DatabaseInitializer
	grpcServer *agentgrpc.Server
	errChan    chan error // 服务器错误通道
}

// NewGRPCServerInitializer 创建 gRPC 服务器初始化器
func NewGRPCServerInitializer(
	opts *options.ServerOptions,
	logger core.Logger,
	registry *RegistryInitializer,
	dispatcher *DispatcherInitializer,
	dbInit *DatabaseInitializer,
) *GRPCServerInitializer {
	return &GRPCServerInitializer{
		opts:       opts,
		logger:     logger,
		registry:   registry,
		dispatcher: dispatcher,
		dbInit:     dbInit,
	}
}

// Name 返回初始化器名称
func (g *GRPCServerInitializer) Name() string {
	return "grpc-server"
}

// Priority 返回初始化优先级
func (g *GRPCServerInitializer) Priority() int {
	return bootstrap.PriorityGRPC
}

// Initialize 执行初始化
func (g *GRPCServerInitializer) Initialize(ctx context.Context) error {
	if !g.opts.GRPC.Enable {
		g.logger.Infow("gRPC server is disabled, skipping initialization")
		return nil
	}

	g.logger.Infow("Initializing gRPC server",
		"host", g.opts.GRPC.Host,
		"port", g.opts.GRPC.Port,
	)

	grpcOpts := &agentgrpc.ServerOptions{
		Host:             g.opts.GRPC.Host,
		Port:             g.opts.GRPC.Port,
		MaxRecvMsgSize:   g.opts.GRPC.MaxRecvMsgSize,
		MaxSendMsgSize:   g.opts.GRPC.MaxSendMsgSize,
		KeepaliveTime:    g.opts.GRPC.KeepAliveTime,
		KeepaliveTimeout: g.opts.GRPC.KeepAliveTimeout,
		Registry:         g.registry.Registry(),
		Dispatcher:       g.dispatcher.Dispatcher(),
		Store:            g.dbInit.Store(),
	}

	grpcServer, err := agentgrpc.NewServer(grpcOpts, g.logger)
	if err != nil {
		return fmt.Errorf("failed to create gRPC server: %w", err)
	}
	g.grpcServer = grpcServer

	// 创建错误通道用于捕获服务器启动错误
	g.errChan = make(chan error, 1)

	// 在后台启动 gRPC 服务器
	go func() {
		if err := g.grpcServer.Start(ctx); err != nil {
			g.logger.Errorw("gRPC server fatal error", "error", err)
			g.errChan <- err
		}
	}()

	// 启动错误监听器
	go g.monitorServerErrors()

	g.logger.Infow("gRPC server initialized successfully",
		"address", g.grpcServer.Address(),
	)
	return nil
}

// monitorServerErrors 监听服务器致命错误
func (g *GRPCServerInitializer) monitorServerErrors() {
	for err := range g.errChan {
		if err != nil {
			g.logger.Fatalw("gRPC server encountered fatal error, shutting down",
				"error", err,
				"component", "grpc-server",
			)
		}
	}
}

// Close 关闭 gRPC 服务器
func (g *GRPCServerInitializer) Close(ctx context.Context) error {
	if g.grpcServer != nil {
		g.logger.Infow("Stopping gRPC server")
		if err := g.grpcServer.Stop(); err != nil {
			return err
		}
	}
	// 关闭错误通道
	if g.errChan != nil {
		close(g.errChan)
	}
	return nil
}
