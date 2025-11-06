package initializers

import (
	"context"

	"github.com/kart-io/k8s-agent/cmd/orchestrator/app/options"
	"github.com/kart-io/k8s-agent/internal/orchestrator/workflow"
	"github.com/kart-io/logger/core"
)

// WorkflowInitializer 工作流引擎初始化器.
type WorkflowInitializer struct {
	opts      *options.ServerOptions
	logger    core.Logger
	dbInit    *DatabaseInitializer
	redisInit *RedisInitializer
	engine    *workflow.Engine
}

// NewWorkflowInitializer 创建工作流引擎初始化器.
func NewWorkflowInitializer(
	opts *options.ServerOptions,
	logger core.Logger,
	dbInit *DatabaseInitializer,
	redisInit *RedisInitializer,
) *WorkflowInitializer {
	return &WorkflowInitializer{
		opts:      opts,
		logger:    logger,
		dbInit:    dbInit,
		redisInit: redisInit,
	}
}

// Name 返回初始化器名称.
func (w *WorkflowInitializer) Name() string {
	return "workflow-engine"
}

// Priority 返回初始化优先级.
func (w *WorkflowInitializer) Priority() int {
	return 550 // 在 Database (300) 和 Redis (400) 之后
}

// Initialize 执行初始化.
func (w *WorkflowInitializer) Initialize(ctx context.Context) error {
	w.logger.Infow("Initializing workflow engine",
		"agent_manager_url", w.opts.AI.AgentManagerURL,
		"reasoning_service_url", w.opts.AI.ReasoningServiceURL,
		"global_timeout", w.opts.Workflow.GlobalTimeout,
		"step_default_timeout", w.opts.Workflow.StepDefaultTimeout,
		"retry_on_timeout", w.opts.Workflow.RetryOnTimeout,
		"max_retries", w.opts.Workflow.MaxRetries,
	)

	// 创建执行器
	executor := workflow.NewExecutor(
		w.opts.AI.AgentManagerURL,
		w.opts.AI.ReasoningServiceURL,
		w.logger,
	)

	// 创建工作流引擎
	w.engine = workflow.NewEngine(
		w.dbInit.Store(),
		w.redisInit.Store(),
		executor,
		w.logger,
	)

	// 设置超时配置
	w.engine.SetTimeoutConfig(
		w.opts.Workflow.GlobalTimeout,
		w.opts.Workflow.StepDefaultTimeout,
		w.opts.Workflow.RetryOnTimeout,
		w.opts.Workflow.MaxRetries,
	)

	w.logger.Info("Workflow engine initialized successfully")
	return nil
}

// Close 关闭工作流引擎.
func (w *WorkflowInitializer) Close(ctx context.Context) error {
	// Workflow engine doesn't need explicit cleanup
	return nil
}

// HealthCheck 检查工作流引擎健康状态.
func (w *WorkflowInitializer) HealthCheck(ctx context.Context) error {
	// Workflow engine health is checked via database and redis
	return nil
}

// Engine 获取工作流引擎实例.
func (w *WorkflowInitializer) Engine() *workflow.Engine {
	return w.engine
}
