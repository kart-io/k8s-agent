package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/cluster-service/internal/service"
	"github.com/kart-io/k8s-agent/common/logger"
	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/k8s-agent/common/validator"
)

// K8sAPIHandler 处理所有 Kubernetes API 请求
// 基于 /api/k8s 路径的完整 K8s 管理接口
type K8sAPIHandler struct {
	clusterService     *service.K8sClusterService
	namespaceService   *service.K8sNamespaceService
	podService         *service.K8sPodService
	deploymentService  *service.K8sDeploymentService
	nodeService        *service.K8sNodeService
	serviceService     *service.K8sServiceService
	statefulsetService *service.K8sStatefulSetService
	daemonsetService   *service.K8sDaemonSetService
	configmapService   *service.K8sConfigMapService
	secretService      *service.K8sSecretService
}

// NewK8sAPIHandler 创建新的 K8s API 处理器
func NewK8sAPIHandler(
	clusterService *service.K8sClusterService,
	namespaceService *service.K8sNamespaceService,
	podService *service.K8sPodService,
	deploymentService *service.K8sDeploymentService,
	nodeService *service.K8sNodeService,
	serviceService *service.K8sServiceService,
	statefulsetService *service.K8sStatefulSetService,
	daemonsetService *service.K8sDaemonSetService,
	configmapService *service.K8sConfigMapService,
	secretService *service.K8sSecretService,
) *K8sAPIHandler {
	return &K8sAPIHandler{
		clusterService:     clusterService,
		namespaceService:   namespaceService,
		podService:         podService,
		deploymentService:  deploymentService,
		nodeService:        nodeService,
		serviceService:     serviceService,
		statefulsetService: statefulsetService,
		daemonsetService:   daemonsetService,
		configmapService:   configmapService,
		secretService:      secretService,
	}
}

// ===========================
// 集群管理接口
// ===========================

// ListClusters GET /api/k8s/clusters
// 获取集群列表（支持分页）
func (h *K8sAPIHandler) ListClusters(c *gin.Context) {
	params := pagination.Parse(c)

	logger.Infow("Listing clusters",
		"page", params.Page,
		"page_size", params.GetPageSize(),
	)

	clusters, total, err := h.clusterService.ListClusters(
		c.Request.Context(),
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list clusters", "error", err.Error())
		response.InternalError(c, "Failed to list clusters", err)
		return
	}

	resp := pagination.NewResponse(clusters, total, params)
	response.Success(c, resp)
}

// GetCluster GET /api/k8s/clusters/:id
// 获取集群详情
func (h *K8sAPIHandler) GetCluster(c *gin.Context) {
	clusterID := c.Param("id")

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Getting cluster details", "cluster_id", clusterID)

	cluster, err := h.clusterService.GetCluster(c.Request.Context(), clusterID)
	if err != nil {
		logger.Errorw("Failed to get cluster", "cluster_id", clusterID, "error", err.Error())
		response.NotFound(c, "Cluster not found", err)
		return
	}

	response.Success(c, cluster)
}

