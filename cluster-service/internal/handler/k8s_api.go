package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/cluster-service/internal/service"
	"github.com/kart-io/k8s-agent/cluster-service/pkg/types"
	"github.com/kart-io/k8s-agent/common/logger"
	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/k8s-agent/common/validator"
	batchv1 "k8s.io/api/batch/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

// K8sAPIHandler 处理所有 Kubernetes API 请求
// 基于 /api/k8s 路径的完整 K8s 管理接口
type K8sAPIHandler struct {
	clusterService         *service.K8sClusterService
	namespaceService       *service.K8sNamespaceService
	podService             *service.K8sPodService
	deploymentService      *service.K8sDeploymentService
	nodeService            *service.K8sNodeService
	serviceService         *service.K8sServiceService
	statefulsetService     *service.K8sStatefulSetService
	daemonsetService       *service.K8sDaemonSetService
	configmapService       *service.K8sConfigMapService
	secretService          *service.K8sSecretService
	endpointService        *service.K8sEndpointService
	pvcService             *service.K8sPVCService
	pvService              *service.K8sPVService
	endpointsliceService   *service.K8sEndpointSliceService
	hpaService             *service.K8sHPAService
	eventService           *service.K8sEventService
	rolebindingService     *service.K8sRoleBindingService
	clusterroleService     *service.K8sClusterRoleService
	priorityclassService   *service.K8sPriorityClassService
	roleService            *service.K8sRoleService
	storageclassService    *service.K8sStorageClassService
	jobService             *service.K8sJobService
	cronjobService         *service.K8sCronJobService
	ingressService         *service.K8sIngressService
	networkpolicyService   *service.K8sNetworkPolicyService
	replicasetService         *service.K8sReplicaSetService
	limitrangeService         *service.K8sLimitRangeService
	serviceaccountService     *service.K8sServiceAccountService
	clusterrolebindingService *service.K8sClusterRoleBindingService
	resourcequotaService      *service.K8sResourceQuotaService
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
	endpointService *service.K8sEndpointService,
	pvcService *service.K8sPVCService,
	pvService *service.K8sPVService,
	endpointsliceService *service.K8sEndpointSliceService,
	hpaService *service.K8sHPAService,
	eventService *service.K8sEventService,
	rolebindingService *service.K8sRoleBindingService,
	clusterroleService *service.K8sClusterRoleService,
	priorityclassService *service.K8sPriorityClassService,
	roleService *service.K8sRoleService,
	storageclassService *service.K8sStorageClassService,
	jobService *service.K8sJobService,
	cronjobService *service.K8sCronJobService,
	ingressService *service.K8sIngressService,
	networkpolicyService *service.K8sNetworkPolicyService,
	replicasetService *service.K8sReplicaSetService,
	limitrangeService *service.K8sLimitRangeService,
	serviceaccountService *service.K8sServiceAccountService,
	clusterrolebindingService *service.K8sClusterRoleBindingService,
	resourcequotaService *service.K8sResourceQuotaService,
) *K8sAPIHandler {
	return &K8sAPIHandler{
		clusterService:            clusterService,
		namespaceService:          namespaceService,
		podService:                podService,
		deploymentService:         deploymentService,
		nodeService:               nodeService,
		serviceService:            serviceService,
		statefulsetService:        statefulsetService,
		daemonsetService:          daemonsetService,
		configmapService:          configmapService,
		secretService:             secretService,
		endpointService:           endpointService,
		pvcService:                pvcService,
		pvService:                 pvService,
		endpointsliceService:      endpointsliceService,
		hpaService:                hpaService,
		eventService:              eventService,
		rolebindingService:        rolebindingService,
		clusterroleService:        clusterroleService,
		priorityclassService:      priorityclassService,
		roleService:               roleService,
		storageclassService:       storageclassService,
		jobService:                jobService,
		cronjobService:            cronjobService,
		ingressService:            ingressService,
		networkpolicyService:      networkpolicyService,
		replicasetService:         replicasetService,
		limitrangeService:         limitrangeService,
		serviceaccountService:     serviceaccountService,
		clusterrolebindingService: clusterrolebindingService,
		resourcequotaService:      resourcequotaService,
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
	var req types.GetClusterRequest

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err)
		return
	}

	if err := validator.ValidateClusterID(req.ClusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Getting cluster details", "cluster_id", req.ClusterID)

	cluster, err := h.clusterService.GetCluster(c.Request.Context(), req.ClusterID)
	if err != nil {
		logger.Errorw("Failed to get cluster", "cluster_id", req.ClusterID, "error", err.Error())
		response.NotFound(c, "Cluster not found", err)
		return
	}

	response.Success(c, cluster)
}

// GetClusterOptions GET /api/k8s/clusters/options
// 获取集群选择器列表（用于下拉选择）
func (h *K8sAPIHandler) GetClusterOptions(c *gin.Context) {
	logger.Infow("Getting cluster options")

	options, err := h.clusterService.GetClusterOptions(c.Request.Context())
	if err != nil {
		logger.Errorw("Failed to get cluster options", "error", err.Error())
		response.InternalError(c, "Failed to get cluster options", err)
		return
	}

	response.Success(c, options)
}

// CreateCluster POST /api/k8s/clusters
// 创建新集群
func (h *K8sAPIHandler) CreateCluster(c *gin.Context) {
	var req types.CreateClusterRequest

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
	var req types.UpdateClusterRequest

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid path parameters", err)
		return
	}

	if err := validator.ValidateClusterID(req.ClusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	// 绑定请求体参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	logger.Infow("Updating cluster", "cluster_id", req.ClusterID)

	cluster, err := h.clusterService.UpdateCluster(
		c.Request.Context(),
		req.ClusterID,
		req.Name,
		req.Description,
		req.Labels,
	)
	if err != nil {
		logger.Errorw("Failed to update cluster", "cluster_id", req.ClusterID, "error", err.Error())
		response.InternalError(c, "Failed to update cluster", err)
		return
	}

	response.SuccessWithMessage(c, "Cluster updated successfully", cluster)
}

// DeleteCluster DELETE /api/k8s/clusters/:id
// 删除集群
func (h *K8sAPIHandler) DeleteCluster(c *gin.Context) {
	var req types.DeleteClusterRequest

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err)
		return
	}

	if err := validator.ValidateClusterID(req.ClusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Deleting cluster", "cluster_id", req.ClusterID)

	if err := h.clusterService.DeleteCluster(c.Request.Context(), req.ClusterID); err != nil {
		logger.Errorw("Failed to delete cluster", "cluster_id", req.ClusterID, "error", err.Error())
		response.InternalError(c, "Failed to delete cluster", err)
		return
	}

	response.SuccessWithMessage(c, "Cluster deleted successfully", gin.H{
		"cluster_id": req.ClusterID,
	})
}

// GetClusterHealth GET /api/k8s/clusters/:id/health
// 获取集群健康状态
func (h *K8sAPIHandler) GetClusterHealthStatus(c *gin.Context) {
	var req types.GetClusterHealthRequest

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err)
		return
	}

	if err := validator.ValidateClusterID(req.ClusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Getting cluster health", "cluster_id", req.ClusterID)

	health, err := h.clusterService.GetClusterHealth(c.Request.Context(), req.ClusterID)
	if err != nil {
		logger.Errorw("Failed to get cluster health", "cluster_id", req.ClusterID, "error", err.Error())
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
	var req types.ListNamespacesRequest
	params := pagination.Parse(c)

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err)
		return
	}

	if err := validator.ValidateClusterID(req.ClusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Listing namespaces",
		"cluster_id", req.ClusterID,
		"page", params.Page,
	)

	namespaces, total, err := h.namespaceService.ListNamespaces(
		c.Request.Context(),
		req.ClusterID,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list namespaces", "cluster_id", req.ClusterID, "error", err.Error())
		response.InternalError(c, "Failed to list namespaces", err)
		return
	}

	resp := pagination.NewResponse(namespaces, total, params)
	response.Success(c, resp)
}

