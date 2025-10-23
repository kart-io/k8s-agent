package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kart-io/logger"
	"github.com/nats-io/nats.go"

	"github.com/kart-io/k8s-agent/internal/orchestrator/config"
	"github.com/kart-io/k8s-agent/internal/orchestrator/storage"
	"github.com/kart-io/k8s-agent/internal/orchestrator/strategy"
	"github.com/kart-io/k8s-agent/internal/orchestrator/subscriber"
	"github.com/kart-io/k8s-agent/internal/orchestrator/workflow"
)

// Server represents the orchestrator server
type Server struct {
	opts            *config.Options
	log             logger.Logger
	pgStore         *storage.PostgresStore
	redisStore      *storage.RedisStore
	natsConn        *nats.Conn
	engine          *workflow.Engine
	strategyManager *strategy.Manager
	subscriber      *subscriber.Subscriber
}

// NewServer creates a new orchestrator server
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

	// Initialize PostgreSQL
	s.log.Infow("Initializing PostgreSQL",
		"host", s.opts.Database.Host,
		"port", s.opts.Database.Port,
		"database", s.opts.Database.Database,
	)
	s.pgStore, err = storage.NewPostgresStore(s.opts.Database, s.log)
	if err != nil {
		return fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}
	s.log.Info("PostgreSQL initialized successfully")

	// Initialize Redis
	s.log.Infow("Initializing Redis",
		"addr", s.opts.Redis.Addr,
	)
	s.redisStore, err = storage.NewRedisStore(s.opts.Redis, s.log)
	if err != nil {
		s.pgStore.Close()
		return fmt.Errorf("failed to initialize Redis: %w", err)
	}
	s.log.Info("Redis initialized successfully")

	// Connect to NATS
	s.log.Infow("Connecting to NATS",
		"url", s.opts.NATS.URL,
	)
	s.natsConn, err = nats.Connect(
		s.opts.NATS.URL,
		nats.Name("orchestrator-service"),
		nats.MaxReconnects(s.opts.NATS.MaxReconnect),
		nats.ReconnectWait(s.opts.NATS.ReconnectWait),
	)
	if err != nil {
		s.redisStore.Close()
		s.pgStore.Close()
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}
	s.log.Infow("NATS connected successfully",
		"server_url", s.natsConn.ConnectedUrl(),
	)

	// Initialize workflow components
	s.log.Infow("Initializing workflow engine",
		"agent_manager_url", s.opts.AI.AgentManagerURL,
		"reasoning_service_url", s.opts.AI.ReasoningServiceURL,
	)

	executor := workflow.NewExecutor(
		s.opts.AI.AgentManagerURL,
		s.opts.AI.ReasoningServiceURL,
		s.log,
	)

	s.engine = workflow.NewEngine(s.pgStore, s.redisStore, executor, s.log)
	s.log.Info("Workflow engine initialized")

	// Initialize strategy manager
	s.log.Info("Initializing strategy manager")
	s.strategyManager = strategy.NewManager(s.pgStore, s.engine, s.log)
	s.log.Info("Strategy manager initialized")

	return nil
}

// Run starts the orchestrator server
func (s *Server) Run(ctx context.Context) error {
	// Initialize event subscriber
	s.log.Info("Initializing event subscriber")
	s.subscriber = subscriber.NewSubscriber(s.natsConn, s.strategyManager, s.log)
	if err := s.subscriber.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event subscriber: %w", err)
	}

	s.log.Info("==========================================================")
	s.log.Info("Orchestrator Service started successfully!")
	s.log.Info("==========================================================")
	s.log.Info("Listening for events on NATS channels:")
	s.log.Info("  - internal.event.critical")
	s.log.Info("  - internal.event.anomaly")
	s.log.Info("  - internal.event.* (debug)")
	s.log.Info("==========================================================")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	s.log.Info("Shutdown signal received")
	s.log.Info("Shutting down gracefully...")

	return s.Shutdown()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	if s.subscriber != nil {
		s.subscriber.Stop()
	}

	if s.natsConn != nil {
		s.natsConn.Close()
	}

	if s.redisStore != nil {
		s.redisStore.Close()
	}

	if s.pgStore != nil {
		s.pgStore.Close()
	}

	s.log.Info("Orchestrator Service shutdown complete")
	return nil
}
