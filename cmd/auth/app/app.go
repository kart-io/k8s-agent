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
	// Use Wire to automatically inject all dependencies
	components, err := InitializeAuthComponents(a.opts)
	if err != nil {
		return fmt.Errorf("failed to initialize components: %w", err)
	}

	// Register components to Bootstrap
	bs.Register(components.DB)
	bs.Register(components.Redis)
	bs.Register(components.Session)
	bs.Register(components.Email)
	bs.Register(components.Audit)
	bs.Register(components.Notification)
	bs.Register(components.ForcedLogout)
	bs.Register(components.GRPC) // gRPC 服务器
	bs.Register(components.HTTP)
	bs.Register(components.Health)

	// Save references for app
	a.dbInit = components.DB
	a.redisInit = components.Redis
	a.sessionInit = components.Session
	a.emailInit = components.Email
	a.auditInit = components.Audit
	a.notificationInit = components.Notification
	a.forcedLogoutInit = components.ForcedLogout
	a.httpInit = components.HTTP
	a.healthInit = components.Health

	return nil
}
