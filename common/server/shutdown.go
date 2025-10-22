package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kart-io/logger/core"
)

// ShutdownHandler 关闭处理器接口
type ShutdownHandler interface {
	Shutdown(ctx context.Context) error
}

// GracefulShutdown 优雅关闭 (监听系统信号)
// logger: 日志记录器
// timeout: 关闭超时时间
// handlers: 需要关闭的资源 (数据库、缓存、HTTP 服务器等)
func GracefulShutdown(logger core.Logger, timeout time.Duration, handlers ...ShutdownHandler) {
	quit := make(chan os.Signal, 1)
	// 监听 SIGINT (Ctrl+C) 和 SIGTERM (kill) 信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待信号
	sig := <-quit
	logger.Infow("Received shutdown signal", "signal", sig.String())
	logger.Infow("Shutting down gracefully...")

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 按顺序关闭所有资源
	for i, handler := range handlers {
		logger.Infow("Shutting down handler", "index", i+1, "total", len(handlers))
		if err := handler.Shutdown(ctx); err != nil {
			logger.Errorw("Shutdown error", "handler", i, "error", err)
		}
	}

	logger.Infow("Server exited successfully")
}

// WaitForShutdown 等待关闭信号但不执行关闭 (用于自定义关闭逻辑)
func WaitForShutdown(logger core.Logger) os.Signal {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Infow("Received shutdown signal", "signal", sig.String())
	return sig
}
