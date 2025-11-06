package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/k8s-agent/common/validator"
	"github.com/kart-io/logger"
)

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
