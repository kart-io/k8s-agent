// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	// Removed: options package
	"github.com/kart-io/k8s-agent/internal/reasoning/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

const (
	// UserAgent is the User-Agent string for the reasoning service.
	UserAgent = "aetherius-reasoning"
)

// Execute runs the reasoning service command.
func Execute() {
	// Create configuration options
	opts := commonapp.NewStandardOptions("Reasoning", UserAgent).
		WithLLM()

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

// ReasoningApp implements commonapp.Application interface.
type ReasoningApp struct {
	opts   *commonapp.StandardOptions // 直接使用ServerOptions
	logger core.Logger

	// Component initializers
	llmInit           *initializers.LLMInitializer
	unifiedServerInit *initializers.UnifiedServerInitializer
	healthInit        *pkginitializers.HealthCheckInitializer
}

// Name returns the application name.
func (a *ReasoningApp) Name() string {
	return "Reasoning Service"
}

// Initialize initializes the application.
func (a *ReasoningApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	// 直接保存ServerOptions，不需要转换
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
func (a *ReasoningApp) Run(ctx context.Context) error {
	// The bootstrap framework handles running all servers
	// This method can be used for additional application logic if needed
	<-ctx.Done()
	return nil
}

// Shutdown gracefully shuts down the application.
func (a *ReasoningApp) Shutdown(ctx context.Context) error {
	// Bootstrap framework handles component shutdown
	// This method can be used for additional cleanup if needed
	return nil
}

// registerComponents registers all component initializers with bootstrap.
func (a *ReasoningApp) registerComponents(bs *bootstrap.Bootstrap) error {
	// Use Wire to automatically inject all dependencies
	components, err := InitializeReasoningComponents(a.opts)
	if err != nil {
		return fmt.Errorf("failed to initialize components: %w", err)
	}

	// Register components to Bootstrap
	bs.Register(components.LLM)
	bs.Register(components.UnifiedServer)
	bs.Register(components.Health)

	// Save references for app
	a.llmInit = components.LLM
	a.unifiedServerInit = components.UnifiedServer
	a.healthInit = components.Health

	return nil
}
