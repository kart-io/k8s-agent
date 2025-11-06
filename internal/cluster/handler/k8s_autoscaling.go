package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/logger"
)

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
