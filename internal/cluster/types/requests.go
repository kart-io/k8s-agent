package types

// ===========================
// 通用请求结构体
// ===========================

// PathParam 通用查询参数结构体 (已改为查询参数风格).
type PathParam struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace"`                    // 命名空间
	Name      string `form:"name"`                         // 资源名称
}

// ===========================
// 集群管理请求
// ===========================

// GetClusterRequest 获取集群详情请求.
type GetClusterRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
}

// CreateClusterRequest 创建集群请求.
type CreateClusterRequest struct {
	Name        string            `json:"name" binding:"required"`       // 集群名称
	Description string            `json:"description"`                   // 集群描述
	Endpoint    string            `json:"endpoint" binding:"required"`   // API Server 地址
	KubeConfig  string            `json:"kubeconfig" binding:"required"` // kubeconfig 内容
	Region      string            `json:"region"`                        // 区域
	Provider    string            `json:"provider"`                      // 提供商
	Labels      map[string]string `json:"labels"`                        // 标签
}

// UpdateClusterRequest 更新集群请求.
type UpdateClusterRequest struct {
	ClusterID   string            `form:"clusterId" binding:"required"` // 集群 ID
	Name        string            `json:"name"`                         // 集群名称
	Description string            `json:"description"`                  // 集群描述
	Endpoint    string            `json:"endpoint"`                     // API Server 地址
	KubeConfig  string            `json:"kubeconfig"`                   // kubeconfig 内容
	Region      string            `json:"region"`                       // 区域
	Provider    string            `json:"provider"`                     // 提供商
	Labels      map[string]string `json:"labels"`                       // 标签
}

// DeleteClusterRequest 删除集群请求.
type DeleteClusterRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
}

// GetClusterHealthRequest 获取集群健康状态请求.
type GetClusterHealthRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
}

// ===========================
// 命名空间管理请求
// ===========================

// ListNamespacesRequest 获取命名空间列表请求.
type ListNamespacesRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
}

// GetNamespaceRequest 获取命名空间详情请求.
type GetNamespaceRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间名称
}

// CreateNamespaceRequest 创建命名空间请求.
type CreateNamespaceRequest struct {
	ClusterID string            `form:"clusterId" binding:"required"` // 集群 ID
	Name      string            `json:"name" binding:"required"`      // 命名空间名称
	Labels    map[string]string `json:"labels"`                       // 标签
}

// DeleteNamespaceRequest 删除命名空间请求.
type DeleteNamespaceRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间名称
}

// ===========================
// Pod 管理请求
// ===========================

// ListPodsRequest 获取 Pod 列表请求.
type ListPodsRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace"`                    // 命名空间
}

// GetPodRequest 获取 Pod 详情请求.
type GetPodRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // Pod 名称
}

// DeletePodRequest 删除 Pod 请求.
type DeletePodRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // Pod 名称
}

// GetPodLogsRequest 获取 Pod 日志请求.
type GetPodLogsRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // Pod 名称
	Container string `form:"container"`                    // 容器名称（可选）
	TailLines string `form:"tailLines"`                    // 尾行数（默认 100）
	Follow    bool   `form:"follow"`                       // 是否跟踪日志
}

// ===========================
// Deployment 管理请求
// ===========================

// ListDeploymentsRequest 获取 Deployment 列表请求.
type ListDeploymentsRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace"`                    // 命名空间
}

// GetDeploymentRequest 获取 Deployment 详情请求.
type GetDeploymentRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // Deployment 名称
}

// ScaleDeploymentRequest 扩缩容 Deployment 请求.
type ScaleDeploymentRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // Deployment 名称
	Replicas  int32  `json:"replicas" binding:"required"`  // 副本数
}

// RestartDeploymentRequest 重启 Deployment 请求.
type RestartDeploymentRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // Deployment 名称
}

// ===========================
// Node 管理请求
// ===========================

// ListNodesRequest 获取 Node 列表请求.
type ListNodesRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
}

// GetNodeRequest 获取 Node 详情请求.
type GetNodeRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Name      string `form:"name" binding:"required"`      // Node 名称
}

// CordonNodeRequest 标记 Node 不可调度请求.
type CordonNodeRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Name      string `form:"name" binding:"required"`      // Node 名称
}

// UncordonNodeRequest 标记 Node 可调度请求.
type UncordonNodeRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Name      string `form:"name" binding:"required"`      // Node 名称
}

// DrainNodeRequest 驱逐 Node 上的 Pod 请求.
type DrainNodeRequest struct {
	ClusterID        string `form:"clusterId" binding:"required"` // 集群 ID
	Name             string `form:"name" binding:"required"`      // Node 名称
	GracePeriod      int32  `json:"gracePeriod"`                  // 优雅终止时间（秒）
	IgnoreDaemonSets bool   `json:"ignoreDaemonSets"`             // 是否忽略 DaemonSet
	Force            bool   `json:"force"`                        // 是否强制驱逐
}

// ===========================
// Service 管理请求
// ===========================

// ListServicesRequest 获取 Service 列表请求.
type ListServicesRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace"`                    // 命名空间
}

