// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/cluster/app/options"
	clusterconfig "github.com/kart-io/k8s-agent/internal/cluster/config"
	"github.com/kart-io/k8s-agent/internal/cluster/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// Execute runs the cluster service command
func Execute() {
	// 创建配置选项
	opts := options.NewServerOptions()

	// 使用组合框架运行应用
	commonapp.RunWithRunner(
		opts,
		&ClusterApp{},
		initLogger,
		commonapp.CommandConfig{
			Use:       "cluster",
			Short:     "Cluster Service",
			Long:      "Cluster Service provides multi-cluster management and K8s resource API",
			EnvPrefix: "CLUSTER",
		},
	)
}

// ClusterApp 实现 commonapp.Application 接口
type ClusterApp struct {
	bootstrap *bootstrap.Bootstrap
	opts      *options.ServerOptions
	config    *clusterconfig.Config
	logger    core.Logger

	// 组件初始化器
	dbInit     *initializers.DatabaseInitializer
	httpInit   *initializers.HTTPServerInitializer
	healthInit *pkginitializers.HealthCheckInitializer
}

// Initialize 初始化应用程序
func (a *ClusterApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	a.opts = opts.(*options.ServerOptions)

	// 初始化日志系统
	logger, err := initLogger(opts)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Infow("Initializing Cluster Service",
		"http_port", a.opts.Server.Port,
		"health_port", a.opts.Health.Port,
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
func (a *ClusterApp) Run(ctx context.Context) error {
	a.logger.Infow("Cluster Service started successfully",
		"http_address", fmt.Sprintf("%s:%d", a.opts.Server.Host, a.opts.Server.Port),
		"health_address", fmt.Sprintf(":%d", a.opts.GetHealthPort()),
	)

	// 使用 bootstrap 的 Run 方法,它会等待信号
	// runFunc 会在所有初始化器完成后调用
	return a.bootstrap.Run(ctx, func() error {
		// 启动 HTTP 服务器（在 goroutine 中）
		go func() {
			if err := a.httpInit.Start(); err != nil {
				a.logger.Fatalw("HTTP server failed to start", "error", err)
			}
		}()

		a.logger.Infow("All services started, waiting for shutdown signal")
		return nil
	})
}

// Shutdown 优雅关闭应用程序
func (a *ClusterApp) Shutdown(ctx context.Context) error {
	a.logger.Infow("Shutting down Cluster Service")
	return a.bootstrap.Shutdown(ctx)
}

// createBootstrap 创建 bootstrap 实例
func (a *ClusterApp) createBootstrap() *bootstrap.Bootstrap {
	return bootstrap.New(a.logger)
}

// registerComponents 注册所有组件初始化器
func (a *ClusterApp) registerComponents() {
	// 1. Database (优先级 300)
	a.dbInit = initializers.NewDatabaseInitializer(a.opts.Database, a.logger)
	a.bootstrap.Register(a.dbInit)

	// 2. HTTP Server (优先级 500)
	a.httpInit = initializers.NewHTTPServerInitializer(
		a.opts.Server,
		a.opts.JWT,
		a.logger,
		a.dbInit,
	)
	a.bootstrap.Register(a.httpInit)

	// 3. Health Check (优先级 600)
	a.healthInit = pkginitializers.NewHealthCheckInitializer(
		fmt.Sprintf(":%d", a.opts.GetHealthPort()),
		a.logger,
	)
	a.bootstrap.Register(a.healthInit)
}

// initLogger 初始化日志系统
func initLogger(opts commonapp.Options) (core.Logger, error) {
	cfg := opts.(*options.ServerOptions)
	return cfg.InitLogger()
}
