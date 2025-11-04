package app

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/kart-io/k8s-agent/common/options"
	commonserver "github.com/kart-io/k8s-agent/common/server"
	httpserver "github.com/kart-io/k8s-agent/common/server/http"
	"github.com/kart-io/k8s-agent/internal/gateway/config"
	"github.com/kart-io/k8s-agent/internal/gateway/middleware"
	"github.com/kart-io/k8s-agent/internal/gateway/router"
	"github.com/kart-io/logger/core"
)

// GatewayService represents the gateway service using common/server
type GatewayService struct {
	opts   *config.Options
	log    core.Logger
	rdb    *redis.Client
	server commonserver.Server
}

// NewServer creates a new gateway service (使用 common/server)
func NewServer(opts *config.Options, log core.Logger) (*GatewayService, error) {
	svc := &GatewayService{
		opts: opts,
		log:  log,
	}

	if err := svc.initialize(); err != nil {
		return nil, err
	}

	return svc, nil
}

// initialize initializes all server components
func (s *GatewayService) initialize() error {
	// Connect to Redis
	s.rdb = s.connectRedis()
	if s.rdb != nil {
		// Initialize rate limiter
		middleware.InitRateLimiter(s.rdb)
	}

	// Setup router with unified logger
	routerHandler := router.Setup(s.log)

	// Create Gin server config using common/server
	ginConfig := httpserver.NewGinServerConfig(&options.ServerOptions{
		Host:         s.opts.Server.Host,
		Port:         s.opts.Server.Port,
		Mode:         s.opts.Server.Mode,
		ReadTimeout:  s.opts.Server.ReadTimeout,
		WriteTimeout: s.opts.Server.WriteTimeout,
		IdleTimeout:  s.opts.Server.IdleTimeout,
	})

	// Create Gin server
	ginServer := httpserver.NewGinServerFromFullConfig(s.log, ginConfig)

	// Register the gateway router as a catch-all handler
	ginServer.GetEngine().Any("/*path", gin.WrapH(routerHandler))

	s.server = ginServer

	return nil
}

// connectRedis connects to Redis
func (s *GatewayService) connectRedis() *redis.Client {
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

// Run starts the gateway service using common/server.Serve()
func (s *GatewayService) Run(ctx context.Context) error {
	s.log.Infow("Starting Gateway Service",
		"addr", fmt.Sprintf("%s:%d", s.opts.Server.Host, s.opts.Server.Port),
		"mode", s.opts.Server.Mode,
	)

	// 使用 common/server 的标准 Serve 方法
	// 它会自动处理信号和优雅关停
	return commonserver.Serve(ctx, s.server, s.log)
}

// GetServer returns the server instance (实现 ServerProvider 接口)
func (s *GatewayService) GetServer() commonserver.Server {
	return s.server
}

// Cleanup cleans up resources
func (s *GatewayService) Cleanup() error {
	if s.rdb != nil {
		return s.rdb.Close()
	}
	return nil
}
