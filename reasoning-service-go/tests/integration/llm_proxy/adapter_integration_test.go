package llm_proxy

import (
	"context"
	"reasoning-service-go/internal/config"
	"reasoning-service-go/pkg/llm/proxy"
	"reasoning-service-go/tests/integration/testutil"
	"testing"
	"time"
)

// TestProxyAdapterIntegration 端到端测试 LLM Proxy Adapter
func TestProxyAdapterIntegration(t *testing.T) {
	// 跳过短测试
	testutil.SkipIfShort(t, "LLM Proxy Adapter integration test")

	// 加载测试配置
	cfg := testutil.LoadTestConfig(t)

	// 设置测试数据目录
	testDataDir := testutil.SetupTestDataDir(t)
	defer testutil.CleanupTestDataDir(t)

	t.Logf("Test data directory: %s", testDataDir)

	// 测试配置加载
	t.Run("Config Loading", func(t *testing.T) {
		if cfg == nil {
			t.Fatal("Config is nil")
		}

		if !cfg.LLM.Enabled {
			t.Error("LLM should be enabled in test config")
		}

		if len(cfg.LLM.Providers) == 0 {
			t.Error("No LLM providers configured")
		}

		t.Logf("Loaded config with %d providers", len(cfg.LLM.Providers))
	})

	// 测试创建 Proxy Adapter (会跳过因为需要真实 API key)
	t.Run("Create Proxy Adapter", func(t *testing.T) {
		// 这个测试需要真实的 API key 或 mock server
		// 在真实集成测试环境中运行

		// 检查是否有测试用的 API key
		hasValidKey := false
		for _, provider := range cfg.LLM.Providers {
			if provider.APIKey != "" && provider.APIKey != "test-api-key-12345" {
				hasValidKey = true
				break
			}
		}

		if !hasValidKey {
			t.Skip("Skipping: no valid API keys configured for integration test")
		}

		adapter, err := proxy.NewProxyAdapter(&cfg.LLM)
		if err != nil {
			t.Fatalf("Failed to create proxy adapter: %v", err)
		}

		if adapter == nil {
			t.Fatal("Adapter is nil")
		}

		t.Log("Successfully created proxy adapter")
	})
}

// TestProxyAdapterWithMock 使用 Mock LLM 测试 Proxy Adapter
func TestProxyAdapterWithMock(t *testing.T) {
	testutil.SkipIfShort(t, "Mock LLM Proxy test")

	// 创建 mock LLM
	mockLLM := testutil.NewMockLLM()

	// 添加预定义响应
	mockLLM.WithResponse(
		"root cause",
		testutil.CreateMockRootCauseResponse(),
	).WithResponse(
		"describe",
		testutil.CreateMockDescriptionResponse(),
	).WithResponse(
		"recommend",
		testutil.CreateMockRecommendationResponse(),
	)

	t.Run("Mock LLM Basic Operations", func(t *testing.T) {
		ctx := context.Background()

		// 测试生成响应
		prompt := testutil.NewMockPrompt("Analyze the root cause of pod failure")
		response, err := mockLLM.Generate(ctx, prompt)

		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		if response == "" {
			t.Error("Response is empty")
		}

		// 验证调用历史
		if mockLLM.GetCallCount() != 1 {
			t.Errorf("Expected 1 call, got %d", mockLLM.GetCallCount())
		}

		lastCall := mockLLM.GetLastCall()
		if lastCall == nil {
			t.Fatal("Last call is nil")
		}

		if lastCall.Response != response {
			t.Error("Last call response doesn't match")
		}

		t.Logf("Mock LLM response: %s", response)
	})

	t.Run("Mock LLM Error Handling", func(t *testing.T) {
		ctx := context.Background()

		// 重置 mock 并设置错误
		mockLLM.Reset()
		mockLLM.WithError("simulated API error")

		prompt := testutil.NewMockPrompt("test")
		_, err := mockLLM.Generate(ctx, prompt)

		if err == nil {
			t.Error("Expected error but got nil")
		}

		if err.Error() != "simulated API error" {
			t.Errorf("Expected 'simulated API error', got '%s'", err.Error())
		}

		// 验证错误记录在历史中
		lastCall := mockLLM.GetLastCall()
		if lastCall == nil {
			t.Fatal("Last call is nil")
		}

		if lastCall.Error == nil {
			t.Error("Last call should have error")
		}
	})
}

// TestConfigValidation 测试配置验证
func TestConfigValidation(t *testing.T) {
	testutil.SkipIfShort(t, "Config validation test")

	cfg := testutil.LoadTestConfig(t)

	tests := []struct {
		name      string
		checkFunc func(*testing.T, *config.Config)
	}{
		{
			name: "Server Config",
			checkFunc: func(t *testing.T, c *config.Config) {
				if c.Server.Port <= 0 || c.Server.Port > 65535 {
					t.Errorf("Invalid server port: %d", c.Server.Port)
				}
				if c.Server.Host == "" {
					t.Error("Server host is empty")
				}
			},
		},
		{
			name: "LLM Config",
			checkFunc: func(t *testing.T, c *config.Config) {
				if !c.LLM.Enabled {
					t.Error("LLM should be enabled")
				}
				if len(c.LLM.Providers) == 0 {
					t.Error("No providers configured")
				}
			},
		},
		{
			name: "Memory Config",
			checkFunc: func(t *testing.T, c *config.Config) {
				if c.Memory.EnableVectorStore {
					if c.Memory.VectorStoreType == "" {
						t.Error("Vector store type is empty")
					}
					if c.Memory.VectorStorePath == "" {
						t.Error("Vector store path is empty")
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.checkFunc(t, cfg)
		})
	}
}

// TestPerformanceBaseline 基准性能测试
func TestPerformanceBaseline(t *testing.T) {
	testutil.SkipIfShort(t, "Performance baseline test")

	mockLLM := testutil.NewMockLLM()
	mockLLM.WithResponse("test", "response")

	ctx := context.Background()
	prompt := testutil.NewMockPrompt("test prompt")

	// 预热
	for i := 0; i < 10; i++ {
		mockLLM.Generate(ctx, prompt)
	}

	// 性能测试
	start := time.Now()
	iterations := 100

	for i := 0; i < iterations; i++ {
		_, err := mockLLM.Generate(ctx, prompt)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
	}

	elapsed := time.Since(start)
	avgLatency := elapsed / time.Duration(iterations)

	t.Logf("Performance baseline:")
	t.Logf("  Total iterations: %d", iterations)
	t.Logf("  Total time: %v", elapsed)
	t.Logf("  Average latency: %v", avgLatency)
	t.Logf("  Requests/sec: %.2f", float64(iterations)/elapsed.Seconds())

	// 基准: 每次调用应该 < 1ms (mock 很快)
	if avgLatency > 1*time.Millisecond {
		t.Errorf("Average latency too high: %v (expected < 1ms)", avgLatency)
	}
}
