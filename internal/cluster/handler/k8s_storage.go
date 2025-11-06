package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/logger"
)

func (h *K8sAPIHandler) ListPVCs(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing pvcs",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	pvcs, total, err := h.pvcService.ListPVCs(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list pvcs",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list pvcs", err)
		return
	}

	resp := pagination.NewResponse(pvcs, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetPVC(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	pvcName := c.Query("name")

	logger.Infow("Getting pvc details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"pvc", pvcName,
	)

	pvc, err := h.pvcService.GetPVC(c.Request.Context(), clusterID, namespace, pvcName)
	if err != nil {
		logger.Errorw("Failed to get pvc",
			"cluster_id", clusterID,
			"namespace", namespace,
			"pvc", pvcName,
			"error", err.Error(),
		)
		response.NotFound(c, "PVC not found", err)
		return
	}

	response.Success(c, pvc)
}

func (h *K8sAPIHandler) DeletePVC(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	pvcName := c.Query("name")

	logger.Infow("Deleting pvc",
		"cluster_id", clusterID,
		"namespace", namespace,
		"pvc", pvcName,
	)

	if err := h.pvcService.DeletePVC(c.Request.Context(), clusterID, namespace, pvcName); err != nil {
		logger.Errorw("Failed to delete pvc",
			"cluster_id", clusterID,
			"namespace", namespace,
			"pvc", pvcName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete pvc", err)
		return
	}

	response.SuccessWithMessage(c, "PVC deleted successfully", gin.H{
		"pvc": pvcName,
	})
}

func (h *K8sAPIHandler) ListPVs(c *gin.Context) {
	clusterID := c.Query("clusterId")
	params := pagination.Parse(c)

	logger.Infow("Listing pvs",
		"cluster_id", clusterID,
	)

	pvs, total, err := h.pvService.ListPVs(
		c.Request.Context(),
		clusterID,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list pvs",
			"cluster_id", clusterID,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list pvs", err)
		return
	}

	resp := pagination.NewResponse(pvs, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetPV(c *gin.Context) {
	clusterID := c.Query("clusterId")
	pvName := c.Query("name")

	logger.Infow("Getting pv details",
		"cluster_id", clusterID,
		"pv", pvName,
	)

	pv, err := h.pvService.GetPV(c.Request.Context(), clusterID, pvName)
	if err != nil {
		logger.Errorw("Failed to get pv",
			"cluster_id", clusterID,
			"pv", pvName,
			"error", err.Error(),
		)
		response.NotFound(c, "PV not found", err)
		return
	}

	response.Success(c, pv)
}

func (h *K8sAPIHandler) DeletePV(c *gin.Context) {
	clusterID := c.Query("clusterId")
	pvName := c.Query("name")

	logger.Infow("Deleting pv",
		"cluster_id", clusterID,
		"pv", pvName,
	)

	if err := h.pvService.DeletePV(c.Request.Context(), clusterID, pvName); err != nil {
		logger.Errorw("Failed to delete pv",
			"cluster_id", clusterID,
			"pv", pvName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete pv", err)
		return
	}

	response.SuccessWithMessage(c, "PV deleted successfully", gin.H{
		"pv": pvName,
	})
}
