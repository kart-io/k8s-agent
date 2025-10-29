package initializers

import (
	"context"

	"github.com/kart-io/k8s-agent/cmd/agent-manager/app/options"
	"github.com/kart-io/k8s-agent/internal/agent-manager/agent"
	"github.com/kart-io/k8s-agent/internal/agent-manager/command"
	"github.com/kart-io/k8s-agent/internal/agent-manager/event"
	"github.com/kart-io/k8s-agent/internal/agent-manager/nats"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
)

// RegistryInitializer Agent Registry 初始化器
type RegistryInitializer struct {
	opts      *options.ServerOptions
	logger    core.Logger
	dbInit    *DatabaseInitializer
	redisInit *RedisInitializer
	registry  *agent.Registry
}

// NewRegistryInitializer 创建 Registry 初始化器
func NewRegistryInitializer(
	opts *options.ServerOptions,
	logger core.Logger,
	dbInit *DatabaseInitializer,
	redisInit *RedisInitializer,
) *RegistryInitializer {
	return &RegistryInitializer{
		opts:      opts,
		logger:    logger,
		dbInit:    dbInit,
		redisInit: redisInit,
	}
}

// Name 返回初始化器名称
func (r *RegistryInitializer) Name() string {
	return "registry"
}

// Priority 返回初始化优先级 (在 Database 和 Redis 之后)
func (r *RegistryInitializer) Priority() int {
	return bootstrap.PriorityCache + 50
}

// Initialize 执行初始化
func (r *RegistryInitializer) Initialize(ctx context.Context) error {
	r.logger.Infow("Initializing agent registry")

	r.registry = agent.NewRegistry(
		r.dbInit.Store(),
		r.redisInit.Store(),
		r.logger,
	)

	// 启动 Registry 后台任务
	if err := r.registry.Start(ctx); err != nil {
		return err
	}

	r.logger.Infow("Agent registry initialized successfully")
	return nil
}

// Close 关闭 Registry
func (r *RegistryInitializer) Close(ctx context.Context) error {
	if r.registry != nil {
		r.logger.Infow("Stopping agent registry")
		return r.registry.Stop()
	}
	return nil
}

// Registry 获取 Registry 实例
func (r *RegistryInitializer) Registry() *agent.Registry {
	return r.registry
}

// NATSInitializer NATS 服务器初始化器
type NATSInitializer struct {
	opts       *options.ServerOptions
	logger     core.Logger
	registry   *RegistryInitializer
	eventProc  *event.Processor
	natsServer *nats.Server
}

// NewNATSInitializer 创建 NATS 初始化器
func NewNATSInitializer(
	opts *options.ServerOptions,
	logger core.Logger,
	registry *RegistryInitializer,
	dbInit *DatabaseInitializer,
	redisInit *RedisInitializer,
) *NATSInitializer {
	// 创建 event processor
	eventProc := event.NewProcessor(
		dbInit.Store(),
		redisInit.Store(),
		nil,
		logger,
	)

	return &NATSInitializer{
		opts:      opts,
		logger:    logger,
		registry:  registry,
		eventProc: eventProc,
	}
}

// Name 返回初始化器名称
func (n *NATSInitializer) Name() string {
	return "nats"
}

// Priority 返回初始化优先级
func (n *NATSInitializer) Priority() int {
	return bootstrap.PriorityMQ
}

// Initialize 执行初始化
func (n *NATSInitializer) Initialize(ctx context.Context) error {
	n.logger.Infow("Initializing NATS server",
		"url", n.opts.NATS.URL,
	)

	n.natsServer = nats.NewServer(
		n.registry.Registry(),
		n.eventProc,
		n.logger,
		nats.WithURL(n.opts.NATS.URL),
		nats.WithMaxReconnect(n.opts.NATS.MaxReconnect),
		nats.WithReconnectWait(n.opts.NATS.ReconnectWait),
		nats.WithPingInterval(n.opts.NATS.PingInterval),
		nats.WithMaxPingsOut(n.opts.NATS.MaxPingsOut),
	)

	// 启动 NATS 服务器
	if err := n.natsServer.Start(ctx); err != nil {
		return err
	}

	n.logger.Infow("NATS server initialized successfully")
	return nil
}

// Close 关闭 NATS 服务器
func (n *NATSInitializer) Close(ctx context.Context) error {
	if n.natsServer != nil {
		n.logger.Infow("Stopping NATS server")
		return n.natsServer.Stop()
	}
	return nil
}

// Server 获取 NATS 服务器实例
func (n *NATSInitializer) Server() *nats.Server {
	return n.natsServer
}

// EventProcessor 获取事件处理器实例
func (n *NATSInitializer) EventProcessor() *event.Processor {
	return n.eventProc
}

// DispatcherInitializer Command Dispatcher 初始化器
type DispatcherInitializer struct {
	opts       *options.ServerOptions
	logger     core.Logger
	dbInit     *DatabaseInitializer
	redisInit  *RedisInitializer
	registry   *RegistryInitializer
	natsInit   *NATSInitializer
	dispatcher *command.Dispatcher
}

// NewDispatcherInitializer 创建 Dispatcher 初始化器
func NewDispatcherInitializer(
	opts *options.ServerOptions,
	logger core.Logger,
	dbInit *DatabaseInitializer,
	redisInit *RedisInitializer,
	registry *RegistryInitializer,
	natsInit *NATSInitializer,
) *DispatcherInitializer {
	return &DispatcherInitializer{
		opts:      opts,
		logger:    logger,
		dbInit:    dbInit,
		redisInit: redisInit,
		registry:  registry,
		natsInit:  natsInit,
	}
}

// Name 返回初始化器名称
func (d *DispatcherInitializer) Name() string {
	return "dispatcher"
}

// Priority 返回初始化优先级 (在 NATS 之后)
func (d *DispatcherInitializer) Priority() int {
	return bootstrap.PriorityMQ + 50
}

// Initialize 执行初始化
func (d *DispatcherInitializer) Initialize(ctx context.Context) error {
	d.logger.Infow("Initializing command dispatcher")

	d.dispatcher = command.NewDispatcher(
		d.dbInit.Store(),
		d.redisInit.Store(),
		d.registry.Registry(),
		d.natsInit.Server(),
		d.logger,
	)

	d.logger.Infow("Command dispatcher initialized successfully")
	return nil
}

// Close 关闭 Dispatcher (目前无需关闭操作)
func (d *DispatcherInitializer) Close(ctx context.Context) error {
	return nil
}

// Dispatcher 获取 Dispatcher 实例
func (d *DispatcherInitializer) Dispatcher() *command.Dispatcher {
	return d.dispatcher
}
