package handler

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/internal/gateway/types"
)

var (
	totalRequests   int64
	successRequests int64
	failedRequests  int64
	totalLatency    int64
	startTime       = time.Now()
)

// MetricsHandler 监控指标处理器
func MetricsHandler(c *gin.Context) {
	total := atomic.LoadInt64(&totalRequests)
	success := atomic.LoadInt64(&successRequests)
	failed := atomic.LoadInt64(&failedRequests)
	latency := atomic.LoadInt64(&totalLatency)

	avgLatency := float64(0)
	if total > 0 {
		avgLatency = float64(latency) / float64(total) / 1000000 // 转换为毫秒
	}

	metrics := types.MetricsData{
		TotalRequests:   total,
		SuccessRequests: success,
		FailedRequests:  failed,
		AvgLatency:      avgLatency,
		ServiceMetrics:  make(map[string]int64),
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics":    metrics,
		"uptime":     time.Since(startTime).Seconds(),
		"timestamp":  time.Now().Unix(),
	})
}

// RecordRequest 记录请求
func RecordRequest(success bool, latency time.Duration) {
	atomic.AddInt64(&totalRequests, 1)
	atomic.AddInt64(&totalLatency, int64(latency))

	if success {
		atomic.AddInt64(&successRequests, 1)
	} else {
		atomic.AddInt64(&failedRequests, 1)
	}
}
