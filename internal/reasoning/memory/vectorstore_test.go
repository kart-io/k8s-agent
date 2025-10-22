package memory

import (
	"context"
	"fmt"
	"math"
	"testing"
)

func TestNewInMemoryVectorStore(t *testing.T) {
	tests := []struct {
		name    string
		config  *VectorStoreConfig
		wantErr bool
	}{
		{
			name:    "nil config uses defaults",
			config:  nil,
			wantErr: false,
		},
		{
			name: "valid config",
			config: &VectorStoreConfig{
				Type:      VectorStoreTypeMemory,
				Dimension: 128,
			},
			wantErr: false,
		},
		{
			name: "invalid dimension",
			config: &VectorStoreConfig{
				Type:      VectorStoreTypeMemory,
				Dimension: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vs, err := NewInMemoryVectorStore(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("NewInMemoryVectorStore() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewInMemoryVectorStore() unexpected error: %v", err)
				return
			}

			if vs == nil {
				t.Error("NewInMemoryVectorStore() returned nil")
			}
		})
	}
}

func TestInMemoryVectorStore_Add(t *testing.T) {
	vs, err := NewInMemoryVectorStore(&VectorStoreConfig{
		Type:      VectorStoreTypeMemory,
		Dimension: 3,
	})
	if err != nil {
		t.Fatalf("Failed to create VectorStore: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name      string
		id        string
		embedding []float64
		metadata  map[string]interface{}
		wantErr   bool
	}{
		{
			name:      "valid vector",
			id:        "vec-1",
			embedding: []float64{1.0, 2.0, 3.0},
			metadata:  map[string]interface{}{"type": "test"},
			wantErr:   false,
		},
		{
			name:      "empty id",
			id:        "",
			embedding: []float64{1.0, 2.0, 3.0},
			metadata:  nil,
			wantErr:   true,
		},
		{
			name:      "nil embedding",
			id:        "vec-2",
			embedding: nil,
			metadata:  nil,
			wantErr:   true,
		},
		{
			name:      "empty embedding",
			id:        "vec-3",
			embedding: []float64{},
			metadata:  nil,
			wantErr:   true,
		},
		{
			name:      "wrong dimension",
			id:        "vec-4",
			embedding: []float64{1.0, 2.0},
			metadata:  nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vs.Add(ctx, tt.id, tt.embedding, tt.metadata)

			if tt.wantErr && err == nil {
				t.Error("Add() expected error but got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Add() unexpected error: %v", err)
			}
		})
	}
}

func TestInMemoryVectorStore_Search(t *testing.T) {
	vs, err := NewInMemoryVectorStore(&VectorStoreConfig{
		Type:      VectorStoreTypeMemory,
		Dimension: 3,
	})
	if err != nil {
		t.Fatalf("Failed to create VectorStore: %v", err)
	}

	ctx := context.Background()

	// 添加一些向量
	vectors := []struct {
		id        string
		embedding []float64
		metadata  map[string]interface{}
	}{
		{"vec-1", []float64{1.0, 0.0, 0.0}, map[string]interface{}{"type": "A"}},
		{"vec-2", []float64{0.9, 0.1, 0.0}, map[string]interface{}{"type": "A"}},
		{"vec-3", []float64{0.0, 1.0, 0.0}, map[string]interface{}{"type": "B"}},
		{"vec-4", []float64{0.0, 0.0, 1.0}, map[string]interface{}{"type": "C"}},
	}

	for _, v := range vectors {
		if err := vs.Add(ctx, v.id, v.embedding, v.metadata); err != nil {
			t.Fatalf("Failed to add vector: %v", err)
		}
	}

	tests := []struct {
		name        string
		embedding   []float64
		limit       int
		wantCount   int
		wantFirstID string
		wantErr     bool
	}{
		{
			name:        "search similar to vec-1",
			embedding:   []float64{1.0, 0.0, 0.0},
			limit:       2,
			wantCount:   2,
			wantFirstID: "vec-1",
			wantErr:     false,
		},
		{
			name:        "search all",
			embedding:   []float64{1.0, 0.0, 0.0},
			limit:       0,
			wantCount:   4,
			wantFirstID: "vec-1",
			wantErr:     false,
		},
		{
			name:      "nil embedding",
			embedding: nil,
			limit:     2,
			wantErr:   true,
		},
		{
			name:      "wrong dimension",
			embedding: []float64{1.0, 0.0},
			limit:     2,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := vs.Search(ctx, tt.embedding, tt.limit)

			if tt.wantErr {
				if err == nil {
					t.Error("Search() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Search() unexpected error: %v", err)
				return
			}

			if len(results) != tt.wantCount {
				t.Errorf("Search() returned %d results, want %d", len(results), tt.wantCount)
			}

			if len(results) > 0 && results[0].ID != tt.wantFirstID {
				t.Errorf("Search() first result ID = %s, want %s", results[0].ID, tt.wantFirstID)
			}

			// 验证结果按相似度降序排列
			for i := 1; i < len(results); i++ {
				if results[i].Score > results[i-1].Score {
					t.Error("Search() results should be sorted by score in descending order")
				}
			}
		})
	}
}

