package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/orchestrator/app/options"
	commonapp "github.com/kart-io/k8s-agent/common/app"
	"github.com/kart-io/k8s-agent/common/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/common/initializers"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	orchestrator "github.com/kart-io/k8s-agent/internal/orchestrator"
	"github.com/kart-io/k8s-agent/internal/orchestrator/initializers"
)

// Execute runs the orchestrator command
func Execute() {
	// 创建配置选项
	opts := options.NewServerOptions()

	// 创建应用实例
	app := &OrchestratorApp{}

	// 使用组合框架运行应用
	commonapp.RunWithRunner(
		opts,
		app,
		commonapp.StandardInitLogger,
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
	*commonapp.StandardBootstrapApplication // 嵌入标准 Bootstrap 应用

	config *orchestrator.Config

	// 组件初始化器
	dbInit       *initializers.DatabaseInitializer
	redisInit    *initializers.RedisInitializer
	natsInit     *initializers.NATSInitializer
	workflowInit *initializers.WorkflowInitializer
	strategyInit *initializers.StrategyInitializer
	subInit      *initializers.SubscriberInitializer
	grpcInit     *initializers.GRPCServerInitializer
	httpInit     *initializers.HTTPServerInitializer
	healthInit   *pkginitializers.HealthCheckInitializer
}

// Initialize 初始化应用程序
func (a *OrchestratorApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	// 转换配置
	serverOpts := opts.(*options.ServerOptions)
	config, err := serverOpts.Config()
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}
	a.config = config

	// 创建标准 Bootstrap 应用
	if a.StandardBootstrapApplication == nil {
		a.StandardBootstrapApplication = commonapp.NewStandardBootstrapApplication("Orchestrator", a)
	}

	// 调用标准初始化
	return a.StandardBootstrapApplication.Initialize(ctx, opts)
}

// RegisterComponents 实现 ComponentRegistrar 接口，注册所有组件初始化器
func (a *OrchestratorApp) RegisterComponents(bs *bootstrap.Bootstrap) error {
	opts := a.GetOptions().(*options.ServerOptions)

	// 1. Database (优先级 300)
	a.dbInit = initializers.NewDatabaseInitializer(opts, a.GetLogger())
	bs.Register(a.dbInit)

	// 2. Redis (优先级 400)
	a.redisInit = initializers.NewRedisInitializer(opts, a.GetLogger())
	bs.Register(a.redisInit)

	// 3. NATS (优先级 500)
	a.natsInit = initializers.NewNATSInitializer(opts, a.GetLogger())
	bs.Register(a.natsInit)

	// 4. Workflow Engine (优先级 550 - 在 Database 和 Redis 之后)
	a.workflowInit = initializers.NewWorkflowInitializer(
		opts,
		a.GetLogger(),
		a.dbInit,
		a.redisInit,
	)
	bs.Register(a.workflowInit)

	// 5. Strategy Manager (优先级 600 - 在 Workflow 之后)
	a.strategyInit = initializers.NewStrategyInitializer(
		opts,
		a.GetLogger(),
		a.dbInit,
		a.workflowInit,
	)
	bs.Register(a.strategyInit)

	// 6. Subscriber (优先级 650 - 在 Strategy 之后)
	a.subInit = initializers.NewSubscriberInitializer(
		opts,
		a.GetLogger(),
		a.natsInit,
		a.strategyInit,
	)
	bs.Register(a.subInit)

	// 7. gRPC Server (优先级 700 - 在 Workflow 和 Strategy 之后)
	a.grpcInit = initializers.NewGRPCServerInitializer(
		opts,
		a.GetLogger(),
		a.workflowInit,
		a.dbInit,
	)
	bs.Register(a.grpcInit)

	// 8. HTTP Server with gRPC-Gateway (优先级 800 - 在 gRPC 之后)
	// HTTP requests will be automatically converted to gRPC calls using the same workflow service!
	a.httpInit = initializers.NewHTTPServerInitializer(
		opts,
		a.GetLogger(),
		a.grpcInit, // Pass gRPC init to get shared service
	)
	bs.Register(a.httpInit)

	// 9. Health Check Server (优先级最低，最后启动)
	healthOpts := commonoptions.NewHealthOptions()
	healthOpts.Port = opts.GetHealthPort()
	a.healthInit = pkginitializers.NewHealthCheckInitializer(healthOpts, a.GetLogger())
	bs.Register(a.healthInit)

	return nil
}

// Run/Shutdown/initLogger 方法已由 StandardBootstrapApplication 提供，无需重复定义
