package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/common/logger"
	"github.com/kart-io/k8s-agent/internal/agent-manager/config"
	"github.com/kart-io/k8s-agent/internal/agent-manager/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// Execute runs the agent-manager command
func Execute() {
	// 创建配置选项
	opts := config.NewOptions()

	// 使用组合框架运行应用
	commonapp.RunWithRunner(
		opts,
		&AgentManagerApp{},
		initLogger,
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
	bootstrap *bootstrap.Bootstrap
	opts      *config.Options
	logger    core.Logger

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
	a.opts = opts.(*config.Options)

	// 初始化日志系统
	logger, err := initLogger(opts)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Infow("Initializing Agent Manager Service",
		"http_port", a.opts.Server.Port,
		"grpc_enabled", a.opts.GRPC.Enable,
		"grpc_port", a.opts.GRPC.Port,
	)

	// 创建 bootstrap 实例,直接使用 kart-io/logger
	a.bootstrap = bootstrap.New(a.logger)

	// 注册所有组件初始化器（但不执行初始化）
	// 初始化将在 Run() 方法中由 bootstrap.Run() 执行
	a.registerComponents()

	a.logger.Infow("Components registered, ready to start")
	return nil
}

// Run 运行应用程序主逻辑
func (a *AgentManagerApp) Run(ctx context.Context) error {
	a.logger.Infow("Agent Manager Service started successfully",
		"http_address", fmt.Sprintf("%s:%d", a.opts.Server.Host, a.opts.Server.Port),
		"grpc_enabled", a.opts.GRPC.Enable,
	)

	// 使用 bootstrap 的 Run 方法,它会等待信号
	return a.bootstrap.Run(ctx, nil)
}

// Shutdown 优雅关闭应用程序
func (a *AgentManagerApp) Shutdown(ctx context.Context) error {
	a.logger.Infow("Shutting down Agent Manager Service")
	return a.bootstrap.Shutdown(ctx)
}

// registerComponents 注册所有组件初始化器
func (a *AgentManagerApp) registerComponents() {
	// 1. Database (优先级 300)
	a.dbInit = initializers.NewDatabaseInitializer(a.opts, a.logger)
	a.bootstrap.Register(a.dbInit)

	// 2. Redis (优先级 400)
	a.redisInit = initializers.NewRedisInitializer(a.opts, a.logger)
	a.bootstrap.Register(a.redisInit)

	// 3. Registry (优先级 450 - 在 Database 和 Redis 之后)
	a.registryInit = initializers.NewRegistryInitializer(
		a.opts,
		a.logger,
		a.dbInit,
		a.redisInit,
	)
	a.bootstrap.Register(a.registryInit)

	// 4. NATS (优先级 500)
	a.natsInit = initializers.NewNATSInitializer(
		a.opts,
		a.logger,
		a.registryInit,
		a.dbInit,
		a.redisInit,
	)
	a.bootstrap.Register(a.natsInit)

	// 5. Dispatcher (优先级 550 - 在 NATS 之后)
	a.dispatcherInit = initializers.NewDispatcherInitializer(
		a.opts,
		a.logger,
		a.dbInit,
		a.redisInit,
		a.registryInit,
		a.natsInit,
	)
	a.bootstrap.Register(a.dispatcherInit)

	// 6. HTTP Server (优先级 600)
	a.httpInit = initializers.NewHTTPServerInitializer(
		a.opts,
		a.logger,
		a.registryInit,
		a.dispatcherInit,
		a.dbInit,
		a.redisInit,
		a.natsInit,
	)
	a.bootstrap.Register(a.httpInit)

	// 7. gRPC Server (优先级 700, 可选)
	if a.opts.GRPC.Enable {
		a.grpcInit = initializers.NewGRPCServerInitializer(
			a.opts,
			a.logger,
			a.registryInit,
			a.dispatcherInit,
			a.dbInit,
		)
		a.bootstrap.Register(a.grpcInit)
	}

	// 8. Health Check Server (优先级最低，最后启动)
	healthPort := a.opts.GetHealthPort()
	healthAddr := fmt.Sprintf(":%d", healthPort)
	a.healthInit = pkginitializers.NewHealthCheckInitializer(healthAddr, a.logger)
	a.bootstrap.Register(a.healthInit)
}

// initLogger 初始化日志系统
func initLogger(opts commonapp.Options) (core.Logger, error) {
	return logger.InitFromOptions(opts.(*config.Options).Logging)
}
