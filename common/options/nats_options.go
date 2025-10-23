package config

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

// NATSOptions NATS消息队列配置
type NATSOptions struct {
	URL              string        `mapstructure:"url" yaml:"url" json:"url"`
	ClusterID        string        `mapstructure:"cluster_id" yaml:"cluster_id" json:"cluster_id"`
	MaxReconnect     int           `mapstructure:"max_reconnect" yaml:"max_reconnect" json:"max_reconnect"`
	ReconnectWait    time.Duration `mapstructure:"reconnect_wait" yaml:"reconnect_wait" json:"reconnect_wait"`
	PingInterval     time.Duration `mapstructure:"ping_interval" yaml:"ping_interval" json:"ping_interval"`
	MaxPingsOut      int           `mapstructure:"max_pings_out" yaml:"max_pings_out" json:"max_pings_out"`
	EnableJetStream  bool          `mapstructure:"enable_jetstream" yaml:"enable_jetstream" json:"enable_jetstream"`
	ReconnectBufSize int           `mapstructure:"reconnect_buf_size" yaml:"reconnect_buf_size" json:"reconnect_buf_size"`
}

// NewNATSOptions 创建默认的NATS配置
func NewNATSOptions() *NATSOptions {
	return &NATSOptions{
		URL:              "nats://localhost:4222",
		ClusterID:        "",
		MaxReconnect:     10,
		ReconnectWait:    2 * time.Second,
		PingInterval:     20 * time.Second,
		MaxPingsOut:      2,
		EnableJetStream:  false,
		ReconnectBufSize: 1024 * 1024,
	}
}

// Validate 验证配置
func (o *NATSOptions) Validate() error {
	if o.URL == "" {
		return fmt.Errorf("nats url is required")
	}
	if o.MaxReconnect < 0 {
		return fmt.Errorf("max_reconnect must be >= 0")
	}
	if o.MaxPingsOut < 1 {
		return fmt.Errorf("max_pings_out must be > 0")
	}
	return nil
}

// AddFlags 添加命令行参数
func (o *NATSOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.URL, "nats.url", o.URL, "NATS server URL")
	fs.StringVar(&o.ClusterID, "nats.cluster-id", o.ClusterID, "NATS cluster ID")
	fs.IntVar(&o.MaxReconnect, "nats.max-reconnect", o.MaxReconnect, "NATS maximum reconnect attempts")
	fs.DurationVar(&o.ReconnectWait, "nats.reconnect-wait", o.ReconnectWait, "NATS reconnect wait duration")
	fs.DurationVar(&o.PingInterval, "nats.ping-interval", o.PingInterval, "NATS ping interval")
	fs.IntVar(&o.MaxPingsOut, "nats.max-pings-out", o.MaxPingsOut, "NATS maximum outstanding pings")
	fs.BoolVar(&o.EnableJetStream, "nats.enable-jetstream", o.EnableJetStream, "Enable NATS JetStream")
	fs.IntVar(&o.ReconnectBufSize, "nats.reconnect-buf-size", o.ReconnectBufSize, "NATS reconnect buffer size")
}

// ApplyTo 将配置应用到目标接口
// 接受一个函数切片指针，将配置转换为 mq.NATSOption 函数选项
func (o *NATSOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	// 类型断言，检查是否为函数选项切片指针
	switch v := target.(type) {
	case *[]interface{}:
		// 将配置转换为通用选项
		*v = append(*v,
			map[string]interface{}{
				"url":              o.URL,
				"clusterID":        o.ClusterID,
				"maxReconnect":     o.MaxReconnect,
				"reconnectWait":    o.ReconnectWait,
				"pingInterval":     o.PingInterval,
				"maxPingsOut":      o.MaxPingsOut,
				"enableJetStream":  o.EnableJetStream,
				"reconnectBufSize": o.ReconnectBufSize,
			},
		)
	}

	return nil
}

// Complete 完成配置初始化
// 设置默认值和计算派生值
func (o *NATSOptions) Complete() error {
	// 如果 URL 为空，设置默认值
	if o.URL == "" {
		o.URL = "nats://localhost:4222"
	}

	// 确保重连参数合理
	if o.MaxReconnect < 0 {
		o.MaxReconnect = 10
	}

	if o.ReconnectWait <= 0 {
		o.ReconnectWait = 2 * time.Second
	}

	if o.PingInterval <= 0 {
		o.PingInterval = 20 * time.Second
	}

	if o.MaxPingsOut < 1 {
		o.MaxPingsOut = 2
	}

	if o.ReconnectBufSize <= 0 {
		o.ReconnectBufSize = 1024 * 1024 // 1MB
	}

	return nil
}

// WithNATSURL 设置NATS服务器地址
func WithNATSURL(url string) func(*NATSOptions) {
	return func(o *NATSOptions) {
		o.URL = url
	}
}

// WithNATSClusterID 设置NATS集群ID
func WithNATSClusterID(clusterID string) func(*NATSOptions) {
	return func(o *NATSOptions) {
		o.ClusterID = clusterID
	}
}

// WithNATSMaxReconnect 设置最大重连次数
func WithNATSMaxReconnect(max int) func(*NATSOptions) {
	return func(o *NATSOptions) {
		o.MaxReconnect = max
	}
}

// WithNATSReconnectWait 设置重连等待时间
func WithNATSReconnectWait(wait time.Duration) func(*NATSOptions) {
	return func(o *NATSOptions) {
		o.ReconnectWait = wait
	}
}

// WithNATSPingInterval 设置Ping间隔时间
func WithNATSPingInterval(interval time.Duration) func(*NATSOptions) {
	return func(o *NATSOptions) {
		o.PingInterval = interval
	}
}

// WithNATSMaxPingsOut 设置最大未响应Ping数量
func WithNATSMaxPingsOut(max int) func(*NATSOptions) {
	return func(o *NATSOptions) {
		o.MaxPingsOut = max
	}
}

// WithNATSEnableJetStream 启用JetStream
func WithNATSEnableJetStream(enable bool) func(*NATSOptions) {
	return func(o *NATSOptions) {
		o.EnableJetStream = enable
	}
}

// WithNATSReconnectBufSize 设置重连缓冲区大小
func WithNATSReconnectBufSize(size int) func(*NATSOptions) {
	return func(o *NATSOptions) {
		o.ReconnectBufSize = size
	}
}
