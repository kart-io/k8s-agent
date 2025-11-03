// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"context"

	"github.com/kart-io/k8s-agent/common/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// GRPCHealthAdapter gRPC 健康检查适配器
// 将 common/health.Checker 适配为 gRPC health check
type GRPCHealthAdapter struct {
	manager *health.Manager
}

// NewGRPCHealthAdapter 创建 gRPC 适配器
func NewGRPCHealthAdapter(manager *health.Manager) *GRPCHealthAdapter {
	return &GRPCHealthAdapter{manager: manager}
}

// Check 实现 gRPC Health Check
func (a *GRPCHealthAdapter) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	if req.Service == "" {
		// 检查所有
		_, allHealthy := a.manager.CheckAll(ctx)
		if allHealthy {
			return &grpc_health_v1.HealthCheckResponse{
				Status: grpc_health_v1.HealthCheckResponse_SERVING,
			}, nil
		}
		return &grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
		}, nil
	}

	// 检查单个服务
	result, err := a.manager.Check(ctx, req.Service)
	if err != nil {
		return &grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN,
		}, nil
	}

	if result.Status == "healthy" {
		return &grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_SERVING,
		}, nil
	}

	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
	}, nil
}

// Watch 实现 gRPC Health Watch (流式)
func (a *GRPCHealthAdapter) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	// TODO: 实现流式健康检查
	return nil
}
