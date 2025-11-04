// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/auth/app/options"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/internal/auth"
	authconfig "github.com/kart-io/k8s-agent/internal/auth/config"
	"github.com/kart-io/k8s-agent/internal/auth/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// Execute runs the auth service command
func Execute() {
	// Create configuration options
	opts := options.NewServerOptions()

	// Create application instance
	app := &AuthApp{}

	// Use the simplified RunWithBootstrap to run the application
	commonapp.RunWithBootstrap(
		app,
		opts,
		commonapp.Config{
			Use:       "auth",
			Short:     "Auth Service",
			Long:      "Auth Service provides user authentication and authorization services",
			EnvPrefix: "AUTH",
		},
		app.registerComponents,
	)
}

// AuthApp implements commonapp.Application interface
type AuthApp struct {
	config *auth.Config
	logger core.Logger

	// Component initializers
	dbInit           *initializers.DatabaseInitializer
	redisInit        *initializers.RedisInitializer
	sessionInit      *initializers.SessionServiceInitializer
	auditInit        *initializers.AuditServiceInitializer
	notificationInit *initializers.NotificationServiceInitializer
	forcedLogoutInit *initializers.ForcedLogoutServiceInitializer
	emailInit        *initializers.EmailClientInitializer
	httpInit         *initializers.HTTPServerInitializer
	healthInit       *pkginitializers.HealthCheckInitializer
}

// Name returns the application name
func (a *AuthApp) Name() string {
	return "Auth Service"
}

// Initialize initializes the application
func (a *AuthApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	// Convert configuration
	serverOpts := opts.(*options.ServerOptions)
	config, err := serverOpts.Config()
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}
	a.config = config

	// Initialize logger
	logger, err := serverOpts.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	return nil
}

// Run runs the application
func (a *AuthApp) Run(ctx context.Context) error {
	// The bootstrap framework handles running all servers
	// This method can be used for additional application logic if needed
	<-ctx.Done()
	return nil
}

// Shutdown gracefully shuts down the application
func (a *AuthApp) Shutdown(ctx context.Context) error {
	// Bootstrap framework handles component shutdown
	// This method can be used for additional cleanup if needed
	return nil
}

// registerComponents registers all component initializers with bootstrap
func (a *AuthApp) registerComponents(bs *bootstrap.Bootstrap) error {
	// Get server options from bootstrap context
	// For now, we'll need to recreate the options since bootstrap doesn't provide them directly
	opts := options.NewServerOptions()
	opts.Complete()

	// Convert to internal options format
	internalOpts := a.convertToInternalOptions(opts)

	// 1. Database (priority 300)
	a.dbInit = initializers.NewDatabaseInitializer(internalOpts, a.logger)
	bs.Register(a.dbInit)

	// 2. Redis (priority 400)
	a.redisInit = initializers.NewRedisInitializer(internalOpts, a.logger)
	bs.Register(a.redisInit)

	// 3. Session Service (priority 450)
	a.sessionInit = initializers.NewSessionServiceInitializer(
		internalOpts,
		a.logger,
		a.dbInit,
		a.redisInit,
	)
	bs.Register(a.sessionInit)

	// 4. Email Client (priority 450)
	a.emailInit = initializers.NewEmailClientInitializer(internalOpts, a.logger)
	bs.Register(a.emailInit)

	// 5. Audit Service (priority 460)
	a.auditInit = initializers.NewAuditServiceInitializer(
		internalOpts,
		a.logger,
		a.dbInit,
	)
	bs.Register(a.auditInit)

	// 6. Notification Service (priority 470)
	a.notificationInit = initializers.NewNotificationServiceInitializer(
		internalOpts,
		a.logger,
		a.dbInit,
		a.emailInit,
	)
	bs.Register(a.notificationInit)

	// 7. Forced Logout Service (priority 490)
	a.forcedLogoutInit = initializers.NewForcedLogoutServiceInitializer(
		internalOpts,
		a.logger,
		a.sessionInit,
		a.auditInit,
		a.notificationInit,
	)
	bs.Register(a.forcedLogoutInit)

	// 8. HTTP Server (priority 600)
	a.httpInit = initializers.NewHTTPServerInitializer(
		internalOpts,
		a.logger,
		a.dbInit,
		a.redisInit,
		a.sessionInit,
		a.auditInit,
		a.notificationInit,
		a.forcedLogoutInit,
		a.emailInit,
	)
	bs.Register(a.httpInit)

	// 9. Health Check Server (priority 2000)
	healthOpts := commonoptions.NewHealthOptions()
	a.healthInit = pkginitializers.NewHealthCheckInitializer(healthOpts, a.logger)
	bs.Register(a.healthInit)

	return nil
}

// convertToInternalOptions converts cmd/auth options to internal/auth/config options
func (a *AuthApp) convertToInternalOptions(serverOpts *options.ServerOptions) *authconfig.Config {
	return &authconfig.Config{
		Server:   serverOpts.Server,
		Database: serverOpts.Database,
		Redis:    serverOpts.Redis,
		JWT:      serverOpts.JWT,
		Logging:  serverOpts.Logging,
		Email:    serverOpts.Email,
	}
}
