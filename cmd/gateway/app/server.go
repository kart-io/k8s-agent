package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kart-io/logger/core"
	"github.com/redis/go-redis/v9"

	"github.com/kart-io/k8s-agent/internal/gateway/config"
	"github.com/kart-io/k8s-agent/internal/gateway/middleware"
	"github.com/kart-io/k8s-agent/internal/gateway/router"
)

// Server represents the gateway server
type Server struct {
	opts    *config.Options
	log     core.Logger
	rdb     *redis.Client
	httpSrv *http.Server
	router  http.Handler
}

// NewServer creates a new gateway server
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
	// Connect to Redis
	s.rdb = s.connectRedis()
	if s.rdb != nil {
		// Initialize rate limiter
		middleware.InitRateLimiter(s.rdb)
	}

	// Setup router with unified logger
	s.router = router.Setup(s.log)

	// Create HTTP server
	serverAddr := fmt.Sprintf("%s:%d", s.opts.Server.Host, s.opts.Server.Port)
	s.httpSrv = &http.Server{
		Addr:         serverAddr,
		Handler:      s.router,
		ReadTimeout:  s.opts.Server.ReadTimeout,
		WriteTimeout: s.opts.Server.WriteTimeout,
	}

	return nil
}

// connectRedis connects to Redis
func (s *Server) connectRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     s.opts.Redis.Addr,
		Password: s.opts.Redis.Password,
		DB:       s.opts.Redis.DB,
		PoolSize: s.opts.Redis.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		s.log.Warnw("Failed to connect to Redis, rate limiting will use local mode",
			"error", err,
		)
		return nil
	}

	s.log.Infow("Connected to Redis",
		"addr", s.opts.Redis.Addr,
	)

	return rdb
}

// Run starts the gateway server
func (s *Server) Run(ctx context.Context) error {
	// Start HTTP server
	go func() {
		s.log.Infow("Gateway server started",
			"addr", s.httpSrv.Addr,
			"mode", s.opts.Server.Mode,
		)
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Fatalw("Failed to start server",
				"error", err,
			)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.log.Info("Shutting down gateway server...")

	return s.Shutdown()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if s.httpSrv != nil {
		if err := s.httpSrv.Shutdown(ctx); err != nil {
			s.log.Errorw("Gateway server forced to shutdown",
				"error", err,
			)
			return err
		}
	}

	if s.rdb != nil {
		s.rdb.Close()
	}

	s.log.Info("Gateway server exited")
	return nil
}
