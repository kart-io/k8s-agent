// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/cluster/app/options"
	commonapp "github.com/kart-io/k8s-agent/common/app"
	"github.com/kart-io/k8s-agent/common/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/common/initializers"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	clusterconfig "github.com/kart-io/k8s-agent/internal/cluster/config"
	"github.com/kart-io/k8s-agent/internal/cluster/initializers"
)

// Execute runs the cluster service command
func Execute() {
	// 创建配置选项
	opts := options.NewServerOptions()

	// 创建应用实例
	app := &ClusterApp{}

	// 使用组合框架运行应用
	commonapp.RunWithRunner(
		opts,
		app,
		commonapp.StandardInitLogger,
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
	*commonapp.StandardBootstrapApplication // 嵌入标准 Bootstrap 应用

	config *clusterconfig.Config

	// 组件初始化器
	dbInit     *initializers.DatabaseInitializer
	httpInit   *initializers.HTTPServerInitializer
	healthInit *pkginitializers.HealthCheckInitializer
}

// Initialize 初始化应用程序
func (a *ClusterApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	// 转换配置
	serverOpts := opts.(*options.ServerOptions)
	config, err := serverOpts.Config()
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}
	a.config = config

	// 创建标准 Bootstrap 应用并设置启动钩子
	if a.StandardBootstrapApplication == nil {
		a.StandardBootstrapApplication = commonapp.NewStandardBootstrapApplication("Cluster", a).
			WithStartupHookFunc(a)
	}

	// 调用标准初始化
	return a.StandardBootstrapApplication.Initialize(ctx, opts)
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
	opts := a.GetOptions().(*options.ServerOptions)

	// 1. Database (优先级 300)
	a.dbInit = initializers.NewDatabaseInitializer(opts.Database, a.GetLogger())
	bs.Register(a.dbInit)

	// 2. HTTP Server (优先级 500)
	a.httpInit = initializers.NewHTTPServerInitializer(
		opts.Server,
		opts.JWT,
		a.GetLogger(),
		a.dbInit,
	)
	bs.Register(a.httpInit)

	// 3. Health Check (优先级 600)
	healthOpts := commonoptions.NewHealthOptions()
	healthOpts.Port = opts.GetHealthPort()
	a.healthInit = pkginitializers.NewHealthCheckInitializer(
		healthOpts,
		a.GetLogger(),
	)
	bs.Register(a.healthInit)

	return nil
}

// Run/Shutdown/initLogger 方法已由 StandardBootstrapApplication 提供，无需重复定义
