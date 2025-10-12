package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/cluster-service/internal/service"
	"github.com/kart-io/k8s-agent/cluster-service/pkg/types"
	"github.com/sirupsen/logrus"
)

type ClusterHandler struct {
	service *service.ClusterService
	log     *logrus.Logger
}

func NewClusterHandler(service *service.ClusterService, logger *logrus.Logger) *ClusterHandler {
	return &ClusterHandler{
		service: service,
		log:     logger,
	}
}

// AddCluster handles POST /api/v1/clusters
func (h *ClusterHandler) AddCluster(c *gin.Context) {
	var cluster types.Cluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		h.log.WithError(err).Error("Failed to bind cluster request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.service.AddCluster(c.Request.Context(), &cluster); err != nil {
		h.log.WithError(err).Error("Failed to add cluster")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add cluster"})
		return
	}

	c.JSON(http.StatusCreated, cluster)
}

// GetClusterHealth handles GET /api/v1/clusters/:id/health
func (h *ClusterHandler) GetClusterHealth(c *gin.Context) {
	clusterID := c.Param("id")

	health, err := h.service.GetClusterHealth(c.Request.Context(), clusterID)
	if err != nil {
		h.log.WithError(err).WithField("cluster_id", clusterID).Error("Failed to get cluster health")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cluster health"})
		return
	}

	c.JSON(http.StatusOK, health)
}

// GetPods handles GET /api/v1/clusters/:cluster_id/namespaces/:namespace/pods
func (h *ClusterHandler) GetPods(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	namespace := c.Param("namespace")

	if namespace == "" {
		namespace = "default"
	}

	pods, err := h.service.GetPods(c.Request.Context(), clusterID, namespace)
	if err != nil {
		h.log.WithError(err).WithFields(logrus.Fields{
			"cluster_id": clusterID,
			"namespace":  namespace,
		}).Error("Failed to get pods")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pods"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": pods,
		"total": len(pods),
	})
}

// HealthCheck handles GET /health
func (h *ClusterHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"service": "cluster-service",
	})
}
