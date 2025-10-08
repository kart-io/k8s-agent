package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/gateway-service/pkg/types"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Proxy 代理处理器
type Proxy struct {
	logger   *zap.Logger
	client   *http.Client
	services map[string]types.ServiceConfig
}

// NewProxy 创建代理处理器
func NewProxy(logger *zap.Logger) *Proxy {
	// 加载服务配置
	var services map[string]types.ServiceConfig
	if err := viper.UnmarshalKey("services", &services); err != nil {
		logger.Fatal("Failed to load service configs", zap.Error(err))
	}

	return &Proxy{
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		services: services,
	}
}

// HandleRequest 处理代理请求
func (p *Proxy) HandleRequest(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取服务配置
		service, exists := p.services[serviceName]
		if !exists {
			c.JSON(http.StatusBadGateway, gin.H{
				"error": fmt.Sprintf("服务不存在: %s", serviceName),
			})
			return
		}

		// 构建目标 URL
		targetURL, err := p.buildTargetURL(service, c.Request.URL)
		if err != nil {
			p.logger.Error("Failed to build target URL",
				zap.Error(err),
				zap.String("service", serviceName),
			)
			c.JSON(http.StatusBadGateway, gin.H{
				"error": "构建目标URL失败",
			})
			return
		}

		// 读取请求体
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, err = io.ReadAll(c.Request.Body)
			if err != nil {
				p.logger.Error("Failed to read request body", zap.Error(err))
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "读取请求体失败",
				})
				return
			}
		}

		// 创建代理请求
		proxyReq, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewReader(bodyBytes))
		if err != nil {
			p.logger.Error("Failed to create proxy request", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "创建代理请求失败",
			})
			return
		}

		// 复制请求头
		p.copyHeaders(c.Request.Header, proxyReq.Header)

		// 添加网关标识
		proxyReq.Header.Set("X-Gateway", "k8s-agent-gateway")
		proxyReq.Header.Set("X-Forwarded-For", c.ClientIP())

		// 发送请求
		startTime := time.Now()
		resp, err := p.client.Do(proxyReq)
		latency := time.Since(startTime)

		if err != nil {
			p.logger.Error("Proxy request failed",
				zap.Error(err),
				zap.String("service", serviceName),
				zap.String("url", targetURL),
				zap.Duration("latency", latency),
			)
			c.JSON(http.StatusBadGateway, gin.H{
				"error": "服务请求失败",
			})
			return
		}
		defer resp.Body.Close()

		// 读取响应体
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			p.logger.Error("Failed to read response body", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "读取响应失败",
			})
			return
		}

		// 记录日志
		p.logger.Info("Proxy request completed",
			zap.String("service", serviceName),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", resp.StatusCode),
			zap.Duration("latency", latency),
		)

		// 复制响应头
		for key, values := range resp.Header {
			for _, value := range values {
				c.Writer.Header().Add(key, value)
			}
		}

		// 返回响应
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
	}
}

// buildTargetURL 构建目标 URL
func (p *Proxy) buildTargetURL(service types.ServiceConfig, reqURL *url.URL) (string, error) {
	targetURL, err := url.Parse(service.URL)
	if err != nil {
		return "", err
	}

	targetURL.Path = reqURL.Path
	targetURL.RawQuery = reqURL.RawQuery

	return targetURL.String(), nil
}

// copyHeaders 复制请求头
func (p *Proxy) copyHeaders(src, dst http.Header) {
	// 不复制的头
	skipHeaders := map[string]bool{
		"Host":              true,
		"Connection":        true,
		"Keep-Alive":        true,
		"Proxy-Connection":  true,
		"Transfer-Encoding": true,
		"Upgrade":           true,
	}

	for key, values := range src {
		if skipHeaders[key] {
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// GetServiceHealth 获取服务健康状态
func (p *Proxy) GetServiceHealth(serviceName string) (*types.HealthStatus, error) {
	service, exists := p.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", serviceName)
	}

	healthURL := strings.TrimRight(service.URL, "/") + service.HealthCheck

	req, err := http.NewRequest("GET", healthURL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &types.HealthStatus{
			Service:   serviceName,
			Status:    "unhealthy",
			Timestamp: time.Now(),
			Message:   err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	status := "healthy"
	if resp.StatusCode != http.StatusOK {
		status = "unhealthy"
	}

	return &types.HealthStatus{
		Service:   serviceName,
		Status:    status,
		Timestamp: time.Now(),
	}, nil
}
