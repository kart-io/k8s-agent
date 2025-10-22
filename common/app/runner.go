package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
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

// LoggerInitFunc 日志初始化函数类型
type LoggerInitFunc func(opts Options) (*zap.Logger, error)

// ApplicationRunner 应用程序运行器
type ApplicationRunner struct {
	opts       Options
	app        Application
	logger     *zap.Logger
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
		defer logger.Sync()
	}

	// 2. 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 3. 设置信号处理
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 4. 初始化应用程序
	if err := r.app.Initialize(ctx, r.opts); err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}

	// 5. 在 goroutine 中运行应用程序
	errCh := make(chan error, 1)
	go func() {
		if err := r.app.Run(ctx); err != nil {
			errCh <- err
		}
	}()

	// 6. 等待信号或错误
	select {
	case sig := <-sigCh:
		if r.logger != nil {
			r.logger.Info("Received signal, shutting down", zap.String("signal", sig.String()))
		}
		cancel()
		return r.app.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}

// RunWithRunner 使用 ApplicationRunner 运行应用程序
func RunWithRunner(opts Options, app Application, loggerInit LoggerInitFunc, cfg CommandConfig) {
	runFunc := func(opts Options) error {
		runner := NewApplicationRunner(opts, app, loggerInit)
		return runner.Run()
	}

	Run(opts, runFunc, cfg)
}