// CreateCluster POST /api/k8s/clusters
// 创建新集群
func (h *K8sAPIHandler) CreateCluster(c *gin.Context) {
	var req struct {
		Name        string            `json:"name" binding:"required"`
		Description string            `json:"description"`
		Endpoint    string            `json:"endpoint" binding:"required"`
		KubeConfig  string            `json:"kubeconfig" binding:"required"`
		Region      string            `json:"region"`
		Provider    string            `json:"provider"`
		Labels      map[string]string `json:"labels"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	// 验证集群名称
	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid cluster name", err)
		return
	}

	logger.Infow("Creating cluster",
		"name", req.Name,
		"endpoint", req.Endpoint,
		"region", req.Region,
	)

	cluster, err := h.clusterService.CreateCluster(
		c.Request.Context(),
		req.Name,
		req.Description,
		req.Endpoint,
		req.KubeConfig,
		req.Region,
		req.Provider,
		req.Labels,
	)
	if err != nil {
		logger.Errorw("Failed to create cluster", "name", req.Name, "error", err.Error())
		response.InternalError(c, "Failed to create cluster", err)
		return
	}

	response.SuccessWithMessage(c, "Cluster created successfully", cluster)
}

// UpdateCluster PUT /api/k8s/clusters/:id
// 更新集群信息
func (h *K8sAPIHandler) UpdateCluster(c *gin.Context) {
	clusterID := c.Param("id")

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	var req struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Labels      map[string]string `json:"labels"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	logger.Infow("Updating cluster", "cluster_id", clusterID)

	cluster, err := h.clusterService.UpdateCluster(
		c.Request.Context(),
		clusterID,
		req.Name,
		req.Description,
		req.Labels,
	)
	if err != nil {
		logger.Errorw("Failed to update cluster", "cluster_id", clusterID, "error", err.Error())
		response.InternalError(c, "Failed to update cluster", err)
		return
	}

	response.SuccessWithMessage(c, "Cluster updated successfully", cluster)
}

// DeleteCluster DELETE /api/k8s/clusters/:id
// 删除集群
func (h *K8sAPIHandler) DeleteCluster(c *gin.Context) {
	clusterID := c.Param("id")

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Deleting cluster", "cluster_id", clusterID)

	if err := h.clusterService.DeleteCluster(c.Request.Context(), clusterID); err != nil {
		logger.Errorw("Failed to delete cluster", "cluster_id", clusterID, "error", err.Error())
		response.InternalError(c, "Failed to delete cluster", err)
		return
	}

	response.SuccessWithMessage(c, "Cluster deleted successfully", gin.H{
		"cluster_id": clusterID,
	})
}

// GetClusterHealth GET /api/k8s/clusters/:id/health
// 获取集群健康状态
func (h *K8sAPIHandler) GetClusterHealthStatus(c *gin.Context) {
	clusterID := c.Param("id")

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Getting cluster health", "cluster_id", clusterID)

	health, err := h.clusterService.GetClusterHealth(c.Request.Context(), clusterID)
	if err != nil {
		logger.Errorw("Failed to get cluster health", "cluster_id", clusterID, "error", err.Error())
		response.InternalError(c, "Failed to get cluster health", err)
		return
	}

	response.Success(c, health)
}

// ===========================
// 命名空间管理接口
// ===========================

// ListNamespaces GET /api/k8s/clusters/:clusterId/namespaces
// 获取命名空间列表
func (h *K8sAPIHandler) ListNamespaces(c *gin.Context) {
	clusterID := c.Param("clusterId")
	params := pagination.Parse(c)

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Listing namespaces",
		"cluster_id", clusterID,
		"page", params.Page,
	)

	namespaces, total, err := h.namespaceService.ListNamespaces(
		c.Request.Context(),
		clusterID,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list namespaces", "cluster_id", clusterID, "error", err.Error())
		response.InternalError(c, "Failed to list namespaces", err)
		return
	}

	resp := pagination.NewResponse(namespaces, total, params)
	response.Success(c, resp)
}

// GetNamespace GET /api/k8s/clusters/:clusterId/namespaces/:name
// 获取命名空间详情
func (h *K8sAPIHandler) GetNamespace(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespaceName := c.Param("name")

	if err := validator.ValidateK8sName(namespaceName); err != nil {
		response.BadRequest(c, "Invalid namespace name", err)
		return
	}

	logger.Infow("Getting namespace details",
		"cluster_id", clusterID,
		"namespace", namespaceName,
	)

	namespace, err := h.namespaceService.GetNamespace(c.Request.Context(), clusterID, namespaceName)
	if err != nil {
		logger.Errorw("Failed to get namespace",
			"cluster_id", clusterID,
			"namespace", namespaceName,
			"error", err.Error(),
		)
		response.NotFound(c, "Namespace not found", err)
		return
	}

	response.Success(c, namespace)
}

