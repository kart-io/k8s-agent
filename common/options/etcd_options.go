package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"github.com/kart-io/k8s-agent/common/options/validation"
)

// EtcdOptions Etcd配置选项
// 参考 OneX 项目设计,用于配置 Etcd 客户端连接
// Etcd 常用于配置中心、服务发现、分布式锁等场景
type EtcdOptions struct {
	// Enable 是否启用Etcd
	Enable bool `mapstructure:"enable" yaml:"enable" json:"enable"`

	// Endpoints Etcd集群端点列表
	Endpoints []string `mapstructure:"endpoints" yaml:"endpoints" json:"endpoints"`

	// DialTimeout 连接超时时间
	DialTimeout time.Duration `mapstructure:"dial_timeout" yaml:"dial_timeout" json:"dial_timeout"`

	// RequestTimeout 请求超时时间
	RequestTimeout time.Duration `mapstructure:"request_timeout" yaml:"request_timeout" json:"request_timeout"`

	// Username 用户名 (可选)
	Username string `mapstructure:"username" yaml:"username" json:"username"`

	// Password 密码 (可选)
	Password string `mapstructure:"password" yaml:"password" json:"password"`

	// TLS TLS配置
	TLS *TLSOptions `mapstructure:"tls" yaml:"tls" json:"tls"`

	// AutoSyncInterval 自动同步间隔时间 (0表示禁用)
	AutoSyncInterval time.Duration `mapstructure:"auto_sync_interval" yaml:"auto_sync_interval" json:"auto_sync_interval"`

	// KeepAliveTime KeepAlive时间间隔
	KeepAliveTime time.Duration `mapstructure:"keep_alive_time" yaml:"keep_alive_time" json:"keep_alive_time"`

	// KeepAliveTimeout KeepAlive超时时间
	KeepAliveTimeout time.Duration `mapstructure:"keep_alive_timeout" yaml:"keep_alive_timeout" json:"keep_alive_timeout"`

	// MaxCallSendMsgSize 最大发送消息大小 (字节)
	MaxCallSendMsgSize int `mapstructure:"max_call_send_msg_size" yaml:"max_call_send_msg_size" json:"max_call_send_msg_size"`

	// MaxCallRecvMsgSize 最大接收消息大小 (字节)
	MaxCallRecvMsgSize int `mapstructure:"max_call_recv_msg_size" yaml:"max_call_recv_msg_size" json:"max_call_recv_msg_size"`
}

// NewEtcdOptions 创建默认的Etcd配置
func NewEtcdOptions() *EtcdOptions {
	return &EtcdOptions{
		Enable:             false, // 默认关闭
		Endpoints:          []string{"127.0.0.1:2379"},
		DialTimeout:        5 * time.Second,
		RequestTimeout:     5 * time.Second,
		Username:           "",
		Password:           "",
		TLS:                NewTLSOptions(),
		AutoSyncInterval:   0, // 禁用自动同步
		KeepAliveTime:      30 * time.Second,
		KeepAliveTimeout:   10 * time.Second,
		MaxCallSendMsgSize: 2 * 1024 * 1024,  // 2MB
		MaxCallRecvMsgSize: 10 * 1024 * 1024, // 10MB
	}
}

