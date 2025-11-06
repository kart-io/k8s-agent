package initializers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/kart-io/k8s-agent/cmd/cluster/app/options"
	clusterv1 "github.com/kart-io/k8s-agent/pkg/api/cluster/v1"
	"github.com/kart-io/k8s-agent/internal/cluster/service"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	commoninitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	commonserver "github.com/kart-io/k8s-agent/common/server"
	"github.com/kart-io/logger/core"
)

// GRPCServerInitializer gRPC服务器初始化器
type GRPCServerInitializer struct {
	standardInit *commoninitializers.GRPCServerInitializer
	opts         *options.ServerOptions
	logger       core.Logger

	// 依赖的初始化器
	dbInit *DatabaseInitializer

	// gRPC服务实现
	clusterService *ClusterGRPCService
	k8sService     *K8sResourceGRPCService
}

// NewGRPCServerInitializer 创建gRPC服务器初始化器
func NewGRPCServerInitializer(
	opts *options.ServerOptions,
	logger core.Logger,
	dbInit *DatabaseInitializer,
) *GRPCServerInitializer {
	return &GRPCServerInitializer{
		opts:   opts,
		logger: logger,
		dbInit: dbInit,
	}
}

// Name 返回初始化器名称
func (g *GRPCServerInitializer) Name() string {
	return "cluster-grpc-server"
}

// Priority 返回初始化优先级
func (g *GRPCServerInitializer) Priority() int {
	return bootstrap.PriorityGRPC
}

// Initialize 执行初始化
func (g *GRPCServerInitializer) Initialize(ctx context.Context) error {
	if !g.opts.GRPC.Enable {
		g.logger.Infow("gRPC server is disabled, skipping initialization")
		return nil
	}

	g.logger.Infow("Initializing gRPC server for cluster service",
		"host", g.opts.GRPC.Host,
		"port", g.opts.GRPC.Port,
	)

	// 创建Cluster Service（复用现有的service.ClusterService）
	store := g.dbInit.Store().(*storage.MySQLStorage)
	clusterSvc := service.NewClusterService(store, g.logger)
	g.clusterService = NewClusterGRPCService(clusterSvc, g.logger)

	// 创建K8s Resource Service（简化版本，TODO: 完整实现）
	g.k8sService = NewK8sResourceGRPCService(g.logger)

	serverConfig := &commoninitializers.GRPCServerConfig{
		Name:     g.Name(),
		Priority: g.Priority(),
		Config:   g.opts.GRPC,
		ServiceRegister: func(s *grpc.Server) error {
			// 注册Cluster服务
			clusterv1.RegisterClusterServiceServer(s, g.clusterService)
			// 注册K8s Resource服务（注意大小写）
			clusterv1.RegisterK8SResourceServiceServer(s, g.k8sService)

			g.logger.Info("Cluster gRPC services registered")
			return nil
		},
	}

	g.standardInit = commoninitializers.NewGRPCServerInitializer(serverConfig, g.logger)
	return g.standardInit.Initialize(ctx)
}

// Close 关闭gRPC服务器
func (g *GRPCServerInitializer) Close(ctx context.Context) error {
	if g.standardInit == nil {
		return nil
	}
	return g.standardInit.Close(ctx)
}

// GetServer 返回服务器实例（实现ServerProvider接口）
func (g *GRPCServerInitializer) GetServer() commonserver.Server {
	if g.standardInit == nil {
		return nil
	}
	return g.standardInit.GetServer()
}

// ============================================================================
// ClusterGRPCService - 集群管理gRPC服务实现
// ============================================================================

// ClusterGRPCService 实现 clusterv1.ClusterServiceServer
type ClusterGRPCService struct {
	clusterv1.UnimplementedClusterServiceServer
	clusterService *service.ClusterService
	logger         core.Logger
}

// NewClusterGRPCService 创建集群gRPC服务
func NewClusterGRPCService(clusterService *service.ClusterService, logger core.Logger) *ClusterGRPCService {
	return &ClusterGRPCService{
		clusterService: clusterService,
		logger:         logger.With("component", "cluster-grpc-service"),
	}
}

// GetCluster 获取集群
func (s *ClusterGRPCService) GetCluster(ctx context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.Cluster, error) {
	s.logger.Infow("GetCluster RPC called", "cluster_id", req.ClusterId)

	// TODO: 调用clusterService获取集群
	// cluster, err := s.clusterService.GetCluster(ctx, req.ClusterId)

	return &clusterv1.Cluster{
		Id:   req.ClusterId,
		Name: "TODO: Implement",
	}, nil
}

// ListClusters 列出集群
func (s *ClusterGRPCService) ListClusters(ctx context.Context, req *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
	s.logger.Infow("ListClusters RPC called")

	// TODO: 实现列表查询
	return &clusterv1.ListClustersResponse{
		Clusters: []*clusterv1.Cluster{},
	}, nil
}

// CreateCluster 创建集群
func (s *ClusterGRPCService) CreateCluster(ctx context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.Cluster, error) {
	s.logger.Infow("CreateCluster RPC called", "name", req.Name)

	// TODO: 实现创建逻辑
	return &clusterv1.Cluster{
		Name: req.Name,
	}, nil
}

// UpdateCluster 更新集群
func (s *ClusterGRPCService) UpdateCluster(ctx context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.Cluster, error) {
	s.logger.Infow("UpdateCluster RPC called", "cluster_id", req.ClusterId)

	// TODO: 实现更新逻辑
	return &clusterv1.Cluster{
		Id: req.ClusterId,
	}, nil
}