// GetNamespace GET /api/k8s/clusters/:clusterId/namespaces/:name
// 获取命名空间详情
func (h *K8sAPIHandler) GetNamespace(c *gin.Context) {
	var req types.GetNamespaceRequest

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err)
		return
	}

	if err := validator.ValidateK8sName(req.Namespace); err != nil {
		response.BadRequest(c, "Invalid namespace name", err)
		return
	}

	logger.Infow("Getting namespace details",
		"cluster_id", req.ClusterID,
		"namespace", req.Namespace,
	)

	namespace, err := h.namespaceService.GetNamespace(c.Request.Context(), req.ClusterID, req.Namespace)
	if err != nil {
		logger.Errorw("Failed to get namespace",
			"cluster_id", req.ClusterID,
			"namespace", req.Namespace,
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
	var req types.CreateNamespaceRequest

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid path parameters", err)
		return
	}

	// 绑定请求体参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid namespace name", err)
		return
	}

	logger.Infow("Creating namespace",
		"cluster_id", req.ClusterID,
		"namespace", req.Name,
	)

	namespace, err := h.namespaceService.CreateNamespace(
		c.Request.Context(),
		req.ClusterID,
		req.Name,
		req.Labels,
	)
	if err != nil {
		logger.Errorw("Failed to create namespace",
			"cluster_id", req.ClusterID,
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
	var req types.DeleteNamespaceRequest

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err)
		return
	}

	if err := validator.ValidateK8sName(req.Namespace); err != nil {
		response.BadRequest(c, "Invalid namespace name", err)
		return
	}

	logger.Infow("Deleting namespace",
		"cluster_id", req.ClusterID,
		"namespace", req.Namespace,
	)

	if err := h.namespaceService.DeleteNamespace(c.Request.Context(), req.ClusterID, req.Namespace); err != nil {
		logger.Errorw("Failed to delete namespace",
			"cluster_id", req.ClusterID,
			"namespace", req.Namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete namespace", err)
		return
	}

	response.SuccessWithMessage(c, "Namespace deleted successfully", gin.H{
		"namespace": req.Namespace,
	})
}

// ===========================
// Pod 管理接口
// ===========================

// ListPods GET /api/k8s/pods
// 获取 Pod 列表
// 支持查询所有命名空间或指定命名空间的 Pods
func (h *K8sAPIHandler) ListPods(c *gin.Context) {
	var req types.ListPodsRequest
	params := pagination.Parse(c)

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err)
		return
	}

	if err := validator.ValidateClusterID(req.ClusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	// namespace 参数是可选的，如果提供则验证
	if req.Namespace != "" {
		if err := validator.ValidateK8sName(req.Namespace); err != nil {
			response.BadRequest(c, "Invalid namespace name", err)
			return
		}
	}

	logger.Infow("Listing pods",
		"cluster_id", req.ClusterID,
		"namespace", req.Namespace,
		"page", params.Page,
	)

	pods, total, err := h.podService.ListPods(
		c.Request.Context(),
		req.ClusterID,
		req.Namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list pods",
			"cluster_id", req.ClusterID,
			"namespace", req.Namespace,
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
	var req types.GetPodRequest

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err)
		return
	}

	if err := validator.ValidateClusterID(req.ClusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	if err := validator.ValidateK8sName(req.Namespace); err != nil {
		response.BadRequest(c, "Invalid namespace name", err)
		return
	}

	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid pod name", err)
		return
	}

	logger.Infow("Getting pod details",
		"cluster_id", req.ClusterID,
		"namespace", req.Namespace,
		"pod", req.Name,
	)

	pod, err := h.podService.GetPod(c.Request.Context(), req.ClusterID, req.Namespace, req.Name)
	if err != nil {
		logger.Errorw("Failed to get pod",
			"cluster_id", req.ClusterID,
			"namespace", req.Namespace,
			"pod", req.Name,
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
	var req types.DeletePodRequest

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err)
		return
	}

	if err := validator.ValidateClusterID(req.ClusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	if err := validator.ValidateK8sName(req.Namespace); err != nil {
		response.BadRequest(c, "Invalid namespace name", err)
		return
	}

	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid pod name", err)
		return
	}

	logger.Infow("Deleting pod",
		"cluster_id", req.ClusterID,
		"namespace", req.Namespace,
		"pod", req.Name,
	)

	if err := h.podService.DeletePod(c.Request.Context(), req.ClusterID, req.Namespace, req.Name); err != nil {
		logger.Errorw("Failed to delete pod",
			"cluster_id", req.ClusterID,
			"namespace", req.Namespace,
			"pod", req.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete pod", err)
		return
	}

	response.SuccessWithMessage(c, "Pod deleted successfully", gin.H{
		"pod": req.Name,
	})
}

// GetPodLogs GET /api/k8s/clusters/:clusterId/namespaces/:namespace/pods/:name/logs
// 获取 Pod 日志
func (h *K8sAPIHandler) GetPodLogs(c *gin.Context) {
	var req types.GetPodLogsRequest

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid path parameters", err)
		return
	}

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err)
		return
	}

	if err := validator.ValidateClusterID(req.ClusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	if err := validator.ValidateK8sName(req.Namespace); err != nil {
		response.BadRequest(c, "Invalid namespace name", err)
		return
	}

	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid pod name", err)
		return
	}

	// 设置默认值
	if req.TailLines == "" {
		req.TailLines = "100"
	}

	logger.Infow("Getting pod logs",
		"cluster_id", req.ClusterID,
		"namespace", req.Namespace,
		"pod", req.Name,
		"container", req.Container,
		"tail_lines", req.TailLines,
		"follow", req.Follow,
	)

	logs, err := h.podService.GetPodLogs(
		c.Request.Context(),
		req.ClusterID,
		req.Namespace,
		req.Name,
		req.Container,
		req.TailLines,
		req.Follow,
	)
	if err != nil {
		logger.Errorw("Failed to get pod logs",
			"cluster_id", req.ClusterID,
			"namespace", req.Namespace,
			"pod", req.Name,
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

// ListDeployments GET /api/k8s/deployments
// 获取 Deployment 列表
// 支持查询所有命名空间或指定命名空间的 Deployments
func (h *K8sAPIHandler) ListDeployments(c *gin.Context) {
	var req types.ListDeploymentsRequest
	params := pagination.Parse(c)

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err)
		return
	}

	if err := validator.ValidateClusterID(req.ClusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	// namespace 参数是可选的，如果提供则验证
	if req.Namespace != "" {
		if err := validator.ValidateK8sName(req.Namespace); err != nil {
			response.BadRequest(c, "Invalid namespace name", err)
			return
		}
	}

	logger.Infow("Listing deployments",
		"cluster_id", req.ClusterID,
		"namespace", req.Namespace,
	)

	deployments, total, err := h.deploymentService.ListDeployments(
		c.Request.Context(),
		req.ClusterID,
		req.Namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list deployments",
			"cluster_id", req.ClusterID,
			"namespace", req.Namespace,
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
	var req types.GetDeploymentRequest

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err)
		return
	}

	if err := validator.ValidateClusterID(req.ClusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	if err := validator.ValidateK8sName(req.Namespace); err != nil {
		response.BadRequest(c, "Invalid namespace name", err)
		return
	}

	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid deployment name", err)
		return
	}

	logger.Infow("Getting deployment details",
		"cluster_id", req.ClusterID,
		"namespace", req.Namespace,
		"deployment", req.Name,
	)

	deployment, err := h.deploymentService.GetDeployment(
		c.Request.Context(),
		req.ClusterID,
		req.Namespace,
		req.Name,
	)
	if err != nil {
		logger.Errorw("Failed to get deployment",
			"cluster_id", req.ClusterID,
			"namespace", req.Namespace,
			"deployment", req.Name,
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
	var req types.ScaleDeploymentRequest

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid path parameters", err)
		return
	}

	// 绑定请求体参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	if err := validator.ValidateClusterID(req.ClusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	if err := validator.ValidateK8sName(req.Namespace); err != nil {
		response.BadRequest(c, "Invalid namespace name", err)
		return
	}

	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid deployment name", err)
		return
	}

	if err := validator.ValidateReplicas(req.Replicas); err != nil {
		response.BadRequest(c, "Invalid replicas count", err)
		return
	}

	logger.Infow("Scaling deployment",
		"cluster_id", req.ClusterID,
		"namespace", req.Namespace,
		"deployment", req.Name,
		"replicas", req.Replicas,
	)

	deployment, err := h.deploymentService.ScaleDeployment(
		c.Request.Context(),
		req.ClusterID,
		req.Namespace,
		req.Name,
		req.Replicas,
	)
	if err != nil {
		logger.Errorw("Failed to scale deployment",
			"cluster_id", req.ClusterID,
			"namespace", req.Namespace,
			"deployment", req.Name,
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
	var req types.RestartDeploymentRequest

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err)
		return
	}

	if err := validator.ValidateClusterID(req.ClusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	if err := validator.ValidateK8sName(req.Namespace); err != nil {
		response.BadRequest(c, "Invalid namespace name", err)
		return
	}

	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid deployment name", err)
		return
	}

	logger.Infow("Restarting deployment",
		"cluster_id", req.ClusterID,
		"namespace", req.Namespace,
		"deployment", req.Name,
	)

	if err := h.deploymentService.RestartDeployment(
		c.Request.Context(),
		req.ClusterID,
		req.Namespace,
		req.Name,
	); err != nil {
		logger.Errorw("Failed to restart deployment",
			"cluster_id", req.ClusterID,
			"namespace", req.Namespace,
			"deployment", req.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to restart deployment", err)
		return
	}

	response.SuccessWithMessage(c, "Deployment restarted successfully", gin.H{
		"deployment": req.Name,
	})
}

