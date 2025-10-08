package types

import "time"

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name        string        `mapstructure:"name"`
	URL         string        `mapstructure:"url"`
	Timeout     time.Duration `mapstructure:"timeout"`
	Retry       int           `mapstructure:"retry"`
	HealthCheck string        `mapstructure:"health_check"`
}

// RouteConfig 路由配置
type RouteConfig struct {
	Path         string `mapstructure:"path"`
	Service      string `mapstructure:"service"`
	StripPrefix  bool   `mapstructure:"strip_prefix"`
	AuthRequired bool   `mapstructure:"auth_required"`
}

// GatewayConfig 网关配置
type GatewayConfig struct {
	Services map[string]ServiceConfig `mapstructure:"services"`
	Routes   []RouteConfig            `mapstructure:"routes"`
}

// ProxyRequest 代理请求
type ProxyRequest struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
	Query   map[string][]string
}

// ProxyResponse 代理响应
type ProxyResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

// HealthStatus 健康状态
type HealthStatus struct {
	Service   string    `json:"service"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message,omitempty"`
}

// MetricsData 监控数据
type MetricsData struct {
	TotalRequests   int64             `json:"total_requests"`
	SuccessRequests int64             `json:"success_requests"`
	FailedRequests  int64             `json:"failed_requests"`
	AvgLatency      float64           `json:"avg_latency_ms"`
	ServiceMetrics  map[string]int64  `json:"service_metrics"`
}
