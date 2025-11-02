package options

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/kart-io/k8s-agent/common/options/validation"
)

// HealthOptions 健康检查配置
// 参考 OneX 项目的设计,用于配置应用程序的健康检查端点
type HealthOptions struct {
	// Enable 是否启用健康检查服务器
	Enable bool `mapstructure:"enable" yaml:"enable" json:"enable"`

	// Host 健康检查服务器监听地址
	Host string `mapstructure:"host" yaml:"host" json:"host"`

	// Port 健康检查服务器监听端口
	Port int `mapstructure:"port" yaml:"port" json:"port"`

	// Path 健康检查请求路径
	Path string `mapstructure:"path" yaml:"path" json:"path"`

	// ReadinessPath 就绪探针请求路径 (K8s readiness probe)
	ReadinessPath string `mapstructure:"readiness_path" yaml:"readiness_path" json:"readiness_path"`

	// LivenessPath 存活探针请求路径 (K8s liveness probe)
	LivenessPath string `mapstructure:"liveness_path" yaml:"liveness_path" json:"liveness_path"`

	// EnablePprof 是否启用 pprof 性能分析端点
	EnablePprof bool `mapstructure:"enable_pprof" yaml:"enable_pprof" json:"enable_pprof"`

	// ShutdownTimeout 优雅关闭超时时间
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" yaml:"shutdown_timeout" json:"shutdown_timeout"`

	// HealthCheckInterval 健康检查间隔时间 (用于主动健康检查)
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval" yaml:"health_check_interval" json:"health_check_interval"`
}

// NewHealthOptions 创建默认的健康检查配置
func NewHealthOptions() *HealthOptions {
	return &HealthOptions{
		Enable:              true,
		Host:                "0.0.0.0",
		Port:                20250, // 参考 OneX 的默认端口
		Path:                "/healthz",
		ReadinessPath:       "/readyz",
		LivenessPath:        "/livez",
		EnablePprof:         false, // 生产环境默认关闭
		ShutdownTimeout:     5 * time.Second,
		HealthCheckInterval: 10 * time.Second,
	}
}

// Validate 验证配置
func (o *HealthOptions) Validate() error {
	if !o.Enable {
		return nil // 健康检查是可选的,如果未启用则跳过验证
	}

	// 使用通用验证器
	if err := validation.ValidateHostPort(o.Host, o.Port, "health check"); err != nil {
		return err
	}

	if err := validation.ValidateRequired(o.Path, "health check path"); err != nil {
		return err
	}

	if err := validation.ValidateRequired(o.ReadinessPath, "readiness path"); err != nil {
		return err
	}

	if err := validation.ValidateRequired(o.LivenessPath, "liveness path"); err != nil {
		return err
	}

	if err := validation.ValidatePositiveDuration(o.ShutdownTimeout, "shutdown timeout"); err != nil {
		return err
	}

	if err := validation.ValidatePositiveDuration(o.HealthCheckInterval, "health check interval"); err != nil {
		return err
	}

	return nil
}

// AddFlags 添加命令行参数
func (o *HealthOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Enable, "health.enable", o.Enable, "Enable health check server")
	fs.StringVar(&o.Host, "health.host", o.Host, "Health check server host address")
	fs.IntVar(&o.Port, "health.port", o.Port, "Health check server port")
	fs.StringVar(&o.Path, "health.path", o.Path, "Health check request path")
	fs.StringVar(&o.ReadinessPath, "health.readiness-path", o.ReadinessPath, "Readiness probe path (K8s)")
	fs.StringVar(&o.LivenessPath, "health.liveness-path", o.LivenessPath, "Liveness probe path (K8s)")
	fs.BoolVar(&o.EnablePprof, "health.enable-pprof", o.EnablePprof, "Enable pprof profiling endpoints")
	fs.DurationVar(&o.ShutdownTimeout, "health.shutdown-timeout", o.ShutdownTimeout, "Health server shutdown timeout")
	fs.DurationVar(&o.HealthCheckInterval, "health.check-interval", o.HealthCheckInterval, "Health check interval")
}

