// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/auth/app/options"
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
	// 创建配置选项
	opts := options.NewServerOptions()

	// 使用组合框架运行应用
	commonapp.RunWithRunner(
		opts,
		&AuthApp{},
		initLogger,
		commonapp.CommandConfig{
			Use:       "auth",
			Short:     "Auth Service",
			Long:      "Auth Service provides user authentication and authorization services",
			EnvPrefix: "AUTH",
		},
	)
}

// AuthApp 实现 commonapp.Application 接口
type AuthApp struct {
	bootstrap *bootstrap.Bootstrap
	opts      *options.ServerOptions
	config    *auth.Config
	logger    core.Logger

	// 组件初始化器
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

// Initialize 初始化应用程序
func (a *AuthApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	a.opts = opts.(*options.ServerOptions)

	// 初始化日志系统
	logger, err := initLogger(opts)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Infow("Initializing Auth Service",
		"http_port", a.opts.Server.Port,
		"health_port", a.opts.Health.Port,
		"email_enabled", a.opts.Email.Enabled,
	)

	// 转换为业务配置
	config, err := a.opts.Config()
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}
	a.config = config

	// 创建 bootstrap 实例
	a.bootstrap = a.createBootstrap()

	// 注册所有组件初始化器（但不执行初始化）
	// 初始化将在 Run() 方法中由 bootstrap.Run() 执行
	a.registerComponents()

	a.logger.Infow("Components registered, ready to start")
	return nil
}

// Run 运行应用程序主逻辑
func (a *AuthApp) Run(ctx context.Context) error {
	a.logger.Infow("Auth Service started successfully",
		"http_address", fmt.Sprintf("%s:%d", a.opts.Server.Host, a.opts.Server.Port),
		"health_address", fmt.Sprintf(":%d", a.opts.GetHealthPort()),
	)

	// 使用 bootstrap 的 Run 方法,它会等待信号
	return a.bootstrap.Run(ctx, nil)
}

// Shutdown 优雅关闭应用程序
func (a *AuthApp) Shutdown(ctx context.Context) error {
	a.logger.Infow("Shutting down Auth Service")
	return a.bootstrap.Shutdown(ctx)
}

// createBootstrap 创建 bootstrap 实例
func (a *AuthApp) createBootstrap() *bootstrap.Bootstrap {
	return bootstrap.New(a.logger)
}

// registerComponents 注册所有组件初始化器
func (a *AuthApp) registerComponents() {
	// Convert to internal options format
	opts := a.convertToInternalOptions()

	// 1. Database (优先级 300)
	a.dbInit = initializers.NewDatabaseInitializer(opts, a.logger)
	a.bootstrap.Register(a.dbInit)

	// 2. Redis (优先级 400)
	a.redisInit = initializers.NewRedisInitializer(opts, a.logger)
	a.bootstrap.Register(a.redisInit)

	// 3. Session Service (优先级 450)
	a.sessionInit = initializers.NewSessionServiceInitializer(
		opts,
		a.logger,
		a.dbInit,
		a.redisInit,
	)
	a.bootstrap.Register(a.sessionInit)

	// 4. Email Client (优先级 450)
	a.emailInit = initializers.NewEmailClientInitializer(opts, a.logger)
	a.bootstrap.Register(a.emailInit)

	// 5. Audit Service (优先级 460)
	a.auditInit = initializers.NewAuditServiceInitializer(
		opts,
		a.logger,
		a.dbInit,
	)
	a.bootstrap.Register(a.auditInit)

	// 6. Notification Service (优先级 470)
	a.notificationInit = initializers.NewNotificationServiceInitializer(
		opts,
		a.logger,
		a.dbInit,
		a.emailInit,
	)
	a.bootstrap.Register(a.notificationInit)

	// 7. Forced Logout Service (优先级 490)
	a.forcedLogoutInit = initializers.NewForcedLogoutServiceInitializer(
		opts,
		a.logger,
		a.sessionInit,
		a.auditInit,
		a.notificationInit,
	)
	a.bootstrap.Register(a.forcedLogoutInit)

	// 8. HTTP Server (优先级 600)
	a.httpInit = initializers.NewHTTPServerInitializer(
		opts,
		a.logger,
		a.dbInit,
		a.redisInit,
		a.sessionInit,
		a.auditInit,
		a.notificationInit,
		a.forcedLogoutInit,
		a.emailInit,
	)
	a.bootstrap.Register(a.httpInit)

	// 9. Health Check Server (优先级最低，最后启动)
	healthPort := a.opts.GetHealthPort()
	healthAddr := fmt.Sprintf(":%d", healthPort)
	a.healthInit = pkginitializers.NewHealthCheckInitializer(healthAddr, a.logger)
	a.bootstrap.Register(a.healthInit)
}

// convertToInternalOptions converts cmd/auth options to internal/auth/config options
func (a *AuthApp) convertToInternalOptions() *authconfig.Config {
	return &authconfig.Config{
		Server:   a.opts.Server,
		Database: a.opts.Database,
		Redis:    a.opts.Redis,
		JWT:      a.opts.JWT,
		Logging:  a.opts.Logging,
		Email:    a.opts.Email,
	}
}

// initLogger 初始化日志系统
func initLogger(opts commonapp.Options) (core.Logger, error) {
	serverOpts := opts.(*options.ServerOptions)
	return serverOpts.InitLogger()
}
