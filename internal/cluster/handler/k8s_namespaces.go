package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/k8s-agent/common/validator"
	"github.com/kart-io/k8s-agent/internal/cluster/types"
	"github.com/kart-io/logger"
)

// ===========================
// 命名空间管理接口
// ===========================

// ListNamespaces GET /api/k8s/clusters/:clusterId/namespaces
// 获取命名空间列表.
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
// 获取命名空间详情.
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
// 创建命名空间.
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
// 删除命名空间.
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
