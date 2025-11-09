package initializers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	// "github.com/kart-io/k8s-agent/common/middleware" // TODO: Re-enable when needed.
	commonserver "github.com/kart-io/k8s-agent/common/server"
	"github.com/kart-io/k8s-agent/internal/agent-manager/api"
	agentgrpc "github.com/kart-io/k8s-agent/internal/agent-manager/grpc"
	agentv1 "github.com/kart-io/k8s-agent/pkg/api/agent/v1"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	// "github.com/kart-io/k8s-agent/pkg/idempotent" // TODO: Re-enable when idempotent middleware is implemented.
	commoninitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/k8s-agent/pkg/types"
	"github.com/kart-io/logger/core"
)

// HTTPServerInitializer is a wrapper around the common HTTP server initializer.
// Refactored to follow ServiceInitializer pattern: only depends on ServiceInitializer.
type HTTPServerInitializer struct {
	standardInit *commoninitializers.HTTPServerInitializer
	logger       core.Logger
	opts         *commonapp.StandardOptions
	apiServer    *api.Server // Reused for its handler methods and dependency injection

	// Single dependency: ServiceInitializer (delegates service access)
	serviceInit *ServiceInitializer
	dbInit      *DatabaseInitializer
	redisInit   *RedisInitializer
}

// NewHTTPServerInitializer creates a new HTTP server initializer.
// Dependencies: ONLY ServiceInitializer (no direct access to Registry/Dispatcher/NATS)
func NewHTTPServerInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	serviceInit *ServiceInitializer,
	dbInit *DatabaseInitializer,
	redisInit *RedisInitializer,
) *HTTPServerInitializer {
	return &HTTPServerInitializer{
		opts:        opts,
		logger:      logger,
		serviceInit: serviceInit,
		dbInit:      dbInit,
		redisInit:   redisInit,
	}
}

func (h *HTTPServerInitializer) Name() string {
	return "agent-manager-http-server"
}

func (h *HTTPServerInitializer) Priority() int {
	return bootstrap.PriorityHTTP
}

func (h *HTTPServerInitializer) Initialize(ctx context.Context) error {
	// Create the api.Server which holds all dependencies and handler logic.
	// We will use it to register routes, but not to start the server itself.
	// Get services from ServiceInitializer (delegation pattern)
	h.apiServer = api.NewServer(
		types.ServerConfig{
			// This config is now managed by the common server,
			// but api.NewServer needs it. We pass it for consistency.
			Host: h.opts.Server.Host,
			Port: h.opts.Server.Port,
		},
		h.serviceInit.Registry(),       // Get from ServiceInitializer
		h.serviceInit.EventProcessor(), // Get from ServiceInitializer
		h.serviceInit.Dispatcher(),     // Get from ServiceInitializer
		h.dbInit.Store(),
		h.redisInit.Store(),
		h.logger,
	)

	serverConfig := &commoninitializers.HTTPServerConfig{
		Name:     h.Name(),
		Priority: h.Priority(),
		Config:   h.opts.Server,
		RouteSetup: func(engine *gin.Engine) error {
			// The standard GinServer already adds recovery, logging, cors, requestid.
			// TODO: Re-enable idempotency middleware once it's properly implemented as Gin middleware
			// Currently the idempotent package doesn't provide a Gin middleware function
			/*
				if h.redisInit.Store() != nil && h.redisInit.Store().Client != nil {
					redisStore := idempotent.NewRedisStore(h.redisInit.Store().Client, "agent-manager")
					idempotentHandler := idempotent.NewHandler(redisStore, 24*time.Hour, 5*time.Minute)
					// Need to create a Gin middleware wrapper for the idempotent handler
					// engine.Use(idempotentMiddleware(idempotentHandler))
					h.logger.Info("Idempotency middleware enabled for POST operations")
				} else {
					h.logger.Warn("Redis not available, idempotency middleware disabled")
				}
			*/

			// Register all API routes using handlers from h.apiServer
			// Health endpoints
			health := engine.Group("/health")
			{
				health.GET("/live", h.apiServer.HandleLiveness)
				health.GET("/ready", h.apiServer.HandleReadiness)
				health.GET("/status", h.apiServer.HandleStatus)
			}

			// Metrics endpoint
			engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

			// API v1
			v1 := engine.Group("/api/v1")
			{
				// Agent management
				agents := v1.Group("/agents")
				{
					agents.GET("", h.apiServer.HandleListAgents)
					agents.GET("/:id", h.apiServer.HandleGetAgent)
					agents.DELETE("/:id", h.apiServer.HandleDeleteAgent)
				}

				// Cluster management
				clusters := v1.Group("/clusters")
				{
					clusters.GET("", h.apiServer.HandleListClusters)
					clusters.GET("/:id", h.apiServer.HandleGetCluster)
					clusters.POST("", h.apiServer.HandleCreateCluster)
					clusters.PUT("/:id", h.apiServer.HandleUpdateCluster)
					clusters.DELETE("/:id", h.apiServer.HandleDeleteCluster)
					clusters.GET("/:id/health", h.apiServer.HandleClusterHealth)
				}

				// Event management
				events := v1.Group("/events")
				{
					events.GET("", h.apiServer.HandleListEvents)
					events.GET("/:id", h.apiServer.HandleGetEvent)
					events.POST("/search", h.apiServer.HandleSearchEvents)
				}

				// Command management
				commands := v1.Group("/commands")
				{
					commands.POST("", h.apiServer.HandleSendCommand)
					commands.GET("/:id", h.apiServer.HandleGetCommand)
					commands.GET("/:id/result", h.apiServer.HandleGetCommandResult)
					commands.GET("/:id/events", h.apiServer.HandleGetCommandEvents)
					commands.GET("", h.apiServer.HandleListPendingCommands)
				}

				// Operation tracking
				operations := v1.Group("/operations")
				{
					operations.POST("", h.apiServer.HandleRecordOperation)
					operations.GET("/:id/events", h.apiServer.HandleGetOperationEvents)
				}
			}
			h.logger.Info("All HTTP API routes registered")
			return nil
		},
	}

	h.standardInit = commoninitializers.NewHTTPServerInitializer(serverConfig, h.logger)
	return h.standardInit.Initialize(ctx)
}