func TestInMemoryVectorStore_Delete(t *testing.T) {
	vs, err := NewInMemoryVectorStore(&VectorStoreConfig{
		Type:      VectorStoreTypeMemory,
		Dimension: 3,
	})
	if err != nil {
		t.Fatalf("Failed to create VectorStore: %v", err)
	}

	ctx := context.Background()

	// 添加向量
	err = vs.Add(ctx, "vec-1", []float64{1.0, 2.0, 3.0}, nil)
	if err != nil {
		t.Fatalf("Failed to add vector: %v", err)
	}

	// 删除向量
	err = vs.Delete(ctx, "vec-1")
	if err != nil {
		t.Errorf("Delete() unexpected error: %v", err)
	}

	// 验证向量已删除
	results, err := vs.Search(ctx, []float64{1.0, 2.0, 3.0}, 10)
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results after delete, got %d", len(results))
	}

	// 删除不存在的向量应返回错误
	err = vs.Delete(ctx, "non-existent")
	if err == nil {
		t.Error("Delete() should return error for non-existent vector")
	}
}

func TestInMemoryVectorStore_Clear(t *testing.T) {
	vs, err := NewInMemoryVectorStore(&VectorStoreConfig{
		Type:      VectorStoreTypeMemory,
		Dimension: 3,
	})
	if err != nil {
		t.Fatalf("Failed to create VectorStore: %v", err)
	}

	ctx := context.Background()

	// 添加多个向量
	for i := 0; i < 5; i++ {
		err = vs.Add(ctx, fmt.Sprintf("vec-%d", i), []float64{float64(i), 0.0, 0.0}, nil)
		if err != nil {
			t.Fatalf("Failed to add vector: %v", err)
		}
	}

	// 验证向量存在
	results, err := vs.Search(ctx, []float64{1.0, 0.0, 0.0}, 10)
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(results) != 5 {
		t.Fatalf("Expected 5 vectors, got %d", len(results))
	}

	// 清空存储
	err = vs.Clear(ctx)
	if err != nil {
		t.Errorf("Clear() unexpected error: %v", err)
	}

	// 验证所有向量已删除
	results, err = vs.Search(ctx, []float64{1.0, 0.0, 0.0}, 10)
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results after clear, got %d", len(results))
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
	}{
		{
			name:     "identical vectors",
			a:        []float64{1.0, 0.0, 0.0},
			b:        []float64{1.0, 0.0, 0.0},
			expected: 1.0,
		},
		{
			name:     "orthogonal vectors",
			a:        []float64{1.0, 0.0, 0.0},
			b:        []float64{0.0, 1.0, 0.0},
			expected: 0.0,
		},
		{
			name:     "opposite vectors",
			a:        []float64{1.0, 0.0, 0.0},
			b:        []float64{-1.0, 0.0, 0.0},
			expected: -1.0,
		},
		{
			name:     "similar vectors",
			a:        []float64{1.0, 0.0, 0.0},
			b:        []float64{0.9, 0.1, 0.0},
			expected: 0.9938,
		},
		{
			name:     "different lengths",
			a:        []float64{1.0, 0.0},
			b:        []float64{1.0, 0.0, 0.0},
			expected: 0.0,
		},
		{
			name:     "zero vector",
			a:        []float64{0.0, 0.0, 0.0},
			b:        []float64{1.0, 0.0, 0.0},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cosineSimilarity(tt.a, tt.b)

			// 允许小的浮点误差
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("cosineSimilarity() = %f, want %f", result, tt.expected)
			}
		})
	}
}

func TestInMemoryVectorStore_DataIsolation(t *testing.T) {
	vs, err := NewInMemoryVectorStore(&VectorStoreConfig{
		Type:      VectorStoreTypeMemory,
		Dimension: 3,
	})
	if err != nil {
		t.Fatalf("Failed to create VectorStore: %v", err)
	}

	ctx := context.Background()

	// 原始数据
	embedding := []float64{1.0, 2.0, 3.0}
	metadata := map[string]interface{}{"key": "value"}

	// 添加向量
	err = vs.Add(ctx, "vec-1", embedding, metadata)
	if err != nil {
		t.Fatalf("Failed to add vector: %v", err)
	}

	// 修改原始数据
	embedding[0] = 999.0
	metadata["key"] = "modified"

	// 搜索并验证存储的数据未被修改
	results, err := vs.Search(ctx, []float64{1.0, 2.0, 3.0}, 1)
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Embedding[0] != 1.0 {
		t.Errorf("Embedding[0] = %f, want 1.0 (data should be isolated)", results[0].Embedding[0])
	}

	if results[0].Metadata["key"] != "value" {
		t.Errorf("Metadata[key] = %v, want 'value' (data should be isolated)", results[0].Metadata["key"])
	}

	// 修改搜索结果
	results[0].Embedding[0] = 888.0
	results[0].Metadata["key"] = "modified_again"

	// 再次搜索并验证存储的数据未被修改
	results2, err := vs.Search(ctx, []float64{1.0, 2.0, 3.0}, 1)
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}

	if results2[0].Embedding[0] != 1.0 {
		t.Errorf("Embedding[0] = %f, want 1.0 (stored data should be protected)", results2[0].Embedding[0])
	}

	if results2[0].Metadata["key"] != "value" {
		t.Errorf("Metadata[key] = %v, want 'value' (stored data should be protected)", results2[0].Metadata["key"])
	}
}
