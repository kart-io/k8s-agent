package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/logger/core"

	"github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/internal/agent-manager/agent"
	"github.com/kart-io/k8s-agent/internal/agent-manager/api"
	"github.com/kart-io/k8s-agent/internal/agent-manager/command"
	"github.com/kart-io/k8s-agent/internal/agent-manager/config"
	"github.com/kart-io/k8s-agent/internal/agent-manager/event"
	agentgrpc "github.com/kart-io/k8s-agent/internal/agent-manager/grpc"
	"github.com/kart-io/k8s-agent/internal/agent-manager/nats"
	"github.com/kart-io/k8s-agent/internal/agent-manager/storage"
	"github.com/kart-io/k8s-agent/pkg/types"
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
	grpcServer     *agentgrpc.Server
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

	// Start gRPC server if enabled
	if s.opts.GRPC.Enable {
		go func() {
			if err := s.grpcServer.Start(ctx); err != nil {
				s.logger.Errorw("gRPC server error", "error", err)
			}
		}()
	}

	// Start API server
	errCh := make(chan error, 1) // Error channel for API server
	go func() {
		if err := s.apiServer.Start(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("failed to start API server: %w", err)
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		s.logger.Infow("Shutting down server")
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

	// Initialize gRPC server if enabled
	if s.opts.GRPC.Enable {
		grpcOpts := &agentgrpc.ServerOptions{
			Host:             s.opts.GRPC.Host,
			Port:             s.opts.GRPC.Port,
			MaxRecvMsgSize:   s.opts.GRPC.MaxRecvMsgSize,
			MaxSendMsgSize:   s.opts.GRPC.MaxSendMsgSize,
			KeepaliveTime:    s.opts.GRPC.KeepAliveTime,
			KeepaliveTimeout: s.opts.GRPC.KeepAliveTimeout,
			Registry:         s.registry,
			Dispatcher:       s.dispatcher,
			Store:            s.pgStore,
		}

		grpcServer, err := agentgrpc.NewServer(grpcOpts, s.logger)
		if err != nil {
			return fmt.Errorf("failed to create gRPC server: %w", err)
		}
		s.grpcServer = grpcServer

		s.logger.Infow("gRPC server initialized",
			"address", grpcServer.Address(),
		)
	}

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

	s.logger.Infow("Database initialized",
		"host", s.opts.Database.Host,
		"database", s.opts.Database.Database,
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

	s.logger.Infow("Redis initialized",
		"addr", s.opts.Redis.Addr,
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
			s.logger.Warnw("Failed to stop API server", "error", err)
		}
	}

	// Stop gRPC server
	if s.grpcServer != nil {
		if err := s.grpcServer.Stop(); err != nil {
			s.logger.Warnw("Failed to stop gRPC server", "error", err)
		}
	}

	// Stop NATS server
	if s.natsServer != nil {
		if err := s.natsServer.Stop(); err != nil {
			s.logger.Warnw("Failed to stop NATS server", "error", err)
		}
	}

	// Close database connection
	if s.pgStore != nil {
		if err := s.pgStore.Close(); err != nil {
			s.logger.Warnw("Failed to close database", "error", err)
		}
	}

	// Close Redis connection
	if s.redisStore != nil {
		if err := s.redisStore.Close(); err != nil {
			s.logger.Warnw("Failed to close Redis", "error", err)
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
