// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package health

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

// Server 健康检查服务器接口
type Server interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	GetManager() *Manager
}

// HTTPServer HTTP 健康检查服务器实现
type HTTPServer struct {
	opts    *options.HealthOptions
	logger  core.Logger
	manager *Manager
	srv     *http.Server
}

// NewHTTPServer 创建 HTTP 健康检查服务器
func NewHTTPServer(opts *options.HealthOptions, logger core.Logger) Server {
	if opts == nil {
		opts = options.NewHealthOptions()
	}

	// 如果禁用，返回空实现
	if !opts.Enable {
		return &DisabledServer{}
	}

	// 完成配置
	if err := opts.Complete(); err != nil {
		opts = options.NewHealthOptions()
		_ = opts.Complete()
	}

	return &HTTPServer{
		opts:    opts,
		logger:  logger,
		manager: NewManager(),
	}
}

// Start 启动服务器
func (s *HTTPServer) Start(ctx context.Context) error {
	if s.srv != nil {
		return fmt.Errorf("health server already started")
	}

	mux := http.NewServeMux()

	// 注册健康检查端点
	mux.HandleFunc(s.opts.Path, s.handleHealth)
	mux.HandleFunc(s.opts.ReadinessPath, s.handleReadiness)
	mux.HandleFunc(s.opts.LivenessPath, s.handleLiveness)

	// 可选注册 pprof
	if s.opts.EnablePprof {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	s.srv = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.opts.Host, s.opts.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// 启动
	go func() {
		if s.logger != nil {
			s.logger.Infow("Health server starting",
				"addr", s.srv.Addr,
				"endpoints", []string{s.opts.Path, s.opts.ReadinessPath, s.opts.LivenessPath},
				"pprof", s.opts.EnablePprof,
			)
		}

		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			if s.logger != nil {
				s.logger.Errorw("Health server error", "error", err)
			}
		}
	}()

	return nil
}

// Stop 停止服务器
func (s *HTTPServer) Stop(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}

	if s.logger != nil {
		s.logger.Info("Stopping health server")
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, s.opts.ShutdownTimeout)
	defer cancel()

	return s.srv.Shutdown(shutdownCtx)
}

// GetManager 获取健康检查管理器
func (s *HTTPServer) GetManager() *Manager {
	return s.manager
}

// handleHealth 基本健康检查（不检查依赖）
func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// handleLiveness 存活探针（不检查依赖）
func (s *HTTPServer) handleLiveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("alive"))
}

// handleReadiness 就绪探针（检查所有依赖）
func (s *HTTPServer) handleReadiness(w http.ResponseWriter, r *http.Request) {
	results, allHealthy := s.manager.CheckAll(r.Context())

	status := "ready"
	statusCode := http.StatusOK

	if !allHealthy {
		status = "not ready"
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// 简单 JSON 输出
	fmt.Fprintf(w, `{"status":"%s","checks":%d}`, status, len(results))
}

// DisabledServer 禁用的服务器（空实现）
type DisabledServer struct{}

func (s *DisabledServer) Start(ctx context.Context) error { return nil }
func (s *DisabledServer) Stop(ctx context.Context) error  { return nil }
func (s *DisabledServer) GetManager() *Manager            { return nil }
