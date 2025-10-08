package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	LLM         LLMConfig         `yaml:"llm"`
	Analysis    AnalysisConfig    `yaml:"analysis"`
	Prediction  PredictionConfig  `yaml:"prediction"`
	Learning    LearningConfig    `yaml:"learning"`
	Performance PerformanceConfig `yaml:"performance"`
	Logging     LoggingConfig     `yaml:"logging"`
	Features    FeaturesConfig    `yaml:"features"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	LogLevel string `yaml:"log_level"`
}

// LLMConfig represents LLM provider configuration
type LLMConfig struct {
	Enabled   bool              `yaml:"enabled"`
	Providers []LLMProviderConfig `yaml:"providers"`
}

// LLMProviderConfig represents a single LLM provider
type LLMProviderConfig struct {
	Name        string  `yaml:"name"`        // "openai", "gemini", "deepseek"
	APIKey      string  `yaml:"api_key"`     // Can be set via env var
	BaseURL     string  `yaml:"base_url"`
	Model       string  `yaml:"model"`
	MaxTokens   int     `yaml:"max_tokens"`
	Temperature float64 `yaml:"temperature"`
	Timeout     int     `yaml:"timeout"` // in seconds
	Priority    int     `yaml:"priority"` // Higher priority providers are tried first
}

// AnalysisConfig represents analysis settings
type AnalysisConfig struct {
	MinConfidence       float64 `yaml:"min_confidence"`
	MaxRecommendations  int     `yaml:"max_recommendations"`
	IncludeSimilarCases bool    `yaml:"include_similar_cases"`
	SimilarityThreshold float64 `yaml:"similarity_threshold"`
	UseLLMFallback      bool    `yaml:"use_llm_fallback"` // Use LLM if rule-based analysis fails
}

// PredictionConfig represents prediction settings
type PredictionConfig struct {
	TimeWindows       []string          `yaml:"time_windows"`
	AnomalyDetection  AnomalyConfig     `yaml:"anomaly_detection"`
}

// AnomalyConfig represents anomaly detection settings
type AnomalyConfig struct {
	Contamination float64 `yaml:"contamination"`
	NEstimators   int     `yaml:"n_estimators"`
}

// LearningConfig represents learning system settings
type LearningConfig struct {
	EnableFeedback         bool   `yaml:"enable_feedback"`
	MinSamplesForAccuracy  int    `yaml:"min_samples_for_accuracy"`
	AccuracyUpdateInterval string `yaml:"accuracy_update_interval"`
	ExportLearningData     bool   `yaml:"export_learning_data"`
	ExportPath             string `yaml:"export_path"`
}

// PerformanceConfig represents performance settings
type PerformanceConfig struct {
	MaxWorkers      int    `yaml:"max_workers"`
	RequestTimeout  string `yaml:"request_timeout"`
	MaxContextSize  int    `yaml:"max_context_size"` // Max characters in logs/context
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string         `yaml:"level"`
	Format string         `yaml:"format"` // "json", "text"
	Output []string       `yaml:"output"` // "stdout", "stderr", "file"
	File   FileLogConfig  `yaml:"file"`
}

// FileLogConfig represents file logging configuration
type FileLogConfig struct {
	Path       string `yaml:"path"`
	MaxSize    string `yaml:"max_size"`
	MaxAge     string `yaml:"max_age"`
	MaxBackups int    `yaml:"max_backups"`
	Compress   bool   `yaml:"compress"`
}

// FeaturesConfig represents feature flags
type FeaturesConfig struct {
	EnablePrediction       bool `yaml:"enable_prediction"`
	EnableLearning         bool `yaml:"enable_learning"`
	EnableKnowledgeGraph   bool `yaml:"enable_knowledge_graph"`
	EnableAnomalyDetection bool `yaml:"enable_anomaly_detection"`
	EnableCaseSimilarity   bool `yaml:"enable_case_similarity"`
}

// Load loads configuration from file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply environment variable overrides
	applyEnvOverrides(&config)

	// Validate configuration
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// applyEnvOverrides applies environment variable overrides
func applyEnvOverrides(config *Config) {
	// Server overrides
	if host := os.Getenv("SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &config.Server.Port)
	}

	// LLM API keys from environment
	for i := range config.LLM.Providers {
		provider := &config.LLM.Providers[i]
		envKey := fmt.Sprintf("%s_API_KEY", provider.Name)
		// Try uppercase version
		if apiKey := os.Getenv(envKey); apiKey != "" {
			provider.APIKey = apiKey
		}
		// Common environment variable names
		switch provider.Name {
		case "openai":
			if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
				provider.APIKey = apiKey
			}
		case "gemini":
			if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" || os.Getenv("GOOGLE_API_KEY") != "" {
				if apiKey == "" {
					apiKey = os.Getenv("GOOGLE_API_KEY")
				}
				provider.APIKey = apiKey
			}
		case "deepseek":
			if apiKey := os.Getenv("DEEPSEEK_API_KEY"); apiKey != "" {
				provider.APIKey = apiKey
			}
		case "siliconflow":
			if apiKey := os.Getenv("SILICONFLOW_API_KEY"); apiKey != "" {
				provider.APIKey = apiKey
			}
		case "kimi":
			if apiKey := os.Getenv("KIMI_API_KEY"); apiKey != "" {
				provider.APIKey = apiKey
			} else if apiKey := os.Getenv("MOONSHOT_API_KEY"); apiKey != "" {
				provider.APIKey = apiKey
			}
		case "custom":
			if apiKey := os.Getenv("CUSTOM_LLM_API_KEY"); apiKey != "" {
				provider.APIKey = apiKey
			}
			// Also support base URL override
			if baseURL := os.Getenv("CUSTOM_LLM_BASE_URL"); baseURL != "" {
				provider.BaseURL = baseURL
			}
			// Support model override
			if model := os.Getenv("CUSTOM_LLM_MODEL"); model != "" {
				provider.Model = model
			}
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
