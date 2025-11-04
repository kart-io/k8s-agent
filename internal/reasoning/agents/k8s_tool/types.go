package k8s_tool

import (
	"context"
	"time"
)

// Tool 定义 K8s 工具接口.
type Tool interface {
	// Name 返回工具名称
	Name() string

	// Description 返回工具描述
	Description() string

	// Execute 执行工具操作
	Execute(ctx context.Context, input *ToolInput) (*ToolOutput, error)
}

// ToolInput 工具输入.
type ToolInput struct {
	// 基本参数
	ClusterID    string `json:"cluster_id"`    // 集群 ID
	Namespace    string `json:"namespace"`     // 命名空间
	ResourceType string `json:"resource_type"` // 资源类型: "pod", "deployment", "service", etc.
	ResourceName string `json:"resource_name"` // 资源名称

	// 操作参数
	Action     string            `json:"action"`     // 操作: "get", "describe", "logs", "events", etc.
	Parameters map[string]string `json:"parameters"` // 额外参数

	// 查询选项
	LabelSelector string `json:"label_selector"` // 标签选择器
	FieldSelector string `json:"field_selector"` // 字段选择器
	Limit         int    `json:"limit"`          // 结果限制
}

// ToolOutput 工具输出.
type ToolOutput struct {
	// 执行结果
	Success  bool        `json:"success"`   // 是否成功
	Data     interface{} `json:"data"`      // 返回数据
	ErrorMsg string      `json:"error_msg"` // 错误信息

	// 元数据
	ToolName  string        `json:"tool_name"` // 工具名称
	Action    string        `json:"action"`    // 执行的操作
	Latency   time.Duration `json:"latency"`   // 延迟
	Timestamp time.Time     `json:"timestamp"` // 时间戳
}

// PodInfo Pod 信息.
type PodInfo struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Status      string            `json:"status"`
	Phase       string            `json:"phase"`
	Ready       string            `json:"ready"`
	Restarts    int32             `json:"restarts"`
	Age         string            `json:"age"`
	Node        string            `json:"node"`
	IP          string            `json:"ip"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Containers  []ContainerInfo   `json:"containers"`
	Conditions  []PodCondition    `json:"conditions"`
}

// ContainerInfo 容器信息.
type ContainerInfo struct {
	Name         string               `json:"name"`
	Image        string               `json:"image"`
	Ready        bool                 `json:"ready"`
	RestartCount int32                `json:"restart_count"`
	State        string               `json:"state"`
	Reason       string               `json:"reason"`
	Message      string               `json:"message"`
	Resources    ResourceRequirements `json:"resources"`
}

// ResourceRequirements 资源需求.
type ResourceRequirements struct {
	Requests map[string]string `json:"requests"`
	Limits   map[string]string `json:"limits"`
}

// PodCondition Pod 条件.
type PodCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastTransitionTime time.Time `json:"last_transition_time"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
}

// EventInfo 事件信息.
type EventInfo struct {
	Type           string    `json:"type"`
	Reason         string    `json:"reason"`
	Message        string    `json:"message"`
	Count          int32     `json:"count"`
	FirstTimestamp time.Time `json:"first_timestamp"`
	LastTimestamp  time.Time `json:"last_timestamp"`
	Source         string    `json:"source"`
	Object         string    `json:"object"`
}

// LogsOptions 日志选项.
type LogsOptions struct {
	Container    string `json:"container"`     // 容器名称
	Follow       bool   `json:"follow"`        // 是否跟随
	Previous     bool   `json:"previous"`      // 获取之前容器的日志
	TailLines    int64  `json:"tail_lines"`    // 尾部行数
	SinceSeconds int64  `json:"since_seconds"` // 最近多少秒的日志
	Timestamps   bool   `json:"timestamps"`    // 是否包含时间戳
}

// DeploymentInfo Deployment 信息.
type DeploymentInfo struct {
	Name              string                `json:"name"`
	Namespace         string                `json:"namespace"`
	Replicas          int32                 `json:"replicas"`
	ReadyReplicas     int32                 `json:"ready_replicas"`
	AvailableReplicas int32                 `json:"available_replicas"`
	UpdatedReplicas   int32                 `json:"updated_replicas"`
	Strategy          string                `json:"strategy"`
	Labels            map[string]string     `json:"labels"`
	Selector          map[string]string     `json:"selector"`
	Age               string                `json:"age"`
	Conditions        []DeploymentCondition `json:"conditions"`
}

