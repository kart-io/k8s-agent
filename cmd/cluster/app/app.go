// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/cluster/app/options"
	"github.com/kart-io/k8s-agent/internal/cluster/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// Execute runs the cluster service command.
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

// ClusterApp implements commonapp.Application interface.
type ClusterApp struct {
	opts   *options.ServerOptions // 直接使用ServerOptions
	logger core.Logger

	// Component initializers
	dbInit     *initializers.DatabaseInitializer
	httpInit   *initializers.HTTPServerInitializer
	healthInit *pkginitializers.HealthCheckInitializer
}

// Name returns the application name.
func (a *ClusterApp) Name() string {
	return "Cluster Service"
}

// Initialize initializes the application.
func (a *ClusterApp) Initialize(ctx context.Context, opts commonapp.Options) error {
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
func (a *ClusterApp) Run(ctx context.Context) error {
	// The bootstrap framework handles running all servers
	// This method can be used for additional application logic if needed
	a.logger.Infow("Cluster service running, waiting for shutdown signal")

	// Wait for shutdown signal
	<-ctx.Done()
	return nil
}

// Shutdown gracefully shuts down the application.
func (a *ClusterApp) Shutdown(ctx context.Context) error {
	// Bootstrap framework handles component shutdown
	// This method can be used for additional cleanup if needed
	return nil
}

// registerComponents registers all component initializers with bootstrap.
func (a *ClusterApp) registerComponents(bs *bootstrap.Bootstrap) error {
	// Use Wire to automatically inject all dependencies
	components, err := InitializeClusterComponents(a.opts)
	if err != nil {
		return fmt.Errorf("failed to initialize components: %w", err)
	}

	// Register components to Bootstrap
	bs.Register(components.DB)
	bs.Register(components.HTTP)
	bs.Register(components.Health)

	// Save references for app
	a.dbInit = components.DB
	a.httpInit = components.HTTP
	a.healthInit = components.Health

	return nil
}
