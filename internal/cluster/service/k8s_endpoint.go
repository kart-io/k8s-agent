package service

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/logger"
)

// K8sEndpointService Endpoints 管理服务
type K8sEndpointService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sEndpointService 创建新的 Endpoints 服务
func NewK8sEndpointService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sEndpointService {
	return &K8sEndpointService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// EndpointAddress 端点地址信息
type EndpointAddress struct {
	IP       string `json:"ip"`
	NodeName string `json:"nodeName,omitempty"`
	Hostname string `json:"hostname,omitempty"`
}

// EndpointPort 端点端口信息
type EndpointPort struct {
	Name     string `json:"name,omitempty"`
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
}

// EndpointSubset 端点子集
type EndpointSubset struct {
	Addresses         []EndpointAddress `json:"addresses,omitempty"`
	NotReadyAddresses []EndpointAddress `json:"notReadyAddresses,omitempty"`
	Ports             []EndpointPort    `json:"ports,omitempty"`
}

// EndpointInfo Endpoints 信息
type EndpointInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Subsets   []EndpointSubset  `json:"subsets,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt string            `json:"createdAt"`
}

// ListEndpoints 获取 Endpoints 列表
func (s *K8sEndpointService) ListEndpoints(ctx context.Context, clusterID, namespace string, offset, limit int) ([]EndpointInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	endpoints, err := client.Clientset().CoreV1().Endpoints(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list endpoints: %w", err))
	}

	total := int64(len(endpoints.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(endpoints.Items) {
		start = len(endpoints.Items)
	}
	if end > len(endpoints.Items) {
		end = len(endpoints.Items)
	}

	result := make([]EndpointInfo, 0)
	for i := start; i < end; i++ {
		ep := endpoints.Items[i]
		result = append(result, s.convertEndpointInfo(&ep))
	}

	return result, total, nil
}

// GetEndpoint 获取 Endpoint 详情
func (s *K8sEndpointService) GetEndpoint(ctx context.Context, clusterID, namespace, endpointName string) (*EndpointInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	endpoint, err := client.Clientset().CoreV1().Endpoints(namespace).Get(ctx, endpointName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get endpoint: %w", err))
	}

	epInfo := s.convertEndpointInfo(endpoint)
	return &epInfo, nil
}

// DeleteEndpoint 删除 Endpoint
func (s *K8sEndpointService) DeleteEndpoint(ctx context.Context, clusterID, namespace, endpointName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().CoreV1().Endpoints(namespace).Delete(ctx, endpointName, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete endpoint: %w", err))
	}

	logger.Infow("Endpoint deleted",
		"cluster_id", clusterID,
		"namespace", namespace,
		"endpoint", endpointName,
	)

	return nil
}

// convertEndpointInfo 转换 Endpoints 信息
func (s *K8sEndpointService) convertEndpointInfo(endpoint *corev1.Endpoints) EndpointInfo {
	subsets := make([]EndpointSubset, 0, len(endpoint.Subsets))

	for _, subset := range endpoint.Subsets {
		// 转换地址
		addresses := make([]EndpointAddress, 0, len(subset.Addresses))
		for _, addr := range subset.Addresses {
			address := EndpointAddress{
				IP:       addr.IP,
				Hostname: addr.Hostname,
			}
			if addr.NodeName != nil {
				address.NodeName = *addr.NodeName
			}
			addresses = append(addresses, address)
		}

		// 转换未就绪地址
		notReadyAddresses := make([]EndpointAddress, 0, len(subset.NotReadyAddresses))
		for _, addr := range subset.NotReadyAddresses {
			address := EndpointAddress{
				IP:       addr.IP,
				Hostname: addr.Hostname,
			}
			if addr.NodeName != nil {
				address.NodeName = *addr.NodeName
			}
			notReadyAddresses = append(notReadyAddresses, address)
		}

		// 转换端口
		ports := make([]EndpointPort, 0, len(subset.Ports))
		for _, port := range subset.Ports {
			ports = append(ports, EndpointPort{
				Name:     port.Name,
				Port:     port.Port,
				Protocol: string(port.Protocol),
			})
		}

		subsets = append(subsets, EndpointSubset{
			Addresses:         addresses,
			NotReadyAddresses: notReadyAddresses,
			Ports:             ports,
		})
	}

	return EndpointInfo{
		Name:      endpoint.Name,
		Namespace: endpoint.Namespace,
		Subsets:   subsets,
		Labels:    endpoint.Labels,
		CreatedAt: endpoint.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}
