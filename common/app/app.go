package app

import (
	"fmt"
	"os"
	"runtime"

	"github.com/fsnotify/fsnotify"
	"github.com/kart-io/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Options 定义应用程序配置选项接口
type Options interface {
	// Complete 完成配置初始化，设置默认值
	Complete() error
	// Validate 验证配置的有效性
	Validate() []error
	// AddFlags 添加命令行参数
	AddFlags(fs *pflag.FlagSet)
}

// RunFunc 定义应用程序运行函数类型
type RunFunc func(opts Options) error

// HealthCheckFunc 定义健康检查函数类型
type HealthCheckFunc func() error

// CommandConfig 定义命令配置
type CommandConfig struct {
	// Use 命令使用说明
	Use string
	// Short 命令简短描述
	Short string
	// Long 命令详细描述
	Long string
	// EnvPrefix 环境变量前缀（如 "AGENT_MANAGER"）
	EnvPrefix string
	// ConfigFileFlag 配置文件参数名称（默认 "config"）
	ConfigFileFlag string
	// ConfigFileShort 配置文件参数短名称（默认 "c"）
	ConfigFileShort string
}

// App 封装应用程序的配置和运行逻辑
type App struct {
	opts    Options
	runFunc RunFunc
	config  CommandConfig
	cmd     *cobra.Command

	// 可选功能
	healthCheckFunc HealthCheckFunc
	silence         bool
	watch           bool
	printVersion    bool
	printRuntime    bool
}

// AppOption 定义应用程序功能选项
type AppOption func(*App)

// WithHealthCheck 设置健康检查函数
func WithHealthCheck(fn HealthCheckFunc) AppOption {
	return func(app *App) {
		app.healthCheckFunc = fn
	}
}

// WithSilence 启用静默模式（不打印启动和配置信息）
func WithSilence() AppOption {
	return func(app *App) {
		app.silence = true
	}
}

// WithWatch 启用配置文件监听
func WithWatch() AppOption {
	return func(app *App) {
		app.watch = true
	}
}

// WithPrintVersion 启用版本信息打印
func WithPrintVersion() AppOption {
	return func(app *App) {
		app.printVersion = true
	}
}

// WithPrintRuntime 启用运行时信息打印
func WithPrintRuntime() AppOption {
	return func(app *App) {
		app.printRuntime = true
	}
}

// NewApp 创建一个新的应用程序实例
func NewApp(opts Options, runFunc RunFunc, cfg CommandConfig, appOpts ...AppOption) *App {
	app := &App{
		opts:    opts,
		runFunc: runFunc,
		config:  cfg,
	}

	// 应用功能选项
	for _, opt := range appOpts {
		opt(app)
	}

	// 构建 Cobra 命令
	app.buildCommand()

	return app
}

// buildCommand 构建 Cobra 命令
func (app *App) buildCommand() {
	// 设置默认值
	if app.config.ConfigFileFlag == "" {
		app.config.ConfigFileFlag = "config"
	}
	if app.config.ConfigFileShort == "" {
		app.config.ConfigFileShort = "c"
	}

	cmd := &cobra.Command{
		Use:   app.config.Use,
		Short: app.config.Short,
		Long:  app.config.Long,
		// 禁用错误时自动显示使用说明
		SilenceUsage: true,
		// 禁用错误自动打印（我们在 Execute 中处理）
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 检查是否请求了版本信息
			version.PrintAndExitIfRequested()
			return nil
		},
		RunE: app.runCommand,
	}

	// 添加版本标志
	version.AddFlags(cmd.Flags())

	// 添加配置文件参数
	cmd.Flags().StringP(
		app.config.ConfigFileFlag,
		app.config.ConfigFileShort,
		"",
		"Path to config file",
	)
	viper.BindPFlag(app.config.ConfigFileFlag, cmd.Flags().Lookup(app.config.ConfigFileFlag))

	// 添加选项特定的参数
	app.opts.AddFlags(cmd.Flags())

	// 绑定所有参数到 viper
	viper.BindPFlags(cmd.Flags())

	// 配置环境变量
	if app.config.EnvPrefix != "" {
		viper.SetEnvPrefix(app.config.EnvPrefix)
		viper.AutomaticEnv()
	}

	app.cmd = cmd
}

