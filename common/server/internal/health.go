// Package internal provides shared internal utilities for server implementations.
package internal

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthChecker 健康检查接口
type HealthChecker interface {
	Health(ctx context.Context) error
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Checkers map[string]HealthChecker
}

// RegisterHealthEndpoints 注册健康检查端点
func RegisterHealthEndpoints(engine *gin.Engine, checkers map[string]HealthChecker) {
	engine.GET("/health", healthHandler(checkers))
	engine.GET("/healthz", healthHandler(checkers)) // k8s 风格
	engine.GET("/readiness", readinessHandler(checkers))
	engine.GET("/liveness", livenessHandler())
}

// healthHandler 健康检查处理器 (简单存活检查)
func healthHandler(checkers map[string]HealthChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}

// livenessHandler 存活检查 (不检查依赖)
func livenessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "alive",
		})
	}
}

// readinessHandler 就绪检查处理器 (检查所有依赖)
func readinessHandler(checkers map[string]HealthChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		results := make(map[string]interface{})
		allHealthy := true

		for name, checker := range checkers {
			if err := checker.Health(ctx); err != nil {
				results[name] = map[string]interface{}{
					"status": "unhealthy",
					"error":  err.Error(),
				}
				allHealthy = false
			} else {
				results[name] = map[string]interface{}{
					"status": "healthy",
				}
			}
		}

		statusCode := http.StatusOK
		status := "ready"
		if !allHealthy {
			statusCode = http.StatusServiceUnavailable
			status = "not ready"
		}

		c.JSON(statusCode, gin.H{
			"status": status,
			"checks": results,
		})
	}
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status string                 `json:"status"`
	Checks map[string]CheckResult `json:"checks,omitempty"`
}

// CheckResult 单个检查结果
type CheckResult struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
