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
	*commonapp.BaseBootstrapApp // 组合基础 Bootstrap 应用

	opts   *options.ServerOptions
	config *clusterconfig.Config

	// 组件初始化器
	dbInit     *initializers.DatabaseInitializer
	httpInit   *initializers.HTTPServerInitializer
	healthInit *pkginitializers.HealthCheckInitializer
}

// Initialize 初始化应用程序
func (a *ClusterApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	a.opts = opts.(*options.ServerOptions)

	// 转换为业务配置
	config, err := a.opts.Config()
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}
	a.config = config

	// 创建 BaseBootstrapApp 实例并设置启动钩子
	a.BaseBootstrapApp = commonapp.NewBaseBootstrapApp("Cluster", a).
		WithStartupHook(a)

	// 使用基础类的初始化方法
	return a.BaseInitialize(ctx, a.opts)
}

// Run 运行应用程序主逻辑
func (a *ClusterApp) Run(ctx context.Context) error {
	// 使用基础类的运行方法
	return a.BaseRun(ctx, a.opts)
}

// Shutdown 优雅关闭应用程序
func (a *ClusterApp) Shutdown(ctx context.Context) error {
	// 使用基础类的关闭方法
	return a.BaseShutdown(ctx)
}

// OnStartup 实现 StartupHook 接口，在 bootstrap.Run() 中执行
func (a *ClusterApp) OnStartup(ctx context.Context) error {
	// 启动 HTTP 服务器（在 goroutine 中）
	go func() {
		if err := a.httpInit.Start(); err != nil {
			a.GetLogger().Fatalw("HTTP server failed to start", "error", err)
		}
	}()

	a.GetLogger().Infow("All services started, waiting for shutdown signal")
	return nil
}

// RegisterComponents 实现 ComponentRegistrar 接口，注册所有组件初始化器
func (a *ClusterApp) RegisterComponents(bs *bootstrap.Bootstrap) error {
	// 1. Database (优先级 300)
	a.dbInit = initializers.NewDatabaseInitializer(a.opts.Database, a.GetLogger())
	bs.Register(a.dbInit)

	// 2. HTTP Server (优先级 500)
	a.httpInit = initializers.NewHTTPServerInitializer(
		a.opts.Server,
		a.opts.JWT,
		a.GetLogger(),
		a.dbInit,
	)
	bs.Register(a.httpInit)

	// 3. Health Check (优先级 600)
	a.healthInit = pkginitializers.NewHealthCheckInitializer(
		fmt.Sprintf(":%d", a.opts.GetHealthPort()),
		a.GetLogger(),
	)
	bs.Register(a.healthInit)

	return nil
}

// initLogger 初始化日志系统
func initLogger(opts commonapp.Options) (core.Logger, error) {
	return commonapp.StandardInitLogger(opts)
}
