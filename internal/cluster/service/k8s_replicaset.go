package service

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/logger"
)

// K8sReplicaSetService ReplicaSet 管理服务.
type K8sReplicaSetService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sReplicaSetService 创建新的 ReplicaSet 服务.
func NewK8sReplicaSetService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sReplicaSetService {
	return &K8sReplicaSetService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// ReplicaSetInfo ReplicaSet 信息.
type ReplicaSetInfo struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int32             `json:"replicas"`
	ReadyReplicas     int32             `json:"readyReplicas"`
	AvailableReplicas int32             `json:"availableReplicas"`
	Labels            map[string]string `json:"labels"`
	Selector          map[string]string `json:"selector"`
	OwnerReferences   []OwnerReference  `json:"ownerReferences,omitempty"`
	CreatedAt         string            `json:"createdAt"`
}

// OwnerReference 所有者引用.
type OwnerReference struct {
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	UID        string `json:"uid"`
	Controller bool   `json:"controller"`
}

// ListReplicaSets 获取 ReplicaSet 列表.
func (s *K8sReplicaSetService) ListReplicaSets(ctx context.Context, clusterID, namespace string, offset, limit int) ([]ReplicaSetInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	replicaSets, err := client.Clientset().AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list replicasets: %w", err))
	}

	total := int64(len(replicaSets.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(replicaSets.Items) {
		start = len(replicaSets.Items)
	}
	if end > len(replicaSets.Items) {
		end = len(replicaSets.Items)
	}

	result := make([]ReplicaSetInfo, 0)
	for i := start; i < end; i++ {
		rs := replicaSets.Items[i]
		result = append(result, s.convertReplicaSetInfo(&rs))
	}

	return result, total, nil
}

// GetReplicaSet 获取 ReplicaSet 详情.
func (s *K8sReplicaSetService) GetReplicaSet(ctx context.Context, clusterID, namespace, replicaSetName string) (*ReplicaSetInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	replicaSet, err := client.Clientset().AppsV1().ReplicaSets(namespace).Get(ctx, replicaSetName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get replicaset: %w", err))
	}

	replicaSetInfo := s.convertReplicaSetInfo(replicaSet)
	return &replicaSetInfo, nil
}

// ScaleReplicaSet 扩缩容 ReplicaSet.
func (s *K8sReplicaSetService) ScaleReplicaSet(ctx context.Context, clusterID, namespace, replicaSetName string, replicas int32) (*ReplicaSetInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	// 获取当前 ReplicaSet
	replicaSet, err := client.Clientset().AppsV1().ReplicaSets(namespace).Get(ctx, replicaSetName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get replicaset: %w", err))
	}

	// 更新副本数
	replicaSet.Spec.Replicas = &replicas

	// 更新 ReplicaSet
	updatedReplicaSet, err := client.Clientset().AppsV1().ReplicaSets(namespace).Update(ctx, replicaSet, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to scale replicaset: %w", err))
	}

	logger.Infow("ReplicaSet scaled",
		"cluster_id", clusterID,
		"namespace", namespace,
		"replicaset", replicaSetName,
		"replicas", replicas,
	)

	replicaSetInfo := s.convertReplicaSetInfo(updatedReplicaSet)
	return &replicaSetInfo, nil
}

// DeleteReplicaSet 删除 ReplicaSet.
func (s *K8sReplicaSetService) DeleteReplicaSet(ctx context.Context, clusterID, namespace, replicaSetName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().AppsV1().ReplicaSets(namespace).Delete(ctx, replicaSetName, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete replicaset: %w", err))
	}

	logger.Infow("ReplicaSet deleted",
		"cluster_id", clusterID,
		"namespace", namespace,
		"replicaset", replicaSetName,
	)

	return nil
}

// convertReplicaSetInfo 转换 ReplicaSet 信息.
func (s *K8sReplicaSetService) convertReplicaSetInfo(rs *appsv1.ReplicaSet) ReplicaSetInfo {
	// 确保 Labels 和 Selector 不为 nil
	labels := rs.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	selector := rs.Spec.Selector.MatchLabels
	if selector == nil {
		selector = make(map[string]string)
	}

	replicaSetInfo := ReplicaSetInfo{
		Name:              rs.Name,
		Namespace:         rs.Namespace,
		Replicas:          0,
		ReadyReplicas:     rs.Status.ReadyReplicas,
		AvailableReplicas: rs.Status.AvailableReplicas,
		Labels:            labels,
		Selector:          selector,
		OwnerReferences:   make([]OwnerReference, 0),
		CreatedAt:         rs.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}

	if rs.Spec.Replicas != nil {
		replicaSetInfo.Replicas = *rs.Spec.Replicas
	}

	// 转换 OwnerReferences
	for _, owner := range rs.OwnerReferences {
		isController := false
		if owner.Controller != nil {
			isController = *owner.Controller
		}

		replicaSetInfo.OwnerReferences = append(replicaSetInfo.OwnerReferences, OwnerReference{
			Kind:       owner.Kind,
			Name:       owner.Name,
			UID:        string(owner.UID),
			Controller: isController,
		})
	}

	return replicaSetInfo
}
