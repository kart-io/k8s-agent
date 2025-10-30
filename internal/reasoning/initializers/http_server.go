// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package initializers

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/reasoning/api"
	"github.com/kart-io/k8s-agent/internal/reasoning/config"
	"github.com/kart-io/logger/core"
)

// HTTPServerInitializer initializes the HTTP API server
type HTTPServerInitializer struct {
	config    *config.Config
	logger    core.Logger
	llmInit   *LLMInitializer
	apiServer *api.Server
}

// NewHTTPServerInitializer creates a new HTTP server initializer
func NewHTTPServerInitializer(
	config *config.Config,
	logger core.Logger,
	llmInit *LLMInitializer,
) *HTTPServerInitializer {
	return &HTTPServerInitializer{
		config:  config,
		logger:  logger,
		llmInit: llmInit,
	}
}

// Initialize initializes the HTTP server
func (i *HTTPServerInitializer) Initialize(ctx context.Context) error {
	i.logger.Infow("Initializing HTTP API server",
		"host", i.config.Server.Host,
		"port", i.config.Server.Port,
	)

	// Log Memory configuration
	if i.config.Memory.EnableVectorStore {
		i.logger.Infow("Memory vector store enabled",
			"type", i.config.Memory.VectorStoreType,
			"path", i.config.Memory.VectorStorePath,
			"embedding_model", i.config.Memory.EmbeddingModel,
			"embedding_provider", i.config.Memory.EmbeddingProvider,
		)
	} else {
		i.logger.Info("Memory vector store disabled")
	}

	// Get LLM clients from LLM initializer
	llmClients := i.llmInit.GetClients()

	// Create API server
	i.apiServer = api.NewServer(i.config, llmClients)

	i.logger.Infow("HTTP API server initialized",
		"port", i.config.Server.Port,
		"llm_providers", len(llmClients),
	)

	return nil
}

// Shutdown stops the HTTP server
func (i *HTTPServerInitializer) Shutdown(ctx context.Context) error {
	i.logger.Info("Shutting down HTTP API server")
	// Note: The reasoning API server doesn't have explicit shutdown method
	// Server will be stopped when the process exits
	return nil
}

// Priority returns the initialization priority (higher = earlier)
func (i *HTTPServerInitializer) Priority() int {
	return 500 // HTTP server should be initialized after LLM
}

// Name returns the name of this initializer
func (i *HTTPServerInitializer) Name() string {
	return "HTTPServer"
}

// GetServer returns the initialized server instance
func (i *HTTPServerInitializer) GetServer() *api.Server {
	return i.apiServer
}

// Start starts the HTTP server
func (i *HTTPServerInitializer) Start() error {
	if i.apiServer == nil {
		return fmt.Errorf("server not initialized")
	}
	i.logger.Info("Starting Reasoning Service API server...")
	return i.apiServer.Start()
}
