package initializers

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/reasoning/app/options"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/internal/reasoning/analyzer"
	"github.com/kart-io/k8s-agent/internal/reasoning/config"
	"github.com/kart-io/k8s-agent/internal/reasoning/handler"
	"github.com/kart-io/k8s-agent/internal/reasoning/server"
	"github.com/kart-io/logger/core"
)

// UnifiedServerInitializer initializes the unified server (both gRPC and HTTP using Kratos).
// This follows the OneX architecture pattern where a single handler serves both protocols.
type UnifiedServerInitializer struct {
	opts    *options.ServerOptions
	logger  core.Logger
	llmInit *LLMInitializer

	server  *server.Server
	handler *handler.ReasoningHandler
}

// NewUnifiedServerInitializer creates a new unified server initializer.
func NewUnifiedServerInitializer(
	opts *options.ServerOptions,
	logger core.Logger,
	llmInit *LLMInitializer,
) *UnifiedServerInitializer {
	return &UnifiedServerInitializer{
		opts:    opts,
		logger:  logger.With("component", "unified-server-initializer"),
		llmInit: llmInit,
	}
}

// Name returns the initializer name.
func (i *UnifiedServerInitializer) Name() string {
	return "UnifiedServer"
}

// Priority returns the initialization priority (should be after LLM clients).
func (i *UnifiedServerInitializer) Priority() int {
	return 450 // After LLM (400)
}

// Initialize sets up and starts the unified server (both gRPC and HTTP).
func (i *UnifiedServerInitializer) Initialize(ctx context.Context) error {
	i.logger.Infow("Initializing unified server (gRPC + HTTP)",
		"grpc_host", i.opts.GRPC.Host,
		"grpc_port", i.opts.GRPC.Port,
		"http_host", i.opts.Server.Host,
		"http_port", i.opts.Server.Port,
	)

	// Get LLM clients and create analyzer
	llmClients := i.llmInit.GetClients()
	cfg := config.NewConfigFromOptions(i.opts)
	rootCauseAnalyzer := analyzer.NewRootCauseAnalyzer(cfg, llmClients)

	// Create unified handler (implements both gRPC and HTTP interfaces)
	i.handler = handler.NewReasoningHandler(rootCauseAnalyzer, i.logger)

	// Prepare server configuration using common/options types
	httpOpts := &commonoptions.ServerOptions{
		Host:         i.opts.Server.Host,
		Port:         i.opts.Server.Port,
		Mode:         i.opts.Server.Mode,
		ReadTimeout:  i.opts.Server.ReadTimeout,
		WriteTimeout: i.opts.Server.WriteTimeout,
		IdleTimeout:  i.opts.Server.IdleTimeout,
		GracefulStop: i.opts.Server.GracefulStop,
	}

	grpcOpts := &commonoptions.GRPCOptions{
		Enable:                i.opts.GRPC.Enable,
		Host:                  i.opts.GRPC.Host,
		Port:                  i.opts.GRPC.Port,
		MaxRecvMsgSize:        i.opts.GRPC.MaxRecvMsgSize,
		MaxSendMsgSize:        i.opts.GRPC.MaxSendMsgSize,
		ConnectionTimeout:     i.opts.GRPC.ConnectionTimeout,
		KeepAliveTime:         i.opts.GRPC.KeepAliveTime,
		KeepAliveTimeout:      i.opts.GRPC.KeepAliveTimeout,
		MaxConnectionIdle:     i.opts.GRPC.MaxConnectionIdle,
		MaxConnectionAge:      i.opts.GRPC.MaxConnectionAge,
		MaxConnectionAgeGrace: i.opts.GRPC.MaxConnectionAgeGrace,
		EnableReflection:      i.opts.GRPC.EnableReflection,
		EnableHealthCheck:     i.opts.GRPC.EnableHealthCheck,
	}

	// Create server configuration
	serverCfg := &server.Config{
		HTTP:    httpOpts,
		GRPC:    grpcOpts,
		Handler: i.handler,
	}

	// Create unified server with both gRPC and HTTP transports
	unifiedServer, err := server.NewServer(serverCfg, i.logger)
	if err != nil {
		return fmt.Errorf("failed to create unified server: %w", err)
	}

	i.server = unifiedServer

	// Start unified server in background (starts both gRPC and HTTP)
	go func() {
		i.logger.Infow("Starting unified server in background")
		if err := unifiedServer.Start(ctx); err != nil {
			i.logger.Errorw("Unified server error", "error", err)
		}
	}()

	i.logger.Infow("Unified server initialized successfully")

	return nil
}

// Shutdown stops the unified server (both gRPC and HTTP).
func (i *UnifiedServerInitializer) Shutdown(ctx context.Context) error {
	if i.server == nil {
		return nil
	}

	i.logger.Infow("Shutting down unified server")

	if err := i.server.Stop(); err != nil {
		return fmt.Errorf("failed to shutdown unified server: %w", err)
	}

	i.logger.Infow("Unified server shutdown complete")
	return nil
}

// GetServer returns the unified server instance.
func (i *UnifiedServerInitializer) GetServer() *server.Server {
	return i.server
}

// GetHandler returns the unified handler instance.
func (i *UnifiedServerInitializer) GetHandler() *handler.ReasoningHandler {
	return i.handler
}
