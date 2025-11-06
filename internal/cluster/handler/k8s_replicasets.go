package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/k8s-agent/common/validator"
	"github.com/kart-io/logger"
)

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