// GetServiceRequest 获取 Service 详情请求.
type GetServiceRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // Service 名称
}

// CreateServiceRequest 创建 Service 请求.
type CreateServiceRequest struct {
	ClusterID string                 `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string                 `form:"namespace" binding:"required"` // 命名空间
	Spec      map[string]interface{} `json:"spec" binding:"required"`      // Service 规范
}

// UpdateServiceRequest 更新 Service 请求.
type UpdateServiceRequest struct {
	ClusterID string                 `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string                 `form:"namespace" binding:"required"` // 命名空间
	Name      string                 `form:"name" binding:"required"`      // Service 名称
	Spec      map[string]interface{} `json:"spec" binding:"required"`      // Service 规范
}

// DeleteServiceRequest 删除 Service 请求.
type DeleteServiceRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // Service 名称
}

// ===========================
// StatefulSet 管理请求
// ===========================

// ListStatefulSetsRequest 获取 StatefulSet 列表请求.
type ListStatefulSetsRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
}

// GetStatefulSetRequest 获取 StatefulSet 详情请求.
type GetStatefulSetRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // StatefulSet 名称
}

// ScaleStatefulSetRequest 扩缩容 StatefulSet 请求.
type ScaleStatefulSetRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // StatefulSet 名称
	Replicas  int32  `json:"replicas" binding:"required"`  // 副本数
}

// RestartStatefulSetRequest 重启 StatefulSet 请求.
type RestartStatefulSetRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // StatefulSet 名称
}

// DeleteStatefulSetRequest 删除 StatefulSet 请求.
type DeleteStatefulSetRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // StatefulSet 名称
}

// ===========================
// DaemonSet 管理请求
// ===========================

// ListDaemonSetsRequest 获取 DaemonSet 列表请求.
type ListDaemonSetsRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
}

// GetDaemonSetRequest 获取 DaemonSet 详情请求.
type GetDaemonSetRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // DaemonSet 名称
}

// RestartDaemonSetRequest 重启 DaemonSet 请求.
type RestartDaemonSetRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // DaemonSet 名称
}

// DeleteDaemonSetRequest 删除 DaemonSet 请求.
type DeleteDaemonSetRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // DaemonSet 名称
}

// ===========================
// ConfigMap 管理请求
// ===========================

// ListConfigMapsRequest 获取 ConfigMap 列表请求.
type ListConfigMapsRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
}

// GetConfigMapRequest 获取 ConfigMap 详情请求.
type GetConfigMapRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // ConfigMap 名称
}

// CreateConfigMapRequest 创建 ConfigMap 请求.
type CreateConfigMapRequest struct {
	ClusterID string            `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string            `form:"namespace" binding:"required"` // 命名空间
	Name      string            `json:"name" binding:"required"`      // ConfigMap 名称
	Data      map[string]string `json:"data"`                         // 数据
}

// UpdateConfigMapRequest 更新 ConfigMap 请求.
type UpdateConfigMapRequest struct {
	ClusterID string            `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string            `form:"namespace" binding:"required"` // 命名空间
	Name      string            `form:"name" binding:"required"`      // ConfigMap 名称
	Data      map[string]string `json:"data" binding:"required"`      // 数据
}

// DeleteConfigMapRequest 删除 ConfigMap 请求.
type DeleteConfigMapRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // ConfigMap 名称
}

// ===========================
// Secret 管理请求
// ===========================

// ListSecretsRequest 获取 Secret 列表请求.
type ListSecretsRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
}

// GetSecretRequest 获取 Secret 详情请求.
type GetSecretRequest struct {
	ClusterID   string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace   string `form:"namespace" binding:"required"` // 命名空间
	Name        string `form:"name" binding:"required"`      // Secret 名称
	IncludeData bool   `form:"includeData"`                  // 是否包含敏感数据
}

// CreateSecretRequest 创建 Secret 请求.
type CreateSecretRequest struct {
	ClusterID string            `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string            `form:"namespace" binding:"required"` // 命名空间
	Name      string            `json:"name" binding:"required"`      // Secret 名称
	Type      string            `json:"type"`                         // Secret 类型
	Data      map[string]string `json:"data"`                         // 数据（base64 编码）
}

// UpdateSecretRequest 更新 Secret 请求.
type UpdateSecretRequest struct {
	ClusterID string            `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string            `form:"namespace" binding:"required"` // 命名空间
	Name      string            `form:"name" binding:"required"`      // Secret 名称
	Data      map[string]string `json:"data" binding:"required"`      // 数据（base64 编码）
}

// DeleteSecretRequest 删除 Secret 请求.
type DeleteSecretRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // Secret 名称
}

// ===========================
// Endpoint 管理请求
// ===========================

// ListEndpointsRequest 获取 Endpoint 列表请求.
type ListEndpointsRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
}

// GetEndpointRequest 获取 Endpoint 详情请求.
type GetEndpointRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // Endpoint 名称
}

// DeleteEndpointRequest 删除 Endpoint 请求.
type DeleteEndpointRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // Endpoint 名称
}

// ===========================
// PVC 管理请求
// ===========================

// ListPVCsRequest 获取 PVC 列表请求.
type ListPVCsRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
}

