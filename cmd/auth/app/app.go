// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/auth/app/options"
	"github.com/kart-io/k8s-agent/internal/auth/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// Execute runs the auth service command.
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

// AuthApp implements commonapp.Application interface.
type AuthApp struct {
	opts   *options.ServerOptions // 直接使用 ServerOptions，不转换
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

// Name returns the application name.
func (a *AuthApp) Name() string {
	return "Auth Service"
}

// Initialize initializes the application.
func (a *AuthApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	// 直接保存 ServerOptions，不需要转换
	a.opts = opts.(*options.ServerOptions)

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
	// The bootstrap framework handles running all servers
	// This method can be used for additional application logic if needed
	<-ctx.Done()
	return nil
}

// Shutdown gracefully shuts down the application.
func (a *AuthApp) Shutdown(ctx context.Context) error {
	// Bootstrap framework handles component shutdown
	// This method can be used for additional cleanup if needed
	return nil
}

// registerComponents registers all component initializers with bootstrap.
func (a *AuthApp) registerComponents(bs *bootstrap.Bootstrap) error {
	// 直接使用已有的 opts，不需要重新创建或转换

	// 1. Database (priority 300)
	a.dbInit = initializers.NewDatabaseInitializer(a.opts, a.logger)
	bs.Register(a.dbInit)

	// 2. Redis (priority 400)
	a.redisInit = initializers.NewRedisInitializer(a.opts, a.logger)
	bs.Register(a.redisInit)

	// 3. Session Service (priority 450)
	a.sessionInit = initializers.NewSessionServiceInitializer(
		a.opts,
		a.logger,
		a.dbInit,
		a.redisInit,
	)
	bs.Register(a.sessionInit)

	// 4. Email Client (priority 450)
	a.emailInit = initializers.NewEmailClientInitializer(a.opts, a.logger)
	bs.Register(a.emailInit)

	// 5. Audit Service (priority 460)
	a.auditInit = initializers.NewAuditServiceInitializer(
		a.opts,
		a.logger,
		a.dbInit,
	)
	bs.Register(a.auditInit)

	// 6. Notification Service (priority 470)
	a.notificationInit = initializers.NewNotificationServiceInitializer(
		a.opts,
		a.logger,
		a.dbInit,
		a.emailInit,
	)
	bs.Register(a.notificationInit)

	// 7. Forced Logout Service (priority 490)
	a.forcedLogoutInit = initializers.NewForcedLogoutServiceInitializer(
		a.opts,
		a.logger,
		a.sessionInit,
		a.auditInit,
		a.notificationInit,
	)
	bs.Register(a.forcedLogoutInit)

	// 8. HTTP Server (priority 600)
	a.httpInit = initializers.NewHTTPServerInitializer(
		a.opts,
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
	a.healthInit = pkginitializers.NewHealthCheckInitializer(a.opts.Health, a.logger)
	bs.Register(a.healthInit)

	return nil
}
