package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"

	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/internal/orchestrator/config"
	"github.com/kart-io/k8s-agent/internal/orchestrator/storage"
	"github.com/kart-io/k8s-agent/internal/orchestrator/strategy"
	"github.com/kart-io/k8s-agent/internal/orchestrator/subscriber"
	"github.com/kart-io/k8s-agent/internal/orchestrator/workflow"
	"github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/logger"
	"github.com/kart-io/logger/core"
	"github.com/kart-io/logger/option"
)

// Server represents the orchestrator server
type Server struct {
	cfg             *config.Config
	log             core.Logger
	pgStore         *storage.PostgresStore
	redisStore      *storage.RedisStore
	natsConn        *nats.Conn
	engine          *workflow.Engine
	strategyManager *strategy.Manager
	subscriber      *subscriber.Subscriber
	healthServer    *app.DefaultHealthCheckServer
}

// NewServer creates a new orchestrator server
func NewServer(cfg *config.Config) (*Server, error) {
	log, err := initLogger(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	log.Info("Starting Aetherius Orchestrator Service")

	srv := &Server{
		cfg: cfg,
		log: log,
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
	s.log.Infow("📦 [1/6] Initializing PostgreSQL",
		"host", s.cfg.Database.Host,
		"port", s.cfg.Database.Port,
		"database", s.cfg.Database.Database)

	// 转换配置为 options.DatabaseOptions
	dbOpts := &options.DatabaseOptions{
		Host:            s.cfg.Database.Host,
		Port:            s.cfg.Database.Port,
		User:            s.cfg.Database.User,
		Password:        s.cfg.Database.Password,
		Database:        s.cfg.Database.Database,
		MaxOpenConns:    s.cfg.Database.MaxOpenConns,
		MaxIdleConns:    s.cfg.Database.MaxIdleConns,
		ConnMaxLifetime: s.cfg.Database.ConnMaxLifetime,
	}

	s.pgStore, err = storage.NewPostgresStore(dbOpts, s.log)
	if err != nil {
		return fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}
	s.log.Info("✅ PostgreSQL initialized successfully")

	// Initialize Redis
	s.log.Infow("📦 [2/6] Initializing Redis",
		"addr", s.cfg.Redis.Addr)

	// 转换配置为 options.RedisOptions
	redisOpts := &options.RedisOptions{
		Addr:         s.cfg.Redis.Addr,
		Password:     s.cfg.Redis.Password,
		DB:           s.cfg.Redis.DB,
		PoolSize:     s.cfg.Redis.PoolSize,
		MinIdleConns: s.cfg.Redis.MinIdleConns,
		DialTimeout:  s.cfg.Redis.DialTimeout,
	}

	s.redisStore, err = storage.NewRedisStore(redisOpts, s.log)
	if err != nil {
		s.pgStore.Close()
		return fmt.Errorf("failed to initialize Redis: %w", err)
	}
	s.log.Info("✅ Redis initialized successfully")

	// Connect to NATS
	s.log.Infow("📡 [3/6] Connecting to NATS",
		"url", s.cfg.NATS.URL)
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
	s.log.Infow("✅ NATS connected successfully",
		"server_url", s.natsConn.ConnectedUrl())

	// Initialize workflow components
	s.log.Infow("⚙️  [4/6] Initializing workflow engine",
		"agent_manager_url", s.cfg.AI.AgentManagerURL,
		"reasoning_service_url", s.cfg.AI.ReasoningServiceURL)

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

	// Start health check server
	healthPort := s.cfg.Server.HealthPort
	if healthPort == 0 {
		healthPort = 8092 // 默认端口
	}
	healthAddr := fmt.Sprintf(":%d", healthPort)
	s.log.Infow("🏥 Starting health check server", "addr", healthAddr)
	s.healthServer = app.NewDefaultHealthCheckServer(healthAddr)
	if err := s.healthServer.Start(); err != nil {
		return fmt.Errorf("failed to start health check server: %w", err)
	}
	s.log.Info("✅ Health check server started (endpoints: /healthz, /readyz)")

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

	if s.healthServer != nil {
		s.log.Info("Stopping health check server")
		s.healthServer.Stop()
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
		s.log.Flush()
	}

	s.log.Info("Orchestrator Service shutdown complete")
	return nil
}

func initLogger(cfg config.LoggingConfig) (core.Logger, error) {
	// 使用 kart-io/logger 的 option.LogOption
	logOpt := &option.LogOption{
		Engine:      "slog",
		Level:       cfg.Level,
		Format:      cfg.Format,
		OutputPaths: []string{cfg.OutputPath},
		InitialFields: map[string]interface{}{
			"service": "orchestrator",
		},
	}

	return logger.New(logOpt)
}