// ===========================
// Node 管理接口
// ===========================

// ListNodes GET /api/k8s/clusters/:clusterId/nodes
// 获取 Node 列表
func (h *K8sAPIHandler) ListNodes(c *gin.Context) {
	clusterID := c.Query("clusterId")
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
	clusterID := c.Query("clusterId")
	nodeName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	nodeName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	nodeName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	nodeName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	serviceName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	serviceName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	serviceName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	statefulsetName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	statefulsetName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	statefulsetName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	statefulsetName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	daemonsetName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	daemonsetName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	daemonsetName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	configmapName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	configmapName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	configmapName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	secretName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	secretName := c.Query("name")

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
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	secretName := c.Query("name")

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

// ===========================
// Endpoints 管理接口
// ===========================

// ListEndpoints GET /api/k8s/clusters/:clusterId/namespaces/:namespace/endpoints
// 获取 Endpoints 列表
func (h *K8sAPIHandler) ListEndpoints(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing endpoints",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	endpoints, total, err := h.endpointService.ListEndpoints(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list endpoints",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list endpoints", err)
		return
	}

	resp := pagination.NewResponse(endpoints, total, params)
	response.Success(c, resp)
}

// GetEndpoint GET /api/k8s/clusters/:clusterId/namespaces/:namespace/endpoints/:name
// 获取 Endpoint 详情
func (h *K8sAPIHandler) GetEndpoint(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	endpointName := c.Query("name")

	logger.Infow("Getting endpoint details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"endpoint", endpointName,
	)

	endpoint, err := h.endpointService.GetEndpoint(
		c.Request.Context(),
		clusterID,
		namespace,
		endpointName,
	)
	if err != nil {
		logger.Errorw("Failed to get endpoint",
			"cluster_id", clusterID,
			"namespace", namespace,
			"endpoint", endpointName,
			"error", err.Error(),
		)
		response.NotFound(c, "Endpoint not found", err)
		return
	}

	response.Success(c, endpoint)
}

// DeleteEndpoint DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/endpoints/:name
// 删除 Endpoint
func (h *K8sAPIHandler) DeleteEndpoint(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	endpointName := c.Query("name")

	logger.Infow("Deleting endpoint",
		"cluster_id", clusterID,
		"namespace", namespace,
		"endpoint", endpointName,
	)

	if err := h.endpointService.DeleteEndpoint(c.Request.Context(), clusterID, namespace, endpointName); err != nil {
		logger.Errorw("Failed to delete endpoint",
			"cluster_id", clusterID,
			"namespace", namespace,
			"endpoint", endpointName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete endpoint", err)
		return
	}

	response.SuccessWithMessage(c, "Endpoint deleted successfully", gin.H{
		"endpoint": endpointName,
	})
}

// ===========================
// PersistentVolumeClaim 管理接口
// ===========================

// ListPVCs GET /api/k8s/clusters/:clusterId/namespaces/:namespace/persistentvolumeclaims
// 获取 PVC 列表
func (h *K8sAPIHandler) ListPVCs(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing pvcs",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	pvcs, total, err := h.pvcService.ListPVCs(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list pvcs",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list pvcs", err)
		return
	}

	resp := pagination.NewResponse(pvcs, total, params)
	response.Success(c, resp)
}

// GetPVC GET /api/k8s/clusters/:clusterId/namespaces/:namespace/persistentvolumeclaims/:name
// 获取 PVC 详情
func (h *K8sAPIHandler) GetPVC(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	pvcName := c.Query("name")

	logger.Infow("Getting pvc details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"pvc", pvcName,
	)

	pvc, err := h.pvcService.GetPVC(c.Request.Context(), clusterID, namespace, pvcName)
	if err != nil {
		logger.Errorw("Failed to get pvc",
			"cluster_id", clusterID,
			"namespace", namespace,
			"pvc", pvcName,
			"error", err.Error(),
		)
		response.NotFound(c, "PVC not found", err)
		return
	}

	response.Success(c, pvc)
}

