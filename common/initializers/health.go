package initializers

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/common/app"
	"github.com/kart-io/k8s-agent/common/bootstrap"
	"github.com/kart-io/logger/core"
)

// HealthCheckInitializer 健康检查初始化器（通用实现）
type HealthCheckInitializer struct {
	logger core.Logger
	port   string
	server *app.DefaultHealthCheckServer
}

// NewHealthCheckInitializer 创建健康检查初始化器
// port: 健康检查服务器端口，例如 ":8091"
// logger: 日志记录器
func NewHealthCheckInitializer(port string, logger core.Logger) *HealthCheckInitializer {
	if port == "" {
		port = ":8080" // 默认健康检查端口
	}

	return &HealthCheckInitializer{
		logger: logger,
		port:   port,
	}
}

// Name 返回初始化器名称
func (h *HealthCheckInitializer) Name() string {
	return "HealthCheck"
}

// Priority 返回初始化优先级
// 健康检查应该最后启动，优先级最低
func (h *HealthCheckInitializer) Priority() int {
	return bootstrap.PriorityLowest
}

// Initialize 初始化健康检查服务器
func (h *HealthCheckInitializer) Initialize(ctx context.Context) error {
	h.server = app.NewDefaultHealthCheckServer(h.port)
	if err := h.server.Start(); err != nil {
		return fmt.Errorf("failed to start health check server: %w", err)
	}

	h.logger.Infow("Health check server started",
		"port", h.port,
		"endpoints", []string{"/healthz", "/readyz"},
	)

	return nil
}

// Shutdown 关闭健康检查服务器
func (h *HealthCheckInitializer) Shutdown(ctx context.Context) error {
	if h.server != nil {
		h.logger.Infow("Stopping health check server")
		return h.server.Stop()
	}
	return nil
}
