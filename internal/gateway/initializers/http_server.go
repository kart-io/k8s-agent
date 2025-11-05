package initializers

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/cmd/gateway/app/options"
	"github.com/kart-io/k8s-agent/internal/gateway/handler"
	"github.com/kart-io/k8s-agent/internal/gateway/middleware"
	"github.com/kart-io/k8s-agent/internal/gateway/proxy"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// HTTPServerInitializer wraps the common HTTP server initializer.
type HTTPServerInitializer struct {
	*pkginitializers.HTTPServerInitializer
	cfg          *options.ServerOptions
	logger       core.Logger
	redisInit    *RedisInitializer
	proxyHandler *proxy.Proxy
}

// NewHTTPServerInitializer creates a new HTTP server initializer.
func NewHTTPServerInitializer(
	cfg *options.ServerOptions,
	logger core.Logger,
	redisInit *RedisInitializer,
) *HTTPServerInitializer {
	h := &HTTPServerInitializer{
		cfg:       cfg,
		logger:    logger,
		redisInit: redisInit,
	}

	// Create standard HTTP server config
	serverConfig := &pkginitializers.HTTPServerConfig{
		Name:       "gateway-http-server",
		Priority:   bootstrap.PriorityHTTP,
		Config:     cfg.Server,
		RouteSetup: h.setupRoutes,
	}

	// Create standard HTTP server initializer
	h.HTTPServerInitializer = pkginitializers.NewHTTPServerInitializer(serverConfig, logger)
	return h
}

// Initialize initializes the HTTP server and creates gateway components.
func (h *HTTPServerInitializer) Initialize(ctx context.Context) error {
	// Initialize Redis-based middleware if Redis is available
	if h.redisInit.Client() != nil {
		middleware.InitRateLimiter(h.redisInit.Client())
	}

	// Create proxy handler
	h.proxyHandler = proxy.NewProxy(h.logger)

	// Initialize HTTP server (creates server and calls setupRoutes)
	return h.HTTPServerInitializer.Initialize(ctx)
}

