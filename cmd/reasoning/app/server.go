package app

import (
	"context"
	"fmt"
	"sort"

	"github.com/kart-io/k8s-agent/internal/reasoning/api"
	"github.com/kart-io/k8s-agent/internal/reasoning/config"
	"github.com/kart-io/k8s-agent/internal/reasoning/llm"
	"github.com/kart-io/logger/core"
)

// Server represents the reasoning server
type Server struct {
	opts       *config.Options
	log        core.Logger
	llmClients []llm.Client
	apiServer  *api.Server
}

// NewServer creates a new reasoning server
func NewServer(opts *config.Options, log core.Logger) (*Server, error) {
	srv := &Server{
		opts: opts,
		log:  log,
	}

	if err := srv.initialize(); err != nil {
		return nil, err
	}

	return srv, nil
}

// initialize initializes all server components
func (s *Server) initialize() error {
	// Initialize LLM clients
	if s.opts.LLM.Enabled {
		s.log.Info("Initializing LLM providers")

		// Sort providers by priority
		sort.Slice(s.opts.LLM.Providers, func(i, j int) bool {
			return s.opts.LLM.Providers[i].Priority < s.opts.LLM.Providers[j].Priority
		})

		for _, providerCfg := range s.opts.LLM.Providers {
			if providerCfg.APIKey == "" {
				s.log.Warnw("Skipping LLM provider - no API key",
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
				s.log.Errorw("Failed to initialize LLM provider",
					"provider", providerCfg.Name,
					"error", err,
				)
				continue
			}

			s.llmClients = append(s.llmClients, client)
			s.log.Infow("LLM provider initialized",
				"provider", providerCfg.Name,
				"model", providerCfg.Model,
				"priority", providerCfg.Priority,
			)
		}

		if len(s.llmClients) == 0 {
			s.log.Warn("No LLM providers available")
		}
	} else {
		s.log.Info("LLM support disabled")
	}

	// Log Memory configuration
	if s.opts.Memory.EnableVectorStore {
		s.log.Infow("Memory vector store enabled",
			"type", s.opts.Memory.VectorStoreType,
			"path", s.opts.Memory.VectorStorePath,
			"embedding_model", s.opts.Memory.EmbeddingModel,
			"embedding_provider", s.opts.Memory.EmbeddingProvider,
		)
	} else {
		s.log.Info("Memory vector store disabled")
	}

	// Create API server with Config
	cfg := s.opts.Config()
	s.apiServer = api.NewServer(cfg, s.llmClients)

	return nil
}

// Run starts the reasoning server
func (s *Server) Run(ctx context.Context) error {
	s.log.Info("Starting Reasoning Service API server...")

	if err := s.apiServer.Start(); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	s.log.Info("Reasoning Service shutdown complete")
	return nil
}