// GetServer implements commonserver.ServerProvider.
func (h *HTTPServerInitializer) GetServer() commonserver.Server {
	if h.standardInit == nil {
		return nil
	}
	return h.standardInit.GetServer()
}

// Close is a no-op because the server lifecycle is managed by bootstrap.
func (h *HTTPServerInitializer) Close(ctx context.Context) error {
	return nil
}

// GRPCServerInitializer is a wrapper around the common gRPC server initializer.
// Refactored to follow ServiceInitializer pattern: only depends on ServiceInitializer.
type GRPCServerInitializer struct {
	standardInit *commoninitializers.GRPCServerInitializer
	logger       core.Logger
	opts         *commonapp.StandardOptions

	// Single dependency: ServiceInitializer (delegates service access)
	serviceInit *ServiceInitializer
	dbInit      *DatabaseInitializer
}

// NewGRPCServerInitializer creates a new gRPC server initializer.
// Dependencies: ONLY ServiceInitializer (no direct access to Registry/Dispatcher)
func NewGRPCServerInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	serviceInit *ServiceInitializer,
	dbInit *DatabaseInitializer,
) *GRPCServerInitializer {
	return &GRPCServerInitializer{
		opts:        opts,
		logger:      logger,
		serviceInit: serviceInit,
		dbInit:      dbInit,
	}
}

func (g *GRPCServerInitializer) Name() string {
	return "agent-manager-grpc-server"
}

func (g *GRPCServerInitializer) Priority() int {
	return bootstrap.PriorityGRPC
}

func (g *GRPCServerInitializer) Initialize(ctx context.Context) error {
	if !g.opts.GRPC.Enable {
		g.logger.Infow("gRPC server is disabled, skipping initialization")
		return nil
	}

	// Get services from ServiceInitializer (delegation pattern)
	// This ensures services are created only once and shared between HTTP/gRPC
	registry := g.serviceInit.Registry()
	dispatcher := g.serviceInit.Dispatcher()

	// Create gRPC service servers (protocol wrappers)
	agentService := agentgrpc.NewAgentServiceServer(registry, g.logger)
	commandService := agentgrpc.NewCommandServiceServer(dispatcher, g.dbInit.Store(), g.logger)

	serverConfig := &commoninitializers.GRPCServerConfig{
		Name:     g.Name(),
		Priority: g.Priority(),
		Config:   g.opts.GRPC,
		ServiceRegister: func(s *grpc.Server) error {
			agentv1.RegisterAgentServiceServer(s, agentService)
			agentv1.RegisterCommandServiceServer(s, commandService)
			g.logger.Info("All gRPC services registered")
			return nil
		},
	}

	g.standardInit = commoninitializers.NewGRPCServerInitializer(serverConfig, g.logger)
	return g.standardInit.Initialize(ctx)
}

// GetServer implements commonserver.ServerProvider.
func (g *GRPCServerInitializer) GetServer() commonserver.Server {
	if g.standardInit == nil {
		return nil
	}
	return g.standardInit.GetServer()
}

// Close is a no-op because the server lifecycle is managed by bootstrap.
func (g *GRPCServerInitializer) Close(ctx context.Context) error {
	return nil
}
