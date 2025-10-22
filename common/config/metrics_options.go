package config

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

// MetricsOptions 指标采集配置
type MetricsOptions struct {
	Enabled             bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Path                string        `mapstructure:"path" yaml:"path" json:"path"`
	Port                int           `mapstructure:"port" yaml:"port" json:"port"`
	RetentionDays       int           `mapstructure:"retention_days" yaml:"retention_days" json:"retention_days"`
	AggregationInterval time.Duration `mapstructure:"aggregation_interval" yaml:"aggregation_interval" json:"aggregation_interval"`
}

// NewMetricsOptions 创建默认的指标配置
func NewMetricsOptions() *MetricsOptions {
	return &MetricsOptions{
		Enabled:             true,
		Path:                "/metrics",
		Port:                9090,
		RetentionDays:       30,
		AggregationInterval: 1 * time.Minute,
	}
}

// Validate 验证配置
func (o *MetricsOptions) Validate() error {
	if o.Enabled {
		if o.Path == "" {
			return fmt.Errorf("metrics path is required when enabled")
		}
		if o.Port < 1 || o.Port > 65535 {
			return fmt.Errorf("invalid metrics port: %d", o.Port)
		}
		if o.RetentionDays < 1 {
			return fmt.Errorf("retention_days must be > 0")
		}
	}
	return nil
}

// AddFlags 添加命令行参数
func (o *MetricsOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Enabled, "metrics.enabled", o.Enabled, "Enable metrics collection")
	fs.StringVar(&o.Path, "metrics.path", o.Path, "Metrics endpoint path")
	fs.IntVar(&o.Port, "metrics.port", o.Port, "Metrics server port")
	fs.IntVar(&o.RetentionDays, "metrics.retention-days", o.RetentionDays, "Metrics retention days")
	fs.DurationVar(&o.AggregationInterval, "metrics.aggregation-interval", o.AggregationInterval, "Metrics aggregation interval")
}

// ApplyTo 将配置应用到目标接口
// 接受一个配置结构指针，将指标配置应用到目标
func (o *MetricsOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	// 类型断言，检查是否为配置结构指针
	switch v := target.(type) {
	case *[]interface{}:
		// 将配置转换为通用选项
		*v = append(*v,
			map[string]interface{}{
				"enabled":             o.Enabled,
				"path":                o.Path,
				"port":                o.Port,
				"retentionDays":       o.RetentionDays,
				"aggregationInterval": o.AggregationInterval,
			},
		)
	}

	return nil
}

// Complete 完成配置初始化
// 设置默认值和计算派生值
func (o *MetricsOptions) Complete() error {
	if !o.Enabled {
		// 如果指标收集未启用，无需完成其他配置
		return nil
	}

	// 确保路径不为空
	if o.Path == "" {
		o.Path = "/metrics"
	}

	// 确保路径以 / 开头
	if o.Path[0] != '/' {
		o.Path = "/" + o.Path
	}

	// 确保端口在有效范围内
	if o.Port <= 0 || o.Port > 65535 {
		o.Port = 9090
	}

	// 确保保留天数合理 (1-365天)
	if o.RetentionDays < 1 {
		o.RetentionDays = 30
	} else if o.RetentionDays > 365 {
		o.RetentionDays = 365
	}

	// 确保聚合间隔合理 (1秒-1小时)
	if o.AggregationInterval < 1*time.Second {
		o.AggregationInterval = 1 * time.Minute
	} else if o.AggregationInterval > 1*time.Hour {
		o.AggregationInterval = 1 * time.Hour
	}

	return nil
}


// WithMetricsEnabled 启用指标采集
func WithMetricsEnabled(enabled bool) func(*MetricsOptions) {
	return func(o *MetricsOptions) {
		o.Enabled = enabled
	}
}

// WithMetricsPath 设置指标路径
func WithMetricsPath(path string) func(*MetricsOptions) {
	return func(o *MetricsOptions) {
		o.Path = path
	}
}

// WithMetricsPort 设置指标端口
func WithMetricsPort(port int) func(*MetricsOptions) {
	return func(o *MetricsOptions) {
		o.Port = port
	}
}

// WithMetricsRetentionDays 设置指标保留天数
func WithMetricsRetentionDays(days int) func(*MetricsOptions) {
	return func(o *MetricsOptions) {
		o.RetentionDays = days
	}
}

// WithMetricsAggregationInterval 设置聚合间隔
func WithMetricsAggregationInterval(interval time.Duration) func(*MetricsOptions) {
	return func(o *MetricsOptions) {
		o.AggregationInterval = interval
	}
}