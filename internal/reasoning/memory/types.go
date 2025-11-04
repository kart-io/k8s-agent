package memory

import (
	"context"
	"time"
)

// Manager 定义 Memory 管理器接口.
type Manager interface {
	// AddConversation 添加对话到记忆
	AddConversation(ctx context.Context, conv *Conversation) error

	// GetConversationHistory 获取对话历史
	GetConversationHistory(ctx context.Context, sessionID string, limit int) ([]*Conversation, error)

	// SearchSimilarCases 搜索相似案例
	SearchSimilarCases(ctx context.Context, query string, limit int) ([]*CaseMemory, error)

	// AddCase 添加案例到向量存储
	AddCase(ctx context.Context, caseMemory *CaseMemory) error

	// Clear 清空指定会话的记忆
	Clear(ctx context.Context, sessionID string) error
}

// Conversation 对话记录.
type Conversation struct {
	ID        string                 `json:"id"`         // 对话 ID
	SessionID string                 `json:"session_id"` // 会话 ID
	Role      string                 `json:"role"`       // 角色: "user", "assistant", "system"
	Content   string                 `json:"content"`    // 内容
	Timestamp time.Time              `json:"timestamp"`  // 时间戳
	Metadata  map[string]interface{} `json:"metadata"`   // 元数据
}

// CaseMemory 案例记忆.
type CaseMemory struct {
	ID           string                 `json:"id"`            // 案例 ID
	Description  string                 `json:"description"`   // 描述
	RootCause    string                 `json:"root_cause"`    // 根因
	Solution     string                 `json:"solution"`      // 解决方案
	FailureType  string                 `json:"failure_type"`  // 故障类型
	ResourceType string                 `json:"resource_type"` // 资源类型
	Symptoms     []string               `json:"symptoms"`      // 症状
	Tags         []string               `json:"tags"`          // 标签
	Embedding    []float64              `json:"embedding"`     // 向量嵌入
	Similarity   float64                `json:"similarity"`    // 相似度（搜索结果）
	CreatedAt    time.Time              `json:"created_at"`    // 创建时间
	UpdatedAt    time.Time              `json:"updated_at"`    // 更新时间
	Metadata     map[string]interface{} `json:"metadata"`      // 元数据
}

// ConversationMemory 对话记忆接口.
type ConversationMemory interface {
	// Add 添加对话
	Add(ctx context.Context, conv *Conversation) error

	// Get 获取对话历史
	Get(ctx context.Context, sessionID string, limit int) ([]*Conversation, error)

	// Clear 清空会话
	Clear(ctx context.Context, sessionID string) error

	// Count 获取会话对话数量
	Count(ctx context.Context, sessionID string) (int, error)
}

// VectorStore 向量存储接口.
type VectorStore interface {
	// Add 添加向量
	Add(ctx context.Context, id string, embedding []float64, metadata map[string]interface{}) error

	// Search 搜索相似向量
	Search(ctx context.Context, embedding []float64, limit int) ([]*SearchResult, error)

	// Delete 删除向量
	Delete(ctx context.Context, id string) error

	// Clear 清空存储
	Clear(ctx context.Context) error
}

// SearchResult 搜索结果.
type SearchResult struct {
	ID        string                 `json:"id"`        // ID
	Score     float64                `json:"score"`     // 相似度分数
	Embedding []float64              `json:"embedding"` // 向量
	Metadata  map[string]interface{} `json:"metadata"`  // 元数据
}

// Embedder 嵌入生成器接口.
type Embedder interface {
	// Embed 生成文本嵌入
	Embed(ctx context.Context, text string) ([]float64, error)

	// EmbedBatch 批量生成嵌入
	EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
}

// ManagerConfig Memory 管理器配置.
type ManagerConfig struct {
	// ConversationMemory 配置
	EnableConversation    bool `json:"enable_conversation"`     // 是否启用对话记忆
	MaxConversationLength int  `json:"max_conversation_length"` // 最大对话长度

	// VectorStore 配置
	EnableVectorStore  bool   `json:"enable_vector_store"` // 是否启用向量存储
	VectorStoreType    string `json:"vector_store_type"`   // 向量存储类型
	VectorStorePath    string `json:"vector_store_path"`   // 向量存储路径
	EmbeddingDimension int    `json:"embedding_dimension"` // 嵌入维度

	// Embedder 配置
	EmbeddingModel    string `json:"embedding_model"`    // 嵌入模型
	EmbeddingProvider string `json:"embedding_provider"` // 嵌入提供商

	// 搜索配置
	DefaultSearchLimit  int     `json:"default_search_limit"` // 默认搜索结果数量
	SimilarityThreshold float64 `json:"similarity_threshold"` // 相似度阈值
}

// ConversationMemoryConfig 对话记忆配置.
type ConversationMemoryConfig struct {
	MaxLength int           `json:"max_length"` // 最大记忆长度
	TTL       time.Duration `json:"ttl"`        // 过期时间
}

// VectorStoreConfig 向量存储配置.
type VectorStoreConfig struct {
	Type      string `json:"type"`      // 类型: "chroma", "memory"
	Path      string `json:"path"`      // 路径
	Dimension int    `json:"dimension"` // 维度
}

// EmbedderConfig 嵌入生成器配置.
type EmbedderConfig struct {
	Model    string `json:"model"`    // 模型名称
	Provider string `json:"provider"` // 提供商: "openai", "local"
	APIKey   string `json:"api_key"`  // API Key
}

// 支持的向量存储类型.
const (
	VectorStoreTypeMemory = "memory" // 内存存储（用于测试）
	VectorStoreTypeChroma = "chroma" // Chroma 向量数据库
)

// 支持的嵌入提供商.
const (
	EmbedderProviderOpenAI = "openai" // OpenAI
	EmbedderProviderLocal  = "local"  // 本地模型
	EmbedderProviderMock   = "mock"   // Mock（用于测试）
)

// 默认配置值.
const (
	DefaultMaxConversationLength = 10             // 默认最大对话长度
	DefaultSearchLimit           = 5              // 默认搜索结果数量
	DefaultSimilarityThreshold   = 0.7            // 默认相似度阈值
	DefaultEmbeddingDimension    = 1536           // OpenAI text-embedding-ada-002 维度
	DefaultConversationTTL       = 24 * time.Hour // 默认对话过期时间
)
