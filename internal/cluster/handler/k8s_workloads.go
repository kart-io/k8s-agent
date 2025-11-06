package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/k8s-agent/common/validator"
	"github.com/kart-io/logger"
)

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
