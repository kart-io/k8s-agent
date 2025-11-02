package commoncore

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// LoadableConfig 定义可加载的配置接口
// 实现此接口的配置可以使用 LoadOptions 和 LoadOptionsWithCallback 函数
type LoadableConfig interface {
	Complete() error
	Validate() []error
}

// LoadConfig 加载配置文件
// configPath: 配置文件路径
// cfg: 配置结构体指针
// Deprecated: 使用 LoadOptions 替代，支持 Complete 和 Validate
func LoadConfig(configPath string, cfg interface{}) error {
	v := viper.New()

	// 设置配置文件路径
	v.SetConfigFile(configPath)

	// 环境变量支持 (支持嵌套字段，如 server.port -> SERVER_PORT)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析到结构体
	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// LoadOptions 加载配置并完成初始化和验证
// opts: 实现了 LoadableConfig 接口的配置结构
// configPath: 配置文件路径（空字符串使用默认搜索）
// envBindings: 自定义环境变量绑定（可选，key 为 viper 配置路径，value 为环境变量名）
func LoadOptions(opts LoadableConfig, configPath string, envBindings map[string]string) error {
	v := viper.New()

	// 设置配置文件
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// 使用默认配置文件搜索
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	// 读取配置文件（如果不存在不报错，使用默认值）
	if err := v.ReadInConfig(); err != nil {
		// 如果配置文件不存在，不报错，继续使用默认值和环境变量
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// 环境变量支持（自动映射，如 server.port -> SERVER_PORT）
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// 自定义环境变量绑定
	for viperKey, envKey := range envBindings {
		if err := v.BindEnv(viperKey, envKey); err != nil {
			return fmt.Errorf("failed to bind env %s: %w", envKey, err)
		}
	}

	// 解析配置到 opts
	if err := v.Unmarshal(opts); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 调用 Complete() 完成配置
	if err := opts.Complete(); err != nil {
		return fmt.Errorf("failed to complete config: %w", err)
	}

	// 调用 Validate() 验证配置
	if errs := opts.Validate(); len(errs) > 0 {
		return fmt.Errorf("config validation failed: %v", errs)
	}

	return nil
}

// PostUnmarshalCallback 定义 Unmarshal 后的回调函数类型
// 用于在 Unmarshal 之后、Complete 之前执行自定义逻辑
type PostUnmarshalCallback func(*viper.Viper, LoadableConfig) error

// LoadOptionsWithCallback 加载配置并支持回调
// 支持在 Unmarshal 之后、Complete 之前执行自定义逻辑
// 适用于需要特殊处理的场景，如 reasoning 服务的 LLM 环境变量覆盖
func LoadOptionsWithCallback(
	opts LoadableConfig,
	configPath string,
	envBindings map[string]string,
	postUnmarshal PostUnmarshalCallback,
) error {
	v := viper.New()

	// 设置配置文件
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// 环境变量支持
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// 自定义环境变量绑定
	for viperKey, envKey := range envBindings {
		if err := v.BindEnv(viperKey, envKey); err != nil {
			return fmt.Errorf("failed to bind env %s: %w", envKey, err)
		}
	}

	// 解析配置
	if err := v.Unmarshal(opts); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 执行回调（如 reasoning 的 applyLLMEnvOverrides）
	if postUnmarshal != nil {
		if err := postUnmarshal(v, opts); err != nil {
			return fmt.Errorf("post-unmarshal callback failed: %w", err)
		}
	}

	// 调用 Complete() 完成配置
	if err := opts.Complete(); err != nil {
		return fmt.Errorf("failed to complete config: %w", err)
	}

	// 调用 Validate() 验证配置
	if errs := opts.Validate(); len(errs) > 0 {
		return fmt.Errorf("config validation failed: %v", errs)
	}

	return nil
}

// LoadConfigWithDefaults 加载配置 (支持默认值)
// configPath: 配置文件路径
// cfg: 配置结构体指针
// defaults: 默认值 map (key 使用点分隔，如 "server.port")
func LoadConfigWithDefaults(configPath string, cfg interface{}, defaults map[string]interface{}) error {
	v := viper.New()

	// 设置默认值
	for key, value := range defaults {
		v.SetDefault(key, value)
	}

	v.SetConfigFile(configPath)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// LoadConfigFromEnv 仅从环境变量加载配置 (不读取文件)
func LoadConfigFromEnv(prefix string, cfg interface{}) error {
	v := viper.New()

	v.SetEnvPrefix(prefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config from env: %w", err)
	}

	return nil
}
