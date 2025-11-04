package api

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/middleware"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	commonserver "github.com/kart-io/k8s-agent/common/server"
	httpserver "github.com/kart-io/k8s-agent/common/server/http"
	handlerpkg "github.com/kart-io/k8s-agent/internal/cluster/handler"
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
	handler        *handlerpkg.ClusterHandler
	k8sAPIHandler  *handlerpkg.K8sAPIHandler
	versionHandler *handlerpkg.VersionHandler
	log            core.Logger
	ginServer      commonserver.Server // 使用 common/server 的 Server 接口
}

// NewServer 创建新的服务器实例，使用 common/server 框架
func NewServer(
	config *ServerConfig,
	handler *handlerpkg.ClusterHandler,
	k8sAPIHandler *handlerpkg.K8sAPIHandler,
	logger core.Logger,
) *Server {
	// 创建 common/server 的配置
	serverOpts := &commonoptions.ServerOptions{
		Host:         "",
		Port:         config.Port,
		Mode:         config.Mode,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// 创建 Gin 服务器配置
	ginConfig := httpserver.NewGinServerConfig(serverOpts)

	// 创建 Gin 服务器
	ginServer := httpserver.NewGinServerFromFullConfig(logger, ginConfig)
	engine := ginServer.GetEngine()

	// 设置额外的中间件（除了 common/server 已经添加的）
	engine.Use(middleware.RateLimitByIP(100, 200)) // 每秒 100 个请求，桶容量 200
	engine.Use(middleware.Timeout(30 * time.Second))

	s := &Server{
		config:         config,
		handler:        handler,
		k8sAPIHandler:  k8sAPIHandler,
		versionHandler: handlerpkg.NewVersionHandler(),
		log:            logger,
		ginServer:      ginServer,
	}

	// 设置路由
	s.setupRoutes(engine)

	return s
}

// setupRoutes 设置 K8s API 路由
func (s *Server) setupRoutes(engine *gin.Engine) {
	// Health check
	engine.GET("/health", s.handler.HealthCheck)

	// Version endpoints
	version := engine.Group("/version")
	{
		version.GET("", s.versionHandler.GetVersion)              // 完整版本信息
		version.GET("/simple", s.versionHandler.GetVersionSimple) // 简化版本信息
		version.GET("/text", s.versionHandler.GetVersionText)     // 文本格式
		version.GET("/json", s.versionHandler.GetVersionJSON)     // JSON 格式
	}

	// K8s API 路由 - 扁平化路由，使用查询参数
	k8sAPI := engine.Group("/api/k8s")
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
		// 工作负载资源
		// ===========================
		// Pods
		k8sAPI.GET("/pods", s.k8sAPIHandler.ListPods)       // 获取 Pod 列表 ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/pod", s.k8sAPIHandler.GetPod)          // 获取 Pod 详情 ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/pod", s.k8sAPIHandler.DeletePod)    // 删除 Pod ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.GET("/pod/logs", s.k8sAPIHandler.GetPodLogs) // 获取 Pod 日志 ?clusterId=xxx&namespace=xxx&name=xxx&container=xxx

		// Deployments
		k8sAPI.GET("/deployments", s.k8sAPIHandler.ListDeployments)          // 获取 Deployment 列表
		k8sAPI.GET("/deployment", s.k8sAPIHandler.GetDeployment)             // 获取 Deployment 详情
		k8sAPI.PUT("/deployment/scale", s.k8sAPIHandler.ScaleDeployment)     // 扩缩容 Deployment
		k8sAPI.PUT("/deployment/restart", s.k8sAPIHandler.RestartDeployment) // 重启 Deployment

		// StatefulSets
		k8sAPI.GET("/statefulsets", s.k8sAPIHandler.ListStatefulSets)          // 获取 StatefulSet 列表
		k8sAPI.GET("/statefulset", s.k8sAPIHandler.GetStatefulSet)             // 获取 StatefulSet 详情
		k8sAPI.DELETE("/statefulset", s.k8sAPIHandler.DeleteStatefulSet)       // 删除 StatefulSet
		k8sAPI.PUT("/statefulset/scale", s.k8sAPIHandler.ScaleStatefulSet)     // 扩缩容 StatefulSet
		k8sAPI.PUT("/statefulset/restart", s.k8sAPIHandler.RestartStatefulSet) // 重启 StatefulSet

		// DaemonSets
		k8sAPI.GET("/daemonsets", s.k8sAPIHandler.ListDaemonSets)          // 获取 DaemonSet 列表
		k8sAPI.GET("/daemonset", s.k8sAPIHandler.GetDaemonSet)             // 获取 DaemonSet 详情
		k8sAPI.DELETE("/daemonset", s.k8sAPIHandler.DeleteDaemonSet)       // 删除 DaemonSet
		k8sAPI.PUT("/daemonset/restart", s.k8sAPIHandler.RestartDaemonSet) // 重启 DaemonSet

		// ReplicaSets
		k8sAPI.GET("/replicasets", s.k8sAPIHandler.ListReplicaSets)      // 获取 ReplicaSet 列表
		k8sAPI.GET("/replicaset", s.k8sAPIHandler.GetReplicaSet)         // 获取 ReplicaSet 详情
		k8sAPI.DELETE("/replicaset", s.k8sAPIHandler.DeleteReplicaSet)   // 删除 ReplicaSet
		k8sAPI.PUT("/replicaset/scale", s.k8sAPIHandler.ScaleReplicaSet) // 扩缩容 ReplicaSet

		// Jobs
		k8sAPI.GET("/jobs", s.k8sAPIHandler.ListJobs)    // 获取 Job 列表
		k8sAPI.GET("/job", s.k8sAPIHandler.GetJob)       // 获取 Job 详情
		k8sAPI.POST("/jobs", s.k8sAPIHandler.CreateJob)  // 创建 Job
		k8sAPI.DELETE("/job", s.k8sAPIHandler.DeleteJob) // 删除 Job

		// CronJobs
		k8sAPI.GET("/cronjobs", s.k8sAPIHandler.ListCronJobs)    // 获取 CronJob 列表
		k8sAPI.GET("/cronjob", s.k8sAPIHandler.GetCronJob)       // 获取 CronJob 详情
		k8sAPI.POST("/cronjobs", s.k8sAPIHandler.CreateCronJob)  // 创建 CronJob
		k8sAPI.PUT("/cronjob", s.k8sAPIHandler.UpdateCronJob)    // 更新 CronJob
		k8sAPI.DELETE("/cronjob", s.k8sAPIHandler.DeleteCronJob) // 删除 CronJob

		// ===========================
		// 服务与发现
		// ===========================
		// Services
		k8sAPI.GET("/services", s.k8sAPIHandler.ListServices)    // 获取 Service 列表
		k8sAPI.GET("/service", s.k8sAPIHandler.GetService)       // 获取 Service 详情
		k8sAPI.POST("/services", s.k8sAPIHandler.CreateService)  // 创建 Service
		k8sAPI.PUT("/service", s.k8sAPIHandler.UpdateService)    // 更新 Service
		k8sAPI.DELETE("/service", s.k8sAPIHandler.DeleteService) // 删除 Service

		// Endpoints
		k8sAPI.GET("/endpoints", s.k8sAPIHandler.ListEndpoints) // 获取 Endpoint 列表
		k8sAPI.GET("/endpoint", s.k8sAPIHandler.GetEndpoint)    // 获取 Endpoint 详情

		// EndpointSlices
		k8sAPI.GET("/endpointslices", s.k8sAPIHandler.ListEndpointSlices) // 获取 EndpointSlice 列表
		k8sAPI.GET("/endpointslice", s.k8sAPIHandler.GetEndpointSlice)    // 获取 EndpointSlice 详情

		// Ingresses
		k8sAPI.GET("/ingresses", s.k8sAPIHandler.ListIngresses)  // 获取 Ingress 列表
		k8sAPI.GET("/ingress", s.k8sAPIHandler.GetIngress)       // 获取 Ingress 详情
		k8sAPI.POST("/ingresses", s.k8sAPIHandler.CreateIngress) // 创建 Ingress
		k8sAPI.PUT("/ingress", s.k8sAPIHandler.UpdateIngress)    // 更新 Ingress
		k8sAPI.DELETE("/ingress", s.k8sAPIHandler.DeleteIngress) // 删除 Ingress

		// NetworkPolicies
		k8sAPI.GET("/networkpolicies", s.k8sAPIHandler.ListNetworkPolicies)  // 获取 NetworkPolicy 列表
		k8sAPI.GET("/networkpolicy", s.k8sAPIHandler.GetNetworkPolicy)       // 获取 NetworkPolicy 详情
		k8sAPI.POST("/networkpolicies", s.k8sAPIHandler.CreateNetworkPolicy) // 创建 NetworkPolicy
		k8sAPI.PUT("/networkpolicy", s.k8sAPIHandler.UpdateNetworkPolicy)    // 更新 NetworkPolicy
		k8sAPI.DELETE("/networkpolicy", s.k8sAPIHandler.DeleteNetworkPolicy) // 删除 NetworkPolicy

		// ===========================
		// 配置与存储
		// ===========================
		// ConfigMaps
		k8sAPI.GET("/configmaps", s.k8sAPIHandler.ListConfigMaps)    // 获取 ConfigMap 列表
		k8sAPI.GET("/configmap", s.k8sAPIHandler.GetConfigMap)       // 获取 ConfigMap 详情
		k8sAPI.POST("/configmaps", s.k8sAPIHandler.CreateConfigMap)  // 创建 ConfigMap
		k8sAPI.PUT("/configmap", s.k8sAPIHandler.UpdateConfigMap)    // 更新 ConfigMap
		k8sAPI.DELETE("/configmap", s.k8sAPIHandler.DeleteConfigMap) // 删除 ConfigMap

		// Secrets
		k8sAPI.GET("/secrets", s.k8sAPIHandler.ListSecrets)    // 获取 Secret 列表
		k8sAPI.GET("/secret", s.k8sAPIHandler.GetSecret)       // 获取 Secret 详情
		k8sAPI.POST("/secrets", s.k8sAPIHandler.CreateSecret)  // 创建 Secret
		k8sAPI.PUT("/secret", s.k8sAPIHandler.UpdateSecret)    // 更新 Secret
		k8sAPI.DELETE("/secret", s.k8sAPIHandler.DeleteSecret) // 删除 Secret

		// PersistentVolumeClaims
		k8sAPI.GET("/pvcs", s.k8sAPIHandler.ListPVCs)    // 获取 PVC 列表
		k8sAPI.GET("/pvc", s.k8sAPIHandler.GetPVC)       // 获取 PVC 详情
		k8sAPI.DELETE("/pvc", s.k8sAPIHandler.DeletePVC) // 删除 PVC

		// PersistentVolumes
		k8sAPI.GET("/pvs", s.k8sAPIHandler.ListPVs) // 获取 PV 列表
		k8sAPI.GET("/pv", s.k8sAPIHandler.GetPV)    // 获取 PV 详情

		// StorageClasses
		k8sAPI.GET("/storageclasses", s.k8sAPIHandler.ListStorageClasses) // 获取 StorageClass 列表
		k8sAPI.GET("/storageclass", s.k8sAPIHandler.GetStorageClass)      // 获取 StorageClass 详情

		// ===========================
		// 集群资源
		// ===========================
		// Nodes
		k8sAPI.GET("/nodes", s.k8sAPIHandler.ListNodes)            // 获取节点列表
		k8sAPI.GET("/node", s.k8sAPIHandler.GetNode)               // 获取节点详情
		k8sAPI.PUT("/node/cordon", s.k8sAPIHandler.CordonNode)     // 封锁节点
		k8sAPI.PUT("/node/uncordon", s.k8sAPIHandler.UncordonNode) // 解封节点
		k8sAPI.PUT("/node/drain", s.k8sAPIHandler.DrainNode)       // 驱逐节点

		// ===========================
		// 安全与 RBAC
		// ===========================
		// ServiceAccounts
		k8sAPI.GET("/serviceaccounts", s.k8sAPIHandler.ListServiceAccounts)    // 获取 ServiceAccount 列表
		k8sAPI.GET("/serviceaccount", s.k8sAPIHandler.GetServiceAccount)       // 获取 ServiceAccount 详情
		k8sAPI.DELETE("/serviceaccount", s.k8sAPIHandler.DeleteServiceAccount) // 删除 ServiceAccount

		// Roles
		k8sAPI.GET("/roles", s.k8sAPIHandler.ListRoles) // 获取 Role 列表
		k8sAPI.GET("/role", s.k8sAPIHandler.GetRole)    // 获取 Role 详情

		// RoleBindings
		k8sAPI.GET("/rolebindings", s.k8sAPIHandler.ListRoleBindings) // 获取 RoleBinding 列表
		k8sAPI.GET("/rolebinding", s.k8sAPIHandler.GetRoleBinding)    // 获取 RoleBinding 详情

		// ClusterRoles
		k8sAPI.GET("/clusterroles", s.k8sAPIHandler.ListClusterRoles) // 获取 ClusterRole 列表
		k8sAPI.GET("/clusterrole", s.k8sAPIHandler.GetClusterRole)    // 获取 ClusterRole 详情

		// ClusterRoleBindings
		k8sAPI.GET("/clusterrolebindings", s.k8sAPIHandler.ListClusterRoleBindings) // 获取 ClusterRoleBinding 列表
		k8sAPI.GET("/clusterrolebinding", s.k8sAPIHandler.GetClusterRoleBinding)    // 获取 ClusterRoleBinding 详情

		// ===========================
		// 监控与扩缩容
		// ===========================
		// HorizontalPodAutoscalers
		k8sAPI.GET("/hpas", s.k8sAPIHandler.ListHPAs)    // 获取 HPA 列表
		k8sAPI.GET("/hpa", s.k8sAPIHandler.GetHPA)       // 获取 HPA 详情
		k8sAPI.DELETE("/hpa", s.k8sAPIHandler.DeleteHPA) // 删除 HPA

		// Events
		k8sAPI.GET("/events", s.k8sAPIHandler.ListEvents) // 获取事件列表

		// ===========================
		// 资源管理
		// ===========================
		// LimitRanges
		k8sAPI.GET("/limitranges", s.k8sAPIHandler.ListLimitRanges) // 获取 LimitRange 列表
		k8sAPI.GET("/limitrange", s.k8sAPIHandler.GetLimitRange)    // 获取 LimitRange 详情

		// ResourceQuotas
		k8sAPI.GET("/resourcequotas", s.k8sAPIHandler.ListResourceQuotas) // 获取 ResourceQuota 列表
		k8sAPI.GET("/resourcequota", s.k8sAPIHandler.GetResourceQuota)    // 获取 ResourceQuota 详情

		// PriorityClasses
		k8sAPI.GET("/priorityclasses", s.k8sAPIHandler.ListPriorityClasses) // 获取 PriorityClass 列表
		k8sAPI.GET("/priorityclass", s.k8sAPIHandler.GetPriorityClass)      // 获取 PriorityClass 详情
	}

	if s.log != nil {
		s.log.Infow("Cluster service routes configured",
			"port", s.config.Port,
			"mode", s.config.Mode,
			"health", "/health",
			"api", "/api/k8s",
		)
	}
}

