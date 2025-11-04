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

// K8sServiceAccountService ServiceAccount 管理服务.
type K8sServiceAccountService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sServiceAccountService 创建新的 ServiceAccount 服务.
func NewK8sServiceAccountService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sServiceAccountService {
	return &K8sServiceAccountService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// ServiceAccountInfo ServiceAccount 信息.
type ServiceAccountInfo struct {
	Name                    string            `json:"name"`
	Namespace               string            `json:"namespace"`
	Secrets                 int               `json:"secrets"`
	ImagePullSecrets        int               `json:"imagePullSecrets"`
	AutomountServiceAccount *bool             `json:"automountServiceAccountToken,omitempty"`
	Labels                  map[string]string `json:"labels"`
	Annotations             map[string]string `json:"annotations"`
	CreatedAt               string            `json:"createdAt"`
}

// ListServiceAccounts 获取 ServiceAccount 列表.
func (s *K8sServiceAccountService) ListServiceAccounts(ctx context.Context, clusterID, namespace string, offset, limit int) ([]ServiceAccountInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	serviceAccounts, err := client.Clientset().CoreV1().ServiceAccounts(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list serviceaccounts: %w", err))
	}

	total := int64(len(serviceAccounts.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(serviceAccounts.Items) {
		start = len(serviceAccounts.Items)
	}
	if end > len(serviceAccounts.Items) {
		end = len(serviceAccounts.Items)
	}

	result := make([]ServiceAccountInfo, 0)
	for i := start; i < end; i++ {
		sa := serviceAccounts.Items[i]
		result = append(result, s.convertServiceAccountInfo(&sa))
	}

	return result, total, nil
}

// GetServiceAccount 获取 ServiceAccount 详情.
func (s *K8sServiceAccountService) GetServiceAccount(ctx context.Context, clusterID, namespace, serviceAccountName string) (*corev1.ServiceAccount, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	serviceAccount, err := client.Clientset().CoreV1().ServiceAccounts(namespace).Get(ctx, serviceAccountName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get serviceaccount: %w", err))
	}

	return serviceAccount, nil
}

// CreateServiceAccount 创建 ServiceAccount.
func (s *K8sServiceAccountService) CreateServiceAccount(ctx context.Context, clusterID, namespace string, serviceAccount *corev1.ServiceAccount) (*corev1.ServiceAccount, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	createdServiceAccount, err := client.Clientset().CoreV1().ServiceAccounts(namespace).Create(ctx, serviceAccount, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to create serviceaccount: %w", err))
	}

	logger.Infow("ServiceAccount created",
		"cluster_id", clusterID,
		"namespace", namespace,
		"serviceaccount", createdServiceAccount.Name,
	)

	return createdServiceAccount, nil
}

// UpdateServiceAccount 更新 ServiceAccount.
func (s *K8sServiceAccountService) UpdateServiceAccount(ctx context.Context, clusterID, namespace string, serviceAccount *corev1.ServiceAccount) (*corev1.ServiceAccount, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	updatedServiceAccount, err := client.Clientset().CoreV1().ServiceAccounts(namespace).Update(ctx, serviceAccount, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to update serviceaccount: %w", err))
	}

	logger.Infow("ServiceAccount updated",
		"cluster_id", clusterID,
		"namespace", namespace,
		"serviceaccount", updatedServiceAccount.Name,
	)

	return updatedServiceAccount, nil
}

// DeleteServiceAccount 删除 ServiceAccount.
func (s *K8sServiceAccountService) DeleteServiceAccount(ctx context.Context, clusterID, namespace, serviceAccountName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().CoreV1().ServiceAccounts(namespace).Delete(ctx, serviceAccountName, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete serviceaccount: %w", err))
	}

	logger.Infow("ServiceAccount deleted",
		"cluster_id", clusterID,
		"namespace", namespace,
		"serviceaccount", serviceAccountName,
	)

	return nil
}

// convertServiceAccountInfo 转换 ServiceAccount 信息.
func (s *K8sServiceAccountService) convertServiceAccountInfo(sa *corev1.ServiceAccount) ServiceAccountInfo {
	// 确保 Labels 和 Annotations 不为 nil
	labels := sa.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	annotations := sa.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}

	return ServiceAccountInfo{
		Name:                    sa.Name,
		Namespace:               sa.Namespace,
		Secrets:                 len(sa.Secrets),
		ImagePullSecrets:        len(sa.ImagePullSecrets),
		AutomountServiceAccount: sa.AutomountServiceAccountToken,
		Labels:                  labels,
		Annotations:             annotations,
		CreatedAt:               sa.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}
