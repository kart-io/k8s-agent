package server

import (
	"context"
	"sync"

	"google.golang.org/grpc/health/grpc_health_v1"
)

// GRPCHealthChecker gRPC 健康检查器接口
// 与 HTTP HealthChecker 类似，但用于 gRPC 服务
type GRPCHealthChecker interface {
	// HealthCheck 执行健康检查
	// 返回 true 表示健康，false 表示不健康
	HealthCheck() (bool, error)
}

// HealthCheckManager 健康检查管理器
// 用于管理多个服务的健康状态
type HealthCheckManager struct {
	mu       sync.RWMutex
	checkers map[string]GRPCHealthChecker
	statuses map[string]grpc_health_v1.HealthCheckResponse_ServingStatus
}

// NewHealthCheckManager 创建健康检查管理器
func NewHealthCheckManager() *HealthCheckManager {
	return &HealthCheckManager{
		checkers: make(map[string]GRPCHealthChecker),
		statuses: make(map[string]grpc_health_v1.HealthCheckResponse_ServingStatus),
	}
}

// RegisterChecker 注册健康检查器
func (m *HealthCheckManager) RegisterChecker(service string, checker GRPCHealthChecker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkers[service] = checker
	m.statuses[service] = grpc_health_v1.HealthCheckResponse_SERVING
}

// UnregisterChecker 注销健康检查器
func (m *HealthCheckManager) UnregisterChecker(service string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.checkers, service)
	delete(m.statuses, service)
}

// SetStatus 设置服务状态
func (m *HealthCheckManager) SetStatus(service string, status grpc_health_v1.HealthCheckResponse_ServingStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statuses[service] = status
}

// GetStatus 获取服务状态
func (m *HealthCheckManager) GetStatus(service string) grpc_health_v1.HealthCheckResponse_ServingStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, ok := m.statuses[service]
	if !ok {
		return grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN
	}

	return status
}

// Check 执行健康检查
func (m *HealthCheckManager) Check(ctx context.Context, service string) grpc_health_v1.HealthCheckResponse_ServingStatus {
	m.mu.RLock()
	checker, ok := m.checkers[service]
	m.mu.RUnlock()

	if !ok {
		return grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN
	}

	healthy, err := checker.HealthCheck()
	if err != nil || !healthy {
		m.SetStatus(service, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		return grpc_health_v1.HealthCheckResponse_NOT_SERVING
	}

	m.SetStatus(service, grpc_health_v1.HealthCheckResponse_SERVING)
	return grpc_health_v1.HealthCheckResponse_SERVING
}

// CheckAll 检查所有服务
func (m *HealthCheckManager) CheckAll(ctx context.Context) map[string]grpc_health_v1.HealthCheckResponse_ServingStatus {
	m.mu.RLock()
	services := make([]string, 0, len(m.checkers))
	for service := range m.checkers {
		services = append(services, service)
	}
	m.mu.RUnlock()

	results := make(map[string]grpc_health_v1.HealthCheckResponse_ServingStatus)
	for _, service := range services {
		results[service] = m.Check(ctx, service)
	}

	return results
}
