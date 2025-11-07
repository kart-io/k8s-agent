package initializers

import (
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	
	reasoningv1 "github.com/kart-io/k8s-agent/pkg/api/reasoning/v1"
	"github.com/kart-io/logger/core"
)

// HTTPServerInitializer initializes the HTTP server with gRPC-Gateway.
type HTTPServerInitializer struct {
	opts     *commonapp.StandardOptions
	logger   core.Logger
	grpcInit *GRPCServerInitializer

	server *http.Server
	mux    *runtime.ServeMux
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
	return 500 // After gRPC server (450)
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

	// Get shared reasoning service from gRPC initializer
	reasoningService := i.grpcInit.GetReasoningService()
	if reasoningService == nil {
		return fmt.Errorf("reasoning service not initialized")
	}

	// Register gRPC service handler for HTTP (gRPC-Gateway magic!)
	// This allows HTTP requests to be automatically converted to gRPC calls
	if err := reasoningv1.RegisterReasoningServiceHandlerServer(
		ctx,
		i.mux,
		reasoningService,
	); err != nil {
		return fmt.Errorf("failed to register reasoning service handler: %w", err)
	}

	// Create HTTP server
	i.server = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", i.opts.Server.Host, i.opts.Server.Port),
		Handler:           i.mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start HTTP server in background
	go func() {
		i.logger.Infow("Starting HTTP server with gRPC-Gateway",
			"address", i.server.Addr,
		)
		if err := i.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			i.logger.Errorw("HTTP server error", "error", err)
		}
	}()

	i.logger.Infow("HTTP server initialized successfully",
		"address", i.server.Addr,
		"note", "HTTP requests will be automatically converted to gRPC calls",
	)

	return nil
}

// Shutdown stops the HTTP server.
func (i *HTTPServerInitializer) Shutdown(ctx context.Context) error {
	if i.server == nil {
		return nil
	}

	i.logger.Infow("Shutting down HTTP server")

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := i.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	i.logger.Infow("HTTP server shutdown complete")
	return nil
}
