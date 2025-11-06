package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/k8s-agent/common/validator"
	"github.com/kart-io/logger"
)

func (h *K8sAPIHandler) ListStorageClasses(c *gin.Context) {
	clusterID := c.Query("clusterId")
	params := pagination.Parse(c)

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Listing storageclasses",
		"cluster_id", clusterID,
		"page", params.Page,
	)

	storageClasses, total, err := h.storageclassService.ListStorageClasses(
		c.Request.Context(),
		clusterID,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list storageclasses", "cluster_id", clusterID, "error", err.Error())
		response.InternalError(c, "Failed to list storageclasses", err)
		return
	}

	resp := pagination.NewResponse(storageClasses, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetStorageClass(c *gin.Context) {
	clusterID := c.Query("clusterId")
	name := c.Query("name")

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	if err := validator.ValidateK8sName(name); err != nil {
		response.BadRequest(c, "Invalid storageclass name", err)
		return
	}

	logger.Infow("Getting storageclass details",
		"cluster_id", clusterID,
		"storageclass", name,
	)

	storageClass, err := h.storageclassService.GetStorageClass(c.Request.Context(), clusterID, name)
	if err != nil {
		logger.Errorw("Failed to get storageclass",
			"cluster_id", clusterID,
			"storageclass", name,
			"error", err.Error(),
		)
		response.NotFound(c, "StorageClass not found", err)
		return
	}

	response.Success(c, storageClass)
}

func (h *K8sAPIHandler) DeleteStorageClass(c *gin.Context) {
	clusterID := c.Query("clusterId")
	name := c.Query("name")

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	if err := validator.ValidateK8sName(name); err != nil {
		response.BadRequest(c, "Invalid storageclass name", err)
		return
	}

	logger.Infow("Deleting storageclass",
		"cluster_id", clusterID,
		"storageclass", name,
	)

	if err := h.storageclassService.DeleteStorageClass(c.Request.Context(), clusterID, name); err != nil {
		logger.Errorw("Failed to delete storageclass",
			"cluster_id", clusterID,
			"storageclass", name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete storageclass", err)
		return
	}

	response.SuccessWithMessage(c, "StorageClass deleted successfully", gin.H{
		"storageclass": name,
	})
}
