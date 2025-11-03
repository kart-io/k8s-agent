package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/common/server"
	"github.com/kart-io/logger/core"
)

// Application 定义应用程序接口
type Application interface {
	// Initialize 初始化应用程序
	Initialize(ctx context.Context, opts Options) error
	// Run 运行应用程序
	Run(ctx context.Context) error
	// Shutdown 优雅关闭应用程序
	Shutdown(ctx context.Context) error
}

// ServerProvider 定义提供 Server 实例的接口
// Application 可以实现此接口来让 ApplicationRunner 管理其 servers
type ServerProvider interface {
	// GetServer 返回 server 实例
	// 如果有多个 servers，应该返回它们的组合或者主 server
	GetServer() server.Server
}

// MultiServerProvider 定义提供多个 Server 实例的接口
// Application 可以实现此接口来提供多个 servers（如 HTTP + gRPC）
type MultiServerProvider interface {
	// GetServers 返回所有 server 实例
	GetServers() []server.Server
}

// LoggerInitFunc 日志初始化函数类型
type LoggerInitFunc func(opts Options) (core.Logger, error)

// ApplicationRunner 应用程序运行器
type ApplicationRunner struct {
	opts       Options
	app        Application
	logger     core.Logger
	loggerInit LoggerInitFunc
}

// NewApplicationRunner 创建新的应用程序运行器
func NewApplicationRunner(opts Options, app Application, loggerInit LoggerInitFunc) *ApplicationRunner {
	return &ApplicationRunner{
		opts:       opts,
		app:        app,
		loggerInit: loggerInit,
	}
}

// Run 运行应用程序
func (r *ApplicationRunner) Run() error {
	// 1. 初始化日志
	if r.loggerInit != nil {
		logger, err := r.loggerInit(r.opts)
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}
		r.logger = logger
	}

	// 2. 创建上下文
	ctx := context.Background()

	// 3. 初始化应用程序
	if err := r.app.Initialize(ctx, r.opts); err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}

	// 4. 收集 servers（如果 app 实现了 ServerProvider 或 MultiServerProvider）
	var servers []server.Server

	// 优先检查 MultiServerProvider
	if multiProvider, ok := r.app.(MultiServerProvider); ok {
		servers = multiProvider.GetServers()
		if r.logger != nil && len(servers) > 0 {
			r.logger.Infow("Collected servers from MultiServerProvider", "count", len(servers))
		}
	} else if provider, ok := r.app.(ServerProvider); ok {
		if srv := provider.GetServer(); srv != nil {
			servers = append(servers, srv)
			if r.logger != nil {
				r.logger.Infow("Collected server from ServerProvider")
			}
		}
	}

	// 5. 启动 servers（如果有）
	// 注意：servers 的启动由 server 包管理，这里只是触发
	// 实际上，如果 app 已经在 Initialize 或 Run 中启动了 servers，这里就不需要重复启动
	// 但为了完整性，我们提供这个选项

	// 这里不启动 servers，因为：
	// 1. Bootstrap 模式：servers 由 bootstrap.Run() 管理
	// 2. Simple 模式：servers 应该在 app.Run() 中启动
	// 3. 这样保持了向后兼容性

	// 6. 运行应用程序
	// 注意：app.Run() 会阻塞直到收到信号或发生错误
	// 信号处理由 bootstrap.Run() 或 app.Run() 负责，不需要在这里处理
	if err := r.app.Run(ctx); err != nil {
		return fmt.Errorf("application run failed: %w", err)
	}

	// 7. 正常退出（由信号触发）
	if r.logger != nil {
		r.logger.Infow("Application exited normally")
	}

	return nil
}

// RunWithRunner 使用 ApplicationRunner 运行应用程序
func RunWithRunner(opts Options, app Application, loggerInit LoggerInitFunc, cfg CommandConfig) {
	runFunc := func(opts Options) error {
		runner := NewApplicationRunner(opts, app, loggerInit)
		return runner.Run()
	}

	RunWithOptions(opts, runFunc, cfg)
}
