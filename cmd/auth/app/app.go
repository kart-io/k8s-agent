package app

import (
	"context"
	"fmt"

	commonlogger "github.com/kart-io/k8s-agent/common/logger"
	authconfig "github.com/kart-io/k8s-agent/internal/auth/config"
	"github.com/kart-io/k8s-agent/internal/auth/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
	"github.com/spf13/pflag"
)

// Execute runs the auth service command
func Execute() {
	// 创建配置选项 (使用auth service的config)
	opts := NewConfigOptions()

	// 使用组合框架运行应用
	commonapp.RunWithRunner(
		opts,
		&AuthApp{},
		initLogger,
		commonapp.CommandConfig{
			Use:       "auth",
			Short:     "Authentication Service",
			Long:      "Authentication Service provides user authentication and authorization",
			EnvPrefix: "AUTH",
		},
	)
}

// ConfigOptions 实现 commonapp.Options 接口
type ConfigOptions struct {
	Config     *authconfig.Config
	ConfigFile string
}

// NewConfigOptions 创建新的配置选项
func NewConfigOptions() *ConfigOptions {
	return &ConfigOptions{}
}

// AddFlags 添加命令行标志
func (o *ConfigOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.ConfigFile, "config", "c", "", "Path to configuration file")
}

// Complete 完成配置
func (o *ConfigOptions) Complete() error {
	if o.Config == nil {
		var cfg *authconfig.Config
		var err error

		if o.ConfigFile != "" {
			cfg, err = authconfig.LoadFromPath(o.ConfigFile)
		} else {
			cfg, err = authconfig.Load()
		}

		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		o.Config = cfg
	}
	return nil
}

// Validate 验证配置
func (o *ConfigOptions) Validate() []error {
	var errs []error

	if o.Config == nil {
		errs = append(errs, fmt.Errorf("config is nil"))
		return errs
	}

	// Config validation is done in config.Load()
	return errs
}

// AuthApp 实现 commonapp.Application 接口
type AuthApp struct {
	bootstrap *bootstrap.Bootstrap
	opts      *authconfig.Config
	logger    core.Logger

	// 组件初始化器
	dbInit           *initializers.DatabaseInitializer
	redisInit        *initializers.RedisInitializer
	sessionInit      *initializers.SessionServiceInitializer
	auditInit        *initializers.AuditServiceInitializer
	notificationInit *initializers.NotificationServiceInitializer
	forcedLogoutInit *initializers.ForcedLogoutServiceInitializer
	emailInit        *initializers.EmailClientInitializer
	httpInit         *initializers.HTTPServerInitializer
}

// Initialize 初始化应用程序
func (a *AuthApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	a.opts = opts.(*ConfigOptions).Config

	a.logger.Infow("Initializing Auth Service",
		"host", a.opts.Server.Host,
		"port", a.opts.Server.Port,
	)

	// 创建 bootstrap 实例
	a.bootstrap = bootstrap.New(a.logger)

	// 注册所有组件初始化器
	a.registerComponents()

	// 执行初始化
	if err := a.bootstrap.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize components: %w", err)
	}

	a.logger.Infow("All components initialized successfully")
	return nil
}

// Run 运行应用程序主逻辑
func (a *AuthApp) Run(ctx context.Context) error {
	a.logger.Infow("Auth Service started successfully",
		"address", fmt.Sprintf("%s:%d", a.opts.Server.Host, a.opts.Server.Port),
	)

	// 使用 bootstrap 的 Run 方法,它会等待信号
	return a.bootstrap.Run(ctx, nil)
}

// Shutdown 优雅关闭应用程序
func (a *AuthApp) Shutdown(ctx context.Context) error {
	a.logger.Infow("Shutting down Auth Service")
	return a.bootstrap.Shutdown(ctx)
}

// registerComponents 注册所有组件初始化器
func (a *AuthApp) registerComponents() {
	// 1. Database (优先级 300)
	a.dbInit = initializers.NewDatabaseInitializer(a.opts, a.logger)
	a.bootstrap.Register(a.dbInit)

	// 2. Redis (优先级 400)
	a.redisInit = initializers.NewRedisInitializer(a.opts, a.logger)
	a.bootstrap.Register(a.redisInit)

	// 3. Session Service (优先级 450)
	a.sessionInit = initializers.NewSessionServiceInitializer(
		a.opts,
		a.logger,
		a.dbInit,
		a.redisInit,
	)
	a.bootstrap.Register(a.sessionInit)

	// 4. Email Client (优先级 450)
	a.emailInit = initializers.NewEmailClientInitializer(a.opts, a.logger)
	a.bootstrap.Register(a.emailInit)

	// 5. Audit Service (优先级 460)
	a.auditInit = initializers.NewAuditServiceInitializer(
		a.opts,
		a.logger,
		a.dbInit,
	)
	a.bootstrap.Register(a.auditInit)

	// 6. Notification Service (优先级 470) - 需要 dbInit 和 emailInit
	a.notificationInit = initializers.NewNotificationServiceInitializer(
		a.opts,
		a.logger,
		a.dbInit,
		a.emailInit,
	)
	a.bootstrap.Register(a.notificationInit)

	// 7. Forced Logout Service (优先级 490) - 需要 session, audit, notification
	a.forcedLogoutInit = initializers.NewForcedLogoutServiceInitializer(
		a.opts,
		a.logger,
		a.sessionInit,
		a.auditInit,
		a.notificationInit,
	)
	a.bootstrap.Register(a.forcedLogoutInit)

	// 8. HTTP Server (优先级 600)
	a.httpInit = initializers.NewHTTPServerInitializer(
		a.opts,
		a.logger,
		a.dbInit,
		a.redisInit,
		a.sessionInit,
		a.auditInit,
		a.notificationInit,
		a.forcedLogoutInit,
		a.emailInit,
	)
	a.bootstrap.Register(a.httpInit)
}

// initLogger 初始化日志系统
func initLogger(opts commonapp.Options) (core.Logger, error) {
	cfg := opts.(*ConfigOptions).Config

	// cfg.Logging 已经是 commonoptions.LoggingOptions 类型，直接使用
	return commonlogger.InitFromOptions(&cfg.Logging)
}
