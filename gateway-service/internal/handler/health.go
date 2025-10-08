package handler

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/gateway-service/internal/proxy"
	"github.com/kart-io/k8s-agent/gateway-service/pkg/types"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	proxy  *proxy.Proxy
	logger *zap.Logger
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(proxy *proxy.Proxy, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		proxy:  proxy,
		logger: logger,
	}
}

// GatewayHealth 网关健康检查
func (h *HealthHandler) GatewayHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "gateway",
		"timestamp": time.Now().Unix(),
	})
}

// ServicesHealth 所有服务健康检查
func (h *HealthHandler) ServicesHealth(c *gin.Context) {
	var services map[string]types.ServiceConfig
	if err := viper.UnmarshalKey("services", &services); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "加载服务配置失败",
		})
		return
	}

	// 并发检查所有服务
	var wg sync.WaitGroup
	healthStatuses := make([]types.HealthStatus, 0, len(services))
	statusChan := make(chan types.HealthStatus, len(services))

	for name := range services {
		wg.Add(1)
		go func(serviceName string) {
			defer wg.Done()
			status, err := h.proxy.GetServiceHealth(serviceName)
			if err != nil {
				h.logger.Error("Failed to get service health",
					zap.String("service", serviceName),
					zap.Error(err),
				)
				statusChan <- types.HealthStatus{
					Service:   serviceName,
					Status:    "unknown",
					Timestamp: time.Now(),
					Message:   err.Error(),
				}
				return
			}
			statusChan <- *status
		}(name)
	}

	// 等待所有检查完成
	go func() {
		wg.Wait()
		close(statusChan)
	}()

	// 收集结果
	for status := range statusChan {
		healthStatuses = append(healthStatuses, status)
	}

	c.JSON(http.StatusOK, gin.H{
		"services":  healthStatuses,
		"timestamp": time.Now().Unix(),
	})
}

// ServiceHealth 单个服务健康检查
func (h *HealthHandler) ServiceHealth(c *gin.Context) {
	serviceName := c.Param("service")

	status, err := h.proxy.GetServiceHealth(serviceName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}
