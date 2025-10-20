package memory

import (
	"context"
	"fmt"
	"log"
	"time"
)

// DefaultManager Memory 管理器默认实现
type DefaultManager struct {
	// 对话记忆
	conversationMemory ConversationMemory

	// 向量存储
	vectorStore VectorStore

	// 嵌入生成器
	embedder Embedder

	// 配置
	config *ManagerConfig
}

// NewManager 创建新的 Memory 管理器
func NewManager(config *ManagerConfig) (*DefaultManager, error) {
	if config == nil {
		config = defaultManagerConfig()
	}

	// 验证配置
	if err := validateManagerConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	m := &DefaultManager{
		config: config,
	}

	// 初始化对话记忆
	if config.EnableConversation {
		convConfig := &ConversationMemoryConfig{
			MaxLength: config.MaxConversationLength,
			TTL:       DefaultConversationTTL,
		}
		convMem, err := NewInMemoryConversationMemory(convConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create conversation memory: %w", err)
		}
		m.conversationMemory = convMem
	}

	// 初始化向量存储
	if config.EnableVectorStore {
		vsConfig := &VectorStoreConfig{
			Type:      config.VectorStoreType,
			Path:      config.VectorStorePath,
			Dimension: config.EmbeddingDimension,
		}
		vs, err := NewInMemoryVectorStore(vsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create vector store: %w", err)
		}
		m.vectorStore = vs

		// 初始化嵌入生成器
		embedder, err := NewMockEmbedder(config.EmbeddingDimension)
		if err != nil {
			return nil, fmt.Errorf("failed to create embedder: %w", err)
		}
		m.embedder = embedder
	}

	return m, nil
}

// AddConversation 添加对话到记忆
func (m *DefaultManager) AddConversation(ctx context.Context, conv *Conversation) error {
	if !m.config.EnableConversation {
		return fmt.Errorf("conversation memory is disabled")
	}

	if m.conversationMemory == nil {
		return fmt.Errorf("conversation memory not initialized")
	}

	return m.conversationMemory.Add(ctx, conv)
}

// GetConversationHistory 获取对话历史
func (m *DefaultManager) GetConversationHistory(ctx context.Context, sessionID string, limit int) ([]*Conversation, error) {
	if !m.config.EnableConversation {
		return nil, fmt.Errorf("conversation memory is disabled")
	}

	if m.conversationMemory == nil {
		return nil, fmt.Errorf("conversation memory not initialized")
	}

	if limit <= 0 {
		limit = m.config.MaxConversationLength
	}

	return m.conversationMemory.Get(ctx, sessionID, limit)
}

// SearchSimilarCases 搜索相似案例
func (m *DefaultManager) SearchSimilarCases(ctx context.Context, query string, limit int) ([]*CaseMemory, error) {
	if !m.config.EnableVectorStore {
		return nil, fmt.Errorf("vector store is disabled")
	}

	if m.vectorStore == nil || m.embedder == nil {
		return nil, fmt.Errorf("vector store not initialized")
	}

	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	if limit <= 0 {
		limit = m.config.DefaultSearchLimit
	}

	// 生成查询的嵌入
	embedding, err := m.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// 搜索相似向量
	results, err := m.vectorStore.Search(ctx, embedding, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search vector store: %w", err)
	}

	// 转换为 CaseMemory
	cases := make([]*CaseMemory, 0, len(results))
	for _, result := range results {
		// 过滤低相似度结果
		if result.Score < m.config.SimilarityThreshold {
			continue
		}

		caseMemory := &CaseMemory{
			ID:         result.ID,
			Similarity: result.Score,
			Embedding:  result.Embedding,
			Metadata:   result.Metadata,
		}

		// 从元数据中提取案例信息
		if desc, ok := result.Metadata["description"].(string); ok {
			caseMemory.Description = desc
		}
		if rc, ok := result.Metadata["root_cause"].(string); ok {
			caseMemory.RootCause = rc
		}
		if sol, ok := result.Metadata["solution"].(string); ok {
			caseMemory.Solution = sol
		}
		if ft, ok := result.Metadata["failure_type"].(string); ok {
			caseMemory.FailureType = ft
		}
		if rt, ok := result.Metadata["resource_type"].(string); ok {
			caseMemory.ResourceType = rt
		}
		if symptoms, ok := result.Metadata["symptoms"].([]string); ok {
			caseMemory.Symptoms = symptoms
		}
		if tags, ok := result.Metadata["tags"].([]string); ok {
			caseMemory.Tags = tags
		}

		cases = append(cases, caseMemory)
	}

	log.Printf("SearchSimilarCases: found %d cases (filtered from %d results)", len(cases), len(results))

	return cases, nil
}

