// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"

	"github.com/kart-io/k8s-agent/cmd/reasoning/app/options"
	commonapp "github.com/kart-io/k8s-agent/common/app"
	"github.com/kart-io/k8s-agent/common/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/common/initializers"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	reasoningconfig "github.com/kart-io/k8s-agent/internal/reasoning/config"
	"github.com/kart-io/k8s-agent/internal/reasoning/initializers"
)

// Execute runs the reasoning service command
func Execute() {
	// 创建配置选项
	opts := options.NewServerOptions()

	// 创建应用实例
	app := &ReasoningApp{}

	// 使用组合框架运行应用
	commonapp.RunWithRunner(
		opts,
		app,
		commonapp.StandardInitLogger,
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
	*commonapp.StandardBootstrapApplication // 嵌入标准 Bootstrap 应用

	config *reasoningconfig.Config

	// 组件初始化器
	llmInit           *initializers.LLMInitializer
	unifiedServerInit *initializers.UnifiedServerInitializer
	healthInit        *pkginitializers.HealthCheckInitializer
}

// Initialize 初始化应用程序
func (a *ReasoningApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	// 转换配置
	serverOpts := opts.(*options.ServerOptions)
	config := serverOpts.Config()
	a.config = config

	// 创建标准 Bootstrap 应用
	if a.StandardBootstrapApplication == nil {
		a.StandardBootstrapApplication = commonapp.NewStandardBootstrapApplication("Reasoning", a)
	}

	// 调用标准初始化
	return a.StandardBootstrapApplication.Initialize(ctx, opts)
}

// RegisterComponents 实现 ComponentRegistrar 接口，注册所有组件初始化器
func (a *ReasoningApp) RegisterComponents(bs *bootstrap.Bootstrap) error {
	opts := a.GetOptions().(*options.ServerOptions)

	// 1. LLM Clients (优先级 400)
	a.llmInit = initializers.NewLLMInitializer(opts.LLM, a.GetLogger())
	bs.Register(a.llmInit)

	// 2. Unified Server (gRPC + HTTP using Kratos framework, OneX architecture pattern)
	// Priority 450 - after LLM initialization
	// A single handler implements both ReasoningServiceServer (gRPC) and ReasoningServiceHTTPServer (HTTP)
	// This follows the OneX pattern where both protocols share the same handler methods
	a.unifiedServerInit = initializers.NewUnifiedServerInitializer(
		opts,
		a.GetLogger(),
		a.llmInit,
	)
	bs.Register(a.unifiedServerInit)

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
