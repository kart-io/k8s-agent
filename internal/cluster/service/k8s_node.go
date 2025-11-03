package service

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/logger"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sNodeService Node 管理服务
type K8sNodeService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sNodeService 创建新的 Node 服务
func NewK8sNodeService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sNodeService {
	return &K8sNodeService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// NodeInfo Node 信息
type NodeInfo struct {
	Name             string            `json:"name"`
	Status           string            `json:"status"`
	Roles            []string          `json:"roles"`
	Version          string            `json:"version"`
	InternalIP       string            `json:"internalIP"`
	ExternalIP       string            `json:"externalIP"`
	OSImage          string            `json:"osImage"`
	KernelVersion    string            `json:"kernelVersion"`
	ContainerRuntime string            `json:"containerRuntime"`
	Capacity         map[string]string `json:"capacity"`
	Allocatable      map[string]string `json:"allocatable"`
	Conditions       []NodeCondition   `json:"conditions"`
	Labels           map[string]string `json:"labels"`
	Annotations      map[string]string `json:"annotations"`
	Taints           []NodeTaint       `json:"taints"`
	CreatedAt        string            `json:"createdAt"`
}

// NodeCondition Node 条件
type NodeCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

// NodeTaint Node 污点
type NodeTaint struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"`
}

// ListNodes 获取 Node 列表
func (s *K8sNodeService) ListNodes(ctx context.Context, clusterID string, offset, limit int) ([]NodeInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	nodes, err := client.Clientset().CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list nodes: %w", err))
	}

	total := int64(len(nodes.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(nodes.Items) {
		start = len(nodes.Items)
	}
	if end > len(nodes.Items) {
		end = len(nodes.Items)
	}

	result := make([]NodeInfo, 0)
	for i := start; i < end; i++ {
		node := nodes.Items[i]
		result = append(result, s.convertNodeInfo(&node))
	}

	return result, total, nil
}

// GetNode 获取 Node 详情
func (s *K8sNodeService) GetNode(ctx context.Context, clusterID, nodeName string) (*NodeInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	node, err := client.Clientset().CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get node: %w", err))
	}

	nodeInfo := s.convertNodeInfo(node)
	return &nodeInfo, nil
}

// CordonNode 标记 Node 为不可调度
func (s *K8sNodeService) CordonNode(ctx context.Context, clusterID, nodeName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	node, err := client.Clientset().CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to get node: %w", err))
	}

	// 设置不可调度
	node.Spec.Unschedulable = true

	_, err = client.Clientset().CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to cordon node: %w", err))
	}

	logger.Infow("Node cordoned",
		"cluster_id", clusterID,
		"node", nodeName,
	)

	return nil
}

// UncordonNode 标记 Node 为可调度
func (s *K8sNodeService) UncordonNode(ctx context.Context, clusterID, nodeName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	node, err := client.Clientset().CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to get node: %w", err))
	}

	// 设置可调度
	node.Spec.Unschedulable = false

	_, err = client.Clientset().CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to uncordon node: %w", err))
	}

	logger.Infow("Node uncordoned",
		"cluster_id", clusterID,
		"node", nodeName,
	)

	return nil
}

// DrainNode 驱逐 Node 上的 Pod
func (s *K8sNodeService) DrainNode(ctx context.Context, clusterID, nodeName string, force bool) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	// 首先 cordon node
	if err := s.CordonNode(ctx, clusterID, nodeName); err != nil {
		return err
	}

	// 获取 node 上的所有 pods
	pods, err := client.Clientset().CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to list pods on node: %w", err))
	}

	// 驱逐 pods
	for _, pod := range pods.Items {
		// 跳过 DaemonSet 管理的 Pod（除非 force=true）
		if !force && s.isDaemonSetPod(&pod) {
			continue
		}

		// 删除 Pod
		err := client.Clientset().CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
		if err != nil {
			logger.Warnw("Failed to delete pod during drain",
				"cluster_id", clusterID,
				"node", nodeName,
				"pod", pod.Name,
				"namespace", pod.Namespace,
				"error", err.Error(),
			)
		}
	}

	logger.Infow("Node drained",
		"cluster_id", clusterID,
		"node", nodeName,
		"pods_evicted", len(pods.Items),
	)

	return nil
}

// convertNodeInfo 转换 Node 信息
func (s *K8sNodeService) convertNodeInfo(node *corev1.Node) NodeInfo {
	// 提取角色
	roles := make([]string, 0)
	for label := range node.Labels {
		if label == "node-role.kubernetes.io/master" || label == "node-role.kubernetes.io/control-plane" {
			roles = append(roles, "master")
		} else if label == "node-role.kubernetes.io/worker" {
			roles = append(roles, "worker")
		}
	}
	if len(roles) == 0 {
		roles = append(roles, "worker")
	}

	// 提取 IP 地址
	var internalIP, externalIP string
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			internalIP = addr.Address
		} else if addr.Type == corev1.NodeExternalIP {
			externalIP = addr.Address
		}
	}

	// 提取状态
	status := "Unknown"
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				status = "Ready"
			} else {
				status = "NotReady"
			}
			break
		}
	}

	// 转换 Conditions
	conditions := make([]NodeCondition, 0)
	for _, cond := range node.Status.Conditions {
		conditions = append(conditions, NodeCondition{
			Type:    string(cond.Type),
			Status:  string(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	// 转换 Taints
	taints := make([]NodeTaint, 0)
	for _, taint := range node.Spec.Taints {
		taints = append(taints, NodeTaint{
			Key:    taint.Key,
			Value:  taint.Value,
			Effect: string(taint.Effect),
		})
	}

	// 转换 Capacity 和 Allocatable
	capacity := make(map[string]string)
	for k, v := range node.Status.Capacity {
		capacity[string(k)] = v.String()
	}

	allocatable := make(map[string]string)
	for k, v := range node.Status.Allocatable {
		allocatable[string(k)] = v.String()
	}

	return NodeInfo{
		Name:             node.Name,
		Status:           status,
		Roles:            roles,
		Version:          node.Status.NodeInfo.KubeletVersion,
		InternalIP:       internalIP,
		ExternalIP:       externalIP,
		OSImage:          node.Status.NodeInfo.OSImage,
		KernelVersion:    node.Status.NodeInfo.KernelVersion,
		ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
		Capacity:         capacity,
		Allocatable:      allocatable,
		Conditions:       conditions,
		Labels:           node.Labels,
		Annotations:      node.Annotations,
		Taints:           taints,
		CreatedAt:        node.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}

// isDaemonSetPod 判断是否为 DaemonSet 管理的 Pod
func (s *K8sNodeService) isDaemonSetPod(pod *corev1.Pod) bool {
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "DaemonSet" {
			return true
		}
	}
	return false
}
