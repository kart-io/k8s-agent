package options

import (
	"errors"
	"time"

	"github.com/spf13/pflag"

	"github.com/kart-io/k8s-agent/common/loggerutil"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

// Configuration validation errors.
var (
	ErrInvalidRateLimit = errors.New("requests_per_second must be greater than 0")
	ErrInvalidBurst     = errors.New("burst must be greater than 0")
)

// ServerOptions 定义 gateway 服务的配置选项
// 实现 pkg/app.Options 接口
type ServerOptions struct {
	Server  *commonoptions.ServerOptions  `json:"server" mapstructure:"server"`
	Logging *commonoptions.LoggingOptions `json:"logging" mapstructure:"logging"`
	Health  *commonoptions.HealthOptions  `json:"health" mapstructure:"health"`
	Redis   *commonoptions.RedisOptions   `json:"redis" mapstructure:"redis"`
	JWT     *commonoptions.JWTOptions     `json:"jwt" mapstructure:"jwt"`

	// 使用 common/options 标准配置
	RateLimit *commonoptions.RateLimitOptions `json:"rate_limit" mapstructure:"rate_limit"`
	CORS      *commonoptions.CORSOptions      `json:"cors" mapstructure:"cors"`
	Metrics   *commonoptions.MetricsOptions   `json:"metrics" mapstructure:"metrics"`

	// Gateway 特有配置
	Services    ServicesOptions    `json:"services" mapstructure:"services"`
	Routes      []RouteOptions     `json:"routes" mapstructure:"routes"`
	HealthCheck HealthCheckOptions `json:"health_check" mapstructure:"health_check"`
}

// ServiceOptions holds individual service configuration.
type ServiceOptions struct {
	Name        string        `json:"name" mapstructure:"name"`
	URL         string        `json:"url" mapstructure:"url"`
	Timeout     time.Duration `json:"timeout" mapstructure:"timeout"`
	Retry       int           `json:"retry" mapstructure:"retry"`
	HealthCheck string        `json:"health_check" mapstructure:"health_check"`
}

// ServicesOptions holds all backend services configuration.
type ServicesOptions struct {
	Auth         ServiceOptions `json:"auth" mapstructure:"auth"`
	AgentManager ServiceOptions `json:"agent_manager" mapstructure:"agent_manager"`
	Orchestrator ServiceOptions `json:"orchestrator" mapstructure:"orchestrator"`
	Reasoning    ServiceOptions `json:"reasoning" mapstructure:"reasoning"`
}

// RouteOptions holds route configuration.
type RouteOptions struct {
	Path         string `json:"path" mapstructure:"path"`
	Service      string `json:"service" mapstructure:"service"`
	StripPrefix  bool   `json:"strip_prefix" mapstructure:"strip_prefix"`
	AuthRequired bool   `json:"auth_required" mapstructure:"auth_required"`
}

// HealthCheckOptions holds health check configuration.
type HealthCheckOptions struct {
	Enabled  bool          `json:"enabled" mapstructure:"enabled"`
	Interval time.Duration `json:"interval" mapstructure:"interval"`
	Timeout  time.Duration `json:"timeout" mapstructure:"timeout"`
}

// NewServerOptions 创建新的 ServerOptions 实例，使用默认值
func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		Server:  commonoptions.NewServerOptions(),
		Logging: commonoptions.NewLoggingOptions(),
		Health:  commonoptions.NewHealthOptions(),
		Redis:   commonoptions.NewRedisOptions(),
		JWT:     commonoptions.NewJWTOptions(),

		// 使用 common/options 标准配置
		RateLimit: commonoptions.NewRateLimitOptions(),
		CORS:      commonoptions.NewCORSOptions(),
		Metrics:   commonoptions.NewMetricsOptions(),

		Services: ServicesOptions{
			Auth: ServiceOptions{
				Name:        "auth",
				URL:         "http://localhost:8080",
				Timeout:     10 * time.Second,
				Retry:       3,
				HealthCheck: "/health",
			},
			AgentManager: ServiceOptions{
				Name:        "agent-manager",
				URL:         "http://localhost:8081",
				Timeout:     10 * time.Second,
				Retry:       3,
				HealthCheck: "/health",
			},
			Orchestrator: ServiceOptions{
				Name:        "orchestrator",
				URL:         "http://localhost:8082",
				Timeout:     10 * time.Second,
				Retry:       3,
				HealthCheck: "/health",
			},
			Reasoning: ServiceOptions{
				Name:        "reasoning",
				URL:         "http://localhost:8083",
				Timeout:     10 * time.Second,
				Retry:       3,
				HealthCheck: "/health",
			},
		},
		HealthCheck: HealthCheckOptions{
			Enabled:  true,
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
		},
	}
}

// Validate 验证所有必需的配置选项
func (o *ServerOptions) Validate() []error {
	// 使用通用工具函数统一验证所有子选项
	errs := commonoptions.ValidateAll(o)

	// Validate gateway specific options
	if o.RateLimit != nil && o.RateLimit.Enable {
		if o.RateLimit.Rate <= 0 {
			errs = append(errs, ErrInvalidRateLimit)
		}
		if o.RateLimit.Burst <= 0 {
			errs = append(errs, ErrInvalidBurst)
		}
	}

	return errs
}

// Complete 填充未设置但需要有效数据的字段
func (o *ServerOptions) Complete() error {
	// 使用通用工具函数统一完成所有子选项
	if err := commonoptions.CompleteAll(o); err != nil {
		return err
	}

	// Set defaults for gateway specific options
	if o.HealthCheck.Interval == 0 {
		o.HealthCheck.Interval = 30 * time.Second
	}

	if o.HealthCheck.Timeout == 0 {
		o.HealthCheck.Timeout = 5 * time.Second
	}

	return nil
}

// AddFlags 添加 flags 到 flag set
// 注意: --config/-c flag 由 pkg/app 框架自动添加
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	// 使用通用工具函数统一添加所有子选项的 flags
	commonoptions.AddFlagsAll(o, fs)

	// Add gateway specific flags
	fs.BoolVar(&o.HealthCheck.Enabled, "health-check.enabled", o.HealthCheck.Enabled,
		"Enable health check for backend services")

	fs.DurationVar(&o.HealthCheck.Interval, "health-check.interval", o.HealthCheck.Interval,
		"Health check interval")
}

// InitLogger 基于配置初始化 logger
func (o *ServerOptions) InitLogger() (core.Logger, error) {
	return loggerutil.InitFromOptions(o.Logging)
}

// GetServiceName 返回服务名称
func (o *ServerOptions) GetServiceName() string {
	return "Gateway"
}

// GetLogFields 返回初始化日志的字段
func (o *ServerOptions) GetLogFields() []interface{} {
	return []interface{}{
		"http_port", o.Server.Port,
		"health_port", o.Health.Port,
		"rate_limit_enabled", o.RateLimit.Enable,
		"cors_enabled", o.CORS.Enabled,
	}
}
