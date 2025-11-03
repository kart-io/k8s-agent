package initializers

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/common/bootstrap"
	"github.com/kart-io/k8s-agent/common/options"
	commonserver "github.com/kart-io/k8s-agent/common/server"
	grpcserver "github.com/kart-io/k8s-agent/common/server/grpc"
	"github.com/kart-io/logger/core"
	"google.golang.org/grpc"
)

// GRPCServerConfig holds configuration for the gRPC server initializer.
type GRPCServerConfig struct {
	Name            string
	Priority        int
	Config          *options.GRPCOptions
	ServiceRegister func(*grpc.Server) error
}

// GRPCServerInitializer initializes a standard gRPC server and integrates it with the bootstrap lifecycle.
// It implements bootstrap.Initializer and bootstrap.ServerProvider.
type GRPCServerInitializer struct {
	config *GRPCServerConfig
	logger core.Logger
	server commonserver.Server
}

// NewGRPCServerInitializer creates a new GRPCServerInitializer.
func NewGRPCServerInitializer(config *GRPCServerConfig, logger core.Logger) *GRPCServerInitializer {
	// Set default name and priority if not provided
	if config.Name == "" {
		config.Name = "grpc-server"
	}
	if config.Priority == 0 {
		config.Priority = bootstrap.PriorityGRPC
	}
	return &GRPCServerInitializer{
		config: config,
		logger: logger,
	}
}

// Name returns the initializer name.
func (i *GRPCServerInitializer) Name() string {
	return i.config.Name
}

// Priority returns the initializer priority.
func (i *GRPCServerInitializer) Priority() int {
	return i.config.Priority
}

// Initialize creates and configures the gRPC server.
func (i *GRPCServerInitializer) Initialize(ctx context.Context) error {
	i.logger.Infow("Initializing gRPC server", "name", i.config.Name)

	// Create gRPC server from config
	grpcServer, err := grpcserver.NewStandardGRPCServerFromConfig(i.logger, i.config.Config)
	if err != nil {
		return fmt.Errorf("failed to create gRPC server for %s: %w", i.config.Name, err)
	}
	i.server = grpcServer

	// Register services
	if i.config.ServiceRegister != nil {
		if err := i.config.ServiceRegister(grpcServer.GetServer()); err != nil {
			return fmt.Errorf("failed to register gRPC services for %s: %w", i.config.Name, err)
		}
		i.logger.Infow("gRPC services registered", "name", i.config.Name)
	}

	i.logger.Infow("gRPC server initialized successfully", "name", i.config.Name)
	return nil
}

// GetServer returns the server instance to be managed by bootstrap.
// This implements the bootstrap.ServerProvider interface.
func (i *GRPCServerInitializer) GetServer() commonserver.Server {
	return i.server
}

// Close is a no-op for this initializer because the server's lifecycle
// is managed by the bootstrap's Run method, which calls GracefulStop.
func (i *GRPCServerInitializer) Close(ctx context.Context) error {
	return nil
}
