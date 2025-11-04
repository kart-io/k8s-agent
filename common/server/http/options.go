// Copyright 2025 Kart Project. All rights reserved.
// HTTP server implementation using Options pattern
// 参考 OneX 项目设计

package server

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"

	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

// HTTPOptionsServer 使用 Options 配置模式的 HTTP 服务器
// 参考 OneX 项目设计，实现 Server 接口
type HTTPOptionsServer struct {
	srv *http.Server
	log core.Logger
}

// NewHTTPOptionsServer 创建使用 Options 配置的 HTTP 服务器
// 参考 OneX 项目的 NewHTTPServer 设计
//
// 参数:
//   - httpOpts: HTTP 服务器配置选项
//   - tlsOpts: TLS 配置选项 (可选，为 nil 则使用 HTTP)
//   - handler: HTTP 请求处理器
//   - log: 日志记录器
//
// 使用示例:
//
//	// HTTP 服务器
//	httpOpts := options.NewHTTPServerOptions()
//	httpSrv := server.NewHTTPOptionsServer(httpOpts, nil, handler, log)
//
//	// HTTPS 服务器
//	tlsOpts := options.NewTLSOptions()
//	tlsOpts.UseTLS = true
//	httpsSrv := server.NewHTTPOptionsServer(httpOpts, tlsOpts, handler, log)
//
//	// 启动服务器
//	ctx, cancel := context.WithCancel(context.Background())
//	if err := server.Serve(ctx, httpSrv, log); err != nil {
//	    log.Fatalw("Server failed", "err", err)
//	}
func NewHTTPOptionsServer(
	httpOpts *options.HTTPServerOptions,
	tlsOpts *options.TLSOptions,
	handler http.Handler,
	log core.Logger,
) *HTTPOptionsServer {
	// 准备 TLS 配置
	var tlsConfig *tls.Config
	if tlsOpts != nil && tlsOpts.UseTLS {
		// 加载 TLS 证书和密钥
		cert, err := tls.LoadX509KeyPair(tlsOpts.Cert, tlsOpts.Key)
		if err != nil {
			log.Fatalw("Failed to load TLS certificate", "err", err)
		}

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tlsOpts.MinVersion,
		}
	}

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:           httpOpts.Addr,
		Handler:        handler,
		TLSConfig:      tlsConfig,
		ReadTimeout:    httpOpts.ReadTimeout,
		WriteTimeout:   httpOpts.WriteTimeout,
		IdleTimeout:    httpOpts.IdleTimeout,
		MaxHeaderBytes: httpOpts.MaxHeaderBytes,
	}

	return &HTTPOptionsServer{
		srv: srv,
		log: log.With("component", "http-options-server", "addr", srv.Addr),
	}
}

// RunOrDie 启动 HTTP 服务器并在出错时记录致命错误
// 实现 Server 接口
func (s *HTTPOptionsServer) RunOrDie() {
	protocol := "http"
	if s.srv.TLSConfig != nil {
		protocol = "https"
	}

	s.log.Infow("Starting HTTP server",
		"protocol", protocol,
		"addr", s.srv.Addr,
	)

	// 根据是否配置 TLS 选择启动方式
	var err error
	if s.srv.TLSConfig != nil {
		// HTTPS 服务器 (证书已在 TLSConfig 中配置)
		err = s.srv.ListenAndServeTLS("", "")
	} else {
		// HTTP 服务器
		err = s.srv.ListenAndServe()
	}

	// 正常关闭不算错误
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.log.Fatalw("HTTP server failed", "protocol", protocol, "err", err)
	}
}

// GracefulStop 优雅地关闭 HTTP 服务器
// 实现 Server 接口
func (s *HTTPOptionsServer) GracefulStop(ctx context.Context) {
	s.log.Infow("Gracefully stopping HTTP server")

	if err := s.srv.Shutdown(ctx); err != nil {
		s.log.Errorw("HTTP server forced to shutdown", "err", err)
		return
	}

	s.log.Infow("HTTP server stopped gracefully")
}

// GetHTTPServer 返回底层的 *http.Server (用于高级配置)
func (s *HTTPOptionsServer) GetHTTPServer() *http.Server {
	return s.srv
}

// Addr 返回服务器监听地址
func (s *HTTPOptionsServer) Addr() string {
	return s.srv.Addr
}
