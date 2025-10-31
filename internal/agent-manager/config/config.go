package config

import (
	"fmt"

	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/pkg/types"
)

// Load loads configuration from file and environment variables
func Load() (*types.Config, error) {
	return LoadFromPath("")
}

// LoadFromPath loads configuration from a specific file path
func LoadFromPath(configPath string) (*types.Config, error) {
	config := &types.Config{}

	// 使用通用配置加载器
	// 注意：types.Config 需要实现 Complete() 和 Validate() 方法
	// 这里为了兼容，使用一个包装类型
	wrapper := &configWrapper{Config: config}

	envBindings := map[string]string{
		"database.host":     "DB_HOST",
		"database.port":     "DB_PORT",
		"database.user":     "DB_USER",
		"database.password": "DB_PASSWORD",
		"database.database": "DB_NAME",
		"redis.addr":        "REDIS_ADDR",
		"redis.password":    "REDIS_PASSWORD",
		"nats.url":          "NATS_URL",
	}

	if err := commonoptions.LoadOptions(wrapper, configPath, envBindings); err != nil {
		return nil, err
	}

	return config, nil
}

// configWrapper 包装 types.Config 以实现 Options 接口
type configWrapper struct {
	*types.Config
}

// Complete 实现 Options 接口
func (w *configWrapper) Complete() error {
	// agent-manager 的 Config 不需要特殊的 Complete 逻辑
	return nil
}

// Validate 实现 Options 接口
func (w *configWrapper) Validate() []error {
	if err := validate(w.Config); err != nil {
		return []error{err}
	}
	return nil
}

// validate validates configuration
func validate(cfg *types.Config) error {
	if cfg.Server.Port == 0 {
		return fmt.Errorf("server.port is required")
	}
	if cfg.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if cfg.Database.Database == "" {
		return fmt.Errorf("database.database is required")
	}
	if cfg.Redis.Addr == "" {
		return fmt.Errorf("redis.addr is required")
	}
	if cfg.NATS.URL == "" {
		return fmt.Errorf("nats.url is required")
	}
	return nil
}
