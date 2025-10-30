// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/reasoning/app/options"
	reasoningconfig "github.com/kart-io/k8s-agent/internal/reasoning/config"
	"github.com/kart-io/k8s-agent/internal/reasoning/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// Execute runs the reasoning service command
func Execute() {
	// 创建配置选项
	opts := options.NewServerOptions()

	// 使用组合框架运行应用
	commonapp.RunWithRunner(
		opts,
		&ReasoningApp{},
		initLogger,
		commonapp.CommandConfig{
			Use:       "reasoning",
			Short:     "Reasoning Service",
			Long:      "Reasoning Service provides AI-driven root cause analysis and intelligent recommendations",
			EnvPrefix: "REASONING",
		},
	)
}

// ReasoningApp 实现 commonapp.Application 接口
type ReasoningApp struct {
	bootstrap *bootstrap.Bootstrap
	opts      *options.ServerOptions
	config    *reasoningconfig.Config
	logger    core.Logger

	// 组件初始化器
	llmInit    *initializers.LLMInitializer
	httpInit   *initializers.HTTPServerInitializer
	healthInit *pkginitializers.HealthCheckInitializer
}

// Initialize 初始化应用程序
func (a *ReasoningApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	a.opts = opts.(*options.ServerOptions)

	// 初始化日志系统
	logger, err := initLogger(opts)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Info("Aetherius Reasoning Service (Go)")
	a.logger.Info("=================================")
	a.logger.Infow("Initializing Reasoning Service",
		"http_port", a.opts.Server.Port,
		"health_port", a.opts.Health.Port,
		"llm_enabled", a.opts.LLM.Enabled,
		"memory_enabled", a.opts.Memory.EnableVectorStore,
	)

	// 转换为业务配置
	config := a.opts.Config()
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
func (a *ReasoningApp) Run(ctx context.Context) error {
	a.logger.Infow("Reasoning Service started successfully",
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
func (a *ReasoningApp) Shutdown(ctx context.Context) error {
	a.logger.Infow("Shutting down Reasoning Service")
	return a.bootstrap.Shutdown(ctx)
}

// createBootstrap 创建 bootstrap 实例
func (a *ReasoningApp) createBootstrap() *bootstrap.Bootstrap {
	return bootstrap.New(a.logger)
}

// registerComponents 注册所有组件初始化器
func (a *ReasoningApp) registerComponents() {
	// 1. LLM Clients (优先级 400)
	a.llmInit = initializers.NewLLMInitializer(a.opts.LLM, a.logger)
	a.bootstrap.Register(a.llmInit)

	// 2. HTTP Server (优先级 500)
	a.httpInit = initializers.NewHTTPServerInitializer(
		a.config,
		a.logger,
		a.llmInit,
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
