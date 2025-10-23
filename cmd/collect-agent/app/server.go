package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/internal/collect-agent/agent"
	"github.com/kart-io/k8s-agent/internal/collect-agent/config"
	"github.com/kart-io/logger/core"
)

// Server represents the collect-agent server
type Server struct {
	opts          *config.Options
	log           core.Logger
	agentInstance *agent.Agent
	healthServer  *agent.HealthServer
}

// NewServer creates a new collect-agent server
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
	var err error

	// Convert Options to AgentConfig for backward compatibility with agent package
	// TODO: Update agent package to use config.Options directly
	agentConfig := s.opts.ToAgentConfig()

	// Create Zap logger for agent (agent package still uses zap.Logger)
	// TODO: Update agent package to use logger.Logger
	zapLogger, err := createZapLogger(s.log)
	if err != nil {
		return fmt.Errorf("failed to create zap logger: %w", err)
	}

	// Create agent instance
	s.agentInstance, err = agent.New(agentConfig, zapLogger)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	// Determine health port
	port := s.opts.Agent.HealthPort
	if port == 0 {
		port = 8080 // default
	}

	// Create health server
	s.healthServer = agent.NewHealthServer(s.agentInstance, port, zapLogger)

	return nil
}

// createZapLogger creates a zap.Logger from a core.Logger
// This is a temporary bridge until agent package is updated to use logger.Logger
func createZapLogger(log core.Logger) (*zap.Logger, error) {
	// For now, create a new zap logger with similar config
	// In a real implementation, we would want to extract the underlying zap logger
	// or update the agent package to use logger.Logger
	zapCfg := zap.NewProductionConfig()
	return zapCfg.Build()
}

// Run starts the collect-agent server
func (s *Server) Run(ctx context.Context) error {
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		s.log.Infow("Received shutdown signal",
			"signal", sig.String(),
		)
		cancel()
	}()

	// Start health server
	if err := s.healthServer.Start(); err != nil {
		return fmt.Errorf("failed to start health server: %w", err)
	}

	// Start agent
	s.log.Info("Starting agent services...")
	if err := s.agentInstance.Start(ctx); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	s.log.Info("Agent shutdown complete")
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	if s.healthServer != nil {
		s.healthServer.Stop()
	}

	s.log.Info("Collect Agent shutdown complete")
	return nil
}
