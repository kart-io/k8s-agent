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
	llmInit    *initializers.LLMInitializer
	grpcInit   *initializers.GRPCServerInitializer
	httpInit   *initializers.HTTPServerInitializer
	healthInit *pkginitializers.HealthCheckInitializer
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

	// 2. gRPC Server (优先级 450 - 在 LLM 之后，HTTP 之前)
	a.grpcInit = initializers.NewGRPCServerInitializer(
		opts,
		a.GetLogger(),
		a.llmInit,
	)
	bs.Register(a.grpcInit)

	// 3. HTTP Server with gRPC-Gateway (优先级 500 - 在 gRPC 之后)
	// HTTP requests will be automatically converted to gRPC calls using the same reasoning service!
	a.httpInit = initializers.NewHTTPServerInitializer(
		opts,
		a.GetLogger(),
		a.grpcInit, // Pass gRPC init to get shared service
	)
	bs.Register(a.httpInit)

	// 4. Health Check (优先级 600)
	a.healthInit = pkginitializers.NewHealthCheckInitializer(
		fmt.Sprintf(":%d", opts.GetHealthPort()),
		a.GetLogger(),
	)
	bs.Register(a.healthInit)

	return nil
}

// Run/Shutdown/initLogger 方法已由 StandardBootstrapApplication 提供，无需重复定义
