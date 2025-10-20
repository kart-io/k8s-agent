package memory

import (
	"context"
	"fmt"
	"testing"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name    string
		config  *ManagerConfig
		wantErr bool
	}{
		{
			name:    "nil config uses defaults",
			config:  nil,
			wantErr: false,
		},
		{
			name: "valid config",
			config: &ManagerConfig{
				EnableConversation:    true,
				MaxConversationLength: 10,
				EnableVectorStore:     true,
				VectorStoreType:       VectorStoreTypeMemory,
				EmbeddingDimension:    128,
				DefaultSearchLimit:    5,
				SimilarityThreshold:   0.7,
			},
			wantErr: false,
		},
		{
			name: "invalid max_conversation_length",
			config: &ManagerConfig{
				EnableConversation:    true,
				MaxConversationLength: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid embedding_dimension",
			config: &ManagerConfig{
				EnableConversation:    true,
				MaxConversationLength: 10,
				EnableVectorStore:     true,
				EmbeddingDimension:    -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewManager(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("NewManager() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewManager() unexpected error: %v", err)
				return
			}

			if manager == nil {
				t.Error("NewManager() returned nil")
			}
		})
	}
}

func TestManager_ConversationOperations(t *testing.T) {
	manager, err := NewManager(&ManagerConfig{
		EnableConversation:    true,
		MaxConversationLength: 10,
		EnableVectorStore:     false,
	})
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()
	sessionID := "session-1"

	// 添加对话
	conv := &Conversation{
		SessionID: sessionID,
		Role:      "user",
		Content:   "Hello",
	}
	err = manager.AddConversation(ctx, conv)
	if err != nil {
		t.Errorf("AddConversation() error: %v", err)
	}

	// 获取对话历史
	history, err := manager.GetConversationHistory(ctx, sessionID, 0)
	if err != nil {
		t.Errorf("GetConversationHistory() error: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("GetConversationHistory() returned %d conversations, want 1", len(history))
	}

	if history[0].Content != "Hello" {
		t.Errorf("Conversation content = %s, want 'Hello'", history[0].Content)
	}

	// 清空会话
	err = manager.Clear(ctx, sessionID)
	if err != nil {
		t.Errorf("Clear() error: %v", err)
	}

	// 验证已清空
	history, err = manager.GetConversationHistory(ctx, sessionID, 0)
	if err != nil {
		t.Errorf("GetConversationHistory() error: %v", err)
	}

	if len(history) != 0 {
		t.Errorf("Expected 0 conversations after clear, got %d", len(history))
	}
}

func TestManager_ConversationDisabled(t *testing.T) {
	manager, err := NewManager(&ManagerConfig{
		EnableConversation:    false,
		MaxConversationLength: 10,
		EnableVectorStore:     false,
	})
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// 尝试添加对话应该失败
	err = manager.AddConversation(ctx, &Conversation{
		SessionID: "session-1",
		Role:      "user",
		Content:   "Hello",
	})
	if err == nil {
		t.Error("AddConversation() should return error when disabled")
	}

	// 尝试获取历史应该失败
	_, err = manager.GetConversationHistory(ctx, "session-1", 0)
	if err == nil {
		t.Error("GetConversationHistory() should return error when disabled")
	}
}

func TestManager_CaseOperations(t *testing.T) {
	manager, err := NewManager(&ManagerConfig{
		EnableConversation:    false,
		MaxConversationLength: 10, // Required even when disabled
		EnableVectorStore:     true,
		VectorStoreType:       VectorStoreTypeMemory,
		EmbeddingDimension:    128,
		DefaultSearchLimit:    5,
		SimilarityThreshold:   -1.0, // Negative threshold to include all results
	})
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// 添加案例
	caseMemory := &CaseMemory{
		Description:  "Pod OOMKilled due to memory limit",
		RootCause:    "Insufficient memory limit configuration",
		Solution:     "Increase memory limit in deployment",
		FailureType:  "pod_failure",
		ResourceType: "pod",
		Symptoms:     []string{"OOMKilled", "CrashLoopBackOff"},
		Tags:         []string{"memory", "oom"},
	}

	err = manager.AddCase(ctx, caseMemory)
	if err != nil {
		t.Errorf("AddCase() error: %v", err)
	}

	if caseMemory.ID == "" {
		t.Error("Case ID should be auto-generated")
	}

	if len(caseMemory.Embedding) != 128 {
		t.Errorf("Embedding length = %d, want 128", len(caseMemory.Embedding))
	}

	// 搜索相似案例
	query := "Pod memory exceeded"
	cases, err := manager.SearchSimilarCases(ctx, query, 5)
	if err != nil {
		t.Errorf("SearchSimilarCases() error: %v", err)
	}

	if len(cases) == 0 {
		t.Error("SearchSimilarCases() should return at least one case")
	}

	// 验证返回的案例
	found := false
	for _, c := range cases {
		if c.ID == caseMemory.ID {
			found = true
			if c.Description != caseMemory.Description {
				t.Errorf("Description = %s, want %s", c.Description, caseMemory.Description)
			}
			if c.RootCause != caseMemory.RootCause {
				t.Errorf("RootCause = %s, want %s", c.RootCause, caseMemory.RootCause)
			}
			break
		}
	}

	if !found {
		t.Error("SearchSimilarCases() should return the added case")
	}
}

func TestManager_VectorStoreDisabled(t *testing.T) {
	manager, err := NewManager(&ManagerConfig{
		EnableConversation:    true,
		MaxConversationLength: 10,
		EnableVectorStore:     false,
	})
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// 尝试添加案例应该失败
	err = manager.AddCase(ctx, &CaseMemory{
		Description: "Test case",
	})
	if err == nil {
		t.Error("AddCase() should return error when disabled")
	}

	// 尝试搜索应该失败
	_, err = manager.SearchSimilarCases(ctx, "test query", 5)
	if err == nil {
		t.Error("SearchSimilarCases() should return error when disabled")
	}
}

func TestManager_MultipleCases(t *testing.T) {
	manager, err := NewManager(&ManagerConfig{
		EnableConversation:    false,
		MaxConversationLength: 10,
		EnableVectorStore:     true,
		VectorStoreType:       VectorStoreTypeMemory,
		EmbeddingDimension:    128,
		DefaultSearchLimit:    5,
		SimilarityThreshold:   -1.0, // Negative threshold to include all results
	})
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// 添加多个案例
	cases := []*CaseMemory{
		{
			Description:  "Pod OOMKilled",
			RootCause:    "Memory limit too low",
			Solution:     "Increase memory limit",
			FailureType:  "pod_failure",
			ResourceType: "pod",
		},
		{
			Description:  "Image pull error",
			RootCause:    "Invalid image name",
			Solution:     "Fix image name in deployment",
			FailureType:  "image_pull_error",
			ResourceType: "pod",
		},
		{
			Description:  "Node not ready",
			RootCause:    "Kubelet stopped",
			Solution:     "Restart kubelet service",
			FailureType:  "node_failure",
			ResourceType: "node",
		},
	}

	for _, c := range cases {
		err = manager.AddCase(ctx, c)
		if err != nil {
			t.Errorf("AddCase() error: %v", err)
		}
	}

	// 搜索相关的案例
	query := "Pod memory problem"
	results, err := manager.SearchSimilarCases(ctx, query, 10)
	if err != nil {
		t.Errorf("SearchSimilarCases() error: %v", err)
	}

	if len(results) == 0 {
		t.Error("SearchSimilarCases() should return results")
	}

	// 验证结果按相似度排序
	for i := 1; i < len(results); i++ {
		if results[i].Similarity > results[i-1].Similarity {
			t.Error("Results should be sorted by similarity in descending order")
		}
	}
}

func TestManager_SimilarityThreshold(t *testing.T) {
	manager, err := NewManager(&ManagerConfig{
		EnableConversation:    false,
		MaxConversationLength: 10,
		EnableVectorStore:     true,
		VectorStoreType:       VectorStoreTypeMemory,
		EmbeddingDimension:    128,
		DefaultSearchLimit:    10,
		SimilarityThreshold:   0.9, // 高阈值
	})
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// 添加一个案例
	err = manager.AddCase(ctx, &CaseMemory{
		Description:  "Specific pod failure scenario",
		RootCause:    "Very specific root cause",
		Solution:     "Very specific solution",
		FailureType:  "pod_failure",
		ResourceType: "pod",
	})
	if err != nil {
		t.Errorf("AddCase() error: %v", err)
	}

	// 使用不太相关的查询搜索
	query := "Node network issue"
	results, err := manager.SearchSimilarCases(ctx, query, 10)
	if err != nil {
		t.Errorf("SearchSimilarCases() error: %v", err)
	}

	// 由于相似度阈值高，可能没有结果
	t.Logf("SearchSimilarCases() returned %d results with threshold 0.9", len(results))
}

func TestManager_CaseMetadata(t *testing.T) {
	manager, err := NewManager(&ManagerConfig{
		EnableConversation:    false,
		MaxConversationLength: 10,
		EnableVectorStore:     true,
		VectorStoreType:       VectorStoreTypeMemory,
		EmbeddingDimension:    128,
		DefaultSearchLimit:    5,
		SimilarityThreshold:   0.5,
	})
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// 添加带有自定义元数据的案例
	caseMemory := &CaseMemory{
		Description:  "Test case",
		RootCause:    "Test root cause",
		Solution:     "Test solution",
		FailureType:  "test_failure",
		ResourceType: "test_resource",
		Metadata: map[string]interface{}{
			"custom_field_1": "value1",
			"custom_field_2": 123,
		},
	}

	err = manager.AddCase(ctx, caseMemory)
	if err != nil {
		t.Errorf("AddCase() error: %v", err)
	}

	// 验证时间戳已设置
	if caseMemory.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}

	if caseMemory.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
}

