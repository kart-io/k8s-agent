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

// K8sDaemonSetService DaemonSet 管理服务
type K8sDaemonSetService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sDaemonSetService 创建新的 DaemonSet 服务
func NewK8sDaemonSetService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sDaemonSetService {
	return &K8sDaemonSetService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// DaemonSetInfo DaemonSet 信息
type DaemonSetInfo struct {
	Name                   string            `json:"name"`
	Namespace              string            `json:"namespace"`
	DesiredNumberScheduled int32             `json:"desiredNumberScheduled"`
	CurrentNumberScheduled int32             `json:"currentNumberScheduled"`
	NumberReady            int32             `json:"numberReady"`
	NumberAvailable        int32             `json:"numberAvailable"`
	NumberMisscheduled     int32             `json:"numberMisscheduled"`
	UpdatedNumberScheduled int32             `json:"updatedNumberScheduled"`
	Selector               map[string]string `json:"selector"`
	Labels                 map[string]string `json:"labels"`
	UpdateStrategy         string            `json:"updateStrategy"`
	CreatedAt              string            `json:"createdAt"`
}

// ListDaemonSets 获取 DaemonSet 列表
func (s *K8sDaemonSetService) ListDaemonSets(ctx context.Context, clusterID, namespace string, offset, limit int) ([]DaemonSetInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	daemonsets, err := client.Clientset().AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list daemonsets: %w", err))
	}

	total := int64(len(daemonsets.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(daemonsets.Items) {
		start = len(daemonsets.Items)
	}
	if end > len(daemonsets.Items) {
		end = len(daemonsets.Items)
	}

	result := make([]DaemonSetInfo, 0)
	for i := start; i < end; i++ {
		ds := daemonsets.Items[i]
		result = append(result, s.convertDaemonSetInfo(&ds))
	}

	return result, total, nil
}

// GetDaemonSet 获取 DaemonSet 详情
func (s *K8sDaemonSetService) GetDaemonSet(ctx context.Context, clusterID, namespace, daemonsetName string) (*DaemonSetInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	daemonset, err := client.Clientset().AppsV1().DaemonSets(namespace).Get(ctx, daemonsetName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get daemonset: %w", err))
	}

	dsInfo := s.convertDaemonSetInfo(daemonset)
	return &dsInfo, nil
}

// RestartDaemonSet 重启 DaemonSet
func (s *K8sDaemonSetService) RestartDaemonSet(ctx context.Context, clusterID, namespace, daemonsetName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	// 获取当前的 DaemonSet
	daemonset, err := client.Clientset().AppsV1().DaemonSets(namespace).Get(ctx, daemonsetName, metav1.GetOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to get daemonset: %w", err))
	}

	// 添加重启注解（通过修改 Pod 模板的注解来触发重启）
	if daemonset.Spec.Template.ObjectMeta.Annotations == nil {
		daemonset.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	daemonset.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	// 更新 DaemonSet
	_, err = client.Clientset().AppsV1().DaemonSets(namespace).Update(ctx, daemonset, metav1.UpdateOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to restart daemonset: %w", err))
	}

	logger.Infow("DaemonSet restarted",
		"cluster_id", clusterID,
		"namespace", namespace,
		"daemonset", daemonsetName,
	)

	return nil
}

// DeleteDaemonSet 删除 DaemonSet
func (s *K8sDaemonSetService) DeleteDaemonSet(ctx context.Context, clusterID, namespace, daemonsetName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().AppsV1().DaemonSets(namespace).Delete(ctx, daemonsetName, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete daemonset: %w", err))
	}

	logger.Infow("DaemonSet deleted",
		"cluster_id", clusterID,
		"namespace", namespace,
		"daemonset", daemonsetName,
	)

	return nil
}

// convertDaemonSetInfo 转换 DaemonSet 信息
func (s *K8sDaemonSetService) convertDaemonSetInfo(daemonset *appsv1.DaemonSet) DaemonSetInfo {
	updateStrategy := string(daemonset.Spec.UpdateStrategy.Type)

	return DaemonSetInfo{
		Name:                   daemonset.Name,
		Namespace:              daemonset.Namespace,
		DesiredNumberScheduled: daemonset.Status.DesiredNumberScheduled,
		CurrentNumberScheduled: daemonset.Status.CurrentNumberScheduled,
		NumberReady:            daemonset.Status.NumberReady,
		NumberAvailable:        daemonset.Status.NumberAvailable,
		NumberMisscheduled:     daemonset.Status.NumberMisscheduled,
		UpdatedNumberScheduled: daemonset.Status.UpdatedNumberScheduled,
		Selector:               daemonset.Spec.Selector.MatchLabels,
		Labels:                 daemonset.Labels,
		UpdateStrategy:         updateStrategy,
		CreatedAt:              daemonset.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}
