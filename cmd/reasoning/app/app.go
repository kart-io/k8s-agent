// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"
	"sort"

	"github.com/kart-io/k8s-agent/internal/reasoning/analyzer"
	"github.com/kart-io/k8s-agent/internal/reasoning/config"
	"github.com/kart-io/k8s-agent/internal/reasoning/handler"
	"github.com/kart-io/k8s-agent/internal/reasoning/llm"
	"github.com/kart-io/k8s-agent/internal/reasoning/server"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
)

const (
	// UserAgent is the User-Agent string for the reasoning service.
	UserAgent = "aetherius-reasoning"
)

// Execute runs the reasoning service command.
func Execute() {
	opts := commonapp.NewStandardOptions("Reasoning", UserAgent).
		WithLLM()

	app := &ReasoningApp{}

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
	opts   *commonapp.StandardOptions
	logger core.Logger

	// Server component
	server *server.Server
}

// Name returns the application name.
func (a *ReasoningApp) Name() string {
	return "Reasoning Service"
}

// Initialize initializes the application.
func (a *ReasoningApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	a.opts = opts.(*commonapp.StandardOptions)

	// Initialize logger
	logger, err := a.opts.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Infow("Starting Reasoning Service",
		"llm_enabled", a.opts.LLM != nil && a.opts.LLM.Enabled,
	)

	return nil
}

// Run runs the application.
func (a *ReasoningApp) Run(ctx context.Context) error {
	// Bootstrap framework handles running the server
	<-ctx.Done()
	return nil
}

// Shutdown gracefully shuts down the application.
func (a *ReasoningApp) Shutdown(ctx context.Context) error {
	// Bootstrap framework handles component shutdown
	return nil
}

// registerComponents registers all component initializers with bootstrap.
// This uses direct initialization without Wire for simplicity.
func (a *ReasoningApp) registerComponents(bs *bootstrap.Bootstrap) error {
	// Initialize LLM clients
	llmClients := a.initializeLLMClients()

	// Create config from options
	cfg := config.NewConfigFromOptions(a.opts)

	// Initialize business logic components
	rootCauseAnalyzer := analyzer.NewRootCauseAnalyzer(cfg, llmClients)
	reasoningHandler := handler.NewReasoningHandler(rootCauseAnalyzer, a.logger)

	// Initialize unified server (HTTP + gRPC)
	srv, err := server.NewServer(&server.Config{
		HTTP:    a.opts.Server,
		GRPC:    a.opts.GRPC,
		Handler: reasoningHandler,
	}, a.logger)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	a.server = srv

	// Register server as an initializer
	bs.Register(&ServerInitializer{
		server: srv,
		logger: a.logger,
	})

	a.logger.Infow("Components registered",
		"http_port", a.opts.Server.Port,
		"grpc_port", a.opts.GRPC.Port,
		"llm_providers", len(llmClients),
	)

	return nil
}

// initializeLLMClients initializes LLM clients from options.
func (a *ReasoningApp) initializeLLMClients() []llm.Client {
	if a.opts.LLM == nil || !a.opts.LLM.Enabled {
		a.logger.Info("LLM support disabled")
		return nil
	}

	a.logger.Info("Initializing LLM providers")

	// Sort providers by priority
	sort.Slice(a.opts.LLM.Providers, func(i, j int) bool {
		return a.opts.LLM.Providers[i].Priority < a.opts.LLM.Providers[j].Priority
	})

	clients := make([]llm.Client, 0, len(a.opts.LLM.Providers))
	for _, providerCfg := range a.opts.LLM.Providers {
		if providerCfg.APIKey == "" {
			a.logger.Warnw("Skipping LLM provider - no API key", "provider", providerCfg.Name)
			continue
		}

		llmConfig := &llm.Config{
			Provider:    llm.Provider(providerCfg.Name),
			APIKey:      providerCfg.APIKey,
			BaseURL:     providerCfg.BaseURL,
			Model:       providerCfg.Model,
			MaxTokens:   providerCfg.MaxTokens,
			Temperature: providerCfg.Temperature,
			Timeout:     providerCfg.Timeout,
		}

		client, err := llm.NewClient(llmConfig)
		if err != nil {
			a.logger.Errorw("Failed to initialize LLM provider",
				"provider", providerCfg.Name,
				"error", err,
			)
			continue
		}

		clients = append(clients, client)
		a.logger.Infow("LLM provider initialized",
			"provider", providerCfg.Name,
			"model", providerCfg.Model,
			"priority", providerCfg.Priority,
		)
	}

	if len(clients) == 0 {
		a.logger.Warn("No LLM providers available")
	}

	return clients
}

// ServerInitializer wraps the server for bootstrap compatibility.
type ServerInitializer struct {
	server *server.Server
	logger core.Logger
}

// Name returns the initializer name.
func (s *ServerInitializer) Name() string {
	return "UnifiedServer"
}

// Priority returns the initialization priority.
func (s *ServerInitializer) Priority() int {
	return 1000
}

// Initialize starts the server.
func (s *ServerInitializer) Initialize(ctx context.Context) error {
	s.logger.Infow("Starting unified server")
	go func() {
		if err := s.server.Start(ctx); err != nil {
			s.logger.Errorw("Server error", "error", err)
		}
	}()
	return nil
}

// Shutdown stops the server.
func (s *ServerInitializer) Shutdown(ctx context.Context) error {
	s.logger.Infow("Shutting down unified server")
	return s.server.Stop()
}
