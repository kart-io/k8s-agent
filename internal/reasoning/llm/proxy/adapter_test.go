package proxy

import (
	"testing"

	"github.com/kart-io/k8s-agent/internal/reasoning/config"
)

func TestNewProxyAdapter(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.LLMConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
			errMsg:  "LLM config is nil",
		},
		{
			name: "disabled LLM",
			cfg: &config.LLMConfig{
				Enabled: false,
			},
			wantErr: true,
			errMsg:  "LLM is disabled in config",
		},
		{
			name: "no providers",
			cfg: &config.LLMConfig{
				Enabled:   true,
				Providers: []config.LLMProviderConfig{},
			},
			wantErr: true,
			errMsg:  "no LLM providers configured",
		},
		{
			name: "all providers missing API keys",
			cfg: &config.LLMConfig{
				Enabled: true,
				Providers: []config.LLMProviderConfig{
					{Name: "openai", APIKey: "", Priority: 1},
					{Name: "gemini", APIKey: "", Priority: 2},
				},
			},
			wantErr: true,
			errMsg:  "no valid LLM providers available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewProxyAdapter(tt.cfg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewProxyAdapter() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewProxyAdapter() unexpected error: %v", err)
				return
			}

			if adapter == nil {
				t.Errorf("NewProxyAdapter() returned nil adapter")
				return
			}

			if adapter.metrics == nil {
				t.Errorf("NewProxyAdapter() metrics not initialized")
			}

			if adapter.metrics.ProviderStats == nil {
				t.Errorf("NewProxyAdapter() provider stats not initialized")
			}
		})
	}

	// Note: 测试实际的 gollm 客户端创建需要真实的 API key
	// 这些测试应该在集成测试中进行,而不是在单元测试中
	t.Log("Tests for actual gollm client creation require real API keys and should be done in integration tests")
}

func TestProviderPrioritySorting(t *testing.T) {
	t.Skip("Skipping test that requires real API keys - priority sorting is already tested in basic config tests")
	// 这个测试需要真实的 gollm 客户端,应该在集成测试中进行
}

func TestGetMetrics(t *testing.T) {
	t.Skip("Skipping test that requires real API keys - metrics functionality will be tested in integration tests")
	// 这个测试需要真实的 gollm 客户端,应该在集成测试中进行
}

func TestGetProviderStatus(t *testing.T) {
	t.Skip("Skipping test that requires real API keys - provider status will be tested in integration tests")
	// 这个测试需要真实的 gollm 客户端,应该在集成测试中进行
}
