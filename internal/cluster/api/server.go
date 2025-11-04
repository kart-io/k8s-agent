package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/middleware"
	"github.com/kart-io/k8s-agent/internal/cluster/handler"
	handler2 "github.com/kart-io/k8s-agent/internal/cluster/handler"
	"github.com/kart-io/logger/core"
)

type ServerConfig struct {
	Port         int
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	JWTSecret    string
}

type Server struct {
	config         *ServerConfig
	handler        *handler.ClusterHandler
	k8sAPIHandler  *handler.K8sAPIHandler
	versionHandler *handler.VersionHandler
	log            core.Logger
	engine         *gin.Engine
	server         *http.Server
}

// NewServer 创建新的服务器实例
func NewServer(
	config *ServerConfig,
	handler *handler.ClusterHandler,
	k8sAPIHandler *handler.K8sAPIHandler,
	logger core.Logger,
) *Server {
	gin.SetMode(config.Mode)
	engine := gin.New()

	// 使用 common 包中的中间件
	engine.Use(middleware.Recovery())
	engine.Use(middleware.RequestID())
	engine.Use(middleware.RequestLogger())
	engine.Use(middleware.CORS())
	engine.Use(middleware.RateLimitByIP(100, 200)) // 每秒 100 个请求，桶容量 200
	engine.Use(middleware.Timeout(30 * time.Second))

	s := &Server{
		config:         config,
		handler:        handler,
		k8sAPIHandler:  k8sAPIHandler,
		versionHandler: handler2.NewVersionHandler(),
		log:            logger,
		engine:         engine,
	}

	s.setupRoutes()
	return s
}

