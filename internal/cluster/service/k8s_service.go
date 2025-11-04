package service

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/logger"
)

// K8sServiceService Service 管理服务
type K8sServiceService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sServiceService 创建新的 Service 服务
func NewK8sServiceService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sServiceService {
	return &K8sServiceService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// ServiceInfo Service 信息
type ServiceInfo struct {
	Name           string            `json:"name"`
	Namespace      string            `json:"namespace"`
	Type           string            `json:"type"`
	ClusterIP      string            `json:"clusterIP"`
	ExternalIPs    []string          `json:"externalIPs"`
	LoadBalancerIP string            `json:"loadBalancerIP,omitempty"`
	Ports          []ServicePort     `json:"ports"`
	Selector       map[string]string `json:"selector"`
	Labels         map[string]string `json:"labels"`
	Annotations    map[string]string `json:"annotations"`
	CreatedAt      string            `json:"createdAt"`
}

// ServicePort Service 端口
type ServicePort struct {
	Name       string `json:"name"`
	Protocol   string `json:"protocol"`
	Port       int32  `json:"port"`
	TargetPort string `json:"targetPort"`
	NodePort   int32  `json:"nodePort,omitempty"`
}

// CreateServiceRequest 创建 Service 请求
type CreateServiceRequest struct {
	Name        string            `json:"name" binding:"required"`
	Namespace   string            `json:"namespace" binding:"required"`
	Type        string            `json:"type"`
	Ports       []ServicePort     `json:"ports" binding:"required"`
	Selector    map[string]string `json:"selector" binding:"required"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// ListServices 获取 Service 列表
func (s *K8sServiceService) ListServices(ctx context.Context, clusterID, namespace string, offset, limit int) ([]ServiceInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	services, err := client.Clientset().CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list services: %w", err))
	}

	total := int64(len(services.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(services.Items) {
		start = len(services.Items)
	}
	if end > len(services.Items) {
		end = len(services.Items)
	}

	result := make([]ServiceInfo, 0)
	for i := start; i < end; i++ {
		svc := services.Items[i]
		result = append(result, s.convertServiceInfo(&svc))
	}

	return result, total, nil
}

// GetService 获取 Service 详情
func (s *K8sServiceService) GetService(ctx context.Context, clusterID, namespace, serviceName string) (*ServiceInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	service, err := client.Clientset().CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get service: %w", err))
	}

	serviceInfo := s.convertServiceInfo(service)
	return &serviceInfo, nil
}

// CreateService 创建 Service
func (s *K8sServiceService) CreateService(ctx context.Context, clusterID string, req *CreateServiceRequest) (*ServiceInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	// 构建 Service 对象
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   req.Namespace,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceType(req.Type),
			Selector: req.Selector,
			Ports:    s.convertToK8sPorts(req.Ports),
		},
	}

	// 如果没有指定类型，默认为 ClusterIP
	if service.Spec.Type == "" {
		service.Spec.Type = corev1.ServiceTypeClusterIP
	}

	created, err := client.Clientset().CoreV1().Services(req.Namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to create service: %w", err))
	}

	logger.Infow("Service created",
		"cluster_id", clusterID,
		"namespace", req.Namespace,
		"service", req.Name,
		"type", req.Type,
	)

	serviceInfo := s.convertServiceInfo(created)
	return &serviceInfo, nil
}

// UpdateService 更新 Service
func (s *K8sServiceService) UpdateService(ctx context.Context, clusterID, namespace, serviceName string, req *CreateServiceRequest) (*ServiceInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	// 获取现有 Service
	existing, err := client.Clientset().CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get service: %w", err))
	}

	// 更新字段
	existing.Labels = req.Labels
	existing.Annotations = req.Annotations
	existing.Spec.Selector = req.Selector
	existing.Spec.Ports = s.convertToK8sPorts(req.Ports)

	if req.Type != "" {
		existing.Spec.Type = corev1.ServiceType(req.Type)
	}

	updated, err := client.Clientset().CoreV1().Services(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to update service: %w", err))
	}

	logger.Infow("Service updated",
		"cluster_id", clusterID,
		"namespace", namespace,
		"service", serviceName,
	)

	serviceInfo := s.convertServiceInfo(updated)
	return &serviceInfo, nil
}

// DeleteService 删除 Service
func (s *K8sServiceService) DeleteService(ctx context.Context, clusterID, namespace, serviceName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().CoreV1().Services(namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete service: %w", err))
	}

	logger.Infow("Service deleted",
		"cluster_id", clusterID,
		"namespace", namespace,
		"service", serviceName,
	)

	return nil
}

// convertServiceInfo 转换 Service 信息
func (s *K8sServiceService) convertServiceInfo(service *corev1.Service) ServiceInfo {
	ports := make([]ServicePort, 0)
	for _, port := range service.Spec.Ports {
		targetPort := port.TargetPort.String()
		ports = append(ports, ServicePort{
			Name:       port.Name,
			Protocol:   string(port.Protocol),
			Port:       port.Port,
			TargetPort: targetPort,
			NodePort:   port.NodePort,
		})
	}

	externalIPs := service.Spec.ExternalIPs
	if externalIPs == nil {
		externalIPs = []string{}
	}

	loadBalancerIP := ""
	if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
		if len(service.Status.LoadBalancer.Ingress) > 0 {
			if service.Status.LoadBalancer.Ingress[0].IP != "" {
				loadBalancerIP = service.Status.LoadBalancer.Ingress[0].IP
			} else if service.Status.LoadBalancer.Ingress[0].Hostname != "" {
				loadBalancerIP = service.Status.LoadBalancer.Ingress[0].Hostname
			}
		}
	}

	return ServiceInfo{
		Name:           service.Name,
		Namespace:      service.Namespace,
		Type:           string(service.Spec.Type),
		ClusterIP:      service.Spec.ClusterIP,
		ExternalIPs:    externalIPs,
		LoadBalancerIP: loadBalancerIP,
		Ports:          ports,
		Selector:       service.Spec.Selector,
		Labels:         service.Labels,
		Annotations:    service.Annotations,
		CreatedAt:      service.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}

// convertToK8sPorts 转换为 K8s ServicePort 数组
func (s *K8sServiceService) convertToK8sPorts(ports []ServicePort) []corev1.ServicePort {
	k8sPorts := make([]corev1.ServicePort, 0)
	for _, port := range ports {
		targetPort := intstr.Parse(port.TargetPort)

		k8sPort := corev1.ServicePort{
			Name:       port.Name,
			Protocol:   corev1.Protocol(port.Protocol),
			Port:       port.Port,
			TargetPort: targetPort,
		}

		if port.NodePort > 0 {
			k8sPort.NodePort = port.NodePort
		}

		k8sPorts = append(k8sPorts, k8sPort)
	}
	return k8sPorts
}
