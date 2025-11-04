package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/collect-agent/app/options"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/logger/core"
)

// Execute runs the collect-agent command.
func Execute() {
	// Create configuration options
	opts := options.NewServerOptions()

	// Create application instance
	app := &CollectAgentApp{}

	// Use the simplified framework (no bootstrap needed for simple services)
	commonapp.Run(
		app,
		opts,
		commonapp.Config{
			Use:       "collect-agent",
			Short:     "Collect Agent",
			Long:      "Collect Agent monitors K8s cluster events and collects metrics from edge clusters",
			EnvPrefix: "COLLECT_AGENT",
		},
	)
}

// CollectAgentApp implements commonapp.Application interface.
type CollectAgentApp struct {
	opts   *options.ServerOptions // 使用 ServerOptions
	logger core.Logger
	server *CollectAgentService
}

// Name returns the application name.
func (a *CollectAgentApp) Name() string {
	return "Collect Agent"
}

// Initialize initializes the application.
func (a *CollectAgentApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	// 保存 ServerOptions
	a.opts = opts.(*options.ServerOptions)

	// Initialize logger
	logger, err := a.opts.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Infow("Starting Aetherius Collect Agent",
		"cluster_id", a.opts.Agent.ClusterID,
		"central_endpoint", a.opts.Agent.CentralEndpoint,
		"health_port", a.opts.Health.Port,
	)

	// Create server
	srv, err := NewServer(a.opts, logger)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	a.server = srv

	return nil
}

// Run runs the application.
func (a *CollectAgentApp) Run(ctx context.Context) error {
	// Start server
	return a.server.Run(ctx)
}

// Shutdown gracefully shuts down the application.
func (a *CollectAgentApp) Shutdown(ctx context.Context) error {
	if a.logger != nil {
		_ = a.logger.Flush() // Best effort flush, ignore errors during shutdown
	}
	return nil
}
