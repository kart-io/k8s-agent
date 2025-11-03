package initializers

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
	"github.com/nats-io/nats.go"
)

// NATSInitializer 通用 NATS 初始化器
//
// 负责初始化NATS消息队列连接，提供以下功能：
// - NATS 连接初始化
// - 自动重连配置
// - 健康检查
// - 优雅关闭
type NATSInitializer struct {
	opts   *options.NATSOptions
	logger core.Logger
	conn   *nats.Conn
}

// NewNATSInitializer 创建 NATS 初始化器
//
// 参数：
//   - opts: NATS 配置选项
//   - logger: 日志记录器
//
// 返回：
//   - *NATSInitializer: NATS 初始化器实例
func NewNATSInitializer(
	opts *options.NATSOptions,
	logger core.Logger,
) *NATSInitializer {
	return &NATSInitializer{
		opts:   opts,
		logger: logger,
	}
}

// Name 返回初始化器名称
func (n *NATSInitializer) Name() string {
	return "nats"
}

// Priority 返回初始化优先级
//
// NATS 初始化器应该在数据库和缓存之后，业务逻辑之前初始化
func (n *NATSInitializer) Priority() int {
	return bootstrap.PriorityMQ
}

// Initialize 执行初始化
//
// 初始化流程：
//  1. 创建 NATS 连接
//  2. 配置自动重连参数
//  3. 测试连接状态
//
// 参数：
//   - ctx: 上下文对象
//
// 返回：
//   - error: 初始化错误
func (n *NATSInitializer) Initialize(ctx context.Context) error {
	n.logger.Infow("Initializing NATS connection",
		"url", n.opts.URL,
		"max_reconnect", n.opts.MaxReconnect,
	)

	// 准备 NATS 连接选项
	natsOpts := []nats.Option{
		nats.MaxReconnects(n.opts.MaxReconnect),
		nats.ReconnectWait(n.opts.ReconnectWait),
		nats.PingInterval(n.opts.PingInterval),
		nats.MaxPingsOutstanding(n.opts.MaxPingsOut),
		nats.ReconnectBufSize(n.opts.ReconnectBufSize),

		// 重连事件处理
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			if err != nil {
				n.logger.Warnw("NATS disconnected", "error", err)
			} else {
				n.logger.Warnw("NATS disconnected")
			}
		}),
		nats.ReconnectHandler(func(conn *nats.Conn) {
			n.logger.Infow("NATS reconnected", "url", conn.ConnectedUrl())
		}),
		nats.ClosedHandler(func(conn *nats.Conn) {
			n.logger.Warnw("NATS connection closed")
		}),
	}

	// 如果设置了 ClusterID，则添加集群配置
	if n.opts.ClusterID != "" {
		natsOpts = append(natsOpts, nats.Name(n.opts.ClusterID))
	}

	// 连接到 NATS 服务器
	conn, err := nats.Connect(n.opts.URL, natsOpts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	n.conn = conn

	// 记录连接信息
	n.logger.Infow("NATS initialized successfully",
		"url", conn.ConnectedUrl(),
		"server_id", conn.ConnectedServerId(),
	)

	return nil
}

// Close 关闭 NATS 连接
//
// 优雅关闭 NATS 连接，等待所有未完成的消息处理完毕。
//
// 参数：
//   - ctx: 上下文对象（用于超时控制）
//
// 返回：
//   - error: 关闭错误
func (n *NATSInitializer) Close(ctx context.Context) error {
	if n.conn != nil {
		n.logger.Infow("Closing NATS connection")

		// Drain 会等待所有订阅的消息处理完毕
		if err := n.conn.Drain(); err != nil {
			n.logger.Warnw("Failed to drain NATS connection", "error", err)
			// 即使 Drain 失败，也要执行 Close
		}

		// 关闭连接
		n.conn.Close()
		n.logger.Infow("NATS connection closed")
	}
	return nil
}

// HealthCheck 检查 NATS 健康状态
//
// 检查 NATS 连接是否正常。
//
// 参数：
//   - ctx: 上下文对象（用于超时控制）
//
// 返回：
//   - error: 健康检查错误
func (n *NATSInitializer) HealthCheck(ctx context.Context) error {
	if n.conn == nil {
		return fmt.Errorf("NATS not initialized")
	}

	// 检查连接状态
	if !n.conn.IsConnected() {
		return fmt.Errorf("NATS disconnected")
	}

	// 检查是否已关闭
	if n.conn.IsClosed() {
		return fmt.Errorf("NATS connection closed")
	}

	return nil
}

// Connection 获取 NATS 连接
//
// 返回 NATS 连接实例，供业务逻辑使用。
//
// 返回：
//   - *nats.Conn: NATS 连接实例
func (n *NATSInitializer) Connection() *nats.Conn {
	return n.conn
}

// Conn 获取 NATS 连接（别名方法）
//
// 与 Connection() 功能相同，提供更简短的方法名。
//
// 返回：
//   - *nats.Conn: NATS 连接实例
func (n *NATSInitializer) Conn() *nats.Conn {
	return n.conn
}

// Publish 发布消息到指定主题
//
// 这是一个便捷方法，直接使用底层连接发布消息。
//
// 参数：
//   - subject: 主题名称
//   - data: 消息数据
//
// 返回：
//   - error: 发布错误
func (n *NATSInitializer) Publish(subject string, data []byte) error {
	if n.conn == nil {
		return fmt.Errorf("NATS not initialized")
	}
	return n.conn.Publish(subject, data)
}

// Subscribe 订阅主题
//
// 这是一个便捷方法，直接使用底层连接订阅主题。
//
// 参数：
//   - subject: 主题名称
//   - handler: 消息处理函数
//
// 返回：
//   - *nats.Subscription: 订阅实例
//   - error: 订阅错误
func (n *NATSInitializer) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	if n.conn == nil {
		return nil, fmt.Errorf("NATS not initialized")
	}
	return n.conn.Subscribe(subject, handler)
}
