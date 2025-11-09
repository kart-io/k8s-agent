// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/kart-io/k8s-agent/cmd/gateway/app/options"
	"github.com/kart-io/k8s-agent/internal/gateway/handler"
	"github.com/kart-io/k8s-agent/internal/gateway/middleware"
	"github.com/kart-io/k8s-agent/internal/gateway/proxy"
	"github.com/kart-io/k8s-agent/internal/gateway/types"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// GatewayContainer holds all component initializers.
type GatewayContainer struct {
	Redis  *RedisInitializer
	HTTP   *HTTPServerInitializer
	Health *pkginitializers.HealthCheckInitializer
}

// NewGatewayContainer creates component bundle (called by Wire).
func NewGatewayContainer(
	redis *RedisInitializer,
	http *HTTPServerInitializer,
	health *pkginitializers.HealthCheckInitializer,
) *GatewayContainer {
	return &GatewayContainer{
		Redis:  redis,
		HTTP:   http,
		Health: health,
	}
}

// ====================================================================
// RedisInitializer - Wrapper around pkg/initializers with optional init
// ====================================================================

// RedisInitializer wraps pkg/initializers.RedisInitializer with optional initialization
// for gateway service (non-fatal if Redis is unavailable).
type RedisInitializer struct {
	*pkginitializers.RedisInitializer
	opts          *options.ServerOptions
	logger        core.Logger
	isInitialized bool
}

// NewRedisInitializer creates Redis initializer for gateway.
// Redis is optional for gateway - if unavailable, rate limiting falls back to local mode.
func NewRedisInitializer(opts *options.ServerOptions, logger core.Logger) *RedisInitializer {
	wrapper := &RedisInitializer{
		opts:   opts,
		logger: logger,
	}

	// Only create base initializer if Redis is configured
	if opts.Redis != nil && opts.Redis.Addr != "" {
		wrapper.RedisInitializer = pkginitializers.NewRedisInitializer(opts.Redis, logger)
	}

	return wrapper
}

// Name returns initializer name.
func (r *RedisInitializer) Name() string {
	return "gateway-redis"
}

// Priority returns initialization priority.
func (r *RedisInitializer) Priority() int {
	return bootstrap.PriorityRedis
}

// Initialize connects to Redis (non-fatal if fails).
func (r *RedisInitializer) Initialize(ctx context.Context) error {
	// If Redis not configured, skip initialization
	if r.RedisInitializer == nil {
		r.logger.Info("Redis is not configured for gateway")
		return nil
	}

	// Try to initialize Redis
	if err := r.RedisInitializer.Initialize(ctx); err != nil {
		r.logger.Warnw("Failed to connect to Redis (rate limiting will use local mode)", "error", err)
		// Non-fatal for gateway - continue without Redis
		return nil
	}

	// Verify connection works
	if err := r.HealthCheck(ctx); err != nil {
		r.logger.Warnw("Redis health check failed (rate limiting will use local mode)", "error", err)
		// Non-fatal for gateway - continue without Redis
		return nil
	}

	r.isInitialized = true
	r.logger.Infow("Redis connected successfully",
		"addr", r.opts.Redis.Addr,
	)
	return nil
}

// Client returns Redis client (may be nil if not initialized).
func (r *RedisInitializer) Client() *redis.Client {
	if r.RedisInitializer == nil || !r.isInitialized {
		return nil
	}
	return r.RedisInitializer.Client()
}

// IsAvailable returns true if Redis is available and initialized.
func (r *RedisInitializer) IsAvailable() bool {
	return r.isInitialized && r.Client() != nil
}

// ====================================================================
// HTTPServerInitializer - Wrapper around pkg/initializers
// ====================================================================

// HTTPServerInitializer wraps pkg/initializers.HTTPServerInitializer with gateway-specific logic.
type HTTPServerInitializer struct {
	*pkginitializers.HTTPServerInitializer
	opts         *options.ServerOptions
	logger       core.Logger
	redisInit    *RedisInitializer
	proxyHandler *proxy.Proxy
}

// NewHTTPServerInitializer creates HTTP server initializer for gateway.
func NewHTTPServerInitializer(
	opts *options.ServerOptions,
	logger core.Logger,
	redisInit *RedisInitializer,
) *HTTPServerInitializer {
	h := &HTTPServerInitializer{
		opts:      opts,
		logger:    logger,
		redisInit: redisInit,
	}

	// Create standard HTTP server config with route setup callback
	serverConfig := &pkginitializers.HTTPServerConfig{
		Name:       "gateway-http-server",
		Priority:   bootstrap.PriorityHTTP,
		Config:     opts.Server,
		RouteSetup: h.setupRoutes,
		CORS:       opts.CORS,
	}

	h.HTTPServerInitializer = pkginitializers.NewHTTPServerInitializer(serverConfig, logger)
	return h
}

