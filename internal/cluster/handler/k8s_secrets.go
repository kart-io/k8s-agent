package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/k8s-agent/common/validator"
	"github.com/kart-io/k8s-agent/internal/cluster/service"
	"github.com/kart-io/logger"
)

func (h *K8sAPIHandler) ListSecrets(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing secrets",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	secrets, total, err := h.secretService.ListSecrets(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list secrets",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list secrets", err)
		return
	}

	resp := pagination.NewResponse(secrets, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetSecret(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	secretName := c.Query("name")

	// 查询参数：是否包含敏感数据
	includeData := c.Query("includeData") == "true"

	logger.Infow("Getting secret details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"secret", secretName,
		"include_data", includeData,
	)

	secret, err := h.secretService.GetSecret(c.Request.Context(), clusterID, namespace, secretName, includeData)
	if err != nil {
		logger.Errorw("Failed to get secret",
			"cluster_id", clusterID,
			"namespace", namespace,
			"secret", secretName,
			"error", err.Error(),
		)
		response.NotFound(c, "Secret not found", err)
		return
	}

	response.Success(c, secret)
}

func (h *K8sAPIHandler) CreateSecret(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

	var req service.CreateSecretRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	// 确保 namespace 一致
	req.Namespace = namespace

	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid secret name", err)
		return
	}

	logger.Infow("Creating secret",
		"cluster_id", clusterID,
		"namespace", namespace,
		"secret", req.Name,
	)

	secret, err := h.secretService.CreateSecret(c.Request.Context(), clusterID, &req)
	if err != nil {
		logger.Errorw("Failed to create secret",
			"cluster_id", clusterID,
			"namespace", namespace,
			"secret", req.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to create secret", err)
		return
	}

	response.SuccessWithMessage(c, "Secret created successfully", secret)
}

func (h *K8sAPIHandler) UpdateSecret(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	secretName := c.Query("name")

	var req service.CreateSecretRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	logger.Infow("Updating secret",
		"cluster_id", clusterID,
		"namespace", namespace,
		"secret", secretName,
	)

	secret, err := h.secretService.UpdateSecret(c.Request.Context(), clusterID, namespace, secretName, &req)
	if err != nil {
		logger.Errorw("Failed to update secret",
			"cluster_id", clusterID,
			"namespace", namespace,
			"secret", secretName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to update secret", err)
		return
	}

	response.SuccessWithMessage(c, "Secret updated successfully", secret)
}

func (h *K8sAPIHandler) DeleteSecret(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	secretName := c.Query("name")

	logger.Infow("Deleting secret",
		"cluster_id", clusterID,
		"namespace", namespace,
		"secret", secretName,
	)

	if err := h.secretService.DeleteSecret(c.Request.Context(), clusterID, namespace, secretName); err != nil {
		logger.Errorw("Failed to delete secret",
			"cluster_id", clusterID,
			"namespace", namespace,
			"secret", secretName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete secret", err)
		return
	}

	response.SuccessWithMessage(c, "Secret deleted successfully", gin.H{
		"secret": secretName,
	})
}
