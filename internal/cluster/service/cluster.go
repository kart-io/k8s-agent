package service

import (
	"context"
	"fmt"
	"time"

	"github.com/kart-io/k8s-agent/internal/cluster/k8s"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/k8s-agent/internal/cluster/types"
	"github.com/kart-io/logger/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterService struct {
	storage *storage.MySQLStorage
	clients map[string]*k8s.Client // cluster_id -> client
	log     core.Logger
}

func NewClusterService(storage *storage.MySQLStorage, logger core.Logger) *ClusterService {
	return &ClusterService{
		storage: storage,
		clients: make(map[string]*k8s.Client),
		log:     logger,
	}
}

// AddCluster 添加集群
func (s *ClusterService) AddCluster(ctx context.Context, cluster *types.Cluster) error {
	// 测试连接
	client, err := k8s.NewClientFromKubeConfig([]byte(cluster.KubeConfig))
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}

	if err := client.CheckConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}

	// 获取集群版本
	version, err := client.GetServerVersion(ctx)
	if err != nil {
		s.log.Warnw("Failed to get server version", "error", err)
	} else {
		cluster.Version = version
	}

	cluster.Status = "healthy"
	cluster.CreatedAt = time.Now()
	cluster.UpdatedAt = time.Now()

	// 保存到数据库
	query := `
		INSERT INTO clusters (id, name, description, endpoint, version, status, region, provider, kubeconfig, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = s.storage.DB().ExecContext(ctx, query,
		cluster.ID, cluster.Name, cluster.Description, cluster.Endpoint,
		cluster.Version, cluster.Status, cluster.Region, cluster.Provider,
		cluster.KubeConfig, cluster.CreatedAt, cluster.UpdatedAt,
	)

	if err != nil {
		return err
	}

	// 缓存客户端
	s.clients[cluster.ID] = client

	return nil
}

// GetClusterHealth 获取集群健康状态
func (s *ClusterService) GetClusterHealth(ctx context.Context, clusterID string) (*types.ClusterHealth, error) {
	client, err := s.getClient(clusterID)
	if err != nil {
		return nil, err
	}

	// 获取节点列表
	nodes, err := client.Clientset().CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
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
		return nil, fmt.Errorf("failed to list pods: %w", err)
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

	return &types.ClusterHealth{
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
func (s *ClusterService) GetPods(ctx context.Context, clusterID, namespace string) ([]types.Pod, error) {
	client, err := s.getClient(clusterID)
	if err != nil {
		return nil, err
	}

	pods, err := client.Clientset().CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]types.Pod, 0, len(pods.Items))
	for _, pod := range pods.Items {
		containers := make([]types.Container, 0, len(pod.Status.ContainerStatuses))
		for _, cs := range pod.Status.ContainerStatuses {
			state := "unknown"
			if cs.State.Running != nil {
				state = "running"
			} else if cs.State.Waiting != nil {
				state = "waiting"
			} else if cs.State.Terminated != nil {
				state = "terminated"
			}

			containers = append(containers, types.Container{
				Name:         cs.Name,
				Image:        cs.Image,
				Ready:        cs.Ready,
				RestartCount: cs.RestartCount,
				State:        state,
			})
		}

		result = append(result, types.Pod{
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

func (s *ClusterService) getClient(clusterID string) (*k8s.Client, error) {
	// 先从缓存获取
	if client, ok := s.clients[clusterID]; ok {
		return client, nil
	}

	// 从数据库加载
	var kubeconfigData string
	query := "SELECT kubeconfig FROM clusters WHERE id = ?"
	err := s.storage.DB().QueryRow(query, clusterID).Scan(&kubeconfigData)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	client, err := k8s.NewClientFromKubeConfig([]byte(kubeconfigData))
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	s.clients[clusterID] = client
	return client, nil
}
