package memory

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
)

// InMemoryVectorStore 内存向量存储实现
type InMemoryVectorStore struct {
	// vectors 存储向量数据
	vectors map[string]*vectorEntry

	// config 配置
	config *VectorStoreConfig

	// mu 互斥锁
	mu sync.RWMutex
}

// vectorEntry 向量条目
type vectorEntry struct {
	ID        string
	Embedding []float64
	Metadata  map[string]interface{}
}

// NewInMemoryVectorStore 创建新的内存向量存储
func NewInMemoryVectorStore(config *VectorStoreConfig) (*InMemoryVectorStore, error) {
	if config == nil {
		config = &VectorStoreConfig{
			Type:      VectorStoreTypeMemory,
			Dimension: DefaultEmbeddingDimension,
		}
	}

	// 验证配置
	if config.Dimension <= 0 {
		return nil, fmt.Errorf("dimension must be positive")
	}

	vs := &InMemoryVectorStore{
		vectors: make(map[string]*vectorEntry),
		config:  config,
	}

	return vs, nil
}

// Add 添加向量
func (vs *InMemoryVectorStore) Add(ctx context.Context, id string, embedding []float64, metadata map[string]interface{}) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}

	if embedding == nil || len(embedding) == 0 {
		return fmt.Errorf("embedding is required")
	}

	if len(embedding) != vs.config.Dimension {
		return fmt.Errorf("embedding dimension mismatch: expected %d, got %d", vs.config.Dimension, len(embedding))
	}

	vs.mu.Lock()
	defer vs.mu.Unlock()

	// 复制嵌入向量以避免外部修改
	embeddingCopy := make([]float64, len(embedding))
	copy(embeddingCopy, embedding)

	// 复制元数据
	metadataCopy := make(map[string]interface{})
	for k, v := range metadata {
		metadataCopy[k] = v
	}

	vs.vectors[id] = &vectorEntry{
		ID:        id,
		Embedding: embeddingCopy,
		Metadata:  metadataCopy,
	}

	return nil
}

// Search 搜索相似向量
func (vs *InMemoryVectorStore) Search(ctx context.Context, embedding []float64, limit int) ([]*SearchResult, error) {
	if embedding == nil || len(embedding) == 0 {
		return nil, fmt.Errorf("embedding is required")
	}

	if len(embedding) != vs.config.Dimension {
		return nil, fmt.Errorf("embedding dimension mismatch: expected %d, got %d", vs.config.Dimension, len(embedding))
	}

	if limit <= 0 {
		limit = DefaultSearchLimit
	}

	vs.mu.RLock()
	defer vs.mu.RUnlock()

	// 计算所有向量的相似度
	results := make([]*SearchResult, 0, len(vs.vectors))

	for _, entry := range vs.vectors {
		similarity := cosineSimilarity(embedding, entry.Embedding)

		// 复制元数据
		metadataCopy := make(map[string]interface{})
		for k, v := range entry.Metadata {
			metadataCopy[k] = v
		}

		// 复制嵌入向量
		embeddingCopy := make([]float64, len(entry.Embedding))
		copy(embeddingCopy, entry.Embedding)

		results = append(results, &SearchResult{
			ID:        entry.ID,
			Score:     similarity,
			Embedding: embeddingCopy,
			Metadata:  metadataCopy,
		})
	}

	// 按相似度排序（降序）
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 限制返回数量
	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}

	return results, nil
}

// Delete 删除向量
func (vs *InMemoryVectorStore) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}

	vs.mu.Lock()
	defer vs.mu.Unlock()

	if _, exists := vs.vectors[id]; !exists {
		return fmt.Errorf("vector with id %s not found", id)
	}

	delete(vs.vectors, id)

	return nil
}

// Clear 清空存储
func (vs *InMemoryVectorStore) Clear(ctx context.Context) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	vs.vectors = make(map[string]*vectorEntry)

	return nil
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
