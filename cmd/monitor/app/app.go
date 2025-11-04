package app

import (
	"context"
	"fmt"
	"os"

	"github.com/kart-io/k8s-agent/cmd/monitor/app/options"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/logger/core"
)

// Execute runs the monitor command.
func Execute() {
	// Create configuration options
	opts := options.NewServerOptions()

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

// MonitorApp implements commonapp.Application interface.
type MonitorApp struct {
	opts   *options.ServerOptions // 使用 ServerOptions
	logger core.Logger
	server *MonitorService
}

// Name returns the application name.
func (a *MonitorApp) Name() string {
	return "Monitor Service"
}

// Initialize initializes the application.
func (a *MonitorApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	// 保存 ServerOptions
	a.opts = opts.(*options.ServerOptions)

	// Initialize logger
	logger, err := a.opts.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Info("Starting Monitor Service...")

	// Create service
	svc, err := NewServer(a.opts, logger)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	a.server = svc

	return nil
}

// Run runs the application.
func (a *MonitorApp) Run(ctx context.Context) error {
	// Start service
	return a.server.Run(ctx)
}

// Shutdown gracefully shuts down the application.
func (a *MonitorApp) Shutdown(ctx context.Context) error {
	if a.logger != nil {
		if err := a.logger.Flush(); err != nil {
			// Can't log the error since logger is being flushed
			fmt.Fprintf(os.Stderr, "Failed to flush logger: %v\n", err)
		}
	}
	return nil
}
