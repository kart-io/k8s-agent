package types

import "time"

// Cluster K8s 集群信息.
type Cluster struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Endpoint    string    `json:"endpoint"`
	Version     string    `json:"version"`
	Status      string    `json:"status"` // healthy, unhealthy, unknown
	Region      string    `json:"region"`
	Provider    string    `json:"provider"` // aws, gcp, azure, self-hosted
	KubeConfig  string    `json:"kubeconfig,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ClusterHealth 集群健康状态.
type ClusterHealth struct {
	ClusterID   string    `json:"cluster_id"`
	Status      string    `json:"status"`
	TotalNodes  int       `json:"total_nodes"`
	ReadyNodes  int       `json:"ready_nodes"`
	TotalPods   int       `json:"total_pods"`
	RunningPods int       `json:"running_pods"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	CheckedAt   time.Time `json:"checked_at"`
}

// Node K8s 节点信息.
type Node struct {
	Name       string            `json:"name"`
	Status     string            `json:"status"`
	Roles      []string          `json:"roles"`
	Version    string            `json:"version"`
	OS         string            `json:"os"`
	Kernel     string            `json:"kernel"`
	CPUCores   int               `json:"cpu_cores"`
	Memory     string            `json:"memory"`
	Pods       int               `json:"pods"`
	Labels     map[string]string `json:"labels"`
	Conditions []NodeCondition   `json:"conditions"`
	CreatedAt  time.Time         `json:"created_at"`
}

// NodeCondition 节点状态条件.
type NodeCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
	LastTransitionTime time.Time `json:"last_transition_time"`
}

// Pod Pod 信息.
type Pod struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Status     string            `json:"status"`
	Phase      string            `json:"phase"`
	NodeName   string            `json:"node_name"`
	PodIP      string            `json:"pod_ip"`
	Labels     map[string]string `json:"labels"`
	Containers []Container       `json:"containers"`
	CreatedAt  time.Time         `json:"created_at"`
}

// Container 容器信息.
type Container struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	Ready        bool   `json:"ready"`
	RestartCount int32  `json:"restart_count"`
	State        string `json:"state"`
}

// Deployment Deployment 信息.
type Deployment struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int32             `json:"replicas"`
	ReadyReplicas     int32             `json:"ready_replicas"`
	AvailableReplicas int32             `json:"available_replicas"`
	UpdatedReplicas   int32             `json:"updated_replicas"`
	Labels            map[string]string `json:"labels"`
	Strategy          string            `json:"strategy"`
	CreatedAt         time.Time         `json:"created_at"`
}

// DeploymentCreate 创建 Deployment 请求.
type DeploymentCreate struct {
	Name      string               `json:"name" binding:"required"`
	Namespace string               `json:"namespace" binding:"required"`
	Replicas  int32                `json:"replicas" binding:"required"`
	Image     string               `json:"image" binding:"required"`
	Labels    map[string]string    `json:"labels"`
	Env       []EnvVar             `json:"env"`
	Ports     []ContainerPort      `json:"ports"`
	Resources ResourceRequirements `json:"resources"`
}

// EnvVar 环境变量.
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ContainerPort 容器端口.
type ContainerPort struct {
	Name          string `json:"name"`
	ContainerPort int32  `json:"container_port"`
	Protocol      string `json:"protocol"`
}

// ResourceRequirements 资源需求.
type ResourceRequirements struct {
	Limits   ResourceList `json:"limits"`
	Requests ResourceList `json:"requests"`
}

// ResourceList 资源列表.
type ResourceList struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// ClusterStats 集群统计信息.
type ClusterStats struct {
	ClusterID        string  `json:"cluster_id"`
	TotalNodes       int     `json:"total_nodes"`
	ReadyNodes       int     `json:"ready_nodes"`
	TotalPods        int     `json:"total_pods"`
	RunningPods      int     `json:"running_pods"`
	PendingPods      int     `json:"pending_pods"`
	FailedPods       int     `json:"failed_pods"`
	TotalNamespaces  int     `json:"total_namespaces"`
	TotalDeployments int     `json:"total_deployments"`
	TotalServices    int     `json:"total_services"`
	CPUCapacity      float64 `json:"cpu_capacity"`
	CPUUsage         float64 `json:"cpu_usage"`
	MemoryCapacity   float64 `json:"memory_capacity"`
	MemoryUsage      float64 `json:"memory_usage"`
}

// K8sEvent K8s 事件.
type K8sEvent struct {
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Object    string    `json:"object"`
	Namespace string    `json:"namespace"`
	Count     int32     `json:"count"`
	FirstTime time.Time `json:"first_time"`
	LastTime  time.Time `json:"last_time"`
}
