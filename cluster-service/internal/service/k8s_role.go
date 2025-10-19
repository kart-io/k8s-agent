package service

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cluster-service/internal/storage"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sRoleService provides Role-related operations
type K8sRoleService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sRoleService creates a new K8sRoleService instance
func NewK8sRoleService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sRoleService {
	return &K8sRoleService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// RoleInfo represents a simplified Role object
type RoleInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	RuleCount int               `json:"ruleCount"`
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt string            `json:"createdAt"`
}

// ListRoles lists all Roles in the specified namespace
func (s *K8sRoleService) ListRoles(ctx context.Context, clusterID, namespace string, offset, limit int) ([]RoleInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get client: %w", err)
	}

	// List Roles
	var roleList *v1.RoleList
	if namespace == "" || namespace == "-" {
		// List all namespaces
		roleList, err = client.Clientset().RbacV1().Roles("").List(ctx, metav1.ListOptions{})
	} else {
		roleList, err = client.Clientset().RbacV1().Roles(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list roles: %w", err)
	}

	total := int64(len(roleList.Items))

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(roleList.Items) {
		start = len(roleList.Items)
	}
	if end > len(roleList.Items) {
		end = len(roleList.Items)
	}

	// Convert to RoleInfo
	roles := make([]RoleInfo, 0, end-start)
	for i := start; i < end; i++ {
		role := roleList.Items[i]

		roleInfo := RoleInfo{
			Name:      role.Name,
			Namespace: role.Namespace,
			RuleCount: len(role.Rules),
			Labels:    role.Labels,
			CreatedAt: role.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
		}
		roles = append(roles, roleInfo)
	}

	return roles, total, nil
}

// GetRole retrieves a specific Role
func (s *K8sRoleService) GetRole(ctx context.Context, clusterID, namespace, name string) (*RoleInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	role, err := client.Clientset().RbacV1().Roles(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	roleInfo := &RoleInfo{
		Name:      role.Name,
		Namespace: role.Namespace,
		RuleCount: len(role.Rules),
		Labels:    role.Labels,
		CreatedAt: role.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}

	return roleInfo, nil
}

// DeleteRole deletes a Role
func (s *K8sRoleService) DeleteRole(ctx context.Context, clusterID, namespace, name string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	err = client.Clientset().RbacV1().Roles(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}