// AddCase 添加案例到向量存储
func (m *DefaultManager) AddCase(ctx context.Context, caseMemory *CaseMemory) error {
	if !m.config.EnableVectorStore {
		return fmt.Errorf("vector store is disabled")
	}

	if m.vectorStore == nil || m.embedder == nil {
		return fmt.Errorf("vector store not initialized")
	}

	if caseMemory == nil {
		return fmt.Errorf("case memory is nil")
	}

	// 验证必需字段
	if caseMemory.Description == "" {
		return fmt.Errorf("description is required")
	}

	// 生成案例 ID（如果未提供）
	if caseMemory.ID == "" {
		caseMemory.ID = generateCaseID()
	}

	// 设置时间戳
	now := time.Now()
	if caseMemory.CreatedAt.IsZero() {
		caseMemory.CreatedAt = now
	}
	caseMemory.UpdatedAt = now

	// 构建用于嵌入的文本
	text := buildCaseText(caseMemory)

	// 生成嵌入（如果未提供）
	var embedding []float64
	var err error
	if len(caseMemory.Embedding) == 0 {
		embedding, err = m.embedder.Embed(ctx, text)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}
		caseMemory.Embedding = embedding
	} else {
		embedding = caseMemory.Embedding
	}

	// 构建元数据
	metadata := map[string]interface{}{
		"description":   caseMemory.Description,
		"root_cause":    caseMemory.RootCause,
		"solution":      caseMemory.Solution,
		"failure_type":  caseMemory.FailureType,
		"resource_type": caseMemory.ResourceType,
		"symptoms":      caseMemory.Symptoms,
		"tags":          caseMemory.Tags,
		"created_at":    caseMemory.CreatedAt,
		"updated_at":    caseMemory.UpdatedAt,
	}

	// 添加自定义元数据
	for k, v := range caseMemory.Metadata {
		metadata[k] = v
	}

	// 添加到向量存储
	err = m.vectorStore.Add(ctx, caseMemory.ID, embedding, metadata)
	if err != nil {
		return fmt.Errorf("failed to add to vector store: %w", err)
	}

	log.Printf("AddCase: added case %s to vector store", caseMemory.ID)

	return nil
}

// Clear 清空指定会话的记忆
func (m *DefaultManager) Clear(ctx context.Context, sessionID string) error {
	if !m.config.EnableConversation {
		return fmt.Errorf("conversation memory is disabled")
	}

	if m.conversationMemory == nil {
		return fmt.Errorf("conversation memory not initialized")
	}

	return m.conversationMemory.Clear(ctx, sessionID)
}

// buildCaseText 构建用于嵌入的案例文本
func buildCaseText(caseMemory *CaseMemory) string {
	text := caseMemory.Description

	if caseMemory.RootCause != "" {
		text += " " + caseMemory.RootCause
	}

	if caseMemory.Solution != "" {
		text += " " + caseMemory.Solution
	}

	if caseMemory.FailureType != "" {
		text += " failure_type:" + caseMemory.FailureType
	}

	if caseMemory.ResourceType != "" {
		text += " resource_type:" + caseMemory.ResourceType
	}

	for _, symptom := range caseMemory.Symptoms {
		text += " " + symptom
	}

	for _, tag := range caseMemory.Tags {
		text += " #" + tag
	}

	return text
}

// generateCaseID 生成案例 ID
func generateCaseID() string {
	return fmt.Sprintf("case_%d", time.Now().UnixNano())
}

// validateManagerConfig 验证配置
func validateManagerConfig(config *ManagerConfig) error {
	if config.MaxConversationLength <= 0 {
		return fmt.Errorf("max_conversation_length must be positive")
	}

	if config.EnableVectorStore {
		if config.EmbeddingDimension <= 0 {
			return fmt.Errorf("embedding_dimension must be positive")
		}

		if config.DefaultSearchLimit <= 0 {
			return fmt.Errorf("default_search_limit must be positive")
		}

		if config.SimilarityThreshold < -1 || config.SimilarityThreshold > 1 {
			return fmt.Errorf("similarity_threshold must be between -1 and 1")
		}
	}

	return nil
}

// defaultManagerConfig 返回默认配置
func defaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		EnableConversation:    true,
		MaxConversationLength: DefaultMaxConversationLength,
		EnableVectorStore:     true,
		VectorStoreType:       VectorStoreTypeMemory,
		EmbeddingDimension:    DefaultEmbeddingDimension,
		DefaultSearchLimit:    DefaultSearchLimit,
		SimilarityThreshold:   DefaultSimilarityThreshold,
		EmbeddingModel:        "text-embedding-ada-002",
		EmbeddingProvider:     EmbedderProviderMock,
	}
}
