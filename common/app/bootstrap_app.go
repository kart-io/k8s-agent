// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/common/bootstrap"
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

// MiddlewareFunc 定义中间件函数类型
// 中间件在组件注册时被调用，可以添加额外的初始化器
type MiddlewareFunc func(bs *bootstrap.Bootstrap, logger core.Logger, opts Options) error

// MiddlewareConfig 定义中间件配置
type MiddlewareConfig struct {
	Name     string         // 中间件名称（用于日志）
	Priority int            // 优先级（数值越小越先执行，默认 500）
	Func     MiddlewareFunc // 中间件函数
}

// Middleware 创建中间件配置
func Middleware(name string, priority int, fn MiddlewareFunc) MiddlewareConfig {
	return MiddlewareConfig{
		Name:     name,
		Priority: priority,
		Func:     fn,
	}
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

	// 如果 registrar 是 middlewareRegistrar，设置 logger
	if mw, ok := b.registrar.(*middlewareRegistrar); ok {
		mw.logger = logger
	}

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

// ================= 以下是新增的通用方法，用于进一步减少重复代码 =================

// ConfigHolder 定义持有配置的接口
// 服务的 App 结构体应该实现此接口来存储 opts
type ConfigHolder interface {
	SetOptions(opts Options)
	GetOptions() Options
}

// BootstrapAppBuilder 提供流式 API 来构建 BaseBootstrapApp
type BootstrapAppBuilder struct {
	app         *BaseBootstrapApp
	serviceName string
	registrar   ComponentRegistrar
	hook        StartupHook
}

// NewBootstrapAppBuilder 创建构建器
func NewBootstrapAppBuilder(serviceName string, registrar ComponentRegistrar) *BootstrapAppBuilder {
	return &BootstrapAppBuilder{
		serviceName: serviceName,
		registrar:   registrar,
	}
}

// WithHook 设置启动钩子
func (b *BootstrapAppBuilder) WithHook(hook StartupHook) *BootstrapAppBuilder {
	b.hook = hook
	return b
}

// Build 构建 BaseBootstrapApp
func (b *BootstrapAppBuilder) Build() *BaseBootstrapApp {
	app := NewBaseBootstrapApp(b.serviceName, b.registrar)
	if b.hook != nil {
		app.WithStartupHook(b.hook)
	}
	return app
}

// StandardInitialize 提供标准的 Initialize 实现
// 这个方法可以被所有 Bootstrap 服务直接使用，进一步减少代码重复
func StandardInitialize(ctx context.Context, opts Options, holder ConfigHolder, serviceName string, registrar ComponentRegistrar, hook StartupHook) (*BaseBootstrapApp, error) {
	// 保存 options
	holder.SetOptions(opts)

	// 创建 BaseBootstrapApp
	builder := NewBootstrapAppBuilder(serviceName, registrar)
	if hook != nil {
		builder.WithHook(hook)
	}
	app := builder.Build()

	// 执行初始化
	config, ok := opts.(BootstrapConfig)
	if !ok {
		return nil, fmt.Errorf("options does not implement BootstrapConfig interface")
	}

	if err := app.BaseInitialize(ctx, config); err != nil {
		return nil, err
	}

	return app, nil
}

// StandardRun 提供标准的 Run 实现
func StandardRun(ctx context.Context, app *BaseBootstrapApp, opts Options) error {
	config, ok := opts.(BootstrapConfig)
	if !ok {
		return fmt.Errorf("options does not implement BootstrapConfig interface")
	}
	return app.BaseRun(ctx, config)
}

// StandardShutdown 提供标准的 Shutdown 实现
func StandardShutdown(ctx context.Context, app *BaseBootstrapApp) error {
	return app.BaseShutdown(ctx)
}

// ================= StandardBootstrapApplication 完全封装的 Bootstrap 应用 =================

// StandardBootstrapApplication 提供完全封装的 Bootstrap 应用实现
// 服务只需组合此结构体，无需编写 Initialize/Run/Shutdown/initLogger 方法
type StandardBootstrapApplication struct {
	*BaseBootstrapApp
	opts          Options
	serviceName   string
	registrar     ComponentRegistrar
	startupHook   StartupHook
	configHandler func(Options) error // 可选的配置处理函数
	middlewares   []MiddlewareConfig  // 中间件列表
}

// NewStandardBootstrapApplication 创建标准 Bootstrap 应用
func NewStandardBootstrapApplication(serviceName string, registrar ComponentRegistrar) *StandardBootstrapApplication {
	return &StandardBootstrapApplication{
		serviceName: serviceName,
		registrar:   registrar,
		middlewares: make([]MiddlewareConfig, 0),
	}
}

// WithStartupHookFunc 设置启动钩子
func (s *StandardBootstrapApplication) WithStartupHookFunc(hook StartupHook) *StandardBootstrapApplication {
	s.startupHook = hook
	return s
}

// WithConfigHandler 设置配置处理函数（用于处理 Config() 转换）
func (s *StandardBootstrapApplication) WithConfigHandler(handler func(Options) error) *StandardBootstrapApplication {
	s.configHandler = handler
	return s
}

// WithMiddleware 添加单个中间件
func (s *StandardBootstrapApplication) WithMiddleware(middleware MiddlewareConfig) *StandardBootstrapApplication {
	s.middlewares = append(s.middlewares, middleware)
	return s
}

// WithMiddlewares 批量添加中间件
func (s *StandardBootstrapApplication) WithMiddlewares(middlewares ...MiddlewareConfig) *StandardBootstrapApplication {
	s.middlewares = append(s.middlewares, middlewares...)
	return s
}

// WithMiddlewareFunc 添加简单的中间件函数（使用默认优先级 500）
func (s *StandardBootstrapApplication) WithMiddlewareFunc(name string, fn MiddlewareFunc) *StandardBootstrapApplication {
	return s.WithMiddleware(Middleware(name, 500, fn))
}

// Initialize 实现 Application 接口
func (s *StandardBootstrapApplication) Initialize(ctx context.Context, opts Options) error {
	s.opts = opts

	// 执行配置处理（如果有）
	if s.configHandler != nil {
		if err := s.configHandler(opts); err != nil {
			return fmt.Errorf("failed to handle config: %w", err)
		}
	}

	// 创建 BaseBootstrapApp，使用中间件包装的 registrar
	wrappedRegistrar := s.createMiddlewareRegistrar()
	s.BaseBootstrapApp = NewBaseBootstrapApp(s.serviceName, wrappedRegistrar)
	if s.startupHook != nil {
		s.BaseBootstrapApp.WithStartupHook(s.startupHook)
	}

	// 执行初始化
	config, ok := opts.(BootstrapConfig)
	if !ok {
		return fmt.Errorf("options does not implement BootstrapConfig interface")
	}

	return s.BaseInitialize(ctx, config)
}

// createMiddlewareRegistrar 创建包装了中间件的组件注册器
func (s *StandardBootstrapApplication) createMiddlewareRegistrar() ComponentRegistrar {
	return &middlewareRegistrar{
		original:    s.registrar,
		middlewares: s.middlewares,
		opts:        s.opts,
		logger:      nil, // logger 会在 RegisterComponents 被调用时通过参数传递
	}
}

// middlewareRegistrar 包装原始的 ComponentRegistrar，添加中间件支持
type middlewareRegistrar struct {
	original    ComponentRegistrar
	middlewares []MiddlewareConfig
	opts        Options
	logger      core.Logger
}

// RegisterComponents 实现 ComponentRegistrar 接口，先执行中间件，再执行原始注册
func (m *middlewareRegistrar) RegisterComponents(bs *bootstrap.Bootstrap) error {
	// 从 BaseBootstrapApp 获取 logger
	// 注意：此时 logger 已经在 BaseInitialize 中初始化
	logger := m.logger
	if logger == nil {
		// 如果没有预先设置 logger，返回错误
		return fmt.Errorf("logger not initialized in middleware registrar")
	}

	// 按优先级排序中间件
	sortedMiddlewares := make([]MiddlewareConfig, len(m.middlewares))
	copy(sortedMiddlewares, m.middlewares)

	// 简单的冒泡排序（中间件数量不多，性能影响可忽略）
	for i := 0; i < len(sortedMiddlewares)-1; i++ {
		for j := 0; j < len(sortedMiddlewares)-i-1; j++ {
			if sortedMiddlewares[j].Priority > sortedMiddlewares[j+1].Priority {
				sortedMiddlewares[j], sortedMiddlewares[j+1] = sortedMiddlewares[j+1], sortedMiddlewares[j]
			}
		}
	}

	// 执行中间件
	for _, middleware := range sortedMiddlewares {
		logger.Infow("Applying middleware", "name", middleware.Name, "priority", middleware.Priority)
		if err := middleware.Func(bs, logger, m.opts); err != nil {
			return fmt.Errorf("middleware %s failed: %w", middleware.Name, err)
		}
	}

	// 执行原始的组件注册
	return m.original.RegisterComponents(bs)
}

// Run 实现 Application 接口
func (s *StandardBootstrapApplication) Run(ctx context.Context) error {
	config, ok := s.opts.(BootstrapConfig)
	if !ok {
		return fmt.Errorf("options does not implement BootstrapConfig interface")
	}
	return s.BaseRun(ctx, config)
}

// Shutdown 实现 Application 接口
func (s *StandardBootstrapApplication) Shutdown(ctx context.Context) error {
	return s.BaseShutdown(ctx)
}

// GetOptions 获取 options
func (s *StandardBootstrapApplication) GetOptions() Options {
	return s.opts
}
