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

// Options defines options for gateway service
// This implements the pkg/app.Options interface.
type Options struct {
	Server  *commonoptions.ServerOptions  `json:"server" mapstructure:"server"`
	Logging *commonoptions.LoggingOptions `json:"logging" mapstructure:"logging"`
	Redis   *commonoptions.RedisOptions   `json:"redis" mapstructure:"redis"`
	JWT     *commonoptions.JWTOptions     `json:"jwt" mapstructure:"jwt"`

	// Gateway specific options
	RateLimit   RateLimitOptions   `json:"rate_limit" mapstructure:"rate_limit"`
	CORS        CORSOptions        `json:"cors" mapstructure:"cors"`
	Services    ServicesOptions    `json:"services" mapstructure:"services"`
	Routes      []RouteOptions     `json:"routes" mapstructure:"routes"`
	HealthCheck HealthCheckOptions `json:"health_check" mapstructure:"health_check"`
	Metrics     MetricsOptions     `json:"metrics" mapstructure:"metrics"`
}

// RateLimitOptions holds rate limiting configuration.
type RateLimitOptions struct {
	Enabled           bool `json:"enabled" mapstructure:"enabled"`
	RequestsPerSecond int  `json:"requests_per_second" mapstructure:"requests_per_second"`
	Burst             int  `json:"burst" mapstructure:"burst"`
}

// CORSOptions holds CORS configuration.
type CORSOptions struct {
	Enabled          bool          `json:"enabled" mapstructure:"enabled"`
	AllowOrigins     []string      `json:"allow_origins" mapstructure:"allow_origins"`
	AllowMethods     []string      `json:"allow_methods" mapstructure:"allow_methods"`
	AllowHeaders     []string      `json:"allow_headers" mapstructure:"allow_headers"`
	ExposeHeaders    []string      `json:"expose_headers" mapstructure:"expose_headers"`
	AllowCredentials bool          `json:"allow_credentials" mapstructure:"allow_credentials"`
	MaxAge           time.Duration `json:"max_age" mapstructure:"max_age"`
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

// MetricsOptions holds metrics configuration.
type MetricsOptions struct {
	Enabled bool   `json:"enabled" mapstructure:"enabled"`
	Path    string `json:"path" mapstructure:"path"`
}

// NewOptions creates a new Options instance with default values.
func NewOptions() *Options {
	return &Options{
		Server:  commonoptions.NewServerOptions(),
		Logging: commonoptions.NewLoggingOptions(),
		Redis:   commonoptions.NewRedisOptions(),
		JWT:     commonoptions.NewJWTOptions(),
		RateLimit: RateLimitOptions{
			Enabled:           true,
			RequestsPerSecond: 100,
			Burst:             200,
		},
		CORS: CORSOptions{
			Enabled:          true,
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		},
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
		Metrics: MetricsOptions{
			Enabled: true,
			Path:    "/metrics",
		},
	}
}

// Validate validates all the required options.
func (o *Options) Validate() []error {
	var errs []error

	if err := o.Server.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Logging.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Redis.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.JWT.Validate(); err != nil {
		errs = append(errs, err)
	}

	// Validate gateway specific options
	if o.RateLimit.Enabled {
		if o.RateLimit.RequestsPerSecond <= 0 {
			errs = append(errs, ErrInvalidRateLimit)
		}
		if o.RateLimit.Burst <= 0 {
			errs = append(errs, ErrInvalidBurst)
		}
	}

	return errs
}

// Complete fills in any fields not set that are required to have valid data.
func (o *Options) Complete() error {
	if err := o.Server.Complete(); err != nil {
		return err
	}

	if err := o.Logging.Complete(); err != nil {
		return err
	}

	if err := o.Redis.Complete(); err != nil {
		return err
	}

	if err := o.JWT.Complete(); err != nil {
		return err
	}

	// Set defaults for gateway specific options
	if o.RateLimit.RequestsPerSecond == 0 {
		o.RateLimit.RequestsPerSecond = 100
	}

	if o.RateLimit.Burst == 0 {
		o.RateLimit.Burst = 200
	}

	if o.HealthCheck.Interval == 0 {
		o.HealthCheck.Interval = 30 * time.Second
	}

	if o.HealthCheck.Timeout == 0 {
		o.HealthCheck.Timeout = 5 * time.Second
	}

	if o.Metrics.Path == "" {
		o.Metrics.Path = "/metrics"
	}

	return nil
}

// GetHealthPort 实现 commonapp.HealthPortProvider 接口
// 简化版本：直接返回固定端口，不使用HealthOptions.
func (o *Options) GetHealthPort() int {
	return 8095 // Gateway 健康检查端口
}

// AddFlags adds flags to the flag set
// Note: --config/-c flag is automatically added by pkg/app framework.
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.Server.AddFlags(fs)
	o.Logging.AddFlags(fs)
	o.Redis.AddFlags(fs)
	o.JWT.AddFlags(fs)

	// Add gateway specific flags
	fs.BoolVar(&o.RateLimit.Enabled, "rate-limit.enabled", o.RateLimit.Enabled,
		"Enable rate limiting")

	fs.IntVar(&o.RateLimit.RequestsPerSecond, "rate-limit.requests-per-second", o.RateLimit.RequestsPerSecond,
		"Maximum requests per second")

	fs.IntVar(&o.RateLimit.Burst, "rate-limit.burst", o.RateLimit.Burst,
		"Burst size for rate limiting")

	fs.BoolVar(&o.CORS.Enabled, "cors.enabled", o.CORS.Enabled,
		"Enable CORS middleware")

	fs.StringSliceVar(&o.CORS.AllowOrigins, "cors.allow-origins", o.CORS.AllowOrigins,
		"Allowed origins for CORS")

	fs.BoolVar(&o.HealthCheck.Enabled, "health-check.enabled", o.HealthCheck.Enabled,
		"Enable health check for backend services")

	fs.DurationVar(&o.HealthCheck.Interval, "health-check.interval", o.HealthCheck.Interval,
		"Health check interval")

	fs.BoolVar(&o.Metrics.Enabled, "metrics.enabled", o.Metrics.Enabled,
		"Enable metrics endpoint")

	fs.StringVar(&o.Metrics.Path, "metrics.path", o.Metrics.Path,
		"Path for metrics endpoint")
}

// InitLogger initializes the logger based on the options.
func (o *Options) InitLogger() (core.Logger, error) {
	return loggerutil.InitFromOptions(o.Logging)
}
