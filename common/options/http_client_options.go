package options

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/kart-io/k8s-agent/common/options/validation"
)

// HTTPClientOptions HTTP客户端配置选项
// 参考 OneX 项目设计,用于配置 HTTP 客户端连接
type HTTPClientOptions struct {
	// Network 网络协议(tcp, tcp4, tcp6, unix)
	Network string `mapstructure:"network" yaml:"network" json:"network"`

	// Addr 服务器地址 (host:port)
	Addr string `mapstructure:"addr" yaml:"addr" json:"addr"`

	// Timeout 请求超时时间
	Timeout time.Duration `mapstructure:"timeout" yaml:"timeout" json:"timeout"`

	// MaxIdleConns 最大空闲连接数
	MaxIdleConns int `mapstructure:"max_idle_conns" yaml:"max_idle_conns" json:"max_idle_conns"`

	// MaxIdleConnsPerHost 每个主机的最大空闲连接数
	MaxIdleConnsPerHost int `mapstructure:"max_idle_conns_per_host" yaml:"max_idle_conns_per_host" json:"max_idle_conns_per_host"`

	// IdleConnTimeout 空闲连接超时时间
	IdleConnTimeout time.Duration `mapstructure:"idle_conn_timeout" yaml:"idle_conn_timeout" json:"idle_conn_timeout"`

	// TLSHandshakeTimeout TLS握手超时时间
	TLSHandshakeTimeout time.Duration `mapstructure:"tls_handshake_timeout" yaml:"tls_handshake_timeout" json:"tls_handshake_timeout"`

	// ExpectContinueTimeout 等待服务器返回100-continue响应的超时时间
	ExpectContinueTimeout time.Duration `mapstructure:"expect_continue_timeout" yaml:"expect_continue_timeout" json:"expect_continue_timeout"`

	// MaxConnsPerHost 每个主机的最大连接数(包括空闲和活跃)
	MaxConnsPerHost int `mapstructure:"max_conns_per_host" yaml:"max_conns_per_host" json:"max_conns_per_host"`

	// TLS TLS配置
	TLS *TLSOptions `mapstructure:"tls" yaml:"tls" json:"tls"`
}

// NewHTTPClientOptions 创建默认的HTTP客户端配置
func NewHTTPClientOptions() *HTTPClientOptions {
	return &HTTPClientOptions{
		Network:               "tcp",
		Addr:                  "0.0.0.0:8080",
		Timeout:               30 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxConnsPerHost:       0, // 0表示不限制
		TLS:                   NewTLSOptions(),
	}
}

