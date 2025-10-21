package service

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cluster-service/internal/storage"
	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/common/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sLimitRangeService LimitRange 管理服务
type K8sLimitRangeService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sLimitRangeService 创建新的 LimitRange 服务
func NewK8sLimitRangeService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sLimitRangeService {
	return &K8sLimitRangeService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// LimitRangeInfo LimitRange 信息
type LimitRangeInfo struct {
	Name        string                 `json:"name"`
	Namespace   string                 `json:"namespace"`
	Limits      []LimitRangeItemInfo   `json:"limits"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	CreatedAt   string                 `json:"createdAt"`
}

// LimitRangeItemInfo LimitRange 限制项
type LimitRangeItemInfo struct {
	Type                 string                       `json:"type"`
	Max                  map[string]string            `json:"max,omitempty"`
	Min                  map[string]string            `json:"min,omitempty"`
	Default              map[string]string            `json:"default,omitempty"`
	DefaultRequest       map[string]string            `json:"defaultRequest,omitempty"`
	MaxLimitRequestRatio map[string]string            `json:"maxLimitRequestRatio,omitempty"`
}

// ListLimitRanges 获取 LimitRange 列表
func (s *K8sLimitRangeService) ListLimitRanges(ctx context.Context, clusterID, namespace string, offset, limit int) ([]LimitRangeInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	limitRanges, err := client.Clientset().CoreV1().LimitRanges(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list limitranges: %w", err))
	}

	total := int64(len(limitRanges.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(limitRanges.Items) {
		start = len(limitRanges.Items)
	}
	if end > len(limitRanges.Items) {
		end = len(limitRanges.Items)
	}

	result := make([]LimitRangeInfo, 0)
	for i := start; i < end; i++ {
		lr := limitRanges.Items[i]
		result = append(result, s.convertLimitRangeInfo(&lr))
	}

	return result, total, nil
}

// GetLimitRange 获取 LimitRange 详情
func (s *K8sLimitRangeService) GetLimitRange(ctx context.Context, clusterID, namespace, limitRangeName string) (*LimitRangeInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	limitRange, err := client.Clientset().CoreV1().LimitRanges(namespace).Get(ctx, limitRangeName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get limitrange: %w", err))
	}

	limitRangeInfo := s.convertLimitRangeInfo(limitRange)
	return &limitRangeInfo, nil
}

// CreateLimitRange 创建 LimitRange
func (s *K8sLimitRangeService) CreateLimitRange(ctx context.Context, clusterID, namespace string, limitRange *corev1.LimitRange) (*LimitRangeInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	createdLimitRange, err := client.Clientset().CoreV1().LimitRanges(namespace).Create(ctx, limitRange, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to create limitrange: %w", err))
	}

	logger.Infow("LimitRange created",
		"cluster_id", clusterID,
		"namespace", namespace,
		"limitrange", createdLimitRange.Name,
	)

	limitRangeInfo := s.convertLimitRangeInfo(createdLimitRange)
	return &limitRangeInfo, nil
}

// UpdateLimitRange 更新 LimitRange
func (s *K8sLimitRangeService) UpdateLimitRange(ctx context.Context, clusterID, namespace string, limitRange *corev1.LimitRange) (*LimitRangeInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	updatedLimitRange, err := client.Clientset().CoreV1().LimitRanges(namespace).Update(ctx, limitRange, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to update limitrange: %w", err))
	}

	logger.Infow("LimitRange updated",
		"cluster_id", clusterID,
		"namespace", namespace,
		"limitrange", updatedLimitRange.Name,
	)

	limitRangeInfo := s.convertLimitRangeInfo(updatedLimitRange)
	return &limitRangeInfo, nil
}

// DeleteLimitRange 删除 LimitRange
func (s *K8sLimitRangeService) DeleteLimitRange(ctx context.Context, clusterID, namespace, limitRangeName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().CoreV1().LimitRanges(namespace).Delete(ctx, limitRangeName, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete limitrange: %w", err))
	}

	logger.Infow("LimitRange deleted",
		"cluster_id", clusterID,
		"namespace", namespace,
		"limitrange", limitRangeName,
	)

	return nil
}

// convertLimitRangeInfo 转换 LimitRange 信息
func (s *K8sLimitRangeService) convertLimitRangeInfo(lr *corev1.LimitRange) LimitRangeInfo {
	// 确保 Labels 和 Annotations 不为 nil
	labels := lr.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	annotations := lr.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}

	limitRangeInfo := LimitRangeInfo{
		Name:        lr.Name,
		Namespace:   lr.Namespace,
		Limits:      make([]LimitRangeItemInfo, 0),
		Labels:      labels,
		Annotations: annotations,
		CreatedAt:   lr.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}

	// 转换 Limits
	for _, item := range lr.Spec.Limits {
		limitItem := LimitRangeItemInfo{
			Type:                 string(item.Type),
			Max:                  convertResourceList(item.Max),
			Min:                  convertResourceList(item.Min),
			Default:              convertResourceList(item.Default),
			DefaultRequest:       convertResourceList(item.DefaultRequest),
			MaxLimitRequestRatio: convertResourceList(item.MaxLimitRequestRatio),
		}
		limitRangeInfo.Limits = append(limitRangeInfo.Limits, limitItem)
	}

	return limitRangeInfo
}

// convertResourceList 转换资源列表为字符串 map
func convertResourceList(resourceList corev1.ResourceList) map[string]string {
	if resourceList == nil {
		return nil
	}

	result := make(map[string]string)
	for k, v := range resourceList {
		result[string(k)] = v.String()
	}
	return result
}
