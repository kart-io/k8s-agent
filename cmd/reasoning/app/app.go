// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/reasoning/app/options"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	reasoningconfig "github.com/kart-io/k8s-agent/internal/reasoning/config"
	"github.com/kart-io/k8s-agent/internal/reasoning/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// Execute runs the reasoning service command
func Execute() {
	// Create configuration options
	opts := options.NewServerOptions()

	// Create application instance
	app := &ReasoningApp{}

	// Use the simplified RunWithBootstrap to run the application
	commonapp.RunWithBootstrap(
		app,
		opts,
		commonapp.Config{
			Use:       "reasoning",
			Short:     "Reasoning Service",
			Long:      "Reasoning Service provides AI-driven root cause analysis and intelligent recommendations",
			EnvPrefix: "REASONING",
		},
		app.registerComponents,
	)
}

// ReasoningApp implements commonapp.Application interface
type ReasoningApp struct {
	config *reasoningconfig.Config
	logger core.Logger

	// Component initializers
	llmInit           *initializers.LLMInitializer
	unifiedServerInit *initializers.UnifiedServerInitializer
	healthInit        *pkginitializers.HealthCheckInitializer
}

// Name returns the application name
func (a *ReasoningApp) Name() string {
	return "Reasoning Service"
}

// Initialize initializes the application
func (a *ReasoningApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	// Convert configuration
	serverOpts := opts.(*options.ServerOptions)
	config := serverOpts.Config()
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
func (a *ReasoningApp) Run(ctx context.Context) error {
	// The bootstrap framework handles running all servers
	// This method can be used for additional application logic if needed
	<-ctx.Done()
	return nil
}

// Shutdown gracefully shuts down the application
func (a *ReasoningApp) Shutdown(ctx context.Context) error {
	// Bootstrap framework handles component shutdown
	// This method can be used for additional cleanup if needed
	return nil
}

// registerComponents registers all component initializers with bootstrap
func (a *ReasoningApp) registerComponents(bs *bootstrap.Bootstrap) error {
	// Get server options from bootstrap context
	// For now, we'll need to recreate the options since bootstrap doesn't provide them directly
	opts := options.NewServerOptions()
	opts.Complete()

	// 1. LLM Clients (priority 400)
	a.llmInit = initializers.NewLLMInitializer(opts.LLM, a.logger)
	bs.Register(a.llmInit)

	// 2. Unified Server (gRPC + HTTP using Kratos framework, OneX architecture pattern)
	// Priority 450 - after LLM initialization
	// A single handler implements both ReasoningServiceServer (gRPC) and ReasoningServiceHTTPServer (HTTP)
	// This follows the OneX pattern where both protocols share the same handler methods
	a.unifiedServerInit = initializers.NewUnifiedServerInitializer(
		opts,
		a.logger,
		a.llmInit,
	)
	bs.Register(a.unifiedServerInit)

	// 3. Health Check (priority 2000)
	healthOpts := commonoptions.NewHealthOptions()
	a.healthInit = pkginitializers.NewHealthCheckInitializer(
		healthOpts,
		a.logger,
	)
	bs.Register(a.healthInit)

	return nil
}
