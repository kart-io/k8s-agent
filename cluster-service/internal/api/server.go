package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/cluster-service/internal/handler"
	handler2 "github.com/kart-io/k8s-agent/cluster-service/internal/handler"
	"github.com/kart-io/k8s-agent/common/logger"
	"github.com/kart-io/k8s-agent/common/middleware"
	"github.com/sirupsen/logrus"
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
	log            *logrus.Logger
	engine         *gin.Engine
	server         *http.Server
}

// NewServer 创建新的服务器实例（保持向后兼容）
func NewServer(config *ServerConfig, handler *handler.ClusterHandler, logger *logrus.Logger) *Server {
	gin.SetMode(config.Mode)
	engine := gin.New()

	// 使用基本中间件
	engine.Use(gin.Recovery())
	engine.Use(ginLogger(logger))

	s := &Server{
		config:         config,
		handler:        handler,
		versionHandler: handler2.NewVersionHandler(),
		log:            logger,
		engine:         engine,
	}

	s.setupRoutes()
	return s
}

// NewServerWithK8sAPI 创建支持完整 K8s API 的服务器实例
func NewServerWithK8sAPI(
	config *ServerConfig,
	handler *handler.ClusterHandler,
	k8sAPIHandler *handler.K8sAPIHandler,
	logger *logrus.Logger,
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

	s.setupK8sAPIRoutes()
	return s
}

// setupRoutes 设置原有的路由（保持向后兼容）
func (s *Server) setupRoutes() {
	// Health check
	s.engine.GET("/health", s.handler.HealthCheck)

	// Version endpoints
	version := s.engine.Group("/version")
	{
		version.GET("", s.versionHandler.GetVersion)            // 完整版本信息
		version.GET("/simple", s.versionHandler.GetVersionSimple) // 简化版本信息
		version.GET("/text", s.versionHandler.GetVersionText)   // 文本格式
		version.GET("/json", s.versionHandler.GetVersionJSON)   // JSON 格式
	}

	// API routes
	v1 := s.engine.Group("/api/v1")
	{
		// Cluster routes
		clusters := v1.Group("/clusters")
		{
			clusters.POST("", s.handler.AddCluster)
			clusters.GET("/:clusterId/health", s.handler.GetClusterHealth)

			// Pod routes
			clusters.GET("/:clusterId/namespaces/:namespace/pods", s.handler.GetPods)
		}
	}
}