// runCommand 执行命令的核心逻辑
func (app *App) runCommand(cmd *cobra.Command, args []string) error {
	// 1. 加载配置文件
	if cfgFile := viper.GetString(app.config.ConfigFileFlag); cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		// 启用配置文件监听
		if app.watch {
			app.setupConfigWatch()
		}
	}

	// 2. 将配置解析到 options 结构
	if err := viper.Unmarshal(app.opts); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 3. 完成配置初始化（设置默认值、计算派生值）
	if err := app.opts.Complete(); err != nil {
		return fmt.Errorf("failed to complete options: %w", err)
	}

	// 4. 验证配置有效性
	if errs := app.opts.Validate(); len(errs) > 0 {
		return fmt.Errorf("invalid options: %v", errs)
	}

	// 5. 打印启动信息（非静默模式）
	if !app.silence {
		app.printStartupInfo()
	}

	// 6. 执行健康检查
	if app.healthCheckFunc != nil {
		if err := app.healthCheckFunc(); err != nil {
			return fmt.Errorf("health check failed: %w", err)
		}
	}

	// 7. 运行应用程序
	return app.runFunc(app.opts)
}

// setupConfigWatch 设置配置文件监听
func (app *App) setupConfigWatch() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("Config file changed: %s\n", e.Name)
		// 重新解析配置
		if err := viper.Unmarshal(app.opts); err != nil {
			fmt.Printf("Failed to reload config: %v\n", err)
		} else {
			fmt.Println("Config reloaded successfully")
		}
	})
}

// printStartupInfo 打印启动信息
func (app *App) printStartupInfo() {
	// 打印版本信息
	if app.printVersion {
		fmt.Printf("Starting %s %s\n", app.config.Use, version.Get().String())
	}

	// 打印运行时信息
	if app.printRuntime {
		fmt.Printf("Go Version: %s\n", runtime.Version())
		fmt.Printf("GOOS: %s, GOARCH: %s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("NumCPU: %d, GOMAXPROCS: %d\n", runtime.NumCPU(), runtime.GOMAXPROCS(0))
		if gogc := os.Getenv("GOGC"); gogc != "" {
			fmt.Printf("GOGC: %s\n", gogc)
		}
	}

	// 打印配置文件信息
	if cfgFile := viper.ConfigFileUsed(); cfgFile != "" {
		fmt.Printf("Using config file: %s\n", cfgFile)
	}
}

// Run 运行应用程序
func (app *App) Run() {
	if err := app.cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Command 返回底层的 Cobra 命令（用于高级定制）
func (app *App) Command() *cobra.Command {
	return app.cmd
}

// ==================== 向后兼容的便捷函数 ====================

// NewCommand 创建一个新的 Cobra 命令（向后兼容）
func NewCommand(opts Options, runFunc RunFunc, cfg CommandConfig) *cobra.Command {
	app := NewApp(opts, runFunc, cfg)
	return app.Command()
}

// Execute 执行命令（向后兼容）
func Execute(cmd *cobra.Command) {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Run 创建并执行命令的便捷函数（向后兼容）
func Run(opts Options, runFunc RunFunc, cfg CommandConfig) {
	app := NewApp(opts, runFunc, cfg)
	app.Run()
}

// RunWithOptions 创建并执行命令的增强版本（支持功能选项）
func RunWithOptions(opts Options, runFunc RunFunc, cfg CommandConfig, appOpts ...AppOption) {
	app := NewApp(opts, runFunc, cfg, appOpts...)
	app.Run()
}
