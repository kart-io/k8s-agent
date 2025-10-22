package config

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

// ServerOptions HTTP服务器配置
type ServerOptions struct {
	Host         string        `mapstructure:"host" yaml:"host" json:"host"`
	Port         int           `mapstructure:"port" yaml:"port" json:"port"`
	Mode         string        `mapstructure:"mode" yaml:"mode" json:"mode"` // debug, release, test
	ReadTimeout  time.Duration `mapstructure:"read_timeout" yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout" yaml:"idle_timeout" json:"idle_timeout"`
	GracefulStop time.Duration `mapstructure:"graceful_stop" yaml:"graceful_stop" json:"graceful_stop"`
}

// NewServerOptions 创建默认的服务器配置
func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		Host:         "0.0.0.0",
		Port:         8080,
		Mode:         "release",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		GracefulStop: 30 * time.Second,
	}
}

// Validate 验证配置
func (o *ServerOptions) Validate() error {
	if o.Port < 1 || o.Port > 65535 {
		return fmt.Errorf("invalid port: %d", o.Port)
	}
	if o.Mode != "debug" && o.Mode != "release" && o.Mode != "test" {
		return fmt.Errorf("invalid mode: %s", o.Mode)
	}
	return nil
}

// AddFlags 添加命令行参数
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Host, "server.host", o.Host, "Server host address")
	fs.IntVar(&o.Port, "server.port", o.Port, "Server port")
	fs.StringVar(&o.Mode, "server.mode", o.Mode, "Server mode (debug, release, test)")
	fs.DurationVar(&o.ReadTimeout, "server.read-timeout", o.ReadTimeout, "Server read timeout")
	fs.DurationVar(&o.WriteTimeout, "server.write-timeout", o.WriteTimeout, "Server write timeout")
	fs.DurationVar(&o.IdleTimeout, "server.idle-timeout", o.IdleTimeout, "Server idle timeout")
	fs.DurationVar(&o.GracefulStop, "server.graceful-stop", o.GracefulStop, "Server graceful stop timeout")
}

// ApplyTo 将配置应用到目标接口
// 接受一个配置结构指针，将服务器配置应用到目标
func (o *ServerOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	// 类型断言，检查是否为配置结构指针
	switch v := target.(type) {
	case *[]interface{}:
		// 将配置转换为通用选项
		*v = append(*v,
			map[string]interface{}{
				"host":         o.Host,
				"port":         o.Port,
				"mode":         o.Mode,
				"readTimeout":  o.ReadTimeout,
				"writeTimeout": o.WriteTimeout,
				"idleTimeout":  o.IdleTimeout,
				"gracefulStop": o.GracefulStop,
			},
		)
	}

	return nil
}

// Complete 完成配置初始化
// 设置默认值和计算派生值
func (o *ServerOptions) Complete() error {
	// 确保主机地址不为空
	if o.Host == "" {
		o.Host = "0.0.0.0"
	}

	// 确保端口在有效范围内
	if o.Port <= 0 || o.Port > 65535 {
		o.Port = 8080
	}

	// 确保模式有效
	if o.Mode != "debug" && o.Mode != "release" && o.Mode != "test" {
		o.Mode = "release"
	}

	// 确保超时时间合理
	if o.ReadTimeout <= 0 {
		o.ReadTimeout = 10 * time.Second
	}

	if o.WriteTimeout <= 0 {
		o.WriteTimeout = 10 * time.Second
	}

	if o.IdleTimeout <= 0 {
		o.IdleTimeout = 60 * time.Second
	}

	if o.GracefulStop <= 0 {
		o.GracefulStop = 30 * time.Second
	}

	// 确保优雅停机时间不超过60秒
	if o.GracefulStop > 60*time.Second {
		o.GracefulStop = 60 * time.Second
	}

	return nil
}

// WithHost 设置服务器地址
func WithHost(host string) func(*ServerOptions) {
	return func(o *ServerOptions) {
		o.Host = host
	}
}

// WithPort 设置服务器端口
func WithPort(port int) func(*ServerOptions) {
	return func(o *ServerOptions) {
		o.Port = port
	}
}

// WithMode 设置服务器模式
func WithMode(mode string) func(*ServerOptions) {
	return func(o *ServerOptions) {
		o.Mode = mode
	}
}

// WithReadTimeout 设置读超时时间
func WithReadTimeout(timeout time.Duration) func(*ServerOptions) {
	return func(o *ServerOptions) {
		o.ReadTimeout = timeout
	}
}

// WithWriteTimeout 设置写超时时间
func WithWriteTimeout(timeout time.Duration) func(*ServerOptions) {
	return func(o *ServerOptions) {
		o.WriteTimeout = timeout
	}
}

// WithIdleTimeout 设置空闲超时时间
func WithIdleTimeout(timeout time.Duration) func(*ServerOptions) {
	return func(o *ServerOptions) {
		o.IdleTimeout = timeout
	}
}

// WithGracefulStop 设置优雅停机时间
func WithGracefulStop(timeout time.Duration) func(*ServerOptions) {
	return func(o *ServerOptions) {
		o.GracefulStop = timeout
	}
}