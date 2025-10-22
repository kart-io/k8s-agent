package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// NATSOptions NATS 配置选项
type NATSOptions struct {
	url              string
	maxReconnects    int
	reconnectWait    time.Duration
	reconnectBufSize int
}

// NATSOption NATS 配置选项函数
type NATSOption func(*NATSOptions)

// WithNATSURL 设置 NATS 服务器地址
func WithNATSURL(url string) NATSOption {
	return func(o *NATSOptions) {
		o.url = url
	}
}

// WithNATSMaxReconnects 设置最大重连次数
func WithNATSMaxReconnects(n int) NATSOption {
	return func(o *NATSOptions) {
		o.maxReconnects = n
	}
}

// WithNATSReconnectWait 设置重连等待时间
func WithNATSReconnectWait(d time.Duration) NATSOption {
	return func(o *NATSOptions) {
		o.reconnectWait = d
	}
}

// WithNATSReconnectBufSize 设置重连缓冲区大小
func WithNATSReconnectBufSize(size int) NATSOption {
	return func(o *NATSOptions) {
		o.reconnectBufSize = size
	}
}

// defaultNATSOptions 返回默认 NATS 配置
func defaultNATSOptions() *NATSOptions {
	return &NATSOptions{
		url:              "nats://localhost:4222",
		maxReconnects:    10,
		reconnectWait:    2 * time.Second,
		reconnectBufSize: 1024 * 1024, // 1MB
	}
}

// NATSClient NATS 客户端
type NATSClient struct {
	Conn   *nats.Conn
	logger *zap.Logger
}

// NewNATS 创建 NATS 连接
func NewNATS(log *zap.Logger, opts ...NATSOption) (*NATSClient, error) {
	// 应用默认配置
	options := defaultNATSOptions()

	// 应用用户配置
	for _, opt := range opts {
		opt(options)
	}

	natsOpts := []nats.Option{
		nats.MaxReconnects(options.maxReconnects),
		nats.ReconnectWait(options.reconnectWait),
		nats.ReconnectBufSize(options.reconnectBufSize),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				log.Error("NATS disconnected", zap.Error(err))
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Info("NATS reconnected", zap.String("url", nc.ConnectedUrl()))
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Warn("NATS connection closed")
		}),
	}

	conn, err := nats.Connect(options.url, natsOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	log.Info("NATS connected", zap.String("url", options.url))

	return &NATSClient{
		Conn:   conn,
		logger: log.With(zap.String("component", "nats")),
	}, nil
}

// Publish 发布消息
func (c *NATSClient) Publish(subject string, data []byte) error {
	if err := c.Conn.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish to %s: %w", subject, err)
	}
	c.logger.Debug("Published message", zap.String("subject", subject), zap.Int("bytes", len(data)))
	return nil
}

// PublishAsync 异步发布消息
func (c *NATSClient) PublishAsync(subject string, data []byte) error {
	return c.Publish(subject, data)
}

// Subscribe 订阅消息
func (c *NATSClient) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	sub, err := c.Conn.Subscribe(subject, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}
	c.logger.Info("Subscribed to subject", zap.String("subject", subject))
	return sub, nil
}

// QueueSubscribe 队列订阅 (负载均衡)
func (c *NATSClient) QueueSubscribe(subject string, queue string, handler nats.MsgHandler) (*nats.Subscription, error) {
	sub, err := c.Conn.QueueSubscribe(subject, queue, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to queue subscribe to %s: %w", subject, err)
	}
	c.logger.Info("Queue subscribed", zap.String("subject", subject), zap.String("queue", queue))
	return sub, nil
}

// Request 请求-响应模式
func (c *NATSClient) Request(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	msg, err := c.Conn.Request(subject, data, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to request %s: %w", subject, err)
	}
	return msg, nil
}

// Health 检查 NATS 健康
func (c *NATSClient) Health(ctx context.Context) error {
	if !c.Conn.IsConnected() {
		return fmt.Errorf("NATS not connected")
	}
	return nil
}

// Close 关闭连接
func (c *NATSClient) Close() error {
	c.logger.Info("Closing NATS connection")
	c.Conn.Close()
	return nil
}

// Shutdown 实现 ShutdownHandler 接口
func (c *NATSClient) Shutdown(ctx context.Context) error {
	return c.Close()
}
