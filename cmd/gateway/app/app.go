package app

import (
	"context"
	"fmt"
	"os"

	"github.com/kart-io/k8s-agent/cmd/gateway/app/options"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/logger/core"
)

// Execute runs the gateway command.
func Execute() {
	// Create configuration options
	opts := options.NewOptions()

	// Create application instance
	app := &GatewayApp{}

	// Use the simplified framework (no bootstrap needed for simple services)
	commonapp.Run(
		app,
		opts,
		commonapp.Config{
			Use:       "gateway",
			Short:     "Gateway Service",
			Long:      "Gateway Service provides API gateway with Traefik integration",
			EnvPrefix: "GATEWAY",
		},
	)
}

// GatewayApp implements commonapp.Application interface.
type GatewayApp struct {
	opts   *options.Options // 直接使用Options
	logger core.Logger
	server *GatewayService
}

// Name returns the application name.
func (a *GatewayApp) Name() string {
	return "Gateway Service"
}

// Initialize initializes the application.
func (a *GatewayApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	// 直接保存Options，不需要转换
	a.opts = opts.(*options.Options)

	// Initialize logger
	logger, err := a.opts.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Info("Starting Gateway Service...")

	// Create service
	svc, err := NewServer(a.opts, logger)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	a.server = svc

	return nil
}

// Run runs the application.
func (a *GatewayApp) Run(ctx context.Context) error {
	// Start service
	return a.server.Run(ctx)
}

// Shutdown gracefully shuts down the application.
func (a *GatewayApp) Shutdown(ctx context.Context) error {
	if a.server != nil {
		if err := a.server.Cleanup(); err != nil {
			a.logger.Errorw("Failed to cleanup server", "error", err)
		}
	}
	if a.logger != nil {
		if err := a.logger.Flush(); err != nil {
			// Can't log the error since logger is being flushed
			fmt.Fprintf(os.Stderr, "Failed to flush logger: %v\n", err)
		}
	}
	return nil
}
