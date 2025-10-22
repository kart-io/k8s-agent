package service

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/common/logger"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sClusterRoleBindingService ClusterRoleBinding 管理服务
type K8sClusterRoleBindingService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sClusterRoleBindingService 创建新的 ClusterRoleBinding 服务
func NewK8sClusterRoleBindingService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sClusterRoleBindingService {
	return &K8sClusterRoleBindingService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// ClusterRoleBindingInfo ClusterRoleBinding 信息
type ClusterRoleBindingInfo struct {
	Name        string            `json:"name"`
	RoleRef     RoleRefInfo       `json:"roleRef"`
	Subjects    []SubjectInfo     `json:"subjects"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   string            `json:"createdAt"`
}

// RoleRefInfo RoleRef 信息
type RoleRefInfo struct {
	APIGroup string `json:"apiGroup"`
	Kind     string `json:"kind"`
	Name     string `json:"name"`
}

// SubjectInfo Subject 信息
type SubjectInfo struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	APIGroup  string `json:"apiGroup,omitempty"`
}

// ListClusterRoleBindings 获取 ClusterRoleBinding 列表
func (s *K8sClusterRoleBindingService) ListClusterRoleBindings(ctx context.Context, clusterID string, offset, limit int) ([]ClusterRoleBindingInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	clusterRoleBindings, err := client.Clientset().RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list clusterrolebindings: %w", err))
	}

	total := int64(len(clusterRoleBindings.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(clusterRoleBindings.Items) {
		start = len(clusterRoleBindings.Items)
	}
	if end > len(clusterRoleBindings.Items) {
		end = len(clusterRoleBindings.Items)
	}

	result := make([]ClusterRoleBindingInfo, 0)
	for i := start; i < end; i++ {
		crb := clusterRoleBindings.Items[i]
		result = append(result, s.convertClusterRoleBindingInfo(&crb))
	}

	return result, total, nil
}

// GetClusterRoleBinding 获取 ClusterRoleBinding 详情
func (s *K8sClusterRoleBindingService) GetClusterRoleBinding(ctx context.Context, clusterID, clusterRoleBindingName string) (*rbacv1.ClusterRoleBinding, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	clusterRoleBinding, err := client.Clientset().RbacV1().ClusterRoleBindings().Get(ctx, clusterRoleBindingName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get clusterrolebinding: %w", err))
	}

	return clusterRoleBinding, nil
}

// CreateClusterRoleBinding 创建 ClusterRoleBinding
func (s *K8sClusterRoleBindingService) CreateClusterRoleBinding(ctx context.Context, clusterID string, clusterRoleBinding *rbacv1.ClusterRoleBinding) (*rbacv1.ClusterRoleBinding, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	createdClusterRoleBinding, err := client.Clientset().RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBinding, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to create clusterrolebinding: %w", err))
	}

	logger.Infow("ClusterRoleBinding created",
		"cluster_id", clusterID,
		"clusterrolebinding", createdClusterRoleBinding.Name,
	)

	return createdClusterRoleBinding, nil
}

// UpdateClusterRoleBinding 更新 ClusterRoleBinding
func (s *K8sClusterRoleBindingService) UpdateClusterRoleBinding(ctx context.Context, clusterID string, clusterRoleBinding *rbacv1.ClusterRoleBinding) (*rbacv1.ClusterRoleBinding, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	updatedClusterRoleBinding, err := client.Clientset().RbacV1().ClusterRoleBindings().Update(ctx, clusterRoleBinding, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to update clusterrolebinding: %w", err))
	}

	logger.Infow("ClusterRoleBinding updated",
		"cluster_id", clusterID,
		"clusterrolebinding", updatedClusterRoleBinding.Name,
	)

	return updatedClusterRoleBinding, nil
}

// DeleteClusterRoleBinding 删除 ClusterRoleBinding
func (s *K8sClusterRoleBindingService) DeleteClusterRoleBinding(ctx context.Context, clusterID, clusterRoleBindingName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().RbacV1().ClusterRoleBindings().Delete(ctx, clusterRoleBindingName, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete clusterrolebinding: %w", err))
	}

	logger.Infow("ClusterRoleBinding deleted",
		"cluster_id", clusterID,
		"clusterrolebinding", clusterRoleBindingName,
	)

	return nil
}

// convertClusterRoleBindingInfo 转换 ClusterRoleBinding 信息
func (s *K8sClusterRoleBindingService) convertClusterRoleBindingInfo(crb *rbacv1.ClusterRoleBinding) ClusterRoleBindingInfo {
	// 确保 Labels 和 Annotations 不为 nil
	labels := crb.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	annotations := crb.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}

	// 转换 RoleRef
	roleRef := RoleRefInfo{
		APIGroup: crb.RoleRef.APIGroup,
		Kind:     crb.RoleRef.Kind,
		Name:     crb.RoleRef.Name,
	}

	// 转换 Subjects
	subjects := make([]SubjectInfo, 0, len(crb.Subjects))
	for _, subject := range crb.Subjects {
		subjects = append(subjects, SubjectInfo{
			Kind:      subject.Kind,
			Name:      subject.Name,
			Namespace: subject.Namespace,
			APIGroup:  subject.APIGroup,
		})
	}

	return ClusterRoleBindingInfo{
		Name:        crb.Name,
		RoleRef:     roleRef,
		Subjects:    subjects,
		Labels:      labels,
		Annotations: annotations,
		CreatedAt:   crb.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}
