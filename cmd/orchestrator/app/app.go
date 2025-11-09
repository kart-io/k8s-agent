// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/orchestrator/startup"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

const (
	// UserAgent is the User-Agent string for the orchestrator service.
	UserAgent = "aetherius-orchestrator"
)

// Execute runs the orchestrator command.
func Execute() {
	// Create configuration options
	opts := commonapp.NewStandardOptions("Orchestrator", UserAgent).
		WithDatabase().WithRedis().WithNATS().WithMetrics().
		WithAI().WithWorkflow()

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
	bootstrap *bootstrap.Bootstrap
	opts      *commonapp.StandardOptions
	logger    core.Logger
}

// Name returns the application name.
func (a *OrchestratorApp) Name() string {
	return "Orchestrator Service"
}

// Initialize initializes the application.
func (a *OrchestratorApp) Initialize(ctx context.Context, opts commonapp.Options) error {
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
func (a *OrchestratorApp) Run(ctx context.Context) error {
	// Bootstrap.Run() is already called by RunWithBootstrap
	// We just need to wait for the context to be cancelled
	<-ctx.Done()
	return ctx.Err()
}

// Shutdown gracefully shuts down the application.
func (a *OrchestratorApp) Shutdown(ctx context.Context) error {
	// Bootstrap shutdown is handled automatically
	return nil
}

// registerComponents registers all component initializers with bootstrap.
// This method defines the complete initialization order for the orchestrator service.
func (a *OrchestratorApp) registerComponents(bs *bootstrap.Bootstrap) error {
	a.bootstrap = bs

	// Layer 1: Infrastructure (Priority 300-500)
	// Database, Redis, and NATS configuration
	infra := startup.NewInfrastructureInitializers(a.opts, a.logger)
	bs.Register(infra.Database)
	bs.Register(infra.Redis)
	bs.Register(infra.NATS)

	// Layer 2: Core Business Services (Priority 600)
	// Workflow engine, strategy manager, and workflow service
	coreServices := startup.NewCoreServicesInitializer(a.opts, a.logger, infra)
	bs.Register(coreServices)

	// Layer 3: Event Processing (Priority 700)
	// Event subscriber for NATS events
	eventSubscriber := startup.NewEventSubscriberInitializer(a.logger, infra, coreServices)
	bs.Register(eventSubscriber)

	// Layer 4: Server Layer (Priority 900-1000)
	// gRPC and HTTP servers that expose the services
	if a.opts.GRPC.Enable {
		grpcInit := startup.NewGRPCServerInitializer(a.opts, a.logger, coreServices)
		bs.Register(grpcInit)
	}

	httpInit := startup.NewHTTPServerInitializer(a.opts, a.logger, coreServices)
	bs.Register(httpInit)

	// Layer 5: Monitoring (Priority 2000)
	// Health check endpoint
	healthInit := pkginitializers.NewHealthCheckInitializer(a.opts.Health, a.logger)
	bs.Register(healthInit)

	return nil
}
