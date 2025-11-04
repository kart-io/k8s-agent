package initializers

import (
	"context"

	"github.com/kart-io/k8s-agent/cmd/orchestrator/app/options"
	"github.com/kart-io/k8s-agent/internal/orchestrator/strategy"
	"github.com/kart-io/logger/core"
)

// StrategyInitializer 策略管理器初始化器.
type StrategyInitializer struct {
	opts         *options.ServerOptions
	logger       core.Logger
	dbInit       *DatabaseInitializer
	workflowInit *WorkflowInitializer
	manager      *strategy.Manager
}

// NewStrategyInitializer 创建策略管理器初始化器.
func NewStrategyInitializer(
	opts *options.ServerOptions,
	logger core.Logger,
	dbInit *DatabaseInitializer,
	workflowInit *WorkflowInitializer,
) *StrategyInitializer {
	return &StrategyInitializer{
		opts:         opts,
		logger:       logger,
		dbInit:       dbInit,
		workflowInit: workflowInit,
	}
}

// Name 返回初始化器名称.
func (s *StrategyInitializer) Name() string {
	return "strategy-manager"
}

// Priority 返回初始化优先级.
func (s *StrategyInitializer) Priority() int {
	return 600 // 在 Workflow (550) 之后
}

// Initialize 执行初始化.
func (s *StrategyInitializer) Initialize(ctx context.Context) error {
	s.logger.Info("Initializing strategy manager")

	s.manager = strategy.NewManager(
		s.dbInit.Store(),
		s.workflowInit.Engine(),
		s.logger,
	)

	s.logger.Info("Strategy manager initialized successfully")
	return nil
}

// Close 关闭策略管理器.
func (s *StrategyInitializer) Close(ctx context.Context) error {
	// Strategy manager doesn't need explicit cleanup
	return nil
}

// HealthCheck 检查策略管理器健康状态.
func (s *StrategyInitializer) HealthCheck(ctx context.Context) error {
	// Strategy manager health is checked via database
	return nil
}

// Manager 获取策略管理器实例.
func (s *StrategyInitializer) Manager() *strategy.Manager {
	return s.manager
}
