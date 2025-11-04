package service

import (
	"context"
	"fmt"

	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
)

// K8sEndpointSliceService EndpointSlice 管理服务
type K8sEndpointSliceService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sEndpointSliceService 创建新的 EndpointSlice 服务
func NewK8sEndpointSliceService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sEndpointSliceService {
	return &K8sEndpointSliceService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// EndpointSliceInfo EndpointSlice 信息
type EndpointSliceInfo struct {
	Name          string            `json:"name"`
	Namespace     string            `json:"namespace"`
	AddressType   string            `json:"addressType"`
	EndpointCount int               `json:"endpointCount"`
	PortCount     int               `json:"portCount"`
	Labels        map[string]string `json:"labels,omitempty"`
	CreatedAt     string            `json:"createdAt"`
}

// ListEndpointSlices 获取 EndpointSlice 列表
func (s *K8sEndpointSliceService) ListEndpointSlices(ctx context.Context, clusterID, namespace string, offset, limit int) ([]EndpointSliceInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	endpointSlices, err := client.Clientset().DiscoveryV1().EndpointSlices(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list endpointslices: %w", err))
	}

	total := int64(len(endpointSlices.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(endpointSlices.Items) {
		start = len(endpointSlices.Items)
	}
	if end > len(endpointSlices.Items) {
		end = len(endpointSlices.Items)
	}

	pagedSlices := endpointSlices.Items[start:end]

	// 转换为 EndpointSliceInfo
	result := make([]EndpointSliceInfo, 0, len(pagedSlices))
	for _, slice := range pagedSlices {
		result = append(result, convertToEndpointSliceInfo(&slice))
	}

	return result, total, nil
}

// GetEndpointSlice 获取单个 EndpointSlice 详情
func (s *K8sEndpointSliceService) GetEndpointSlice(ctx context.Context, clusterID, namespace, name string) (*discoveryv1.EndpointSlice, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	slice, err := client.Clientset().DiscoveryV1().EndpointSlices(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get endpointslice: %w", err))
	}

	return slice, nil
}

// DeleteEndpointSlice 删除 EndpointSlice
func (s *K8sEndpointSliceService) DeleteEndpointSlice(ctx context.Context, clusterID, namespace, name string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().DiscoveryV1().EndpointSlices(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete endpointslice: %w", err))
	}

	return nil
}

// convertToEndpointSliceInfo 转换 EndpointSlice 为 EndpointSliceInfo
func convertToEndpointSliceInfo(slice *discoveryv1.EndpointSlice) EndpointSliceInfo {
	return EndpointSliceInfo{
		Name:          slice.Name,
		Namespace:     slice.Namespace,
		AddressType:   string(slice.AddressType),
		EndpointCount: len(slice.Endpoints),
		PortCount:     len(slice.Ports),
		Labels:        slice.Labels,
		CreatedAt:     slice.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}
