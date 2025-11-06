package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/logger"
)

func (h *K8sAPIHandler) ListLimitRanges(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing limitranges",
		"cluster_id", clusterID,
		"namespace", namespace,
		"page", params.Page,
	)

	limitranges, total, err := h.limitrangeService.ListLimitRanges(c.Request.Context(), clusterID, namespace, params.GetOffset(), params.GetLimit())
	if err != nil {
		logger.Errorw("Failed to list limitranges",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list limitranges", err)
		return
	}

	resp := pagination.NewResponse(limitranges, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetLimitRange(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	limitrangeName := c.Query("name")

	logger.Infow("Getting limitrange",
		"cluster_id", clusterID,
		"namespace", namespace,
		"limitrange", limitrangeName,
	)

	limitrange, err := h.limitrangeService.GetLimitRange(c.Request.Context(), clusterID, namespace, limitrangeName)
	if err != nil {
		logger.Errorw("Failed to get limitrange",
			"cluster_id", clusterID,
			"namespace", namespace,
			"limitrange", limitrangeName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to get limitrange", err)
		return
	}

	response.Success(c, limitrange)
}

func (h *K8sAPIHandler) DeleteLimitRange(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	limitrangeName := c.Query("name")

	logger.Infow("Deleting limitrange",
		"cluster_id", clusterID,
		"namespace", namespace,
		"limitrange", limitrangeName,
	)

	if err := h.limitrangeService.DeleteLimitRange(c.Request.Context(), clusterID, namespace, limitrangeName); err != nil {
		logger.Errorw("Failed to delete limitrange",
			"cluster_id", clusterID,
			"namespace", namespace,
			"limitrange", limitrangeName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete limitrange", err)
		return
	}

	response.SuccessWithMessage(c, "LimitRange deleted successfully", gin.H{
		"limitrange": limitrangeName,
	})
}

func (h *K8sAPIHandler) ListServiceAccounts(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing serviceaccounts",
		"cluster_id", clusterID,
		"namespace", namespace,
		"page", params.Page,
	)

	serviceaccounts, total, err := h.serviceaccountService.ListServiceAccounts(c.Request.Context(), clusterID, namespace, params.GetOffset(), params.GetLimit())
	if err != nil {
		logger.Errorw("Failed to list serviceaccounts",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list serviceaccounts", err)
		return
	}

	resp := pagination.NewResponse(serviceaccounts, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetServiceAccount(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	serviceaccountName := c.Query("name")

	logger.Infow("Getting serviceaccount",
		"cluster_id", clusterID,
		"namespace", namespace,
		"serviceaccount", serviceaccountName,
	)

	serviceaccount, err := h.serviceaccountService.GetServiceAccount(c.Request.Context(), clusterID, namespace, serviceaccountName)
	if err != nil {
		logger.Errorw("Failed to get serviceaccount",
			"cluster_id", clusterID,
			"namespace", namespace,
			"serviceaccount", serviceaccountName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to get serviceaccount", err)
		return
	}

	response.Success(c, serviceaccount)
}

func (h *K8sAPIHandler) DeleteServiceAccount(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	serviceaccountName := c.Query("name")

	logger.Infow("Deleting serviceaccount",
		"cluster_id", clusterID,
		"namespace", namespace,
		"serviceaccount", serviceaccountName,
	)

	if err := h.serviceaccountService.DeleteServiceAccount(c.Request.Context(), clusterID, namespace, serviceaccountName); err != nil {
		logger.Errorw("Failed to delete serviceaccount",
			"cluster_id", clusterID,
			"namespace", namespace,
			"serviceaccount", serviceaccountName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete serviceaccount", err)
		return
	}

	response.SuccessWithMessage(c, "ServiceAccount deleted successfully", gin.H{
		"serviceaccount": serviceaccountName,
	})
}

func (h *K8sAPIHandler) ListClusterRoleBindings(c *gin.Context) {
	clusterID := c.Query("clusterId")
	params := pagination.Parse(c)

	logger.Infow("Listing clusterrolebindings",
		"cluster_id", clusterID,
		"page", params.Page,
	)

	clusterrolebindings, total, err := h.clusterrolebindingService.ListClusterRoleBindings(c.Request.Context(), clusterID, params.GetOffset(), params.GetLimit())
	if err != nil {
		logger.Errorw("Failed to list clusterrolebindings",
			"cluster_id", clusterID,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list clusterrolebindings", err)
		return
	}

	resp := pagination.NewResponse(clusterrolebindings, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetClusterRoleBinding(c *gin.Context) {
	clusterID := c.Query("clusterId")
	clusterrolebindingName := c.Query("name")

	logger.Infow("Getting clusterrolebinding",
		"cluster_id", clusterID,
		"clusterrolebinding", clusterrolebindingName,
	)

	clusterrolebinding, err := h.clusterrolebindingService.GetClusterRoleBinding(c.Request.Context(), clusterID, clusterrolebindingName)
	if err != nil {
		logger.Errorw("Failed to get clusterrolebinding",
			"cluster_id", clusterID,
			"clusterrolebinding", clusterrolebindingName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to get clusterrolebinding", err)
		return
	}

	response.Success(c, clusterrolebinding)
}

func (h *K8sAPIHandler) DeleteClusterRoleBinding(c *gin.Context) {
	clusterID := c.Query("clusterId")
	clusterrolebindingName := c.Query("name")

	logger.Infow("Deleting clusterrolebinding",
		"cluster_id", clusterID,
		"clusterrolebinding", clusterrolebindingName,
	)

	if err := h.clusterrolebindingService.DeleteClusterRoleBinding(c.Request.Context(), clusterID, clusterrolebindingName); err != nil {
		logger.Errorw("Failed to delete clusterrolebinding",
			"cluster_id", clusterID,
			"clusterrolebinding", clusterrolebindingName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete clusterrolebinding", err)
		return
	}

	response.SuccessWithMessage(c, "ClusterRoleBinding deleted successfully", gin.H{
		"clusterrolebinding": clusterrolebindingName,
	})
}

func (h *K8sAPIHandler) ListResourceQuotas(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing resourcequotas",
		"cluster_id", clusterID,
		"namespace", namespace,
		"page", params.Page,
	)

	resourcequotas, total, err := h.resourcequotaService.ListResourceQuotas(c.Request.Context(), clusterID, namespace, params.GetOffset(), params.GetLimit())
	if err != nil {
		logger.Errorw("Failed to list resourcequotas",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list resourcequotas", err)
		return
	}

	resp := pagination.NewResponse(resourcequotas, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetResourceQuota(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	resourcequotaName := c.Query("name")

	logger.Infow("Getting resourcequota",
		"cluster_id", clusterID,
		"namespace", namespace,
		"resourcequota", resourcequotaName,
	)

	resourcequota, err := h.resourcequotaService.GetResourceQuota(c.Request.Context(), clusterID, namespace, resourcequotaName)
	if err != nil {
		logger.Errorw("Failed to get resourcequota",
			"cluster_id", clusterID,
			"namespace", namespace,
			"resourcequota", resourcequotaName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to get resourcequota", err)
		return
	}

	response.Success(c, resourcequota)
}

func (h *K8sAPIHandler) DeleteResourceQuota(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	resourcequotaName := c.Query("name")

	logger.Infow("Deleting resourcequota",
		"cluster_id", clusterID,
		"namespace", namespace,
		"resourcequota", resourcequotaName,
	)

	if err := h.resourcequotaService.DeleteResourceQuota(c.Request.Context(), clusterID, namespace, resourcequotaName); err != nil {
		logger.Errorw("Failed to delete resourcequota",
			"cluster_id", clusterID,
			"namespace", namespace,
			"resourcequota", resourcequotaName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete resourcequota", err)
		return
	}

	response.SuccessWithMessage(c, "ResourceQuota deleted successfully", gin.H{
		"resourcequota": resourcequotaName,
	})
}
