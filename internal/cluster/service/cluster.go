package service

import (
	"context"
	stderr "errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/internal/cluster/k8s"
	clustermodel "github.com/kart-io/k8s-agent/internal/models/cluster"
	"github.com/kart-io/logger/core"
)

// ClusterService 集群管理服务（统一版本）
type ClusterService struct {
	db      *gorm.DB               // 直接使用 GORM DB (来自 pkg/initializers)
	clients map[string]*k8s.Client // cluster_id -> k8s client 缓存
	log     core.Logger
}

// K8sClusterService 是 ClusterService 的别名，用于向后兼容
// DEPRECATED: 其他 K8s 服务应该直接使用 ClusterService
type K8sClusterService = ClusterService

// NewClusterService 创建集群服务
// db 参数应来自 pkg/initializers.DatabaseInitializer.DB()
func NewClusterService(db *gorm.DB, logger core.Logger) *ClusterService {
	return &ClusterService{
		db:      db,
		clients: make(map[string]*k8s.Client),
		log:     logger,
	}
}

// NewK8sClusterService 是 NewClusterService 的别名，用于向后兼容
// DEPRECATED: 使用 NewClusterService 替代
func NewK8sClusterService(db *gorm.DB, logger core.Logger) *K8sClusterService {
	return NewClusterService(db, logger)
}

// ========== CRUD 操作 ==========

// CreateCluster 创建集群
func (s *ClusterService) CreateCluster(ctx context.Context, req *CreateClusterRequest) (*clustermodel.Cluster, error) {
	// 1. 验证 kubeconfig 并测试连接
	client, err := k8s.NewClientFromKubeConfig([]byte(req.KubeConfig))
	if err != nil {
		return nil, errors.NewValidationError(fmt.Errorf("invalid kubeconfig: %w", err))
	}

	if err := client.CheckConnection(ctx); err != nil {
		return nil, errors.NewValidationError(fmt.Errorf("failed to connect to cluster: %w", err))
	}

	// 2. 获取集群版本
	version, err := client.GetServerVersion(ctx)
	if err != nil {
		s.log.Warnw("Failed to get server version", "error", err)
		version = clustermodel.StatusUnknown
	}

	// 3. 创建数据库记录
	cluster := &clustermodel.Cluster{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Endpoint:    req.Endpoint,
		Version:     version,
		Status:      clustermodel.StatusHealthy,
		Region:      req.Region,
		Provider:    req.Provider,
		KubeConfig:  req.KubeConfig,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(cluster).Error; err != nil {
		return nil, errors.NewDatabaseError(fmt.Errorf("failed to create cluster: %w", err))
	}

	// 4. 缓存客户端
	s.clients[cluster.ID] = client

	s.log.Infow("Cluster created",
		"cluster_id", cluster.ID,
		"name", cluster.Name,
		"version", version,
	)

	return cluster, nil
}

// AddCluster 是 CreateCluster 的别名，用于向后兼容
// DEPRECATED: 使用 CreateCluster 替代
func (s *ClusterService) AddCluster(ctx context.Context, req *CreateClusterRequest) (*clustermodel.Cluster, error) {
	return s.CreateCluster(ctx, req)
}

// ListClusters 获取集群列表（分页）
func (s *ClusterService) ListClusters(ctx context.Context, offset, limit int, withStats bool) ([]*clustermodel.Cluster, int64, error) {
	// 1. 查询总数
	var total int64
	if err := s.db.WithContext(ctx).Model(&clustermodel.Cluster{}).Count(&total).Error; err != nil {
		return nil, 0, errors.NewDatabaseError(fmt.Errorf("failed to count clusters: %w", err))
	}

	// 2. 查询列表
	var clusters []*clustermodel.Cluster
	if err := s.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&clusters).Error; err != nil {
		return nil, 0, errors.NewDatabaseError(fmt.Errorf("failed to query clusters: %w", err))
	}

	// 3. 可选: 填充统计信息
	if withStats {
		for i := range clusters {
			if err := s.populateClusterStats(ctx, clusters[i]); err != nil {
				s.log.Warnw("Failed to populate cluster stats",
					"cluster_id", clusters[i].ID,
					"error", err,
				)
				// 统计信息获取失败不影响列表返回
			}
		}
	}

	return clusters, total, nil
}

