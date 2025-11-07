package app

import (
	"context"
	"fmt"

	// Removed: options package
	"github.com/kart-io/k8s-agent/internal/collect-agent/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

const (
	// UserAgent is the User-Agent string for the collect-agent service.
	UserAgent = "aetherius-collect-agent"
)

// Execute runs the collect-agent command.
func Execute() {
	// Create configuration options
	opts := commonapp.NewStandardOptions("Collect Agent", UserAgent).
		WithNATS().
		WithAgent()

	// Create application instance
	app := &CollectAgentApp{}

	// Use the simplified RunWithBootstrap to run the application
	commonapp.RunWithBootstrap(
		app,
		opts,
		commonapp.Config{
			Use:       "collect-agent",
			Short:     "Collect Agent",
			Long:      "Collect Agent monitors K8s cluster events and collects metrics from edge clusters",
			EnvPrefix: "COLLECT_AGENT",
		},
		app.registerComponents,
	)
}

// CollectAgentApp implements commonapp.Application interface.
type CollectAgentApp struct {
	opts   *commonapp.StandardOptions // 使用 ServerOptions
	logger core.Logger

	// Component initializers
	agentInit  *initializers.AgentInitializer
	healthInit *pkginitializers.HealthCheckInitializer
}

// Name returns the application name.
func (a *CollectAgentApp) Name() string {
	return "Collect Agent"
}

// Initialize initializes the application.
func (a *CollectAgentApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	// 保存 ServerOptions
	a.opts = opts.(*commonapp.StandardOptions)

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

	return nil
}

// Run runs the application.
func (a *CollectAgentApp) Run(ctx context.Context) error {
	// The bootstrap framework handles running all components
	// This method can be used for additional application logic if needed
	<-ctx.Done()
	return nil
}

// Shutdown gracefully shuts down the application.
func (a *CollectAgentApp) Shutdown(ctx context.Context) error {
	// Bootstrap framework handles component shutdown
	// This method can be used for additional cleanup if needed
	return nil
}

// registerComponents registers all component initializers with bootstrap.
func (a *CollectAgentApp) registerComponents(bs *bootstrap.Bootstrap) error {
	// Use Wire to automatically inject all dependencies
	components, err := InitializeCollectAgentComponents(a.opts)
	if err != nil {
		return fmt.Errorf("failed to initialize components: %w", err)
	}

	// Register components to Bootstrap
	bs.Register(components.Agent)
	bs.Register(components.Health)

	// Save references for app
	a.agentInit = components.Agent
	a.healthInit = components.Health

	return nil
}