// Initialize initializes gateway-specific components before starting HTTP server.
func (h *HTTPServerInitializer) Initialize(ctx context.Context) error {
	h.logger.Infow("Initializing gateway HTTP server")

	// Initialize Redis-based rate limiter if available
	if h.redisInit.IsAvailable() {
		middleware.InitRateLimiter(h.redisInit.Client())
		h.logger.Infow("Redis-based rate limiter initialized")
	} else {
		h.logger.Infow("Using local rate limiter (Redis not available)")
	}

	// Create proxy handler with services configuration
	// Convert ServicesOptions to map[string]types.ServiceConfig
	services := map[string]types.ServiceConfig{
		"auth": {
			Name:        h.opts.Services.Auth.Name,
			URL:         h.opts.Services.Auth.URL,
			Timeout:     h.opts.Services.Auth.Timeout,
			Retry:       h.opts.Services.Auth.Retry,
			HealthCheck: h.opts.Services.Auth.HealthCheck,
		},
		"agent_manager": {
			Name:        h.opts.Services.AgentManager.Name,
			URL:         h.opts.Services.AgentManager.URL,
			Timeout:     h.opts.Services.AgentManager.Timeout,
			Retry:       h.opts.Services.AgentManager.Retry,
			HealthCheck: h.opts.Services.AgentManager.HealthCheck,
		},
		"orchestrator": {
			Name:        h.opts.Services.Orchestrator.Name,
			URL:         h.opts.Services.Orchestrator.URL,
			Timeout:     h.opts.Services.Orchestrator.Timeout,
			Retry:       h.opts.Services.Orchestrator.Retry,
			HealthCheck: h.opts.Services.Orchestrator.HealthCheck,
		},
		"reasoning": {
			Name:        h.opts.Services.Reasoning.Name,
			URL:         h.opts.Services.Reasoning.URL,
			Timeout:     h.opts.Services.Reasoning.Timeout,
			Retry:       h.opts.Services.Reasoning.Retry,
			HealthCheck: h.opts.Services.Reasoning.HealthCheck,
		},
	}
	h.proxyHandler = proxy.NewProxyWithServices(services, h.logger)

	// Initialize standard HTTP server (calls setupRoutes callback)
	return h.HTTPServerInitializer.Initialize(ctx)
}

