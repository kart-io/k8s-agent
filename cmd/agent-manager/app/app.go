package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/agent-manager/app/options"
	agentmanager "github.com/kart-io/k8s-agent/internal/agent-manager"
	"github.com/kart-io/k8s-agent/internal/agent-manager/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
)

// Execute runs the agent-manager command
func Execute() {
	// 创建配置选项
	opts := options.NewServerOptions()

	// 创建应用实例
	app := &AgentManagerApp{}

	// 使用组合框架运行应用
	commonapp.RunWithRunner(
		opts,
		app,
		commonapp.StandardInitLogger, // 使用标准 initLogger
		commonapp.CommandConfig{
			Use:       "agent-manager",
			Short:     "Agent Manager Service",
			Long:      "Agent Manager Service manages k8s agents across multiple clusters",
			EnvPrefix: "AGENT_MANAGER",
		},
	)
}

// AgentManagerApp 实现 commonapp.Application 接口
type AgentManagerApp struct {
	*commonapp.StandardBootstrapApplication // 嵌入标准 Bootstrap 应用

	config *agentmanager.Config

	// 组件初始化器
	dbInit         *initializers.DatabaseInitializer
	redisInit      *initializers.RedisInitializer
	registryInit   *initializers.RegistryInitializer
	natsInit       *initializers.NATSInitializer
	dispatcherInit *initializers.DispatcherInitializer
	httpInit       *initializers.HTTPServerInitializer
	grpcInit       *initializers.GRPCServerInitializer
	healthInit     *pkginitializers.HealthCheckInitializer
}

// Initialize 初始化应用程序
func (a *AgentManagerApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	// 转换配置
	serverOpts := opts.(*options.ServerOptions)
	config, err := serverOpts.Config()
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}
	a.config = config

	// 创建标准 Bootstrap 应用（只需一次）
	if a.StandardBootstrapApplication == nil {
		a.StandardBootstrapApplication = commonapp.NewStandardBootstrapApplication("Agent Manager", a)
	}

	// 调用标准初始化
	return a.StandardBootstrapApplication.Initialize(ctx, opts)
}

// RegisterComponents 实现 ComponentRegistrar 接口，注册所有组件初始化器
func (a *AgentManagerApp) RegisterComponents(bs *bootstrap.Bootstrap) error {
	opts := a.GetOptions().(*options.ServerOptions)

	// 1. Database (优先级 300)
	a.dbInit = initializers.NewDatabaseInitializer(opts, a.GetLogger())
	bs.Register(a.dbInit)

	// 2. Redis (优先级 400)
	a.redisInit = initializers.NewRedisInitializer(opts, a.GetLogger())
	bs.Register(a.redisInit)

	// 3. Registry (优先级 450 - 在 Database 和 Redis 之后)
	a.registryInit = initializers.NewRegistryInitializer(
		opts,
		a.GetLogger(),
		a.dbInit,
		a.redisInit,
	)
	bs.Register(a.registryInit)

	// 4. NATS (优先级 500)
	a.natsInit = initializers.NewNATSInitializer(
		opts,
		a.GetLogger(),
		a.registryInit,
		a.dbInit,
		a.redisInit,
	)
	bs.Register(a.natsInit)

	// 5. Dispatcher (优先级 550 - 在 NATS 之后)
	a.dispatcherInit = initializers.NewDispatcherInitializer(
		opts,
		a.GetLogger(),
		a.dbInit,
		a.redisInit,
		a.registryInit,
		a.natsInit,
	)
	bs.Register(a.dispatcherInit)

	// 6. HTTP Server (优先级 600)
	a.httpInit = initializers.NewHTTPServerInitializer(
		opts,
		a.GetLogger(),
		a.registryInit,
		a.dispatcherInit,
		a.dbInit,
		a.redisInit,
		a.natsInit,
	)
	bs.Register(a.httpInit)

	// 7. gRPC Server (优先级 700, 可选)
	if opts.GRPC.Enable {
		a.grpcInit = initializers.NewGRPCServerInitializer(
			opts,
			a.GetLogger(),
			a.registryInit,
			a.dispatcherInit,
			a.dbInit,
		)
		bs.Register(a.grpcInit)
	}

	// 8. Health Check Server (优先级最低，最后启动)
	healthPort := opts.GetHealthPort()
	healthAddr := fmt.Sprintf(":%d", healthPort)
	a.healthInit = pkginitializers.NewHealthCheckInitializer(healthAddr, a.GetLogger())
	bs.Register(a.healthInit)

	return nil
}

// Run/Shutdown/initLogger 方法已由 StandardBootstrapApplication 提供，无需重复定义
