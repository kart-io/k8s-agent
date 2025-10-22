package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kart-io/k8s-agent/internal/cluster/k8s"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/common/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sClusterService 集群管理服务
type K8sClusterService struct {
	storage *storage.MySQLStorage
	clients map[string]*k8s.Client // cluster_id -> client
}

// NewK8sClusterService 创建新的集群服务
func NewK8sClusterService(storage *storage.MySQLStorage) *K8sClusterService {
	return &K8sClusterService{
		storage: storage,
		clients: make(map[string]*k8s.Client),
	}
}

// ClusterInfo 集群信息
type ClusterInfo struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	Endpoint       string            `json:"endpoint"`
	Version        string            `json:"version"`
	Status         string            `json:"status"`
	Region         string            `json:"region"`
	Provider       string            `json:"provider"`
	Labels         map[string]string `json:"labels"`
	KubeConfig     string            `json:"kubeconfig,omitempty"`
	NodeCount      int               `json:"nodeCount,omitempty"`
	PodCount       int               `json:"podCount,omitempty"`
	NamespaceCount int               `json:"namespaceCount,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
}

// ClusterHealth 集群健康状态
type ClusterHealth struct {
	ClusterID   string    `json:"clusterId"`
	Status      string    `json:"status"`
	TotalNodes  int       `json:"totalNodes"`
	ReadyNodes  int       `json:"readyNodes"`
	TotalPods   int       `json:"totalPods"`
	RunningPods int       `json:"runningPods"`
	CheckedAt   time.Time `json:"checkedAt"`
}

// ListClusters 获取集群列表
func (s *K8sClusterService) ListClusters(ctx context.Context, offset, limit int) ([]ClusterInfo, int64, error) {
	// 查询总数
	var total int64
	countQuery := "SELECT COUNT(*) FROM clusters"
	if err := s.storage.DB().QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, errors.NewDatabaseError(fmt.Errorf("failed to count clusters: %w", err))
	}

	// 查询集群列表
	query := `
		SELECT id, name, description, endpoint, version, status, region, provider, created_at, updated_at
		FROM clusters
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.storage.DB().QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, errors.NewDatabaseError(fmt.Errorf("failed to query clusters: %w", err))
	}
	defer rows.Close()

	clusters := make([]ClusterInfo, 0)
	for rows.Next() {
		var cluster ClusterInfo
		if err := rows.Scan(
			&cluster.ID,
			&cluster.Name,
			&cluster.Description,
			&cluster.Endpoint,
			&cluster.Version,
			&cluster.Status,
			&cluster.Region,
			&cluster.Provider,
			&cluster.CreatedAt,
			&cluster.UpdatedAt,
		); err != nil {
			logger.Errorw("Failed to scan cluster row", "error", err.Error())
			continue
		}
		clusters = append(clusters, cluster)
	}

	// 为每个集群填充统计信息
	for i := range clusters {
		if err := s.populateClusterStats(ctx, &clusters[i]); err != nil {
			logger.Warnw("Failed to populate cluster stats for list",
				"cluster_id", clusters[i].ID,
				"error", err.Error())
			// 统计信息获取失败不影响列表的返回，保持字段为零值
		}
	}

	return clusters, total, nil
}

// GetCluster 获取集群详情
func (s *K8sClusterService) GetCluster(ctx context.Context, clusterID string) (*ClusterInfo, error) {
	query := `
		SELECT id, name, description, endpoint, version, status, region, provider, kubeconfig, created_at, updated_at
		FROM clusters
		WHERE id = ?
	`

	var cluster ClusterInfo
	err := s.storage.DB().QueryRowContext(ctx, query, clusterID).Scan(
		&cluster.ID,
		&cluster.Name,
		&cluster.Description,
		&cluster.Endpoint,
		&cluster.Version,
		&cluster.Status,
		&cluster.Region,
		&cluster.Provider,
		&cluster.KubeConfig,
		&cluster.CreatedAt,
		&cluster.UpdatedAt,
	)

	if err != nil {
		return nil, errors.ErrClusterNotFound
	}

	// 获取集群统计信息
	if err := s.populateClusterStats(ctx, &cluster); err != nil {
		logger.Warnw("Failed to populate cluster stats", "cluster_id", clusterID, "error", err.Error())
		// 统计信息获取失败不影响集群详情的返回
	}

	return &cluster, nil
}

// populateClusterStats 填充集群统计信息
func (s *K8sClusterService) populateClusterStats(ctx context.Context, cluster *ClusterInfo) error {
	client, err := s.getClient(ctx, cluster.ID)
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

	// 获取 Pod 数量（所有命名空间）
	pods, err := client.Clientset().CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}
	cluster.PodCount = len(pods.Items)

	return nil
}

