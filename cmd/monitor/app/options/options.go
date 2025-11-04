package options

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/kart-io/k8s-agent/common/loggerutil"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

// ServerOptions 定义 monitor 服务的配置选项
// 实现 pkg/app.Options 接口
type ServerOptions struct {
	Server   *commonoptions.ServerOptions   `json:"server" mapstructure:"server"`
	Logging  *commonoptions.LoggingOptions  `json:"logging" mapstructure:"logging"`
	Health   *commonoptions.HealthOptions   `json:"health" mapstructure:"health"`
	Database *commonoptions.DatabaseOptions `json:"database" mapstructure:"database"`
	Redis    *commonoptions.RedisOptions    `json:"redis" mapstructure:"redis"`
	Metrics  *commonoptions.MetricsOptions  `json:"metrics" mapstructure:"metrics"`
	JWT      *commonoptions.JWTOptions      `json:"jwt" mapstructure:"jwt"`

	// Monitor 特有配置
	Prometheus PrometheusConfig `json:"prometheus" mapstructure:"prometheus"`
	Alert      AlertConfig      `json:"alert" mapstructure:"alert"`
}

// PrometheusConfig holds Prometheus configuration.
type PrometheusConfig struct {
	Enabled bool `json:"enabled" mapstructure:"enabled"`
	Port    int  `json:"port" mapstructure:"port"`
}

// AlertConfig holds alert configuration.
type AlertConfig struct {
	CheckInterval string              `json:"check_interval" mapstructure:"check_interval"`
	Channels      AlertChannelsConfig `json:"channels" mapstructure:"channels"`
}

// AlertChannelsConfig holds alert channels configuration.
type AlertChannelsConfig struct {
	Email   EmailAlertConfig   `json:"email" mapstructure:"email"`
	Webhook WebhookAlertConfig `json:"webhook" mapstructure:"webhook"`
	Slack   SlackAlertConfig   `json:"slack" mapstructure:"slack"`
}

// EmailAlertConfig holds email alert configuration.
type EmailAlertConfig struct {
	Enabled  bool   `json:"enabled" mapstructure:"enabled"`
	SMTPHost string `json:"smtp_host" mapstructure:"smtp_host"`
	SMTPPort int    `json:"smtp_port" mapstructure:"smtp_port"`
	From     string `json:"from" mapstructure:"from"`
}

// WebhookAlertConfig holds webhook alert configuration.
type WebhookAlertConfig struct {
	Enabled bool   `json:"enabled" mapstructure:"enabled"`
	URL     string `json:"url" mapstructure:"url"`
}

// SlackAlertConfig holds Slack alert configuration.
type SlackAlertConfig struct {
	Enabled    bool   `json:"enabled" mapstructure:"enabled"`
	WebhookURL string `json:"webhook_url" mapstructure:"webhook_url"`
}

// NewServerOptions 创建新的 ServerOptions 实例，使用默认值
func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		Server:   commonoptions.NewServerOptions(),
		Logging:  commonoptions.NewLoggingOptions(),
		Health:   commonoptions.NewHealthOptions(),
		Database: commonoptions.NewDatabaseOptions(),
		Redis:    commonoptions.NewRedisOptions(),
		Metrics:  commonoptions.NewMetricsOptions(),
		JWT:      commonoptions.NewJWTOptions(),
		Prometheus: PrometheusConfig{
			Enabled: true,
			Port:    9095,
		},
		Alert: AlertConfig{
			CheckInterval: "30s",
		},
	}
}

// Validate 验证所有必需的配置选项
func (o *ServerOptions) Validate() []error {
	var errs []error

	if err := o.Server.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Logging.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Database.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Redis.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.JWT.Validate(); err != nil {
		errs = append(errs, err)
	}

	// Validate monitor specific options
	if o.Prometheus.Enabled && o.Prometheus.Port == 0 {
		errs = append(errs, fmt.Errorf("prometheus.port is required when enabled"))
	}

	return errs
}

// Complete 填充未设置但需要有效数据的字段
func (o *ServerOptions) Complete() error {
	if err := o.Server.Complete(); err != nil {
		return err
	}

	if err := o.Logging.Complete(); err != nil {
		return err
	}

	if err := o.Database.Complete(); err != nil {
		return err
	}

	if err := o.Redis.Complete(); err != nil {
		return err
	}

	if err := o.JWT.Complete(); err != nil {
		return err
	}

	// Set defaults for monitor specific options
	if o.Prometheus.Port == 0 {
		o.Prometheus.Port = 9095
	}

	if o.Alert.CheckInterval == "" {
		o.Alert.CheckInterval = "30s"
	}

	return nil
}

// AddFlags 添加 flags 到 flag set
// 注意: --config/-c flag 由 pkg/app 框架自动添加
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	o.Server.AddFlags(fs)
	o.Logging.AddFlags(fs)
	o.Health.AddFlags(fs)
	o.Database.AddFlags(fs)
	o.Redis.AddFlags(fs)
	o.Metrics.AddFlags(fs)
	o.JWT.AddFlags(fs)

	// Add monitor specific flags
	fs.BoolVar(&o.Prometheus.Enabled, "prometheus.enabled", o.Prometheus.Enabled,
		"Enable Prometheus metrics")

	fs.IntVar(&o.Prometheus.Port, "prometheus.port", o.Prometheus.Port,
		"Prometheus metrics port")

	fs.StringVar(&o.Alert.CheckInterval, "alert.check-interval", o.Alert.CheckInterval,
		"Alert check interval")
}

// InitLogger 基于配置初始化 logger
func (o *ServerOptions) InitLogger() (core.Logger, error) {
	return loggerutil.InitFromOptions(o.Logging)
}

// GetServiceName 返回服务名称
func (o *ServerOptions) GetServiceName() string {
	return "Monitor"
}

// GetLogFields 返回初始化日志的字段
func (o *ServerOptions) GetLogFields() []interface{} {
	return []interface{}{
		"http_port", o.Server.Port,
		"health_port", o.Health.Port,
		"prometheus_enabled", o.Prometheus.Enabled,
		"prometheus_port", o.Prometheus.Port,
	}
}
