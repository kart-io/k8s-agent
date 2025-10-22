package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/logger/core"

	"github.com/kart-io/k8s-agent/internal/agent-manager/agent"
	"github.com/kart-io/k8s-agent/internal/agent-manager/api"
	"github.com/kart-io/k8s-agent/internal/agent-manager/command"
	"github.com/kart-io/k8s-agent/internal/agent-manager/config"
	"github.com/kart-io/k8s-agent/internal/agent-manager/event"
	"github.com/kart-io/k8s-agent/internal/agent-manager/grpc"
	"github.com/kart-io/k8s-agent/internal/agent-manager/nats"
	"github.com/kart-io/k8s-agent/internal/agent-manager/storage"
	"github.com/kart-io/k8s-agent/pkg/types"
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
	grpcServer     *grpc.Server
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
	errCh := make(chan error, 2) // Increased capacity for both servers
	go func() {
		if err := s.apiServer.Start(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("failed to start API server: %w", err)
		}
	}()

	// Start gRPC server if enabled
	if s.opts.GRPC.Enable && s.grpcServer != nil {
		go func() {
			s.logger.Infow("Starting gRPC server",
				"address", fmt.Sprintf("%s:%d", s.opts.GRPC.Host, s.opts.GRPC.Port),
			)
			if err := s.grpcServer.Run(); err != nil {
				errCh <- fmt.Errorf("failed to start gRPC server: %w", err)
			}
		}()
	}

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
		// 准备 gRPC 服务依赖
		deps := &grpc.ServerDependencies{
			Registry:   s.registry,
			Dispatcher: s.dispatcher,
			Processor:  s.eventProcessor,
			AgentStore: &grpcAgentStoreAdapter{store: s.pgStore},
			EventStore: &grpcEventStoreAdapter{store: s.pgStore},
		}

		var err error
		s.grpcServer, err = grpc.NewServer(s.opts.GRPC, s.logger, deps)
		if err != nil {
			return fmt.Errorf("failed to create gRPC server: %w", err)
		}

		s.logger.Infow("gRPC server initialized",
			"enabled", true,
			"address", fmt.Sprintf("%s:%d", s.opts.GRPC.Host, s.opts.GRPC.Port),
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
	// Stop gRPC server
	if s.grpcServer != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.opts.Server.GracefulStop)
		defer cancel()

		if err := s.grpcServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Warnw("Failed to shutdown gRPC server gracefully", "error", err)
		}
	}

	// Stop API server
	if s.apiServer != nil {
		if err := s.apiServer.Stop(); err != nil {
			s.logger.Warnw("Failed to stop API server", "error", err)
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

// grpcAgentStoreAdapter 适配 PostgresStore 到 AgentStore 接口
type grpcAgentStoreAdapter struct {
	store *storage.PostgresStore
}

func (a *grpcAgentStoreAdapter) GetAgentMetrics(ctx context.Context, agentID string, startTime, endTime *time.Time) ([]*types.Metrics, error) {
	// TODO: 实现从数据库查询指标数据的逻辑
	// 目前返回空数组，后续可以根据需求实现
	return []*types.Metrics{}, nil
}

// grpcEventStoreAdapter 适配 PostgresStore 到 EventStore 接口
type grpcEventStoreAdapter struct {
	store *storage.PostgresStore
}

func (a *grpcEventStoreAdapter) GetEvent(ctx context.Context, eventID string) (*types.Event, error) {
	return a.store.GetEvent(ctx, eventID)
}

func (a *grpcEventStoreAdapter) ListEvents(ctx context.Context, filters interface{}) ([]*types.Event, error) {
	// TODO: 实现更复杂的过滤逻辑
	filter := storage.EventFilter{
		Limit: 100,
	}
	return a.store.ListEvents(ctx, filter)
}

func (a *grpcEventStoreAdapter) SearchEvents(ctx context.Context, query interface{}) ([]*types.Event, error) {
	// TODO: 实现搜索逻辑
	filter := storage.EventFilter{
		Limit: 100,
	}
	return a.store.ListEvents(ctx, filter)
}