// setupRoutes configures all gateway routes (called by pkg/initializers.HTTPServerInitializer).
func (h *HTTPServerInitializer) setupRoutes(engine *gin.Engine) error {
	h.logger.Infow("Setting up gateway routes")

	// Global middleware
	if h.opts.RateLimit != nil && h.opts.RateLimit.Enable {
		engine.Use(middleware.RateLimit())
	}

	// Health check handler
	healthHandler := handler.NewHealthHandler(h.proxyHandler, h.logger)
	engine.GET("/health", healthHandler.GatewayHealth)

	// API routes
	api := engine.Group("/api/v1")
	{
		// Auth routes (no auth required)
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", h.proxyHandler.HandleRequest("auth"))
			authGroup.POST("/logout", h.proxyHandler.HandleRequest("auth"))
			authGroup.GET("/me", h.proxyHandler.HandleRequest("auth"))
			authGroup.GET("/menus", h.proxyHandler.HandleRequest("auth"))
			authGroup.POST("/check", h.proxyHandler.HandleRequest("auth"))
		}

		// Authenticated routes
		authenticated := api.Group("")
		authenticated.Use(middleware.JWTAuth())
		{
			// Agent routes
			authenticated.GET("/agents", h.proxyHandler.HandleRequest("agent_manager"))
			authenticated.GET("/agents/:id", h.proxyHandler.HandleRequest("agent_manager"))
			authenticated.POST("/agents", h.proxyHandler.HandleRequest("agent_manager"))
			authenticated.PUT("/agents/:id", h.proxyHandler.HandleRequest("agent_manager"))
			authenticated.DELETE("/agents/:id", h.proxyHandler.HandleRequest("agent_manager"))

			// Cluster routes
			authenticated.GET("/clusters", h.proxyHandler.HandleRequest("agent_manager"))
			authenticated.GET("/clusters/:id", h.proxyHandler.HandleRequest("agent_manager"))
			authenticated.POST("/clusters", h.proxyHandler.HandleRequest("agent_manager"))
			authenticated.PUT("/clusters/:id", h.proxyHandler.HandleRequest("agent_manager"))
			authenticated.DELETE("/clusters/:id", h.proxyHandler.HandleRequest("agent_manager"))

			// Event routes
			authenticated.GET("/events", h.proxyHandler.HandleRequest("agent_manager"))
			authenticated.GET("/events/:id", h.proxyHandler.HandleRequest("agent_manager"))

			// Command routes
			authenticated.GET("/commands", h.proxyHandler.HandleRequest("agent_manager"))
			authenticated.GET("/commands/:id", h.proxyHandler.HandleRequest("agent_manager"))
			authenticated.POST("/commands", h.proxyHandler.HandleRequest("agent_manager"))

			// Workflow routes
			authenticated.GET("/workflows", h.proxyHandler.HandleRequest("orchestrator"))
			authenticated.GET("/workflows/:id", h.proxyHandler.HandleRequest("orchestrator"))
			authenticated.POST("/workflows", h.proxyHandler.HandleRequest("orchestrator"))
			authenticated.PUT("/workflows/:id", h.proxyHandler.HandleRequest("orchestrator"))
			authenticated.DELETE("/workflows/:id", h.proxyHandler.HandleRequest("orchestrator"))

			// Strategy routes
			authenticated.GET("/strategies", h.proxyHandler.HandleRequest("orchestrator"))
			authenticated.GET("/strategies/:id", h.proxyHandler.HandleRequest("orchestrator"))
			authenticated.POST("/strategies", h.proxyHandler.HandleRequest("orchestrator"))
			authenticated.PUT("/strategies/:id", h.proxyHandler.HandleRequest("orchestrator"))
			authenticated.DELETE("/strategies/:id", h.proxyHandler.HandleRequest("orchestrator"))

			// Reasoning routes
			authenticated.POST("/reasoning/analyze", h.proxyHandler.HandleRequest("reasoning"))
			authenticated.POST("/reasoning/suggest", h.proxyHandler.HandleRequest("reasoning"))

			// User routes
			authenticated.GET("/users", h.proxyHandler.HandleRequest("auth"))
			authenticated.GET("/users/:id", h.proxyHandler.HandleRequest("auth"))
			authenticated.POST("/users", h.proxyHandler.HandleRequest("auth"))
			authenticated.PUT("/users/:id", h.proxyHandler.HandleRequest("auth"))
			authenticated.DELETE("/users/:id", h.proxyHandler.HandleRequest("auth"))

			// Role routes
			authenticated.GET("/roles", h.proxyHandler.HandleRequest("auth"))
			authenticated.GET("/roles/:id", h.proxyHandler.HandleRequest("auth"))
			authenticated.POST("/roles", h.proxyHandler.HandleRequest("auth"))
			authenticated.PUT("/roles/:id", h.proxyHandler.HandleRequest("auth"))
			authenticated.DELETE("/roles/:id", h.proxyHandler.HandleRequest("auth"))
			authenticated.POST("/roles/:id/permissions", h.proxyHandler.HandleRequest("auth"))

			// Permission routes
			authenticated.GET("/permissions", h.proxyHandler.HandleRequest("auth"))
			authenticated.GET("/permissions/tree", h.proxyHandler.HandleRequest("auth"))
			authenticated.GET("/permissions/:id", h.proxyHandler.HandleRequest("auth"))
			authenticated.POST("/permissions", h.proxyHandler.HandleRequest("auth"))
			authenticated.PUT("/permissions/:id", h.proxyHandler.HandleRequest("auth"))
			authenticated.DELETE("/permissions/:id", h.proxyHandler.HandleRequest("auth"))
		}

		// Service health check
		api.GET("/health/services", healthHandler.ServicesHealth)
		api.GET("/health/services/:service", healthHandler.ServiceHealth)
	}

	// Metrics endpoint
	if h.opts.Metrics != nil && h.opts.Metrics.Enabled {
		engine.GET(h.opts.Metrics.Path, handler.MetricsHandler)
	}

	// Custom routes
	h.loadCustomRoutes(engine)

	h.logger.Infow("Gateway routes registered", "total_endpoints", "50+")
	return nil
}

// loadCustomRoutes loads custom configured routes (helper for setupRoutes).
func (h *HTTPServerInitializer) loadCustomRoutes(engine *gin.Engine) {
	if len(h.opts.Routes) == 0 {
		return
	}

	for _, route := range h.opts.Routes {
		if route.AuthRequired {
			engine.Any(route.Path, middleware.JWTAuth(), h.proxyHandler.HandleRequest(route.Service))
		} else {
			engine.Any(route.Path, h.proxyHandler.HandleRequest(route.Service))
		}
	}

	h.logger.Infow("Custom routes loaded", "count", len(h.opts.Routes))
}
