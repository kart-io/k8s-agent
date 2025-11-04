package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/orchestrator/app/options"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
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
	// 直接使用已有的opts，不需要重新创建

	// 1. Database (priority 300)
	a.dbInit = initializers.NewDatabaseInitializer(a.opts, a.logger)
	bs.Register(a.dbInit)

	// 2. Redis (priority 400)
	a.redisInit = initializers.NewRedisInitializer(a.opts, a.logger)
	bs.Register(a.redisInit)

	// 3. NATS (priority 500)
	a.natsInit = initializers.NewNATSInitializer(a.opts, a.logger)
	bs.Register(a.natsInit)

	// 4. Workflow Engine (priority 550 - after Database and Redis)
	a.workflowInit = initializers.NewWorkflowInitializer(
		a.opts,
		a.logger,
		a.dbInit,
		a.redisInit,
	)
	bs.Register(a.workflowInit)

	// 5. Strategy Manager (priority 600 - after Workflow)
	a.strategyInit = initializers.NewStrategyInitializer(
		a.opts,
		a.logger,
		a.dbInit,
		a.workflowInit,
	)
	bs.Register(a.strategyInit)

	// 6. Subscriber (priority 650 - after Strategy)
	a.subInit = initializers.NewSubscriberInitializer(
		a.opts,
		a.logger,
		a.natsInit,
		a.strategyInit,
	)
	bs.Register(a.subInit)

	// 7. gRPC Server (priority 700 - after Workflow and Strategy)
	a.grpcInit = initializers.NewGRPCServerInitializer(
		a.opts,
		a.logger,
		a.workflowInit,
		a.dbInit,
	)
	bs.Register(a.grpcInit)

	// 8. HTTP Server with gRPC-Gateway (priority 800 - after gRPC)
	// HTTP requests will be automatically converted to gRPC calls using the same workflow service!
	a.httpInit = initializers.NewHTTPServerInitializer(
		a.opts,
		a.logger,
		a.grpcInit, // Pass gRPC init to get shared service
	)
	bs.Register(a.httpInit)

	// 9. Health Check Server (priority 2000)
	healthOpts := commonoptions.NewHealthOptions()
	a.healthInit = pkginitializers.NewHealthCheckInitializer(healthOpts, a.logger)
	bs.Register(a.healthInit)

	return nil
}
