package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/logger"
)

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
