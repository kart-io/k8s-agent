package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// LoadConfig 加载配置文件
// configPath: 配置文件路径
// cfg: 配置结构体指针
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