// GetCluster 获取集群详情
func (s *ClusterService) GetCluster(ctx context.Context, clusterID string, withStats bool) (*clustermodel.Cluster, error) {
	var cluster clustermodel.Cluster
	if err := s.db.WithContext(ctx).Where("id = ?", clusterID).First(&cluster).Error; err != nil {
		if stderr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrClusterNotFound
		}
		return nil, errors.NewDatabaseError(fmt.Errorf("failed to query cluster: %w", err))
	}

	// 可选: 填充统计信息
	if withStats {
		if err := s.populateClusterStats(ctx, &cluster); err != nil {
			s.log.Warnw("Failed to populate cluster stats",
				"cluster_id", clusterID,
				"error", err,
			)
		}
	}

	return &cluster, nil
}

// UpdateCluster 更新集群信息
func (s *ClusterService) UpdateCluster(ctx context.Context, clusterID string, req *UpdateClusterRequest) (*clustermodel.Cluster, error) {
	// 1. 检查集群是否存在
	if _, err := s.GetCluster(ctx, clusterID, false); err != nil {
		return nil, err
	}

	// 2. 构建更新字段
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}

	// 3. 执行更新
	if err := s.db.WithContext(ctx).
		Model(&clustermodel.Cluster{}).
		Where("id = ?", clusterID).
		Updates(updates).Error; err != nil {
		return nil, errors.NewDatabaseError(fmt.Errorf("failed to update cluster: %w", err))
	}

	s.log.Infow("Cluster updated", "cluster_id", clusterID)

	return s.GetCluster(ctx, clusterID, false)
}

// DeleteCluster 删除集群
func (s *ClusterService) DeleteCluster(ctx context.Context, clusterID string) error {
	result := s.db.WithContext(ctx).Where("id = ?", clusterID).Delete(&clustermodel.Cluster{})
	if result.Error != nil {
		return errors.NewDatabaseError(fmt.Errorf("failed to delete cluster: %w", result.Error))
	}

	if result.RowsAffected == 0 {
		return errors.ErrClusterNotFound
	}

	// 清除缓存
	delete(s.clients, clusterID)

	s.log.Infow("Cluster deleted", "cluster_id", clusterID)

	return nil
}

// ========== K8s 资源查询操作 ==========

// GetClusterHealth 获取集群健康状态
func (s *ClusterService) GetClusterHealth(ctx context.Context, clusterID string) (*clustermodel.ClusterHealth, error) {
	client, err := s.GetClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	// 获取节点列表
	nodes, err := client.Clientset().CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to list nodes: %w", err))
	}

	readyNodes := 0
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == clustermodel.ConditionTypeReady && condition.Status == "True" {
				readyNodes++
				break
			}
		}
	}

	// 获取 Pod 列表
	pods, err := client.Clientset().CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to list pods: %w", err))
	}

	runningPods := 0
	for _, pod := range pods.Items {
		if pod.Status.Phase == "Running" {
			runningPods++
		}
	}

	// 计算状态
	status := clustermodel.StatusHealthy
	if readyNodes < len(nodes.Items) {
		status = clustermodel.StatusDegraded
	}
	if readyNodes == 0 {
		status = clustermodel.StatusUnhealthy
	}

	return &clustermodel.ClusterHealth{
		ClusterID:   clusterID,
		Status:      status,
		TotalNodes:  len(nodes.Items),
		ReadyNodes:  readyNodes,
		TotalPods:   len(pods.Items),
		RunningPods: runningPods,
		CheckedAt:   time.Now(),
	}, nil
}

