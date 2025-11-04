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
	// 直接使用已有的opts，不需要重新创建

	// 1. Database (priority 300)
	a.dbInit = initializers.NewDatabaseInitializer(a.opts, a.logger)
	bs.Register(a.dbInit)

	// 2. Redis (priority 400)
	a.redisInit = initializers.NewRedisInitializer(a.opts, a.logger)
	bs.Register(a.redisInit)

	// 3. Registry (priority 450 - after Database and Redis)
	a.registryInit = initializers.NewRegistryInitializer(
		a.opts,
		a.logger,
		a.dbInit,
		a.redisInit,
	)
	bs.Register(a.registryInit)

	// 4. NATS (priority 500)
	a.natsInit = initializers.NewNATSInitializer(
		a.opts,
		a.logger,
		a.registryInit,
		a.dbInit,
		a.redisInit,
	)
	bs.Register(a.natsInit)

	// 5. Dispatcher (priority 600 - after NATS)
	a.dispatcherInit = initializers.NewDispatcherInitializer(
		a.opts,
		a.logger,
		a.dbInit,
		a.redisInit,
		a.registryInit,
		a.natsInit,
	)
	bs.Register(a.dispatcherInit)

	// 6. HTTP Server (priority 1000)
	a.httpInit = initializers.NewHTTPServerInitializer(
		a.opts,
		a.logger,
		a.registryInit,
		a.dispatcherInit,
		a.dbInit,
		a.redisInit,
		a.natsInit,
	)
	bs.Register(a.httpInit)

	// 7. gRPC Server (priority 1100)
	if a.opts.GRPC.Enable {
		a.grpcInit = initializers.NewGRPCServerInitializer(
			a.opts,
			a.logger,
			a.registryInit,
			a.dispatcherInit,
			a.dbInit,
		)
		bs.Register(a.grpcInit)
	}

	// 8. Health Check Server (priority 2000)
	a.healthInit = pkginitializers.NewHealthCheckInitializer(a.opts.Health, a.logger)
	bs.Register(a.healthInit)

	return nil
}