// DeletePVC DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/persistentvolumeclaims/:name
// 删除 PVC
func (h *K8sAPIHandler) DeletePVC(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	pvcName := c.Query("name")

	logger.Infow("Deleting pvc",
		"cluster_id", clusterID,
		"namespace", namespace,
		"pvc", pvcName,
	)

	if err := h.pvcService.DeletePVC(c.Request.Context(), clusterID, namespace, pvcName); err != nil {
		logger.Errorw("Failed to delete pvc",
			"cluster_id", clusterID,
			"namespace", namespace,
			"pvc", pvcName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete pvc", err)
		return
	}

	response.SuccessWithMessage(c, "PVC deleted successfully", gin.H{
		"pvc": pvcName,
	})
}

// ===========================
// PersistentVolume 管理接口
// ===========================

// ListPVs GET /api/k8s/clusters/:clusterId/persistentvolumes
// 获取 PV 列表
func (h *K8sAPIHandler) ListPVs(c *gin.Context) {
	clusterID := c.Query("clusterId")
	params := pagination.Parse(c)

	logger.Infow("Listing pvs",
		"cluster_id", clusterID,
	)

	pvs, total, err := h.pvService.ListPVs(
		c.Request.Context(),
		clusterID,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list pvs",
			"cluster_id", clusterID,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list pvs", err)
		return
	}

	resp := pagination.NewResponse(pvs, total, params)
	response.Success(c, resp)
}

// GetPV GET /api/k8s/clusters/:clusterId/persistentvolumes/:name
// 获取 PV 详情
func (h *K8sAPIHandler) GetPV(c *gin.Context) {
	clusterID := c.Query("clusterId")
	pvName := c.Query("name")

	logger.Infow("Getting pv details",
		"cluster_id", clusterID,
		"pv", pvName,
	)

	pv, err := h.pvService.GetPV(c.Request.Context(), clusterID, pvName)
	if err != nil {
		logger.Errorw("Failed to get pv",
			"cluster_id", clusterID,
			"pv", pvName,
			"error", err.Error(),
		)
		response.NotFound(c, "PV not found", err)
		return
	}

	response.Success(c, pv)
}

// DeletePV DELETE /api/k8s/clusters/:clusterId/persistentvolumes/:name
// 删除 PV
func (h *K8sAPIHandler) DeletePV(c *gin.Context) {
	clusterID := c.Query("clusterId")
	pvName := c.Query("name")

	logger.Infow("Deleting pv",
		"cluster_id", clusterID,
		"pv", pvName,
	)

	if err := h.pvService.DeletePV(c.Request.Context(), clusterID, pvName); err != nil {
		logger.Errorw("Failed to delete pv",
			"cluster_id", clusterID,
			"pv", pvName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete pv", err)
		return
	}

	response.SuccessWithMessage(c, "PV deleted successfully", gin.H{
		"pv": pvName,
	})
}

// ===========================
// EndpointSlice 管理接口
// ===========================

// ListEndpointSlices GET /api/k8s/clusters/:clusterId/namespaces/:namespace/endpointslices
// 获取 EndpointSlice 列表
func (h *K8sAPIHandler) ListEndpointSlices(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing endpointslices",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	endpointslices, total, err := h.endpointsliceService.ListEndpointSlices(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list endpointslices",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list endpointslices", err)
		return
	}

	resp := pagination.NewResponse(endpointslices, total, params)
	response.Success(c, resp)
}

// GetEndpointSlice GET /api/k8s/clusters/:clusterId/namespaces/:namespace/endpointslices/:name
// 获取 EndpointSlice 详情
func (h *K8sAPIHandler) GetEndpointSlice(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	sliceName := c.Query("name")

	logger.Infow("Getting endpointslice details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"endpointslice", sliceName,
	)

	slice, err := h.endpointsliceService.GetEndpointSlice(c.Request.Context(), clusterID, namespace, sliceName)
	if err != nil {
		logger.Errorw("Failed to get endpointslice",
			"cluster_id", clusterID,
			"namespace", namespace,
			"endpointslice", sliceName,
			"error", err.Error(),
		)
		response.NotFound(c, "EndpointSlice not found", err)
		return
	}

	response.Success(c, slice)
}

// DeleteEndpointSlice DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/endpointslices/:name
// 删除 EndpointSlice
func (h *K8sAPIHandler) DeleteEndpointSlice(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	sliceName := c.Query("name")

	logger.Infow("Deleting endpointslice",
		"cluster_id", clusterID,
		"namespace", namespace,
		"endpointslice", sliceName,
	)

	if err := h.endpointsliceService.DeleteEndpointSlice(c.Request.Context(), clusterID, namespace, sliceName); err != nil {
		logger.Errorw("Failed to delete endpointslice",
			"cluster_id", clusterID,
			"namespace", namespace,
			"endpointslice", sliceName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete endpointslice", err)
		return
	}

	response.SuccessWithMessage(c, "EndpointSlice deleted successfully", gin.H{
		"endpointslice": sliceName,
	})
}

// ===========================
// HorizontalPodAutoscaler 管理接口
// ===========================

// ListHPAs GET /api/k8s/clusters/:clusterId/namespaces/:namespace/horizontalpodautoscalers
// 获取 HPA 列表
func (h *K8sAPIHandler) ListHPAs(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing hpas",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	hpas, total, err := h.hpaService.ListHPAs(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list hpas",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list hpas", err)
		return
	}

	resp := pagination.NewResponse(hpas, total, params)
	response.Success(c, resp)
}

// GetHPA GET /api/k8s/clusters/:clusterId/namespaces/:namespace/horizontalpodautoscalers/:name
// 获取 HPA 详情
func (h *K8sAPIHandler) GetHPA(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	hpaName := c.Query("name")

	logger.Infow("Getting hpa details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"hpa", hpaName,
	)

	hpa, err := h.hpaService.GetHPA(c.Request.Context(), clusterID, namespace, hpaName)
	if err != nil {
		logger.Errorw("Failed to get hpa",
			"cluster_id", clusterID,
			"namespace", namespace,
			"hpa", hpaName,
			"error", err.Error(),
		)
		response.NotFound(c, "HPA not found", err)
		return
	}

	response.Success(c, hpa)
}

// DeleteHPA DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/horizontalpodautoscalers/:name
// 删除 HPA
func (h *K8sAPIHandler) DeleteHPA(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	hpaName := c.Query("name")

	logger.Infow("Deleting hpa",
		"cluster_id", clusterID,
		"namespace", namespace,
		"hpa", hpaName,
	)

	if err := h.hpaService.DeleteHPA(c.Request.Context(), clusterID, namespace, hpaName); err != nil {
		logger.Errorw("Failed to delete hpa",
			"cluster_id", clusterID,
			"namespace", namespace,
			"hpa", hpaName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete hpa", err)
		return
	}

	response.SuccessWithMessage(c, "HPA deleted successfully", gin.H{
		"hpa": hpaName,
	})
}

// ===========================
// Event 管理接口
// ===========================

// ListEvents GET /api/k8s/clusters/:clusterId/namespaces/:namespace/events
// 获取 Event 列表
func (h *K8sAPIHandler) ListEvents(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	// 支持按事件类型过滤
	eventType := c.Query("type")
	// 支持按事件原因过滤
	eventReason := c.Query("reason")

	logger.Infow("Listing events",
		"cluster_id", clusterID,
		"namespace", namespace,
		"type", eventType,
		"reason", eventReason,
	)

	events, total, err := h.eventService.ListEvents(
		c.Request.Context(),
		clusterID,
		namespace,
		eventType,
		eventReason,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list events",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list events", err)
		return
	}

	resp := pagination.NewResponse(events, total, params)
	response.Success(c, resp)
}

// GetEvent GET /api/k8s/clusters/:clusterId/namespaces/:namespace/events/:name
// 获取 Event 详情
func (h *K8sAPIHandler) GetEvent(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	eventName := c.Query("name")

	logger.Infow("Getting event details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"event", eventName,
	)

	event, err := h.eventService.GetEvent(c.Request.Context(), clusterID, namespace, eventName)
	if err != nil {
		logger.Errorw("Failed to get event",
			"cluster_id", clusterID,
			"namespace", namespace,
			"event", eventName,
			"error", err.Error(),
		)
		response.NotFound(c, "Event not found", err)
		return
	}

	response.Success(c, event)
}

// ===========================
// RoleBinding 管理接口
// ===========================

// ListRoleBindings GET /api/k8s/clusters/:clusterId/namespaces/:namespace/rolebindings
// 获取 RoleBinding 列表
func (h *K8sAPIHandler) ListRoleBindings(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing rolebindings",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	rolebindings, total, err := h.rolebindingService.ListRoleBindings(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list rolebindings",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list rolebindings", err)
		return
	}

	resp := pagination.NewResponse(rolebindings, total, params)
	response.Success(c, resp)
}

// GetRoleBinding GET /api/k8s/clusters/:clusterId/namespaces/:namespace/rolebindings/:name
// 获取 RoleBinding 详情
func (h *K8sAPIHandler) GetRoleBinding(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	rbName := c.Query("name")

	logger.Infow("Getting rolebinding details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"rolebinding", rbName,
	)

	rb, err := h.rolebindingService.GetRoleBinding(c.Request.Context(), clusterID, namespace, rbName)
	if err != nil {
		logger.Errorw("Failed to get rolebinding",
			"cluster_id", clusterID,
			"namespace", namespace,
			"rolebinding", rbName,
			"error", err.Error(),
		)
		response.NotFound(c, "RoleBinding not found", err)
		return
	}

	response.Success(c, rb)
}

