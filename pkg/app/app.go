package app

import (
	"fmt"
	"os"

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

// NewCommand 创建一个新的 Cobra 命令
// opts: 配置选项实例
// runFunc: 应用程序运行函数
// cfg: 命令配置
func NewCommand(opts Options, runFunc RunFunc, cfg CommandConfig) *cobra.Command {
	// 设置默认值
	if cfg.ConfigFileFlag == "" {
		cfg.ConfigFileFlag = "config"
	}
	if cfg.ConfigFileShort == "" {
		cfg.ConfigFileShort = "c"
	}

	cmd := &cobra.Command{
		Use:   cfg.Use,
		Short: cfg.Short,
		Long:  cfg.Long,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 检查是否请求了版本信息
			version.PrintAndExitIfRequested()
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. 加载配置文件
			if cfgFile := viper.GetString(cfg.ConfigFileFlag); cfgFile != "" {
				viper.SetConfigFile(cfgFile)
				if err := viper.ReadInConfig(); err != nil {
					return fmt.Errorf("failed to read config file: %w", err)
				}
			}

			// 2. 将配置解析到 options 结构
			if err := viper.Unmarshal(opts); err != nil {
				return fmt.Errorf("failed to unmarshal config: %w", err)
			}

			// 3. 完成配置初始化（设置默认值、计算派生值）
			if err := opts.Complete(); err != nil {
				return fmt.Errorf("failed to complete options: %w", err)
			}

			// 4. 验证配置有效性
			if errs := opts.Validate(); len(errs) > 0 {
				return fmt.Errorf("invalid options: %v", errs)
			}

			// 5. 运行应用程序
			return runFunc(opts)
		},
	}

	// 添加版本标志
	version.AddFlags(cmd.Flags())

	// 添加配置文件参数
	cmd.Flags().StringP(
		cfg.ConfigFileFlag,
		cfg.ConfigFileShort,
		"",
		"Path to config file",
	)
	viper.BindPFlag(cfg.ConfigFileFlag, cmd.Flags().Lookup(cfg.ConfigFileFlag))

	// 添加选项特定的参数
	opts.AddFlags(cmd.Flags())

	// 绑定所有参数到 viper
	viper.BindPFlags(cmd.Flags())

	// 配置环境变量
	if cfg.EnvPrefix != "" {
		viper.SetEnvPrefix(cfg.EnvPrefix)
		viper.AutomaticEnv()
	}

	return cmd
}

// Execute 执行命令
func Execute(cmd *cobra.Command) {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Run 创建并执行命令的便捷函数
func Run(opts Options, runFunc RunFunc, cfg CommandConfig) {
	cmd := NewCommand(opts, runFunc, cfg)
	Execute(cmd)
}
