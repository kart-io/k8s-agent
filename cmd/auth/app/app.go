// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/auth/app/options"
	commonapp "github.com/kart-io/k8s-agent/common/app"
	"github.com/kart-io/k8s-agent/common/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/common/initializers"
	"github.com/kart-io/k8s-agent/internal/auth"
	authconfig "github.com/kart-io/k8s-agent/internal/auth/config"
	"github.com/kart-io/k8s-agent/internal/auth/initializers"
)

// Execute runs the auth service command
func Execute() {
	// 创建配置选项
	opts := options.NewServerOptions()

	// 创建应用实例
	app := &AuthApp{}

	// 使用组合框架运行应用
	commonapp.RunWithRunner(
		opts,
		app,
		commonapp.StandardInitLogger,
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
	*commonapp.StandardBootstrapApplication // 嵌入标准 Bootstrap 应用

	config *auth.Config

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
	// 转换配置
	serverOpts := opts.(*options.ServerOptions)
	config, err := serverOpts.Config()
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}
	a.config = config

	// 创建标准 Bootstrap 应用
	if a.StandardBootstrapApplication == nil {
		a.StandardBootstrapApplication = commonapp.NewStandardBootstrapApplication("Auth", a)
	}

	// 调用标准初始化
	return a.StandardBootstrapApplication.Initialize(ctx, opts)
}

// RegisterComponents 实现 ComponentRegistrar 接口，注册所有组件初始化器
func (a *AuthApp) RegisterComponents(bs *bootstrap.Bootstrap) error {
	// Convert to internal options format
	opts := a.convertToInternalOptions()

	// 1. Database (优先级 300)
	a.dbInit = initializers.NewDatabaseInitializer(opts, a.GetLogger())
	bs.Register(a.dbInit)

	// 2. Redis (优先级 400)
	a.redisInit = initializers.NewRedisInitializer(opts, a.GetLogger())
	bs.Register(a.redisInit)

	// 3. Session Service (优先级 450)
	a.sessionInit = initializers.NewSessionServiceInitializer(
		opts,
		a.GetLogger(),
		a.dbInit,
		a.redisInit,
	)
	bs.Register(a.sessionInit)

	// 4. Email Client (优先级 450)
	a.emailInit = initializers.NewEmailClientInitializer(opts, a.GetLogger())
	bs.Register(a.emailInit)

	// 5. Audit Service (优先级 460)
	a.auditInit = initializers.NewAuditServiceInitializer(
		opts,
		a.GetLogger(),
		a.dbInit,
	)
	bs.Register(a.auditInit)

	// 6. Notification Service (优先级 470)
	a.notificationInit = initializers.NewNotificationServiceInitializer(
		opts,
		a.GetLogger(),
		a.dbInit,
		a.emailInit,
	)
	bs.Register(a.notificationInit)

	// 7. Forced Logout Service (优先级 490)
	a.forcedLogoutInit = initializers.NewForcedLogoutServiceInitializer(
		opts,
		a.GetLogger(),
		a.sessionInit,
		a.auditInit,
		a.notificationInit,
	)
	bs.Register(a.forcedLogoutInit)

	// 8. HTTP Server (优先级 600)
	a.httpInit = initializers.NewHTTPServerInitializer(
		opts,
		a.GetLogger(),
		a.dbInit,
		a.redisInit,
		a.sessionInit,
		a.auditInit,
		a.notificationInit,
		a.forcedLogoutInit,
		a.emailInit,
	)
	bs.Register(a.httpInit)

	// 9. Health Check Server (优先级最低，最后启动)
	serverOpts := a.GetOptions().(*options.ServerOptions)
	a.healthInit = pkginitializers.NewHealthCheckInitializer(fmt.Sprintf(":%d", serverOpts.Server.Port), a.GetLogger())
	bs.Register(a.healthInit)

	return nil
}

// convertToInternalOptions converts cmd/auth options to internal/auth/config options
func (a *AuthApp) convertToInternalOptions() *authconfig.Config {
	serverOpts := a.GetOptions().(*options.ServerOptions)
	return &authconfig.Config{
		Server:   serverOpts.Server,
		Database: serverOpts.Database,
		Redis:    serverOpts.Redis,
		JWT:      serverOpts.JWT,
		Logging:  serverOpts.Logging,
		Email:    serverOpts.Email,
	}
}

// Run/Shutdown/initLogger 方法已由 StandardBootstrapApplication 提供，无需重复定义