// setupRoutes configures all routes for the gateway service.
func (h *HTTPServerInitializer) setupRoutes(engine *gin.Engine) error {
	h.logger.Infow("Setting up gateway service routes")

	// Global middleware
	if h.cfg.CORS.Enabled {
		engine.Use(middleware.CORS())
	}
	if h.cfg.RateLimit.Enabled {
		engine.Use(middleware.RateLimit())
	}

	// Health check handler
	healthHandler := handler.NewHealthHandler(h.proxyHandler, h.logger)

	// Gateway's own health check
	engine.GET("/health", healthHandler.GatewayHealth)

	// API route group
	api := engine.Group("/api/v1")
	{
		// Auth service routes (no authentication required)
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", h.proxyHandler.HandleRequest("auth"))
			authGroup.POST("/logout", h.proxyHandler.HandleRequest("auth"))
			authGroup.GET("/me", h.proxyHandler.HandleRequest("auth"))
			authGroup.GET("/menus", h.proxyHandler.HandleRequest("auth"))
			authGroup.POST("/check", h.proxyHandler.HandleRequest("auth"))
		}

		// Routes requiring authentication
		authenticated := api.Group("")
		authenticated.Use(middleware.JWTAuth())
		{
			// Agent management routes
			agentGroup := authenticated.Group("/agents")
			{
				agentGroup.GET("", h.proxyHandler.HandleRequest("agent_manager"))
				agentGroup.GET("/:id", h.proxyHandler.HandleRequest("agent_manager"))
				agentGroup.POST("", h.proxyHandler.HandleRequest("agent_manager"))
				agentGroup.PUT("/:id", h.proxyHandler.HandleRequest("agent_manager"))
				agentGroup.DELETE("/:id", h.proxyHandler.HandleRequest("agent_manager"))
			}

			// Cluster management routes
			clusterGroup := authenticated.Group("/clusters")
			{
				clusterGroup.GET("", h.proxyHandler.HandleRequest("agent_manager"))
				clusterGroup.GET("/:id", h.proxyHandler.HandleRequest("agent_manager"))
				clusterGroup.POST("", h.proxyHandler.HandleRequest("agent_manager"))
				clusterGroup.PUT("/:id", h.proxyHandler.HandleRequest("agent_manager"))
				clusterGroup.DELETE("/:id", h.proxyHandler.HandleRequest("agent_manager"))
			}

			// Event management routes
			eventGroup := authenticated.Group("/events")
			{
				eventGroup.GET("", h.proxyHandler.HandleRequest("agent_manager"))
				eventGroup.GET("/:id", h.proxyHandler.HandleRequest("agent_manager"))
			}

			// Command management routes
			commandGroup := authenticated.Group("/commands")
			{
				commandGroup.GET("", h.proxyHandler.HandleRequest("agent_manager"))
				commandGroup.GET("/:id", h.proxyHandler.HandleRequest("agent_manager"))
				commandGroup.POST("", h.proxyHandler.HandleRequest("agent_manager"))
			}

			// Workflow routes
			workflowGroup := authenticated.Group("/workflows")
			{
				workflowGroup.GET("", h.proxyHandler.HandleRequest("orchestrator"))
				workflowGroup.GET("/:id", h.proxyHandler.HandleRequest("orchestrator"))
				workflowGroup.POST("", h.proxyHandler.HandleRequest("orchestrator"))
				workflowGroup.PUT("/:id", h.proxyHandler.HandleRequest("orchestrator"))
				workflowGroup.DELETE("/:id", h.proxyHandler.HandleRequest("orchestrator"))
			}

			// Strategy routes
			strategyGroup := authenticated.Group("/strategies")
			{
				strategyGroup.GET("", h.proxyHandler.HandleRequest("orchestrator"))
				strategyGroup.GET("/:id", h.proxyHandler.HandleRequest("orchestrator"))
				strategyGroup.POST("", h.proxyHandler.HandleRequest("orchestrator"))
				strategyGroup.PUT("/:id", h.proxyHandler.HandleRequest("orchestrator"))
				strategyGroup.DELETE("/:id", h.proxyHandler.HandleRequest("orchestrator"))
			}

			// Reasoning service routes
			reasoningGroup := authenticated.Group("/reasoning")
			{
				reasoningGroup.POST("/analyze", h.proxyHandler.HandleRequest("reasoning"))
				reasoningGroup.POST("/suggest", h.proxyHandler.HandleRequest("reasoning"))
			}

			// User management routes
			userGroup := authenticated.Group("/users")
			{
				userGroup.GET("", h.proxyHandler.HandleRequest("auth"))
				userGroup.GET("/:id", h.proxyHandler.HandleRequest("auth"))
				userGroup.POST("", h.proxyHandler.HandleRequest("auth"))
				userGroup.PUT("/:id", h.proxyHandler.HandleRequest("auth"))
				userGroup.DELETE("/:id", h.proxyHandler.HandleRequest("auth"))
			}

			// Role management routes
			roleGroup := authenticated.Group("/roles")
			{
				roleGroup.GET("", h.proxyHandler.HandleRequest("auth"))
				roleGroup.GET("/:id", h.proxyHandler.HandleRequest("auth"))
				roleGroup.POST("", h.proxyHandler.HandleRequest("auth"))
				roleGroup.PUT("/:id", h.proxyHandler.HandleRequest("auth"))
				roleGroup.DELETE("/:id", h.proxyHandler.HandleRequest("auth"))
				roleGroup.POST("/:id/permissions", h.proxyHandler.HandleRequest("auth"))
			}

			// Permission management routes
			permissionGroup := authenticated.Group("/permissions")
			{
				permissionGroup.GET("", h.proxyHandler.HandleRequest("auth"))
				permissionGroup.GET("/tree", h.proxyHandler.HandleRequest("auth"))
				permissionGroup.GET("/:id", h.proxyHandler.HandleRequest("auth"))
				permissionGroup.POST("", h.proxyHandler.HandleRequest("auth"))
				permissionGroup.PUT("/:id", h.proxyHandler.HandleRequest("auth"))
				permissionGroup.DELETE("/:id", h.proxyHandler.HandleRequest("auth"))
			}
		}

		// Service health check
		api.GET("/health/services", healthHandler.ServicesHealth)
		api.GET("/health/services/:service", healthHandler.ServiceHealth)
	}

	// Metrics endpoint
	if h.cfg.Metrics.Enabled {
		engine.GET(h.cfg.Metrics.Path, handler.MetricsHandler)
	}

	// Load custom routes
	h.loadCustomRoutes(engine)

	h.logger.Infow("Gateway service routes registered",
		"health", "GET /health",
		"auth", "POST /api/v1/auth/login, logout, etc.",
		"agents", "GET/POST/PUT/DELETE /api/v1/agents",
		"clusters", "GET/POST/PUT/DELETE /api/v1/clusters",
		"workflows", "GET/POST/PUT/DELETE /api/v1/workflows",
		"strategies", "GET/POST/PUT/DELETE /api/v1/strategies",
		"reasoning", "POST /api/v1/reasoning/analyze, suggest",
		"users", "GET/POST/PUT/DELETE /api/v1/users",
		"roles", "GET/POST/PUT/DELETE /api/v1/roles",
		"permissions", "GET/POST/PUT/DELETE /api/v1/permissions",
		"metrics", h.cfg.Metrics.Enabled,
		"cors", h.cfg.CORS.Enabled,
		"rate_limit", h.cfg.RateLimit.Enabled,
	)

	return nil
}

// loadCustomRoutes loads custom route configurations.
func (h *HTTPServerInitializer) loadCustomRoutes(engine *gin.Engine) {
	if len(h.cfg.Routes) == 0 {
		return
	}

	for _, route := range h.cfg.Routes {
		// Add middleware based on authentication requirement
		if route.AuthRequired {
			engine.Any(route.Path, middleware.JWTAuth(), h.proxyHandler.HandleRequest(route.Service))
		} else {
			engine.Any(route.Path, h.proxyHandler.HandleRequest(route.Service))
		}
	}

	h.logger.Infow("Custom routes loaded", "count", len(h.cfg.Routes))
}