func TestManager_AddCaseValidation(t *testing.T) {
	manager, err := NewManager(&ManagerConfig{
		EnableConversation:    false,
		MaxConversationLength: 10,
		EnableVectorStore:     true,
		VectorStoreType:       VectorStoreTypeMemory,
		EmbeddingDimension:    128,
		DefaultSearchLimit:    5,
		SimilarityThreshold:   0.5,
	})
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name    string
		case_   *CaseMemory
		wantErr bool
	}{
		{
			name: "valid case",
			case_: &CaseMemory{
				Description: "Valid case",
			},
			wantErr: false,
		},
		{
			name:    "nil case",
			case_:   nil,
			wantErr: true,
		},
		{
			name: "missing description",
			case_: &CaseMemory{
				RootCause: "Some cause",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.AddCase(ctx, tt.case_)

			if tt.wantErr && err == nil {
				t.Error("AddCase() expected error but got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("AddCase() unexpected error: %v", err)
			}
		})
	}
}

func TestManager_SearchValidation(t *testing.T) {
	manager, err := NewManager(&ManagerConfig{
		EnableConversation:    false,
		MaxConversationLength: 10,
		EnableVectorStore:     true,
		VectorStoreType:       VectorStoreTypeMemory,
		EmbeddingDimension:    128,
		DefaultSearchLimit:    5,
		SimilarityThreshold:   0.5,
	})
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// 空查询应该失败
	_, err = manager.SearchSimilarCases(ctx, "", 5)
	if err == nil {
		t.Error("SearchSimilarCases() should return error for empty query")
	}
}

func TestManager_DefaultSearchLimit(t *testing.T) {
	manager, err := NewManager(&ManagerConfig{
		EnableConversation:    false,
		MaxConversationLength: 10,
		EnableVectorStore:     true,
		VectorStoreType:       VectorStoreTypeMemory,
		EmbeddingDimension:    128,
		DefaultSearchLimit:    3,
		SimilarityThreshold:   0.1,
	})
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// 添加多个案例
	for i := 0; i < 10; i++ {
		err = manager.AddCase(ctx, &CaseMemory{
			Description: fmt.Sprintf("Case %d", i),
		})
		if err != nil {
			t.Errorf("AddCase() error: %v", err)
		}
	}

	// 使用默认限制（传入 0）
	results, err := manager.SearchSimilarCases(ctx, "test", 0)
	if err != nil {
		t.Errorf("SearchSimilarCases() error: %v", err)
	}

	// 结果数量应该不超过默认限制
	if len(results) > 3 {
		t.Errorf("SearchSimilarCases() returned %d results, want <= 3", len(results))
	}
}
