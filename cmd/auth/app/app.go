// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/auth/startup"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

const (
	// UserAgent is the User-Agent string for the auth service.
	UserAgent = "aetherius-auth"
)

// Execute runs the auth service command.
func Execute() {
	// Create configuration options
	opts := commonapp.NewStandardOptions("Auth", UserAgent).
		WithDatabase().WithRedis().WithJWT().WithEmail()

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

// AuthApp implements commonapp.Application interface.
type AuthApp struct {
	bootstrap *bootstrap.Bootstrap
	opts      *commonapp.StandardOptions
	logger    core.Logger
}

// Name returns the application name.
func (a *AuthApp) Name() string {
	return "Auth Service"
}

// Initialize initializes the application.
func (a *AuthApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	a.opts = opts.(*commonapp.StandardOptions)

	// Initialize logger
	logger, err := a.opts.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	return nil
}

// Run runs the application.
func (a *AuthApp) Run(ctx context.Context) error {
	// Bootstrap.Run() is already called by RunWithBootstrap
	// We just need to wait for the context to be cancelled
	<-ctx.Done()
	return ctx.Err()
}

// Shutdown gracefully shuts down the application.
func (a *AuthApp) Shutdown(ctx context.Context) error {
	// Bootstrap shutdown is handled automatically
	return nil
}

// registerComponents registers all component initializers with bootstrap.
// This method defines the complete initialization order for the auth service.
func (a *AuthApp) registerComponents(bs *bootstrap.Bootstrap) error {
	a.bootstrap = bs

	// Layer 1: Infrastructure (Priority 300-500)
	// Database, Redis, and Email client configuration
	infra := startup.NewInfrastructureInitializers(a.opts, a.logger)
	bs.Register(infra.Database)
	bs.Register(infra.Redis)
	bs.Register(infra.Email)

	// Layer 2: Core Business Services (Priority 600)
	// Auth, User, Role, Permission, and APIKey services
	coreServices := startup.NewCoreServicesInitializer(
		a.opts,
		a.logger,
		infra.Database,
		infra.Redis,
	)
	bs.Register(coreServices)

	// Layer 3: Forced Logout Feature Services (Priority 650-680)
	// Session, Audit, Notification, and ForcedLogout services
	sessionInit := startup.NewSessionServiceInitializer(a.opts, a.logger, infra.Redis)
	bs.Register(sessionInit)

	auditInit := startup.NewAuditServiceInitializer(a.logger, infra.Database)
	bs.Register(auditInit)

	notificationInit := startup.NewNotificationServiceInitializer(
		a.opts,
		a.logger,
		infra.Database,
		infra.Email,
	)
	bs.Register(notificationInit)

	forcedLogoutInit := startup.NewForcedLogoutServiceInitializer(
		a.logger,
		sessionInit,
		auditInit,
		notificationInit,
	)
	bs.Register(forcedLogoutInit)

	// Layer 4: Server Layer (Priority 900-1000)
	// gRPC and HTTP servers that expose the services
	if a.opts.GRPC.Enable {
		grpcInit := startup.NewGRPCServerInitializer(
			a.opts,
			a.logger,
			coreServices,
			sessionInit,
		)
		bs.Register(grpcInit)
	}

	httpInit := startup.NewHTTPServerInitializer(
		a.opts,
		a.logger,
		infra,
		coreServices,
		sessionInit,
		auditInit,
		notificationInit,
		forcedLogoutInit,
	)
	bs.Register(httpInit)

	// Layer 5: Monitoring (Priority 2000)
	// Health check endpoint
	healthInit := pkginitializers.NewHealthCheckInitializer(a.opts.Health, a.logger)
	bs.Register(healthInit)

	return nil
}
