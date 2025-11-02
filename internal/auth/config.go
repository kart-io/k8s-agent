// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package auth

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	commonoptions "github.com/kart-io/k8s-agent/common/options"
	authconfig "github.com/kart-io/k8s-agent/internal/auth/config"
	"github.com/kart-io/k8s-agent/internal/auth/initializers"
	"github.com/kart-io/k8s-agent/common/bootstrap"
	"github.com/kart-io/logger/core"
)

var (
	// Name is the name of the compiled software.
	Name = "aetherius-auth"

	// ID is the hostname of the machine running the auth service.
	ID, _ = os.Hostname()
)

// Config contains application-related configurations for the auth service.
type Config struct {
	Server   *commonoptions.ServerOptions
	Database *commonoptions.DatabaseOptions
	Redis    *commonoptions.RedisOptions
	JWT      *commonoptions.JWTOptions
	Logging  *commonoptions.LoggingOptions
	Email    *commonoptions.EmailOptions
	Metrics  *commonoptions.MetricsOptions
}

// Server represents the auth service server.
type Server struct {
	cfg       *Config
	logger    core.Logger
	bootstrap *bootstrap.Bootstrap

	// Component initializers
	dbInit           *initializers.DatabaseInitializer
	redisInit        *initializers.RedisInitializer
	sessionInit      *initializers.SessionServiceInitializer
	auditInit        *initializers.AuditServiceInitializer
	notificationInit *initializers.NotificationServiceInitializer
	forcedLogoutInit *initializers.ForcedLogoutServiceInitializer
	emailInit        *initializers.EmailClientInitializer
	httpInit         *initializers.HTTPServerInitializer
}

// NewServer initializes and returns a new Server instance.
func (c *Config) NewServer(ctx context.Context, logger core.Logger) (*Server, error) {
	s := &Server{
		cfg:       c,
		logger:    logger,
		bootstrap: bootstrap.New(logger),
	}

	// Register all components
	if err := s.registerComponents(); err != nil {
		return nil, fmt.Errorf("failed to register components: %w", err)
	}

	s.logger.Infow("Auth server initialized successfully",
		"id", ID,
		"name", Name,
	)

	return s, nil
}

// Run starts the server and listens for termination signals.
// It gracefully shuts down the server upon receiving a termination signal.
func (s *Server) Run(ctx context.Context) error {
	s.logger.Infow("Auth service started",
		"address", fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port),
	)

	// Setup signal context for graceful shutdown if not provided
	if ctx == nil {
		ctx = setupSignalContext()
	}

	// Use bootstrap's Run method to handle initialization and graceful shutdown
	return s.bootstrap.Run(ctx, nil)
}

// setupSignalContext creates a context that is cancelled when interrupt or termination signals are received.
func setupSignalContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
		<-c
		os.Exit(1) // Second signal. Exit directly.
	}()
	return ctx
}

// registerComponents registers all component initializers with the bootstrap system.
func (s *Server) registerComponents() error {
	opts := s.convertToInternalOptions()

	// 1. Database (priority 300)
	s.dbInit = initializers.NewDatabaseInitializer(opts, s.logger)
	s.bootstrap.Register(s.dbInit)

	// 2. Redis (priority 400)
	s.redisInit = initializers.NewRedisInitializer(opts, s.logger)
	s.bootstrap.Register(s.redisInit)

	// 3. Session Service (priority 450)
	s.sessionInit = initializers.NewSessionServiceInitializer(
		opts,
		s.logger,
		s.dbInit,
		s.redisInit,
	)
	s.bootstrap.Register(s.sessionInit)

	// 4. Email Client (priority 450)
	s.emailInit = initializers.NewEmailClientInitializer(opts, s.logger)
	s.bootstrap.Register(s.emailInit)

	// 5. Audit Service (priority 460)
	s.auditInit = initializers.NewAuditServiceInitializer(
		opts,
		s.logger,
		s.dbInit,
	)
	s.bootstrap.Register(s.auditInit)

	// 6. Notification Service (priority 470)
	s.notificationInit = initializers.NewNotificationServiceInitializer(
		opts,
		s.logger,
		s.dbInit,
		s.emailInit,
	)
	s.bootstrap.Register(s.notificationInit)

	// 7. Forced Logout Service (priority 490)
	s.forcedLogoutInit = initializers.NewForcedLogoutServiceInitializer(
		opts,
		s.logger,
		s.sessionInit,
		s.auditInit,
		s.notificationInit,
	)
	s.bootstrap.Register(s.forcedLogoutInit)

	// 8. HTTP Server (priority 600)
	s.httpInit = initializers.NewHTTPServerInitializer(
		opts,
		s.logger,
		s.dbInit,
		s.redisInit,
		s.sessionInit,
		s.auditInit,
		s.notificationInit,
		s.forcedLogoutInit,
		s.emailInit,
	)
	s.bootstrap.Register(s.httpInit)

	return nil
}

// convertToInternalOptions converts Config to the internal options format
// required by the initializers.
func (s *Server) convertToInternalOptions() *authconfig.Options {
	// Convert from the new auth.Config structure to the old config.Options structure
	// that initializers expect
	return &authconfig.Options{
		Server:   s.cfg.Server,
		Database: s.cfg.Database,
		Redis:    s.cfg.Redis,
		JWT:      s.cfg.JWT,
		Logging:  s.cfg.Logging,
		Email:    s.cfg.Email,
	}
}
