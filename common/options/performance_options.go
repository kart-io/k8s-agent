package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

// PerformanceOptions 性能配置选项
type PerformanceOptions struct {
	MaxWorkers     int    `mapstructure:"max_workers" yaml:"max_workers" json:"max_workers"`
	RequestTimeout string `mapstructure:"request_timeout" yaml:"request_timeout" json:"request_timeout"`
	MaxContextSize int    `mapstructure:"max_context_size" yaml:"max_context_size" json:"max_context_size"` // 日志/上下文最大字符数
}

// NewPerformanceOptions 创建默认的性能配置
func NewPerformanceOptions() *PerformanceOptions {
	return &PerformanceOptions{
		MaxWorkers:     10,
		RequestTimeout: "30s",
		MaxContextSize: 100000,
	}
}

// Validate 验证配置
func (o *PerformanceOptions) Validate() error {
	if o.MaxWorkers < 1 {
		return fmt.Errorf("max_workers must be at least 1")
	}
	if o.MaxContextSize < 1000 {
		return fmt.Errorf("max_context_size must be at least 1000")
	}

	// 验证 timeout 格式
	if _, err := time.ParseDuration(o.RequestTimeout); err != nil {
		return fmt.Errorf("invalid request_timeout format: %w", err)
	}

	return nil
}

// AddFlags 添加命令行参数
func (o *PerformanceOptions) AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&o.MaxWorkers, "performance.max-workers", o.MaxWorkers, "Maximum number of workers")
	fs.StringVar(&o.RequestTimeout, "performance.request-timeout", o.RequestTimeout, "Request timeout duration")
	fs.IntVar(&o.MaxContextSize, "performance.max-context-size", o.MaxContextSize, "Maximum context size in characters")
}

// ApplyTo 将配置应用到目标接口
func (o *PerformanceOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	switch v := target.(type) {
	case *[]interface{}:
		*v = append(*v,
			map[string]interface{}{
				"maxWorkers":     o.MaxWorkers,
				"requestTimeout": o.RequestTimeout,
				"maxContextSize": o.MaxContextSize,
			},
		)
	}

	return nil
}

// Complete 完成配置初始化
func (o *PerformanceOptions) Complete() error {
	if o.MaxWorkers <= 0 {
		o.MaxWorkers = 10
	}
	if o.RequestTimeout == "" {
		o.RequestTimeout = "30s"
	}
	if o.MaxContextSize <= 0 {
		o.MaxContextSize = 100000
	}

	// 确保超时时间合理
	duration, err := time.ParseDuration(o.RequestTimeout)
	if err == nil {
		if duration < time.Second {
			o.RequestTimeout = "1s"
		} else if duration > 5*time.Minute {
			o.RequestTimeout = "5m"
		}
	}

	return nil
}

// GetRequestTimeoutDuration 返回请求超时时长
func (o *PerformanceOptions) GetRequestTimeoutDuration() time.Duration {
	duration, err := time.ParseDuration(o.RequestTimeout)
	if err != nil {
		return 30 * time.Second // Default
	}
	return duration
}

// WithMaxWorkers 设置最大工作线程数
func WithMaxWorkers(workers int) func(*PerformanceOptions) {
	return func(o *PerformanceOptions) {
		o.MaxWorkers = workers
	}
}

// WithRequestTimeout 设置请求超时时间
func WithRequestTimeout(timeout string) func(*PerformanceOptions) {
	return func(o *PerformanceOptions) {
		o.RequestTimeout = timeout
	}
}

// WithMaxContextSize 设置最大上下文大小
func WithMaxContextSize(size int) func(*PerformanceOptions) {
	return func(o *PerformanceOptions) {
		o.MaxContextSize = size
	}
}
