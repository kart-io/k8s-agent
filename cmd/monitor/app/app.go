package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/common/loggerutil"
	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/internal/monitor/config"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/logger/core"
)

// Execute runs the monitor command
func Execute() {
	// Create configuration options
	opts := config.NewOptions()

	// Create application instance
	app := &MonitorApp{}

	// Use the simplified framework (no bootstrap needed for simple services)
	commonapp.Run(
		app,
		opts,
		commonapp.Config{
			Use:       "monitor",
			Short:     "Monitor Service",
			Long:      "Monitor Service provides monitoring and metrics collection for the platform",
			EnvPrefix: "MONITOR",
		},
	)
}

// MonitorApp implements commonapp.Application interface
type MonitorApp struct {
	config *config.Options
	logger core.Logger
	server *MonitorService
}

// Name returns the application name
func (a *MonitorApp) Name() string {
	return "Monitor Service"
}

// Initialize initializes the application
func (a *MonitorApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	// Convert configuration
	configOpts := opts.(*config.Options)
	a.config = configOpts

	// Convert LoggingConfig to LoggingOptions for compatibility
	logOpts := &options.LoggingOptions{
		Engine:      "slog",
		Level:       configOpts.Logging.Level,
		Format:      configOpts.Logging.Format,
		OutputPaths: []string{configOpts.Logging.Output},
	}

	// Initialize logger
	logger, err := loggerutil.InitFromOptions(logOpts)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Info("Starting Monitor Service...")

	// Create service
	svc, err := NewServer(configOpts, logger)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	a.server = svc

	return nil
}

// Run runs the application
func (a *MonitorApp) Run(ctx context.Context) error {
	// Start service
	return a.server.Run(ctx)
}

// Shutdown gracefully shuts down the application
func (a *MonitorApp) Shutdown(ctx context.Context) error {
	if a.logger != nil {
		a.logger.Flush()
	}
	return nil
}
