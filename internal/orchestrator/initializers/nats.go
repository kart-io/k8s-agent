package initializers

import (
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"context"

	"github.com/nats-io/nats.go"

	
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
)

// NATSInitializer NATS初始化器.
type NATSInitializer struct {
	opts   *commonapp.StandardOptions
	logger core.Logger
	conn   *nats.Conn
}

// NewNATSInitializer 创建NATS初始化器.
func NewNATSInitializer(opts *commonapp.StandardOptions, logger core.Logger) *NATSInitializer {
	return &NATSInitializer{
		opts:   opts,
		logger: logger,
	}
}

// Name 返回初始化器名称.
func (n *NATSInitializer) Name() string {
	return "nats"
}

// Priority 返回初始化优先级.
func (n *NATSInitializer) Priority() int {
	return bootstrap.PriorityMQ
}

// Initialize 执行初始化.
func (n *NATSInitializer) Initialize(ctx context.Context) error {
	n.logger.Infow("Connecting to NATS",
		"url", n.opts.NATS.URL,
	)

	conn, err := nats.Connect(
		n.opts.NATS.URL,
		nats.Name("orchestrator-service"),
		nats.MaxReconnects(n.opts.NATS.MaxReconnect),
		nats.ReconnectWait(n.opts.NATS.ReconnectWait),
	)
	if err != nil {
		return err
	}

	n.conn = conn
	n.logger.Infow("NATS connected successfully",
		"server_url", n.conn.ConnectedUrl(),
	)
	return nil
}

// Close 关闭NATS连接.
func (n *NATSInitializer) Close(ctx context.Context) error {
	if n.conn != nil {
		n.conn.Close()
	}
	return nil
}

// HealthCheck 检查NATS健康状态.
func (n *NATSInitializer) HealthCheck(ctx context.Context) error {
	if n.conn == nil {
		return nil
	}
	if !n.conn.IsConnected() {
		return nats.ErrDisconnected
	}
	return nil
}

// Conn 获取NATS连接.
func (n *NATSInitializer) Conn() *nats.Conn {
	return n.conn
}
