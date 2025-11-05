package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/orchestrator/app/options"
	"github.com/kart-io/k8s-agent/internal/orchestrator/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// Execute runs the orchestrator command.
func Execute() {
	// Create configuration options
	opts := options.NewServerOptions()

	// Create application instance
	app := &OrchestratorApp{}

	// Use the simplified RunWithBootstrap to run the application
	commonapp.RunWithBootstrap(
		app,
		opts,
		commonapp.Config{
			Use:       "orchestrator",
			Short:     "Orchestrator Service",
			Long:      "Orchestrator Service manages workflow execution for automated diagnosis and remediation",
			EnvPrefix: "ORCHESTRATOR",
		},
		app.registerComponents,
	)
}

// OrchestratorApp implements commonapp.Application interface.
type OrchestratorApp struct {
	opts   *options.ServerOptions // 直接使用ServerOptions
	logger core.Logger

	// Component initializers
	dbInit       *initializers.DatabaseInitializer
	redisInit    *initializers.RedisInitializer
	natsInit     *initializers.NATSInitializer
	workflowInit *initializers.WorkflowInitializer
	strategyInit *initializers.StrategyInitializer
	subInit      *initializers.SubscriberInitializer
	grpcInit     *initializers.GRPCServerInitializer
	httpInit     *initializers.HTTPServerInitializer
	healthInit   *pkginitializers.HealthCheckInitializer
}

// Name returns the application name.
func (a *OrchestratorApp) Name() string {
	return "Orchestrator Service"
}

// Initialize initializes the application.
func (a *OrchestratorApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	// 直接保存ServerOptions，不需要转换
	a.opts = opts.(*options.ServerOptions)

	// Initialize logger
	logger, err := a.opts.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	return nil
}

// Run runs the application.
func (a *OrchestratorApp) Run(ctx context.Context) error {
	// The bootstrap framework handles running all servers
	// This method can be used for additional application logic if needed
	<-ctx.Done()
	return nil
}

// Shutdown gracefully shuts down the application.
func (a *OrchestratorApp) Shutdown(ctx context.Context) error {
	// Bootstrap framework handles component shutdown
	// This method can be used for additional cleanup if needed
	return nil
}

// registerComponents registers all component initializers with bootstrap.
func (a *OrchestratorApp) registerComponents(bs *bootstrap.Bootstrap) error {
	// Use Wire to automatically inject all dependencies
	components, err := InitializeOrchestratorComponents(a.opts)
	if err != nil {
		return fmt.Errorf("failed to initialize components: %w", err)
	}

	// Register components to Bootstrap in dependency order
	bs.Register(components.DB)
	bs.Register(components.Redis)
	bs.Register(components.NATS)
	bs.Register(components.Workflow)
	bs.Register(components.Strategy)
	bs.Register(components.Subscriber)
	bs.Register(components.GRPC)
	bs.Register(components.HTTP)
	bs.Register(components.Health)

	// Save references for app
	a.dbInit = components.DB
	a.redisInit = components.Redis
	a.natsInit = components.NATS
	a.workflowInit = components.Workflow
	a.strategyInit = components.Strategy
	a.subInit = components.Subscriber
	a.grpcInit = components.GRPC
	a.httpInit = components.HTTP
	a.healthInit = components.Health

	return nil
}