// DeleteRoleBinding DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/rolebindings/:name
// 删除 RoleBinding
func (h *K8sAPIHandler) DeleteRoleBinding(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	rbName := c.Query("name")

	logger.Infow("Deleting rolebinding",
		"cluster_id", clusterID,
		"namespace", namespace,
		"rolebinding", rbName,
	)

	if err := h.rolebindingService.DeleteRoleBinding(c.Request.Context(), clusterID, namespace, rbName); err != nil {
		logger.Errorw("Failed to delete rolebinding",
			"cluster_id", clusterID,
			"namespace", namespace,
			"rolebinding", rbName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete rolebinding", err)
		return
	}

	response.SuccessWithMessage(c, "RoleBinding deleted successfully", gin.H{
		"rolebinding": rbName,
	})
}

// ===========================
// ClusterRole 管理接口
// ===========================

// ListClusterRoles GET /api/k8s/clusters/:clusterId/clusterroles
// 获取 ClusterRole 列表
func (h *K8sAPIHandler) ListClusterRoles(c *gin.Context) {
	clusterID := c.Query("clusterId")
	params := pagination.Parse(c)

	logger.Infow("Listing clusterroles",
		"cluster_id", clusterID,
	)

	clusterroles, total, err := h.clusterroleService.ListClusterRoles(
		c.Request.Context(),
		clusterID,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list clusterroles",
			"cluster_id", clusterID,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list clusterroles", err)
		return
	}

	resp := pagination.NewResponse(clusterroles, total, params)
	response.Success(c, resp)
}

// GetClusterRole GET /api/k8s/clusters/:clusterId/clusterroles/:name
// 获取 ClusterRole 详情
func (h *K8sAPIHandler) GetClusterRole(c *gin.Context) {
	clusterID := c.Query("clusterId")
	crName := c.Query("name")

	logger.Infow("Getting clusterrole details",
		"cluster_id", clusterID,
		"clusterrole", crName,
	)

	cr, err := h.clusterroleService.GetClusterRole(c.Request.Context(), clusterID, crName)
	if err != nil {
		logger.Errorw("Failed to get clusterrole",
			"cluster_id", clusterID,
			"clusterrole", crName,
			"error", err.Error(),
		)
		response.NotFound(c, "ClusterRole not found", err)
		return
	}

	response.Success(c, cr)
}

// DeleteClusterRole DELETE /api/k8s/clusters/:clusterId/clusterroles/:name
// 删除 ClusterRole
func (h *K8sAPIHandler) DeleteClusterRole(c *gin.Context) {
	clusterID := c.Query("clusterId")
	crName := c.Query("name")

	logger.Infow("Deleting clusterrole",
		"cluster_id", clusterID,
		"clusterrole", crName,
	)

	if err := h.clusterroleService.DeleteClusterRole(c.Request.Context(), clusterID, crName); err != nil {
		logger.Errorw("Failed to delete clusterrole",
			"cluster_id", clusterID,
			"clusterrole", crName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete clusterrole", err)
		return
	}

	response.SuccessWithMessage(c, "ClusterRole deleted successfully", gin.H{
		"clusterrole": crName,
	})
}

// ===========================
// PriorityClass 管理接口
// ===========================

// ListPriorityClasses GET /api/k8s/clusters/:clusterId/priorityclasses
// 获取 PriorityClass 列表
func (h *K8sAPIHandler) ListPriorityClasses(c *gin.Context) {
	clusterID := c.Query("clusterId")
	params := pagination.Parse(c)

	logger.Infow("Listing priorityclasses",
		"cluster_id", clusterID,
	)

	priorityclasses, total, err := h.priorityclassService.ListPriorityClasses(
		c.Request.Context(),
		clusterID,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list priorityclasses",
			"cluster_id", clusterID,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list priorityclasses", err)
		return
	}

	resp := pagination.NewResponse(priorityclasses, total, params)
	response.Success(c, resp)
}

// GetPriorityClass GET /api/k8s/clusters/:clusterId/priorityclasses/:name
// 获取 PriorityClass 详情
func (h *K8sAPIHandler) GetPriorityClass(c *gin.Context) {
	clusterID := c.Query("clusterId")
	pcName := c.Query("name")

	logger.Infow("Getting priorityclass details",
		"cluster_id", clusterID,
		"priorityclass", pcName,
	)

	pc, err := h.priorityclassService.GetPriorityClass(c.Request.Context(), clusterID, pcName)
	if err != nil {
		logger.Errorw("Failed to get priorityclass",
			"cluster_id", clusterID,
			"priorityclass", pcName,
			"error", err.Error(),
		)
		response.NotFound(c, "PriorityClass not found", err)
		return
	}

	response.Success(c, pc)
}

// DeletePriorityClass DELETE /api/k8s/clusters/:clusterId/priorityclasses/:name
// 删除 PriorityClass
func (h *K8sAPIHandler) DeletePriorityClass(c *gin.Context) {
	clusterID := c.Query("clusterId")
	pcName := c.Query("name")

	logger.Infow("Deleting priorityclass",
		"cluster_id", clusterID,
		"priorityclass", pcName,
	)

	if err := h.priorityclassService.DeletePriorityClass(c.Request.Context(), clusterID, pcName); err != nil {
		logger.Errorw("Failed to delete priorityclass",
			"cluster_id", clusterID,
			"priorityclass", pcName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete priorityclass", err)
		return
	}

	response.SuccessWithMessage(c, "PriorityClass deleted successfully", gin.H{
		"priorityclass": pcName,
	})
}

// ===========================
// Role 管理接口
// ===========================

// ListRoles GET /api/k8s/clusters/:clusterId/namespaces/:namespace/roles
// 获取 Role 列表
func (h *K8sAPIHandler) ListRoles(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing roles",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	roles, total, err := h.roleService.ListRoles(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list roles",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list roles", err)
		return
	}

	resp := pagination.NewResponse(roles, total, params)
	response.Success(c, resp)
}

// GetRole GET /api/k8s/clusters/:clusterId/namespaces/:namespace/roles/:name
// 获取 Role 详情
func (h *K8sAPIHandler) GetRole(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	roleName := c.Query("name")

	logger.Infow("Getting role details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"role", roleName,
	)

	role, err := h.roleService.GetRole(c.Request.Context(), clusterID, namespace, roleName)
	if err != nil {
		logger.Errorw("Failed to get role",
			"cluster_id", clusterID,
			"namespace", namespace,
			"role", roleName,
			"error", err.Error(),
		)
		response.NotFound(c, "Role not found", err)
		return
	}

	response.Success(c, role)
}

// DeleteRole DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/roles/:name
// 删除 Role
func (h *K8sAPIHandler) DeleteRole(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	roleName := c.Query("name")

	logger.Infow("Deleting role",
		"cluster_id", clusterID,
		"namespace", namespace,
		"role", roleName,
	)

	if err := h.roleService.DeleteRole(c.Request.Context(), clusterID, namespace, roleName); err != nil {
		logger.Errorw("Failed to delete role",
			"cluster_id", clusterID,
			"namespace", namespace,
			"role", roleName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete role", err)
		return
	}

	response.SuccessWithMessage(c, "Role deleted successfully", gin.H{
		"role": roleName,
	})
}


// ===========================
// StorageClass 管理接口
// ===========================

// ListStorageClasses GET /api/k8s/clusters/:clusterId/storageclasses
// 获取 StorageClass 列表
func (h *K8sAPIHandler) ListStorageClasses(c *gin.Context) {
	clusterID := c.Query("clusterId")
	params := pagination.Parse(c)

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Listing storageclasses",
		"cluster_id", clusterID,
		"page", params.Page,
	)

	storageClasses, total, err := h.storageclassService.ListStorageClasses(
		c.Request.Context(),
		clusterID,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list storageclasses", "cluster_id", clusterID, "error", err.Error())
		response.InternalError(c, "Failed to list storageclasses", err)
		return
	}

	resp := pagination.NewResponse(storageClasses, total, params)
	response.Success(c, resp)
}

// GetStorageClass GET /api/k8s/clusters/:clusterId/storageclasses/:name
// 获取 StorageClass 详情
func (h *K8sAPIHandler) GetStorageClass(c *gin.Context) {
	clusterID := c.Query("clusterId")
	name := c.Query("name")

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	if err := validator.ValidateK8sName(name); err != nil {
		response.BadRequest(c, "Invalid storageclass name", err)
		return
	}

	logger.Infow("Getting storageclass details",
		"cluster_id", clusterID,
		"storageclass", name,
	)

	storageClass, err := h.storageclassService.GetStorageClass(c.Request.Context(), clusterID, name)
	if err != nil {
		logger.Errorw("Failed to get storageclass",
			"cluster_id", clusterID,
			"storageclass", name,
			"error", err.Error(),
		)
		response.NotFound(c, "StorageClass not found", err)
		return
	}

	response.Success(c, storageClass)
}

// DeleteStorageClass DELETE /api/k8s/clusters/:clusterId/storageclasses/:name
// 删除 StorageClass
func (h *K8sAPIHandler) DeleteStorageClass(c *gin.Context) {
	clusterID := c.Query("clusterId")
	name := c.Query("name")

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	if err := validator.ValidateK8sName(name); err != nil {
		response.BadRequest(c, "Invalid storageclass name", err)
		return
	}

	logger.Infow("Deleting storageclass",
		"cluster_id", clusterID,
		"storageclass", name,
	)

	if err := h.storageclassService.DeleteStorageClass(c.Request.Context(), clusterID, name); err != nil {
		logger.Errorw("Failed to delete storageclass",
			"cluster_id", clusterID,
			"storageclass", name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete storageclass", err)
		return
	}

	response.SuccessWithMessage(c, "StorageClass deleted successfully", gin.H{
		"storageclass": name,
	})
}

// ===========================
// Job 管理接口
// ===========================

// ListJobs GET /api/k8s/jobs
// 获取 Job 列表
func (h *K8sAPIHandler) ListJobs(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing jobs",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	jobs, total, err := h.jobService.ListJobs(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list jobs",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list jobs", err)
		return
	}

	resp := pagination.NewResponse(jobs, total, params)
	response.Success(c, resp)
}

// GetJob GET /api/k8s/job
// 获取 Job 详情
func (h *K8sAPIHandler) GetJob(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	jobName := c.Query("name")

	logger.Infow("Getting job details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"job", jobName,
	)

	job, err := h.jobService.GetJob(c.Request.Context(), clusterID, namespace, jobName)
	if err != nil {
		logger.Errorw("Failed to get job",
			"cluster_id", clusterID,
			"namespace", namespace,
			"job", jobName,
			"error", err.Error(),
		)
		response.NotFound(c, "Job not found", err)
		return
	}

	response.Success(c, job)
}

// CreateJob POST /api/k8s/jobs
// 创建 Job
func (h *K8sAPIHandler) CreateJob(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

	var job batchv1.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		response.BadRequest(c, "Invalid job data", err)
		return
	}

	logger.Infow("Creating job",
		"cluster_id", clusterID,
		"namespace", namespace,
		"job", job.Name,
	)

	createdJob, err := h.jobService.CreateJob(c.Request.Context(), clusterID, namespace, &job)
	if err != nil {
		logger.Errorw("Failed to create job",
			"cluster_id", clusterID,
			"namespace", namespace,
			"job", job.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to create job", err)
		return
	}

	response.SuccessWithMessage(c, "Job created successfully", createdJob)
}

// DeleteJob DELETE /api/k8s/job
// 删除 Job
func (h *K8sAPIHandler) DeleteJob(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	jobName := c.Query("name")

	logger.Infow("Deleting job",
		"cluster_id", clusterID,
		"namespace", namespace,
		"job", jobName,
	)

	if err := h.jobService.DeleteJob(c.Request.Context(), clusterID, namespace, jobName); err != nil {
		logger.Errorw("Failed to delete job",
			"cluster_id", clusterID,
			"namespace", namespace,
			"job", jobName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete job", err)
		return
	}

	response.SuccessWithMessage(c, "Job deleted successfully", gin.H{
		"job": jobName,
	})
}

// ===========================
// CronJob 管理接口
// ===========================

// ListCronJobs GET /api/k8s/cronjobs
// 获取 CronJob 列表
func (h *K8sAPIHandler) ListCronJobs(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing cronjobs",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	cronjobs, total, err := h.cronjobService.ListCronJobs(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list cronjobs",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list cronjobs", err)
		return
	}

	resp := pagination.NewResponse(cronjobs, total, params)
	response.Success(c, resp)
}

// GetCronJob GET /api/k8s/cronjob
// 获取 CronJob 详情
func (h *K8sAPIHandler) GetCronJob(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	cronjobName := c.Query("name")

	logger.Infow("Getting cronjob details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"cronjob", cronjobName,
	)

	cronjob, err := h.cronjobService.GetCronJob(c.Request.Context(), clusterID, namespace, cronjobName)
	if err != nil {
		logger.Errorw("Failed to get cronjob",
			"cluster_id", clusterID,
			"namespace", namespace,
			"cronjob", cronjobName,
			"error", err.Error(),
		)
		response.NotFound(c, "CronJob not found", err)
		return
	}

	response.Success(c, cronjob)
}

// CreateCronJob POST /api/k8s/cronjobs
// 创建 CronJob
func (h *K8sAPIHandler) CreateCronJob(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

	var cronjob batchv1.CronJob
	if err := c.ShouldBindJSON(&cronjob); err != nil {
		response.BadRequest(c, "Invalid cronjob data", err)
		return
	}

	logger.Infow("Creating cronjob",
		"cluster_id", clusterID,
		"namespace", namespace,
		"cronjob", cronjob.Name,
	)

	createdCronJob, err := h.cronjobService.CreateCronJob(c.Request.Context(), clusterID, namespace, &cronjob)
	if err != nil {
		logger.Errorw("Failed to create cronjob",
			"cluster_id", clusterID,
			"namespace", namespace,
			"cronjob", cronjob.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to create cronjob", err)
		return
	}

	response.SuccessWithMessage(c, "CronJob created successfully", createdCronJob)
}

// UpdateCronJob PUT /api/k8s/cronjob
// 更新 CronJob
func (h *K8sAPIHandler) UpdateCronJob(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

	var cronjob batchv1.CronJob
	if err := c.ShouldBindJSON(&cronjob); err != nil {
		response.BadRequest(c, "Invalid cronjob data", err)
		return
	}

	logger.Infow("Updating cronjob",
		"cluster_id", clusterID,
		"namespace", namespace,
		"cronjob", cronjob.Name,
	)

	updatedCronJob, err := h.cronjobService.UpdateCronJob(c.Request.Context(), clusterID, namespace, &cronjob)
	if err != nil {
		logger.Errorw("Failed to update cronjob",
			"cluster_id", clusterID,
			"namespace", namespace,
			"cronjob", cronjob.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to update cronjob", err)
		return
	}

	response.SuccessWithMessage(c, "CronJob updated successfully", updatedCronJob)
}

// DeleteCronJob DELETE /api/k8s/cronjob
// 删除 CronJob
func (h *K8sAPIHandler) DeleteCronJob(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	cronjobName := c.Query("name")

	logger.Infow("Deleting cronjob",
		"cluster_id", clusterID,
		"namespace", namespace,
		"cronjob", cronjobName,
	)

	if err := h.cronjobService.DeleteCronJob(c.Request.Context(), clusterID, namespace, cronjobName); err != nil {
		logger.Errorw("Failed to delete cronjob",
			"cluster_id", clusterID,
			"namespace", namespace,
			"cronjob", cronjobName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete cronjob", err)
		return
	}

	response.SuccessWithMessage(c, "CronJob deleted successfully", gin.H{
		"cronjob": cronjobName,
	})
}

// ===========================
// Ingress 管理接口
// ===========================

// ListIngresses GET /api/k8s/ingresses
// 获取 Ingress 列表
func (h *K8sAPIHandler) ListIngresses(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing ingresses",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	ingresses, total, err := h.ingressService.ListIngresses(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list ingresses",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list ingresses", err)
		return
	}

	resp := pagination.NewResponse(ingresses, total, params)
	response.Success(c, resp)
}

// GetIngress GET /api/k8s/ingress
// 获取 Ingress 详情
func (h *K8sAPIHandler) GetIngress(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	ingressName := c.Query("name")

	logger.Infow("Getting ingress details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"ingress", ingressName,
	)

	ingress, err := h.ingressService.GetIngress(c.Request.Context(), clusterID, namespace, ingressName)
	if err != nil {
		logger.Errorw("Failed to get ingress",
			"cluster_id", clusterID,
			"namespace", namespace,
			"ingress", ingressName,
			"error", err.Error(),
		)
		response.NotFound(c, "Ingress not found", err)
		return
	}

	response.Success(c, ingress)
}

// CreateIngress POST /api/k8s/ingresses
// 创建 Ingress
func (h *K8sAPIHandler) CreateIngress(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

	var ingress networkingv1.Ingress
	if err := c.ShouldBindJSON(&ingress); err != nil {
		response.BadRequest(c, "Invalid ingress data", err)
		return
	}

	logger.Infow("Creating ingress",
		"cluster_id", clusterID,
		"namespace", namespace,
		"ingress", ingress.Name,
	)

	createdIngress, err := h.ingressService.CreateIngress(c.Request.Context(), clusterID, namespace, &ingress)
	if err != nil {
		logger.Errorw("Failed to create ingress",
			"cluster_id", clusterID,
			"namespace", namespace,
			"ingress", ingress.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to create ingress", err)
		return
	}

	response.SuccessWithMessage(c, "Ingress created successfully", createdIngress)
}

// UpdateIngress PUT /api/k8s/ingress
// 更新 Ingress
func (h *K8sAPIHandler) UpdateIngress(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

	var ingress networkingv1.Ingress
	if err := c.ShouldBindJSON(&ingress); err != nil {
		response.BadRequest(c, "Invalid ingress data", err)
		return
	}

	logger.Infow("Updating ingress",
		"cluster_id", clusterID,
		"namespace", namespace,
		"ingress", ingress.Name,
	)

	updatedIngress, err := h.ingressService.UpdateIngress(c.Request.Context(), clusterID, namespace, &ingress)
	if err != nil {
		logger.Errorw("Failed to update ingress",
			"cluster_id", clusterID,
			"namespace", namespace,
			"ingress", ingress.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to update ingress", err)
		return
	}

	response.SuccessWithMessage(c, "Ingress updated successfully", updatedIngress)
}

// DeleteIngress DELETE /api/k8s/ingress
// 删除 Ingress
func (h *K8sAPIHandler) DeleteIngress(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	ingressName := c.Query("name")

	logger.Infow("Deleting ingress",
		"cluster_id", clusterID,
		"namespace", namespace,
		"ingress", ingressName,
	)

	if err := h.ingressService.DeleteIngress(c.Request.Context(), clusterID, namespace, ingressName); err != nil {
		logger.Errorw("Failed to delete ingress",
			"cluster_id", clusterID,
			"namespace", namespace,
			"ingress", ingressName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete ingress", err)
		return
	}

	response.SuccessWithMessage(c, "Ingress deleted successfully", gin.H{
		"ingress": ingressName,
	})
}

// ===========================
// NetworkPolicy 管理接口
// ===========================

// ListNetworkPolicies GET /api/k8s/networkpolicies
// 获取 NetworkPolicy 列表
func (h *K8sAPIHandler) ListNetworkPolicies(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing networkpolicies",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	networkpolicies, total, err := h.networkpolicyService.ListNetworkPolicies(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list networkpolicies",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list networkpolicies", err)
		return
	}

	resp := pagination.NewResponse(networkpolicies, total, params)
	response.Success(c, resp)
}

// GetNetworkPolicy GET /api/k8s/networkpolicy
// 获取 NetworkPolicy 详情
func (h *K8sAPIHandler) GetNetworkPolicy(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	networkpolicyName := c.Query("name")

	logger.Infow("Getting networkpolicy details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"networkpolicy", networkpolicyName,
	)

	networkpolicy, err := h.networkpolicyService.GetNetworkPolicy(c.Request.Context(), clusterID, namespace, networkpolicyName)
	if err != nil {
		logger.Errorw("Failed to get networkpolicy",
			"cluster_id", clusterID,
			"namespace", namespace,
			"networkpolicy", networkpolicyName,
			"error", err.Error(),
		)
		response.NotFound(c, "NetworkPolicy not found", err)
		return
	}

	response.Success(c, networkpolicy)
}

// CreateNetworkPolicy POST /api/k8s/networkpolicies
// 创建 NetworkPolicy
func (h *K8sAPIHandler) CreateNetworkPolicy(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

	var networkpolicy networkingv1.NetworkPolicy
	if err := c.ShouldBindJSON(&networkpolicy); err != nil {
		response.BadRequest(c, "Invalid networkpolicy data", err)
		return
	}

	logger.Infow("Creating networkpolicy",
		"cluster_id", clusterID,
		"namespace", namespace,
		"networkpolicy", networkpolicy.Name,
	)

	createdNetworkPolicy, err := h.networkpolicyService.CreateNetworkPolicy(c.Request.Context(), clusterID, namespace, &networkpolicy)
	if err != nil {
		logger.Errorw("Failed to create networkpolicy",
			"cluster_id", clusterID,
			"namespace", namespace,
			"networkpolicy", networkpolicy.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to create networkpolicy", err)
		return
	}

	response.SuccessWithMessage(c, "NetworkPolicy created successfully", createdNetworkPolicy)
}

// UpdateNetworkPolicy PUT /api/k8s/networkpolicy
// 更新 NetworkPolicy
func (h *K8sAPIHandler) UpdateNetworkPolicy(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

	var networkpolicy networkingv1.NetworkPolicy
	if err := c.ShouldBindJSON(&networkpolicy); err != nil {
		response.BadRequest(c, "Invalid networkpolicy data", err)
		return
	}

	logger.Infow("Updating networkpolicy",
		"cluster_id", clusterID,
		"namespace", namespace,
		"networkpolicy", networkpolicy.Name,
	)

	updatedNetworkPolicy, err := h.networkpolicyService.UpdateNetworkPolicy(c.Request.Context(), clusterID, namespace, &networkpolicy)
	if err != nil {
		logger.Errorw("Failed to update networkpolicy",
			"cluster_id", clusterID,
			"namespace", namespace,
			"networkpolicy", networkpolicy.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to update networkpolicy", err)
		return
	}

	response.SuccessWithMessage(c, "NetworkPolicy updated successfully", updatedNetworkPolicy)
}

// DeleteNetworkPolicy DELETE /api/k8s/networkpolicy
// 删除 NetworkPolicy
func (h *K8sAPIHandler) DeleteNetworkPolicy(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	networkpolicyName := c.Query("name")

	logger.Infow("Deleting networkpolicy",
		"cluster_id", clusterID,
		"namespace", namespace,
		"networkpolicy", networkpolicyName,
	)

	if err := h.networkpolicyService.DeleteNetworkPolicy(c.Request.Context(), clusterID, namespace, networkpolicyName); err != nil {
		logger.Errorw("Failed to delete networkpolicy",
			"cluster_id", clusterID,
			"namespace", namespace,
			"networkpolicy", networkpolicyName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete networkpolicy", err)
		return
	}

	response.SuccessWithMessage(c, "NetworkPolicy deleted successfully", gin.H{
		"networkpolicy": networkpolicyName,
	})
}

// ===========================
// ReplicaSet 管理接口
// ===========================

// ListReplicaSets GET /api/k8s/replicasets
// 获取 ReplicaSet 列表
func (h *K8sAPIHandler) ListReplicaSets(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing replicasets",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	replicasets, total, err := h.replicasetService.ListReplicaSets(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list replicasets",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list replicasets", err)
		return
	}

	resp := pagination.NewResponse(replicasets, total, params)
	response.Success(c, resp)
}

// GetReplicaSet GET /api/k8s/replicaset
// 获取 ReplicaSet 详情
func (h *K8sAPIHandler) GetReplicaSet(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	replicasetName := c.Query("name")

	logger.Infow("Getting replicaset details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"replicaset", replicasetName,
	)

	replicaset, err := h.replicasetService.GetReplicaSet(c.Request.Context(), clusterID, namespace, replicasetName)
	if err != nil {
		logger.Errorw("Failed to get replicaset",
			"cluster_id", clusterID,
			"namespace", namespace,
			"replicaset", replicasetName,
			"error", err.Error(),
		)
		response.NotFound(c, "ReplicaSet not found", err)
		return
	}

	response.Success(c, replicaset)
}

// ScaleReplicaSet PUT /api/k8s/replicaset/scale
// 扩缩容 ReplicaSet
func (h *K8sAPIHandler) ScaleReplicaSet(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	replicasetName := c.Query("name")

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

	logger.Infow("Scaling replicaset",
		"cluster_id", clusterID,
		"namespace", namespace,
		"replicaset", replicasetName,
		"replicas", req.Replicas,
	)

	replicaset, err := h.replicasetService.ScaleReplicaSet(
		c.Request.Context(),
		clusterID,
		namespace,
		replicasetName,
		req.Replicas,
	)
	if err != nil {
		logger.Errorw("Failed to scale replicaset",
			"cluster_id", clusterID,
			"namespace", namespace,
			"replicaset", replicasetName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to scale replicaset", err)
		return
	}

	response.SuccessWithMessage(c, "ReplicaSet scaled successfully", replicaset)
}

// DeleteReplicaSet DELETE /api/k8s/replicaset
// 删除 ReplicaSet
func (h *K8sAPIHandler) DeleteReplicaSet(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	replicasetName := c.Query("name")

	logger.Infow("Deleting replicaset",
		"cluster_id", clusterID,
		"namespace", namespace,
		"replicaset", replicasetName,
	)

	if err := h.replicasetService.DeleteReplicaSet(c.Request.Context(), clusterID, namespace, replicasetName); err != nil {
		logger.Errorw("Failed to delete replicaset",
			"cluster_id", clusterID,
			"namespace", namespace,
			"replicaset", replicasetName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete replicaset", err)
		return
	}

	response.SuccessWithMessage(c, "ReplicaSet deleted successfully", gin.H{
		"replicaset": replicasetName,
	})
}

// ===========================
// LimitRange 管理接口
// ===========================

// ListLimitRanges GET /api/k8s/limitranges
// 获取 LimitRange 列表
func (h *K8sAPIHandler) ListLimitRanges(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing limitranges",
		"cluster_id", clusterID,
		"namespace", namespace,
		"page", params.Page,
	)

	limitranges, total, err := h.limitrangeService.ListLimitRanges(c.Request.Context(), clusterID, namespace, params.GetOffset(), params.GetLimit())
	if err != nil {
		logger.Errorw("Failed to list limitranges",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list limitranges", err)
		return
	}

	resp := pagination.NewResponse(limitranges, total, params)
	response.Success(c, resp)
}

// GetLimitRange GET /api/k8s/limitrange
// 获取 LimitRange 详情
func (h *K8sAPIHandler) GetLimitRange(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	limitrangeName := c.Query("name")

	logger.Infow("Getting limitrange",
		"cluster_id", clusterID,
		"namespace", namespace,
		"limitrange", limitrangeName,
	)

	limitrange, err := h.limitrangeService.GetLimitRange(c.Request.Context(), clusterID, namespace, limitrangeName)
	if err != nil {
		logger.Errorw("Failed to get limitrange",
			"cluster_id", clusterID,
			"namespace", namespace,
			"limitrange", limitrangeName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to get limitrange", err)
		return
	}

	response.Success(c, limitrange)
}

// DeleteLimitRange DELETE /api/k8s/limitrange
// 删除 LimitRange
func (h *K8sAPIHandler) DeleteLimitRange(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	limitrangeName := c.Query("name")

	logger.Infow("Deleting limitrange",
		"cluster_id", clusterID,
		"namespace", namespace,
		"limitrange", limitrangeName,
	)

	if err := h.limitrangeService.DeleteLimitRange(c.Request.Context(), clusterID, namespace, limitrangeName); err != nil {
		logger.Errorw("Failed to delete limitrange",
			"cluster_id", clusterID,
			"namespace", namespace,
			"limitrange", limitrangeName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete limitrange", err)
		return
	}

	response.SuccessWithMessage(c, "LimitRange deleted successfully", gin.H{
		"limitrange": limitrangeName,
	})
}

// ===========================
// ServiceAccount 管理接口
// ===========================

// ListServiceAccounts GET /api/k8s/serviceaccounts
// 获取 ServiceAccount 列表
func (h *K8sAPIHandler) ListServiceAccounts(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing serviceaccounts",
		"cluster_id", clusterID,
		"namespace", namespace,
		"page", params.Page,
	)

	serviceaccounts, total, err := h.serviceaccountService.ListServiceAccounts(c.Request.Context(), clusterID, namespace, params.GetOffset(), params.GetLimit())
	if err != nil {
		logger.Errorw("Failed to list serviceaccounts",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list serviceaccounts", err)
		return
	}

	resp := pagination.NewResponse(serviceaccounts, total, params)
	response.Success(c, resp)
}

// GetServiceAccount GET /api/k8s/serviceaccount
// 获取 ServiceAccount 详情
func (h *K8sAPIHandler) GetServiceAccount(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	serviceaccountName := c.Query("name")

	logger.Infow("Getting serviceaccount",
		"cluster_id", clusterID,
		"namespace", namespace,
		"serviceaccount", serviceaccountName,
	)

	serviceaccount, err := h.serviceaccountService.GetServiceAccount(c.Request.Context(), clusterID, namespace, serviceaccountName)
	if err != nil {
		logger.Errorw("Failed to get serviceaccount",
			"cluster_id", clusterID,
			"namespace", namespace,
			"serviceaccount", serviceaccountName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to get serviceaccount", err)
		return
	}

	response.Success(c, serviceaccount)
}

// DeleteServiceAccount DELETE /api/k8s/serviceaccount
// 删除 ServiceAccount
func (h *K8sAPIHandler) DeleteServiceAccount(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	serviceaccountName := c.Query("name")

	logger.Infow("Deleting serviceaccount",
		"cluster_id", clusterID,
		"namespace", namespace,
		"serviceaccount", serviceaccountName,
	)

	if err := h.serviceaccountService.DeleteServiceAccount(c.Request.Context(), clusterID, namespace, serviceaccountName); err != nil {
		logger.Errorw("Failed to delete serviceaccount",
			"cluster_id", clusterID,
			"namespace", namespace,
			"serviceaccount", serviceaccountName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete serviceaccount", err)
		return
	}

	response.SuccessWithMessage(c, "ServiceAccount deleted successfully", gin.H{
		"serviceaccount": serviceaccountName,
	})
}

// ===========================
// ClusterRoleBinding 管理接口
// ===========================

// ListClusterRoleBindings GET /api/k8s/clusterrolebindings
// 获取 ClusterRoleBinding 列表
func (h *K8sAPIHandler) ListClusterRoleBindings(c *gin.Context) {
	clusterID := c.Query("clusterId")
	params := pagination.Parse(c)

	logger.Infow("Listing clusterrolebindings",
		"cluster_id", clusterID,
		"page", params.Page,
	)

	clusterrolebindings, total, err := h.clusterrolebindingService.ListClusterRoleBindings(c.Request.Context(), clusterID, params.GetOffset(), params.GetLimit())
	if err != nil {
		logger.Errorw("Failed to list clusterrolebindings",
			"cluster_id", clusterID,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list clusterrolebindings", err)
		return
	}

	resp := pagination.NewResponse(clusterrolebindings, total, params)
	response.Success(c, resp)
}

// GetClusterRoleBinding GET /api/k8s/clusterrolebinding
// 获取 ClusterRoleBinding 详情
func (h *K8sAPIHandler) GetClusterRoleBinding(c *gin.Context) {
	clusterID := c.Query("clusterId")
	clusterrolebindingName := c.Query("name")

	logger.Infow("Getting clusterrolebinding",
		"cluster_id", clusterID,
		"clusterrolebinding", clusterrolebindingName,
	)

	clusterrolebinding, err := h.clusterrolebindingService.GetClusterRoleBinding(c.Request.Context(), clusterID, clusterrolebindingName)
	if err != nil {
		logger.Errorw("Failed to get clusterrolebinding",
			"cluster_id", clusterID,
			"clusterrolebinding", clusterrolebindingName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to get clusterrolebinding", err)
		return
	}

	response.Success(c, clusterrolebinding)
}

// DeleteClusterRoleBinding DELETE /api/k8s/clusterrolebinding
// 删除 ClusterRoleBinding
func (h *K8sAPIHandler) DeleteClusterRoleBinding(c *gin.Context) {
	clusterID := c.Query("clusterId")
	clusterrolebindingName := c.Query("name")

	logger.Infow("Deleting clusterrolebinding",
		"cluster_id", clusterID,
		"clusterrolebinding", clusterrolebindingName,
	)

	if err := h.clusterrolebindingService.DeleteClusterRoleBinding(c.Request.Context(), clusterID, clusterrolebindingName); err != nil {
		logger.Errorw("Failed to delete clusterrolebinding",
			"cluster_id", clusterID,
			"clusterrolebinding", clusterrolebindingName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete clusterrolebinding", err)
		return
	}

	response.SuccessWithMessage(c, "ClusterRoleBinding deleted successfully", gin.H{
		"clusterrolebinding": clusterrolebindingName,
	})
}

// ===========================
// ResourceQuota 管理接口
// ===========================

// ListResourceQuotas GET /api/k8s/resourcequotas
// 获取 ResourceQuota 列表
func (h *K8sAPIHandler) ListResourceQuotas(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing resourcequotas",
		"cluster_id", clusterID,
		"namespace", namespace,
		"page", params.Page,
	)

	resourcequotas, total, err := h.resourcequotaService.ListResourceQuotas(c.Request.Context(), clusterID, namespace, params.GetOffset(), params.GetLimit())
	if err != nil {
		logger.Errorw("Failed to list resourcequotas",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list resourcequotas", err)
		return
	}

	resp := pagination.NewResponse(resourcequotas, total, params)
	response.Success(c, resp)
}

// GetResourceQuota GET /api/k8s/resourcequota
// 获取 ResourceQuota 详情
func (h *K8sAPIHandler) GetResourceQuota(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	resourcequotaName := c.Query("name")

	logger.Infow("Getting resourcequota",
		"cluster_id", clusterID,
		"namespace", namespace,
		"resourcequota", resourcequotaName,
	)

	resourcequota, err := h.resourcequotaService.GetResourceQuota(c.Request.Context(), clusterID, namespace, resourcequotaName)
	if err != nil {
		logger.Errorw("Failed to get resourcequota",
			"cluster_id", clusterID,
			"namespace", namespace,
			"resourcequota", resourcequotaName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to get resourcequota", err)
		return
	}

	response.Success(c, resourcequota)
}

// DeleteResourceQuota DELETE /api/k8s/resourcequota
// 删除 ResourceQuota
func (h *K8sAPIHandler) DeleteResourceQuota(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	resourcequotaName := c.Query("name")

	logger.Infow("Deleting resourcequota",
		"cluster_id", clusterID,
		"namespace", namespace,
		"resourcequota", resourcequotaName,
	)

	if err := h.resourcequotaService.DeleteResourceQuota(c.Request.Context(), clusterID, namespace, resourcequotaName); err != nil {
		logger.Errorw("Failed to delete resourcequota",
			"cluster_id", clusterID,
			"namespace", namespace,
			"resourcequota", resourcequotaName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete resourcequota", err)
		return
	}

	response.SuccessWithMessage(c, "ResourceQuota deleted successfully", gin.H{
		"resourcequota": resourcequotaName,
	})
}

