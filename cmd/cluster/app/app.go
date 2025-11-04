// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/cluster/app/options"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	clusterconfig "github.com/kart-io/k8s-agent/internal/cluster/config"
	"github.com/kart-io/k8s-agent/internal/cluster/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// Execute runs the cluster service command
func Execute() {
	// Create configuration options
	opts := options.NewServerOptions()

	// Create application instance
	app := &ClusterApp{}

	// Use the simplified RunWithBootstrap to run the application
	commonapp.RunWithBootstrap(
		app,
		opts,
		commonapp.Config{
			Use:       "cluster",
			Short:     "Cluster Service",
			Long:      "Cluster Service provides multi-cluster management and K8s resource API",
			EnvPrefix: "CLUSTER",
		},
		app.registerComponents,
	)
}

// ClusterApp implements commonapp.Application interface
type ClusterApp struct {
	config *clusterconfig.Config
	logger core.Logger

	// Component initializers
	dbInit     *initializers.DatabaseInitializer
	httpInit   *initializers.HTTPServerInitializer
	healthInit *pkginitializers.HealthCheckInitializer
}

// Name returns the application name
func (a *ClusterApp) Name() string {
	return "Cluster Service"
}

// Initialize initializes the application
func (a *ClusterApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	// Convert configuration
	serverOpts := opts.(*options.ServerOptions)
	config, err := serverOpts.Config()
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}
	a.config = config

	// Initialize logger
	logger, err := serverOpts.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	return nil
}

// Run runs the application
func (a *ClusterApp) Run(ctx context.Context) error {
	// Start HTTP server (in goroutine)
	go func() {
		if err := a.httpInit.Start(); err != nil {
			a.logger.Fatalw("HTTP server failed to start", "error", err)
		}
	}()

	a.logger.Infow("All services started, waiting for shutdown signal")

	// Wait for shutdown signal
	<-ctx.Done()
	return nil
}

// Shutdown gracefully shuts down the application
func (a *ClusterApp) Shutdown(ctx context.Context) error {
	// Bootstrap framework handles component shutdown
	// This method can be used for additional cleanup if needed
	return nil
}

// registerComponents registers all component initializers with bootstrap
func (a *ClusterApp) registerComponents(bs *bootstrap.Bootstrap) error {
	// Get server options from bootstrap context
	// For now, we'll need to recreate the options since bootstrap doesn't provide them directly
	opts := options.NewServerOptions()
	opts.Complete()

	// 1. Database (priority 300)
	a.dbInit = initializers.NewDatabaseInitializer(opts.Database, a.logger)
	bs.Register(a.dbInit)

	// 2. HTTP Server (priority 500)
	a.httpInit = initializers.NewHTTPServerInitializer(
		opts.Server,
		opts.JWT,
		a.logger,
		a.dbInit,
	)
	bs.Register(a.httpInit)

	// 3. Health Check (priority 2000)
	healthOpts := commonoptions.NewHealthOptions()
	a.healthInit = pkginitializers.NewHealthCheckInitializer(
		healthOpts,
		a.logger,
	)
	bs.Register(a.healthInit)

	return nil
}
