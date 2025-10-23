package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// LLMOptions LLM提供商配置
type LLMOptions struct {
	Enabled   bool                `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Providers []LLMProviderConfig `mapstructure:"providers" yaml:"providers" json:"providers"`
}

// LLMProviderConfig 单个LLM提供商配置
type LLMProviderConfig struct {
	Name        string  `mapstructure:"name" yaml:"name" json:"name"`             // "openai", "gemini", "deepseek", "kimi", "siliconflow", "ollama", "custom"
	APIKey      string  `mapstructure:"api_key" yaml:"api_key" json:"api_key"`    // API密钥（可通过环境变量设置）
	BaseURL     string  `mapstructure:"base_url" yaml:"base_url" json:"base_url"` // API基础URL
	Model       string  `mapstructure:"model" yaml:"model" json:"model"`          // 模型名称
	MaxTokens   int     `mapstructure:"max_tokens" yaml:"max_tokens" json:"max_tokens"`
	Temperature float64 `mapstructure:"temperature" yaml:"temperature" json:"temperature"`
	Timeout     int     `mapstructure:"timeout" yaml:"timeout" json:"timeout"`    // 超时时间（秒）
	Priority    int     `mapstructure:"priority" yaml:"priority" json:"priority"` // 优先级（数字越大优先级越高）
}

// NewLLMOptions 创建默认的LLM配置
func NewLLMOptions() *LLMOptions {
	return &LLMOptions{
		Enabled:   false,
		Providers: []LLMProviderConfig{},
	}
}

// Validate 验证配置
func (o *LLMOptions) Validate() error {
	if !o.Enabled {
		return nil // LLM disabled, no validation needed
	}

	if len(o.Providers) == 0 {
		return fmt.Errorf("LLM is enabled but no providers configured")
	}

	for i, provider := range o.Providers {
		if provider.Name == "" {
			return fmt.Errorf("provider %d: name is required", i)
		}
		if provider.MaxTokens < 0 {
			return fmt.Errorf("provider %s: max_tokens must be >= 0", provider.Name)
		}
		if provider.Temperature < 0 || provider.Temperature > 2 {
			return fmt.Errorf("provider %s: temperature must be between 0 and 2", provider.Name)
		}
		if provider.Timeout < 0 {
			return fmt.Errorf("provider %s: timeout must be >= 0", provider.Name)
		}
	}

	return nil
}

// AddFlags 添加命令行参数
func (o *LLMOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Enabled, "llm.enabled", o.Enabled, "Enable LLM integration")
	// Note: Provider details are typically configured via config file, not command-line flags
}

// ApplyTo 将配置应用到目标接口
func (o *LLMOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	switch v := target.(type) {
	case *[]interface{}:
		providers := make([]map[string]interface{}, len(o.Providers))
		for i, p := range o.Providers {
			providers[i] = map[string]interface{}{
				"name":        p.Name,
				"apiKey":      p.APIKey,
				"baseURL":     p.BaseURL,
				"model":       p.Model,
				"maxTokens":   p.MaxTokens,
				"temperature": p.Temperature,
				"timeout":     p.Timeout,
				"priority":    p.Priority,
			}
		}
		*v = append(*v,
			map[string]interface{}{
				"enabled":   o.Enabled,
				"providers": providers,
			},
		)
	}

	return nil
}

// Complete 完成配置初始化
func (o *LLMOptions) Complete() error {
	// 设置默认值
	for i := range o.Providers {
		provider := &o.Providers[i]

		if provider.MaxTokens == 0 {
			provider.MaxTokens = 4096
		}

		if provider.Temperature == 0 {
			provider.Temperature = 0.7
		}

		if provider.Timeout == 0 {
			provider.Timeout = 30
		}

		if provider.Priority == 0 {
			provider.Priority = 1
		}
	}

	return nil
}

// WithLLMEnabled 设置是否启用LLM
func WithLLMEnabled(enabled bool) func(*LLMOptions) {
	return func(o *LLMOptions) {
		o.Enabled = enabled
	}
}

// WithLLMProviders 设置LLM提供商列表
func WithLLMProviders(providers []LLMProviderConfig) func(*LLMOptions) {
	return func(o *LLMOptions) {
		o.Providers = providers
	}
}

// AddProvider 添加LLM提供商
func (o *LLMOptions) AddProvider(provider LLMProviderConfig) {
	o.Providers = append(o.Providers, provider)
}
