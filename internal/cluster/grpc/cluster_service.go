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

// ClusterGRPCService implements the gRPC ClusterService interface.
type ClusterGRPCService struct {
	clusterv1.UnimplementedClusterServiceServer
	clusterService *service.ClusterService
	logger         core.Logger
}

// NewClusterGRPCService creates a new ClusterGRPCService instance.
func NewClusterGRPCService(clusterService *service.ClusterService, logger core.Logger) *ClusterGRPCService {
	return &ClusterGRPCService{
		clusterService: clusterService,
		logger:         logger,
	}
}

// GetCluster retrieves cluster information by ID.
func (s *ClusterGRPCService) GetCluster(ctx context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.Cluster, error) {
	s.logger.Infow("GetCluster RPC called", "cluster_id", req.ClusterId)
	// TODO: Call clusterService to get cluster details
	return &clusterv1.Cluster{
		Id:   req.ClusterId,
		Name: "TODO: Implement",
	}, nil
}

// ListClusters lists all clusters.
func (s *ClusterGRPCService) ListClusters(ctx context.Context, req *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
	s.logger.Infow("ListClusters RPC called")
	// TODO: Implement listing logic
	return &clusterv1.ListClustersResponse{
		Clusters: []*clusterv1.Cluster{},
	}, nil
}

// CreateCluster creates a new cluster.
func (s *ClusterGRPCService) CreateCluster(ctx context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.Cluster, error) {
	s.logger.Infow("CreateCluster RPC called", "name", req.Name)
	// TODO: Implement creation logic
	return &clusterv1.Cluster{
		Name: req.Name,
	}, nil
}

// UpdateCluster updates an existing cluster.
func (s *ClusterGRPCService) UpdateCluster(ctx context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.Cluster, error) {
	s.logger.Infow("UpdateCluster RPC called", "cluster_id", req.ClusterId)
	// TODO: Implement update logic
	return &clusterv1.Cluster{
		Id: req.ClusterId,
	}, nil
}

// DeleteCluster deletes a cluster.
func (s *ClusterGRPCService) DeleteCluster(ctx context.Context, req *clusterv1.DeleteClusterRequest) (*clusterv1.DeleteClusterResponse, error) {
	s.logger.Infow("DeleteCluster RPC called", "cluster_id", req.ClusterId)
	// TODO: Implement deletion logic
	return &clusterv1.DeleteClusterResponse{
		Message: "Cluster deleted successfully",
	}, nil
}

// GetClusterHealth retrieves cluster health status.
func (s *ClusterGRPCService) GetClusterHealth(ctx context.Context, req *clusterv1.GetClusterHealthRequest) (*clusterv1.ClusterHealth, error) {
	s.logger.Infow("GetClusterHealth RPC called", "cluster_id", req.ClusterId)
	health, err := s.clusterService.GetClusterHealth(ctx, req.ClusterId)
	if err != nil {
		return nil, err
	}
	_ = health // TODO: Convert to Proto message
	return &clusterv1.ClusterHealth{
		Status: clusterv1.ClusterStatus_HEALTHY,
	}, nil
}

// GetClusterVersion retrieves cluster version information.
func (s *ClusterGRPCService) GetClusterVersion(ctx context.Context, req *clusterv1.GetClusterVersionRequest) (*clusterv1.ClusterVersion, error) {
	s.logger.Infow("GetClusterVersion RPC called", "cluster_id", req.ClusterId)
	// TODO: Implement version query
	return &clusterv1.ClusterVersion{
		KubernetesVersion: "TODO",
	}, nil
}
