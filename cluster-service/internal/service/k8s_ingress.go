package service

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cluster-service/internal/storage"
	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/common/logger"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sIngressService Ingress 管理服务
type K8sIngressService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sIngressService 创建新的 Ingress 服务
func NewK8sIngressService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sIngressService {
	return &K8sIngressService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// IngressInfo Ingress 信息
type IngressInfo struct {
	Name              string                  `json:"name"`
	Namespace         string                  `json:"namespace"`
	IngressClassName  *string                 `json:"ingressClassName,omitempty"`
	Rules             []IngressRule           `json:"rules"`
	TLS               []IngressTLS            `json:"tls,omitempty"`
	LoadBalancerIPs   []string                `json:"loadBalancerIPs,omitempty"`
	Labels            map[string]string       `json:"labels"`
	Annotations       map[string]string       `json:"annotations"`
	CreatedAt         string                  `json:"createdAt"`
}

// IngressRule Ingress 规则
type IngressRule struct {
	Host  string        `json:"host"`
	Paths []IngressPath `json:"paths"`
}

// IngressPath Ingress 路径
type IngressPath struct {
	Path        string             `json:"path"`
	PathType    string             `json:"pathType"`
	ServiceName string             `json:"serviceName"`
	ServicePort int32              `json:"servicePort"`
}

// IngressTLS Ingress TLS 配置
type IngressTLS struct {
	Hosts      []string `json:"hosts"`
	SecretName string   `json:"secretName"`
}

// ListIngresses 获取 Ingress 列表
func (s *K8sIngressService) ListIngresses(ctx context.Context, clusterID, namespace string, offset, limit int) ([]IngressInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	ingresses, err := client.Clientset().NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list ingresses: %w", err))
	}

	total := int64(len(ingresses.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(ingresses.Items) {
		start = len(ingresses.Items)
	}
	if end > len(ingresses.Items) {
		end = len(ingresses.Items)
	}

	result := make([]IngressInfo, 0)
	for i := start; i < end; i++ {
		ing := ingresses.Items[i]
		result = append(result, s.convertIngressInfo(&ing))
	}

	return result, total, nil
}

// GetIngress 获取 Ingress 详情
func (s *K8sIngressService) GetIngress(ctx context.Context, clusterID, namespace, ingressName string) (*IngressInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	ingress, err := client.Clientset().NetworkingV1().Ingresses(namespace).Get(ctx, ingressName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get ingress: %w", err))
	}

	ingressInfo := s.convertIngressInfo(ingress)
	return &ingressInfo, nil
}

// CreateIngress 创建 Ingress
func (s *K8sIngressService) CreateIngress(ctx context.Context, clusterID, namespace string, ingress *networkingv1.Ingress) (*IngressInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	createdIngress, err := client.Clientset().NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to create ingress: %w", err))
	}

	logger.Infow("Ingress created",
		"cluster_id", clusterID,
		"namespace", namespace,
		"ingress", createdIngress.Name,
	)

	ingressInfo := s.convertIngressInfo(createdIngress)
	return &ingressInfo, nil
}

// UpdateIngress 更新 Ingress
func (s *K8sIngressService) UpdateIngress(ctx context.Context, clusterID, namespace string, ingress *networkingv1.Ingress) (*IngressInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	updatedIngress, err := client.Clientset().NetworkingV1().Ingresses(namespace).Update(ctx, ingress, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to update ingress: %w", err))
	}

	logger.Infow("Ingress updated",
		"cluster_id", clusterID,
		"namespace", namespace,
		"ingress", updatedIngress.Name,
	)

	ingressInfo := s.convertIngressInfo(updatedIngress)
	return &ingressInfo, nil
}

// DeleteIngress 删除 Ingress
func (s *K8sIngressService) DeleteIngress(ctx context.Context, clusterID, namespace, ingressName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().NetworkingV1().Ingresses(namespace).Delete(ctx, ingressName, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete ingress: %w", err))
	}

	logger.Infow("Ingress deleted",
		"cluster_id", clusterID,
		"namespace", namespace,
		"ingress", ingressName,
	)

	return nil
}

// convertIngressInfo 转换 Ingress 信息
func (s *K8sIngressService) convertIngressInfo(ingress *networkingv1.Ingress) IngressInfo {
	// 确保 Labels 和 Annotations 不为 nil
	labels := ingress.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	annotations := ingress.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}

	ingressInfo := IngressInfo{
		Name:             ingress.Name,
		Namespace:        ingress.Namespace,
		IngressClassName: ingress.Spec.IngressClassName,
		Labels:           labels,
		Annotations:      annotations,
		CreatedAt:        ingress.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
		Rules:            make([]IngressRule, 0),
		TLS:              make([]IngressTLS, 0),
		LoadBalancerIPs:  make([]string, 0),
	}

	// 转换规则
	for _, rule := range ingress.Spec.Rules {
		ingressRule := IngressRule{
			Host:  rule.Host,
			Paths: make([]IngressPath, 0),
		}

		if rule.HTTP != nil {
			for _, path := range rule.HTTP.Paths {
				pathType := "Prefix"
				if path.PathType != nil {
					pathType = string(*path.PathType)
				}

				servicePort := int32(0)
				if path.Backend.Service != nil && path.Backend.Service.Port.Number > 0 {
					servicePort = path.Backend.Service.Port.Number
				}

				serviceName := ""
				if path.Backend.Service != nil {
					serviceName = path.Backend.Service.Name
				}

				ingressRule.Paths = append(ingressRule.Paths, IngressPath{
					Path:        path.Path,
					PathType:    pathType,
					ServiceName: serviceName,
					ServicePort: servicePort,
				})
			}
		}

		ingressInfo.Rules = append(ingressInfo.Rules, ingressRule)
	}

	// 转换 TLS 配置
	for _, tls := range ingress.Spec.TLS {
		ingressInfo.TLS = append(ingressInfo.TLS, IngressTLS{
			Hosts:      tls.Hosts,
			SecretName: tls.SecretName,
		})
	}

	// 获取 LoadBalancer IP
	for _, lbIngress := range ingress.Status.LoadBalancer.Ingress {
		if lbIngress.IP != "" {
			ingressInfo.LoadBalancerIPs = append(ingressInfo.LoadBalancerIPs, lbIngress.IP)
		}
	}

	return ingressInfo
}