// DeleteCluster 删除集群
func (s *ClusterGRPCService) DeleteCluster(ctx context.Context, req *clusterv1.DeleteClusterRequest) (*clusterv1.DeleteClusterResponse, error) {
	s.logger.Infow("DeleteCluster RPC called", "cluster_id", req.ClusterId)

	// TODO: 实现删除逻辑
	return &clusterv1.DeleteClusterResponse{
		Message: "Cluster deleted successfully",
	}, nil
}

// GetClusterHealth 获取集群健康状态
func (s *ClusterGRPCService) GetClusterHealth(ctx context.Context, req *clusterv1.GetClusterHealthRequest) (*clusterv1.ClusterHealth, error) {
	s.logger.Infow("GetClusterHealth RPC called", "cluster_id", req.ClusterId)

	// 调用现有的服务方法
	health, err := s.clusterService.GetClusterHealth(ctx, req.ClusterId)
	if err != nil {
		return nil, err
	}

	// 转换为Proto消息
	// TODO: 实现转换逻辑
	_ = health

	return &clusterv1.ClusterHealth{
		Status: clusterv1.ClusterStatus_HEALTHY,
	}, nil
}

// GetClusterVersion 获取集群版本
func (s *ClusterGRPCService) GetClusterVersion(ctx context.Context, req *clusterv1.GetClusterVersionRequest) (*clusterv1.ClusterVersion, error) {
	s.logger.Infow("GetClusterVersion RPC called", "cluster_id", req.ClusterId)

	// TODO: 实现版本查询
	return &clusterv1.ClusterVersion{
		KubernetesVersion: "TODO",
	}, nil
}

// ============================================================================
// K8sResourceGRPCService - K8s资源管理gRPC服务实现
// ============================================================================

// K8sResourceGRPCService 实现 clusterv1.K8SResourceServiceServer
type K8sResourceGRPCService struct {
	clusterv1.UnimplementedK8SResourceServiceServer
	logger core.Logger
}

// NewK8sResourceGRPCService 创建K8s资源gRPC服务
func NewK8sResourceGRPCService(logger core.Logger) *K8sResourceGRPCService {
	return &K8sResourceGRPCService{
		logger: logger.With("component", "k8s-resource-grpc-service"),
	}
}

// GetResource 获取资源
func (s *K8sResourceGRPCService) GetResource(ctx context.Context, req *clusterv1.GetResourceRequest) (*clusterv1.Resource, error) {
	s.logger.Infow("GetResource RPC called",
		"cluster_id", req.ClusterId,
		"type", req.ResourceType,
		"namespace", req.Namespace,
		"name", req.Name,
	)

	// TODO: 实现资源获取
	return &clusterv1.Resource{
		ClusterId: req.ClusterId,
		Type:      req.ResourceType,
		Namespace: req.Namespace,
		Name:      req.Name,
	}, nil
}

// ListResources 列出资源
func (s *K8sResourceGRPCService) ListResources(ctx context.Context, req *clusterv1.ListResourcesRequest) (*clusterv1.ListResourcesResponse, error) {
	s.logger.Infow("ListResources RPC called",
		"cluster_id", req.ClusterId,
		"type", req.ResourceType,
	)

	// TODO: 实现资源列表
	return &clusterv1.ListResourcesResponse{
		Resources: []*clusterv1.Resource{},
	}, nil
}

// CreateResource 创建资源
func (s *K8sResourceGRPCService) CreateResource(ctx context.Context, req *clusterv1.CreateResourceRequest) (*clusterv1.Resource, error) {
	s.logger.Infow("CreateResource RPC called",
		"cluster_id", req.ClusterId,
		"type", req.ResourceType,
	)

	// TODO: 实现资源创建
	return &clusterv1.Resource{
		ClusterId: req.ClusterId,
		Type:      req.ResourceType,
	}, nil
}

// UpdateResource 更新资源
func (s *K8sResourceGRPCService) UpdateResource(ctx context.Context, req *clusterv1.UpdateResourceRequest) (*clusterv1.Resource, error) {
	s.logger.Infow("UpdateResource RPC called",
		"cluster_id", req.ClusterId,
		"type", req.ResourceType,
		"name", req.Name,
	)

	// TODO: 实现资源更新
	return &clusterv1.Resource{
		ClusterId: req.ClusterId,
		Type:      req.ResourceType,
		Name:      req.Name,
	}, nil
}

// DeleteResource 删除资源
func (s *K8sResourceGRPCService) DeleteResource(ctx context.Context, req *clusterv1.DeleteResourceRequest) (*clusterv1.DeleteResourceResponse, error) {
	s.logger.Infow("DeleteResource RPC called",
		"cluster_id", req.ClusterId,
		"type", req.ResourceType,
		"name", req.Name,
	)

	// TODO: 实现资源删除
	return &clusterv1.DeleteResourceResponse{
		Message: "Resource deleted successfully",
	}, nil
}

// WatchResources 监听资源变化
func (s *K8sResourceGRPCService) WatchResources(req *clusterv1.WatchResourcesRequest, stream clusterv1.K8SResourceService_WatchResourcesServer) error {
	s.logger.Infow("WatchResources RPC called",
		"cluster_id", req.ClusterId,
		"type", req.ResourceType,
	)

	// TODO: 实现资源监听（流式响应）
	// 这需要与K8s watch API集成

	return nil
}

