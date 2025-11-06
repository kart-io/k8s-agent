package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/logger"
)

func (h *K8sAPIHandler) ListEvents(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	// 支持按事件类型过滤
	eventType := c.Query("type")
	// 支持按事件原因过滤
	eventReason := c.Query("reason")

	logger.Infow("Listing events",
		"cluster_id", clusterID,
		"namespace", namespace,
		"type", eventType,
		"reason", eventReason,
	)

	events, total, err := h.eventService.ListEvents(
		c.Request.Context(),
		clusterID,
		namespace,
		eventType,
		eventReason,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list events",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list events", err)
		return
	}

	resp := pagination.NewResponse(events, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetEvent(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	eventName := c.Query("name")

	logger.Infow("Getting event details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"event", eventName,
	)

	event, err := h.eventService.GetEvent(c.Request.Context(), clusterID, namespace, eventName)
	if err != nil {
		logger.Errorw("Failed to get event",
			"cluster_id", clusterID,
			"namespace", namespace,
			"event", eventName,
			"error", err.Error(),
		)
		response.NotFound(c, "Event not found", err)
		return
	}

	response.Success(c, event)
}
