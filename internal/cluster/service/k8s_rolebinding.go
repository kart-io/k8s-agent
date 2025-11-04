package service

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
)

// K8sRoleBindingService RoleBinding 管理服务.
type K8sRoleBindingService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sRoleBindingService 创建新的 RoleBinding 服务.
func NewK8sRoleBindingService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sRoleBindingService {
	return &K8sRoleBindingService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// RoleBindingInfo RoleBinding 信息.
type RoleBindingInfo struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	RoleRef      string            `json:"roleRef"`
	SubjectCount int               `json:"subjectCount"`
	Labels       map[string]string `json:"labels,omitempty"`
	CreatedAt    string            `json:"createdAt"`
}

// ListRoleBindings 获取 RoleBinding 列表.
func (s *K8sRoleBindingService) ListRoleBindings(ctx context.Context, clusterID, namespace string, offset, limit int) ([]RoleBindingInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	roleBindings, err := client.Clientset().RbacV1().RoleBindings(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list rolebindings: %w", err))
	}

	total := int64(len(roleBindings.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(roleBindings.Items) {
		start = len(roleBindings.Items)
	}
	if end > len(roleBindings.Items) {
		end = len(roleBindings.Items)
	}

	pagedRoleBindings := roleBindings.Items[start:end]

	// 转换为 RoleBindingInfo
	result := make([]RoleBindingInfo, 0, len(pagedRoleBindings))
	for i := range pagedRoleBindings {
		result = append(result, convertToRoleBindingInfo(&pagedRoleBindings[i]))
	}

	return result, total, nil
}

// GetRoleBinding 获取单个 RoleBinding 详情.
func (s *K8sRoleBindingService) GetRoleBinding(ctx context.Context, clusterID, namespace, name string) (*rbacv1.RoleBinding, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	rb, err := client.Clientset().RbacV1().RoleBindings(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get rolebinding: %w", err))
	}

	return rb, nil
}

// DeleteRoleBinding 删除 RoleBinding.
func (s *K8sRoleBindingService) DeleteRoleBinding(ctx context.Context, clusterID, namespace, name string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().RbacV1().RoleBindings(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete rolebinding: %w", err))
	}

	return nil
}

// convertToRoleBindingInfo 转换 RoleBinding 为 RoleBindingInfo.
func convertToRoleBindingInfo(rb *rbacv1.RoleBinding) RoleBindingInfo {
	roleRef := fmt.Sprintf("%s/%s", rb.RoleRef.Kind, rb.RoleRef.Name)

	return RoleBindingInfo{
		Name:         rb.Name,
		Namespace:    rb.Namespace,
		RoleRef:      roleRef,
		SubjectCount: len(rb.Subjects),
		Labels:       rb.Labels,
		CreatedAt:    rb.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}
