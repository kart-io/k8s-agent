package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kart-io/k8s-agent/internal/collect-agent/agent"
	"github.com/kart-io/k8s-agent/internal/collect-agent/config"
	"github.com/kart-io/logger"
)

// Server represents the collect-agent server
type Server struct {
	opts          *config.Options
	log           logger.Logger
	agentInstance *agent.Agent
	healthServer  *agent.HealthServer
}

// NewServer creates a new collect-agent server
func NewServer(opts *config.Options, log logger.Logger) (*Server, error) {
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

	// Create agent instance
	s.agentInstance, err = agent.New(s.opts, s.log)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	// Determine health port
	port := s.opts.HealthPort
	if port == 0 {
		port = 8080 // default
	}

	// Create health server
	s.healthServer = agent.NewHealthServer(s.agentInstance, port, s.log)

	return nil
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
