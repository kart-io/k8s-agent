package service

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/logger"
)

// K8sStatefulSetService StatefulSet 管理服务.
type K8sStatefulSetService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sStatefulSetService 创建新的 StatefulSet 服务.
func NewK8sStatefulSetService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sStatefulSetService {
	return &K8sStatefulSetService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// StatefulSetInfo StatefulSet 信息.
type StatefulSetInfo struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Replicas        int32             `json:"replicas"`
	ReadyReplicas   int32             `json:"readyReplicas"`
	CurrentReplicas int32             `json:"currentReplicas"`
	UpdatedReplicas int32             `json:"updatedReplicas"`
	ServiceName     string            `json:"serviceName"`
	Selector        map[string]string `json:"selector"`
	Labels          map[string]string `json:"labels"`
	UpdateStrategy  string            `json:"updateStrategy"`
	Partition       int32             `json:"partition,omitempty"`
	CreatedAt       string            `json:"createdAt"`
}

// ListStatefulSets 获取 StatefulSet 列表.
func (s *K8sStatefulSetService) ListStatefulSets(ctx context.Context, clusterID, namespace string, offset, limit int) ([]StatefulSetInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	statefulsets, err := client.Clientset().AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list statefulsets: %w", err))
	}

	total := int64(len(statefulsets.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(statefulsets.Items) {
		start = len(statefulsets.Items)
	}
	if end > len(statefulsets.Items) {
		end = len(statefulsets.Items)
	}

	result := make([]StatefulSetInfo, 0)
	for i := start; i < end; i++ {
		sts := statefulsets.Items[i]
		result = append(result, s.convertStatefulSetInfo(&sts))
	}

	return result, total, nil
}

// GetStatefulSet 获取 StatefulSet 详情.
func (s *K8sStatefulSetService) GetStatefulSet(ctx context.Context, clusterID, namespace, statefulsetName string) (*StatefulSetInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	statefulset, err := client.Clientset().AppsV1().StatefulSets(namespace).Get(ctx, statefulsetName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get statefulset: %w", err))
	}

	stsInfo := s.convertStatefulSetInfo(statefulset)
	return &stsInfo, nil
}

// ScaleStatefulSet 扩缩容 StatefulSet.
func (s *K8sStatefulSetService) ScaleStatefulSet(ctx context.Context, clusterID, namespace, statefulsetName string, replicas int32) (*StatefulSetInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	// 获取当前的 StatefulSet
	statefulset, err := client.Clientset().AppsV1().StatefulSets(namespace).Get(ctx, statefulsetName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get statefulset: %w", err))
	}

	// 更新副本数
	statefulset.Spec.Replicas = &replicas

	// 更新 StatefulSet
	updated, err := client.Clientset().AppsV1().StatefulSets(namespace).Update(ctx, statefulset, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to scale statefulset: %w", err))
	}

	logger.Infow("StatefulSet scaled",
		"cluster_id", clusterID,
		"namespace", namespace,
		"statefulset", statefulsetName,
		"replicas", replicas,
	)

	stsInfo := s.convertStatefulSetInfo(updated)
	return &stsInfo, nil
}

// RestartStatefulSet 重启 StatefulSet.
func (s *K8sStatefulSetService) RestartStatefulSet(ctx context.Context, clusterID, namespace, statefulsetName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	// 获取当前的 StatefulSet
	statefulset, err := client.Clientset().AppsV1().StatefulSets(namespace).Get(ctx, statefulsetName, metav1.GetOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to get statefulset: %w", err))
	}

	// 添加重启注解（通过修改 Pod 模板的注解来触发重启）
	if statefulset.Spec.Template.Annotations == nil {
		statefulset.Spec.Template.Annotations = make(map[string]string)
	}
	statefulset.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	// 更新 StatefulSet
	_, err = client.Clientset().AppsV1().StatefulSets(namespace).Update(ctx, statefulset, metav1.UpdateOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to restart statefulset: %w", err))
	}

	logger.Infow("StatefulSet restarted",
		"cluster_id", clusterID,
		"namespace", namespace,
		"statefulset", statefulsetName,
	)

	return nil
}

// DeleteStatefulSet 删除 StatefulSet.
func (s *K8sStatefulSetService) DeleteStatefulSet(ctx context.Context, clusterID, namespace, statefulsetName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().AppsV1().StatefulSets(namespace).Delete(ctx, statefulsetName, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete statefulset: %w", err))
	}

	logger.Infow("StatefulSet deleted",
		"cluster_id", clusterID,
		"namespace", namespace,
		"statefulset", statefulsetName,
	)

	return nil
}

// convertStatefulSetInfo 转换 StatefulSet 信息.
func (s *K8sStatefulSetService) convertStatefulSetInfo(statefulset *appsv1.StatefulSet) StatefulSetInfo {
	var replicas int32
	if statefulset.Spec.Replicas != nil {
		replicas = *statefulset.Spec.Replicas
	}

	updateStrategy := string(statefulset.Spec.UpdateStrategy.Type)
	var partition int32
	if statefulset.Spec.UpdateStrategy.RollingUpdate != nil &&
		statefulset.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
		partition = *statefulset.Spec.UpdateStrategy.RollingUpdate.Partition
	}

	return StatefulSetInfo{
		Name:            statefulset.Name,
		Namespace:       statefulset.Namespace,
		Replicas:        replicas,
		ReadyReplicas:   statefulset.Status.ReadyReplicas,
		CurrentReplicas: statefulset.Status.CurrentReplicas,
		UpdatedReplicas: statefulset.Status.UpdatedReplicas,
		ServiceName:     statefulset.Spec.ServiceName,
		Selector:        statefulset.Spec.Selector.MatchLabels,
		Labels:          statefulset.Labels,
		UpdateStrategy:  updateStrategy,
		Partition:       partition,
		CreatedAt:       statefulset.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}
