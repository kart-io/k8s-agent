package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/internal/cluster/service"
	"github.com/kart-io/logger/core"
)

type ClusterHandler struct {
	service *service.ClusterService
	log     core.Logger
}

func NewClusterHandler(service *service.ClusterService, logger core.Logger) *ClusterHandler {
	return &ClusterHandler{
		service: service,
		log:     logger,
	}
}

// AddCluster handles POST /api/v1/clusters.
func (h *ClusterHandler) AddCluster(c *gin.Context) {
	var req service.CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorw("Failed to bind cluster request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	cluster, err := h.service.CreateCluster(c.Request.Context(), &req)
	if err != nil {
		h.log.Errorw("Failed to add cluster", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add cluster"})
		return
	}

	c.JSON(http.StatusCreated, cluster)
}

// GetClusterHealth handles GET /api/v1/clusters/:id/health.
func (h *ClusterHandler) GetClusterHealth(c *gin.Context) {
	clusterID := c.Param("id")

	health, err := h.service.GetClusterHealth(c.Request.Context(), clusterID)
	if err != nil {
		h.log.Errorw("Failed to get cluster health",
			"error", err,
			"cluster_id", clusterID,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cluster health"})
		return
	}

	c.JSON(http.StatusOK, health)
}

// GetPods handles GET /api/v1/clusters/:cluster_id/namespaces/:namespace/pods.
func (h *ClusterHandler) GetPods(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	namespace := c.Param("namespace")

	if namespace == "" {
		namespace = "default"
	}

	pods, err := h.service.GetPods(c.Request.Context(), clusterID, namespace)
	if err != nil {
		h.log.Errorw("Failed to get pods",
			"error", err,
			"cluster_id", clusterID,
			"namespace", namespace,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pods"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": pods,
		"total": len(pods),
	})
}

// HealthCheck handles GET /health.
func (h *ClusterHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "cluster-service",
	})
}