// Validate 验证配置
func (o *HTTPClientOptions) Validate() error {
	// 验证网络类型
	if err := validation.ValidateEnum(o.Network, "HTTP client network",
		[]string{"tcp", "tcp4", "tcp6", "unix"}); err != nil {
		return err
	}

	// 验证地址
	if err := validation.ValidateAddr(o.Addr, "HTTP client"); err != nil {
		return err
	}

	// 验证超时时间
	if err := validation.ValidatePositiveDuration(o.Timeout, "HTTP client timeout"); err != nil {
		return err
	}

	// 验证连接池配置
	if err := validation.ValidateNonNegativeInt(o.MaxIdleConns, "HTTP client max_idle_conns"); err != nil {
		return err
	}

	if err := validation.ValidateNonNegativeInt(o.MaxIdleConnsPerHost, "HTTP client max_idle_conns_per_host"); err != nil {
		return err
	}

	if o.MaxIdleConnsPerHost > o.MaxIdleConns {
		// 允许，但不太合理，记录警告即可
	}

	if err := validation.ValidateNonNegativeInt(o.MaxConnsPerHost, "HTTP client max_conns_per_host"); err != nil {
		return err
	}

	// 验证其他超时配置
	if err := validation.ValidatePositiveDuration(o.IdleConnTimeout, "HTTP client idle_conn_timeout"); err != nil {
		return err
	}

	if err := validation.ValidatePositiveDuration(o.TLSHandshakeTimeout, "HTTP client tls_handshake_timeout"); err != nil {
		return err
	}

	if err := validation.ValidatePositiveDuration(o.ExpectContinueTimeout, "HTTP client expect_continue_timeout"); err != nil {
		return err
	}

	// 验证TLS配置
	if o.TLS != nil {
		if err := o.TLS.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// AddFlags 添加命令行参数
func (o *HTTPClientOptions) AddFlags(fs *pflag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}

	fs.StringVar(&o.Network, prefix+"http.network", o.Network,
		"Network type for HTTP client (tcp, tcp4, tcp6, unix)")
	fs.StringVar(&o.Addr, prefix+"http.addr", o.Addr,
		"HTTP server address (host:port)")
	fs.DurationVar(&o.Timeout, prefix+"http.timeout", o.Timeout,
		"HTTP client request timeout")
	fs.IntVar(&o.MaxIdleConns, prefix+"http.max-idle-conns", o.MaxIdleConns,
		"Maximum number of idle connections")
	fs.IntVar(&o.MaxIdleConnsPerHost, prefix+"http.max-idle-conns-per-host", o.MaxIdleConnsPerHost,
		"Maximum idle connections per host")
	fs.DurationVar(&o.IdleConnTimeout, prefix+"http.idle-conn-timeout", o.IdleConnTimeout,
		"Idle connection timeout")
	fs.DurationVar(&o.TLSHandshakeTimeout, prefix+"http.tls-handshake-timeout", o.TLSHandshakeTimeout,
		"TLS handshake timeout")
	fs.DurationVar(&o.ExpectContinueTimeout, prefix+"http.expect-continue-timeout", o.ExpectContinueTimeout,
		"Expect 100-continue timeout")
	fs.IntVar(&o.MaxConnsPerHost, prefix+"http.max-conns-per-host", o.MaxConnsPerHost,
		"Maximum connections per host (0 = unlimited)")

	// 添加TLS相关参数
	if o.TLS != nil {
		o.TLS.AddFlags(fs, prefix+"http")
	}
}

// Complete 完成配置初始化
func (o *HTTPClientOptions) Complete() error {
	// 确保网络类型有效
	if o.Network == "" {
		o.Network = "tcp"
	}

	// 确保超时时间合理
	if o.Timeout <= 0 {
		o.Timeout = 30 * time.Second
	}

	if o.IdleConnTimeout <= 0 {
		o.IdleConnTimeout = 90 * time.Second
	}

	if o.TLSHandshakeTimeout <= 0 {
		o.TLSHandshakeTimeout = 10 * time.Second
	}

	if o.ExpectContinueTimeout <= 0 {
		o.ExpectContinueTimeout = 1 * time.Second
	}

	// 确保连接池配置合理
	if o.MaxIdleConns <= 0 {
		o.MaxIdleConns = 100
	}

	if o.MaxIdleConnsPerHost <= 0 {
		o.MaxIdleConnsPerHost = 10
	}

	// 完成TLS配置
	if o.TLS != nil {
		if err := o.TLS.Complete(); err != nil {
			return err
		}
	}

	return nil
}

// ApplyTo 将配置应用到目标接口
func (o *HTTPClientOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	switch v := target.(type) {
	case *[]interface{}:
		config := map[string]interface{}{
			"network":               o.Network,
			"addr":                  o.Addr,
			"timeout":               o.Timeout,
			"maxIdleConns":          o.MaxIdleConns,
			"maxIdleConnsPerHost":   o.MaxIdleConnsPerHost,
			"idleConnTimeout":       o.IdleConnTimeout,
			"tlsHandshakeTimeout":   o.TLSHandshakeTimeout,
			"expectContinueTimeout": o.ExpectContinueTimeout,
			"maxConnsPerHost":       o.MaxConnsPerHost,
		}

		// 添加TLS配置
		if o.TLS != nil && o.TLS.UseTLS {
			var tlsConfigs []interface{}
			_ = o.TLS.ApplyTo(&tlsConfigs)
			if len(tlsConfigs) > 0 {
				config["tls"] = tlsConfigs[0]
			}
		}

		*v = append(*v, config)
	}

	return nil
}

// WithHTTPNetwork 设置网络类型
func WithHTTPNetwork(network string) func(*HTTPClientOptions) {
	return func(o *HTTPClientOptions) {
		o.Network = network
	}
}

// WithHTTPAddr 设置服务器地址
func WithHTTPAddr(addr string) func(*HTTPClientOptions) {
	return func(o *HTTPClientOptions) {
		o.Addr = addr
	}
}

// WithHTTPTimeout 设置请求超时时间
func WithHTTPTimeout(timeout time.Duration) func(*HTTPClientOptions) {
	return func(o *HTTPClientOptions) {
		o.Timeout = timeout
	}
}

// WithHTTPMaxIdleConns 设置最大空闲连接数
func WithHTTPMaxIdleConns(max int) func(*HTTPClientOptions) {
	return func(o *HTTPClientOptions) {
		o.MaxIdleConns = max
	}
}

// WithHTTPMaxIdleConnsPerHost 设置每个主机的最大空闲连接数
func WithHTTPMaxIdleConnsPerHost(max int) func(*HTTPClientOptions) {
	return func(o *HTTPClientOptions) {
		o.MaxIdleConnsPerHost = max
	}
}

// WithHTTPIdleConnTimeout 设置空闲连接超时时间
func WithHTTPIdleConnTimeout(timeout time.Duration) func(*HTTPClientOptions) {
	return func(o *HTTPClientOptions) {
		o.IdleConnTimeout = timeout
	}
}

// WithHTTPTLS 设置TLS配置
func WithHTTPTLS(tls *TLSOptions) func(*HTTPClientOptions) {
	return func(o *HTTPClientOptions) {
		o.TLS = tls
	}
}
