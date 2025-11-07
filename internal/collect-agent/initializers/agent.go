package initializers

import (
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"context"

	"github.com/kart-io/k8s-agent/internal/collect-agent/agent"
	"github.com/kart-io/k8s-agent/internal/collect-agent/types"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
)

// AgentInitializer handles agent initialization and lifecycle.
type AgentInitializer struct {
	cfg      *commonapp.StandardOptions
	logger   core.Logger
	agent    *agent.Agent
	stopChan chan struct{}
}

// NewAgentInitializer creates a new agent initializer.
func NewAgentInitializer(cfg *commonapp.StandardOptions, logger core.Logger) *AgentInitializer {
	return &AgentInitializer{
		cfg:      cfg,
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

// Name returns the initializer name.
func (a *AgentInitializer) Name() string {
	return "collect-agent"
}

// Priority returns initialization priority (lower runs first).
func (a *AgentInitializer) Priority() int {
	return bootstrap.PriorityApplication
}

// Initialize creates the agent instance.
func (a *AgentInitializer) Initialize(ctx context.Context) error {
	a.logger.Infow("Initializing collect agent")

	// Convert StandardOptions to AgentConfig
	agentConfig := &types.AgentConfig{
		ClusterID:         a.cfg.Agent.ClusterID,
		ClusterName:       a.cfg.Agent.ClusterName,
		CentralEndpoint:   a.cfg.Agent.CentralEndpoint,
		ReconnectDelay:    a.cfg.Agent.ReconnectDelay,
		HeartbeatInterval: a.cfg.Agent.HeartbeatInterval,
		MetricsInterval:   a.cfg.Agent.MetricsInterval,
		BufferSize:        a.cfg.Agent.BufferSize,
		MaxRetries:        a.cfg.Agent.MaxRetries,
		LogLevel:          a.cfg.Logging.Level,
		EnableMetrics:     a.cfg.Agent.EnableMetrics,
		EnableEvents:      a.cfg.Agent.EnableEvents,
		HealthPort:        a.cfg.Health.Port,
	}

	// Create agent instance
	agentInstance, err := agent.New(agentConfig, a.logger)
	if err != nil {
		return err
	}

	a.agent = agentInstance
	a.logger.Infow("Collect agent initialized successfully",
		"cluster_id", a.cfg.Agent.ClusterID,
		"cluster_name", a.cfg.Agent.ClusterName,
		"central_endpoint", a.cfg.Agent.CentralEndpoint,
	)
	return nil
}

// Run starts the agent.
func (a *AgentInitializer) Run(ctx context.Context) error {
	if a.agent == nil {
		return nil
	}

	a.logger.Infow("Starting collect agent")

	// Start agent in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := a.agent.Start(ctx); err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	// Wait for context cancellation or error
	select {
	case err := <-errChan:
		if err != nil {
			a.logger.Errorw("Agent failed", "error", err)
			return err
		}
		return nil
	case <-ctx.Done():
		a.logger.Infow("Agent context cancelled")
		return nil
	}
}

// Close stops the agent.
func (a *AgentInitializer) Close(ctx context.Context) error {
	if a.agent != nil {
		a.logger.Infow("Stopping collect agent")
		// Agent will be stopped by context cancellation in Run()
	}
	return nil
}

// Agent returns the agent instance.
func (a *AgentInitializer) Agent() *agent.Agent {
	return a.agent
}

