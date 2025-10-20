package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	LLM         LLMConfig         `mapstructure:"llm"`
	Analysis    AnalysisConfig    `mapstructure:"analysis"`
	Prediction  PredictionConfig  `mapstructure:"prediction"`
	Learning    LearningConfig    `mapstructure:"learning"`
	Performance PerformanceConfig `mapstructure:"performance"`
	Logging     LoggingConfig     `mapstructure:"logging"`
	Features    FeaturesConfig    `mapstructure:"features"`
	Memory      MemoryConfig      `mapstructure:"memory"` // 新增: Memory 系统配置
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	LogLevel string `mapstructure:"log_level"`
}

// LLMConfig represents LLM provider configuration
type LLMConfig struct {
	Enabled   bool                `mapstructure:"enabled"`
	Providers []LLMProviderConfig `mapstructure:"providers"`
}

// LLMProviderConfig represents a single LLM provider
type LLMProviderConfig struct {
	Name        string  `mapstructure:"name"`    // "openai", "gemini", "deepseek"
	APIKey      string  `mapstructure:"api_key"` // Can be set via env var
	BaseURL     string  `mapstructure:"base_url"`
	Model       string  `mapstructure:"model"`
	MaxTokens   int     `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`
	Timeout     int     `mapstructure:"timeout"`  // in seconds
	Priority    int     `mapstructure:"priority"` // Higher priority providers are tried first
}

// AnalysisConfig represents analysis settings
type AnalysisConfig struct {
	MinConfidence       float64 `mapstructure:"min_confidence"`
	MaxRecommendations  int     `mapstructure:"max_recommendations"`
	IncludeSimilarCases bool    `mapstructure:"include_similar_cases"`
	SimilarityThreshold float64 `mapstructure:"similarity_threshold"`
	UseLLMFallback      bool    `mapstructure:"use_llm_fallback"` // Use LLM if rule-based analysis fails
}

// PredictionConfig represents prediction settings
type PredictionConfig struct {
	TimeWindows      []string      `mapstructure:"time_windows"`
	AnomalyDetection AnomalyConfig `mapstructure:"anomaly_detection"`
}

// AnomalyConfig represents anomaly detection settings
type AnomalyConfig struct {
	Contamination float64 `mapstructure:"contamination"`
	NEstimators   int     `mapstructure:"n_estimators"`
}

// LearningConfig represents learning system settings
type LearningConfig struct {
	EnableFeedback         bool   `mapstructure:"enable_feedback"`
	MinSamplesForAccuracy  int    `mapstructure:"min_samples_for_accuracy"`
	AccuracyUpdateInterval string `mapstructure:"accuracy_update_interval"`
	ExportLearningData     bool   `mapstructure:"export_learning_data"`
	ExportPath             string `mapstructure:"export_path"`
}

// PerformanceConfig represents performance settings
type PerformanceConfig struct {
	MaxWorkers     int    `mapstructure:"max_workers"`
	RequestTimeout string `mapstructure:"request_timeout"`
	MaxContextSize int    `mapstructure:"max_context_size"` // Max characters in logs/context
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string        `mapstructure:"level"`
	Format string        `mapstructure:"format"` // "json", "text"
	Output []string      `mapstructure:"output"` // "stdout", "stderr", "file"
	File   FileLogConfig `mapstructure:"file"`
}

// FileLogConfig represents file logging configuration
type FileLogConfig struct {
	Path       string `mapstructure:"path"`
	MaxSize    string `mapstructure:"max_size"`
	MaxAge     string `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

// FeaturesConfig represents feature flags
type FeaturesConfig struct {
	EnablePrediction       bool `mapstructure:"enable_prediction"`
	EnableLearning         bool `mapstructure:"enable_learning"`
	EnableKnowledgeGraph   bool `mapstructure:"enable_knowledge_graph"`
	EnableAnomalyDetection bool `mapstructure:"enable_anomaly_detection"`
	EnableCaseSimilarity   bool `mapstructure:"enable_case_similarity"`
}

// MemoryConfig represents memory system configuration
type MemoryConfig struct {
	EnableVectorStore bool   `mapstructure:"enable_vector_store"` // 启用向量存储
	VectorStoreType   string `mapstructure:"vector_store_type"`   // 向量存储类型: "chroma"
	VectorStorePath   string `mapstructure:"vector_store_path"`   // 向量存储路径
	EmbeddingModel    string `mapstructure:"embedding_model"`     // Embedding 模型
	EmbeddingProvider string `mapstructure:"embedding_provider"`  // Embedding 提供商: "openai", "local"
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
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Analysis.MinConfidence < 0 || config.Analysis.MinConfidence > 1 {
		return fmt.Errorf("invalid min_confidence: %f (must be between 0 and 1)", config.Analysis.MinConfidence)
	}

	if config.Analysis.MaxRecommendations < 1 {
		return fmt.Errorf("max_recommendations must be at least 1")
	}

	// Validate LLM providers if enabled
	if config.LLM.Enabled {
		if len(config.LLM.Providers) == 0 {
			return fmt.Errorf("LLM is enabled but no providers configured")
		}

		for i, provider := range config.LLM.Providers {
			if provider.Name == "" {
				return fmt.Errorf("provider %d: name is required", i)
			}
			if provider.APIKey == "" {
				// Warning: API key not set, provider will be skipped
				fmt.Fprintf(os.Stderr, "Warning: LLM provider %s has no API key, will be skipped\n", provider.Name)
			}
		}
	}

	// Validate Memory configuration (required)
	if config.Memory.EnableVectorStore {
		if config.Memory.VectorStoreType == "" {
			return fmt.Errorf("vector_store_type is required when vector store is enabled")
		}
		// 验证支持的向量存储类型
		validTypes := []string{"chroma", "pinecone", "weaviate"}
		valid := false
		for _, t := range validTypes {
			if config.Memory.VectorStoreType == t {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid vector_store_type: %s (must be one of: chroma, pinecone, weaviate)", config.Memory.VectorStoreType)
		}
	}

	return nil
}

// GetRequestTimeout returns the request timeout duration
func (c *Config) GetRequestTimeout() time.Duration {
	duration, err := time.ParseDuration(c.Performance.RequestTimeout)
	if err != nil {
		return 30 * time.Second // Default
	}
	return duration
}

// GetAccuracyUpdateInterval returns the accuracy update interval duration
func (c *Config) GetAccuracyUpdateInterval() time.Duration {
	duration, err := time.ParseDuration(c.Learning.AccuracyUpdateInterval)
	if err != nil {
		return 1 * time.Hour // Default
	}
	return duration
}