// CreateCluster 创建集群
func (s *K8sClusterService) CreateCluster(
	ctx context.Context,
	name, description, endpoint, kubeconfig, region, provider string,
	labels map[string]string,
) (*ClusterInfo, error) {
	// 测试连接
	client, err := k8s.NewClientFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return nil, errors.NewValidationError(fmt.Errorf("invalid kubeconfig: %w", err))
	}

	if err := client.CheckConnection(ctx); err != nil {
		return nil, errors.NewValidationError(fmt.Errorf("failed to connect to cluster: %w", err))
	}

	// 获取集群版本
	version, err := client.GetServerVersion(ctx)
	if err != nil {
		logger.Warnw("Failed to get server version", "error", err.Error())
		version = "unknown"
	}

	// 创建集群记录
	clusterID := uuid.New().String()
	now := time.Now()

	cluster := &ClusterInfo{
		ID:          clusterID,
		Name:        name,
		Description: description,
		Endpoint:    endpoint,
		Version:     version,
		Status:      "healthy",
		Region:      region,
		Provider:    provider,
		Labels:      labels,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	query := `
		INSERT INTO clusters (id, name, description, endpoint, version, status, region, provider, kubeconfig, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.storage.DB().ExecContext(ctx, query,
		cluster.ID, cluster.Name, cluster.Description, cluster.Endpoint,
		cluster.Version, cluster.Status, cluster.Region, cluster.Provider,
		kubeconfig, cluster.CreatedAt, cluster.UpdatedAt,
	)

	if err != nil {
		return nil, errors.NewDatabaseError(fmt.Errorf("failed to create cluster: %w", err))
	}

	// 缓存客户端
	s.clients[clusterID] = client

	logger.Infow("Cluster created",
		"cluster_id", clusterID,
		"name", name,
		"version", version,
	)

	return cluster, nil
}

// UpdateCluster 更新集群信息
func (s *K8sClusterService) UpdateCluster(
	ctx context.Context,
	clusterID, name, description string,
	labels map[string]string,
) (*ClusterInfo, error) {
	// 检查集群是否存在
	_, err := s.GetCluster(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	query := `
		UPDATE clusters
		SET name = ?, description = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = s.storage.DB().ExecContext(ctx, query, name, description, now, clusterID)
	if err != nil {
		return nil, errors.NewDatabaseError(fmt.Errorf("failed to update cluster: %w", err))
	}

	logger.Infow("Cluster updated", "cluster_id", clusterID)

	return s.GetCluster(ctx, clusterID)
}

// DeleteCluster 删除集群
func (s *K8sClusterService) DeleteCluster(ctx context.Context, clusterID string) error {
	query := "DELETE FROM clusters WHERE id = ?"
	result, err := s.storage.DB().ExecContext(ctx, query, clusterID)
	if err != nil {
		return errors.NewDatabaseError(fmt.Errorf("failed to delete cluster: %w", err))
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.ErrClusterNotFound
	}

	// 清除缓存的客户端
	delete(s.clients, clusterID)

	logger.Infow("Cluster deleted", "cluster_id", clusterID)

	return nil
}

// GetClusterHealth 获取集群健康状态
func (s *K8sClusterService) GetClusterHealth(ctx context.Context, clusterID string) (*ClusterHealth, error) {
	client, err := s.getClient(ctx, clusterID)
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
			if condition.Type == "Ready" && condition.Status == "True" {
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

	status := "healthy"
	if readyNodes < len(nodes.Items) {
		status = "degraded"
	}
	if readyNodes == 0 {
		status = "unhealthy"
	}

	return &ClusterHealth{
		ClusterID:   clusterID,
		Status:      status,
		TotalNodes:  len(nodes.Items),
		ReadyNodes:  readyNodes,
		TotalPods:   len(pods.Items),
		RunningPods: runningPods,
		CheckedAt:   time.Now(),
	}, nil
}

// ClusterOption 集群选择器选项
type ClusterOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// GetClusterOptions 获取集群选择器列表
func (s *K8sClusterService) GetClusterOptions(ctx context.Context) ([]ClusterOption, error) {
	query := `
		SELECT id, name
		FROM clusters
		ORDER BY name ASC
	`

	rows, err := s.storage.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, errors.NewDatabaseError(fmt.Errorf("failed to query cluster options: %w", err))
	}
	defer rows.Close()

	options := make([]ClusterOption, 0)
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			logger.Errorw("Failed to scan cluster option row", "error", err.Error())
			continue
		}
		options = append(options, ClusterOption{
			Label: name,
			Value: id,
		})
	}

	return options, nil
}

// getClient 获取集群客户端
func (s *K8sClusterService) getClient(ctx context.Context, clusterID string) (*k8s.Client, error) {
	// 先从缓存获取
	if client, ok := s.clients[clusterID]; ok {
		return client, nil
	}

	// 从数据库加载
	var kubeconfigData string
	query := "SELECT kubeconfig FROM clusters WHERE id = ?"
	err := s.storage.DB().QueryRowContext(ctx, query, clusterID).Scan(&kubeconfigData)
	if err != nil {
		return nil, errors.ErrClusterNotFound
	}

	client, err := k8s.NewClientFromKubeConfig([]byte(kubeconfigData))
	if err != nil {
		return nil, errors.NewValidationError(fmt.Errorf("failed to create k8s client: %w", err))
	}

	// 缓存客户端
	s.clients[clusterID] = client

	return client, nil
}
