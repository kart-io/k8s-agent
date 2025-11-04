package k8s_tool

import (
	"fmt"
	"time"
)

// Mock 数据生成方法
// 这些方法在实际实现中会被替换为真实的 K8s API 调用

// getMockPodInfo 获取模拟的 Pod 信息.
func (t *K8sTool) getMockPodInfo(namespace, name string) *PodInfo {
	return &PodInfo{
		Name:      name,
		Namespace: namespace,
		Status:    "Running",
		Phase:     "Running",
		Ready:     "1/1",
		Restarts:  0,
		Age:       "2h30m",
		Node:      "node-1",
		IP:        "10.244.0.5",
		Labels: map[string]string{
			"app":     "test-app",
			"version": "v1.0",
		},
		Annotations: map[string]string{
			"prometheus.io/scrape": "true",
		},
		Containers: []ContainerInfo{
			{
				Name:         "main",
				Image:        "test-app:v1.0",
				Ready:        true,
				RestartCount: 0,
				State:        "Running",
				Resources: ResourceRequirements{
					Requests: map[string]string{
						"cpu":    "100m",
						"memory": "128Mi",
					},
					Limits: map[string]string{
						"cpu":    "500m",
						"memory": "512Mi",
					},
				},
			},
		},
		Conditions: []PodCondition{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: time.Now().Add(-2 * time.Hour),
				Reason:             "ContainersReady",
				Message:            "containers with unready status: []",
			},
		},
	}
}

// getMockDeploymentInfo 获取模拟的 Deployment 信息.
func (t *K8sTool) getMockDeploymentInfo(namespace, name string) *DeploymentInfo {
	return &DeploymentInfo{
		Name:              name,
		Namespace:         namespace,
		Replicas:          3,
		ReadyReplicas:     3,
		AvailableReplicas: 3,
		UpdatedReplicas:   3,
		Strategy:          "RollingUpdate",
		Labels: map[string]string{
			"app": "test-app",
		},
		Selector: map[string]string{
			"app": "test-app",
		},
		Age: "1d",
		Conditions: []DeploymentCondition{
			{
				Type:               "Available",
				Status:             "True",
				LastUpdateTime:     time.Now(),
				LastTransitionTime: time.Now().Add(-24 * time.Hour),
				Reason:             "MinimumReplicasAvailable",
				Message:            "Deployment has minimum availability.",
			},
		},
	}
}

// getMockServiceInfo 获取模拟的 Service 信息.
func (t *K8sTool) getMockServiceInfo(namespace, name string) *ServiceInfo {
	return &ServiceInfo{
		Name:       name,
		Namespace:  namespace,
		Type:       "ClusterIP",
		ClusterIP:  "10.96.0.10",
		ExternalIP: []string{},
		Ports: []ServicePort{
			{
				Name:       "http",
				Protocol:   "TCP",
				Port:       80,
				TargetPort: "8080",
			},
		},
		Selector: map[string]string{
			"app": "test-app",
		},
		Labels: map[string]string{
			"app": "test-app",
		},
		Age: "1d",
	}
}

// getMockNodeInfo 获取模拟的 Node 信息.
func (t *K8sTool) getMockNodeInfo(name string) *NodeInfo {
	return &NodeInfo{
		Name:             name,
		Status:           "Ready",
		Roles:            []string{"master"},
		Age:              "10d",
		Version:          "v1.28.0",
		InternalIP:       "192.168.1.10",
		ExternalIP:       "",
		OS:               "linux",
		KernelVersion:    "5.15.0-89-generic",
		ContainerRuntime: "containerd://1.7.2",
		Labels: map[string]string{
			"kubernetes.io/hostname":         name,
			"node-role.kubernetes.io/master": "",
		},
		Capacity: map[string]string{
			"cpu":    "4",
			"memory": "16Gi",
			"pods":   "110",
		},
		Allocatable: map[string]string{
			"cpu":    "3800m",
			"memory": "15Gi",
			"pods":   "110",
		},
		Conditions: []NodeCondition{
			{
				Type:               "Ready",
				Status:             "True",
				LastHeartbeatTime:  time.Now(),
				LastTransitionTime: time.Now().Add(-10 * 24 * time.Hour),
				Reason:             "KubeletReady",
				Message:            "kubelet is posting ready status",
			},
		},
	}
}

