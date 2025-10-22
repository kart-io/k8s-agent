package memory

import (
	"context"
	"testing"
	"time"
)

func TestNewInMemoryConversationMemory(t *testing.T) {
	tests := []struct {
		name    string
		config  *ConversationMemoryConfig
		wantErr bool
	}{
		{
			name:    "nil config uses defaults",
			config:  nil,
			wantErr: false,
		},
		{
			name: "valid config",
			config: &ConversationMemoryConfig{
				MaxLength: 20,
				TTL:       30 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "invalid max_length",
			config: &ConversationMemoryConfig{
				MaxLength: -1,
				TTL:       30 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "invalid ttl",
			config: &ConversationMemoryConfig{
				MaxLength: 10,
				TTL:       -1 * time.Minute,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm, err := NewInMemoryConversationMemory(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("NewInMemoryConversationMemory() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewInMemoryConversationMemory() unexpected error: %v", err)
				return
			}

			if cm == nil {
				t.Error("NewInMemoryConversationMemory() returned nil")
			}
		})
	}
}

func TestInMemoryConversationMemory_Add(t *testing.T) {
	cm, err := NewInMemoryConversationMemory(nil)
	if err != nil {
		t.Fatalf("Failed to create ConversationMemory: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name    string
		conv    *Conversation
		wantErr bool
	}{
		{
			name: "valid conversation",
			conv: &Conversation{
				SessionID: "session-1",
				Role:      "user",
				Content:   "Hello",
			},
			wantErr: false,
		},
		{
			name:    "nil conversation",
			conv:    nil,
			wantErr: true,
		},
		{
			name: "missing session_id",
			conv: &Conversation{
				Role:    "user",
				Content: "Hello",
			},
			wantErr: true,
		},
		{
			name: "missing role",
			conv: &Conversation{
				SessionID: "session-1",
				Content:   "Hello",
			},
			wantErr: true,
		},
		{
			name: "missing content",
			conv: &Conversation{
				SessionID: "session-1",
				Role:      "user",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cm.Add(ctx, tt.conv)

			if tt.wantErr && err == nil {
				t.Error("Add() expected error but got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Add() unexpected error: %v", err)
			}
		})
	}
}

func TestInMemoryConversationMemory_Get(t *testing.T) {
	cm, err := NewInMemoryConversationMemory(&ConversationMemoryConfig{
		MaxLength: 10,
		TTL:       1 * time.Hour,
	})
	if err != nil {
		t.Fatalf("Failed to create ConversationMemory: %v", err)
	}

	ctx := context.Background()
	sessionID := "session-1"

	// 添加一些对话
	conversations := []*Conversation{
		{SessionID: sessionID, Role: "user", Content: "Hello"},
		{SessionID: sessionID, Role: "assistant", Content: "Hi there!"},
		{SessionID: sessionID, Role: "user", Content: "How are you?"},
		{SessionID: sessionID, Role: "assistant", Content: "I'm good, thanks!"},
	}

	for _, conv := range conversations {
		if err := cm.Add(ctx, conv); err != nil {
			t.Fatalf("Failed to add conversation: %v", err)
		}
	}

	tests := []struct {
		name      string
		sessionID string
		limit     int
		wantCount int
		wantErr   bool
	}{
		{
			name:      "get all conversations",
			sessionID: sessionID,
			limit:     0,
			wantCount: 4,
			wantErr:   false,
		},
		{
			name:      "get with limit",
			sessionID: sessionID,
			limit:     2,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "non-existent session",
			sessionID: "non-existent",
			limit:     0,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "empty session_id",
			sessionID: "",
			limit:     0,
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cm.Get(ctx, tt.sessionID, tt.limit)

			if tt.wantErr {
				if err == nil {
					t.Error("Get() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Get() unexpected error: %v", err)
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("Get() returned %d conversations, want %d", len(result), tt.wantCount)
			}

			// 验证返回的是最新的对话
			if tt.limit > 0 && len(result) > 0 {
				if result[len(result)-1].Content != "I'm good, thanks!" {
					t.Error("Get() should return the most recent conversations")
				}
			}
		})
	}
}

func TestInMemoryConversationMemory_Clear(t *testing.T) {
	cm, err := NewInMemoryConversationMemory(nil)
	if err != nil {
		t.Fatalf("Failed to create ConversationMemory: %v", err)
	}

	ctx := context.Background()
	sessionID := "session-1"

	// 添加对话
	err = cm.Add(ctx, &Conversation{
		SessionID: sessionID,
		Role:      "user",
		Content:   "Hello",
	})
	if err != nil {
		t.Fatalf("Failed to add conversation: %v", err)
	}

	// 验证对话存在
	conversations, err := cm.Get(ctx, sessionID, 0)
	if err != nil {
		t.Fatalf("Failed to get conversations: %v", err)
	}
	if len(conversations) != 1 {
		t.Fatalf("Expected 1 conversation, got %d", len(conversations))
	}

	// 清空会话
	err = cm.Clear(ctx, sessionID)
	if err != nil {
		t.Errorf("Clear() unexpected error: %v", err)
	}

	// 验证对话已清空
	conversations, err = cm.Get(ctx, sessionID, 0)
	if err != nil {
		t.Fatalf("Failed to get conversations: %v", err)
	}
	if len(conversations) != 0 {
		t.Errorf("Expected 0 conversations after clear, got %d", len(conversations))
	}
}

func TestInMemoryConversationMemory_Count(t *testing.T) {
	cm, err := NewInMemoryConversationMemory(nil)
	if err != nil {
		t.Fatalf("Failed to create ConversationMemory: %v", err)
	}

	ctx := context.Background()
	sessionID := "session-1"

	// 初始计数应为 0
	count, err := cm.Count(ctx, sessionID)
	if err != nil {
		t.Fatalf("Count() error: %v", err)
	}
	if count != 0 {
		t.Errorf("Count() = %d, want 0", count)
	}

	// 添加对话
	for i := 0; i < 5; i++ {
		err = cm.Add(ctx, &Conversation{
			SessionID: sessionID,
			Role:      "user",
			Content:   "Message",
		})
		if err != nil {
			t.Fatalf("Failed to add conversation: %v", err)
		}
	}

	// 验证计数
	count, err = cm.Count(ctx, sessionID)
	if err != nil {
		t.Fatalf("Count() error: %v", err)
	}
	if count != 5 {
		t.Errorf("Count() = %d, want 5", count)
	}
}

func TestInMemoryConversationMemory_MaxLength(t *testing.T) {
	maxLength := 3
	cm, err := NewInMemoryConversationMemory(&ConversationMemoryConfig{
		MaxLength: maxLength,
		TTL:       1 * time.Hour,
	})
	if err != nil {
		t.Fatalf("Failed to create ConversationMemory: %v", err)
	}

	ctx := context.Background()
	sessionID := "session-1"

	// 添加超过 MaxLength 的对话
	for i := 0; i < 5; i++ {
		err = cm.Add(ctx, &Conversation{
			SessionID: sessionID,
			Role:      "user",
			Content:   string(rune('A' + i)),
		})
		if err != nil {
			t.Fatalf("Failed to add conversation: %v", err)
		}
	}

	// 验证只保留最新的 MaxLength 条
	conversations, err := cm.Get(ctx, sessionID, 0)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	if len(conversations) != maxLength {
		t.Errorf("Expected %d conversations, got %d", maxLength, len(conversations))
	}

	// 验证保留的是最新的对话
	expectedContent := []string{"C", "D", "E"}
	for i, conv := range conversations {
		if conv.Content != expectedContent[i] {
			t.Errorf("Conversation %d content = %s, want %s", i, conv.Content, expectedContent[i])
		}
	}
}

func TestInMemoryConversationMemory_AutoGenerateID(t *testing.T) {
	cm, err := NewInMemoryConversationMemory(nil)
	if err != nil {
		t.Fatalf("Failed to create ConversationMemory: %v", err)
	}

	ctx := context.Background()
	sessionID := "session-1"

	// 添加没有 ID 的对话
	err = cm.Add(ctx, &Conversation{
		SessionID: sessionID,
		Role:      "user",
		Content:   "Hello",
	})
	if err != nil {
		t.Fatalf("Failed to add conversation: %v", err)
	}

	// 获取对话并验证 ID 已生成
	conversations, err := cm.Get(ctx, sessionID, 0)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	if len(conversations) != 1 {
		t.Fatalf("Expected 1 conversation, got %d", len(conversations))
	}

	if conversations[0].ID == "" {
		t.Error("ID should be auto-generated")
	}
}

func TestInMemoryConversationMemory_AutoSetTimestamp(t *testing.T) {
	cm, err := NewInMemoryConversationMemory(nil)
	if err != nil {
		t.Fatalf("Failed to create ConversationMemory: %v", err)
	}

	ctx := context.Background()
	sessionID := "session-1"

	before := time.Now()

	// 添加没有时间戳的对话
	err = cm.Add(ctx, &Conversation{
		SessionID: sessionID,
		Role:      "user",
		Content:   "Hello",
	})
	if err != nil {
		t.Fatalf("Failed to add conversation: %v", err)
	}

	after := time.Now()

	// 获取对话并验证时间戳已设置
	conversations, err := cm.Get(ctx, sessionID, 0)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	if len(conversations) != 1 {
		t.Fatalf("Expected 1 conversation, got %d", len(conversations))
	}

	timestamp := conversations[0].Timestamp
	if timestamp.IsZero() {
		t.Error("Timestamp should be auto-set")
	}

	if timestamp.Before(before) || timestamp.After(after) {
		t.Error("Timestamp should be between before and after time")
	}
}

func TestInMemoryConversationMemory_MultipleSessions(t *testing.T) {
	cm, err := NewInMemoryConversationMemory(nil)
	if err != nil {
		t.Fatalf("Failed to create ConversationMemory: %v", err)
	}

	ctx := context.Background()

	// 添加不同会话的对话
	sessions := []string{"session-1", "session-2", "session-3"}
	for _, sessionID := range sessions {
		for i := 0; i < 3; i++ {
			err = cm.Add(ctx, &Conversation{
				SessionID: sessionID,
				Role:      "user",
				Content:   "Message",
			})
			if err != nil {
				t.Fatalf("Failed to add conversation: %v", err)
			}
		}
	}

	// 验证每个会话的对话数量
	for _, sessionID := range sessions {
		count, err := cm.Count(ctx, sessionID)
		if err != nil {
			t.Fatalf("Count() error for %s: %v", sessionID, err)
		}
		if count != 3 {
			t.Errorf("Session %s: Count() = %d, want 3", sessionID, count)
		}
	}

	// 清空一个会话
	err = cm.Clear(ctx, "session-1")
	if err != nil {
		t.Fatalf("Clear() error: %v", err)
	}

	// 验证只有该会话被清空
	count, err := cm.Count(ctx, "session-1")
	if err != nil {
		t.Fatalf("Count() error: %v", err)
	}
	if count != 0 {
		t.Errorf("Session session-1: Count() = %d, want 0", count)
	}

	count, err = cm.Count(ctx, "session-2")
	if err != nil {
		t.Fatalf("Count() error: %v", err)
	}
	if count != 3 {
		t.Errorf("Session session-2: Count() = %d, want 3", count)
	}
}
