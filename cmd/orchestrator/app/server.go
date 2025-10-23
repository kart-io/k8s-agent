package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/kart-io/k8s-agent/internal/orchestrator/config"
	"github.com/kart-io/k8s-agent/internal/orchestrator/storage"
	"github.com/kart-io/k8s-agent/internal/orchestrator/strategy"
	"github.com/kart-io/k8s-agent/internal/orchestrator/subscriber"
	"github.com/kart-io/k8s-agent/internal/orchestrator/workflow"
)

// Server represents the orchestrator server
type Server struct {
	cfg             *config.Config
	log             *zap.Logger
	pgStore         *storage.PostgresStore
	redisStore      *storage.RedisStore
	natsConn        *nats.Conn
	engine          *workflow.Engine
	strategyManager *strategy.Manager
	subscriber      *subscriber.Subscriber
}

// NewServer creates a new orchestrator server
func NewServer(cfg *config.Config) (*Server, error) {
	logger, err := initLogger(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	logger.Info("Starting Aetherius Orchestrator Service")

	srv := &Server{
		cfg: cfg,
		log: logger,
	}

	if err := srv.initialize(); err != nil {
		return nil, err
	}

	return srv, nil
}

// initialize initializes all server components
func (s *Server) initialize() error {
	var err error

	s.log.Info("==========================================================")
	s.log.Info("     Aetherius Orchestrator Service - Initialization      ")
	s.log.Info("==========================================================")

	// Initialize PostgreSQL
	s.log.Info("📦 [1/6] Initializing PostgreSQL",
		zap.String("host", s.cfg.Database.Host),
		zap.Int("port", s.cfg.Database.Port),
		zap.String("database", s.cfg.Database.Database))
	s.pgStore, err = storage.NewPostgresStore(s.cfg.Database, s.log)
	if err != nil {
		return fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}
	s.log.Info("✅ PostgreSQL initialized successfully")

	// Initialize Redis
	s.log.Info("📦 [2/6] Initializing Redis",
		zap.String("addr", s.cfg.Redis.Addr))
	s.redisStore, err = storage.NewRedisStore(s.cfg.Redis, s.log)
	if err != nil {
		s.pgStore.Close()
		return fmt.Errorf("failed to initialize Redis: %w", err)
	}
	s.log.Info("✅ Redis initialized successfully")

	// Connect to NATS
	s.log.Info("📡 [3/6] Connecting to NATS",
		zap.String("url", s.cfg.NATS.URL))
	s.natsConn, err = nats.Connect(
		s.cfg.NATS.URL,
		nats.Name("orchestrator-service"),
		nats.MaxReconnects(s.cfg.NATS.MaxReconnect),
		nats.ReconnectWait(s.cfg.NATS.ReconnectWait),
	)
	if err != nil {
		s.redisStore.Close()
		s.pgStore.Close()
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}
	s.log.Info("✅ NATS connected successfully",
		zap.String("server_url", s.natsConn.ConnectedUrl()))

	// Initialize workflow components
	s.log.Info("⚙️  [4/6] Initializing workflow engine",
		zap.String("agent_manager_url", s.cfg.AI.AgentManagerURL),
		zap.String("reasoning_service_url", s.cfg.AI.ReasoningServiceURL))

	executor := workflow.NewExecutor(
		s.cfg.AI.AgentManagerURL,
		s.cfg.AI.ReasoningServiceURL,
		s.log,
	)

	s.engine = workflow.NewEngine(s.pgStore, s.redisStore, executor, s.log)
	s.log.Info("✅ Workflow engine initialized")

	// Initialize strategy manager
	s.log.Info("🎯 [5/6] Initializing strategy manager")
	s.strategyManager = strategy.NewManager(s.pgStore, s.engine, s.log)
	s.log.Info("✅ Strategy manager initialized")

	return nil
}

// Run starts the orchestrator server
func (s *Server) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize event subscriber
	s.log.Info("📬 [6/6] Initializing event subscriber")
	s.subscriber = subscriber.NewSubscriber(s.natsConn, s.strategyManager, s.log)
	if err := s.subscriber.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event subscriber: %w", err)
	}

	s.log.Info("==========================================================")
	s.log.Info("✅ Orchestrator Service started successfully!")
	s.log.Info("==========================================================")
	s.log.Info("🎧 Listening for events on NATS channels:")
	s.log.Info("   - internal.event.critical")
	s.log.Info("   - internal.event.anomaly")
	s.log.Info("   - internal.event.* (debug)")
	s.log.Info("==========================================================")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	s.log.Info("🛑 Shutdown signal received")
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

	if s.log != nil {
		s.log.Sync()
	}

	s.log.Info("Orchestrator Service shutdown complete")
	return nil
}

func initLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	var encoderConfig zapcore.EncoderConfig
	if cfg.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      cfg.Format != "json",
		Encoding:         cfg.Format,
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{cfg.OutputPath},
		ErrorOutputPaths: []string{"stderr"},
	}

	return zapConfig.Build()
}
