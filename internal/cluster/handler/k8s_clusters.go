package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/k8s-agent/common/validator"
	"github.com/kart-io/k8s-agent/internal/cluster/service"
	"github.com/kart-io/k8s-agent/internal/cluster/types"
	"github.com/kart-io/logger"
)

// ===========================
// 集群管理接口
// ===========================

// ListClusters GET /api/k8s/clusters
// 获取集群列表（支持分页）.
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
		false, // withStats: 默认不获取统计信息以提升性能
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
// 获取集群详情.
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

	cluster, err := h.clusterService.GetCluster(c.Request.Context(), req.ClusterID, true) // withStats: true 获取详细统计
	if err != nil {
		logger.Errorw("Failed to get cluster", "cluster_id", req.ClusterID, "error", err.Error())
		response.NotFound(c, "Cluster not found", err)
		return
	}

	response.Success(c, cluster)
}

// GetClusterOptions GET /api/k8s/clusters/options
// 获取集群选择器列表（用于下拉选择）.
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
// 创建新集群.
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

	// 转换为 service.CreateClusterRequest
	createReq := &service.CreateClusterRequest{
		Name:        req.Name,
		Description: req.Description,
		Endpoint:    req.Endpoint,
		KubeConfig:  req.KubeConfig,
		Region:      req.Region,
		Provider:    req.Provider,
	}

	cluster, err := h.clusterService.CreateCluster(c.Request.Context(), createReq)
	if err != nil {
		logger.Errorw("Failed to create cluster", "name", req.Name, "error", err.Error())
		response.InternalError(c, "Failed to create cluster", err)
		return
	}

	response.SuccessWithMessage(c, "Cluster created successfully", cluster)
}

// UpdateCluster PUT /api/k8s/clusters/:id
// 更新集群信息.
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

	// 转换为 service.UpdateClusterRequest
	updateReq := &service.UpdateClusterRequest{
		Name:        req.Name,
		Description: req.Description,
	}

	cluster, err := h.clusterService.UpdateCluster(
		c.Request.Context(),
		req.ClusterID,
		updateReq,
	)
	if err != nil {
		logger.Errorw("Failed to update cluster", "cluster_id", req.ClusterID, "error", err.Error())
		response.InternalError(c, "Failed to update cluster", err)
		return
	}

	response.SuccessWithMessage(c, "Cluster updated successfully", cluster)
}

// DeleteCluster DELETE /api/k8s/clusters/:id
// 删除集群.
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

// GetClusterHealthStatus GET /api/k8s/clusters/:id/health
// 获取集群健康状态.
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
