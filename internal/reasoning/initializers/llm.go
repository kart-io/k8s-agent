// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package initializers

import (
	"context"
	"sort"

	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/internal/reasoning/llm"
	"github.com/kart-io/logger/core"
)

// LLMInitializer initializes LLM clients
type LLMInitializer struct {
	opts       *commonoptions.LLMOptions
	logger     core.Logger
	llmClients []llm.Client
}

// NewLLMInitializer creates a new LLM initializer
func NewLLMInitializer(opts *commonoptions.LLMOptions, logger core.Logger) *LLMInitializer {
	return &LLMInitializer{
		opts:   opts,
		logger: logger,
	}
}

// Initialize initializes LLM clients
func (i *LLMInitializer) Initialize(ctx context.Context) error {
	if !i.opts.Enabled {
		i.logger.Info("LLM support disabled")
		return nil
	}

	i.logger.Info("Initializing LLM providers")

	// Sort providers by priority
	sort.Slice(i.opts.Providers, func(j, k int) bool {
		return i.opts.Providers[j].Priority < i.opts.Providers[k].Priority
	})

	for _, providerCfg := range i.opts.Providers {
		if providerCfg.APIKey == "" {
			i.logger.Warnw("Skipping LLM provider - no API key",
				"provider", providerCfg.Name,
			)
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
			i.logger.Errorw("Failed to initialize LLM provider",
				"provider", providerCfg.Name,
				"error", err,
			)
			continue
		}

		i.llmClients = append(i.llmClients, client)
		i.logger.Infow("LLM provider initialized",
			"provider", providerCfg.Name,
			"model", providerCfg.Model,
			"priority", providerCfg.Priority,
		)
	}

	if len(i.llmClients) == 0 {
		i.logger.Warn("No LLM providers available")
	}

	return nil
}

// Shutdown closes all LLM clients
func (i *LLMInitializer) Shutdown(ctx context.Context) error {
	i.logger.Info("Shutting down LLM clients")
	// LLM clients don't need explicit cleanup
	return nil
}

// Priority returns the initialization priority (higher = earlier)
func (i *LLMInitializer) Priority() int {
	return 400 // LLM should be initialized early
}

// Name returns the name of this initializer
func (i *LLMInitializer) Name() string {
	return "LLM"
}

// GetClients returns the initialized LLM clients
func (i *LLMInitializer) GetClients() []llm.Client {
	return i.llmClients
}