// CreateNamespace POST /api/k8s/clusters/:clusterId/namespaces
// 创建命名空间
func (h *K8sAPIHandler) CreateNamespace(c *gin.Context) {
	clusterID := c.Param("clusterId")

	var req struct {
		Name   string            `json:"name" binding:"required"`
		Labels map[string]string `json:"labels"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid namespace name", err)
		return
	}

	logger.Infow("Creating namespace",
		"cluster_id", clusterID,
		"namespace", req.Name,
	)

	namespace, err := h.namespaceService.CreateNamespace(
		c.Request.Context(),
		clusterID,
		req.Name,
		req.Labels,
	)
	if err != nil {
		logger.Errorw("Failed to create namespace",
			"cluster_id", clusterID,
			"namespace", req.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to create namespace", err)
		return
	}

	response.SuccessWithMessage(c, "Namespace created successfully", namespace)
}

// DeleteNamespace DELETE /api/k8s/clusters/:clusterId/namespaces/:name
// 删除命名空间
func (h *K8sAPIHandler) DeleteNamespace(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespaceName := c.Param("name")

	if err := validator.ValidateK8sName(namespaceName); err != nil {
		response.BadRequest(c, "Invalid namespace name", err)
		return
	}

	logger.Infow("Deleting namespace",
		"cluster_id", clusterID,
		"namespace", namespaceName,
	)

	if err := h.namespaceService.DeleteNamespace(c.Request.Context(), clusterID, namespaceName); err != nil {
		logger.Errorw("Failed to delete namespace",
			"cluster_id", clusterID,
			"namespace", namespaceName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete namespace", err)
		return
	}

	response.SuccessWithMessage(c, "Namespace deleted successfully", gin.H{
		"namespace": namespaceName,
	})
}

// ===========================
// Pod 管理接口
// ===========================

// ListPods GET /api/k8s/clusters/:clusterId/namespaces/:namespace/pods
// 获取 Pod 列表
func (h *K8sAPIHandler) ListPods(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing pods",
		"cluster_id", clusterID,
		"namespace", namespace,
		"page", params.Page,
	)

	pods, total, err := h.podService.ListPods(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list pods",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list pods", err)
		return
	}

	resp := pagination.NewResponse(pods, total, params)
	response.Success(c, resp)
}

// GetPod GET /api/k8s/clusters/:clusterId/namespaces/:namespace/pods/:name
// 获取 Pod 详情
func (h *K8sAPIHandler) GetPod(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	podName := c.Param("name")

	logger.Infow("Getting pod details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"pod", podName,
	)

	pod, err := h.podService.GetPod(c.Request.Context(), clusterID, namespace, podName)
	if err != nil {
		logger.Errorw("Failed to get pod",
			"cluster_id", clusterID,
			"namespace", namespace,
			"pod", podName,
			"error", err.Error(),
		)
		response.NotFound(c, "Pod not found", err)
		return
	}

	response.Success(c, pod)
}

// DeletePod DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/pods/:name
// 删除 Pod
func (h *K8sAPIHandler) DeletePod(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	podName := c.Param("name")

	logger.Infow("Deleting pod",
		"cluster_id", clusterID,
		"namespace", namespace,
		"pod", podName,
	)

	if err := h.podService.DeletePod(c.Request.Context(), clusterID, namespace, podName); err != nil {
		logger.Errorw("Failed to delete pod",
			"cluster_id", clusterID,
			"namespace", namespace,
			"pod", podName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete pod", err)
		return
	}

	response.SuccessWithMessage(c, "Pod deleted successfully", gin.H{
		"pod": podName,
	})
}

// GetPodLogs GET /api/k8s/clusters/:clusterId/namespaces/:namespace/pods/:name/logs
// 获取 Pod 日志
func (h *K8sAPIHandler) GetPodLogs(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	podName := c.Param("name")

	// 查询参数
	container := c.Query("container")
	tailLines := c.DefaultQuery("tailLines", "100")
	follow := c.Query("follow") == "true"

	logger.Infow("Getting pod logs",
		"cluster_id", clusterID,
		"namespace", namespace,
		"pod", podName,
		"container", container,
	)

	logs, err := h.podService.GetPodLogs(
		c.Request.Context(),
		clusterID,
		namespace,
		podName,
		container,
		tailLines,
		follow,
	)
	if err != nil {
		logger.Errorw("Failed to get pod logs",
			"cluster_id", clusterID,
			"namespace", namespace,
			"pod", podName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to get pod logs", err)
		return
	}

	response.Success(c, gin.H{
		"logs": logs,
	})
}

// ===========================
// Deployment 管理接口
// ===========================

// ListDeployments GET /api/k8s/clusters/:clusterId/namespaces/:namespace/deployments
// 获取 Deployment 列表
func (h *K8sAPIHandler) ListDeployments(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing deployments",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	deployments, total, err := h.deploymentService.ListDeployments(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list deployments",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list deployments", err)
		return
	}

	resp := pagination.NewResponse(deployments, total, params)
	response.Success(c, resp)
}

// GetDeployment GET /api/k8s/clusters/:clusterId/namespaces/:namespace/deployments/:name
// 获取 Deployment 详情
func (h *K8sAPIHandler) GetDeployment(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	deploymentName := c.Param("name")

	logger.Infow("Getting deployment details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"deployment", deploymentName,
	)

	deployment, err := h.deploymentService.GetDeployment(
		c.Request.Context(),
		clusterID,
		namespace,
		deploymentName,
	)
	if err != nil {
		logger.Errorw("Failed to get deployment",
			"cluster_id", clusterID,
			"namespace", namespace,
			"deployment", deploymentName,
			"error", err.Error(),
		)
		response.NotFound(c, "Deployment not found", err)
		return
	}

	response.Success(c, deployment)
}

// ScaleDeployment PUT /api/k8s/clusters/:clusterId/namespaces/:namespace/deployments/:name/scale
// 扩缩容 Deployment
func (h *K8sAPIHandler) ScaleDeployment(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	deploymentName := c.Param("name")

	var req struct {
		Replicas int32 `json:"replicas" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	if err := validator.ValidateReplicas(req.Replicas); err != nil {
		response.BadRequest(c, "Invalid replicas count", err)
		return
	}

	logger.Infow("Scaling deployment",
		"cluster_id", clusterID,
		"namespace", namespace,
		"deployment", deploymentName,
		"replicas", req.Replicas,
	)

	deployment, err := h.deploymentService.ScaleDeployment(
		c.Request.Context(),
		clusterID,
		namespace,
		deploymentName,
		req.Replicas,
	)
	if err != nil {
		logger.Errorw("Failed to scale deployment",
			"cluster_id", clusterID,
			"namespace", namespace,
			"deployment", deploymentName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to scale deployment", err)
		return
	}

	response.SuccessWithMessage(c, "Deployment scaled successfully", deployment)
}

// RestartDeployment POST /api/k8s/clusters/:clusterId/namespaces/:namespace/deployments/:name/restart
// 重启 Deployment
func (h *K8sAPIHandler) RestartDeployment(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	deploymentName := c.Param("name")

	logger.Infow("Restarting deployment",
		"cluster_id", clusterID,
		"namespace", namespace,
		"deployment", deploymentName,
	)

	if err := h.deploymentService.RestartDeployment(
		c.Request.Context(),
		clusterID,
		namespace,
		deploymentName,
	); err != nil {
		logger.Errorw("Failed to restart deployment",
			"cluster_id", clusterID,
			"namespace", namespace,
			"deployment", deploymentName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to restart deployment", err)
		return
	}

	response.SuccessWithMessage(c, "Deployment restarted successfully", gin.H{
		"deployment": deploymentName,
	})
}

// ===========================
// Node 管理接口
// ===========================

// ListNodes GET /api/k8s/clusters/:clusterId/nodes
// 获取 Node 列表
func (h *K8sAPIHandler) ListNodes(c *gin.Context) {
	clusterID := c.Param("clusterId")
	params := pagination.Parse(c)

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Listing nodes",
		"cluster_id", clusterID,
		"page", params.Page,
	)

	nodes, total, err := h.nodeService.ListNodes(
		c.Request.Context(),
		clusterID,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list nodes", "cluster_id", clusterID, "error", err.Error())
		response.InternalError(c, "Failed to list nodes", err)
		return
	}

	resp := pagination.NewResponse(nodes, total, params)
	response.Success(c, resp)
}

// GetNode GET /api/k8s/clusters/:clusterId/nodes/:name
// 获取 Node 详情
func (h *K8sAPIHandler) GetNode(c *gin.Context) {
	clusterID := c.Param("clusterId")
	nodeName := c.Param("name")

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Getting node details",
		"cluster_id", clusterID,
		"node", nodeName,
	)

	node, err := h.nodeService.GetNode(c.Request.Context(), clusterID, nodeName)
	if err != nil {
		logger.Errorw("Failed to get node",
			"cluster_id", clusterID,
			"node", nodeName,
			"error", err.Error(),
		)
		response.NotFound(c, "Node not found", err)
		return
	}

	response.Success(c, node)
}

// CordonNode POST /api/k8s/clusters/:clusterId/nodes/:name/cordon
// 标记 Node 为不可调度
func (h *K8sAPIHandler) CordonNode(c *gin.Context) {
	clusterID := c.Param("clusterId")
	nodeName := c.Param("name")

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Cordoning node",
		"cluster_id", clusterID,
		"node", nodeName,
	)

	if err := h.nodeService.CordonNode(c.Request.Context(), clusterID, nodeName); err != nil {
		logger.Errorw("Failed to cordon node",
			"cluster_id", clusterID,
			"node", nodeName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to cordon node", err)
		return
	}

	response.SuccessWithMessage(c, "Node cordoned successfully", gin.H{
		"node": nodeName,
	})
}

// UncordonNode POST /api/k8s/clusters/:clusterId/nodes/:name/uncordon
// 标记 Node 为可调度
func (h *K8sAPIHandler) UncordonNode(c *gin.Context) {
	clusterID := c.Param("clusterId")
	nodeName := c.Param("name")

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Uncordoning node",
		"cluster_id", clusterID,
		"node", nodeName,
	)

	if err := h.nodeService.UncordonNode(c.Request.Context(), clusterID, nodeName); err != nil {
		logger.Errorw("Failed to uncordon node",
			"cluster_id", clusterID,
			"node", nodeName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to uncordon node", err)
		return
	}

	response.SuccessWithMessage(c, "Node uncordoned successfully", gin.H{
		"node": nodeName,
	})
}

// DrainNode POST /api/k8s/clusters/:clusterId/nodes/:name/drain
// 驱逐 Node 上的 Pod
func (h *K8sAPIHandler) DrainNode(c *gin.Context) {
	clusterID := c.Param("clusterId")
	nodeName := c.Param("name")

	var req struct {
		Force bool `json:"force"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有请求体，使用默认值
		req.Force = false
	}

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Draining node",
		"cluster_id", clusterID,
		"node", nodeName,
		"force", req.Force,
	)

	if err := h.nodeService.DrainNode(c.Request.Context(), clusterID, nodeName, req.Force); err != nil {
		logger.Errorw("Failed to drain node",
			"cluster_id", clusterID,
			"node", nodeName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to drain node", err)
		return
	}

	response.SuccessWithMessage(c, "Node drained successfully", gin.H{
		"node": nodeName,
	})
}

// ===========================
// Service 管理接口
// ===========================

// ListServices GET /api/k8s/clusters/:clusterId/namespaces/:namespace/services
// 获取 Service 列表
func (h *K8sAPIHandler) ListServices(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing services",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	services, total, err := h.serviceService.ListServices(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list services",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list services", err)
		return
	}

	resp := pagination.NewResponse(services, total, params)
	response.Success(c, resp)
}

// GetService GET /api/k8s/clusters/:clusterId/namespaces/:namespace/services/:name
// 获取 Service 详情
func (h *K8sAPIHandler) GetService(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	serviceName := c.Param("name")

	logger.Infow("Getting service details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"service", serviceName,
	)

	service, err := h.serviceService.GetService(c.Request.Context(), clusterID, namespace, serviceName)
	if err != nil {
		logger.Errorw("Failed to get service",
			"cluster_id", clusterID,
			"namespace", namespace,
			"service", serviceName,
			"error", err.Error(),
		)
		response.NotFound(c, "Service not found", err)
		return
	}

	response.Success(c, service)
}

// CreateService POST /api/k8s/clusters/:clusterId/namespaces/:namespace/services
// 创建 Service
func (h *K8sAPIHandler) CreateService(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")

	var req service.CreateServiceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	// 确保 namespace 一致
	req.Namespace = namespace

	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid service name", err)
		return
	}

	logger.Infow("Creating service",
		"cluster_id", clusterID,
		"namespace", namespace,
		"service", req.Name,
	)

	svc, err := h.serviceService.CreateService(c.Request.Context(), clusterID, &req)
	if err != nil {
		logger.Errorw("Failed to create service",
			"cluster_id", clusterID,
			"namespace", namespace,
			"service", req.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to create service", err)
		return
	}

	response.SuccessWithMessage(c, "Service created successfully", svc)
}

// UpdateService PUT /api/k8s/clusters/:clusterId/namespaces/:namespace/services/:name
// 更新 Service
func (h *K8sAPIHandler) UpdateService(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	serviceName := c.Param("name")

	var req service.CreateServiceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	logger.Infow("Updating service",
		"cluster_id", clusterID,
		"namespace", namespace,
		"service", serviceName,
	)

	svc, err := h.serviceService.UpdateService(c.Request.Context(), clusterID, namespace, serviceName, &req)
	if err != nil {
		logger.Errorw("Failed to update service",
			"cluster_id", clusterID,
			"namespace", namespace,
			"service", serviceName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to update service", err)
		return
	}

	response.SuccessWithMessage(c, "Service updated successfully", svc)
}

// DeleteService DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/services/:name
// 删除 Service
func (h *K8sAPIHandler) DeleteService(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	serviceName := c.Param("name")

	logger.Infow("Deleting service",
		"cluster_id", clusterID,
		"namespace", namespace,
		"service", serviceName,
	)

	if err := h.serviceService.DeleteService(c.Request.Context(), clusterID, namespace, serviceName); err != nil {
		logger.Errorw("Failed to delete service",
			"cluster_id", clusterID,
			"namespace", namespace,
			"service", serviceName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete service", err)
		return
	}

	response.SuccessWithMessage(c, "Service deleted successfully", gin.H{
		"service": serviceName,
	})
}

// ===========================
// StatefulSet 管理接口
// ===========================

// ListStatefulSets GET /api/k8s/clusters/:clusterId/namespaces/:namespace/statefulsets
// 获取 StatefulSet 列表
func (h *K8sAPIHandler) ListStatefulSets(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing statefulsets",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	statefulsets, total, err := h.statefulsetService.ListStatefulSets(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list statefulsets",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list statefulsets", err)
		return
	}

	resp := pagination.NewResponse(statefulsets, total, params)
	response.Success(c, resp)
}

// GetStatefulSet GET /api/k8s/clusters/:clusterId/namespaces/:namespace/statefulsets/:name
// 获取 StatefulSet 详情
func (h *K8sAPIHandler) GetStatefulSet(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	statefulsetName := c.Param("name")

	logger.Infow("Getting statefulset details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"statefulset", statefulsetName,
	)

	statefulset, err := h.statefulsetService.GetStatefulSet(
		c.Request.Context(),
		clusterID,
		namespace,
		statefulsetName,
	)
	if err != nil {
		logger.Errorw("Failed to get statefulset",
			"cluster_id", clusterID,
			"namespace", namespace,
			"statefulset", statefulsetName,
			"error", err.Error(),
		)
		response.NotFound(c, "StatefulSet not found", err)
		return
	}

	response.Success(c, statefulset)
}

// ScaleStatefulSet PUT /api/k8s/clusters/:clusterId/namespaces/:namespace/statefulsets/:name/scale
// 扩缩容 StatefulSet
func (h *K8sAPIHandler) ScaleStatefulSet(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	statefulsetName := c.Param("name")

	var req struct {
		Replicas int32 `json:"replicas" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	if err := validator.ValidateReplicas(req.Replicas); err != nil {
		response.BadRequest(c, "Invalid replicas count", err)
		return
	}

	logger.Infow("Scaling statefulset",
		"cluster_id", clusterID,
		"namespace", namespace,
		"statefulset", statefulsetName,
		"replicas", req.Replicas,
	)

	statefulset, err := h.statefulsetService.ScaleStatefulSet(
		c.Request.Context(),
		clusterID,
		namespace,
		statefulsetName,
		req.Replicas,
	)
	if err != nil {
		logger.Errorw("Failed to scale statefulset",
			"cluster_id", clusterID,
			"namespace", namespace,
			"statefulset", statefulsetName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to scale statefulset", err)
		return
	}

	response.SuccessWithMessage(c, "StatefulSet scaled successfully", statefulset)
}

// RestartStatefulSet POST /api/k8s/clusters/:clusterId/namespaces/:namespace/statefulsets/:name/restart
// 重启 StatefulSet
func (h *K8sAPIHandler) RestartStatefulSet(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	statefulsetName := c.Param("name")

	logger.Infow("Restarting statefulset",
		"cluster_id", clusterID,
		"namespace", namespace,
		"statefulset", statefulsetName,
	)

	if err := h.statefulsetService.RestartStatefulSet(
		c.Request.Context(),
		clusterID,
		namespace,
		statefulsetName,
	); err != nil {
		logger.Errorw("Failed to restart statefulset",
			"cluster_id", clusterID,
			"namespace", namespace,
			"statefulset", statefulsetName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to restart statefulset", err)
		return
	}

	response.SuccessWithMessage(c, "StatefulSet restarted successfully", gin.H{
		"statefulset": statefulsetName,
	})
}

// DeleteStatefulSet DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/statefulsets/:name
// 删除 StatefulSet
func (h *K8sAPIHandler) DeleteStatefulSet(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	statefulsetName := c.Param("name")

	logger.Infow("Deleting statefulset",
		"cluster_id", clusterID,
		"namespace", namespace,
		"statefulset", statefulsetName,
	)

	if err := h.statefulsetService.DeleteStatefulSet(c.Request.Context(), clusterID, namespace, statefulsetName); err != nil {
		logger.Errorw("Failed to delete statefulset",
			"cluster_id", clusterID,
			"namespace", namespace,
			"statefulset", statefulsetName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete statefulset", err)
		return
	}

	response.SuccessWithMessage(c, "StatefulSet deleted successfully", gin.H{
		"statefulset": statefulsetName,
	})
}

// ===========================
// DaemonSet 管理接口
// ===========================

// ListDaemonSets GET /api/k8s/clusters/:clusterId/namespaces/:namespace/daemonsets
// 获取 DaemonSet 列表
func (h *K8sAPIHandler) ListDaemonSets(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing daemonsets",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	daemonsets, total, err := h.daemonsetService.ListDaemonSets(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list daemonsets",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list daemonsets", err)
		return
	}

	resp := pagination.NewResponse(daemonsets, total, params)
	response.Success(c, resp)
}

// GetDaemonSet GET /api/k8s/clusters/:clusterId/namespaces/:namespace/daemonsets/:name
// 获取 DaemonSet 详情
func (h *K8sAPIHandler) GetDaemonSet(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	daemonsetName := c.Param("name")

	logger.Infow("Getting daemonset details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"daemonset", daemonsetName,
	)

	daemonset, err := h.daemonsetService.GetDaemonSet(
		c.Request.Context(),
		clusterID,
		namespace,
		daemonsetName,
	)
	if err != nil {
		logger.Errorw("Failed to get daemonset",
			"cluster_id", clusterID,
			"namespace", namespace,
			"daemonset", daemonsetName,
			"error", err.Error(),
		)
		response.NotFound(c, "DaemonSet not found", err)
		return
	}

	response.Success(c, daemonset)
}

// RestartDaemonSet POST /api/k8s/clusters/:clusterId/namespaces/:namespace/daemonsets/:name/restart
// 重启 DaemonSet
func (h *K8sAPIHandler) RestartDaemonSet(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	daemonsetName := c.Param("name")

	logger.Infow("Restarting daemonset",
		"cluster_id", clusterID,
		"namespace", namespace,
		"daemonset", daemonsetName,
	)

	if err := h.daemonsetService.RestartDaemonSet(
		c.Request.Context(),
		clusterID,
		namespace,
		daemonsetName,
	); err != nil {
		logger.Errorw("Failed to restart daemonset",
			"cluster_id", clusterID,
			"namespace", namespace,
			"daemonset", daemonsetName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to restart daemonset", err)
		return
	}

	response.SuccessWithMessage(c, "DaemonSet restarted successfully", gin.H{
		"daemonset": daemonsetName,
	})
}

// DeleteDaemonSet DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/daemonsets/:name
// 删除 DaemonSet
func (h *K8sAPIHandler) DeleteDaemonSet(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	daemonsetName := c.Param("name")

	logger.Infow("Deleting daemonset",
		"cluster_id", clusterID,
		"namespace", namespace,
		"daemonset", daemonsetName,
	)

	if err := h.daemonsetService.DeleteDaemonSet(c.Request.Context(), clusterID, namespace, daemonsetName); err != nil {
		logger.Errorw("Failed to delete daemonset",
			"cluster_id", clusterID,
			"namespace", namespace,
			"daemonset", daemonsetName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete daemonset", err)
		return
	}

	response.SuccessWithMessage(c, "DaemonSet deleted successfully", gin.H{
		"daemonset": daemonsetName,
	})
}

// ===========================
// ConfigMap 管理接口
// ===========================

// ListConfigMaps GET /api/k8s/clusters/:clusterId/namespaces/:namespace/configmaps
// 获取 ConfigMap 列表
func (h *K8sAPIHandler) ListConfigMaps(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing configmaps",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	configmaps, total, err := h.configmapService.ListConfigMaps(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list configmaps",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list configmaps", err)
		return
	}

	resp := pagination.NewResponse(configmaps, total, params)
	response.Success(c, resp)
}

// GetConfigMap GET /api/k8s/clusters/:clusterId/namespaces/:namespace/configmaps/:name
// 获取 ConfigMap 详情
func (h *K8sAPIHandler) GetConfigMap(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	configmapName := c.Param("name")

	logger.Infow("Getting configmap details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"configmap", configmapName,
	)

	configmap, err := h.configmapService.GetConfigMap(c.Request.Context(), clusterID, namespace, configmapName)
	if err != nil {
		logger.Errorw("Failed to get configmap",
			"cluster_id", clusterID,
			"namespace", namespace,
			"configmap", configmapName,
			"error", err.Error(),
		)
		response.NotFound(c, "ConfigMap not found", err)
		return
	}

	response.Success(c, configmap)
}

// CreateConfigMap POST /api/k8s/clusters/:clusterId/namespaces/:namespace/configmaps
// 创建 ConfigMap
func (h *K8sAPIHandler) CreateConfigMap(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")

	var req service.CreateConfigMapRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	// 确保 namespace 一致
	req.Namespace = namespace

	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid configmap name", err)
		return
	}

	logger.Infow("Creating configmap",
		"cluster_id", clusterID,
		"namespace", namespace,
		"configmap", req.Name,
	)

	configmap, err := h.configmapService.CreateConfigMap(c.Request.Context(), clusterID, &req)
	if err != nil {
		logger.Errorw("Failed to create configmap",
			"cluster_id", clusterID,
			"namespace", namespace,
			"configmap", req.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to create configmap", err)
		return
	}

	response.SuccessWithMessage(c, "ConfigMap created successfully", configmap)
}

// UpdateConfigMap PUT /api/k8s/clusters/:clusterId/namespaces/:namespace/configmaps/:name
// 更新 ConfigMap
func (h *K8sAPIHandler) UpdateConfigMap(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	configmapName := c.Param("name")

	var req service.CreateConfigMapRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	logger.Infow("Updating configmap",
		"cluster_id", clusterID,
		"namespace", namespace,
		"configmap", configmapName,
	)

	configmap, err := h.configmapService.UpdateConfigMap(c.Request.Context(), clusterID, namespace, configmapName, &req)
	if err != nil {
		logger.Errorw("Failed to update configmap",
			"cluster_id", clusterID,
			"namespace", namespace,
			"configmap", configmapName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to update configmap", err)
		return
	}

	response.SuccessWithMessage(c, "ConfigMap updated successfully", configmap)
}

// DeleteConfigMap DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/configmaps/:name
// 删除 ConfigMap
func (h *K8sAPIHandler) DeleteConfigMap(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	configmapName := c.Param("name")

	logger.Infow("Deleting configmap",
		"cluster_id", clusterID,
		"namespace", namespace,
		"configmap", configmapName,
	)

	if err := h.configmapService.DeleteConfigMap(c.Request.Context(), clusterID, namespace, configmapName); err != nil {
		logger.Errorw("Failed to delete configmap",
			"cluster_id", clusterID,
			"namespace", namespace,
			"configmap", configmapName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete configmap", err)
		return
	}

	response.SuccessWithMessage(c, "ConfigMap deleted successfully", gin.H{
		"configmap": configmapName,
	})
}

// ===========================
// Secret 管理接口
// ===========================

// ListSecrets GET /api/k8s/clusters/:clusterId/namespaces/:namespace/secrets
// 获取 Secret 列表
func (h *K8sAPIHandler) ListSecrets(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing secrets",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	secrets, total, err := h.secretService.ListSecrets(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list secrets",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list secrets", err)
		return
	}

	resp := pagination.NewResponse(secrets, total, params)
	response.Success(c, resp)
}

// GetSecret GET /api/k8s/clusters/:clusterId/namespaces/:namespace/secrets/:name
// 获取 Secret 详情
func (h *K8sAPIHandler) GetSecret(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	secretName := c.Param("name")

	// 查询参数：是否包含敏感数据
	includeData := c.Query("includeData") == "true"

	logger.Infow("Getting secret details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"secret", secretName,
		"include_data", includeData,
	)

	secret, err := h.secretService.GetSecret(c.Request.Context(), clusterID, namespace, secretName, includeData)
	if err != nil {
		logger.Errorw("Failed to get secret",
			"cluster_id", clusterID,
			"namespace", namespace,
			"secret", secretName,
			"error", err.Error(),
		)
		response.NotFound(c, "Secret not found", err)
		return
	}

	response.Success(c, secret)
}

// CreateSecret POST /api/k8s/clusters/:clusterId/namespaces/:namespace/secrets
// 创建 Secret
func (h *K8sAPIHandler) CreateSecret(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")

	var req service.CreateSecretRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	// 确保 namespace 一致
	req.Namespace = namespace

	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid secret name", err)
		return
	}

	logger.Infow("Creating secret",
		"cluster_id", clusterID,
		"namespace", namespace,
		"secret", req.Name,
	)

	secret, err := h.secretService.CreateSecret(c.Request.Context(), clusterID, &req)
	if err != nil {
		logger.Errorw("Failed to create secret",
			"cluster_id", clusterID,
			"namespace", namespace,
			"secret", req.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to create secret", err)
		return
	}

	response.SuccessWithMessage(c, "Secret created successfully", secret)
}

// UpdateSecret PUT /api/k8s/clusters/:clusterId/namespaces/:namespace/secrets/:name
// 更新 Secret
func (h *K8sAPIHandler) UpdateSecret(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	secretName := c.Param("name")

	var req service.CreateSecretRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	logger.Infow("Updating secret",
		"cluster_id", clusterID,
		"namespace", namespace,
		"secret", secretName,
	)

	secret, err := h.secretService.UpdateSecret(c.Request.Context(), clusterID, namespace, secretName, &req)
	if err != nil {
		logger.Errorw("Failed to update secret",
			"cluster_id", clusterID,
			"namespace", namespace,
			"secret", secretName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to update secret", err)
		return
	}

	response.SuccessWithMessage(c, "Secret updated successfully", secret)
}

// DeleteSecret DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/secrets/:name
// 删除 Secret
func (h *K8sAPIHandler) DeleteSecret(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	secretName := c.Param("name")

	logger.Infow("Deleting secret",
		"cluster_id", clusterID,
		"namespace", namespace,
		"secret", secretName,
	)

	if err := h.secretService.DeleteSecret(c.Request.Context(), clusterID, namespace, secretName); err != nil {
		logger.Errorw("Failed to delete secret",
			"cluster_id", clusterID,
			"namespace", namespace,
			"secret", secretName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete secret", err)
		return
	}

	response.SuccessWithMessage(c, "Secret deleted successfully", gin.H{
		"secret": secretName,
	})
}