// GetPods 获取 Pod 列表
func (s *ClusterService) GetPods(ctx context.Context, clusterID, namespace string) ([]*clustermodel.Pod, error) {
	client, err := s.GetClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	pods, err := client.Clientset().CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to list pods: %w", err))
	}

	result := make([]*clustermodel.Pod, 0, len(pods.Items))
	for _, pod := range pods.Items {
		containers := make([]clustermodel.Container, 0, len(pod.Status.ContainerStatuses))
		for _, cs := range pod.Status.ContainerStatuses {
			state := clustermodel.StatusUnknown
			if cs.State.Running != nil {
				state = "running"
			} else if cs.State.Waiting != nil {
				state = "waiting"
			} else if cs.State.Terminated != nil {
				state = "terminated"
			}

			containers = append(containers, clustermodel.Container{
				Name:         cs.Name,
				Image:        cs.Image,
				Ready:        cs.Ready,
				RestartCount: cs.RestartCount,
				State:        state,
			})
		}

		result = append(result, &clustermodel.Pod{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Status:     string(pod.Status.Phase),
			Phase:      string(pod.Status.Phase),
			NodeName:   pod.Spec.NodeName,
			PodIP:      pod.Status.PodIP,
			Labels:     pod.Labels,
			Containers: containers,
			CreatedAt:  pod.CreationTimestamp.Time,
		})
	}

	return result, nil
}

// GetClusterOptions 获取集群选择器列表（用于下拉框）
func (s *ClusterService) GetClusterOptions(ctx context.Context) ([]*clustermodel.ClusterOption, error) {
	var clusters []*clustermodel.Cluster
	if err := s.db.WithContext(ctx).
		Select("id, name").
		Order("name ASC").
		Find(&clusters).Error; err != nil {
		return nil, errors.NewDatabaseError(fmt.Errorf("failed to query cluster options: %w", err))
	}

	options := make([]*clustermodel.ClusterOption, 0, len(clusters))
	for _, c := range clusters {
		options = append(options, &clustermodel.ClusterOption{
			Label: c.Name,
			Value: c.ID,
		})
	}

	return options, nil
}

// ========== 内部辅助方法 ==========

// populateClusterStats 填充集群统计信息（NodeCount, PodCount, NamespaceCount）
func (s *ClusterService) populateClusterStats(ctx context.Context, cluster *clustermodel.Cluster) error {
	client, err := s.GetClient(ctx, cluster.ID)
	if err != nil {
		return err
	}

	// 获取节点数量
	nodes, err := client.Clientset().CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}
	cluster.NodeCount = len(nodes.Items)

	// 获取命名空间数量
	namespaces, err := client.Clientset().CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list namespaces: %w", err)
	}
	cluster.NamespaceCount = len(namespaces.Items)

	// 获取 Pod 数量
	pods, err := client.Clientset().CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}
	cluster.PodCount = len(pods.Items)

	return nil
}

// GetClient 获取或创建 K8s 客户端（带缓存）
// 公开此方法以供其他 K8s 服务使用
func (s *ClusterService) GetClient(ctx context.Context, clusterID string) (*k8s.Client, error) {
	// 1. 尝试从缓存获取
	if client, ok := s.clients[clusterID]; ok {
		return client, nil
	}

	// 2. 从数据库加载 kubeconfig
	var cluster clustermodel.Cluster
	if err := s.db.WithContext(ctx).
		Select("kubeconfig").
		Where("id = ?", clusterID).
		First(&cluster).Error; err != nil {
		if stderr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrClusterNotFound
		}
		return nil, errors.NewDatabaseError(fmt.Errorf("failed to query cluster: %w", err))
	}

	// 3. 创建客户端
	client, err := k8s.NewClientFromKubeConfig([]byte(cluster.KubeConfig))
	if err != nil {
		return nil, errors.NewValidationError(fmt.Errorf("failed to create k8s client: %w", err))
	}

	// 4. 缓存客户端
	s.clients[clusterID] = client

	return client, nil
}

// getClient 提供小写版本以保持向后兼容性
// DEPRECATED: 使用 GetClient 替代
func (s *ClusterService) getClient(ctx context.Context, clusterID string) (*k8s.Client, error) {
	return s.GetClient(ctx, clusterID)
}

// ========== 请求/响应类型 ==========

// CreateClusterRequest 创建集群请求
type CreateClusterRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint" binding:"required"`
	KubeConfig  string `json:"kubeconfig" binding:"required"`
	Region      string `json:"region"`
	Provider    string `json:"provider"`
}

// UpdateClusterRequest 更新集群请求
type UpdateClusterRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
