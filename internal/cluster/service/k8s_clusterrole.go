package service

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
)

// K8sClusterRoleService ClusterRole 管理服务.
type K8sClusterRoleService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sClusterRoleService 创建新的 ClusterRole 服务.
func NewK8sClusterRoleService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sClusterRoleService {
	return &K8sClusterRoleService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// ClusterRoleInfo ClusterRole 信息.
type ClusterRoleInfo struct {
	Name      string            `json:"name"`
	RuleCount int               `json:"ruleCount"`
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt string            `json:"createdAt"`
}

// ListClusterRoles 获取 ClusterRole 列表.
func (s *K8sClusterRoleService) ListClusterRoles(ctx context.Context, clusterID string, offset, limit int) ([]ClusterRoleInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	clusterRoles, err := client.Clientset().RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list clusterroles: %w", err))
	}

	total := int64(len(clusterRoles.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(clusterRoles.Items) {
		start = len(clusterRoles.Items)
	}
	if end > len(clusterRoles.Items) {
		end = len(clusterRoles.Items)
	}

	pagedClusterRoles := clusterRoles.Items[start:end]

	// 转换为 ClusterRoleInfo
	result := make([]ClusterRoleInfo, 0, len(pagedClusterRoles))
	for i := range pagedClusterRoles {
		result = append(result, convertToClusterRoleInfo(&pagedClusterRoles[i]))
	}

	return result, total, nil
}

// GetClusterRole 获取单个 ClusterRole 详情.
func (s *K8sClusterRoleService) GetClusterRole(ctx context.Context, clusterID, name string) (*rbacv1.ClusterRole, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	cr, err := client.Clientset().RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get clusterrole: %w", err))
	}

	return cr, nil
}

// DeleteClusterRole 删除 ClusterRole.
func (s *K8sClusterRoleService) DeleteClusterRole(ctx context.Context, clusterID, name string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().RbacV1().ClusterRoles().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete clusterrole: %w", err))
	}

	return nil
}

// convertToClusterRoleInfo 转换 ClusterRole 为 ClusterRoleInfo.
func convertToClusterRoleInfo(cr *rbacv1.ClusterRole) ClusterRoleInfo {
	return ClusterRoleInfo{
		Name:      cr.Name,
		RuleCount: len(cr.Rules),
		Labels:    cr.Labels,
		CreatedAt: cr.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}
