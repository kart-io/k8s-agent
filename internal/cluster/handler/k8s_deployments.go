package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/k8s-agent/common/validator"
	"github.com/kart-io/k8s-agent/internal/cluster/types"
	"github.com/kart-io/logger"
)

func (h *K8sAPIHandler) ListDeployments(c *gin.Context) {
	var req types.ListDeploymentsRequest
	params := pagination.Parse(c)

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err)
		return
	}

	if err := validator.ValidateClusterID(req.ClusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	// namespace 参数是可选的，如果提供则验证
	if req.Namespace != "" {
		if err := validator.ValidateK8sName(req.Namespace); err != nil {
			response.BadRequest(c, "Invalid namespace name", err)
			return
		}
	}

	logger.Infow("Listing deployments",
		"cluster_id", req.ClusterID,
		"namespace", req.Namespace,
	)

	deployments, total, err := h.deploymentService.ListDeployments(
		c.Request.Context(),
		req.ClusterID,
		req.Namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list deployments",
			"cluster_id", req.ClusterID,
			"namespace", req.Namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list deployments", err)
		return
	}

	resp := pagination.NewResponse(deployments, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetDeployment(c *gin.Context) {
	var req types.GetDeploymentRequest

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err)
		return
	}

	if err := validator.ValidateClusterID(req.ClusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	if err := validator.ValidateK8sName(req.Namespace); err != nil {
		response.BadRequest(c, "Invalid namespace name", err)
		return
	}

	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid deployment name", err)
		return
	}

	logger.Infow("Getting deployment details",
		"cluster_id", req.ClusterID,
		"namespace", req.Namespace,
		"deployment", req.Name,
	)

	deployment, err := h.deploymentService.GetDeployment(
		c.Request.Context(),
		req.ClusterID,
		req.Namespace,
		req.Name,
	)
	if err != nil {
		logger.Errorw("Failed to get deployment",
			"cluster_id", req.ClusterID,
			"namespace", req.Namespace,
			"deployment", req.Name,
			"error", err.Error(),
		)
		response.NotFound(c, "Deployment not found", err)
		return
	}

	response.Success(c, deployment)
}

func (h *K8sAPIHandler) ScaleDeployment(c *gin.Context) {
	var req types.ScaleDeploymentRequest

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid path parameters", err)
		return
	}

	// 绑定请求体参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	if err := validator.ValidateClusterID(req.ClusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	if err := validator.ValidateK8sName(req.Namespace); err != nil {
		response.BadRequest(c, "Invalid namespace name", err)
		return
	}

	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid deployment name", err)
		return
	}

	if err := validator.ValidateReplicas(req.Replicas); err != nil {
		response.BadRequest(c, "Invalid replicas count", err)
		return
	}

	logger.Infow("Scaling deployment",
		"cluster_id", req.ClusterID,
		"namespace", req.Namespace,
		"deployment", req.Name,
		"replicas", req.Replicas,
	)

	deployment, err := h.deploymentService.ScaleDeployment(
		c.Request.Context(),
		req.ClusterID,
		req.Namespace,
		req.Name,
		req.Replicas,
	)
	if err != nil {
		logger.Errorw("Failed to scale deployment",
			"cluster_id", req.ClusterID,
			"namespace", req.Namespace,
			"deployment", req.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to scale deployment", err)
		return
	}

	response.SuccessWithMessage(c, "Deployment scaled successfully", deployment)
}

func (h *K8sAPIHandler) RestartDeployment(c *gin.Context) {
	var req types.RestartDeploymentRequest

	// 绑定路径参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err)
		return
	}

	if err := validator.ValidateClusterID(req.ClusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	if err := validator.ValidateK8sName(req.Namespace); err != nil {
		response.BadRequest(c, "Invalid namespace name", err)
		return
	}

	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid deployment name", err)
		return
	}

	logger.Infow("Restarting deployment",
		"cluster_id", req.ClusterID,
		"namespace", req.Namespace,
		"deployment", req.Name,
	)

	if err := h.deploymentService.RestartDeployment(
		c.Request.Context(),
		req.ClusterID,
		req.Namespace,
		req.Name,
	); err != nil {
		logger.Errorw("Failed to restart deployment",
			"cluster_id", req.ClusterID,
			"namespace", req.Namespace,
			"deployment", req.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to restart deployment", err)
		return
	}

	response.SuccessWithMessage(c, "Deployment restarted successfully", gin.H{
		"deployment": req.Name,
	})
}
