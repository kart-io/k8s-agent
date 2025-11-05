package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/agent-manager/app/options"
	"github.com/kart-io/k8s-agent/internal/agent-manager/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// Execute runs the agent-manager command.
func Execute() {
	// Create configuration options
	opts := options.NewServerOptions()

	// Create application instance
	app := &AgentManagerApp{}

	// Use the simplified RunWithBootstrap to run the application
	commonapp.RunWithBootstrap(
		app,
		opts,
		commonapp.Config{
			Use:       "agent-manager",
			Short:     "Agent Manager Service",
			Long:      "Agent Manager Service manages k8s agents across multiple clusters",
			EnvPrefix: "AGENT_MANAGER",
		},
		app.registerComponents,
	)
}

// AgentManagerApp implements commonapp.Application interface.
type AgentManagerApp struct {
	opts   *options.ServerOptions // 直接使用ServerOptions
	logger core.Logger

	// Component initializers
	dbInit         *initializers.DatabaseInitializer
	redisInit      *initializers.RedisInitializer
	registryInit   *initializers.RegistryInitializer
	natsInit       *initializers.NATSInitializer
	dispatcherInit *initializers.DispatcherInitializer
	httpInit       *initializers.HTTPServerInitializer
	grpcInit       *initializers.GRPCServerInitializer
	healthInit     *pkginitializers.HealthCheckInitializer
}

// Name returns the application name.
func (a *AgentManagerApp) Name() string {
	return "Agent Manager"
}

// Initialize initializes the application.
func (a *AgentManagerApp) Initialize(ctx context.Context, opts commonapp.Options) error {
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
func (a *AgentManagerApp) Run(ctx context.Context) error {
	// The bootstrap framework handles running all servers
	// This method can be used for additional application logic if needed
	<-ctx.Done()
	return nil
}

// Shutdown gracefully shuts down the application.
func (a *AgentManagerApp) Shutdown(ctx context.Context) error {
	// Bootstrap framework handles component shutdown
	// This method can be used for additional cleanup if needed
	return nil
}

// registerComponents registers all component initializers with bootstrap.
func (a *AgentManagerApp) registerComponents(bs *bootstrap.Bootstrap) error {
	// Use Wire to automatically inject all dependencies
	components, err := InitializeAgentManagerComponents(a.opts)
	if err != nil {
		return fmt.Errorf("failed to initialize components: %w", err)
	}

	// Register components to Bootstrap
	bs.Register(components.DB)
	bs.Register(components.Redis)
	bs.Register(components.Registry)
	bs.Register(components.NATS)
	bs.Register(components.Dispatcher)
	bs.Register(components.HTTP)

	// Register gRPC if enabled
	if a.opts.GRPC.Enable {
		a.grpcInit = initializers.NewGRPCServerInitializer(
			a.opts,
			a.logger,
			components.Registry,
			components.Dispatcher,
			components.DB,
		)
		bs.Register(a.grpcInit)
	}

	bs.Register(components.Health)

	// Save references for app
	a.dbInit = components.DB
	a.redisInit = components.Redis
	a.registryInit = components.Registry
	a.natsInit = components.NATS
	a.dispatcherInit = components.Dispatcher
	a.httpInit = components.HTTP
	a.healthInit = components.Health

	return nil
}
