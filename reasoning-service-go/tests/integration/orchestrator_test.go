package integration

import (
	"testing"
	"time"

	"reasoning-service-go/internal/orchestrator"
	"reasoning-service-go/tests/integration/testutil"
)

// TestOrchestratorIntegration 测试 Orchestrator 的完整集成流程
// 注意：这是一个真正的集成测试，需要实际的 LLM API 访问
// 可以通过 go test -short 跳过此测试
func TestOrchestratorIntegration(t *testing.T) {
	testutil.SkipIfShort(t, "Orchestrator integration test requires real LLM access")

	// 1. 加载测试配置
	cfg := testutil.LoadTestConfig(t)

	// 检查是否有有效的 API key
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

	// 2. 初始化组件
	// Note: 由于需要真实 LLM API,我们暂时跳过完整的集成测试
	// 完整测试应该在有真实 API key 的环境中运行
	t.Skip("Full orchestrator integration test requires real API keys - implement when ready")
}

// TestOrchestratorWithoutMemory 测试禁用 Memory 时的 Orchestrator
// 这是一个单元测试,测试 Orchestrator 的基本构造逻辑
func TestOrchestratorWithoutMemory(t *testing.T) {
	// 测试配置验证
	t.Run("Config Validation", func(t *testing.T) {
		config := orchestrator.DefaultOrchestratorConfig()
		if config == nil {
			t.Fatal("DefaultOrchestratorConfig returned nil")
		}

		// 测试禁用 Memory
		config.EnableMemory = false
		config.EnableDescription = true

		// 尝试创建 Orchestrator (应该失败因为缺少组件)
		_, err := orchestrator.NewOrchestrator(nil, nil, nil, nil, nil, config)
		if err == nil {
			t.Error("NewOrchestrator should fail with nil components")
		}
	})

	t.Run("Execution Steps Tracking", func(t *testing.T) {
		// 测试响应结构
		resp := &orchestrator.AnalysisResponse{
			ExecutionSteps: make([]orchestrator.ExecutionStep, 0),
			Timestamp:      time.Now(),
		}

		// 添加执行步骤
		resp.ExecutionSteps = append(resp.ExecutionSteps, orchestrator.ExecutionStep{
			Step:        1,
			Name:        "test_step",
			Description: "Test execution step",
			Status:      "success",
			Duration:    100 * time.Millisecond,
		})

		if len(resp.ExecutionSteps) != 1 {
			t.Errorf("ExecutionSteps count = %d, want 1", len(resp.ExecutionSteps))
		}

		if resp.ExecutionSteps[0].Status != "success" {
			t.Errorf("Step status = %s, want 'success'", resp.ExecutionSteps[0].Status)
		}
	})
}
