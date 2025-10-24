package app

import (
	"fmt"
	"net/http"
	"time"
)

// DefaultHealthCheckServer 默认的健康检查服务器
type DefaultHealthCheckServer struct {
	addr string
	srv  *http.Server
}

// NewDefaultHealthCheckServer 创建默认的健康检查服务器
// addr: 监听地址，如 ":8090"
func NewDefaultHealthCheckServer(addr string) *DefaultHealthCheckServer {
	if addr == "" {
		addr = ":8090"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready"))
	})

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	return &DefaultHealthCheckServer{
		addr: addr,
		srv:  srv,
	}
}

// Start 启动健康检查服务器（非阻塞）
func (s *DefaultHealthCheckServer) Start() error {
	go func() {
		fmt.Printf("Health check server listening on %s\n", s.addr)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Health check server error: %v\n", err)
		}
	}()
	return nil
}

// Stop 停止健康检查服务器
func (s *DefaultHealthCheckServer) Stop() error {
	if s.srv != nil {
		return s.srv.Close()
	}
	return nil
}

// DefaultHealthCheckFunc 返回默认的健康检查函数
// 它会启动一个简单的 HTTP 服务器，监听 /healthz 和 /readyz 端点
func DefaultHealthCheckFunc(addr string) HealthCheckFunc {
	return func() error {
		server := NewDefaultHealthCheckServer(addr)
		return server.Start()
	}
}

// GetHealthCheckAddr 从 Options 中获取健康检查地址
// 如果 Options 实现了 HealthPortProvider 接口，则使用其提供的端口
// 否则返回默认地址 ":8090"
func GetHealthCheckAddr(opts Options) string {
	if provider, ok := opts.(HealthPortProvider); ok {
		port := provider.GetHealthPort()
		if port > 0 {
			return fmt.Sprintf(":%d", port)
		}
	}
	return ":8090" // 默认端口
}

// DefaultHealthCheckFuncFromOptions 从 Options 创建健康检查函数
// 它会自动检测 Options 是否实现了 HealthPortProvider 接口
func DefaultHealthCheckFuncFromOptions(opts Options) HealthCheckFunc {
	addr := GetHealthCheckAddr(opts)
	return DefaultHealthCheckFunc(addr)
}
