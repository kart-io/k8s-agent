// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/agent-manager/startup"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

const (
	// UserAgent is the userAgent name when starting agent-manager server.
	UserAgent = "aetherius-agent-manager"
)

// Execute runs the agent-manager command.
func Execute() {
	// Create configuration options using StandardOptions
	opts := commonapp.NewStandardOptions("Agent Manager", UserAgent).
		WithDatabase().
		WithRedis().
		WithNATS().
		WithMetrics()

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
	bootstrap *bootstrap.Bootstrap
	opts      *commonapp.StandardOptions
	logger    core.Logger
}

// Name returns the application name.
func (a *AgentManagerApp) Name() string {
	return "Agent Manager"
}

// Initialize initializes the application.
func (a *AgentManagerApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	a.opts = opts.(*commonapp.StandardOptions)

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
	// Bootstrap.Run() is already called by RunWithBootstrap
	// We just need to wait for the context to be cancelled
	<-ctx.Done()
	return ctx.Err()
}

// Shutdown gracefully shuts down the application.
func (a *AgentManagerApp) Shutdown(ctx context.Context) error {
	// Bootstrap shutdown is handled automatically
	return nil
}

// registerComponents registers all component initializers with bootstrap.
// This method defines the complete initialization order for the agent-manager service.
func (a *AgentManagerApp) registerComponents(bs *bootstrap.Bootstrap) error {
	a.bootstrap = bs

	// Layer 1: Infrastructure (Priority 300-400)
	// Database and Redis configuration
	infra := startup.NewInfrastructureInitializers(a.opts, a.logger)
	bs.Register(infra.Database)
	bs.Register(infra.Redis)

	// Layer 2: Core Business Services (Priority 600)
	// Registry, EventProcessor, and Dispatcher services
	businessServices := startup.NewBusinessServicesInitializer(a.opts, a.logger, infra)
	bs.Register(businessServices)

	// Layer 3: NATS and Command Dispatcher (Priority 500-550)
	// NATS server and dispatcher wiring
	natsInit := startup.NewNATSInitializer(a.opts, a.logger, businessServices)
	bs.Register(natsInit)

	dispatcherInit := startup.NewDispatcherInitializer(a.logger, businessServices, natsInit)
	bs.Register(dispatcherInit)

	// Layer 4: Server Layer (Priority 900-1000)
	// gRPC and HTTP servers that expose the services
	if a.opts.GRPC.Enable {
		grpcInit := startup.NewGRPCServerInitializer(a.opts, a.logger, infra, businessServices)
		bs.Register(grpcInit)
	}

	httpInit := startup.NewHTTPServerInitializer(a.opts, a.logger, infra, businessServices)
	bs.Register(httpInit)

	// Layer 5: Monitoring (Priority 2000)
	// Health check endpoint
	healthInit := pkginitializers.NewHealthCheckInitializer(a.opts.Health, a.logger)
	bs.Register(healthInit)

	return nil
}
