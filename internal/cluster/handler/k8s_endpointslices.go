package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/logger"
)

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
