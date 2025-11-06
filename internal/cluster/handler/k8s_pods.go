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
// Pod 管理接口
// ===========================

// ListPods GET /api/k8s/pods
// 获取 Pod 列表
// 支持查询所有命名空间或指定命名空间的 Pods.
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
// 获取 Pod 详情.
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
// 删除 Pod.
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
// 获取 Pod 日志.
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
