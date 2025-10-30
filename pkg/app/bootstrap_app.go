// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
)

// BootstrapConfig 定义 Bootstrap 应用需要的配置接口
type BootstrapConfig interface {
	// InitLogger 初始化日志系统
	InitLogger() (core.Logger, error)

	// GetServiceName 获取服务名称
	GetServiceName() string

	// GetLogFields 获取初始化日志字段
	GetLogFields() []interface{}
}

// ComponentRegistrar 定义组件注册接口
type ComponentRegistrar interface {
	// RegisterComponents 注册所有组件初始化器
	RegisterComponents(bootstrap *bootstrap.Bootstrap) error
}

// StartupHook 定义启动钩子接口（可选）
type StartupHook interface {
	// OnStartup 在 bootstrap.Run() 中执行的启动逻辑
	OnStartup(ctx context.Context) error
}

// BaseBootstrapApp 提供 Bootstrap 应用的通用功能
// 各服务可以组合此结构体来减少重复代码
type BaseBootstrapApp struct {
	bootstrap   *bootstrap.Bootstrap
	logger      core.Logger
	serviceName string
	registrar   ComponentRegistrar
	startupHook StartupHook
}

// NewBaseBootstrapApp 创建基础 Bootstrap 应用
func NewBaseBootstrapApp(serviceName string, registrar ComponentRegistrar) *BaseBootstrapApp {
	return &BaseBootstrapApp{
		serviceName: serviceName,
		registrar:   registrar,
	}
}

// WithStartupHook 设置启动钩子
func (b *BaseBootstrapApp) WithStartupHook(hook StartupHook) *BaseBootstrapApp {
	b.startupHook = hook
	return b
}

// BaseInitialize 执行基础初始化逻辑
// 此方法封装了所有 Bootstrap 服务的通用初始化步骤
func (b *BaseBootstrapApp) BaseInitialize(ctx context.Context, config BootstrapConfig) error {
	// 初始化日志系统
	logger, err := config.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	b.logger = logger

	// 记录初始化日志
	fields := append([]interface{}{"service", b.serviceName}, config.GetLogFields()...)
	b.logger.Infow(fmt.Sprintf("Initializing %s Service", b.serviceName), fields...)

	// 创建 bootstrap 实例
	b.bootstrap = bootstrap.New(b.logger)

	// 注册所有组件初始化器
	if err := b.registrar.RegisterComponents(b.bootstrap); err != nil {
		return fmt.Errorf("failed to register components: %w", err)
	}

	b.logger.Infow("Components registered, ready to start")
	return nil
}

// BaseRun 执行基础运行逻辑
// 此方法封装了所有 Bootstrap 服务的通用运行步骤
func (b *BaseBootstrapApp) BaseRun(ctx context.Context, config BootstrapConfig) error {
	// 记录启动日志
	b.logger.Infow(
		fmt.Sprintf("%s Service started successfully", b.serviceName),
		config.GetLogFields()...,
	)

	// 构造 runFunc
	var runFunc func() error
	if b.startupHook != nil {
		runFunc = func() error {
			return b.startupHook.OnStartup(ctx)
		}
	}

	// 使用 bootstrap 的 Run 方法，它会等待信号
	return b.bootstrap.Run(ctx, runFunc)
}

// BaseShutdown 执行基础关闭逻辑
// 此方法封装了所有 Bootstrap 服务的通用关闭步骤
func (b *BaseBootstrapApp) BaseShutdown(ctx context.Context) error {
	b.logger.Infow(fmt.Sprintf("Shutting down %s Service", b.serviceName))
	return b.bootstrap.Shutdown(ctx)
}

// GetBootstrap 获取 bootstrap 实例
func (b *BaseBootstrapApp) GetBootstrap() *bootstrap.Bootstrap {
	return b.bootstrap
}

// GetLogger 获取 logger 实例
func (b *BaseBootstrapApp) GetLogger() core.Logger {
	return b.logger
}

// StandardInitLogger 提供标准的 initLogger 实现
// 各服务的 app.go 可以直接使用此函数
func StandardInitLogger(opts Options) (core.Logger, error) {
	config, ok := opts.(BootstrapConfig)
	if !ok {
		return nil, fmt.Errorf("options does not implement BootstrapConfig interface")
	}
	return config.InitLogger()
}
