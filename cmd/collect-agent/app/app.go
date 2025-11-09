package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/common/health"
	"github.com/kart-io/k8s-agent/internal/collect-agent/agent"
	"github.com/kart-io/k8s-agent/internal/collect-agent/types"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/logger/core"
)

const (
	// UserAgent is the User-Agent string for the collect-agent service.
	UserAgent = "aetherius-collect-agent"
)

// Execute runs the collect-agent command.
func Execute() {
	opts := commonapp.NewStandardOptions("Collect Agent", UserAgent).
		WithAgent()

	app := &CollectAgentApp{}

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
// This is the ULTRA-SIMPLE pattern: no Bootstrap, no Wire, direct initialization.
type CollectAgentApp struct {
	opts         *commonapp.StandardOptions
	logger       core.Logger
	agent        *agent.Agent
	healthServer health.Server
}

// Name returns the application name.
func (a *CollectAgentApp) Name() string {
	return "Collect Agent"
}

// Initialize initializes the application components directly.
func (a *CollectAgentApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	a.opts = opts.(*commonapp.StandardOptions)

	// Initialize logger
	logger, err := a.opts.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Infow("Initializing Collect Agent",
		"cluster_id", a.opts.Agent.ClusterID,
		"central_endpoint", a.opts.Agent.CentralEndpoint,
		"health_port", a.opts.Health.Port,
	)

	// Create agent instance
	agentConfig := &types.AgentConfig{
		ClusterID:         a.opts.Agent.ClusterID,
		ClusterName:       a.opts.Agent.ClusterName,
		CentralEndpoint:   a.opts.Agent.CentralEndpoint,
		ReconnectDelay:    a.opts.Agent.ReconnectDelay,
		HeartbeatInterval: a.opts.Agent.HeartbeatInterval,
		MetricsInterval:   a.opts.Agent.MetricsInterval,
		BufferSize:        a.opts.Agent.BufferSize,
		MaxRetries:        a.opts.Agent.MaxRetries,
		LogLevel:          a.opts.Logging.Level,
		EnableMetrics:     a.opts.Agent.EnableMetrics,
		EnableEvents:      a.opts.Agent.EnableEvents,
		HealthPort:        a.opts.Health.Port,
	}

	a.agent, err = agent.New(agentConfig, logger)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	// Create health check server
	a.healthServer = health.NewHTTPServer(a.opts.Health, logger)
	if err := a.healthServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start health server: %w", err)
	}

	return nil
}

// Run runs the application.
func (a *CollectAgentApp) Run(ctx context.Context) error {
	a.logger.Info("Starting Collect Agent")

	// Start agent (blocks until context cancelled)
	return a.agent.Start(ctx)
}

// Shutdown gracefully shuts down the application.
func (a *CollectAgentApp) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down Collect Agent")

	// Shutdown health check server
	if a.healthServer != nil {
		if err := a.healthServer.Stop(ctx); err != nil {
			a.logger.Warnw("Health server shutdown error", "error", err)
		}
	}

	// Agent shutdown is handled by context cancellation in Start()
	a.logger.Info("Collect Agent shutdown complete")
	return nil
}
