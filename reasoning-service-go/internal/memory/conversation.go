package memory

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// InMemoryConversationMemory 内存对话记忆实现
type InMemoryConversationMemory struct {
	// sessions 存储每个会话的对话列表
	sessions map[string][]*Conversation

	// config 配置
	config *ConversationMemoryConfig

	// mu 互斥锁
	mu sync.RWMutex

	// lastAccess 记录每个会话的最后访问时间
	lastAccess map[string]time.Time
}

// NewInMemoryConversationMemory 创建新的内存对话记忆
func NewInMemoryConversationMemory(config *ConversationMemoryConfig) (*InMemoryConversationMemory, error) {
	if config == nil {
		config = &ConversationMemoryConfig{
			MaxLength: DefaultMaxConversationLength,
			TTL:       DefaultConversationTTL,
		}
	}

	// 验证配置
	if config.MaxLength <= 0 {
		return nil, fmt.Errorf("max_length must be positive")
	}
	if config.TTL <= 0 {
		return nil, fmt.Errorf("ttl must be positive")
	}

	cm := &InMemoryConversationMemory{
		sessions:   make(map[string][]*Conversation),
		config:     config,
		lastAccess: make(map[string]time.Time),
	}

	// 启动清理 goroutine
	go cm.cleanupExpired()

	return cm, nil
}

// Add 添加对话
func (cm *InMemoryConversationMemory) Add(ctx context.Context, conv *Conversation) error {
	if conv == nil {
		return fmt.Errorf("conversation is nil")
	}

	if conv.SessionID == "" {
		return fmt.Errorf("session_id is required")
	}

	if conv.Role == "" {
		return fmt.Errorf("role is required")
	}

	if conv.Content == "" {
		return fmt.Errorf("content is required")
	}

	// 生成 ID
	if conv.ID == "" {
		conv.ID = generateConversationID()
	}

	// 设置时间戳
	if conv.Timestamp.IsZero() {
		conv.Timestamp = time.Now()
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 获取或创建会话
	conversations := cm.sessions[conv.SessionID]

	// 添加新对话
	conversations = append(conversations, conv)

	// 限制长度
	if len(conversations) > cm.config.MaxLength {
		// 保留最新的对话
		conversations = conversations[len(conversations)-cm.config.MaxLength:]
	}

	cm.sessions[conv.SessionID] = conversations
	cm.lastAccess[conv.SessionID] = time.Now()

	return nil
}

// Get 获取对话历史
func (cm *InMemoryConversationMemory) Get(ctx context.Context, sessionID string, limit int) ([]*Conversation, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conversations := cm.sessions[sessionID]
	if conversations == nil {
		return []*Conversation{}, nil
	}

	// 更新最后访问时间
	cm.lastAccess[sessionID] = time.Now()

	// 应用限制
	if limit > 0 && limit < len(conversations) {
		// 返回最新的 limit 条对话
		conversations = conversations[len(conversations)-limit:]
	}

	// 复制切片以避免外部修改
	result := make([]*Conversation, len(conversations))
	copy(result, conversations)

	return result, nil
}

// Clear 清空会话
func (cm *InMemoryConversationMemory) Clear(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session_id is required")
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.sessions, sessionID)
	delete(cm.lastAccess, sessionID)

	return nil
}

// Count 获取会话对话数量
func (cm *InMemoryConversationMemory) Count(ctx context.Context, sessionID string) (int, error) {
	if sessionID == "" {
		return 0, fmt.Errorf("session_id is required")
	}

	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conversations := cm.sessions[sessionID]
	return len(conversations), nil
}

// cleanupExpired 定期清理过期会话
func (cm *InMemoryConversationMemory) cleanupExpired() {
	ticker := time.NewTicker(cm.config.TTL / 2)
	defer ticker.Stop()

	for range ticker.C {
		cm.mu.Lock()

		now := time.Now()
		for sessionID, lastAccess := range cm.lastAccess {
			if now.Sub(lastAccess) > cm.config.TTL {
				delete(cm.sessions, sessionID)
				delete(cm.lastAccess, sessionID)
			}
		}

		cm.mu.Unlock()
	}
}

// generateConversationID 生成对话 ID
func generateConversationID() string {
	return fmt.Sprintf("conv_%d", time.Now().UnixNano())
}
