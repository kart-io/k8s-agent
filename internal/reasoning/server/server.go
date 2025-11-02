// Package server provides unified gRPC and HTTP servers using Kratos framework.
// This package follows the OneX architecture pattern where both servers share the same handler.
package server

import (
	"context"
	"fmt"

	kratosgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/grpc"

	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/internal/reasoning/handler"
	reasoningv1 "github.com/kart-io/k8s-agent/pkg/api/reasoning/v1"
	"github.com/kart-io/logger/core"
)

const (
	// Error channel buffer size for concurrent server errors.
	errorChannelBuffer = 2
)

// Config holds configuration for the unified server.
type Config struct {
	// HTTP server options (from common/options)
	HTTP *options.ServerOptions

	// gRPC server options (from common/options)
	GRPC *options.GRPCOptions

	// Handler (unified for both gRPC and HTTP)
	Handler *handler.ReasoningHandler
}

// Server manages both gRPC and HTTP servers using Kratos framework.
type Server struct {
	httpServer *kratoshttp.Server
	grpcServer *kratosgrpc.Server
	logger     core.Logger
}

// NewServer creates a new server with both gRPC and HTTP transports.
// Following the OneX architecture pattern, both servers use the same handler.
func NewServer(cfg *Config, logger core.Logger) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Complete and validate HTTP options
	if cfg.HTTP == nil {
		cfg.HTTP = options.NewServerOptions()
	}
	if err := cfg.HTTP.Complete(); err != nil {
		return nil, fmt.Errorf("failed to complete HTTP options: %w", err)
	}
	if err := cfg.HTTP.Validate(); err != nil {
		return nil, fmt.Errorf("invalid HTTP options: %w", err)
	}

	// Complete and validate gRPC options
	if cfg.GRPC == nil {
		cfg.GRPC = options.NewGRPCOptions()
	}
	if err := cfg.GRPC.Complete(); err != nil {
		return nil, fmt.Errorf("failed to complete gRPC options: %w", err)
	}
	if err := cfg.GRPC.Validate(); err != nil {
		return nil, fmt.Errorf("invalid gRPC options: %w", err)
	}

	// Create gRPC server
	grpcAddr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
	grpcServer := kratosgrpc.NewServer(
		kratosgrpc.Address(grpcAddr),
		kratosgrpc.Timeout(cfg.GRPC.ConnectionTimeout),
		kratosgrpc.Options(
			grpc.MaxRecvMsgSize(cfg.GRPC.MaxRecvMsgSize),
			grpc.MaxSendMsgSize(cfg.GRPC.MaxSendMsgSize),
		),
	)

	// Register gRPC service with the unified handler
	reasoningv1.RegisterReasoningServiceServer(grpcServer, cfg.Handler)

	logger.Infow("gRPC server configured",
		"address", grpcAddr,
		"max_recv_msg_size", cfg.GRPC.MaxRecvMsgSize,
		"max_send_msg_size", cfg.GRPC.MaxSendMsgSize,
	)

	// Create HTTP server
	httpAddr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	httpServer := kratoshttp.NewServer(
		kratoshttp.Address(httpAddr),
		kratoshttp.Timeout(cfg.HTTP.ReadTimeout),
	)

	// Register HTTP service with the unified handler
	reasoningv1.RegisterReasoningServiceHTTPServer(httpServer, cfg.Handler)

	logger.Infow("HTTP server configured",
		"address", httpAddr,
		"read_timeout", cfg.HTTP.ReadTimeout,
		"write_timeout", cfg.HTTP.WriteTimeout,
	)

	return &Server{
		httpServer: httpServer,
		grpcServer: grpcServer,
		logger:     logger.With("component", "unified-server"),
	}, nil
}

// Start starts both gRPC and HTTP servers.
func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, errorChannelBuffer)

	// Start gRPC server
	go func() {
		s.logger.Infow("Starting gRPC server")
		if err := s.grpcServer.Start(ctx); err != nil {
			s.logger.Errorw("gRPC server error", "error", err)
			errCh <- fmt.Errorf("gRPC server failed: %w", err)
		}
	}()

	// Start HTTP server
	go func() {
		s.logger.Infow("Starting HTTP server")
		if err := s.httpServer.Start(ctx); err != nil {
			s.logger.Errorw("HTTP server error", "error", err)
			errCh <- fmt.Errorf("HTTP server failed: %w", err)
		}
	}()

	// Wait for first error
	select {
	case err := <-errCh:
		// If one server fails, stop the other
		_ = s.Stop()
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stop stops both servers gracefully.
func (s *Server) Stop() error {
	s.logger.Infow("Stopping servers")

	// Create a timeout context for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), options.NewServerOptions().GracefulStop)
	defer cancel()

	// Stop HTTP server
	if err := s.httpServer.Stop(ctx); err != nil {
		s.logger.Errorw("Failed to stop HTTP server", "error", err)
		return fmt.Errorf("failed to stop HTTP server: %w", err)
	}
	s.logger.Infow("HTTP server stopped")

	// Stop gRPC server
	if err := s.grpcServer.Stop(ctx); err != nil {
		s.logger.Errorw("Failed to stop gRPC server", "error", err)
		return fmt.Errorf("failed to stop gRPC server: %w", err)
	}
	s.logger.Infow("gRPC server stopped")

	s.logger.Infow("All servers stopped successfully")
	return nil
}

// GRPCServer returns the underlying gRPC server for advanced configuration.
func (s *Server) GRPCServer() *kratosgrpc.Server {
	return s.grpcServer
}

// HTTPServer returns the underlying HTTP server for advanced configuration.
func (s *Server) HTTPServer() *kratoshttp.Server {
	return s.httpServer
}
