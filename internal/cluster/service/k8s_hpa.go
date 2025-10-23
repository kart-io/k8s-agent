package service

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sHPAService HorizontalPodAutoscaler 管理服务
type K8sHPAService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sHPAService 创建新的 HPA 服务
func NewK8sHPAService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sHPAService {
	return &K8sHPAService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// HPAInfo HPA 信息
type HPAInfo struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	ScaleTargetRef  string            `json:"scaleTargetRef"`
	MinReplicas     int32             `json:"minReplicas"`
	MaxReplicas     int32             `json:"maxReplicas"`
	CurrentReplicas int32             `json:"currentReplicas"`
	DesiredReplicas int32             `json:"desiredReplicas"`
	Labels          map[string]string `json:"labels,omitempty"`
	CreatedAt       string            `json:"createdAt"`
}

// ListHPAs 获取 HPA 列表
func (s *K8sHPAService) ListHPAs(ctx context.Context, clusterID, namespace string, offset, limit int) ([]HPAInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	hpas, err := client.Clientset().AutoscalingV2().HorizontalPodAutoscalers(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list hpas: %w", err))
	}

	total := int64(len(hpas.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(hpas.Items) {
		start = len(hpas.Items)
	}
	if end > len(hpas.Items) {
		end = len(hpas.Items)
	}

	pagedHPAs := hpas.Items[start:end]

	// 转换为 HPAInfo
	result := make([]HPAInfo, 0, len(pagedHPAs))
	for _, hpa := range pagedHPAs {
		result = append(result, convertToHPAInfo(&hpa))
	}

	return result, total, nil
}

// GetHPA 获取单个 HPA 详情
func (s *K8sHPAService) GetHPA(ctx context.Context, clusterID, namespace, name string) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	hpa, err := client.Clientset().AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get hpa: %w", err))
	}

	return hpa, nil
}

// DeleteHPA 删除 HPA
func (s *K8sHPAService) DeleteHPA(ctx context.Context, clusterID, namespace, name string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().AutoscalingV2().HorizontalPodAutoscalers(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete hpa: %w", err))
	}

	return nil
}

// convertToHPAInfo 转换 HPA 为 HPAInfo
func convertToHPAInfo(hpa *autoscalingv2.HorizontalPodAutoscaler) HPAInfo {
	minReplicas := int32(1)
	if hpa.Spec.MinReplicas != nil {
		minReplicas = *hpa.Spec.MinReplicas
	}

	scaleTargetRef := fmt.Sprintf("%s/%s", hpa.Spec.ScaleTargetRef.Kind, hpa.Spec.ScaleTargetRef.Name)

	return HPAInfo{
		Name:            hpa.Name,
		Namespace:       hpa.Namespace,
		ScaleTargetRef:  scaleTargetRef,
		MinReplicas:     minReplicas,
		MaxReplicas:     hpa.Spec.MaxReplicas,
		CurrentReplicas: hpa.Status.CurrentReplicas,
		DesiredReplicas: hpa.Status.DesiredReplicas,
		Labels:          hpa.Labels,
		CreatedAt:       hpa.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}
