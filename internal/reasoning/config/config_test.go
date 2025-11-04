package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid basic config",
			config: &Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Analysis: AnalysisConfig{
					MinConfidence:      0.7,
					MaxRecommendations: 5,
				},
				LLM: LLMConfig{
					Enabled:   false,
					Providers: []LLMProviderConfig{},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid server port - too low",
			config: &Config{
				Server: ServerConfig{
					Port: 0,
				},
				Analysis: AnalysisConfig{
					MinConfidence:      0.7,
					MaxRecommendations: 5,
				},
			},
			wantErr: true,
			errMsg:  "invalid server port",
		},
		{
			name: "invalid server port - too high",
			config: &Config{
				Server: ServerConfig{
					Port: 70000,
				},
				Analysis: AnalysisConfig{
					MinConfidence:      0.7,
					MaxRecommendations: 5,
				},
			},
			wantErr: true,
			errMsg:  "invalid server port",
		},
		{
			name: "invalid min_confidence - too low",
			config: &Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Analysis: AnalysisConfig{
					MinConfidence:      -0.1,
					MaxRecommendations: 5,
				},
			},
			wantErr: true,
			errMsg:  "invalid min_confidence",
		},
		{
			name: "invalid min_confidence - too high",
			config: &Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Analysis: AnalysisConfig{
					MinConfidence:      1.5,
					MaxRecommendations: 5,
				},
			},
			wantErr: true,
			errMsg:  "invalid min_confidence",
		},
		{
			name: "invalid max_recommendations",
			config: &Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Analysis: AnalysisConfig{
					MinConfidence:      0.7,
					MaxRecommendations: 0,
				},
			},
			wantErr: true,
			errMsg:  "max_recommendations must be at least 1",
		},
		{
			name: "LLM enabled but no providers",
			config: &Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Analysis: AnalysisConfig{
					MinConfidence:      0.7,
					MaxRecommendations: 5,
				},
				LLM: LLMConfig{
					Enabled:   true,
					Providers: []LLMProviderConfig{},
				},
			},
			wantErr: true,
			errMsg:  "LLM is enabled but no providers configured",
		},
		{
			name: "memory configuration with valid vector store",
			config: &Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Analysis: AnalysisConfig{
					MinConfidence:      0.7,
					MaxRecommendations: 5,
				},
				Memory: MemoryConfig{
					EnableVectorStore: true,
					VectorStoreType:   "chroma",
					VectorStorePath:   "./data/chroma",
				},
			},
			wantErr: false,
		},
		{
			name: "memory configuration with invalid vector store type",
			config: &Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Analysis: AnalysisConfig{
					MinConfidence:      0.7,
					MaxRecommendations: 5,
				},
				Memory: MemoryConfig{
					EnableVectorStore: true,
					VectorStoreType:   "invalid",
				},
			},
			wantErr: true,
			errMsg:  "invalid vector_store_type",
		},
		{
			name: "memory configuration with missing vector store type",
			config: &Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Analysis: AnalysisConfig{
					MinConfidence:      0.7,
					MaxRecommendations: 5,
				},
				Memory: MemoryConfig{
					EnableVectorStore: true,
					VectorStoreType:   "",
				},
			},
			wantErr: true,
			errMsg:  "vector_store_type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Suppress warning output during tests
			oldStderr := os.Stderr
			os.Stderr, _ = os.Open(os.DevNull)
			defer func() { os.Stderr = oldStderr }()

			err := validate(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validate() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("validate() unexpected error: %v", err)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
server:
  host: "0.0.0.0"
  port: 8080
  log_level: "info"

llm:
  enabled: false
  providers: []

analysis:
  min_confidence: 0.7
  max_recommendations: 5
  include_similar_cases: false
  similarity_threshold: 0.8
  use_llm_fallback: true

memory:
  enable_vector_store: false
  vector_store_type: "chroma"
  vector_store_path: "./data/chroma"
  embedding_model: "text-embedding-ada-002"
  embedding_provider: "openai"

performance:
  max_workers: 10
  request_timeout: "30s"
  max_context_size: 10000

logging:
  level: "info"
  format: "json"
`

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// 加载配置
	config, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("LoadFromPath() failed: %v", err)
	}

	// 验证配置加载正确
	if config.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", config.Server.Port)
	}

	if config.Analysis.MinConfidence != 0.7 {
		t.Errorf("Analysis.MinConfidence = %f, want 0.7", config.Analysis.MinConfidence)
	}

	// 验证 Memory 配置
	if config.Memory.VectorStoreType != "chroma" {
		t.Errorf("Memory.VectorStoreType = %s, want chroma", config.Memory.VectorStoreType)
	}

	if config.Memory.EmbeddingModel != "text-embedding-ada-002" {
		t.Errorf("Memory.EmbeddingModel = %s, want text-embedding-ada-002", config.Memory.EmbeddingModel)
	}
}

func TestLLMEnvOverrides(t *testing.T) {
	tests := []struct {
		name           string
		configAPIKey   string
		envVars        map[string]string
		providerName   string
		expectedAPIKey string
		description    string
	}{
		{
			name:           "config file has API key - should not override",
			configAPIKey:   "config-key-123",
			envVars:        map[string]string{"OPENAI_API_KEY": "env-key-456"},
			providerName:   "openai",
			expectedAPIKey: "config-key-123",
			description:    "Config file API key takes precedence",
		},
		{
			name:           "config file empty, env var set - should use env var",
			configAPIKey:   "",
			envVars:        map[string]string{"OPENAI_API_KEY": "env-key-789"},
			providerName:   "openai",
			expectedAPIKey: "env-key-789",
			description:    "Environment variable fills empty config",
		},
		{
			name:           "gemini with GEMINI_API_KEY",
			configAPIKey:   "",
			envVars:        map[string]string{"GEMINI_API_KEY": "gemini-key-123"},
			providerName:   "gemini",
			expectedAPIKey: "gemini-key-123",
			description:    "Gemini uses GEMINI_API_KEY",
		},
		{
			name:           "gemini with GOOGLE_API_KEY fallback",
			configAPIKey:   "",
			envVars:        map[string]string{"GOOGLE_API_KEY": "google-key-456"},
			providerName:   "gemini",
			expectedAPIKey: "google-key-456",
			description:    "Gemini falls back to GOOGLE_API_KEY",
		},
		{
			name:           "gemini prefers GEMINI_API_KEY over GOOGLE_API_KEY",
			configAPIKey:   "",
			envVars:        map[string]string{"GEMINI_API_KEY": "gemini-key", "GOOGLE_API_KEY": "google-key"},
			providerName:   "gemini",
			expectedAPIKey: "gemini-key",
			description:    "GEMINI_API_KEY has priority over GOOGLE_API_KEY",
		},
		{
			name:           "kimi with KIMI_API_KEY",
			configAPIKey:   "",
			envVars:        map[string]string{"KIMI_API_KEY": "kimi-key-123"},
			providerName:   "kimi",
			expectedAPIKey: "kimi-key-123",
			description:    "Kimi uses KIMI_API_KEY",
		},
		{
			name:           "kimi with MOONSHOT_API_KEY fallback",
			configAPIKey:   "",
			envVars:        map[string]string{"MOONSHOT_API_KEY": "moonshot-key-456"},
			providerName:   "kimi",
			expectedAPIKey: "moonshot-key-456",
			description:    "Kimi falls back to MOONSHOT_API_KEY",
		},
		{
			name:           "deepseek with DEEPSEEK_API_KEY",
			configAPIKey:   "",
			envVars:        map[string]string{"DEEPSEEK_API_KEY": "deepseek-key-123"},
			providerName:   "deepseek",
			expectedAPIKey: "deepseek-key-123",
			description:    "DeepSeek uses DEEPSEEK_API_KEY",
		},
		{
			name:           "siliconflow with SILICONFLOW_API_KEY",
			configAPIKey:   "",
			envVars:        map[string]string{"SILICONFLOW_API_KEY": "silicon-key-123"},
			providerName:   "siliconflow",
			expectedAPIKey: "silicon-key-123",
			description:    "SiliconFlow uses SILICONFLOW_API_KEY",
		},
		{
			name:           "custom provider with CUSTOM_LLM_API_KEY",
			configAPIKey:   "",
			envVars:        map[string]string{"CUSTOM_LLM_API_KEY": "custom-key-123"},
			providerName:   "custom",
			expectedAPIKey: "custom-key-123",
			description:    "Custom provider uses CUSTOM_LLM_API_KEY",
		},
		{
			name:           "ollama with OLLAMA_API_KEY (optional)",
			configAPIKey:   "",
			envVars:        map[string]string{"OLLAMA_API_KEY": "ollama-key-123"},
			providerName:   "ollama",
			expectedAPIKey: "ollama-key-123",
			description:    "Ollama can use OLLAMA_API_KEY",
		},
		{
			name:           "unknown provider uses uppercase pattern",
			configAPIKey:   "",
			envVars:        map[string]string{"MYAPI_API_KEY": "myapi-key-123"},
			providerName:   "myapi",
			expectedAPIKey: "myapi-key-123",
			description:    "Unknown provider uses {UPPERCASE_NAME}_API_KEY",
		},
		{
			name:           "no env var set - API key remains empty",
			configAPIKey:   "",
			envVars:        map[string]string{},
			providerName:   "openai",
			expectedAPIKey: "",
			description:    "No env var means API key stays empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			for key := range tt.envVars {
				_ = os.Unsetenv(key)
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}
			defer func() {
				// Clean up
				for key := range tt.envVars {
					_ = os.Unsetenv(key)
				}
			}()

			// Create test config
			cfg := &Config{
				LLM: LLMConfig{
					Enabled: true,
					Providers: []LLMProviderConfig{
						{
							Name:   tt.providerName,
							APIKey: tt.configAPIKey,
							Model:  "test-model",
						},
					},
				},
			}

			// Apply environment variable overrides
			applyLLMEnvOverrides(cfg)

			// Verify result
			actualAPIKey := cfg.LLM.Providers[0].APIKey
			if actualAPIKey != tt.expectedAPIKey {
				t.Errorf("%s: expected API key '%s', got '%s'",
					tt.description, tt.expectedAPIKey, actualAPIKey)
			}
		})
	}
}

func TestCustomProviderEnvOverrides(t *testing.T) {
	// Test that custom provider also supports base URL and model overrides
	t.Setenv("CUSTOM_LLM_API_KEY", "custom-key-123")
	t.Setenv("CUSTOM_LLM_BASE_URL", "https://custom.example.com/v1")
	t.Setenv("CUSTOM_LLM_MODEL", "custom-model-v2")

	cfg := &Config{
		LLM: LLMConfig{
			Enabled: true,
			Providers: []LLMProviderConfig{
				{
					Name:    "custom",
					APIKey:  "",
					BaseURL: "https://default.example.com",
					Model:   "default-model",
				},
			},
		},
	}

	applyLLMEnvOverrides(cfg)

	provider := cfg.LLM.Providers[0]
	if provider.APIKey != "custom-key-123" {
		t.Errorf("Expected API key 'custom-key-123', got '%s'", provider.APIKey)
	}
	if provider.BaseURL != "https://custom.example.com/v1" {
		t.Errorf("Expected base URL 'https://custom.example.com/v1', got '%s'", provider.BaseURL)
	}
	if provider.Model != "custom-model-v2" {
		t.Errorf("Expected model 'custom-model-v2', got '%s'", provider.Model)
	}
}
