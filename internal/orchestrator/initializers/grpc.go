package initializers

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/orchestrator/app/options"
	grpcserver "github.com/kart-io/k8s-agent/internal/orchestrator/grpc"
	"github.com/kart-io/k8s-agent/internal/orchestrator/service"
	"github.com/kart-io/logger/core"
)

// GRPCServerInitializer initializes the gRPC server
type GRPCServerInitializer struct {
	opts         *options.ServerOptions
	logger       core.Logger
	workflowInit *WorkflowInitializer
	dbInit       *DatabaseInitializer

	server          *grpcserver.Server
	workflowService *service.WorkflowServiceServer
}

// NewGRPCServerInitializer creates a new gRPC server initializer
func NewGRPCServerInitializer(
	opts *options.ServerOptions,
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

// Name returns the initializer name
func (i *GRPCServerInitializer) Name() string {
	return "GRPCServer"
}

// Priority returns the initialization priority (should be after workflow engine)
func (i *GRPCServerInitializer) Priority() int {
	return 700 // After workflow engine (550) and strategy (600)
}

// Initialize sets up and starts the gRPC server
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

	// Create gRPC server with shared service
	grpcOpts := &grpcserver.ServerOptions{
		Host:            i.opts.GRPC.Host,
		Port:            i.opts.GRPC.Port,
		WorkflowService: i.workflowService, // Pass shared service
	}

	server, err := grpcserver.NewServer(grpcOpts, i.logger)
	if err != nil {
		return fmt.Errorf("failed to create gRPC server: %w", err)
	}

	i.server = server

	// Start gRPC server in background
	go func() {
		i.logger.Infow("Starting gRPC server in background",
			"address", server.Address(),
		)
		if err := server.Start(ctx); err != nil {
			i.logger.Errorw("gRPC server error",
				"error", err,
			)
		}
	}()

	i.logger.Infow("gRPC server initialized successfully",
		"address", server.Address(),
	)

	return nil
}

// GetWorkflowService returns the shared workflow service instance
func (i *GRPCServerInitializer) GetWorkflowService() *service.WorkflowServiceServer {
	return i.workflowService
}

// Shutdown stops the gRPC server
func (i *GRPCServerInitializer) Shutdown(ctx context.Context) error {
	if i.server == nil {
		return nil
	}

	i.logger.Infow("Shutting down gRPC server")

	if err := i.server.Stop(); err != nil {
		return fmt.Errorf("failed to shutdown gRPC server: %w", err)
	}

	i.logger.Infow("gRPC server shutdown complete")
	return nil
}

// GetServer returns the gRPC server instance
func (i *GRPCServerInitializer) GetServer() *grpcserver.Server {
	return i.server
}
