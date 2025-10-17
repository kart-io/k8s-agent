package service

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cluster-service/internal/storage"
	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/common/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sNamespaceService 命名空间管理服务
type K8sNamespaceService struct {
	storage       *storage.PostgresStorage
	clusterService *K8sClusterService
}

// NewK8sNamespaceService 创建新的命名空间服务
func NewK8sNamespaceService(storage *storage.PostgresStorage, clusterService *K8sClusterService) *K8sNamespaceService {
	return &K8sNamespaceService{
		storage:       storage,
		clusterService: clusterService,
	}
}

// NamespaceInfo 命名空间信息
type NamespaceInfo struct {
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	Labels    map[string]string `json:"labels"`
	CreatedAt string            `json:"createdAt"`
}

// ListNamespaces 获取命名空间列表
func (s *K8sNamespaceService) ListNamespaces(ctx context.Context, clusterID string, offset, limit int) ([]NamespaceInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	namespaces, err := client.Clientset().CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list namespaces: %w", err))
	}

	total := int64(len(namespaces.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(namespaces.Items) {
		start = len(namespaces.Items)
	}
	if end > len(namespaces.Items) {
		end = len(namespaces.Items)
	}

	result := make([]NamespaceInfo, 0)
	for i := start; i < end; i++ {
		ns := namespaces.Items[i]
		result = append(result, NamespaceInfo{
			Name:      ns.Name,
			Status:    string(ns.Status.Phase),
			Labels:    ns.Labels,
			CreatedAt: ns.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
		})
	}

	return result, total, nil
}

// GetNamespace 获取命名空间详情
func (s *K8sNamespaceService) GetNamespace(ctx context.Context, clusterID, namespaceName string) (*NamespaceInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	namespace, err := client.Clientset().CoreV1().Namespaces().Get(ctx, namespaceName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get namespace: %w", err))
	}

	return &NamespaceInfo{
		Name:      namespace.Name,
		Status:    string(namespace.Status.Phase),
		Labels:    namespace.Labels,
		CreatedAt: namespace.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}, nil
}

// CreateNamespace 创建命名空间
func (s *K8sNamespaceService) CreateNamespace(ctx context.Context, clusterID, name string, labels map[string]string) (*NamespaceInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}

	created, err := client.Clientset().CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to create namespace: %w", err))
	}

	logger.Infow("Namespace created",
		"cluster_id", clusterID,
		"namespace", name,
	)

	return &NamespaceInfo{
		Name:      created.Name,
		Status:    string(created.Status.Phase),
		Labels:    created.Labels,
		CreatedAt: created.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}, nil
}

// DeleteNamespace 删除命名空间
func (s *K8sNamespaceService) DeleteNamespace(ctx context.Context, clusterID, namespaceName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().CoreV1().Namespaces().Delete(ctx, namespaceName, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete namespace: %w", err))
	}

	logger.Infow("Namespace deleted",
		"cluster_id", clusterID,
		"namespace", namespaceName,
	)

	return nil
}
