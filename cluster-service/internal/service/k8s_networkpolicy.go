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

// K8sNetworkPolicyService NetworkPolicy 管理服务
type K8sNetworkPolicyService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sNetworkPolicyService 创建新的 NetworkPolicy 服务
func NewK8sNetworkPolicyService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sNetworkPolicyService {
	return &K8sNetworkPolicyService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// NetworkPolicyInfo NetworkPolicy 信息
type NetworkPolicyInfo struct {
	Name        string                    `json:"name"`
	Namespace   string                    `json:"namespace"`
	PodSelector map[string]string         `json:"podSelector"`
	PolicyTypes []string                  `json:"policyTypes"`
	Ingress     []NetworkPolicyIngressRule `json:"ingress,omitempty"`
	Egress      []NetworkPolicyEgressRule  `json:"egress,omitempty"`
	Labels      map[string]string         `json:"labels"`
	Annotations map[string]string         `json:"annotations"`
	CreatedAt   string                    `json:"createdAt"`
}

// NetworkPolicyIngressRule Ingress 规则
type NetworkPolicyIngressRule struct {
	Ports []NetworkPolicyPort `json:"ports,omitempty"`
	From  []NetworkPolicyPeer `json:"from,omitempty"`
}

// NetworkPolicyEgressRule Egress 规则
type NetworkPolicyEgressRule struct {
	Ports []NetworkPolicyPort `json:"ports,omitempty"`
	To    []NetworkPolicyPeer `json:"to,omitempty"`
}

// NetworkPolicyPort 端口规则
type NetworkPolicyPort struct {
	Protocol string `json:"protocol"`
	Port     string `json:"port"`
}

// NetworkPolicyPeer 网络策略对等方
type NetworkPolicyPeer struct {
	PodSelector       map[string]string `json:"podSelector,omitempty"`
	NamespaceSelector map[string]string `json:"namespaceSelector,omitempty"`
	IPBlock           *IPBlock          `json:"ipBlock,omitempty"`
}

// IPBlock IP 块
type IPBlock struct {
	CIDR   string   `json:"cidr"`
	Except []string `json:"except,omitempty"`
}

// ListNetworkPolicies 获取 NetworkPolicy 列表
func (s *K8sNetworkPolicyService) ListNetworkPolicies(ctx context.Context, clusterID, namespace string, offset, limit int) ([]NetworkPolicyInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	networkPolicies, err := client.Clientset().NetworkingV1().NetworkPolicies(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list networkpolicies: %w", err))
	}

	total := int64(len(networkPolicies.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(networkPolicies.Items) {
		start = len(networkPolicies.Items)
	}
	if end > len(networkPolicies.Items) {
		end = len(networkPolicies.Items)
	}

	result := make([]NetworkPolicyInfo, 0)
	for i := start; i < end; i++ {
		np := networkPolicies.Items[i]
		result = append(result, s.convertNetworkPolicyInfo(&np))
	}

	return result, total, nil
}

// GetNetworkPolicy 获取 NetworkPolicy 详情
func (s *K8sNetworkPolicyService) GetNetworkPolicy(ctx context.Context, clusterID, namespace, networkPolicyName string) (*NetworkPolicyInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	networkPolicy, err := client.Clientset().NetworkingV1().NetworkPolicies(namespace).Get(ctx, networkPolicyName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get networkpolicy: %w", err))
	}

	networkPolicyInfo := s.convertNetworkPolicyInfo(networkPolicy)
	return &networkPolicyInfo, nil
}

// CreateNetworkPolicy 创建 NetworkPolicy
func (s *K8sNetworkPolicyService) CreateNetworkPolicy(ctx context.Context, clusterID, namespace string, networkPolicy *networkingv1.NetworkPolicy) (*NetworkPolicyInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	createdNetworkPolicy, err := client.Clientset().NetworkingV1().NetworkPolicies(namespace).Create(ctx, networkPolicy, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to create networkpolicy: %w", err))
	}

	logger.Infow("NetworkPolicy created",
		"cluster_id", clusterID,
		"namespace", namespace,
		"networkpolicy", createdNetworkPolicy.Name,
	)

	networkPolicyInfo := s.convertNetworkPolicyInfo(createdNetworkPolicy)
	return &networkPolicyInfo, nil
}

// UpdateNetworkPolicy 更新 NetworkPolicy
func (s *K8sNetworkPolicyService) UpdateNetworkPolicy(ctx context.Context, clusterID, namespace string, networkPolicy *networkingv1.NetworkPolicy) (*NetworkPolicyInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	updatedNetworkPolicy, err := client.Clientset().NetworkingV1().NetworkPolicies(namespace).Update(ctx, networkPolicy, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to update networkpolicy: %w", err))
	}

	logger.Infow("NetworkPolicy updated",
		"cluster_id", clusterID,
		"namespace", namespace,
		"networkpolicy", updatedNetworkPolicy.Name,
	)

	networkPolicyInfo := s.convertNetworkPolicyInfo(updatedNetworkPolicy)
	return &networkPolicyInfo, nil
}

// DeleteNetworkPolicy 删除 NetworkPolicy
func (s *K8sNetworkPolicyService) DeleteNetworkPolicy(ctx context.Context, clusterID, namespace, networkPolicyName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().NetworkingV1().NetworkPolicies(namespace).Delete(ctx, networkPolicyName, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete networkpolicy: %w", err))
	}

	logger.Infow("NetworkPolicy deleted",
		"cluster_id", clusterID,
		"namespace", namespace,
		"networkpolicy", networkPolicyName,
	)

	return nil
}

// convertNetworkPolicyInfo 转换 NetworkPolicy 信息
func (s *K8sNetworkPolicyService) convertNetworkPolicyInfo(np *networkingv1.NetworkPolicy) NetworkPolicyInfo {
	// 确保 Labels 和 Annotations 不为 nil
	labels := np.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	annotations := np.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}
	podSelector := np.Spec.PodSelector.MatchLabels
	if podSelector == nil {
		podSelector = make(map[string]string)
	}

	networkPolicyInfo := NetworkPolicyInfo{
		Name:        np.Name,
		Namespace:   np.Namespace,
		PodSelector: podSelector,
		PolicyTypes: make([]string, 0),
		Ingress:     make([]NetworkPolicyIngressRule, 0),
		Egress:      make([]NetworkPolicyEgressRule, 0),
		Labels:      labels,
		Annotations: annotations,
		CreatedAt:   np.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}

	// 转换策略类型
	for _, policyType := range np.Spec.PolicyTypes {
		networkPolicyInfo.PolicyTypes = append(networkPolicyInfo.PolicyTypes, string(policyType))
	}

	// 转换 Ingress 规则
	for _, ingress := range np.Spec.Ingress {
		ingressRule := NetworkPolicyIngressRule{
			Ports: make([]NetworkPolicyPort, 0),
			From:  make([]NetworkPolicyPeer, 0),
		}

		// 转换端口
		for _, port := range ingress.Ports {
			protocol := "TCP"
			if port.Protocol != nil {
				protocol = string(*port.Protocol)
			}

			portValue := ""
			if port.Port != nil {
				portValue = port.Port.String()
			}

			ingressRule.Ports = append(ingressRule.Ports, NetworkPolicyPort{
				Protocol: protocol,
				Port:     portValue,
			})
		}

		// 转换来源
		for _, from := range ingress.From {
			peer := s.convertNetworkPolicyPeer(&from)
			ingressRule.From = append(ingressRule.From, peer)
		}

		networkPolicyInfo.Ingress = append(networkPolicyInfo.Ingress, ingressRule)
	}

	// 转换 Egress 规则
	for _, egress := range np.Spec.Egress {
		egressRule := NetworkPolicyEgressRule{
			Ports: make([]NetworkPolicyPort, 0),
			To:    make([]NetworkPolicyPeer, 0),
		}

		// 转换端口
		for _, port := range egress.Ports {
			protocol := "TCP"
			if port.Protocol != nil {
				protocol = string(*port.Protocol)
			}

			portValue := ""
			if port.Port != nil {
				portValue = port.Port.String()
			}

			egressRule.Ports = append(egressRule.Ports, NetworkPolicyPort{
				Protocol: protocol,
				Port:     portValue,
			})
		}

		// 转换目标
		for _, to := range egress.To {
			peer := s.convertNetworkPolicyPeer(&to)
			egressRule.To = append(egressRule.To, peer)
		}

		networkPolicyInfo.Egress = append(networkPolicyInfo.Egress, egressRule)
	}

	return networkPolicyInfo
}

// convertNetworkPolicyPeer 转换网络策略对等方
func (s *K8sNetworkPolicyService) convertNetworkPolicyPeer(peer *networkingv1.NetworkPolicyPeer) NetworkPolicyPeer {
	policyPeer := NetworkPolicyPeer{}

	if peer.PodSelector != nil {
		policyPeer.PodSelector = peer.PodSelector.MatchLabels
	}

	if peer.NamespaceSelector != nil {
		policyPeer.NamespaceSelector = peer.NamespaceSelector.MatchLabels
	}

	if peer.IPBlock != nil {
		policyPeer.IPBlock = &IPBlock{
			CIDR:   peer.IPBlock.CIDR,
			Except: peer.IPBlock.Except,
		}
	}

	return policyPeer
}
