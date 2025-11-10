// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package grpc

import (
	"context"

	"github.com/kart-io/k8s-agent/internal/cluster/service"
	clusterv1 "github.com/kart-io/k8s-agent/pkg/api/cluster/v1"
	"github.com/kart-io/logger/core"
)

// K8sResourceGRPCService implements the gRPC K8sResourceService interface.
type K8sResourceGRPCService struct {
	clusterv1.UnimplementedK8SResourceServiceServer
	registry *service.K8sServiceRegistry
	logger   core.Logger
}

// NewK8sResourceGRPCService creates a new K8sResourceGRPCService instance.
func NewK8sResourceGRPCService(registry *service.K8sServiceRegistry, logger core.Logger) *K8sResourceGRPCService {
	return &K8sResourceGRPCService{
		registry: registry,
		logger:   logger,
	}
}

// GetResource retrieves a Kubernetes resource.
func (s *K8sResourceGRPCService) GetResource(ctx context.Context, req *clusterv1.GetResourceRequest) (*clusterv1.Resource, error) {
	s.logger.Infow("GetResource RPC called",
		"cluster_id", req.ClusterId,
		"type", req.ResourceType,
		"namespace", req.Namespace,
		"name", req.Name,
	)
	// TODO: Implement resource retrieval
	return &clusterv1.Resource{
		ClusterId: req.ClusterId,
		Type:      req.ResourceType,
		Namespace: req.Namespace,
		Name:      req.Name,
	}, nil
}

// ListResources lists Kubernetes resources.
func (s *K8sResourceGRPCService) ListResources(ctx context.Context, req *clusterv1.ListResourcesRequest) (*clusterv1.ListResourcesResponse, error) {
	s.logger.Infow("ListResources RPC called",
		"cluster_id", req.ClusterId,
		"type", req.ResourceType,
	)
	// TODO: Implement resource listing
	return &clusterv1.ListResourcesResponse{
		Resources: []*clusterv1.Resource{},
	}, nil
}

// CreateResource creates a new Kubernetes resource.
func (s *K8sResourceGRPCService) CreateResource(ctx context.Context, req *clusterv1.CreateResourceRequest) (*clusterv1.Resource, error) {
	s.logger.Infow("CreateResource RPC called",
		"cluster_id", req.ClusterId,
		"type", req.ResourceType,
	)
	// TODO: Implement resource creation
	return &clusterv1.Resource{
		ClusterId: req.ClusterId,
		Type:      req.ResourceType,
	}, nil
}

// UpdateResource updates a Kubernetes resource.
func (s *K8sResourceGRPCService) UpdateResource(ctx context.Context, req *clusterv1.UpdateResourceRequest) (*clusterv1.Resource, error) {
	s.logger.Infow("UpdateResource RPC called",
		"cluster_id", req.ClusterId,
		"type", req.ResourceType,
		"name", req.Name,
	)
	// TODO: Implement resource update
	return &clusterv1.Resource{
		ClusterId: req.ClusterId,
		Type:      req.ResourceType,
		Name:      req.Name,
	}, nil
}

// DeleteResource deletes a Kubernetes resource.
func (s *K8sResourceGRPCService) DeleteResource(ctx context.Context, req *clusterv1.DeleteResourceRequest) (*clusterv1.DeleteResourceResponse, error) {
	s.logger.Infow("DeleteResource RPC called",
		"cluster_id", req.ClusterId,
		"type", req.ResourceType,
		"name", req.Name,
	)
	// TODO: Implement resource deletion
	return &clusterv1.DeleteResourceResponse{
		Message: "Resource deleted successfully",
	}, nil
}

// WatchResources watches Kubernetes resource changes.
func (s *K8sResourceGRPCService) WatchResources(req *clusterv1.WatchResourcesRequest, stream clusterv1.K8SResourceService_WatchResourcesServer) error {
	s.logger.Infow("WatchResources RPC called",
		"cluster_id", req.ClusterId,
		"type", req.ResourceType,
	)
	// TODO: Implement streaming resource watch
	return nil
}
