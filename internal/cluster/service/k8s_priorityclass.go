package service

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/k8s-agent/common/errors"
	schedulingv1 "k8s.io/api/scheduling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sPriorityClassService PriorityClass 管理服务
type K8sPriorityClassService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sPriorityClassService 创建新的 PriorityClass 服务
func NewK8sPriorityClassService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sPriorityClassService {
	return &K8sPriorityClassService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// PriorityClassInfo PriorityClass 信息
type PriorityClassInfo struct {
	Name            string            `json:"name"`
	Value           int32             `json:"value"`
	GlobalDefault   bool              `json:"globalDefault"`
	PreemptionPolicy string           `json:"preemptionPolicy,omitempty"`
	Description     string            `json:"description,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	CreatedAt       string            `json:"createdAt"`
}

// ListPriorityClasses 获取 PriorityClass 列表
func (s *K8sPriorityClassService) ListPriorityClasses(ctx context.Context, clusterID string, offset, limit int) ([]PriorityClassInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	priorityClasses, err := client.Clientset().SchedulingV1().PriorityClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list priorityclasses: %w", err))
	}

	total := int64(len(priorityClasses.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(priorityClasses.Items) {
		start = len(priorityClasses.Items)
	}
	if end > len(priorityClasses.Items) {
		end = len(priorityClasses.Items)
	}

	pagedPriorityClasses := priorityClasses.Items[start:end]

	// 转换为 PriorityClassInfo
	result := make([]PriorityClassInfo, 0, len(pagedPriorityClasses))
	for _, pc := range pagedPriorityClasses {
		result = append(result, convertToPriorityClassInfo(&pc))
	}

	return result, total, nil
}

// GetPriorityClass 获取单个 PriorityClass 详情
func (s *K8sPriorityClassService) GetPriorityClass(ctx context.Context, clusterID, name string) (*schedulingv1.PriorityClass, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	pc, err := client.Clientset().SchedulingV1().PriorityClasses().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get priorityclass: %w", err))
	}

	return pc, nil
}

// DeletePriorityClass 删除 PriorityClass
func (s *K8sPriorityClassService) DeletePriorityClass(ctx context.Context, clusterID, name string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().SchedulingV1().PriorityClasses().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete priorityclass: %w", err))
	}

	return nil
}

// convertToPriorityClassInfo 转换 PriorityClass 为 PriorityClassInfo
func convertToPriorityClassInfo(pc *schedulingv1.PriorityClass) PriorityClassInfo {
	// 确保 Labels 不为 nil
	labels := pc.Labels
	if labels == nil {
		labels = make(map[string]string)
	}

	preemptionPolicy := ""
	if pc.PreemptionPolicy != nil {
		preemptionPolicy = string(*pc.PreemptionPolicy)
	}

	return PriorityClassInfo{
		Name:             pc.Name,
		Value:            pc.Value,
		GlobalDefault:    pc.GlobalDefault,
		PreemptionPolicy: preemptionPolicy,
		Description:      pc.Description,
		Labels:           labels,
		CreatedAt:        pc.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}
