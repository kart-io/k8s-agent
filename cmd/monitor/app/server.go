package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/internal/monitor/api"
	"github.com/kart-io/k8s-agent/internal/monitor/config"
	"github.com/kart-io/k8s-agent/internal/monitor/handler"
	"github.com/kart-io/k8s-agent/internal/monitor/service"
	"github.com/kart-io/k8s-agent/internal/monitor/storage"
	"github.com/kart-io/logger/core"
)

// Server represents the monitor server
type Server struct {
	opts           *config.Options
	log            core.Logger
	pgStorage      *storage.PostgresStorage
	redisStorage   *storage.RedisStorage
	monitorService *service.MonitorService
	apiServer      *api.Server
}

// NewServer creates a new monitor server
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

	// Initialize PostgreSQL storage
	dbOpts := &options.DatabaseOptions{
		Host:         s.opts.Database.Host,
		Port:         s.opts.Database.Port,
		User:         s.opts.Database.User,
		Password:     s.opts.Database.Password,
		Database:     s.opts.Database.DBName,
		MaxOpenConns: s.opts.Database.MaxOpenConns,
		MaxIdleConns: s.opts.Database.MaxIdleConns,
	}
	s.pgStorage, err = storage.NewPostgresStorage(dbOpts, s.log)
	if err != nil {
		return fmt.Errorf("failed to initialize PostgreSQL storage: %w", err)
	}

	// Initialize Redis storage
	s.redisStorage, err = storage.NewRedisStorage(&storage.RedisConfig{
		Host:     s.opts.Redis.Host,
		Port:     s.opts.Redis.Port,
		Password: s.opts.Redis.Password,
		DB:       s.opts.Redis.DB,
		PoolSize: s.opts.Redis.PoolSize,
	}, s.log)
	if err != nil {
		s.pgStorage.Close()
		return fmt.Errorf("failed to initialize Redis storage: %w", err)
	}

	// Initialize monitor service
	s.monitorService = service.NewMonitorService(s.pgStorage, s.redisStorage, s.log)

	// Initialize metrics handler
	metricsHandler := handler.NewMetricsHandler(s.monitorService, s.log)

	// Parse timeout durations
	readTimeout, _ := time.ParseDuration(s.opts.Server.ReadTimeout)
	writeTimeout, _ := time.ParseDuration(s.opts.Server.WriteTimeout)

	// Determine metrics port
	metricsPort := 0
	if s.opts.Prometheus.Enabled {
		metricsPort = s.opts.Prometheus.Port
	}

	// Create API server
	s.apiServer = api.NewServer(&api.ServerConfig{
		Port:         s.opts.Server.Port,
		Mode:         s.opts.Server.Mode,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		JWTSecret:    s.opts.JWT.Secret,
		MetricsPort:  metricsPort,
	}, metricsHandler, s.log)

	return nil
}

// Run starts the monitor server
func (s *Server) Run(ctx context.Context) error {
	// Start API server
	go func() {
		if err := s.apiServer.Start(); err != nil {
			s.log.Fatalw("Server failed to start",
				"error", err,
			)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.log.Info("Shutting down server...")

	return s.Shutdown()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if s.apiServer != nil {
		if err := s.apiServer.Shutdown(ctx); err != nil {
			s.log.Errorw("Server forced to shutdown",
				"error", err,
			)
		}
	}

	if s.redisStorage != nil {
		s.redisStorage.Close()
	}

	if s.pgStorage != nil {
		s.pgStorage.Close()
	}

	s.log.Info("Server exited")
	return nil
}
