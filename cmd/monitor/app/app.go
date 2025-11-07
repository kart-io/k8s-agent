package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/monitor/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

const (
	// UserAgent is the User-Agent string for the monitor service.
	UserAgent = "aetherius-monitor"
)

// Execute runs the monitor command.
func Execute() {
	// Create configuration options using StandardOptions
	opts := commonapp.NewStandardOptions("Monitor", UserAgent).
		WithDatabase().
		WithRedis().
		WithPrometheus().
		WithAlert()

	// Create application instance
	app := &MonitorApp{}

	// Use the simplified RunWithBootstrap to run the application
	commonapp.RunWithBootstrap(
		app,
		opts,
		commonapp.Config{
			Use:       "monitor",
			Short:     "Monitor Service",
			Long:      "Monitor Service provides monitoring and metrics collection for the platform",
			EnvPrefix: "MONITOR",
		},
		app.registerComponents,
	)
}

// MonitorApp implements commonapp.Application interface.
type MonitorApp struct {
	opts   *commonapp.StandardOptions // Use StandardOptions
	logger core.Logger

	// Component initializers
	dbInit     *initializers.DatabaseInitializer
	redisInit  *initializers.RedisInitializer
	httpInit   *initializers.HTTPServerInitializer
	healthInit *pkginitializers.HealthCheckInitializer
}

// Name returns the application name.
func (a *MonitorApp) Name() string {
	return "Monitor Service"
}

// Initialize initializes the application.
func (a *MonitorApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	// 保存 ServerOptions
	a.opts = opts.(*commonapp.StandardOptions)

	// Initialize logger
	logger, err := a.opts.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Info("Starting Monitor Service...")

	return nil
}

// Run runs the application.
func (a *MonitorApp) Run(ctx context.Context) error {
	// The bootstrap framework handles running all servers
	// This method can be used for additional application logic if needed
	<-ctx.Done()
	return nil
}

// Shutdown gracefully shuts down the application.
func (a *MonitorApp) Shutdown(ctx context.Context) error {
	// Bootstrap framework handles component shutdown
	// This method can be used for additional cleanup if needed
	return nil
}

// registerComponents registers all component initializers with bootstrap.
func (a *MonitorApp) registerComponents(bs *bootstrap.Bootstrap) error {
	// Use Wire to automatically inject all dependencies
	components, err := InitializeMonitorComponents(a.opts)
	if err != nil {
		return fmt.Errorf("failed to initialize components: %w", err)
	}

	// Register components to Bootstrap
	bs.Register(components.DB)
	bs.Register(components.Redis)
	bs.Register(components.HTTP)
	bs.Register(components.GRPC)
	bs.Register(components.Health)

	// Save references for app
	a.dbInit = components.DB
	a.redisInit = components.Redis
	a.httpInit = components.HTTP
	a.healthInit = components.Health

	return nil
}
