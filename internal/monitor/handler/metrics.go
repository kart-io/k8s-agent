package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/internal/monitor/service"
	"github.com/kart-io/logger/core"
)

type MetricsHandler struct {
	monitorSvc *service.MonitorService
	log        core.Logger
}

func NewMetricsHandler(monitorSvc *service.MonitorService, logger core.Logger) *MetricsHandler {
	return &MetricsHandler{
		monitorSvc: monitorSvc,
		log:        logger,
	}
}

// GetSummary 获取监控概览.
func (h *MetricsHandler) GetSummary(c *gin.Context) {
	summary, err := h.monitorSvc.GetMetricsSummary(c.Request.Context())
	if err != nil {
		h.log.Errorw("Failed to get metrics summary", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get metrics summary",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summary,
	})
}

// GetAgentMetrics 获取 Agent 指标.
func (h *MetricsHandler) GetAgentMetrics(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	metrics, err := h.monitorSvc.GetAgentMetrics(c.Request.Context(), limit, offset)
	if err != nil {
		h.log.Errorw("Failed to get agent metrics", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get agent metrics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// GetTrends 获取趋势数据.
func (h *MetricsHandler) GetTrends(c *gin.Context) {
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))

	trends, err := h.monitorSvc.GetTrendData(c.Request.Context(), hours)
	if err != nil {
		h.log.Errorw("Failed to get trend data", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get trend data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    trends,
	})
}
