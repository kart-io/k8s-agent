package k8sutils

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConvertToMap 将 Kubernetes 资源转换为 map
func ConvertToMap(obj interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// ExtractMetadata 提取 Kubernetes 资源的 metadata 信息
func ExtractMetadata(obj metav1.Object) map[string]interface{} {
	return map[string]interface{}{
		"name":              obj.GetName(),
		"namespace":         obj.GetNamespace(),
		"uid":               string(obj.GetUID()),
		"resourceVersion":   obj.GetResourceVersion(),
		"creationTimestamp": obj.GetCreationTimestamp().Time,
		"labels":            obj.GetLabels(),
		"annotations":       obj.GetAnnotations(),
	}
}

// ExtractPodInfo 提取 Pod 基本信息
func ExtractPodInfo(pod *corev1.Pod) map[string]interface{} {
	containerStatuses := make([]map[string]interface{}, 0, len(pod.Status.ContainerStatuses))
	for _, cs := range pod.Status.ContainerStatuses {
		containerStatuses = append(containerStatuses, map[string]interface{}{
			"name":         cs.Name,
			"ready":        cs.Ready,
			"restartCount": cs.RestartCount,
			"image":        cs.Image,
			"imageID":      cs.ImageID,
		})
	}

	return map[string]interface{}{
		"metadata": ExtractMetadata(pod),
		"spec": map[string]interface{}{
			"nodeName":           pod.Spec.NodeName,
			"serviceAccountName": pod.Spec.ServiceAccountName,
			"restartPolicy":      pod.Spec.RestartPolicy,
		},
		"status": map[string]interface{}{
			"phase":             pod.Status.Phase,
			"podIP":             pod.Status.PodIP,
			"hostIP":            pod.Status.HostIP,
			"startTime":         pod.Status.StartTime,
			"containerStatuses": containerStatuses,
		},
	}
}

// ExtractNodeInfo 提取 Node 基本信息
func ExtractNodeInfo(node *corev1.Node) map[string]interface{} {
	// 提取节点条件
	conditions := make([]map[string]interface{}, 0, len(node.Status.Conditions))
	for _, condition := range node.Status.Conditions {
		conditions = append(conditions, map[string]interface{}{
			"type":               condition.Type,
			"status":             condition.Status,
			"lastHeartbeatTime":  condition.LastHeartbeatTime.Time,
			"lastTransitionTime": condition.LastTransitionTime.Time,
			"reason":             condition.Reason,
			"message":            condition.Message,
		})
	}

	// 提取节点地址
	addresses := make([]map[string]interface{}, 0, len(node.Status.Addresses))
	for _, addr := range node.Status.Addresses {
		addresses = append(addresses, map[string]interface{}{
			"type":    addr.Type,
			"address": addr.Address,
		})
	}

	return map[string]interface{}{
		"metadata": ExtractMetadata(node),
		"spec": map[string]interface{}{
			"podCIDR":       node.Spec.PodCIDR,
			"unschedulable": node.Spec.Unschedulable,
		},
		"status": map[string]interface{}{
			"capacity":    node.Status.Capacity,
			"allocatable": node.Status.Allocatable,
			"conditions":  conditions,
			"addresses":   addresses,
			"nodeInfo": map[string]interface{}{
				"kubeletVersion":          node.Status.NodeInfo.KubeletVersion,
				"osImage":                 node.Status.NodeInfo.OSImage,
				"containerRuntimeVersion": node.Status.NodeInfo.ContainerRuntimeVersion,
				"architecture":            node.Status.NodeInfo.Architecture,
			},
		},
	}
}

// IsResourceReady 判断资源是否就绪
func IsResourceReady(conditions []metav1.Condition) bool {
	for _, condition := range conditions {
		if condition.Type == "Ready" && condition.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}

// IsPodReady 判断 Pod 是否就绪
func IsPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// IsPodRunning 判断 Pod 是否运行中
func IsPodRunning(pod *corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodRunning
}

// IsNodeReady 判断 Node 是否就绪
func IsNodeReady(node *corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
