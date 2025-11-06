package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/k8s-agent/common/validator"
	"github.com/kart-io/k8s-agent/internal/cluster/service"
	"github.com/kart-io/logger"
)

func (h *K8sAPIHandler) ListServices(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing services",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	services, total, err := h.serviceService.ListServices(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list services",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list services", err)
		return
	}

	resp := pagination.NewResponse(services, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetService(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	serviceName := c.Query("name")

	logger.Infow("Getting service details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"service", serviceName,
	)

	service, err := h.serviceService.GetService(c.Request.Context(), clusterID, namespace, serviceName)
	if err != nil {
		logger.Errorw("Failed to get service",
			"cluster_id", clusterID,
			"namespace", namespace,
			"service", serviceName,
			"error", err.Error(),
		)
		response.NotFound(c, "Service not found", err)
		return
	}

	response.Success(c, service)
}

func (h *K8sAPIHandler) CreateService(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

	var req service.CreateServiceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	// 确保 namespace 一致
	req.Namespace = namespace

	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid service name", err)
		return
	}

	logger.Infow("Creating service",
		"cluster_id", clusterID,
		"namespace", namespace,
		"service", req.Name,
	)

	svc, err := h.serviceService.CreateService(c.Request.Context(), clusterID, &req)
	if err != nil {
		logger.Errorw("Failed to create service",
			"cluster_id", clusterID,
			"namespace", namespace,
			"service", req.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to create service", err)
		return
	}

	response.SuccessWithMessage(c, "Service created successfully", svc)
}

func (h *K8sAPIHandler) UpdateService(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	serviceName := c.Query("name")

	var req service.CreateServiceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	logger.Infow("Updating service",
		"cluster_id", clusterID,
		"namespace", namespace,
		"service", serviceName,
	)

	svc, err := h.serviceService.UpdateService(c.Request.Context(), clusterID, namespace, serviceName, &req)
	if err != nil {
		logger.Errorw("Failed to update service",
			"cluster_id", clusterID,
			"namespace", namespace,
			"service", serviceName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to update service", err)
		return
	}

	response.SuccessWithMessage(c, "Service updated successfully", svc)
}

func (h *K8sAPIHandler) DeleteService(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	serviceName := c.Query("name")

	logger.Infow("Deleting service",
		"cluster_id", clusterID,
		"namespace", namespace,
		"service", serviceName,
	)

	if err := h.serviceService.DeleteService(c.Request.Context(), clusterID, namespace, serviceName); err != nil {
		logger.Errorw("Failed to delete service",
			"cluster_id", clusterID,
			"namespace", namespace,
			"service", serviceName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete service", err)
		return
	}

	response.SuccessWithMessage(c, "Service deleted successfully", gin.H{
		"service": serviceName,
	})
}
