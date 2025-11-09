// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package startup

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	orchestratorv1 "github.com/kart-io/k8s-agent/pkg/api/orchestrator/v1"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// GRPCServerInitializer initializes gRPC server.
type GRPCServerInitializer struct {
	opts         *commonapp.StandardOptions
	logger       core.Logger
	coreServices *CoreServicesInitializer

	standardInit *pkginitializers.GRPCServerInitializer
}

// NewGRPCServerInitializer creates a new gRPC server initializer.
func NewGRPCServerInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	coreServices *CoreServicesInitializer,
) *GRPCServerInitializer {
	return &GRPCServerInitializer{
		opts:         opts,
		logger:       logger,
		coreServices: coreServices,
	}
}

// Name returns the initializer name.
func (g *GRPCServerInitializer) Name() string {
	return "grpc-server"
}

// Priority returns initialization priority.
func (g *GRPCServerInitializer) Priority() int {
	return 900 // After core services (600)
}

// Initialize starts the gRPC server.
func (g *GRPCServerInitializer) Initialize(ctx context.Context) error {
	if !g.opts.GRPC.Enable {
		g.logger.Infow("gRPC server is disabled, skipping initialization")
		return nil
	}

	g.logger.Infow("Initializing gRPC server",
		"host", g.opts.GRPC.Host,
		"port", g.opts.GRPC.Port,
	)

	// Get workflow service
	services := g.coreServices.Services()
	if services == nil {
		return fmt.Errorf("core services not initialized")
	}
	if services.WorkflowService == nil {
		return fmt.Errorf("workflow service not initialized")
	}

	// Configure gRPC server
	serverConfig := &pkginitializers.GRPCServerConfig{
		Name:     g.Name(),
		Priority: g.Priority(),
		Config:   g.opts.GRPC,
		ServiceRegister: func(grpcServer *grpc.Server) error {
			orchestratorv1.RegisterWorkflowServiceServer(grpcServer, services.WorkflowService)
			g.logger.Infow("Registered WorkflowService to gRPC server")
			return nil
		},
	}

	g.standardInit = pkginitializers.NewGRPCServerInitializer(serverConfig, g.logger)
	return g.standardInit.Initialize(ctx)
}

// Close stops the gRPC server.
func (g *GRPCServerInitializer) Close(ctx context.Context) error {
	if g.standardInit == nil {
		return nil
	}
	return g.standardInit.Close(ctx)
}

// HTTPServerInitializer initializes HTTP server with gRPC-Gateway.
type HTTPServerInitializer struct {
	opts         *commonapp.StandardOptions
	logger       core.Logger
	coreServices *CoreServicesInitializer

	standardInit *pkginitializers.HTTPServerInitializer
	mux          *runtime.ServeMux
}

// NewHTTPServerInitializer creates a new HTTP server initializer.
func NewHTTPServerInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	coreServices *CoreServicesInitializer,
) *HTTPServerInitializer {
	return &HTTPServerInitializer{
		opts:         opts,
		logger:       logger,
		coreServices: coreServices,
	}
}

// Name returns the initializer name.
func (h *HTTPServerInitializer) Name() string {
	return "http-server"
}

// Priority returns initialization priority.
func (h *HTTPServerInitializer) Priority() int {
	return 1000 // After gRPC server (900)
}

// Initialize starts the HTTP server with gRPC-Gateway.
func (h *HTTPServerInitializer) Initialize(ctx context.Context) error {
	h.logger.Infow("Initializing HTTP server with gRPC-Gateway",
		"host", h.opts.Server.Host,
		"port", h.opts.Server.Port,
	)

	// Get workflow service
	services := h.coreServices.Services()
	if services == nil {
		return fmt.Errorf("core services not initialized")
	}
	if services.WorkflowService == nil {
		return fmt.Errorf("workflow service not initialized")
	}

	// Create gRPC-Gateway mux with custom options
	h.mux = runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions:   runtime.JSONPb{}.MarshalOptions,
			UnmarshalOptions: runtime.JSONPb{}.UnmarshalOptions,
		}),
	)

	// Register gRPC service handler for HTTP (gRPC-Gateway magic!)
	// This allows HTTP requests to be automatically converted to gRPC calls
	if err := orchestratorv1.RegisterWorkflowServiceHandlerServer(
		ctx,
		h.mux,
		services.WorkflowService,
	); err != nil {
		return fmt.Errorf("failed to register workflow service handler: %w", err)
	}

	// Configure HTTP server
	h.standardInit = pkginitializers.NewHTTPServerInitializer(
		&pkginitializers.HTTPServerConfig{
			Name:     h.Name(),
			Priority: h.Priority(),
			Config:   h.opts.Server,
			RouteSetup: func(engine *gin.Engine) error {
				// Mount gRPC-Gateway mux to Gin routes
				// All /api/* requests are forwarded to gRPC service via gRPC-Gateway
				engine.Any("/api/*path", gin.WrapH(h.mux))

				h.logger.Infow("Registered gRPC-Gateway routes",
					"pattern", "/api/*",
					"note", "HTTP requests will be automatically converted to gRPC calls",
				)
				return nil
			},
		},
		h.logger,
	)

	return h.standardInit.Initialize(ctx)
}

// Close stops the HTTP server.
func (h *HTTPServerInitializer) Close(ctx context.Context) error {
	if h.standardInit == nil {
		return nil
	}
	return h.standardInit.Close(ctx)
}
