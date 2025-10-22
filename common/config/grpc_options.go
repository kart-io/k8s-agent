package config

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

// GRPCOptions gRPC服务器配置
type GRPCOptions struct {
	// Enable 是否启用gRPC服务器
	Enable bool `mapstructure:"enable" yaml:"enable" json:"enable"`

	// Host gRPC服务器监听地址
	Host string `mapstructure:"host" yaml:"host" json:"host"`

	// Port gRPC服务器监听端口
	Port int `mapstructure:"port" yaml:"port" json:"port"`

	// MaxRecvMsgSize 最大接收消息大小（字节）
	MaxRecvMsgSize int `mapstructure:"max_recv_msg_size" yaml:"max_recv_msg_size" json:"max_recv_msg_size"`

	// MaxSendMsgSize 最大发送消息大小（字节）
	MaxSendMsgSize int `mapstructure:"max_send_msg_size" yaml:"max_send_msg_size" json:"max_send_msg_size"`

	// ConnectionTimeout 连接超时时间
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout" yaml:"connection_timeout" json:"connection_timeout"`

	// KeepAliveTime KeepAlive时间间隔
	KeepAliveTime time.Duration `mapstructure:"keep_alive_time" yaml:"keep_alive_time" json:"keep_alive_time"`

	// KeepAliveTimeout KeepAlive超时时间
	KeepAliveTimeout time.Duration `mapstructure:"keep_alive_timeout" yaml:"keep_alive_timeout" json:"keep_alive_timeout"`

	// MaxConnectionIdle 最大连接空闲时间
	MaxConnectionIdle time.Duration `mapstructure:"max_connection_idle" yaml:"max_connection_idle" json:"max_connection_idle"`

	// MaxConnectionAge 最大连接存活时间
	MaxConnectionAge time.Duration `mapstructure:"max_connection_age" yaml:"max_connection_age" json:"max_connection_age"`

	// MaxConnectionAgeGrace 连接关闭宽限时间
	MaxConnectionAgeGrace time.Duration `mapstructure:"max_connection_age_grace" yaml:"max_connection_age_grace" json:"max_connection_age_grace"`

	// EnableReflection 启用gRPC反射服务
	EnableReflection bool `mapstructure:"enable_reflection" yaml:"enable_reflection" json:"enable_reflection"`

	// EnableHealthCheck 启用gRPC健康检查服务
	EnableHealthCheck bool `mapstructure:"enable_health_check" yaml:"enable_health_check" json:"enable_health_check"`

	// Logging gRPC日志配置（可选，如果为nil则使用全局日志配置）
	Logging *LoggingOptions `mapstructure:"logging" yaml:"logging" json:"logging"`
}

// NewGRPCOptions 创建默认的gRPC配置
func NewGRPCOptions() *GRPCOptions {
	return &GRPCOptions{
		Enable:                true,
		Host:                  "0.0.0.0",
		Port:                  9090,
		MaxRecvMsgSize:        10 * 1024 * 1024, // 10MB
		MaxSendMsgSize:        10 * 1024 * 1024, // 10MB
		ConnectionTimeout:     2 * time.Minute,
		KeepAliveTime:         30 * time.Second,
		KeepAliveTimeout:      10 * time.Second,
		MaxConnectionIdle:     5 * time.Minute,
		MaxConnectionAge:      30 * time.Minute,
		MaxConnectionAgeGrace: 5 * time.Second,
		EnableReflection:      true,
		EnableHealthCheck:     true,
	}
}

// Validate 验证配置
func (o *GRPCOptions) Validate() error {
	if !o.Enable {
		return nil // gRPC是可选的，如果未启用则跳过验证
	}

	if o.Port < 1 || o.Port > 65535 {
		return fmt.Errorf("invalid gRPC port: %d", o.Port)
	}

	if o.MaxRecvMsgSize < 0 {
		return fmt.Errorf("invalid max_recv_msg_size: %d", o.MaxRecvMsgSize)
	}

	if o.MaxSendMsgSize < 0 {
		return fmt.Errorf("invalid max_send_msg_size: %d", o.MaxSendMsgSize)
	}

	if o.ConnectionTimeout < 0 {
		return fmt.Errorf("invalid connection_timeout: %s", o.ConnectionTimeout)
	}

	return nil
}

// Complete 填充默认值
func (o *GRPCOptions) Complete() error {
	// 所有字段都有默认值，无需额外处理
	return nil
}

// AddFlags 添加命令行参数
func (o *GRPCOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Enable, "grpc.enable", o.Enable, "Enable gRPC server")
	fs.StringVar(&o.Host, "grpc.host", o.Host, "gRPC server host address")
	fs.IntVar(&o.Port, "grpc.port", o.Port, "gRPC server port")
	fs.IntVar(&o.MaxRecvMsgSize, "grpc.max-recv-msg-size", o.MaxRecvMsgSize, "Maximum message size in bytes the server can receive")
	fs.IntVar(&o.MaxSendMsgSize, "grpc.max-send-msg-size", o.MaxSendMsgSize, "Maximum message size in bytes the server can send")
	fs.DurationVar(&o.ConnectionTimeout, "grpc.connection-timeout", o.ConnectionTimeout, "Connection timeout")
	fs.DurationVar(&o.KeepAliveTime, "grpc.keep-alive-time", o.KeepAliveTime, "Keep-alive time")
	fs.DurationVar(&o.KeepAliveTimeout, "grpc.keep-alive-timeout", o.KeepAliveTimeout, "Keep-alive timeout")
	fs.DurationVar(&o.MaxConnectionIdle, "grpc.max-connection-idle", o.MaxConnectionIdle, "Maximum connection idle time")
	fs.DurationVar(&o.MaxConnectionAge, "grpc.max-connection-age", o.MaxConnectionAge, "Maximum connection age")
	fs.DurationVar(&o.MaxConnectionAgeGrace, "grpc.max-connection-age-grace", o.MaxConnectionAgeGrace, "Maximum connection age grace")
	fs.BoolVar(&o.EnableReflection, "grpc.enable-reflection", o.EnableReflection, "Enable gRPC reflection service")
	fs.BoolVar(&o.EnableHealthCheck, "grpc.enable-health-check", o.EnableHealthCheck, "Enable gRPC health check service")
}
