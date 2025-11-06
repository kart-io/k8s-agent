package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/logger"
)

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
