package initializers

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/reasoning/app/options"
	"github.com/kart-io/k8s-agent/internal/reasoning/analyzer"
	grpcserver "github.com/kart-io/k8s-agent/internal/reasoning/grpc"
	"github.com/kart-io/k8s-agent/internal/reasoning/service"
	"github.com/kart-io/logger/core"
)

// GRPCServerInitializer initializes the gRPC server.
type GRPCServerInitializer struct {
	opts    *options.ServerOptions
	logger  core.Logger
	llmInit *LLMInitializer

	server           *grpcserver.Server
	reasoningService *service.ReasoningServiceServer
}

// NewGRPCServerInitializer creates a new gRPC server initializer.
func NewGRPCServerInitializer(
	opts *options.ServerOptions,
	logger core.Logger,
	llmInit *LLMInitializer,
) *GRPCServerInitializer {
	return &GRPCServerInitializer{
		opts:    opts,
		logger:  logger.With("component", "grpc-server-initializer"),
		llmInit: llmInit,
	}
}

// Name returns the initializer name.
func (i *GRPCServerInitializer) Name() string {
	return "GRPCServer"
}

// Priority returns the initialization priority (should be after LLM clients).
func (i *GRPCServerInitializer) Priority() int {
	return 450 // After LLM (400), before HTTP (500)
}

// Initialize sets up and starts the gRPC server.
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

	// Get LLM clients and create analyzer
	llmClients := i.llmInit.GetClients()
	cfg := i.opts.Config()
	rootCauseAnalyzer := analyzer.NewRootCauseAnalyzer(cfg, llmClients)

	// Create shared reasoning service
	i.reasoningService = service.NewReasoningServiceServer(rootCauseAnalyzer, i.logger)

	// Create gRPC server with shared service
	grpcOpts := &grpcserver.ServerOptions{
		Host:             i.opts.GRPC.Host,
		Port:             i.opts.GRPC.Port,
		ReasoningService: i.reasoningService,
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

// Shutdown stops the gRPC server.
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

// GetServer returns the gRPC server instance.
func (i *GRPCServerInitializer) GetServer() *grpcserver.Server {
	return i.server
}

// GetReasoningService returns the shared reasoning service instance.
func (i *GRPCServerInitializer) GetReasoningService() *service.ReasoningServiceServer {
	return i.reasoningService
}
