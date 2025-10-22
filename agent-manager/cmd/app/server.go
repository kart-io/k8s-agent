package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/logger/core"
	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/agent-manager/internal/agent"
	"github.com/kart-io/k8s-agent/agent-manager/internal/api"
	"github.com/kart-io/k8s-agent/agent-manager/internal/command"
	"github.com/kart-io/k8s-agent/agent-manager/internal/config"
	"github.com/kart-io/k8s-agent/agent-manager/internal/event"
	"github.com/kart-io/k8s-agent/agent-manager/internal/nats"
	"github.com/kart-io/k8s-agent/agent-manager/internal/storage"
	"github.com/kart-io/k8s-agent/agent-manager/pkg/types"
	"github.com/kart-io/k8s-agent/common/db"
)

// Server represents the agent-manager server
type Server struct {
	opts   *config.Options
	logger core.Logger

	// Components
	pgStore        *storage.PostgresStore
	redisStore     *storage.RedisStore
	registry       *agent.Registry
	eventProcessor *event.Processor
	natsServer     *nats.Server
	dispatcher     *command.Dispatcher
	apiServer      *api.Server
}

// NewServer creates a new server instance
func NewServer(opts *config.Options, logger core.Logger) (*Server, error) {
	return &Server{
		opts:   opts,
		logger: logger,
	}, nil
}

// Run starts the server
func (s *Server) Run(ctx context.Context) error {
	// Initialize components
	if err := s.initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	// Start NATS server
	if err := s.natsServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start NATS server: %w", err)
	}
	defer s.natsServer.Stop()

	// Start API server
	errCh := make(chan error, 1)
	go func() {
		if err := s.apiServer.Start(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("failed to start API server: %w", err)
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		s.logger.Info("Shutting down server")
		return s.shutdown()
	case err := <-errCh:
		return err
	}
}

// initialize initializes all components
func (s *Server) initialize(ctx context.Context) error {
	// Initialize database
	if err := s.initDatabase(); err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}

	// Initialize Redis
	if err := s.initRedis(); err != nil {
		return fmt.Errorf("failed to init Redis: %w", err)
	}

	// Initialize registry
	s.registry = agent.NewRegistry(s.pgStore, s.redisStore, s.logger)

	// Initialize event processor
	s.eventProcessor = event.NewProcessor(s.pgStore, s.redisStore, nil, s.logger)

	// Initialize NATS server
	s.natsServer = nats.NewServer(
		s.registry,
		s.eventProcessor,
		s.logger,
		nats.WithURL(s.opts.NATS.URL),
		nats.WithMaxReconnect(s.opts.NATS.MaxReconnect),
		nats.WithReconnectWait(s.opts.NATS.ReconnectWait),
		nats.WithPingInterval(s.opts.NATS.PingInterval),
		nats.WithMaxPingsOut(s.opts.NATS.MaxPingsOut),
	)

	// Initialize command dispatcher
	s.dispatcher = command.NewDispatcher(s.pgStore, s.redisStore, s.registry, s.natsServer, s.logger)

	// Initialize API server
	s.apiServer = api.NewServer(
		s.convertServerConfig(),
		s.registry,
		s.eventProcessor,
		s.dispatcher,
		s.pgStore,
		s.redisStore,
		s.logger,
	)

	return nil
}

// initDatabase initializes the database connection
func (s *Server) initDatabase() error {
	// Create MySQL client using options
	mysqlClient, err := db.NewMySQL(s.logger,
		db.WithHost(s.opts.Database.Host),
		db.WithPort(s.opts.Database.Port),
		db.WithUser(s.opts.Database.User),
		db.WithPassword(s.opts.Database.Password),
		db.WithDatabase(s.opts.Database.Database),
		db.WithMaxOpenConns(s.opts.Database.MaxOpenConns),
		db.WithMaxIdleConns(s.opts.Database.MaxIdleConns),
		db.WithConnMaxLifetime(s.opts.Database.ConnMaxLifetime),
	)
	if err != nil {
		return fmt.Errorf("failed to create MySQL client: %w", err)
	}

	// Create storage with embedded MySQL client
	s.pgStore = &storage.PostgresStore{
		MySQLClient: mysqlClient,
	}

	// Auto-migrate schemas
	if s.opts.Database.AutoMigrate {
		if err := s.pgStore.AutoMigrate(
			&types.Agent{},
			&types.Event{},
			&types.Metrics{},
			&types.Command{},
			&types.CommandResult{},
			&types.Cluster{},
			&types.AlertRule{},
			&types.Alert{},
		); err != nil {
			return fmt.Errorf("failed to auto-migrate: %w", err)
		}
	}

	s.logger.Info("Database initialized",
		zap.String("host", s.opts.Database.Host),
		zap.String("database", s.opts.Database.Database),
	)

	return nil
}

// initRedis initializes the Redis connection
func (s *Server) initRedis() error {
	// Create Redis client using options
	redisClient, err := db.NewRedis(s.logger,
		db.WithAddr(s.opts.Redis.Addr),
		db.WithRedisPassword(s.opts.Redis.Password),
		db.WithRedisDB(s.opts.Redis.DB),
		db.WithRedisPoolSize(s.opts.Redis.PoolSize),
		db.WithRedisMinIdleConns(s.opts.Redis.MinIdleConns),
		db.WithRedisDialTimeout(s.opts.Redis.DialTimeout),
		db.WithRedisReadTimeout(s.opts.Redis.ReadTimeout),
		db.WithRedisWriteTimeout(s.opts.Redis.WriteTimeout),
	)
	if err != nil {
		return fmt.Errorf("failed to create Redis client: %w", err)
	}

	// Create storage with embedded Redis client
	s.redisStore = &storage.RedisStore{
		RedisClient: redisClient,
	}

	s.logger.Info("Redis initialized",
		zap.String("addr", s.opts.Redis.Addr),
	)

	return nil
}

// convertServerConfig converts config.ServerOptions to types.ServerConfig
func (s *Server) convertServerConfig() types.ServerConfig {
	return types.ServerConfig{
		Host:         s.opts.Server.Host,
		Port:         s.opts.Server.Port,
		ReadTimeout:  s.opts.Server.ReadTimeout,
		WriteTimeout: s.opts.Server.WriteTimeout,
		GracefulStop: s.opts.Server.GracefulStop,
	}
}

// shutdown gracefully shuts down the server
func (s *Server) shutdown() error {
	// Stop API server
	if s.apiServer != nil {
		if err := s.apiServer.Stop(); err != nil {
			s.logger.Warn("Failed to stop API server", zap.Error(err))
		}
	}

	// Stop NATS server
	if s.natsServer != nil {
		if err := s.natsServer.Stop(); err != nil {
			s.logger.Warn("Failed to stop NATS server", zap.Error(err))
		}
	}

	// Close database connection
	if s.pgStore != nil {
		if err := s.pgStore.Close(); err != nil {
			s.logger.Warn("Failed to close database", zap.Error(err))
		}
	}

	// Close Redis connection
	if s.redisStore != nil {
		if err := s.redisStore.Close(); err != nil {
			s.logger.Warn("Failed to close Redis", zap.Error(err))
		}
	}

	return nil
}

// SetupGinMode sets the gin mode based on server mode
func SetupGinMode(mode string) {
	switch mode {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
}
