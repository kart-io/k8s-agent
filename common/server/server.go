// Copyright 2025 Kart Project. All rights reserved.
// Server package provides unified HTTP and gRPC server implementations
// with graceful shutdown support.

package server

import (
	"context"
	"time"

	"github.com/kart-io/logger/core"
)

// Server 定义所有服务器类型的统一接口
// 参考 OneX 项目设计，提供标准化的服务器生命周期管理
type Server interface {
	// RunOrDie 运行服务器，如果运行失败会退出程序
	// 这是一个阻塞方法，会持续运行直到服务器停止
	RunOrDie()

	// GracefulStop 优雅关停服务器
	// ctx 控制关停的超时时间
	GracefulStop(ctx context.Context)
}

// Serve 启动服务器并阻塞直到 context 被取消
// 参考 OneX 项目设计，提供统一的服务器启动和关停逻辑
//
// 使用示例:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	// 监听信号
//	go func() {
//	    sigCh := make(chan os.Signal, 1)
//	    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
//	    <-sigCh
//	    cancel()
//	}()
//
//	// 启动服务器
//	if err := server.Serve(ctx, srv, log); err != nil {
//	    log.Fatalw("Server failed", "err", err)
//	}
func Serve(ctx context.Context, srv Server, log core.Logger) error {
	// 在 goroutine 中启动服务器
	go srv.RunOrDie()

	// 阻塞直到 context 被取消或终止
	<-ctx.Done()

	// 优雅关停服务器
	log.Infow("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 优雅停止服务器
	srv.GracefulStop(shutdownCtx)

	log.Infow("Server exited successfully")

	return nil
}

// MultiServe 同时启动多个服务器（例如同时启动 HTTP 和 gRPC）
// 当 context 被取消时，所有服务器会并发执行优雅关停
//
// 使用示例:
//
//	httpSrv := server.NewHTTPServer(httpOpts, tlsOpts, handler, log)
//	grpcSrv, _ := server.NewGRPCServer(grpcOpts, tlsOpts, nil, registerFunc, log)
//
//	if err := server.MultiServe(ctx, log, httpSrv, grpcSrv); err != nil {
//	    log.Fatalw("Servers failed", "err", err)
//	}
func MultiServe(ctx context.Context, log core.Logger, servers ...Server) error {
	// 启动所有服务器
	for _, srv := range servers {
		go srv.RunOrDie()
	}

	// 阻塞直到 context 被取消
	<-ctx.Done()

	// 优雅关停所有服务器
	log.Infow("Shutting down servers...", "count", len(servers))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 并发关停所有服务器
	done := make(chan struct{})
	for _, srv := range servers {
		srv := srv // 避免闭包问题
		go func() {
			srv.GracefulStop(shutdownCtx)
			done <- struct{}{}
		}()
	}

	// 等待所有服务器关停完成
	for i := 0; i < len(servers); i++ {
		<-done
	}

	log.Infow("All servers exited successfully")

	return nil
}
