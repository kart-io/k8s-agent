package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/orchestrator/app/options"
	orchestrator "github.com/kart-io/k8s-agent/internal/orchestrator"
	"github.com/kart-io/k8s-agent/internal/orchestrator/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// Execute runs the orchestrator command
func Execute() {
	// 创建配置选项
	opts := options.NewServerOptions()

	// 使用组合框架运行应用
	commonapp.RunWithRunner(
		opts,
		&OrchestratorApp{},
		initLogger,
		commonapp.CommandConfig{
			Use:       "orchestrator",
			Short:     "Orchestrator Service",
			Long:      "Orchestrator Service manages workflow execution for automated diagnosis and remediation",
			EnvPrefix: "ORCHESTRATOR",
		},
	)
}

// OrchestratorApp 实现 commonapp.Application 接口
type OrchestratorApp struct {
	bootstrap *bootstrap.Bootstrap
	opts      *options.ServerOptions
	config    *orchestrator.Config
	logger    core.Logger

	// 组件初始化器
	dbInit       *initializers.DatabaseInitializer
	redisInit    *initializers.RedisInitializer
	natsInit     *initializers.NATSInitializer
	workflowInit *initializers.WorkflowInitializer
	strategyInit *initializers.StrategyInitializer
	subInit      *initializers.SubscriberInitializer
	healthInit   *pkginitializers.HealthCheckInitializer
}

// Initialize 初始化应用程序
func (a *OrchestratorApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	a.opts = opts.(*options.ServerOptions)

	// 初始化日志系统
	logger, err := initLogger(opts)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Infow("Initializing Orchestrator Service",
		"http_port", a.opts.Server.Port,
		"health_port", a.opts.Health.Port,
		"reasoning_service_url", a.opts.AI.ReasoningServiceURL,
		"agent_manager_url", a.opts.AI.AgentManagerURL,
	)

	// 转换为业务配置
	config, err := a.opts.Config()
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}
	a.config = config

	// 创建 bootstrap 实例
	a.bootstrap = bootstrap.New(a.logger)

	// 注册所有组件初始化器（但不执行初始化）
	// 初始化将在 Run() 方法中由 bootstrap.Run() 执行
	a.registerComponents()

	a.logger.Infow("Components registered, ready to start")
	return nil
}

// Run 运行应用程序主逻辑
func (a *OrchestratorApp) Run(ctx context.Context) error {
	a.logger.Infow("Orchestrator Service started successfully",
		"health_address", fmt.Sprintf(":%d", a.opts.GetHealthPort()),
	)

	// 使用 bootstrap 的 Run 方法,它会等待信号
	return a.bootstrap.Run(ctx, nil)
}

// Shutdown 优雅关闭应用程序
func (a *OrchestratorApp) Shutdown(ctx context.Context) error {
	a.logger.Infow("Shutting down Orchestrator Service")
	return a.bootstrap.Shutdown(ctx)
}

// registerComponents 注册所有组件初始化器
func (a *OrchestratorApp) registerComponents() {
	// 1. Database (优先级 300)
	a.dbInit = initializers.NewDatabaseInitializer(a.opts, a.logger)
	a.bootstrap.Register(a.dbInit)

	// 2. Redis (优先级 400)
	a.redisInit = initializers.NewRedisInitializer(a.opts, a.logger)
	a.bootstrap.Register(a.redisInit)

	// 3. NATS (优先级 500)
	a.natsInit = initializers.NewNATSInitializer(a.opts, a.logger)
	a.bootstrap.Register(a.natsInit)

	// 4. Workflow Engine (优先级 550 - 在 Database 和 Redis 之后)
	a.workflowInit = initializers.NewWorkflowInitializer(
		a.opts,
		a.logger,
		a.dbInit,
		a.redisInit,
	)
	a.bootstrap.Register(a.workflowInit)

	// 5. Strategy Manager (优先级 600 - 在 Workflow 之后)
	a.strategyInit = initializers.NewStrategyInitializer(
		a.opts,
		a.logger,
		a.dbInit,
		a.workflowInit,
	)
	a.bootstrap.Register(a.strategyInit)

	// 6. Subscriber (优先级 650 - 在 Strategy 之后)
	a.subInit = initializers.NewSubscriberInitializer(
		a.opts,
		a.logger,
		a.natsInit,
		a.strategyInit,
	)
	a.bootstrap.Register(a.subInit)

	// 7. Health Check Server (优先级最低，最后启动)
	healthPort := a.opts.GetHealthPort()
	healthAddr := fmt.Sprintf(":%d", healthPort)
	a.healthInit = pkginitializers.NewHealthCheckInitializer(healthAddr, a.logger)
	a.bootstrap.Register(a.healthInit)
}

// initLogger 初始化日志系统
func initLogger(opts commonapp.Options) (core.Logger, error) {
	serverOpts := opts.(*options.ServerOptions)
	return serverOpts.InitLogger()
}
