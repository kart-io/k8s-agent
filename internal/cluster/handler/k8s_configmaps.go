package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/k8s-agent/common/validator"
	"github.com/kart-io/k8s-agent/internal/cluster/service"
	"github.com/kart-io/logger"
)

func (h *K8sAPIHandler) ListConfigMaps(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing configmaps",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	configmaps, total, err := h.configmapService.ListConfigMaps(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list configmaps",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list configmaps", err)
		return
	}

	resp := pagination.NewResponse(configmaps, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetConfigMap(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	configmapName := c.Query("name")

	logger.Infow("Getting configmap details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"configmap", configmapName,
	)

	configmap, err := h.configmapService.GetConfigMap(c.Request.Context(), clusterID, namespace, configmapName)
	if err != nil {
		logger.Errorw("Failed to get configmap",
			"cluster_id", clusterID,
			"namespace", namespace,
			"configmap", configmapName,
			"error", err.Error(),
		)
		response.NotFound(c, "ConfigMap not found", err)
		return
	}

	response.Success(c, configmap)
}

func (h *K8sAPIHandler) CreateConfigMap(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

	var req service.CreateConfigMapRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	// 确保 namespace 一致
	req.Namespace = namespace

	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid configmap name", err)
		return
	}

	logger.Infow("Creating configmap",
		"cluster_id", clusterID,
		"namespace", namespace,
		"configmap", req.Name,
	)

	configmap, err := h.configmapService.CreateConfigMap(c.Request.Context(), clusterID, &req)
	if err != nil {
		logger.Errorw("Failed to create configmap",
			"cluster_id", clusterID,
			"namespace", namespace,
			"configmap", req.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to create configmap", err)
		return
	}

	response.SuccessWithMessage(c, "ConfigMap created successfully", configmap)
}

func (h *K8sAPIHandler) UpdateConfigMap(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	configmapName := c.Query("name")

	var req service.CreateConfigMapRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	logger.Infow("Updating configmap",
		"cluster_id", clusterID,
		"namespace", namespace,
		"configmap", configmapName,
	)

	configmap, err := h.configmapService.UpdateConfigMap(c.Request.Context(), clusterID, namespace, configmapName, &req)
	if err != nil {
		logger.Errorw("Failed to update configmap",
			"cluster_id", clusterID,
			"namespace", namespace,
			"configmap", configmapName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to update configmap", err)
		return
	}

	response.SuccessWithMessage(c, "ConfigMap updated successfully", configmap)
}

func (h *K8sAPIHandler) DeleteConfigMap(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	configmapName := c.Query("name")

	logger.Infow("Deleting configmap",
		"cluster_id", clusterID,
		"namespace", namespace,
		"configmap", configmapName,
	)

	if err := h.configmapService.DeleteConfigMap(c.Request.Context(), clusterID, namespace, configmapName); err != nil {
		logger.Errorw("Failed to delete configmap",
			"cluster_id", clusterID,
			"namespace", namespace,
			"configmap", configmapName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete configmap", err)
		return
	}

	response.SuccessWithMessage(c, "ConfigMap deleted successfully", gin.H{
		"configmap": configmapName,
	})
}
