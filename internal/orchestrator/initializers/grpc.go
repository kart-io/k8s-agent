package initializers

import (
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"context"
	"fmt"

	"google.golang.org/grpc"

	
	commonserver "github.com/kart-io/k8s-agent/common/server"
	"github.com/kart-io/k8s-agent/internal/orchestrator/service"
	orchestratorv1 "github.com/kart-io/k8s-agent/pkg/api/orchestrator/v1"
	commoninitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// GRPCServerInitializer initializes the gRPC server
// 使用 common/initializers 的标准 gRPC 服务器初始化器.
type GRPCServerInitializer struct {
	opts         *commonapp.StandardOptions
	logger       core.Logger
	workflowInit *WorkflowInitializer
	dbInit       *DatabaseInitializer

	// 使用标准初始化器
	standardInit    *commoninitializers.GRPCServerInitializer
	workflowService *service.WorkflowServiceServer
}

// NewGRPCServerInitializer creates a new gRPC server initializer.
func NewGRPCServerInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	workflowInit *WorkflowInitializer,
	dbInit *DatabaseInitializer,
) *GRPCServerInitializer {
	return &GRPCServerInitializer{
		opts:         opts,
		logger:       logger.With("component", "grpc-server-initializer"),
		workflowInit: workflowInit,
		dbInit:       dbInit,
	}
}

// Name returns the initializer name.
func (i *GRPCServerInitializer) Name() string {
	return "GRPCServer"
}

// Priority returns the initialization priority (should be after workflow engine).
func (i *GRPCServerInitializer) Priority() int {
	return 700 // After workflow engine (550) and strategy (600)
}

// Initialize sets up the gRPC server.
func (i *GRPCServerInitializer) Initialize(ctx context.Context) error {
	// Check if gRPC is enabled
	if !i.opts.GRPC.Enable {
		i.logger.Infow("gRPC server is disabled, skipping initialization")
		return nil
	}

	i.logger.Infow("Initializing gRPC server",
		"host", i.opts.GRPC.Host,
		"port", i.opts.GRPC.Port,
	)

	// Get engine and store from other initializers
	engine := i.workflowInit.Engine()
	if engine == nil {
		return fmt.Errorf("workflow engine not initialized")
	}

	store := i.dbInit.Store()
	if store == nil {
		return fmt.Errorf("database store not initialized")
	}

	// Create shared workflow service
	i.workflowService = service.NewWorkflowServiceServer(engine, store, i.logger)

	// 使用标准 gRPC 服务器初始化器
	i.standardInit = commoninitializers.NewGRPCServerInitializer(
		&commoninitializers.GRPCServerConfig{
			Name:     "OrchestratorGRPC",
			Priority: i.Priority(),
			Config:   i.opts.GRPC,
			ServiceRegister: func(grpcServer *grpc.Server) error {
				// 注册 workflow 服务
				orchestratorv1.RegisterWorkflowServiceServer(grpcServer, i.workflowService)
				i.logger.Infow("Registered WorkflowService to gRPC server")
				return nil
			},
		},
		i.logger,
	)

	// 调用标准初始化器的 Initialize
	return i.standardInit.Initialize(ctx)
}

// GetWorkflowService returns the shared workflow service instance.
func (i *GRPCServerInitializer) GetWorkflowService() *service.WorkflowServiceServer {
	return i.workflowService
}

// Close stops the gRPC server.
func (i *GRPCServerInitializer) Close(ctx context.Context) error {
	if i.standardInit == nil {
		return nil
	}

	return i.standardInit.Close(ctx)
}

// GetServer returns the server instance (implements ServerProvider).
func (i *GRPCServerInitializer) GetServer() commonserver.Server {
	if i.standardInit == nil {
		return nil
	}
	return i.standardInit.GetServer()
}
