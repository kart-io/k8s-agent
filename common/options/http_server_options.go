package options

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/kart-io/k8s-agent/common/options/validation"
)

// HTTPServerOptions HTTP 服务器配置选项
//
// Deprecated: 此类型已废弃，请使用 ServerOptions 代替。
// ServerOptions 现在包含了所有 HTTPServerOptions 的功能，并提供了更统一的接口。
//
// 迁移指南：
//   - HTTPServerOptions.Network -> ServerOptions.Network
//   - HTTPServerOptions.Addr -> ServerOptions.GetAddr() 或分别使用 Host 和 Port
//   - HTTPServerOptions.Timeout -> 不再需要（这是客户端超时，不是服务器配置）
//   - HTTPServerOptions.ReadTimeout -> ServerOptions.ReadTimeout
//   - HTTPServerOptions.WriteTimeout -> ServerOptions.WriteTimeout
//   - HTTPServerOptions.IdleTimeout -> ServerOptions.IdleTimeout
//   - HTTPServerOptions.MaxHeaderBytes -> ServerOptions.MaxHeaderBytes
//
// 参考 OneX 项目设计，用于配置 HTTP 服务器基础参数
// 这是一个简化版本，适用于不需要 TLS 的 HTTP 服务器场景
type HTTPServerOptions struct {
	// Network 网络类型（tcp, tcp4, tcp6, unix, unixpacket）
	Network string `mapstructure:"network" yaml:"network" json:"network"`

	// Addr 监听地址（格式: host:port 或 :port）
	// 例如: "0.0.0.0:8080", ":8080", "127.0.0.1:8080"
	Addr string `mapstructure:"addr" yaml:"addr" json:"addr"`

	// Timeout HTTP 客户端超时时间（用于反向代理场景）
	// 这是客户端侧的超时时间，不是服务器侧
	Timeout time.Duration `mapstructure:"timeout" yaml:"timeout" json:"timeout"`

	// ReadTimeout 服务器读取超时时间
	// 从连接建立到读取请求体完成的最大时间
	ReadTimeout time.Duration `mapstructure:"read_timeout" yaml:"read_timeout" json:"read_timeout"`

	// WriteTimeout 服务器写入超时时间
	// 从读取请求头完成到写入响应完成的最大时间
	WriteTimeout time.Duration `mapstructure:"write_timeout" yaml:"write_timeout" json:"write_timeout"`

	// IdleTimeout 空闲连接超时时间
	// Keep-alive 连接的空闲超时时间
	IdleTimeout time.Duration `mapstructure:"idle_timeout" yaml:"idle_timeout" json:"idle_timeout"`

	// MaxHeaderBytes 最大请求头大小（字节）
	MaxHeaderBytes int `mapstructure:"max_header_bytes" yaml:"max_header_bytes" json:"max_header_bytes"`
}

// NewHTTPServerOptions 创建默认的 HTTP 服务器配置
func NewHTTPServerOptions() *HTTPServerOptions {
	return &HTTPServerOptions{
		Network:        "tcp",
		Addr:           "0.0.0.0:8080",
		Timeout:        30 * time.Second, // 客户端超时
		ReadTimeout:    15 * time.Second, // 服务器读超时
		WriteTimeout:   15 * time.Second, // 服务器写超时
		IdleTimeout:    60 * time.Second, // 空闲超时
		MaxHeaderBytes: 1 << 20,          // 1 MB
	}
}

// Validate 验证配置
func (o *HTTPServerOptions) Validate() error {
	// 验证监听地址
	if err := validation.ValidateAddr(o.Addr, "http server"); err != nil {
		return err
	}

	// 验证网络类型
	validNetworks := []string{"tcp", "tcp4", "tcp6", "unix", "unixpacket"}
	if !ContainsString(validNetworks, o.Network) {
		return validation.ValidateEnum(o.Network, "network", validNetworks)
	}

	// 验证超时时间
	if err := validation.ValidatePositiveDuration(o.Timeout, "timeout"); err != nil {
		return err
	}

	if err := validation.ValidatePositiveDuration(o.ReadTimeout, "read_timeout"); err != nil {
		return err
	}

	if err := validation.ValidatePositiveDuration(o.WriteTimeout, "write_timeout"); err != nil {
		return err
	}

	if err := validation.ValidatePositiveDuration(o.IdleTimeout, "idle_timeout"); err != nil {
		return err
	}

	// 验证最大请求头大小
	if err := validation.ValidatePositiveInt(o.MaxHeaderBytes, "max_header_bytes"); err != nil {
		return err
	}

	return nil
}