// GetPVCRequest 获取 PVC 详情请求.
type GetPVCRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // PVC 名称
}

// DeletePVCRequest 删除 PVC 请求.
type DeletePVCRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // PVC 名称
}

// ===========================
// PV 管理请求
// ===========================

// ListPVsRequest 获取 PV 列表请求.
type ListPVsRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
}

// GetPVRequest 获取 PV 详情请求.
type GetPVRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Name      string `form:"name" binding:"required"`      // PV 名称
}

// DeletePVRequest 删除 PV 请求.
type DeletePVRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Name      string `form:"name" binding:"required"`      // PV 名称
}

// ===========================
// EndpointSlice 管理请求
// ===========================

// ListEndpointSlicesRequest 获取 EndpointSlice 列表请求.
type ListEndpointSlicesRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
}

// GetEndpointSliceRequest 获取 EndpointSlice 详情请求.
type GetEndpointSliceRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // EndpointSlice 名称
}

// DeleteEndpointSliceRequest 删除 EndpointSlice 请求.
type DeleteEndpointSliceRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // EndpointSlice 名称
}

// ===========================
// HPA 管理请求
// ===========================

// ListHPAsRequest 获取 HPA 列表请求.
type ListHPAsRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
}

// GetHPARequest 获取 HPA 详情请求.
type GetHPARequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // HPA 名称
}

// DeleteHPARequest 删除 HPA 请求.
type DeleteHPARequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // HPA 名称
}

// ===========================
// Event 管理请求
// ===========================

// ListEventsRequest 获取 Event 列表请求.
type ListEventsRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace"`                    // 命名空间（可选，不传则查询所有命名空间）
	Type      string `form:"type"`                         // 事件类型过滤
}

// GetEventRequest 获取 Event 详情请求.
type GetEventRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // Event 名称
}

// ===========================
// RoleBinding 管理请求
// ===========================

// ListRoleBindingsRequest 获取 RoleBinding 列表请求.
type ListRoleBindingsRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
}

// GetRoleBindingRequest 获取 RoleBinding 详情请求.
type GetRoleBindingRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // RoleBinding 名称
}

// DeleteRoleBindingRequest 删除 RoleBinding 请求.
type DeleteRoleBindingRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // RoleBinding 名称
}

// ===========================
// ClusterRole 管理请求
// ===========================

// ListClusterRolesRequest 获取 ClusterRole 列表请求.
type ListClusterRolesRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
}

// GetClusterRoleRequest 获取 ClusterRole 详情请求.
type GetClusterRoleRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Name      string `form:"name" binding:"required"`      // ClusterRole 名称
}

// DeleteClusterRoleRequest 删除 ClusterRole 请求.
type DeleteClusterRoleRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Name      string `form:"name" binding:"required"`      // ClusterRole 名称
}

// ===========================
// PriorityClass 管理请求
// ===========================

// ListPriorityClassesRequest 获取 PriorityClass 列表请求.
type ListPriorityClassesRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
}

// GetPriorityClassRequest 获取 PriorityClass 详情请求.
type GetPriorityClassRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Name      string `form:"name" binding:"required"`      // PriorityClass 名称
}

// DeletePriorityClassRequest 删除 PriorityClass 请求.
type DeletePriorityClassRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Name      string `form:"name" binding:"required"`      // PriorityClass 名称
}

// ===========================
// Role 管理请求
// ===========================

// ListRolesRequest 获取 Role 列表请求.
type ListRolesRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
}

// GetRoleRequest 获取 Role 详情请求.
type GetRoleRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // Role 名称
}

// DeleteRoleRequest 删除 Role 请求.
type DeleteRoleRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // Role 名称
}

// ===========================
// StorageClass 管理请求
// ===========================

// ListStorageClassesRequest 获取 StorageClass 列表请求.
type ListStorageClassesRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
}

// GetStorageClassRequest 获取 StorageClass 详情请求.
type GetStorageClassRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Name      string `form:"name" binding:"required"`      // StorageClass 名称
}

// DeleteStorageClassRequest 删除 StorageClass 请求.
type DeleteStorageClassRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Name      string `form:"name" binding:"required"`      // StorageClass 名称
}

// ===========================
// Job 管理请求
// ===========================

// ListJobsRequest 获取 Job 列表请求.
type ListJobsRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
}

// GetJobRequest 获取 Job 详情请求.
type GetJobRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // Job 名称
}

// DeleteJobRequest 删除 Job 请求.
type DeleteJobRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // Job 名称
}

// ===========================
// CronJob 管理请求
// ===========================

// ListCronJobsRequest 获取 CronJob 列表请求.
type ListCronJobsRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
}

// GetCronJobRequest 获取 CronJob 详情请求.
type GetCronJobRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // CronJob 名称
}

// DeleteCronJobRequest 删除 CronJob 请求.
type DeleteCronJobRequest struct {
	ClusterID string `form:"clusterId" binding:"required"` // 集群 ID
	Namespace string `form:"namespace" binding:"required"` // 命名空间
	Name      string `form:"name" binding:"required"`      // CronJob 名称
}