// setupK8sAPIRoutes 设置完整的 K8s API 路由
func (s *Server) setupK8sAPIRoutes() {
	// Health check
	s.engine.GET("/health", s.handler.HealthCheck)

	// Version endpoints
	version := s.engine.Group("/version")
	{
		version.GET("", s.versionHandler.GetVersion)            // 完整版本信息
		version.GET("/simple", s.versionHandler.GetVersionSimple) // 简化版本信息
		version.GET("/text", s.versionHandler.GetVersionText)   // 文本格式
		version.GET("/json", s.versionHandler.GetVersionJSON)   // JSON 格式
	}

	// 原有的 v1 API（保持向后兼容）
	v1 := s.engine.Group("/api/v1")
	{
		clusters := v1.Group("/clusters")
		{
			clusters.POST("", s.handler.AddCluster)
			clusters.GET("/:clusterId/health", s.handler.GetClusterHealth)
			clusters.GET("/:clusterId/namespaces/:namespace/pods", s.handler.GetPods)
		}
	}

	// 新的 K8s API 路由 - 基于 /api/k8s 路径
	k8sAPI := s.engine.Group("/api/k8s")
	{
		// ===========================
		// 集群管理 API
		// ===========================
		clusters := k8sAPI.Group("/clusters")
		{
			clusters.GET("", s.k8sAPIHandler.ListClusters)                     // 获取集群列表
			clusters.GET("/options", s.k8sAPIHandler.GetClusterOptions)        // 获取集群选择器列表
			clusters.POST("", s.k8sAPIHandler.CreateCluster)                   // 创建集群
			clusters.GET("/:clusterId", s.k8sAPIHandler.GetCluster)            // 获取集群详情
			clusters.PUT("/:clusterId", s.k8sAPIHandler.UpdateCluster)         // 更新集群
			clusters.DELETE("/:clusterId", s.k8sAPIHandler.DeleteCluster)      // 删除集群
			clusters.GET("/:clusterId/health", s.k8sAPIHandler.GetClusterHealthStatus) // 获取集群健康状态

			// 命名空间管理 API - 放在集群组下
			clusters.GET("/:clusterId/namespaces", s.k8sAPIHandler.ListNamespaces)          // 获取命名空间列表
			clusters.POST("/:clusterId/namespaces", s.k8sAPIHandler.CreateNamespace)        // 创建命名空间
			clusters.GET("/:clusterId/ns/:namespace", s.k8sAPIHandler.GetNamespace)         // 获取命名空间详情 (使用 ns 避免冲突)
			clusters.DELETE("/:clusterId/ns/:namespace", s.k8sAPIHandler.DeleteNamespace)   // 删除命名空间 (使用 ns 避免冲突)
		}

		// ===========================
		// Pod 管理 API
		// ===========================
		pods := k8sAPI.Group("/clusters/:clusterId/namespaces/:namespace/pods")
		{
			pods.GET("", s.k8sAPIHandler.ListPods)              // 获取 Pod 列表
			pods.GET("/:name", s.k8sAPIHandler.GetPod)          // 获取 Pod 详情
			pods.DELETE("/:name", s.k8sAPIHandler.DeletePod)    // 删除 Pod
			pods.GET("/:name/logs", s.k8sAPIHandler.GetPodLogs) // 获取 Pod 日志
		}

		// ===========================
		// Deployment 管理 API
		// ===========================
		deployments := k8sAPI.Group("/clusters/:clusterId/namespaces/:namespace/deployments")
		{
			deployments.GET("", s.k8sAPIHandler.ListDeployments)                    // 获取 Deployment 列表
			deployments.GET("/:name", s.k8sAPIHandler.GetDeployment)                // 获取 Deployment 详情
			deployments.PUT("/:name/scale", s.k8sAPIHandler.ScaleDeployment)        // 扩缩容
			deployments.POST("/:name/restart", s.k8sAPIHandler.RestartDeployment)   // 重启
		}

		// ===========================
		// Node 管理 API
		// ===========================
		nodes := k8sAPI.Group("/clusters/:clusterId/nodes")
		{
			nodes.GET("", s.k8sAPIHandler.ListNodes)                    // 获取 Node 列表
			nodes.GET("/:name", s.k8sAPIHandler.GetNode)                // 获取 Node 详情
			nodes.POST("/:name/cordon", s.k8sAPIHandler.CordonNode)     // 标记为不可调度
			nodes.POST("/:name/uncordon", s.k8sAPIHandler.UncordonNode) // 标记为可调度
			nodes.POST("/:name/drain", s.k8sAPIHandler.DrainNode)       // 驱逐 Pod
		}

		// ===========================
		// Service 管理 API
		// ===========================
		services := k8sAPI.Group("/clusters/:clusterId/namespaces/:namespace/services")
		{
			services.GET("", s.k8sAPIHandler.ListServices)             // 获取 Service 列表
			services.POST("", s.k8sAPIHandler.CreateService)           // 创建 Service
			services.GET("/:name", s.k8sAPIHandler.GetService)         // 获取 Service 详情
			services.PUT("/:name", s.k8sAPIHandler.UpdateService)      // 更新 Service
			services.DELETE("/:name", s.k8sAPIHandler.DeleteService)   // 删除 Service
		}

		// ===========================
		// StatefulSet 管理 API
		// ===========================
		statefulsets := k8sAPI.Group("/clusters/:clusterId/namespaces/:namespace/statefulsets")
		{
			statefulsets.GET("", s.k8sAPIHandler.ListStatefulSets)                      // 获取 StatefulSet 列表
			statefulsets.GET("/:name", s.k8sAPIHandler.GetStatefulSet)                  // 获取 StatefulSet 详情
			statefulsets.PUT("/:name/scale", s.k8sAPIHandler.ScaleStatefulSet)          // 扩缩容
			statefulsets.POST("/:name/restart", s.k8sAPIHandler.RestartStatefulSet)     // 重启
			statefulsets.DELETE("/:name", s.k8sAPIHandler.DeleteStatefulSet)            // 删除
		}

		// ===========================
		// DaemonSet 管理 API
		// ===========================
		daemonsets := k8sAPI.Group("/clusters/:clusterId/namespaces/:namespace/daemonsets")
		{
			daemonsets.GET("", s.k8sAPIHandler.ListDaemonSets)                      // 获取 DaemonSet 列表
			daemonsets.GET("/:name", s.k8sAPIHandler.GetDaemonSet)                  // 获取 DaemonSet 详情
			daemonsets.POST("/:name/restart", s.k8sAPIHandler.RestartDaemonSet)     // 重启
			daemonsets.DELETE("/:name", s.k8sAPIHandler.DeleteDaemonSet)            // 删除
		}

		// ===========================
		// ConfigMap 管理 API
		// ===========================
		configmaps := k8sAPI.Group("/clusters/:clusterId/namespaces/:namespace/configmaps")
		{
			configmaps.GET("", s.k8sAPIHandler.ListConfigMaps)                     // 获取 ConfigMap 列表
			configmaps.GET("/:name", s.k8sAPIHandler.GetConfigMap)                 // 获取 ConfigMap 详情
			configmaps.POST("", s.k8sAPIHandler.CreateConfigMap)                   // 创建 ConfigMap
			configmaps.PUT("/:name", s.k8sAPIHandler.UpdateConfigMap)              // 更新 ConfigMap
			configmaps.DELETE("/:name", s.k8sAPIHandler.DeleteConfigMap)           // 删除 ConfigMap
		}

		// ===========================
		// Secret 管理 API
		// ===========================
		secrets := k8sAPI.Group("/clusters/:clusterId/namespaces/:namespace/secrets")
		{
			secrets.GET("", s.k8sAPIHandler.ListSecrets)                           // 获取 Secret 列表
			secrets.GET("/:name", s.k8sAPIHandler.GetSecret)                       // 获取 Secret 详情
			secrets.POST("", s.k8sAPIHandler.CreateSecret)                         // 创建 Secret
			secrets.PUT("/:name", s.k8sAPIHandler.UpdateSecret)                    // 更新 Secret
			secrets.DELETE("/:name", s.k8sAPIHandler.DeleteSecret)                 // 删除 Secret
		}

		// ===========================
		// Endpoints 管理 API
		// ===========================
		endpoints := k8sAPI.Group("/clusters/:clusterId/namespaces/:namespace/endpoints")
		{
			endpoints.GET("", s.k8sAPIHandler.ListEndpoints)                       // 获取 Endpoints 列表
			endpoints.GET("/:name", s.k8sAPIHandler.GetEndpoint)                   // 获取 Endpoint 详情
			endpoints.DELETE("/:name", s.k8sAPIHandler.DeleteEndpoint)             // 删除 Endpoint
		}

		// ===========================
		// PersistentVolumeClaim 管理 API
		// ===========================
		pvcs := k8sAPI.Group("/clusters/:clusterId/namespaces/:namespace/persistentvolumeclaims")
		{
			pvcs.GET("", s.k8sAPIHandler.ListPVCs)                                 // 获取 PVC 列表
			pvcs.GET("/:name", s.k8sAPIHandler.GetPVC)                             // 获取 PVC 详情
			pvcs.DELETE("/:name", s.k8sAPIHandler.DeletePVC)                       // 删除 PVC
		}

		// ===========================
		// PersistentVolume 管理 API (cluster-scoped)
		// ===========================
		pvs := k8sAPI.Group("/clusters/:clusterId/persistentvolumes")
		{
			pvs.GET("", s.k8sAPIHandler.ListPVs)                                   // 获取 PV 列表
			pvs.GET("/:name", s.k8sAPIHandler.GetPV)                               // 获取 PV 详情
			pvs.DELETE("/:name", s.k8sAPIHandler.DeletePV)                         // 删除 PV
		}

		// ===========================
		// EndpointSlice 管理 API
		// ===========================
		endpointslices := k8sAPI.Group("/clusters/:clusterId/namespaces/:namespace/endpointslices")
		{
			endpointslices.GET("", s.k8sAPIHandler.ListEndpointSlices)             // 获取 EndpointSlice 列表
			endpointslices.GET("/:name", s.k8sAPIHandler.GetEndpointSlice)         // 获取 EndpointSlice 详情
			endpointslices.DELETE("/:name", s.k8sAPIHandler.DeleteEndpointSlice)   // 删除 EndpointSlice
		}

		// ===========================
		// HorizontalPodAutoscaler 管理 API
		// ===========================
		hpas := k8sAPI.Group("/clusters/:clusterId/namespaces/:namespace/horizontalpodautoscalers")
		{
			hpas.GET("", s.k8sAPIHandler.ListHPAs)                                 // 获取 HPA 列表
			hpas.GET("/:name", s.k8sAPIHandler.GetHPA)                             // 获取 HPA 详情
			hpas.DELETE("/:name", s.k8sAPIHandler.DeleteHPA)                       // 删除 HPA
		}

		// ===========================
		// Event 管理 API
		// ===========================
		events := k8sAPI.Group("/clusters/:clusterId/namespaces/:namespace/events")
		{
			events.GET("", s.k8sAPIHandler.ListEvents)                             // 获取 Event 列表 (支持 ?type= 过滤)
			events.GET("/:name", s.k8sAPIHandler.GetEvent)                         // 获取 Event 详情
		}

		// ===========================
		// RoleBinding 管理 API
		// ===========================
		rolebindings := k8sAPI.Group("/clusters/:clusterId/namespaces/:namespace/rolebindings")
		{
			rolebindings.GET("", s.k8sAPIHandler.ListRoleBindings)                 // 获取 RoleBinding 列表
			rolebindings.GET("/:name", s.k8sAPIHandler.GetRoleBinding)             // 获取 RoleBinding 详情
			rolebindings.DELETE("/:name", s.k8sAPIHandler.DeleteRoleBinding)       // 删除 RoleBinding
		}

		// ===========================
		// ClusterRole 管理 API (cluster-scoped)
		// ===========================
		clusterroles := k8sAPI.Group("/clusters/:clusterId/clusterroles")
		{
			clusterroles.GET("", s.k8sAPIHandler.ListClusterRoles)                 // 获取 ClusterRole 列表
			clusterroles.GET("/:name", s.k8sAPIHandler.GetClusterRole)             // 获取 ClusterRole 详情
			clusterroles.DELETE("/:name", s.k8sAPIHandler.DeleteClusterRole)       // 删除 ClusterRole
		}

		// ===========================
		// PriorityClass 管理 API (cluster-scoped)
		// ===========================
		priorityclasses := k8sAPI.Group("/clusters/:clusterId/priorityclasses")
		{
			priorityclasses.GET("", s.k8sAPIHandler.ListPriorityClasses)           // 获取 PriorityClass 列表
			priorityclasses.GET("/:name", s.k8sAPIHandler.GetPriorityClass)        // 获取 PriorityClass 详情
			priorityclasses.DELETE("/:name", s.k8sAPIHandler.DeletePriorityClass)  // 删除 PriorityClass
		}

		// ===========================
		// Role 管理 API (namespace-scoped)
		// ===========================
		roles := k8sAPI.Group("/clusters/:clusterId/namespaces/:namespace/roles")
		{
			roles.GET("", s.k8sAPIHandler.ListRoles)              // 获取 Role 列表
			roles.GET("/:name", s.k8sAPIHandler.GetRole)          // 获取 Role 详情
			roles.DELETE("/:name", s.k8sAPIHandler.DeleteRole)    // 删除 Role
		}

		// ===========================
		// StorageClass 管理 API (cluster-scoped)
		// ===========================
		storageclasses := k8sAPI.Group("/clusters/:clusterId/storageclasses")
		{
			storageclasses.GET("", s.k8sAPIHandler.ListStorageClasses)           // 获取 StorageClass 列表
			storageclasses.GET("/:name", s.k8sAPIHandler.GetStorageClass)        // 获取 StorageClass 详情
			storageclasses.DELETE("/:name", s.k8sAPIHandler.DeleteStorageClass)  // 删除 StorageClass
		}
	}

	// 记录注册的路由
	logger.Infow("K8s API routes registered",
		"base_path", "/api/k8s",
		"cluster_endpoints", 6,
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
	)
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.engine,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	s.log.WithField("port", s.config.Port).Info("Starting cluster service")
	logger.Infow("Server starting",
		"port", s.config.Port,
		"mode", s.config.Mode,
		"read_timeout", s.config.ReadTimeout,
		"write_timeout", s.config.WriteTimeout,
	)

	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("Shutting down server...")
	logger.Info("Server shutdown initiated")

	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// ginLogger returns a gin middleware that logs requests (用于 logrus 兼容)
func ginLogger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		entry := logger.WithFields(logrus.Fields{
			"status":     status,
			"method":     c.Request.Method,
			"path":       path,
			"query":      query,
			"ip":         c.ClientIP(),
			"latency":    latency,
			"user_agent": c.Request.UserAgent(),
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.String())
		} else {
			entry.Info("Request completed")
		}
	}
}