// setupRoutes 设置 K8s API 路由
func (s *Server) setupRoutes() {
	// Health check
	s.engine.GET("/health", s.handler.HealthCheck)

	// Version endpoints
	version := s.engine.Group("/version")
	{
		version.GET("", s.versionHandler.GetVersion)              // 完整版本信息
		version.GET("/simple", s.versionHandler.GetVersionSimple) // 简化版本信息
		version.GET("/text", s.versionHandler.GetVersionText)     // 文本格式
		version.GET("/json", s.versionHandler.GetVersionJSON)     // JSON 格式
	}

	// K8s API 路由 - 扁平化路由，使用查询参数
	k8sAPI := s.engine.Group("/api/k8s")
	{
		// ===========================
		// 集群管理 API
		// ===========================
		k8sAPI.GET("/clusters", s.k8sAPIHandler.ListClusters)                 // 获取集群列表
		k8sAPI.GET("/clusters/options", s.k8sAPIHandler.GetClusterOptions)    // 获取集群选择器列表
		k8sAPI.POST("/clusters", s.k8sAPIHandler.CreateCluster)               // 创建集群
		k8sAPI.GET("/cluster", s.k8sAPIHandler.GetCluster)                    // 获取集群详情 ?clusterId=xxx
		k8sAPI.PUT("/cluster", s.k8sAPIHandler.UpdateCluster)                 // 更新集群 ?clusterId=xxx
		k8sAPI.DELETE("/cluster", s.k8sAPIHandler.DeleteCluster)              // 删除集群 ?clusterId=xxx
		k8sAPI.GET("/cluster/health", s.k8sAPIHandler.GetClusterHealthStatus) // 获取集群健康状态 ?clusterId=xxx

		// ===========================
		// 命名空间管理 API
		// ===========================
		k8sAPI.GET("/namespaces", s.k8sAPIHandler.ListNamespaces)    // 获取命名空间列表 ?clusterId=xxx
		k8sAPI.POST("/namespaces", s.k8sAPIHandler.CreateNamespace)  // 创建命名空间 ?clusterId=xxx
		k8sAPI.GET("/namespace", s.k8sAPIHandler.GetNamespace)       // 获取命名空间详情 ?clusterId=xxx&namespace=xxx
		k8sAPI.DELETE("/namespace", s.k8sAPIHandler.DeleteNamespace) // 删除命名空间 ?clusterId=xxx&namespace=xxx

		// ===========================
		// Pod 管理 API
		// ===========================
		k8sAPI.GET("/pods", s.k8sAPIHandler.ListPods)       // 获取 Pod 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/pod", s.k8sAPIHandler.GetPod)          // 获取 Pod 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/pod", s.k8sAPIHandler.DeletePod)    // 删除 Pod ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.GET("/pod/logs", s.k8sAPIHandler.GetPodLogs) // 获取 Pod 日志 ?clusterId=xxx&namespace=xxx&name=xxx&container=xxx

		// ===========================
		// Deployment 管理 API
		// ===========================
		k8sAPI.GET("/deployments", s.k8sAPIHandler.ListDeployments)           // 获取 Deployment 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/deployment", s.k8sAPIHandler.GetDeployment)              // 获取 Deployment 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.PUT("/deployment/scale", s.k8sAPIHandler.ScaleDeployment)      // 扩缩容 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.POST("/deployment/restart", s.k8sAPIHandler.RestartDeployment) // 重启 ?clusterId=xxx&namespace=xxx&name=xxx

		// ===========================
		// Node 管理 API
		// ===========================
		k8sAPI.GET("/nodes", s.k8sAPIHandler.ListNodes)             // 获取 Node 列表 ?clusterId=xxx
		k8sAPI.GET("/node", s.k8sAPIHandler.GetNode)                // 获取 Node 详情 ?clusterId=xxx&name=xxx
		k8sAPI.POST("/node/cordon", s.k8sAPIHandler.CordonNode)     // 标记为不可调度 ?clusterId=xxx&name=xxx
		k8sAPI.POST("/node/uncordon", s.k8sAPIHandler.UncordonNode) // 标记为可调度 ?clusterId=xxx&name=xxx
		k8sAPI.POST("/node/drain", s.k8sAPIHandler.DrainNode)       // 驱逐 Pod ?clusterId=xxx&name=xxx

		// ===========================
		// Service 管理 API
		// ===========================
		k8sAPI.GET("/services", s.k8sAPIHandler.ListServices)    // 获取 Service 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.POST("/services", s.k8sAPIHandler.CreateService)  // 创建 Service ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/service", s.k8sAPIHandler.GetService)       // 获取 Service 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.PUT("/service", s.k8sAPIHandler.UpdateService)    // 更新 Service ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/service", s.k8sAPIHandler.DeleteService) // 删除 Service ?clusterId=xxx&namespace=xxx&name=xxx

		// ===========================
		// StatefulSet 管理 API
		// ===========================
		k8sAPI.GET("/statefulsets", s.k8sAPIHandler.ListStatefulSets)           // 获取 StatefulSet 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/statefulset", s.k8sAPIHandler.GetStatefulSet)              // 获取 StatefulSet 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.PUT("/statefulset/scale", s.k8sAPIHandler.ScaleStatefulSet)      // 扩缩容 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.POST("/statefulset/restart", s.k8sAPIHandler.RestartStatefulSet) // 重启 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/statefulset", s.k8sAPIHandler.DeleteStatefulSet)        // 删除 ?clusterId=xxx&namespace=xxx&name=xxx

		// ===========================
		// DaemonSet 管理 API
		// ===========================
		k8sAPI.GET("/daemonsets", s.k8sAPIHandler.ListDaemonSets)           // 获取 DaemonSet 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/daemonset", s.k8sAPIHandler.GetDaemonSet)              // 获取 DaemonSet 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.POST("/daemonset/restart", s.k8sAPIHandler.RestartDaemonSet) // 重启 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/daemonset", s.k8sAPIHandler.DeleteDaemonSet)        // 删除 ?clusterId=xxx&namespace=xxx&name=xxx

		// ===========================
		// Job 管理 API
		// ===========================
		k8sAPI.GET("/jobs", s.k8sAPIHandler.ListJobs)    // 获取 Job 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/job", s.k8sAPIHandler.GetJob)       // 获取 Job 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.POST("/jobs", s.k8sAPIHandler.CreateJob)  // 创建 Job ?clusterId=xxx&namespace=xxx
		k8sAPI.DELETE("/job", s.k8sAPIHandler.DeleteJob) // 删除 Job ?clusterId=xxx&namespace=xxx&name=xxx

		// ===========================
		// CronJob 管理 API
		// ===========================
		k8sAPI.GET("/cronjobs", s.k8sAPIHandler.ListCronJobs)    // 获取 CronJob 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/cronjob", s.k8sAPIHandler.GetCronJob)       // 获取 CronJob 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.POST("/cronjobs", s.k8sAPIHandler.CreateCronJob)  // 创建 CronJob ?clusterId=xxx&namespace=xxx
		k8sAPI.PUT("/cronjob", s.k8sAPIHandler.UpdateCronJob)    // 更新 CronJob ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/cronjob", s.k8sAPIHandler.DeleteCronJob) // 删除 CronJob ?clusterId=xxx&namespace=xxx&name=xxx

		// ===========================
		// ConfigMap 管理 API
		// ===========================
		k8sAPI.GET("/configmaps", s.k8sAPIHandler.ListConfigMaps)    // 获取 ConfigMap 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.POST("/configmaps", s.k8sAPIHandler.CreateConfigMap)  // 创建 ConfigMap ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/configmap", s.k8sAPIHandler.GetConfigMap)       // 获取 ConfigMap 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.PUT("/configmap", s.k8sAPIHandler.UpdateConfigMap)    // 更新 ConfigMap ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/configmap", s.k8sAPIHandler.DeleteConfigMap) // 删除 ConfigMap ?clusterId=xxx&namespace=xxx&name=xxx

		// ===========================
		// Secret 管理 API
		// ===========================
		k8sAPI.GET("/secrets", s.k8sAPIHandler.ListSecrets)    // 获取 Secret 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.POST("/secrets", s.k8sAPIHandler.CreateSecret)  // 创建 Secret ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/secret", s.k8sAPIHandler.GetSecret)       // 获取 Secret 详情 ?clusterId=xxx&namespace=xxx&name=xxx&includeData=true
		k8sAPI.PUT("/secret", s.k8sAPIHandler.UpdateSecret)    // 更新 Secret ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/secret", s.k8sAPIHandler.DeleteSecret) // 删除 Secret ?clusterId=xxx&namespace=xxx&name=xxx

		// ===========================
		// Endpoints 管理 API
		// ===========================
		k8sAPI.GET("/endpoints", s.k8sAPIHandler.ListEndpoints)    // 获取 Endpoints 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/endpoint", s.k8sAPIHandler.GetEndpoint)       // 获取 Endpoint 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/endpoint", s.k8sAPIHandler.DeleteEndpoint) // 删除 Endpoint ?clusterId=xxx&namespace=xxx&name=xxx

		// ===========================
		// PersistentVolumeClaim 管理 API
		// ===========================
		k8sAPI.GET("/pvcs", s.k8sAPIHandler.ListPVCs)    // 获取 PVC 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/pvc", s.k8sAPIHandler.GetPVC)       // 获取 PVC 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/pvc", s.k8sAPIHandler.DeletePVC) // 删除 PVC ?clusterId=xxx&namespace=xxx&name=xxx

		// 别名路由（兼容完整资源名）
		k8sAPI.GET("/persistentvolumeclaims", s.k8sAPIHandler.ListPVCs)    // 别名：获取 PVC 列表
		k8sAPI.GET("/persistentvolumeclaim", s.k8sAPIHandler.GetPVC)       // 别名：获取 PVC 详情
		k8sAPI.DELETE("/persistentvolumeclaim", s.k8sAPIHandler.DeletePVC) // 别名：删除 PVC

		// ===========================
		// PersistentVolume 管理 API (cluster-scoped)
		// ===========================
		k8sAPI.GET("/pvs", s.k8sAPIHandler.ListPVs)    // 获取 PV 列表 ?clusterId=xxx
		k8sAPI.GET("/pv", s.k8sAPIHandler.GetPV)       // 获取 PV 详情 ?clusterId=xxx&name=xxx
		k8sAPI.DELETE("/pv", s.k8sAPIHandler.DeletePV) // 删除 PV ?clusterId=xxx&name=xxx

		// 别名路由（兼容完整资源名）
		k8sAPI.GET("/persistentvolumes", s.k8sAPIHandler.ListPVs)    // 别名：获取 PV 列表
		k8sAPI.GET("/persistentvolume", s.k8sAPIHandler.GetPV)       // 别名：获取 PV 详情
		k8sAPI.DELETE("/persistentvolume", s.k8sAPIHandler.DeletePV) // 别名：删除 PV

		// ===========================
		// EndpointSlice 管理 API
		// ===========================
		k8sAPI.GET("/endpointslices", s.k8sAPIHandler.ListEndpointSlices)    // 获取 EndpointSlice 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/endpointslice", s.k8sAPIHandler.GetEndpointSlice)       // 获取 EndpointSlice 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/endpointslice", s.k8sAPIHandler.DeleteEndpointSlice) // 删除 EndpointSlice ?clusterId=xxx&namespace=xxx&name=xxx

		// ===========================
		// HorizontalPodAutoscaler 管理 API
		// ===========================
		k8sAPI.GET("/hpas", s.k8sAPIHandler.ListHPAs)    // 获取 HPA 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/hpa", s.k8sAPIHandler.GetHPA)       // 获取 HPA 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/hpa", s.k8sAPIHandler.DeleteHPA) // 删除 HPA ?clusterId=xxx&namespace=xxx&name=xxx

		// 别名路由（兼容完整资源名）
		k8sAPI.GET("/horizontalpodautoscalers", s.k8sAPIHandler.ListHPAs)    // 别名：获取 HPA 列表
		k8sAPI.GET("/horizontalpodautoscaler", s.k8sAPIHandler.GetHPA)       // 别名：获取 HPA 详情
		k8sAPI.DELETE("/horizontalpodautoscaler", s.k8sAPIHandler.DeleteHPA) // 别名：删除 HPA

		// ===========================
		// Event 管理 API
		// ===========================
		k8sAPI.GET("/events", s.k8sAPIHandler.ListEvents) // 获取 Event 列表 ?clusterId=xxx&namespace=xxx&type=xxx
		k8sAPI.GET("/event", s.k8sAPIHandler.GetEvent)    // 获取 Event 详情 ?clusterId=xxx&namespace=xxx&name=xxx

		// ===========================
		// RoleBinding 管理 API
		// ===========================
		k8sAPI.GET("/rolebindings", s.k8sAPIHandler.ListRoleBindings)    // 获取 RoleBinding 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/rolebinding", s.k8sAPIHandler.GetRoleBinding)       // 获取 RoleBinding 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/rolebinding", s.k8sAPIHandler.DeleteRoleBinding) // 删除 RoleBinding ?clusterId=xxx&namespace=xxx&name=xxx

		// ===========================
		// ClusterRole 管理 API (cluster-scoped)
		// ===========================
		k8sAPI.GET("/clusterroles", s.k8sAPIHandler.ListClusterRoles)    // 获取 ClusterRole 列表 ?clusterId=xxx
		k8sAPI.GET("/clusterrole", s.k8sAPIHandler.GetClusterRole)       // 获取 ClusterRole 详情 ?clusterId=xxx&name=xxx
		k8sAPI.DELETE("/clusterrole", s.k8sAPIHandler.DeleteClusterRole) // 删除 ClusterRole ?clusterId=xxx&name=xxx

		// ===========================
		// PriorityClass 管理 API (cluster-scoped)
		// ===========================
		k8sAPI.GET("/priorityclasses", s.k8sAPIHandler.ListPriorityClasses)  // 获取 PriorityClass 列表 ?clusterId=xxx
		k8sAPI.GET("/priorityclass", s.k8sAPIHandler.GetPriorityClass)       // 获取 PriorityClass 详情 ?clusterId=xxx&name=xxx
		k8sAPI.DELETE("/priorityclass", s.k8sAPIHandler.DeletePriorityClass) // 删除 PriorityClass ?clusterId=xxx&name=xxx

		// ===========================
		// Role 管理 API (namespace-scoped)
		// ===========================
		k8sAPI.GET("/roles", s.k8sAPIHandler.ListRoles)    // 获取 Role 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/role", s.k8sAPIHandler.GetRole)       // 获取 Role 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/role", s.k8sAPIHandler.DeleteRole) // 删除 Role ?clusterId=xxx&namespace=xxx&name=xxx

		// ===========================
		// StorageClass 管理 API (cluster-scoped)
		// ===========================
		k8sAPI.GET("/storageclasses", s.k8sAPIHandler.ListStorageClasses)  // 获取 StorageClass 列表 ?clusterId=xxx
		k8sAPI.GET("/storageclass", s.k8sAPIHandler.GetStorageClass)       // 获取 StorageClass 详情 ?clusterId=xxx&name=xxx
		k8sAPI.DELETE("/storageclass", s.k8sAPIHandler.DeleteStorageClass) // 删除 StorageClass ?clusterId=xxx&name=xxx

		// ===========================
		// Ingress 管理 API (namespace-scoped)
		// ===========================
		k8sAPI.GET("/ingresses", s.k8sAPIHandler.ListIngresses)  // 获取 Ingress 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/ingress", s.k8sAPIHandler.GetIngress)       // 获取 Ingress 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.POST("/ingresses", s.k8sAPIHandler.CreateIngress) // 创建 Ingress ?clusterId=xxx&namespace=xxx
		k8sAPI.PUT("/ingress", s.k8sAPIHandler.UpdateIngress)    // 更新 Ingress ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/ingress", s.k8sAPIHandler.DeleteIngress) // 删除 Ingress ?clusterId=xxx&namespace=xxx&name=xxx

		// ===========================
		// NetworkPolicy 管理 API (namespace-scoped)
		// ===========================
		k8sAPI.GET("/networkpolicies", s.k8sAPIHandler.ListNetworkPolicies)  // 获取 NetworkPolicy 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/networkpolicy", s.k8sAPIHandler.GetNetworkPolicy)       // 获取 NetworkPolicy 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.POST("/networkpolicies", s.k8sAPIHandler.CreateNetworkPolicy) // 创建 NetworkPolicy ?clusterId=xxx&namespace=xxx
		k8sAPI.PUT("/networkpolicy", s.k8sAPIHandler.UpdateNetworkPolicy)    // 更新 NetworkPolicy ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/networkpolicy", s.k8sAPIHandler.DeleteNetworkPolicy) // 删除 NetworkPolicy ?clusterId=xxx&namespace=xxx&name=xxx

		// ===========================
		// ReplicaSet 管理 API (namespace-scoped)
		// ===========================
		k8sAPI.GET("/replicasets", s.k8sAPIHandler.ListReplicaSets)      // 获取 ReplicaSet 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/replicaset", s.k8sAPIHandler.GetReplicaSet)         // 获取 ReplicaSet 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.PUT("/replicaset/scale", s.k8sAPIHandler.ScaleReplicaSet) // 扩缩容 ReplicaSet ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/replicaset", s.k8sAPIHandler.DeleteReplicaSet)   // 删除 ReplicaSet ?clusterId=xxx&namespace=xxx&name=xxx

		// ===========================
		// LimitRange 管理 API (namespace-scoped)
		// ===========================
		k8sAPI.GET("/limitranges", s.k8sAPIHandler.ListLimitRanges)    // 获取 LimitRange 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/limitrange", s.k8sAPIHandler.GetLimitRange)       // 获取 LimitRange 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/limitrange", s.k8sAPIHandler.DeleteLimitRange) // 删除 LimitRange ?clusterId=xxx&namespace=xxx&name=xxx

		// ===========================
		// ServiceAccount 管理 API (namespace-scoped)
		// ===========================
		k8sAPI.GET("/serviceaccounts", s.k8sAPIHandler.ListServiceAccounts)    // 获取 ServiceAccount 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/serviceaccount", s.k8sAPIHandler.GetServiceAccount)       // 获取 ServiceAccount 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/serviceaccount", s.k8sAPIHandler.DeleteServiceAccount) // 删除 ServiceAccount ?clusterId=xxx&namespace=xxx&name=xxx

		// ===========================
		// ClusterRoleBinding 管理 API (cluster-scoped)
		// ===========================
		k8sAPI.GET("/clusterrolebindings", s.k8sAPIHandler.ListClusterRoleBindings)    // 获取 ClusterRoleBinding 列表 ?clusterId=xxx
		k8sAPI.GET("/clusterrolebinding", s.k8sAPIHandler.GetClusterRoleBinding)       // 获取 ClusterRoleBinding 详情 ?clusterId=xxx&name=xxx
		k8sAPI.DELETE("/clusterrolebinding", s.k8sAPIHandler.DeleteClusterRoleBinding) // 删除 ClusterRoleBinding ?clusterId=xxx&name=xxx

		// ===========================
		// ResourceQuota 管理 API (namespace-scoped)
		// ===========================
		k8sAPI.GET("/resourcequotas", s.k8sAPIHandler.ListResourceQuotas)    // 获取 ResourceQuota 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/resourcequota", s.k8sAPIHandler.GetResourceQuota)       // 获取 ResourceQuota 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/resourcequota", s.k8sAPIHandler.DeleteResourceQuota) // 删除 ResourceQuota ?clusterId=xxx&namespace=xxx&name=xxx
	}

	// 记录注册的路由
	if s.log != nil {
		s.log.Infow("K8s API routes registered (query parameter style)",
			"base_path", "/api/k8s",
			"style", "query_parameters",
			"cluster_endpoints", 7,
			"namespace_endpoints", 4,
			"pod_endpoints", 4,
			"deployment_endpoints", 4,
			"node_endpoints", 5,
			"service_endpoints", 5,
			"statefulset_endpoints", 5,
			"daemonset_endpoints", 4,
			"configmap_endpoints", 5,
			"secret_endpoints", 5,
			"endpoint_endpoints", 3,
			"pvc_endpoints", 3,
			"pv_endpoints", 3,
			"endpointslice_endpoints", 3,
			"hpa_endpoints", 3,
			"event_endpoints", 2,
			"rolebinding_endpoints", 3,
			"clusterrole_endpoints", 3,
			"priorityclass_endpoints", 3,
			"role_endpoints", 3,
			"storageclass_endpoints", 3,
		)
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.engine,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	if s.log != nil {
		s.log.Infow("Server starting",
			"port", s.config.Port,
			"mode", s.config.Mode,
			"read_timeout", s.config.ReadTimeout,
			"write_timeout", s.config.WriteTimeout,
		)
	}

	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.log != nil {
		s.log.Info("Server shutdown initiated")
	}

	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}
