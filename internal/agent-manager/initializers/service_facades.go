package initializers

import (
	"context"

	"github.com/kart-io/k8s-agent/internal/agent-manager/agent"
	"github.com/kart-io/k8s-agent/internal/agent-manager/command"
	"github.com/kart-io/k8s-agent/internal/agent-manager/event"
	"github.com/kart-io/k8s-agent/internal/agent-manager/nats"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
)

// RegistryInitializer Agent Registry 初始化器.
// Refactored to use ServiceInitializer instead of creating Registry internally.
// This initializer now only provides delegation methods for backward compatibility.
type RegistryInitializer struct {
	opts        *commonapp.StandardOptions
	logger      core.Logger
	serviceInit *ServiceInitializer
}

// NewRegistryInitializer 创建 Registry 初始化器.
// Now depends on ServiceInitializer to get Registry.
func NewRegistryInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	serviceInit *ServiceInitializer,
) *RegistryInitializer {
	return &RegistryInitializer{
		opts:        opts,
		logger:      logger,
		serviceInit: serviceInit,
	}
}

// Name 返回初始化器名称.
func (r *RegistryInitializer) Name() string {
	return "registry"
}

// Priority 返回初始化优先级 (在 Database 和 Redis 之后).
func (r *RegistryInitializer) Priority() int {
	return bootstrap.PriorityCache + 50
}

// Initialize 执行初始化.
// Registry is now created in ServiceInitializer, so this is a no-op.
func (r *RegistryInitializer) Initialize(ctx context.Context) error {
	r.logger.Infow("RegistryInitializer: delegating to ServiceInitializer (backward compatibility)")
	// Registry is created and started in ServiceInitializer.Initialize()
	return nil
}

// Close 关闭 Registry.
// Delegated to ServiceInitializer.
func (r *RegistryInitializer) Close(ctx context.Context) error {
	// Registry cleanup is handled in ServiceInitializer.Close()
	return nil
}

// Registry 获取 Registry 实例.
// Delegated to ServiceInitializer.
func (r *RegistryInitializer) Registry() *agent.Registry {
	return r.serviceInit.Registry()
}

// NATSInitializer NATS 服务器初始化器.
// Refactored to use ServiceInitializer instead of creating EventProcessor internally.
type NATSInitializer struct {
	opts        *commonapp.StandardOptions
	logger      core.Logger
	serviceInit *ServiceInitializer
	natsServer  *nats.Server
}

// NewNATSInitializer 创建 NATS 初始化器.
// Now depends on ServiceInitializer to get Registry and EventProcessor.
func NewNATSInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	serviceInit *ServiceInitializer,
) *NATSInitializer {
	return &NATSInitializer{
		opts:        opts,
		logger:      logger,
		serviceInit: serviceInit,
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
	n.logger.Infow("Initializing NATS server",
		"url", n.opts.NATS.URL,
		"max_reconnect", n.opts.NATS.MaxReconnect,
		"reconnect_delay_initial", n.opts.NATS.ReconnectDelayInitial.String(),
		"reconnect_delay_max", n.opts.NATS.ReconnectDelayMax.String(),
	)

	// Get Registry and EventProcessor from ServiceInitializer
	registry := n.serviceInit.Registry()
	eventProc := n.serviceInit.EventProcessor()

	n.natsServer = nats.NewServer(
		registry,
		eventProc,
		n.logger,
		nats.WithURL(n.opts.NATS.URL),
		nats.WithMaxReconnect(n.opts.NATS.MaxReconnect),
		nats.WithReconnectWait(n.opts.NATS.ReconnectWait),
		nats.WithPingInterval(n.opts.NATS.PingInterval),
		nats.WithMaxPingsOut(n.opts.NATS.MaxPingsOut),
		nats.WithReconnectDelayInitial(n.opts.NATS.ReconnectDelayInitial),
		nats.WithReconnectDelayMax(n.opts.NATS.ReconnectDelayMax),
		nats.WithReconnectBackoffFactor(n.opts.NATS.ReconnectBackoffFactor),
	)

	// 启动 NATS 服务器
	if err := n.natsServer.Start(ctx); err != nil {
		return err
	}

	n.logger.Infow("NATS server initialized successfully")
	return nil
}

// Close 关闭 NATS 服务器.
func (n *NATSInitializer) Close(ctx context.Context) error {
	if n.natsServer != nil {
		n.logger.Infow("Stopping NATS server")
		return n.natsServer.Stop()
	}
	return nil
}

// Server 获取 NATS 服务器实例.
func (n *NATSInitializer) Server() *nats.Server {
	return n.natsServer
}

// EventProcessor 获取事件处理器实例.
// Delegated to ServiceInitializer.
func (n *NATSInitializer) EventProcessor() *event.Processor {
	return n.serviceInit.EventProcessor()
}

// DispatcherInitializer Command Dispatcher 初始化器.
// Refactored to use ServiceInitializer instead of creating Dispatcher internally.
type DispatcherInitializer struct {
	opts        *commonapp.StandardOptions
	logger      core.Logger
	serviceInit *ServiceInitializer
	natsInit    *NATSInitializer
}

// NewDispatcherInitializer 创建 Dispatcher 初始化器.
// Now depends on ServiceInitializer to get Dispatcher.
func NewDispatcherInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	serviceInit *ServiceInitializer,
	natsInit *NATSInitializer,
) *DispatcherInitializer {
	return &DispatcherInitializer{
		opts:        opts,
		logger:      logger,
		serviceInit: serviceInit,
		natsInit:    natsInit,
	}
}

// Name 返回初始化器名称.
func (d *DispatcherInitializer) Name() string {
	return "dispatcher"
}

// Priority 返回初始化优先级 (在 NATS 之后).
func (d *DispatcherInitializer) Priority() int {
	return bootstrap.PriorityMQ + 50
}

// Initialize 执行初始化.
func (d *DispatcherInitializer) Initialize(ctx context.Context) error {
	d.logger.Infow("Initializing command dispatcher")

	// Get Dispatcher from ServiceInitializer
	dispatcher := d.serviceInit.Dispatcher()

	// Wire up command result handler: NATS server calls dispatcher's HandleCommandResult
	d.natsInit.Server().SetCommandResultHandler(dispatcher.HandleCommandResult)

	d.logger.Infow("Command dispatcher initialized successfully with NATS result handler")
	return nil
}

// Close 关闭 Dispatcher (目前无需关闭操作).
func (d *DispatcherInitializer) Close(ctx context.Context) error {
	return nil
}

// Dispatcher 获取 Dispatcher 实例.
// Delegated to ServiceInitializer.
func (d *DispatcherInitializer) Dispatcher() *command.Dispatcher {
	return d.serviceInit.Dispatcher()
}