// getMockLogs 获取模拟的日志.
func (t *K8sTool) getMockLogs(podName string, opts *LogsOptions) string {
	timestamp := ""
	if opts.Timestamps {
		timestamp = time.Now().Format(time.RFC3339) + " "
	}

	logs := fmt.Sprintf("%sINFO: Application started\n", timestamp)
	logs += fmt.Sprintf("%sINFO: Connected to database\n", timestamp)
	logs += fmt.Sprintf("%sINFO: Server listening on port 8080\n", timestamp)

	if opts.TailLines > 0 && opts.TailLines < 3 {
		// 返回最后几行
		lines := []string{
			fmt.Sprintf("%sINFO: Application started", timestamp),
			fmt.Sprintf("%sINFO: Connected to database", timestamp),
			fmt.Sprintf("%sINFO: Server listening on port 8080", timestamp),
		}
		start := len(lines) - int(opts.TailLines)
		if start < 0 {
			start = 0
		}
		result := ""
		for i := start; i < len(lines); i++ {
			result += lines[i] + "\n"
		}
		return result
	}

	return logs
}

// getMockEvents 获取模拟的事件.
func (t *K8sTool) getMockEvents(namespace, resourceName string) []EventInfo {
	return []EventInfo{
		{
			Type:           "Normal",
			Reason:         "Scheduled",
			Message:        fmt.Sprintf("Successfully assigned %s/%s to node-1", namespace, resourceName),
			Count:          1,
			FirstTimestamp: time.Now().Add(-2 * time.Hour),
			LastTimestamp:  time.Now().Add(-2 * time.Hour),
			Source:         "default-scheduler",
			Object:         fmt.Sprintf("Pod/%s", resourceName),
		},
		{
			Type:           "Normal",
			Reason:         "Pulled",
			Message:        "Container image already present on machine",
			Count:          1,
			FirstTimestamp: time.Now().Add(-2 * time.Hour),
			LastTimestamp:  time.Now().Add(-2 * time.Hour),
			Source:         "kubelet",
			Object:         fmt.Sprintf("Pod/%s", resourceName),
		},
		{
			Type:           "Normal",
			Reason:         "Created",
			Message:        "Created container main",
			Count:          1,
			FirstTimestamp: time.Now().Add(-2 * time.Hour),
			LastTimestamp:  time.Now().Add(-2 * time.Hour),
			Source:         "kubelet",
			Object:         fmt.Sprintf("Pod/%s", resourceName),
		},
		{
			Type:           "Normal",
			Reason:         "Started",
			Message:        "Started container main",
			Count:          1,
			FirstTimestamp: time.Now().Add(-2 * time.Hour),
			LastTimestamp:  time.Now().Add(-2 * time.Hour),
			Source:         "kubelet",
			Object:         fmt.Sprintf("Pod/%s", resourceName),
		},
	}
}

// getMockPodList 获取模拟的 Pod 列表.
func (t *K8sTool) getMockPodList(namespace string) []PodInfo {
	return []PodInfo{
		*t.getMockPodInfo(namespace, "test-pod-1"),
		*t.getMockPodInfo(namespace, "test-pod-2"),
		*t.getMockPodInfo(namespace, "test-pod-3"),
	}
}

// getMockDeploymentList 获取模拟的 Deployment 列表.
func (t *K8sTool) getMockDeploymentList(namespace string) []DeploymentInfo {
	return []DeploymentInfo{
		*t.getMockDeploymentInfo(namespace, "test-deployment-1"),
		*t.getMockDeploymentInfo(namespace, "test-deployment-2"),
	}
}

// getMockMetrics 获取模拟的指标.
func (t *K8sTool) getMockMetrics(namespace, resourceName, resourceType string) *MetricsInfo {
	return &MetricsInfo{
		ResourceType: resourceType,
		ResourceName: resourceName,
		Namespace:    namespace,
		Timestamp:    time.Now(),
		CPU: ResourceMetric{
			Current:     "150m",
			Limit:       "500m",
			Request:     "100m",
			Utilization: 30.0,
		},
		Memory: ResourceMetric{
			Current:     "256Mi",
			Limit:       "512Mi",
			Request:     "128Mi",
			Utilization: 50.0,
		},
		Storage: ResourceMetric{
			Current:     "1Gi",
			Limit:       "10Gi",
			Request:     "5Gi",
			Utilization: 10.0,
		},
		Network: NetworkMetric{
			RxBytes:  1024 * 1024 * 100, // 100 MB
			TxBytes:  1024 * 1024 * 50,  // 50 MB
			RxErrors: 0,
			TxErrors: 0,
		},
	}
}
