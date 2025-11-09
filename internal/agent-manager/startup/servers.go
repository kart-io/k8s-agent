// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package startup

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	commonserver "github.com/kart-io/k8s-agent/common/server"
	"github.com/kart-io/k8s-agent/internal/agent-manager/api"
	agentgrpc "github.com/kart-io/k8s-agent/internal/agent-manager/grpc"
	agentv1 "github.com/kart-io/k8s-agent/pkg/api/agent/v1"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/k8s-agent/pkg/types"
	"github.com/kart-io/logger/core"
)

// HTTPServerInitializer initializes the HTTP API server.
type HTTPServerInitializer struct {
	opts         *commonapp.StandardOptions
	logger       core.Logger
	infra        *InfrastructureInitializers
	services     *BusinessServicesInitializer
	standardInit *pkginitializers.HTTPServerInitializer
}

// NewHTTPServerInitializer creates a new HTTP server initializer.
func NewHTTPServerInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	infra *InfrastructureInitializers,
	services *BusinessServicesInitializer,
) *HTTPServerInitializer {
	return &HTTPServerInitializer{
		opts:     opts,
		logger:   logger,
		infra:    infra,
		services: services,
	}
}

// Name returns the initializer name.
func (h *HTTPServerInitializer) Name() string {
	return "agent-manager-http-server"
}

// Priority returns initialization priority.
func (h *HTTPServerInitializer) Priority() int {
	return 1000 // After all services
}

// Initialize creates and starts HTTP server.
func (h *HTTPServerInitializer) Initialize(ctx context.Context) error {
	h.logger.Infow("Initializing HTTP server")

	// Get business services
	businessServices := h.services.Services()

	// Create API server with all dependencies
	apiServer := api.NewServer(
		types.ServerConfig{
			Host: h.opts.Server.Host,
			Port: h.opts.Server.Port,
		},
		businessServices.Registry,
		businessServices.EventProcessor,
		businessServices.Dispatcher,
		h.infra.Database.Store(),
		h.infra.Redis.Store(),
		h.logger,
	)

	// Configure HTTP server
	serverConfig := &pkginitializers.HTTPServerConfig{
		Name:     h.Name(),
		Priority: h.Priority(),
		Config:   h.opts.Server,
		RouteSetup: func(engine *gin.Engine) error {
			// Health endpoints
			health := engine.Group("/health")
			{
				health.GET("/live", apiServer.HandleLiveness)
				health.GET("/ready", apiServer.HandleReadiness)
				health.GET("/status", apiServer.HandleStatus)
			}

			// Metrics endpoint
			engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

			// API v1
			v1 := engine.Group("/api/v1")
			{
				// Agent management
				agents := v1.Group("/agents")
				{
					agents.GET("", apiServer.HandleListAgents)
					agents.GET("/:id", apiServer.HandleGetAgent)
					agents.DELETE("/:id", apiServer.HandleDeleteAgent)
				}

				// Cluster management
				clusters := v1.Group("/clusters")
				{
					clusters.GET("", apiServer.HandleListClusters)
					clusters.GET("/:id", apiServer.HandleGetCluster)
					clusters.POST("", apiServer.HandleCreateCluster)
					clusters.PUT("/:id", apiServer.HandleUpdateCluster)
					clusters.DELETE("/:id", apiServer.HandleDeleteCluster)
					clusters.GET("/:id/health", apiServer.HandleClusterHealth)
				}

				// Event management
				events := v1.Group("/events")
				{
					events.GET("", apiServer.HandleListEvents)
					events.GET("/:id", apiServer.HandleGetEvent)
					events.POST("/search", apiServer.HandleSearchEvents)
				}

				// Command management
				commands := v1.Group("/commands")
				{
					commands.POST("", apiServer.HandleSendCommand)
					commands.GET("/:id", apiServer.HandleGetCommand)
					commands.GET("/:id/result", apiServer.HandleGetCommandResult)
					commands.GET("/:id/events", apiServer.HandleGetCommandEvents)
					commands.GET("", apiServer.HandleListPendingCommands)
				}

				// Operation tracking
				operations := v1.Group("/operations")
				{
					operations.POST("", apiServer.HandleRecordOperation)
					operations.GET("/:id/events", apiServer.HandleGetOperationEvents)
				}
			}

			h.logger.Info("All HTTP API routes registered")
			return nil
		},
	}

	h.standardInit = pkginitializers.NewHTTPServerInitializer(serverConfig, h.logger)
	return h.standardInit.Initialize(ctx)
}

// GetServer returns the HTTP server instance.
func (h *HTTPServerInitializer) GetServer() commonserver.Server {
	if h.standardInit == nil {
		return nil
	}
	return h.standardInit.GetServer()
}

// Close is a no-op (server lifecycle managed by bootstrap).
func (h *HTTPServerInitializer) Close(ctx context.Context) error {
	return nil
}

// GRPCServerInitializer initializes the gRPC server.
type GRPCServerInitializer struct {
	opts         *commonapp.StandardOptions
	logger       core.Logger
	infra        *InfrastructureInitializers
	services     *BusinessServicesInitializer
	standardInit *pkginitializers.GRPCServerInitializer
}

// NewGRPCServerInitializer creates a new gRPC server initializer.
func NewGRPCServerInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	infra *InfrastructureInitializers,
	services *BusinessServicesInitializer,
) *GRPCServerInitializer {
	return &GRPCServerInitializer{
		opts:     opts,
		logger:   logger,
		infra:    infra,
		services: services,
	}
}

// Name returns the initializer name.
func (g *GRPCServerInitializer) Name() string {
	return "agent-manager-grpc-server"
}

// Priority returns initialization priority.
func (g *GRPCServerInitializer) Priority() int {
	return 900 // Before HTTP server
}

// Initialize creates and starts gRPC server.
func (g *GRPCServerInitializer) Initialize(ctx context.Context) error {
	if !g.opts.GRPC.Enable {
		g.logger.Infow("gRPC server is disabled, skipping initialization")
		return nil
	}

	g.logger.Infow("Initializing gRPC server")

	// Get business services
	businessServices := g.services.Services()

	// Create gRPC service servers
	agentService := agentgrpc.NewAgentServiceServer(businessServices.Registry, g.logger)
	commandService := agentgrpc.NewCommandServiceServer(
		businessServices.Dispatcher,
		g.infra.Database.Store(),
		g.logger,
	)

	serverConfig := &pkginitializers.GRPCServerConfig{
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

	g.standardInit = pkginitializers.NewGRPCServerInitializer(serverConfig, g.logger)
	return g.standardInit.Initialize(ctx)
}

// GetServer returns the gRPC server instance.
func (g *GRPCServerInitializer) GetServer() commonserver.Server {
	if g.standardInit == nil {
		return nil
	}
	return g.standardInit.GetServer()
}

// Close is a no-op (server lifecycle managed by bootstrap).
func (g *GRPCServerInitializer) Close(ctx context.Context) error {
	return nil
}