// DeploymentCondition Deployment 条件.
type DeploymentCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastUpdateTime     time.Time `json:"last_update_time"`
	LastTransitionTime time.Time `json:"last_transition_time"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
}

// ServiceInfo Service 信息.
type ServiceInfo struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Type       string            `json:"type"`
	ClusterIP  string            `json:"cluster_ip"`
	ExternalIP []string          `json:"external_ip"`
	Ports      []ServicePort     `json:"ports"`
	Selector   map[string]string `json:"selector"`
	Labels     map[string]string `json:"labels"`
	Age        string            `json:"age"`
}

// ServicePort Service 端口.
type ServicePort struct {
	Name       string `json:"name"`
	Protocol   string `json:"protocol"`
	Port       int32  `json:"port"`
	TargetPort string `json:"target_port"`
	NodePort   int32  `json:"node_port"`
}

// NodeInfo Node 信息.
type NodeInfo struct {
	Name             string            `json:"name"`
	Status           string            `json:"status"`
	Roles            []string          `json:"roles"`
	Age              string            `json:"age"`
	Version          string            `json:"version"`
	InternalIP       string            `json:"internal_ip"`
	ExternalIP       string            `json:"external_ip"`
	OS               string            `json:"os"`
	KernelVersion    string            `json:"kernel_version"`
	ContainerRuntime string            `json:"container_runtime"`
	Labels           map[string]string `json:"labels"`
	Capacity         map[string]string `json:"capacity"`
	Allocatable      map[string]string `json:"allocatable"`
	Conditions       []NodeCondition   `json:"conditions"`
}

// NodeCondition Node 条件.
type NodeCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastHeartbeatTime  time.Time `json:"last_heartbeat_time"`
	LastTransitionTime time.Time `json:"last_transition_time"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
}

// MetricsInfo 指标信息.
type MetricsInfo struct {
	ResourceType string         `json:"resource_type"`
	ResourceName string         `json:"resource_name"`
	Namespace    string         `json:"namespace"`
	Timestamp    time.Time      `json:"timestamp"`
	CPU          ResourceMetric `json:"cpu"`
	Memory       ResourceMetric `json:"memory"`
	Storage      ResourceMetric `json:"storage"`
	Network      NetworkMetric  `json:"network"`
}

// ResourceMetric 资源指标.
type ResourceMetric struct {
	Current     string  `json:"current"`     // 当前使用量
	Limit       string  `json:"limit"`       // 限制
	Request     string  `json:"request"`     // 请求
	Utilization float64 `json:"utilization"` // 利用率百分比
}

// NetworkMetric 网络指标.
type NetworkMetric struct {
	RxBytes  int64 `json:"rx_bytes"`  // 接收字节数
	TxBytes  int64 `json:"tx_bytes"`  // 发送字节数
	RxErrors int64 `json:"rx_errors"` // 接收错误数
	TxErrors int64 `json:"tx_errors"` // 发送错误数
}

// ToolConfig K8s 工具配置.
type ToolConfig struct {
	// Kubeconfig 路径
	KubeconfigPath string `json:"kubeconfig_path"`

	// 默认命名空间
	DefaultNamespace string `json:"default_namespace"`

	// 超时配置
	Timeout time.Duration `json:"timeout"`

	// 日志配置
	MaxLogLines int64 `json:"max_log_lines"` // 最大日志行数

	// 缓存配置
	EnableCache bool          `json:"enable_cache"` // 是否启用缓存
	CacheTTL    time.Duration `json:"cache_ttl"`    // 缓存过期时间
}

// SupportedActions 支持的操作.
var SupportedActions = []string{
	"get",      // 获取资源
	"describe", // 描述资源详情
	"logs",     // 获取日志
	"events",   // 获取事件
	"list",     // 列出资源
	"metrics",  // 获取指标
	"top",      // 获取资源使用情况
}

// SupportedResourceTypes 支持的资源类型.
var SupportedResourceTypes = []string{
	"pod",
	"deployment",
	"replicaset",
	"service",
	"configmap",
	"secret",
	"node",
	"namespace",
	"persistentvolume",
	"persistentvolumeclaim",
	"ingress",
	"daemonset",
	"statefulset",
	"job",
	"cronjob",
}
