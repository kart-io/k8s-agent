package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/kart-io/k8s-agent/internal/agent-manager/agent"
	"github.com/kart-io/k8s-agent/internal/agent-manager/command"
	"github.com/kart-io/k8s-agent/internal/agent-manager/storage"
	agentv1 "github.com/kart-io/k8s-agent/pkg/api/agent/v1"
	"github.com/kart-io/logger/core"
)

// Server represents the gRPC server
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	logger     core.Logger

	// Services
	agentService   *AgentServiceServer
	commandService *CommandServiceServer

	// Configuration
	host string
	port int
}

// ServerOptions holds gRPC server configuration
type ServerOptions struct {
	Host string
	Port int

	// gRPC options
	MaxRecvMsgSize int
	MaxSendMsgSize int

	// Keepalive options
	KeepaliveTime    time.Duration
	KeepaliveTimeout time.Duration

	// Components
	Registry   *agent.Registry
	Dispatcher *command.Dispatcher
	Store      *storage.PostgresStore
}

// NewServer creates a new gRPC server
func NewServer(opts *ServerOptions, logger core.Logger) (*Server, error) {
	// Set defaults
	if opts.MaxRecvMsgSize == 0 {
		opts.MaxRecvMsgSize = 4 * 1024 * 1024 // 4MB
	}
	if opts.MaxSendMsgSize == 0 {
		opts.MaxSendMsgSize = 4 * 1024 * 1024 // 4MB
	}
	if opts.KeepaliveTime == 0 {
		opts.KeepaliveTime = 2 * time.Hour
	}
	if opts.KeepaliveTimeout == 0 {
		opts.KeepaliveTimeout = 20 * time.Second
	}

	// Create gRPC server with options
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(opts.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(opts.MaxSendMsgSize),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    opts.KeepaliveTime,
			Timeout: opts.KeepaliveTimeout,
		}),
	)

	// Create service instances
	agentService := NewAgentServiceServer(opts.Registry, logger)
	commandService := NewCommandServiceServer(opts.Dispatcher, opts.Store, logger)

	// Register services
	agentv1.RegisterAgentServiceServer(grpcServer, agentService)
	agentv1.RegisterCommandServiceServer(grpcServer, commandService)

	// Register reflection service (for development/testing)
	reflection.Register(grpcServer)

	// Create listener
	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	logger.Infow("gRPC server created",
		"address", addr,
	)

	return &Server{
		grpcServer:     grpcServer,
		listener:       listener,
		logger:         logger.With("component", "grpc-server"),
		agentService:   agentService,
		commandService: commandService,
		host:           opts.Host,
		port:           opts.Port,
	}, nil
}

// Start starts the gRPC server
func (s *Server) Start(ctx context.Context) error {
	s.logger.Infow("Starting gRPC server",
		"address", s.listener.Addr().String(),
	)

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := s.grpcServer.Serve(s.listener); err != nil {
			errCh <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		s.logger.Infow("Stopping gRPC server")
		return s.Stop()
	case err := <-errCh:
		return err
	}
}

// Stop stops the gRPC server gracefully
func (s *Server) Stop() error {
	s.logger.Infow("Stopping gRPC server")

	// Graceful stop with timeout
	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()

	// Wait for graceful stop or timeout
	select {
	case <-stopped:
		s.logger.Infow("gRPC server stopped gracefully")
	case <-time.After(30 * time.Second):
		s.logger.Warnw("gRPC server graceful stop timeout, forcing stop")
		s.grpcServer.Stop()
	}

	return nil
}

// Address returns the server address
func (s *Server) Address() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return fmt.Sprintf("%s:%d", s.host, s.port)
}
