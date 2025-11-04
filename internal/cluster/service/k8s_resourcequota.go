package service

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/logger"
)

// K8sResourceQuotaService ResourceQuota 管理服务
type K8sResourceQuotaService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sResourceQuotaService 创建新的 ResourceQuota 服务
func NewK8sResourceQuotaService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sResourceQuotaService {
	return &K8sResourceQuotaService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// ResourceQuotaInfo ResourceQuota 信息
type ResourceQuotaInfo struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Hard        map[string]string `json:"hard"`
	Used        map[string]string `json:"used"`
	Scopes      []string          `json:"scopes,omitempty"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   string            `json:"createdAt"`
}

// ListResourceQuotas 获取 ResourceQuota 列表
func (s *K8sResourceQuotaService) ListResourceQuotas(ctx context.Context, clusterID, namespace string, offset, limit int) ([]ResourceQuotaInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	resourceQuotas, err := client.Clientset().CoreV1().ResourceQuotas(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list resourcequotas: %w", err))
	}

	total := int64(len(resourceQuotas.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(resourceQuotas.Items) {
		start = len(resourceQuotas.Items)
	}
	if end > len(resourceQuotas.Items) {
		end = len(resourceQuotas.Items)
	}

	result := make([]ResourceQuotaInfo, 0)
	for i := start; i < end; i++ {
		rq := resourceQuotas.Items[i]
		result = append(result, s.convertResourceQuotaInfo(&rq))
	}

	return result, total, nil
}

// GetResourceQuota 获取 ResourceQuota 详情
func (s *K8sResourceQuotaService) GetResourceQuota(ctx context.Context, clusterID, namespace, resourceQuotaName string) (*corev1.ResourceQuota, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	resourceQuota, err := client.Clientset().CoreV1().ResourceQuotas(namespace).Get(ctx, resourceQuotaName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get resourcequota: %w", err))
	}

	return resourceQuota, nil
}

// CreateResourceQuota 创建 ResourceQuota
func (s *K8sResourceQuotaService) CreateResourceQuota(ctx context.Context, clusterID, namespace string, resourceQuota *corev1.ResourceQuota) (*corev1.ResourceQuota, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	createdResourceQuota, err := client.Clientset().CoreV1().ResourceQuotas(namespace).Create(ctx, resourceQuota, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to create resourcequota: %w", err))
	}

	logger.Infow("ResourceQuota created",
		"cluster_id", clusterID,
		"namespace", namespace,
		"resourcequota", createdResourceQuota.Name,
	)

	return createdResourceQuota, nil
}

// UpdateResourceQuota 更新 ResourceQuota
func (s *K8sResourceQuotaService) UpdateResourceQuota(ctx context.Context, clusterID, namespace string, resourceQuota *corev1.ResourceQuota) (*corev1.ResourceQuota, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	updatedResourceQuota, err := client.Clientset().CoreV1().ResourceQuotas(namespace).Update(ctx, resourceQuota, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to update resourcequota: %w", err))
	}

	logger.Infow("ResourceQuota updated",
		"cluster_id", clusterID,
		"namespace", namespace,
		"resourcequota", updatedResourceQuota.Name,
	)

	return updatedResourceQuota, nil
}

// DeleteResourceQuota 删除 ResourceQuota
func (s *K8sResourceQuotaService) DeleteResourceQuota(ctx context.Context, clusterID, namespace, resourceQuotaName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().CoreV1().ResourceQuotas(namespace).Delete(ctx, resourceQuotaName, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete resourcequota: %w", err))
	}

	logger.Infow("ResourceQuota deleted",
		"cluster_id", clusterID,
		"namespace", namespace,
		"resourcequota", resourceQuotaName,
	)

	return nil
}

// convertResourceQuotaInfo 转换 ResourceQuota 信息
func (s *K8sResourceQuotaService) convertResourceQuotaInfo(rq *corev1.ResourceQuota) ResourceQuotaInfo {
	// 确保 Labels 和 Annotations 不为 nil
	labels := rq.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	annotations := rq.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}

	// 转换 Hard 限制
	hard := make(map[string]string)
	for k, v := range rq.Spec.Hard {
		hard[string(k)] = v.String()
	}

	// 转换 Used 使用量
	used := make(map[string]string)
	for k, v := range rq.Status.Used {
		used[string(k)] = v.String()
	}

	// 转换 Scopes
	scopes := make([]string, 0, len(rq.Spec.Scopes))
	for _, scope := range rq.Spec.Scopes {
		scopes = append(scopes, string(scope))
	}

	return ResourceQuotaInfo{
		Name:        rq.Name,
		Namespace:   rq.Namespace,
		Hard:        hard,
		Used:        used,
		Scopes:      scopes,
		Labels:      labels,
		Annotations: annotations,
		CreatedAt:   rq.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}
