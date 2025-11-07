package initializers

import (
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	
	commonserver "github.com/kart-io/k8s-agent/common/server"
	orchestratorv1 "github.com/kart-io/k8s-agent/pkg/api/orchestrator/v1"
	commoninitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// HTTPServerInitializer initializes the HTTP server with gRPC-Gateway
// 使用 common/initializers 的标准 HTTP 服务器初始化器.
type HTTPServerInitializer struct {
	opts     *commonapp.StandardOptions
	logger   core.Logger
	grpcInit *GRPCServerInitializer

	// 使用标准初始化器
	standardInit *commoninitializers.HTTPServerInitializer
	mux          *runtime.ServeMux
}

// NewHTTPServerInitializer creates a new HTTP server initializer.
func NewHTTPServerInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	grpcInit *GRPCServerInitializer,
) *HTTPServerInitializer {
	return &HTTPServerInitializer{
		opts:     opts,
		logger:   logger.With("component", "http-server-initializer"),
		grpcInit: grpcInit,
	}
}

// Name returns the initializer name.
func (i *HTTPServerInitializer) Name() string {
	return "HTTPServer"
}

// Priority returns the initialization priority.
func (i *HTTPServerInitializer) Priority() int {
	return 800 // After gRPC server (700)
}

// Initialize sets up and starts the HTTP server with gRPC-Gateway.
func (i *HTTPServerInitializer) Initialize(ctx context.Context) error {
	i.logger.Infow("Initializing HTTP server with gRPC-Gateway",
		"host", i.opts.Server.Host,
		"port", i.opts.Server.Port,
	)

	// Create gRPC-Gateway mux with custom options
	i.mux = runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions:   runtime.JSONPb{}.MarshalOptions,
			UnmarshalOptions: runtime.JSONPb{}.UnmarshalOptions,
		}),
	)

	// Get shared workflow service from gRPC initializer
	workflowService := i.grpcInit.GetWorkflowService()
	if workflowService == nil {
		return fmt.Errorf("workflow service not initialized")
	}

	// Register gRPC service handler for HTTP (gRPC-Gateway magic!)
	// This allows HTTP requests to be automatically converted to gRPC calls
	if err := orchestratorv1.RegisterWorkflowServiceHandlerServer(
		ctx,
		i.mux,
		workflowService,
	); err != nil {
		return fmt.Errorf("failed to register workflow service handler: %w", err)
	}

	// 使用标准 HTTP 服务器初始化器
	i.standardInit = commoninitializers.NewHTTPServerInitializer(
		&commoninitializers.HTTPServerConfig{
			Name:     "OrchestratorHTTP",
			Priority: i.Priority(),
			Config:   i.opts.Server,
			RouteSetup: func(engine *gin.Engine) error {
				// 将 gRPC-Gateway 的 mux 挂载到 Gin 的路由上
				// 所有 /api/* 的请求都通过 gRPC-Gateway 转发到 gRPC 服务
				engine.Any("/api/*path", gin.WrapH(i.mux))

				i.logger.Infow("Registered gRPC-Gateway routes",
					"pattern", "/api/*",
					"note", "HTTP requests will be automatically converted to gRPC calls",
				)
				return nil
			},
		},
		i.logger,
	)

	// 调用标准初始化器的 Initialize
	return i.standardInit.Initialize(ctx)
}

// Close stops the HTTP server.
func (i *HTTPServerInitializer) Close(ctx context.Context) error {
	if i.standardInit == nil {
		return nil
	}

	return i.standardInit.Close(ctx)
}

// GetServer returns the server instance (implements ServerProvider).
func (i *HTTPServerInitializer) GetServer() commonserver.Server {
	if i.standardInit == nil {
		return nil
	}
	return i.standardInit.GetServer()
}
