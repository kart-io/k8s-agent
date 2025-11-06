package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/logger"
)

func (h *K8sAPIHandler) ListRoleBindings(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing rolebindings",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	rolebindings, total, err := h.rolebindingService.ListRoleBindings(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list rolebindings",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list rolebindings", err)
		return
	}

	resp := pagination.NewResponse(rolebindings, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetRoleBinding(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	rbName := c.Query("name")

	logger.Infow("Getting rolebinding details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"rolebinding", rbName,
	)

	rb, err := h.rolebindingService.GetRoleBinding(c.Request.Context(), clusterID, namespace, rbName)
	if err != nil {
		logger.Errorw("Failed to get rolebinding",
			"cluster_id", clusterID,
			"namespace", namespace,
			"rolebinding", rbName,
			"error", err.Error(),
		)
		response.NotFound(c, "RoleBinding not found", err)
		return
	}

	response.Success(c, rb)
}

func (h *K8sAPIHandler) DeleteRoleBinding(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	rbName := c.Query("name")

	logger.Infow("Deleting rolebinding",
		"cluster_id", clusterID,
		"namespace", namespace,
		"rolebinding", rbName,
	)

	if err := h.rolebindingService.DeleteRoleBinding(c.Request.Context(), clusterID, namespace, rbName); err != nil {
		logger.Errorw("Failed to delete rolebinding",
			"cluster_id", clusterID,
			"namespace", namespace,
			"rolebinding", rbName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete rolebinding", err)
		return
	}

	response.SuccessWithMessage(c, "RoleBinding deleted successfully", gin.H{
		"rolebinding": rbName,
	})
}

func (h *K8sAPIHandler) ListClusterRoles(c *gin.Context) {
	clusterID := c.Query("clusterId")
	params := pagination.Parse(c)

	logger.Infow("Listing clusterroles",
		"cluster_id", clusterID,
	)

	clusterroles, total, err := h.clusterroleService.ListClusterRoles(
		c.Request.Context(),
		clusterID,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list clusterroles",
			"cluster_id", clusterID,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list clusterroles", err)
		return
	}

	resp := pagination.NewResponse(clusterroles, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetClusterRole(c *gin.Context) {
	clusterID := c.Query("clusterId")
	crName := c.Query("name")

	logger.Infow("Getting clusterrole details",
		"cluster_id", clusterID,
		"clusterrole", crName,
	)

	cr, err := h.clusterroleService.GetClusterRole(c.Request.Context(), clusterID, crName)
	if err != nil {
		logger.Errorw("Failed to get clusterrole",
			"cluster_id", clusterID,
			"clusterrole", crName,
			"error", err.Error(),
		)
		response.NotFound(c, "ClusterRole not found", err)
		return
	}

	response.Success(c, cr)
}

func (h *K8sAPIHandler) DeleteClusterRole(c *gin.Context) {
	clusterID := c.Query("clusterId")
	crName := c.Query("name")

	logger.Infow("Deleting clusterrole",
		"cluster_id", clusterID,
		"clusterrole", crName,
	)

	if err := h.clusterroleService.DeleteClusterRole(c.Request.Context(), clusterID, crName); err != nil {
		logger.Errorw("Failed to delete clusterrole",
			"cluster_id", clusterID,
			"clusterrole", crName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete clusterrole", err)
		return
	}

	response.SuccessWithMessage(c, "ClusterRole deleted successfully", gin.H{
		"clusterrole": crName,
	})
}

func (h *K8sAPIHandler) ListRoles(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing roles",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	roles, total, err := h.roleService.ListRoles(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list roles",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list roles", err)
		return
	}

	resp := pagination.NewResponse(roles, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetRole(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	roleName := c.Query("name")

	logger.Infow("Getting role details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"role", roleName,
	)

	role, err := h.roleService.GetRole(c.Request.Context(), clusterID, namespace, roleName)
	if err != nil {
		logger.Errorw("Failed to get role",
			"cluster_id", clusterID,
			"namespace", namespace,
			"role", roleName,
			"error", err.Error(),
		)
		response.NotFound(c, "Role not found", err)
		return
	}

	response.Success(c, role)
}

func (h *K8sAPIHandler) DeleteRole(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	roleName := c.Query("name")

	logger.Infow("Deleting role",
		"cluster_id", clusterID,
		"namespace", namespace,
		"role", roleName,
	)

	if err := h.roleService.DeleteRole(c.Request.Context(), clusterID, namespace, roleName); err != nil {
		logger.Errorw("Failed to delete role",
			"cluster_id", clusterID,
			"namespace", namespace,
			"role", roleName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete role", err)
		return
	}

	response.SuccessWithMessage(c, "Role deleted successfully", gin.H{
		"role": roleName,
	})
}
