package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// ShutdownHandler 关闭处理器接口
type ShutdownHandler interface {
	Shutdown(ctx context.Context) error
}

// GracefulShutdown 优雅关闭 (监听系统信号)
// logger: 日志记录器
// timeout: 关闭超时时间
// handlers: 需要关闭的资源 (数据库、缓存、HTTP 服务器等)
func GracefulShutdown(logger *zap.Logger, timeout time.Duration, handlers ...ShutdownHandler) {
	quit := make(chan os.Signal, 1)
	// 监听 SIGINT (Ctrl+C) 和 SIGTERM (kill) 信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待信号
	sig := <-quit
	logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
	logger.Info("Shutting down gracefully...")

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 按顺序关闭所有资源
	for i, handler := range handlers {
		logger.Info("Shutting down handler", zap.Int("index", i+1), zap.Int("total", len(handlers)))
		if err := handler.Shutdown(ctx); err != nil {
			logger.Error("Shutdown error", zap.Int("handler", i), zap.Error(err))
		}
	}

	logger.Info("Server exited successfully")
}

// WaitForShutdown 等待关闭信号但不执行关闭 (用于自定义关闭逻辑)
func WaitForShutdown(logger *zap.Logger) os.Signal {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
	return sig
}
