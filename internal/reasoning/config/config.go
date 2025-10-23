package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/spf13/viper"
)

// Type aliases for backward compatibility
type (
	LLMConfig         = commonoptions.LLMOptions
	LLMProviderConfig = commonoptions.LLMProviderConfig
	MemoryConfig      = commonoptions.MemoryOptions
	AnalysisConfig    = commonoptions.AnalysisOptions
	PredictionConfig  = commonoptions.PredictionOptions
	LearningConfig    = commonoptions.LearningOptions
	PerformanceConfig = commonoptions.PerformanceOptions
	ServerConfig      = commonoptions.ServerOptions
	LoggingConfig     = commonoptions.LoggingOptions
	AnomalyConfig     = commonoptions.AnomalyOptions
)

// Config represents the application configuration
type Config struct {
	Server      commonoptions.ServerOptions      `mapstructure:"server"`
	LLM         commonoptions.LLMOptions         `mapstructure:"llm"`
	Analysis    commonoptions.AnalysisOptions    `mapstructure:"analysis"`
	Prediction  commonoptions.PredictionOptions  `mapstructure:"prediction"`
	Learning    commonoptions.LearningOptions    `mapstructure:"learning"`
	Performance commonoptions.PerformanceOptions `mapstructure:"performance"`
	Logging     commonoptions.LoggingOptions     `mapstructure:"logging"`
	Memory      commonoptions.MemoryOptions      `mapstructure:"memory"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	return LoadFromPath("")
}

// LoadFromPath loads configuration from a specific file path
func LoadFromPath(configPath string) (*Config, error) {
	v := viper.New()

	if configPath != "" {
		// Use specified config file path
		v.SetConfigFile(configPath)
	} else {
		// Use default config file search
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	// Read configuration file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Allow environment variable overrides
	v.AutomaticEnv()

	// Apply environment variable overrides before unmarshaling
	applyEnvOverrides(v)

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Apply LLM API key overrides from environment variables
	applyLLMEnvOverrides(&config)

	// Validate configuration
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// applyEnvOverrides applies environment variable overrides using viper
func applyEnvOverrides(v *viper.Viper) {
	// Bind specific environment variables
	v.BindEnv("server.host", "SERVER_HOST")
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.log_level", "SERVER_LOG_LEVEL")

	// Bind LLM API keys - these will be applied after unmarshaling
	// Note: We handle LLM provider API keys in a post-processing step
	// since they require iteration over the providers array
}

// applyLLMEnvOverrides applies LLM API key overrides from environment variables
func applyLLMEnvOverrides(config *Config) {
	// LLM API keys from environment
	for i := range config.LLM.Providers {
		provider := &config.LLM.Providers[i]

		// Skip if API key is already set in config file
		if provider.APIKey != "" {
			continue
		}

		// Try to get API key from environment variables
		// Priority: specific env var > uppercase provider name
		var apiKey string

		switch provider.Name {
		case "openai":
			apiKey = os.Getenv("OPENAI_API_KEY")
		case "gemini":
			apiKey = os.Getenv("GEMINI_API_KEY")
			if apiKey == "" {
				apiKey = os.Getenv("GOOGLE_API_KEY")
			}
		case "deepseek":
			apiKey = os.Getenv("DEEPSEEK_API_KEY")
		case "siliconflow":
			apiKey = os.Getenv("SILICONFLOW_API_KEY")
		case "kimi":
			apiKey = os.Getenv("KIMI_API_KEY")
			if apiKey == "" {
				apiKey = os.Getenv("MOONSHOT_API_KEY")
			}
		case "custom":
			apiKey = os.Getenv("CUSTOM_LLM_API_KEY")
			// Also support base URL override for custom provider
			if baseURL := os.Getenv("CUSTOM_LLM_BASE_URL"); baseURL != "" {
				provider.BaseURL = baseURL
			}
			// Support model override for custom provider
			if model := os.Getenv("CUSTOM_LLM_MODEL"); model != "" {
				provider.Model = model
			}
		case "ollama":
			// Ollama typically doesn't require API key (local deployment)
			apiKey = os.Getenv("OLLAMA_API_KEY")
		default:
			// For any other provider, try uppercase version of provider name
			// e.g., "myapi" -> "MYAPI_API_KEY"
			envKey := fmt.Sprintf("%s_API_KEY", strings.ToUpper(provider.Name))
			apiKey = os.Getenv(envKey)
		}

		// Set API key if found in environment
		if apiKey != "" {
			provider.APIKey = apiKey
		}
	}
}

// validate validates the configuration
func validate(config *Config) error {
	if err := config.Server.Validate(); err != nil {
		return fmt.Errorf("server validation failed: %w", err)
	}

	if err := config.LLM.Validate(); err != nil {
		return fmt.Errorf("llm validation failed: %w", err)
	}

	if err := config.Analysis.Validate(); err != nil {
		return fmt.Errorf("analysis validation failed: %w", err)
	}

	if err := config.Prediction.Validate(); err != nil {
		return fmt.Errorf("prediction validation failed: %w", err)
	}

	if err := config.Learning.Validate(); err != nil {
		return fmt.Errorf("learning validation failed: %w", err)
	}

	if err := config.Performance.Validate(); err != nil {
		return fmt.Errorf("performance validation failed: %w", err)
	}

	if err := config.Logging.Validate(); err != nil {
		return fmt.Errorf("logging validation failed: %w", err)
	}

	if err := config.Memory.Validate(); err != nil {
		return fmt.Errorf("memory validation failed: %w", err)
	}

	// 额外的环境变量警告（针对 LLM providers）
	if config.LLM.Enabled {
		for _, provider := range config.LLM.Providers {
			if provider.APIKey == "" {
				fmt.Fprintf(os.Stderr, "Warning: LLM provider %s has no API key, will be skipped\n", provider.Name)
			}
		}
	}

	return nil
}

// GetRequestTimeout returns the request timeout duration
func (c *Config) GetRequestTimeout() time.Duration {
	return c.Performance.GetRequestTimeoutDuration()
}

// GetAccuracyUpdateInterval returns the accuracy update interval duration
func (c *Config) GetAccuracyUpdateInterval() time.Duration {
	duration, err := time.ParseDuration(c.Learning.AccuracyUpdateInterval)
	if err != nil {
		return 1 * time.Hour // Default
	}
	return duration
}