// Start 启动服务器 - 使用 common/server 框架
// 注意：这个方法现在由 initializer 或 bootstrap 调用
func (s *Server) Start() error {
	// 现在由 common/server 框架处理
	// 这个方法主要是为了向后兼容
	return s.Run(context.Background())
}

// Run 运行服务器 - 使用 common/server 框架
func (s *Server) Run(ctx context.Context) error {
	if s.log != nil {
		s.log.Infow("Starting cluster service with common/server framework",
			"port", s.config.Port,
			"mode", s.config.Mode,
		)
	}

	// 使用 common/server 的 Serve 方法来管理生命周期
	return commonserver.Serve(ctx, s.ginServer, s.log)
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown(ctx context.Context) error {
	if s.log != nil {
		s.log.Info("Cluster service shutting down")
	}

	// common/server 的 Serve 方法会自动处理优雅关闭
	// 这里主要是为了向后兼容
	return nil
}

// GetServer 返回底层的 common/server.Server 实例
func (s *Server) GetServer() commonserver.Server {
	return s.ginServer
}

// GetEngine 返回 Gin Engine 实例（用于测试或其他需要）
func (s *Server) GetEngine() *gin.Engine {
	if ginSrv, ok := s.ginServer.(*httpserver.GinServer); ok {
		return ginSrv.GetEngine()
	}
	return nil
}