// AddFlags 添加命令行参数
func (o *HTTPServerOptions) AddFlags(fs *pflag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}

	fs.StringVar(&o.Network, prefix+"http.network", o.Network,
		"HTTP server network type (tcp, tcp4, tcp6, unix, unixpacket)")
	fs.StringVar(&o.Addr, prefix+"http.addr", o.Addr,
		"HTTP server bind address (e.g., :8080, 0.0.0.0:8080)")
	fs.DurationVar(&o.Timeout, prefix+"http.timeout", o.Timeout,
		"HTTP client timeout (for reverse proxy scenarios)")
	fs.DurationVar(&o.ReadTimeout, prefix+"http.read-timeout", o.ReadTimeout,
		"HTTP server read timeout")
	fs.DurationVar(&o.WriteTimeout, prefix+"http.write-timeout", o.WriteTimeout,
		"HTTP server write timeout")
	fs.DurationVar(&o.IdleTimeout, prefix+"http.idle-timeout", o.IdleTimeout,
		"HTTP server idle timeout (keep-alive)")
	fs.IntVar(&o.MaxHeaderBytes, prefix+"http.max-header-bytes", o.MaxHeaderBytes,
		"Maximum size of request headers in bytes")
}

// Complete 完成配置初始化
func (o *HTTPServerOptions) Complete() error {
	// 确保网络类型合理
	if o.Network == "" {
		o.Network = "tcp"
	}

	// 确保地址不为空
	if o.Addr == "" {
		o.Addr = "0.0.0.0:8080"
	}

	// 确保超时时间合理
	if o.Timeout <= 0 {
		o.Timeout = 30 * time.Second
	}

	if o.ReadTimeout <= 0 {
		o.ReadTimeout = 15 * time.Second
	}

	if o.WriteTimeout <= 0 {
		o.WriteTimeout = 15 * time.Second
	}

	if o.IdleTimeout <= 0 {
		o.IdleTimeout = 60 * time.Second
	}

	// 确保最大请求头大小合理
	if o.MaxHeaderBytes <= 0 {
		o.MaxHeaderBytes = 1 << 20 // 1 MB
	}

	return nil
}

// ApplyTo 将配置应用到目标接口
func (o *HTTPServerOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	switch v := target.(type) {
	case *[]interface{}:
		*v = append(*v,
			map[string]interface{}{
				"network":        o.Network,
				"addr":           o.Addr,
				"timeout":        o.Timeout,
				"readTimeout":    o.ReadTimeout,
				"writeTimeout":   o.WriteTimeout,
				"idleTimeout":    o.IdleTimeout,
				"maxHeaderBytes": o.MaxHeaderBytes,
			},
		)
	}

	return nil
}

// WithHTTPServerNetwork 设置网络类型
func WithHTTPServerNetwork(network string) func(*HTTPServerOptions) {
	return func(o *HTTPServerOptions) {
		o.Network = network
	}
}

// WithHTTPServerAddr 设置监听地址
func WithHTTPServerAddr(addr string) func(*HTTPServerOptions) {
	return func(o *HTTPServerOptions) {
		o.Addr = addr
	}
}

// WithHTTPServerTimeout 设置客户端超时时间
func WithHTTPServerTimeout(timeout time.Duration) func(*HTTPServerOptions) {
	return func(o *HTTPServerOptions) {
		o.Timeout = timeout
	}
}

// WithHTTPServerReadTimeout 设置读超时时间
func WithHTTPServerReadTimeout(timeout time.Duration) func(*HTTPServerOptions) {
	return func(o *HTTPServerOptions) {
		o.ReadTimeout = timeout
	}
}

// WithHTTPServerWriteTimeout 设置写超时时间
func WithHTTPServerWriteTimeout(timeout time.Duration) func(*HTTPServerOptions) {
	return func(o *HTTPServerOptions) {
		o.WriteTimeout = timeout
	}
}

// WithHTTPServerIdleTimeout 设置空闲超时时间
func WithHTTPServerIdleTimeout(timeout time.Duration) func(*HTTPServerOptions) {
	return func(o *HTTPServerOptions) {
		o.IdleTimeout = timeout
	}
}

// WithHTTPServerMaxHeaderBytes 设置最大请求头大小
func WithHTTPServerMaxHeaderBytes(size int) func(*HTTPServerOptions) {
	return func(o *HTTPServerOptions) {
		o.MaxHeaderBytes = size
	}
}
