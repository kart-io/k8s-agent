package initializers

import (
	"context"

	"github.com/kart-io/k8s-agent/cmd/orchestrator/app/options"
	"github.com/kart-io/k8s-agent/internal/orchestrator/subscriber"
	"github.com/kart-io/logger/core"
)

// SubscriberInitializer 事件订阅器初始化器
type SubscriberInitializer struct {
	opts         *options.ServerOptions
	logger       core.Logger
	natsInit     *NATSInitializer
	strategyInit *StrategyInitializer
	subscriber   *subscriber.Subscriber
}

// NewSubscriberInitializer 创建事件订阅器初始化器
func NewSubscriberInitializer(
	opts *options.ServerOptions,
	logger core.Logger,
	natsInit *NATSInitializer,
	strategyInit *StrategyInitializer,
) *SubscriberInitializer {
	return &SubscriberInitializer{
		opts:         opts,
		logger:       logger,
		natsInit:     natsInit,
		strategyInit: strategyInit,
	}
}

// Name 返回初始化器名称
func (s *SubscriberInitializer) Name() string {
	return "event-subscriber"
}

// Priority 返回初始化优先级
func (s *SubscriberInitializer) Priority() int {
	return 650 // 在 Strategy (600) 之后
}

// Initialize 执行初始化
func (s *SubscriberInitializer) Initialize(ctx context.Context) error {
	s.logger.Info("Initializing event subscriber")

	s.subscriber = subscriber.NewSubscriber(
		s.natsInit.Conn(),
		s.strategyInit.Manager(),
		s.logger,
	)

	if err := s.subscriber.Start(ctx); err != nil {
		return err
	}

	s.logger.Info("Event subscriber started successfully")
	s.logger.Info("Listening for events on NATS channels:")
	s.logger.Info("  - internal.event.critical")
	s.logger.Info("  - internal.event.anomaly")
	s.logger.Info("  - internal.event.* (debug)")
	return nil
}

// Close 关闭事件订阅器
func (s *SubscriberInitializer) Close(ctx context.Context) error {
	if s.subscriber != nil {
		s.subscriber.Stop()
	}
	return nil
}

// HealthCheck 检查事件订阅器健康状态
func (s *SubscriberInitializer) HealthCheck(ctx context.Context) error {
	// Subscriber health is checked via NATS connection
	return nil
}

// Subscriber 获取订阅器实例
func (s *SubscriberInitializer) Subscriber() *subscriber.Subscriber {
	return s.subscriber
}
