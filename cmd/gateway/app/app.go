package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/gateway/app/options"
	"github.com/kart-io/k8s-agent/internal/gateway/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// Execute runs the gateway command.
func Execute() {
	// Create configuration options
	opts := options.NewServerOptions()

	// Create application instance
	app := &GatewayApp{}

	// Use the simplified RunWithBootstrap to run the application
	commonapp.RunWithBootstrap(
		app,
		opts,
		commonapp.Config{
			Use:       "gateway",
			Short:     "Gateway Service",
			Long:      "Gateway Service provides API gateway with proxy capabilities",
			EnvPrefix: "GATEWAY",
		},
		app.registerComponents,
	)
}

// GatewayApp implements commonapp.Application interface.
type GatewayApp struct {
	opts   *options.ServerOptions // 使用 ServerOptions
	logger core.Logger

	// Component initializers
	redisInit  *initializers.RedisInitializer
	httpInit   *initializers.HTTPServerInitializer
	healthInit *pkginitializers.HealthCheckInitializer
}

// Name returns the application name.
func (a *GatewayApp) Name() string {
	return "Gateway Service"
}

// Initialize initializes the application.
func (a *GatewayApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	// 保存 ServerOptions
	a.opts = opts.(*options.ServerOptions)

	// Initialize logger
	logger, err := a.opts.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Info("Starting Gateway Service...")

	return nil
}

// Run runs the application.
func (a *GatewayApp) Run(ctx context.Context) error {
	// The bootstrap framework handles running all servers
	// This method can be used for additional application logic if needed
	<-ctx.Done()
	return nil
}

// Shutdown gracefully shuts down the application.
func (a *GatewayApp) Shutdown(ctx context.Context) error {
	// Bootstrap framework handles component shutdown
	// This method can be used for additional cleanup if needed
	return nil
}

// registerComponents registers all component initializers with bootstrap.
func (a *GatewayApp) registerComponents(bs *bootstrap.Bootstrap) error {
	// 1. Redis (priority 400)
	a.redisInit = initializers.NewRedisInitializer(a.opts, a.logger)
	bs.Register(a.redisInit)

	// 2. HTTP Server (priority 500)
	a.httpInit = initializers.NewHTTPServerInitializer(
		a.opts,
		a.logger,
		a.redisInit,
	)
	bs.Register(a.httpInit)

	// 3. Health Check (priority 2000)
	a.healthInit = pkginitializers.NewHealthCheckInitializer(
		a.opts.Health,
		a.logger,
	)
	bs.Register(a.healthInit)

	return nil
}
