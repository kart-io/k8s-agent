package memory

import (
	"context"
	"fmt"
	"hash/fnv"
	"math"
)

// MockEmbedder Mock 嵌入生成器（用于测试）
type MockEmbedder struct {
	dimension int
}

// NewMockEmbedder 创建新的 Mock 嵌入生成器
func NewMockEmbedder(dimension int) (*MockEmbedder, error) {
	if dimension <= 0 {
		return nil, fmt.Errorf("dimension must be positive")
	}

	return &MockEmbedder{
		dimension: dimension,
	}, nil
}

// Embed 生成文本嵌入
// 使用简单的哈希算法生成确定性的嵌入向量
func (e *MockEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
	if text == "" {
		return nil, fmt.Errorf("text is required")
	}

	// 使用 FNV 哈希生成确定性的嵌入
	embedding := make([]float64, e.dimension)

	// 为每个维度生成一个值
	for i := 0; i < e.dimension; i++ {
		h := fnv.New64a()
		// 混合文本和维度索引
		h.Write([]byte(fmt.Sprintf("%s-%d", text, i)))
		hashValue := h.Sum64()

		// 将哈希值归一化到 [-1, 1] 范围
		embedding[i] = float64(hashValue%10000)/10000.0*2.0 - 1.0
	}

	// 归一化向量
	embedding = normalizeVector(embedding)

	return embedding, nil
}

// EmbedBatch 批量生成嵌入
func (e *MockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts is required")
	}

	embeddings := make([][]float64, len(texts))

	for i, text := range texts {
		embedding, err := e.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to embed text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// normalizeVector 归一化向量到单位长度
func normalizeVector(v []float64) []float64 {
	var norm float64
	for _, val := range v {
		norm += val * val
	}
	norm = math.Sqrt(norm)

	if norm == 0 {
		return v
	}

	normalized := make([]float64, len(v))
	for i, val := range v {
		normalized[i] = val / norm
	}

	return normalized
}
