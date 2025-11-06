package handler

import (
	"github.com/gin-gonic/gin"
	networkingv1 "k8s.io/api/networking/v1"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/logger"
)

func (h *K8sAPIHandler) ListIngresses(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing ingresses",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	ingresses, total, err := h.ingressService.ListIngresses(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list ingresses",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list ingresses", err)
		return
	}

	resp := pagination.NewResponse(ingresses, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetIngress(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	ingressName := c.Query("name")

	logger.Infow("Getting ingress details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"ingress", ingressName,
	)

	ingress, err := h.ingressService.GetIngress(c.Request.Context(), clusterID, namespace, ingressName)
	if err != nil {
		logger.Errorw("Failed to get ingress",
			"cluster_id", clusterID,
			"namespace", namespace,
			"ingress", ingressName,
			"error", err.Error(),
		)
		response.NotFound(c, "Ingress not found", err)
		return
	}

	response.Success(c, ingress)
}

func (h *K8sAPIHandler) CreateIngress(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

	var ingress networkingv1.Ingress
	if err := c.ShouldBindJSON(&ingress); err != nil {
		response.BadRequest(c, "Invalid ingress data", err)
		return
	}

	logger.Infow("Creating ingress",
		"cluster_id", clusterID,
		"namespace", namespace,
		"ingress", ingress.Name,
	)

	createdIngress, err := h.ingressService.CreateIngress(c.Request.Context(), clusterID, namespace, &ingress)
	if err != nil {
		logger.Errorw("Failed to create ingress",
			"cluster_id", clusterID,
			"namespace", namespace,
			"ingress", ingress.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to create ingress", err)
		return
	}

	response.SuccessWithMessage(c, "Ingress created successfully", createdIngress)
}

func (h *K8sAPIHandler) UpdateIngress(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

	var ingress networkingv1.Ingress
	if err := c.ShouldBindJSON(&ingress); err != nil {
		response.BadRequest(c, "Invalid ingress data", err)
		return
	}

	logger.Infow("Updating ingress",
		"cluster_id", clusterID,
		"namespace", namespace,
		"ingress", ingress.Name,
	)

	updatedIngress, err := h.ingressService.UpdateIngress(c.Request.Context(), clusterID, namespace, &ingress)
	if err != nil {
		logger.Errorw("Failed to update ingress",
			"cluster_id", clusterID,
			"namespace", namespace,
			"ingress", ingress.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to update ingress", err)
		return
	}

	response.SuccessWithMessage(c, "Ingress updated successfully", updatedIngress)
}

func (h *K8sAPIHandler) DeleteIngress(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	ingressName := c.Query("name")

	logger.Infow("Deleting ingress",
		"cluster_id", clusterID,
		"namespace", namespace,
		"ingress", ingressName,
	)

	if err := h.ingressService.DeleteIngress(c.Request.Context(), clusterID, namespace, ingressName); err != nil {
		logger.Errorw("Failed to delete ingress",
			"cluster_id", clusterID,
			"namespace", namespace,
			"ingress", ingressName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete ingress", err)
		return
	}

	response.SuccessWithMessage(c, "Ingress deleted successfully", gin.H{
		"ingress": ingressName,
	})
}

func (h *K8sAPIHandler) ListNetworkPolicies(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing networkpolicies",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	networkpolicies, total, err := h.networkpolicyService.ListNetworkPolicies(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list networkpolicies",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list networkpolicies", err)
		return
	}

	resp := pagination.NewResponse(networkpolicies, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetNetworkPolicy(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	networkpolicyName := c.Query("name")

	logger.Infow("Getting networkpolicy details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"networkpolicy", networkpolicyName,
	)

	networkpolicy, err := h.networkpolicyService.GetNetworkPolicy(c.Request.Context(), clusterID, namespace, networkpolicyName)
	if err != nil {
		logger.Errorw("Failed to get networkpolicy",
			"cluster_id", clusterID,
			"namespace", namespace,
			"networkpolicy", networkpolicyName,
			"error", err.Error(),
		)
		response.NotFound(c, "NetworkPolicy not found", err)
		return
	}

	response.Success(c, networkpolicy)
}

func (h *K8sAPIHandler) CreateNetworkPolicy(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

	var networkpolicy networkingv1.NetworkPolicy
	if err := c.ShouldBindJSON(&networkpolicy); err != nil {
		response.BadRequest(c, "Invalid networkpolicy data", err)
		return
	}

	logger.Infow("Creating networkpolicy",
		"cluster_id", clusterID,
		"namespace", namespace,
		"networkpolicy", networkpolicy.Name,
	)

	createdNetworkPolicy, err := h.networkpolicyService.CreateNetworkPolicy(c.Request.Context(), clusterID, namespace, &networkpolicy)
	if err != nil {
		logger.Errorw("Failed to create networkpolicy",
			"cluster_id", clusterID,
			"namespace", namespace,
			"networkpolicy", networkpolicy.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to create networkpolicy", err)
		return
	}

	response.SuccessWithMessage(c, "NetworkPolicy created successfully", createdNetworkPolicy)
}

func (h *K8sAPIHandler) UpdateNetworkPolicy(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

	var networkpolicy networkingv1.NetworkPolicy
	if err := c.ShouldBindJSON(&networkpolicy); err != nil {
		response.BadRequest(c, "Invalid networkpolicy data", err)
		return
	}

	logger.Infow("Updating networkpolicy",
		"cluster_id", clusterID,
		"namespace", namespace,
		"networkpolicy", networkpolicy.Name,
	)

	updatedNetworkPolicy, err := h.networkpolicyService.UpdateNetworkPolicy(c.Request.Context(), clusterID, namespace, &networkpolicy)
	if err != nil {
		logger.Errorw("Failed to update networkpolicy",
			"cluster_id", clusterID,
			"namespace", namespace,
			"networkpolicy", networkpolicy.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to update networkpolicy", err)
		return
	}

	response.SuccessWithMessage(c, "NetworkPolicy updated successfully", updatedNetworkPolicy)
}

func (h *K8sAPIHandler) DeleteNetworkPolicy(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	networkpolicyName := c.Query("name")

	logger.Infow("Deleting networkpolicy",
		"cluster_id", clusterID,
		"namespace", namespace,
		"networkpolicy", networkpolicyName,
	)

	if err := h.networkpolicyService.DeleteNetworkPolicy(c.Request.Context(), clusterID, namespace, networkpolicyName); err != nil {
		logger.Errorw("Failed to delete networkpolicy",
			"cluster_id", clusterID,
			"namespace", namespace,
			"networkpolicy", networkpolicyName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete networkpolicy", err)
		return
	}

	response.SuccessWithMessage(c, "NetworkPolicy deleted successfully", gin.H{
		"networkpolicy": networkpolicyName,
	})
}