// ApplyTo 将配置应用到目标接口
func (o *HealthOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	// 类型断言,检查是否为配置切片指针
	switch v := target.(type) {
	case *[]interface{}:
		// 将配置转换为通用选项
		*v = append(*v,
			map[string]interface{}{
				"enable":              o.Enable,
				"host":                o.Host,
				"port":                o.Port,
				"path":                o.Path,
				"readinessPath":       o.ReadinessPath,
				"livenessPath":        o.LivenessPath,
				"enablePprof":         o.EnablePprof,
				"shutdownTimeout":     o.ShutdownTimeout,
				"healthCheckInterval": o.HealthCheckInterval,
			},
		)
	}

	return nil
}

// Complete 完成配置初始化
// 设置默认值和计算派生值
func (o *HealthOptions) Complete() error {
	// 如果未启用,跳过初始化
	if !o.Enable {
		return nil
	}

	// 确保 Host 不为空
	if o.Host == "" {
		o.Host = "0.0.0.0"
	}

	// 确保端口在有效范围内
	if o.Port <= 0 || o.Port > 65535 {
		o.Port = 20250
	}

	// 确保路径以 / 开头
	if o.Path == "" {
		o.Path = "/healthz"
	}
	if o.Path[0] != '/' {
		o.Path = "/" + o.Path
	}

	if o.ReadinessPath == "" {
		o.ReadinessPath = "/readyz"
	}
	if o.ReadinessPath[0] != '/' {
		o.ReadinessPath = "/" + o.ReadinessPath
	}

	if o.LivenessPath == "" {
		o.LivenessPath = "/livez"
	}
	if o.LivenessPath[0] != '/' {
		o.LivenessPath = "/" + o.LivenessPath
	}

	// 确保超时时间合理
	if o.ShutdownTimeout <= 0 {
		o.ShutdownTimeout = 5 * time.Second
	}

	if o.HealthCheckInterval <= 0 {
		o.HealthCheckInterval = 10 * time.Second
	}

	return nil
}

// WithHealthEnable 设置是否启用健康检查
func WithHealthEnable(enable bool) func(*HealthOptions) {
	return func(o *HealthOptions) {
		o.Enable = enable
	}
}

// WithHealthHost 设置健康检查服务器地址
func WithHealthHost(host string) func(*HealthOptions) {
	return func(o *HealthOptions) {
		o.Host = host
	}
}

// WithHealthPort 设置健康检查服务器端口
func WithHealthPort(port int) func(*HealthOptions) {
	return func(o *HealthOptions) {
		o.Port = port
	}
}

// WithHealthPath 设置健康检查路径
func WithHealthPath(path string) func(*HealthOptions) {
	return func(o *HealthOptions) {
		o.Path = path
	}
}

// WithReadinessPath 设置就绪探针路径
func WithReadinessPath(path string) func(*HealthOptions) {
	return func(o *HealthOptions) {
		o.ReadinessPath = path
	}
}

// WithLivenessPath 设置存活探针路径
func WithLivenessPath(path string) func(*HealthOptions) {
	return func(o *HealthOptions) {
		o.LivenessPath = path
	}
}

// WithPprofEnabled 设置是否启用 pprof
func WithPprofEnabled(enable bool) func(*HealthOptions) {
	return func(o *HealthOptions) {
		o.EnablePprof = enable
	}
}

// WithShutdownTimeout 设置优雅关闭超时时间
func WithShutdownTimeout(timeout time.Duration) func(*HealthOptions) {
	return func(o *HealthOptions) {
		o.ShutdownTimeout = timeout
	}
}

// WithHealthCheckInterval 设置健康检查间隔
func WithHealthCheckInterval(interval time.Duration) func(*HealthOptions) {
	return func(o *HealthOptions) {
		o.HealthCheckInterval = interval
	}
}
