package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/gateway/app/options"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

const (
	// UserAgent is the User-Agent string for the gateway service.
	UserAgent = "aetherius-gateway"
)

// Execute runs the gateway command.
func Execute() {
	opts := options.NewServerOptions()

	app := &GatewayApp{}

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
	opts   *options.ServerOptions
	logger core.Logger

	// Component initializers (Wire will inject these)
	redisInit  *RedisInitializer
	httpInit   *HTTPServerInitializer
	healthInit *pkginitializers.HealthCheckInitializer
}

// Name returns the application name.
func (a *GatewayApp) Name() string {
	return "Gateway Service"
}

// Initialize initializes the application.
func (a *GatewayApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	a.opts = opts.(*options.ServerOptions)

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
	<-ctx.Done()
	return nil
}

// Shutdown gracefully shuts down the application.
func (a *GatewayApp) Shutdown(ctx context.Context) error {
	return nil
}

// registerComponents registers all component initializers with bootstrap.
func (a *GatewayApp) registerComponents(bs *bootstrap.Bootstrap) error {
	// Use Wire to inject all dependencies
	container, err := InitializeGatewayContainer(a.opts, a.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize container: %w", err)
	}

	// Register to Bootstrap
	bs.Register(container.Redis)
	bs.Register(container.HTTP)
	bs.Register(container.Health)

	// Save references
	a.redisInit = container.Redis
	a.httpInit = container.HTTP
	a.healthInit = container.Health

	return nil
}
