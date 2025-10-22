package grpc

import (
	"context"

	"github.com/kart-io/logger/core"

	"github.com/kart-io/k8s-agent/internal/agent-manager/grpc/services"
	"github.com/kart-io/k8s-agent/common/config"
	"github.com/kart-io/k8s-agent/common/server"

	agentv1 "github.com/kart-io/k8s-agent/api/proto/gen/agentmanager/agent/v1"
	commandv1 "github.com/kart-io/k8s-agent/api/proto/gen/agentmanager/command/v1"
	eventv1 "github.com/kart-io/k8s-agent/api/proto/gen/agentmanager/event/v1"
)

// Server gRPC 服务器包装
type Server struct {
	grpcServer *server.GRPCServer
	logger     core.Logger

	// Services
	agentService   *services.AgentServiceServer
	commandService *services.CommandServiceServer
	eventService   *services.EventServiceServer
}

// ServerDependencies gRPC 服务器依赖
type ServerDependencies struct {
	Registry   services.AgentRegistry
	Dispatcher services.CommandDispatcher
	Processor  services.EventProcessor
	AgentStore services.AgentStore
	EventStore services.EventStore
}

// NewServer 从配置创建 gRPC 服务器
func NewServer(cfg *config.GRPCOptions, logger core.Logger, deps *ServerDependencies) (*Server, error) {
	// 使用 common 库的 NewGRPCServerFromConfig
	grpcServer, err := server.NewGRPCServerFromConfig(logger, cfg)
	if err != nil {
		return nil, err
	}

	// 创建服务实例并注入依赖
	agentService := services.NewAgentServiceServer(logger, deps.Registry, deps.AgentStore)
	commandService := services.NewCommandServiceServer(logger, deps.Dispatcher)
	eventService := services.NewEventServiceServer(logger, deps.Processor, deps.EventStore)

	// 注册服务
	agentv1.RegisterAgentServiceServer(grpcServer.Server(), agentService)
	commandv1.RegisterCommandServiceServer(grpcServer.Server(), commandService)
	eventv1.RegisterEventServiceServer(grpcServer.Server(), eventService)

	logger.Infow("gRPC services registered",
		"host", cfg.Host,
		"port", cfg.Port,
		"reflection", cfg.EnableReflection,
		"health_check", cfg.EnableHealthCheck,
	)

	return &Server{
		grpcServer:     grpcServer,
		logger:         logger.With("component", "grpc-server"),
		agentService:   agentService,
		commandService: commandService,
		eventService:   eventService,
	}, nil
}

// Run 启动服务器（阻塞）
func (s *Server) Run() error {
	s.logger.Infow("Starting gRPC server")
	return s.grpcServer.Run()
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Infow("Shutting down gRPC server")
	return s.grpcServer.Shutdown(ctx)
}

// Stop 强制停止服务器
func (s *Server) Stop() {
	s.logger.Warnw("Forcing gRPC server to stop")
	s.grpcServer.Stop()
}
