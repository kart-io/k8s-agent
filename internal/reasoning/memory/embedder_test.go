package memory

import (
	"context"
	"math"
	"testing"
)

func TestNewMockEmbedder(t *testing.T) {
	tests := []struct {
		name      string
		dimension int
		wantErr   bool
	}{
		{
			name:      "valid dimension",
			dimension: 128,
			wantErr:   false,
		},
		{
			name:      "zero dimension",
			dimension: 0,
			wantErr:   true,
		},
		{
			name:      "negative dimension",
			dimension: -1,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			embedder, err := NewMockEmbedder(tt.dimension)

			if tt.wantErr {
				if err == nil {
					t.Error("NewMockEmbedder() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewMockEmbedder() unexpected error: %v", err)
				return
			}

			if embedder == nil {
				t.Error("NewMockEmbedder() returned nil")
			}
		})
	}
}

func TestMockEmbedder_Embed(t *testing.T) {
	embedder, err := NewMockEmbedder(128)
	if err != nil {
		t.Fatalf("Failed to create embedder: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{
			name:    "valid text",
			text:    "Hello world",
			wantErr: false,
		},
		{
			name:    "empty text",
			text:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			embedding, err := embedder.Embed(ctx, tt.text)

			if tt.wantErr {
				if err == nil {
					t.Error("Embed() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Embed() unexpected error: %v", err)
				return
			}

			if len(embedding) != 128 {
				t.Errorf("Embed() returned embedding of length %d, want 128", len(embedding))
			}

			// 验证向量是归一化的
			var norm float64
			for _, val := range embedding {
				norm += val * val
			}
			norm = math.Sqrt(norm)

			if math.Abs(norm-1.0) > 0.0001 {
				t.Errorf("Embedding norm = %f, want 1.0 (should be normalized)", norm)
			}
		})
	}
}

func TestMockEmbedder_Deterministic(t *testing.T) {
	embedder, err := NewMockEmbedder(128)
	if err != nil {
		t.Fatalf("Failed to create embedder: %v", err)
	}

	ctx := context.Background()
	text := "Hello world"

	// 生成两次嵌入
	embedding1, err := embedder.Embed(ctx, text)
	if err != nil {
		t.Fatalf("Embed() error: %v", err)
	}

	embedding2, err := embedder.Embed(ctx, text)
	if err != nil {
		t.Fatalf("Embed() error: %v", err)
	}

	// 验证两次结果相同（确定性）
	if len(embedding1) != len(embedding2) {
		t.Fatal("Embeddings have different lengths")
	}

	for i := 0; i < len(embedding1); i++ {
		if embedding1[i] != embedding2[i] {
			t.Errorf("Embedding mismatch at index %d: %f != %f", i, embedding1[i], embedding2[i])
		}
	}
}

func TestMockEmbedder_DifferentTexts(t *testing.T) {
	embedder, err := NewMockEmbedder(128)
	if err != nil {
		t.Fatalf("Failed to create embedder: %v", err)
	}

	ctx := context.Background()

	// 生成不同文本的嵌入
	embedding1, err := embedder.Embed(ctx, "Hello world")
	if err != nil {
		t.Fatalf("Embed() error: %v", err)
	}

	embedding2, err := embedder.Embed(ctx, "Goodbye world")
	if err != nil {
		t.Fatalf("Embed() error: %v", err)
	}

	// 验证不同文本产生不同嵌入
	same := true
	for i := 0; i < len(embedding1); i++ {
		if embedding1[i] != embedding2[i] {
			same = false
			break
		}
	}

	if same {
		t.Error("Different texts should produce different embeddings")
	}
}

func TestMockEmbedder_EmbedBatch(t *testing.T) {
	embedder, err := NewMockEmbedder(128)
	if err != nil {
		t.Fatalf("Failed to create embedder: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name      string
		texts     []string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "valid texts",
			texts:     []string{"Hello", "World", "Test"},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "single text",
			texts:     []string{"Hello"},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:    "empty slice",
			texts:   []string{},
			wantErr: true,
		},
		{
			name:    "nil slice",
			texts:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			embeddings, err := embedder.EmbedBatch(ctx, tt.texts)

			if tt.wantErr {
				if err == nil {
					t.Error("EmbedBatch() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("EmbedBatch() unexpected error: %v", err)
				return
			}

			if len(embeddings) != tt.wantCount {
				t.Errorf("EmbedBatch() returned %d embeddings, want %d", len(embeddings), tt.wantCount)
			}

			// 验证每个嵌入的长度
			for i, embedding := range embeddings {
				if len(embedding) != 128 {
					t.Errorf("Embedding %d has length %d, want 128", i, len(embedding))
				}
			}
		})
	}
}

func TestMockEmbedder_BatchConsistency(t *testing.T) {
	embedder, err := NewMockEmbedder(128)
	if err != nil {
		t.Fatalf("Failed to create embedder: %v", err)
	}

	ctx := context.Background()
	texts := []string{"Hello", "World"}

	// 批量生成
	batchEmbeddings, err := embedder.EmbedBatch(ctx, texts)
	if err != nil {
		t.Fatalf("EmbedBatch() error: %v", err)
	}

	// 单独生成
	for i, text := range texts {
		singleEmbedding, err := embedder.Embed(ctx, text)
		if err != nil {
			t.Fatalf("Embed() error: %v", err)
		}

		// 验证批量和单独生成的结果一致
		for j := 0; j < len(singleEmbedding); j++ {
			if singleEmbedding[j] != batchEmbeddings[i][j] {
				t.Errorf("Embedding mismatch for text %d at index %d: %f != %f",
					i, j, singleEmbedding[j], batchEmbeddings[i][j])
			}
		}
	}
}

func TestNormalizeVector(t *testing.T) {
	tests := []struct {
		name     string
		vector   []float64
		wantNorm float64
	}{
		{
			name:     "non-normalized vector",
			vector:   []float64{3.0, 4.0, 0.0},
			wantNorm: 1.0,
		},
		{
			name:     "already normalized",
			vector:   []float64{1.0, 0.0, 0.0},
			wantNorm: 1.0,
		},
		{
			name:     "negative values",
			vector:   []float64{-3.0, 4.0, 0.0},
			wantNorm: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized := normalizeVector(tt.vector)

			// 计算范数
			var norm float64
			for _, val := range normalized {
				norm += val * val
			}
			norm = math.Sqrt(norm)

			if math.Abs(norm-tt.wantNorm) > 0.0001 {
				t.Errorf("normalizeVector() norm = %f, want %f", norm, tt.wantNorm)
			}
		})
	}
}