// Validate 验证配置
func (o *EtcdOptions) Validate() error {
	if !o.Enable {
		return nil // Etcd是可选的,如果未启用则跳过验证
	}

	// 验证端点列表
	if len(o.Endpoints) == 0 {
		return fmt.Errorf("etcd endpoints cannot be empty when etcd is enabled")
	}

	// 验证每个端点地址格式
	for i, endpoint := range o.Endpoints {
		if endpoint == "" {
			return fmt.Errorf("etcd endpoint[%d] cannot be empty", i)
		}
		// 端点可以是 host:port 或 http://host:port 或 https://host:port
		// 这里不做严格验证,由 etcd 客户端库处理
	}

	// 验证超时时间
	if err := validation.ValidatePositiveDuration(o.DialTimeout, "etcd dial_timeout"); err != nil {
		return err
	}

	if err := validation.ValidatePositiveDuration(o.RequestTimeout, "etcd request_timeout"); err != nil {
		return err
	}

	// 验证 KeepAlive 配置
	if o.KeepAliveTime > 0 {
		if err := validation.ValidatePositiveDuration(o.KeepAliveTime, "etcd keep_alive_time"); err != nil {
			return err
		}
	}

	if o.KeepAliveTimeout > 0 {
		if err := validation.ValidatePositiveDuration(o.KeepAliveTimeout, "etcd keep_alive_timeout"); err != nil {
			return err
		}
	}

	// 验证消息大小限制
	if err := validation.ValidatePositiveInt(o.MaxCallSendMsgSize, "etcd max_call_send_msg_size"); err != nil {
		return err
	}

	if err := validation.ValidatePositiveInt(o.MaxCallRecvMsgSize, "etcd max_call_recv_msg_size"); err != nil {
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
func (o *EtcdOptions) AddFlags(fs *pflag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}

	fs.BoolVar(&o.Enable, prefix+"etcd.enable", o.Enable,
		"Enable Etcd client")
	fs.StringSliceVar(&o.Endpoints, prefix+"etcd.endpoints", o.Endpoints,
		"Etcd cluster endpoints (e.g., http://127.0.0.1:2379)")
	fs.DurationVar(&o.DialTimeout, prefix+"etcd.dial-timeout", o.DialTimeout,
		"Etcd dial timeout")
	fs.DurationVar(&o.RequestTimeout, prefix+"etcd.request-timeout", o.RequestTimeout,
		"Etcd request timeout")
	fs.StringVar(&o.Username, prefix+"etcd.username", o.Username,
		"Etcd username for authentication")
	fs.StringVar(&o.Password, prefix+"etcd.password", o.Password,
		"Etcd password for authentication")
	fs.DurationVar(&o.AutoSyncInterval, prefix+"etcd.auto-sync-interval", o.AutoSyncInterval,
		"Auto sync interval (0 to disable)")
	fs.DurationVar(&o.KeepAliveTime, prefix+"etcd.keep-alive-time", o.KeepAliveTime,
		"KeepAlive time interval")
	fs.DurationVar(&o.KeepAliveTimeout, prefix+"etcd.keep-alive-timeout", o.KeepAliveTimeout,
		"KeepAlive timeout")
	fs.IntVar(&o.MaxCallSendMsgSize, prefix+"etcd.max-call-send-msg-size", o.MaxCallSendMsgSize,
		"Maximum gRPC send message size in bytes")
	fs.IntVar(&o.MaxCallRecvMsgSize, prefix+"etcd.max-call-recv-msg-size", o.MaxCallRecvMsgSize,
		"Maximum gRPC receive message size in bytes")

	// 添加TLS相关参数
	if o.TLS != nil {
		o.TLS.AddFlags(fs, prefix+"etcd")
	}
}

// Complete 完成配置初始化
func (o *EtcdOptions) Complete() error {
	if !o.Enable {
		return nil
	}

	// 确保端点列表不为空
	if len(o.Endpoints) == 0 {
		o.Endpoints = []string{"127.0.0.1:2379"}
	}

	// 确保超时时间合理
	if o.DialTimeout <= 0 {
		o.DialTimeout = 5 * time.Second
	}

	if o.RequestTimeout <= 0 {
		o.RequestTimeout = 5 * time.Second
	}

	if o.KeepAliveTime <= 0 {
		o.KeepAliveTime = 30 * time.Second
	}

	if o.KeepAliveTimeout <= 0 {
		o.KeepAliveTimeout = 10 * time.Second
	}

	// 确保消息大小限制合理
	if o.MaxCallSendMsgSize <= 0 {
		o.MaxCallSendMsgSize = 2 * 1024 * 1024 // 2MB
	}

	if o.MaxCallRecvMsgSize <= 0 {
		o.MaxCallRecvMsgSize = 10 * 1024 * 1024 // 10MB
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
func (o *EtcdOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	switch v := target.(type) {
	case *[]interface{}:
		config := map[string]interface{}{
			"enable":             o.Enable,
			"endpoints":          o.Endpoints,
			"dialTimeout":        o.DialTimeout,
			"requestTimeout":     o.RequestTimeout,
			"username":           o.Username,
			"password":           o.Password,
			"autoSyncInterval":   o.AutoSyncInterval,
			"keepAliveTime":      o.KeepAliveTime,
			"keepAliveTimeout":   o.KeepAliveTimeout,
			"maxCallSendMsgSize": o.MaxCallSendMsgSize,
			"maxCallRecvMsgSize": o.MaxCallRecvMsgSize,
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

// WithEtcdEnable 设置是否启用Etcd
func WithEtcdEnable(enable bool) func(*EtcdOptions) {
	return func(o *EtcdOptions) {
		o.Enable = enable
	}
}

// WithEtcdEndpoints 设置Etcd端点列表
func WithEtcdEndpoints(endpoints []string) func(*EtcdOptions) {
	return func(o *EtcdOptions) {
		o.Endpoints = endpoints
	}
}

// WithEtcdDialTimeout 设置连接超时时间
func WithEtcdDialTimeout(timeout time.Duration) func(*EtcdOptions) {
	return func(o *EtcdOptions) {
		o.DialTimeout = timeout
	}
}

// WithEtcdRequestTimeout 设置请求超时时间
func WithEtcdRequestTimeout(timeout time.Duration) func(*EtcdOptions) {
	return func(o *EtcdOptions) {
		o.RequestTimeout = timeout
	}
}

// WithEtcdUsername 设置用户名
func WithEtcdUsername(username string) func(*EtcdOptions) {
	return func(o *EtcdOptions) {
		o.Username = username
	}
}

// WithEtcdPassword 设置密码
func WithEtcdPassword(password string) func(*EtcdOptions) {
	return func(o *EtcdOptions) {
		o.Password = password
	}
}

// WithEtcdTLS 设置TLS配置
func WithEtcdTLS(tls *TLSOptions) func(*EtcdOptions) {
	return func(o *EtcdOptions) {
		o.TLS = tls
	}
}

// WithEtcdAutoSyncInterval 设置自动同步间隔
func WithEtcdAutoSyncInterval(interval time.Duration) func(*EtcdOptions) {
	return func(o *EtcdOptions) {
		o.AutoSyncInterval = interval
	}
}
