package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// MemoryOptions 内存/向量存储配置选项
type MemoryOptions struct {
	EnableVectorStore bool   `mapstructure:"enable_vector_store" yaml:"enable_vector_store" json:"enable_vector_store"` // 启用向量存储
	VectorStoreType   string `mapstructure:"vector_store_type" yaml:"vector_store_type" json:"vector_store_type"`       // 向量存储类型: "chroma", "pinecone", "weaviate"
	VectorStorePath   string `mapstructure:"vector_store_path" yaml:"vector_store_path" json:"vector_store_path"`       // 向量存储路径
	EmbeddingModel    string `mapstructure:"embedding_model" yaml:"embedding_model" json:"embedding_model"`             // Embedding 模型
	EmbeddingProvider string `mapstructure:"embedding_provider" yaml:"embedding_provider" json:"embedding_provider"`    // Embedding 提供商: "openai", "local"
}

// NewMemoryOptions 创建默认的内存配置
func NewMemoryOptions() *MemoryOptions {
	return &MemoryOptions{
		EnableVectorStore: false,
		VectorStoreType:   "chroma",
		VectorStorePath:   "./data/vectorstore",
		EmbeddingModel:    "text-embedding-ada-002",
		EmbeddingProvider: "openai",
	}
}

// Validate 验证配置
func (o *MemoryOptions) Validate() error {
	if !o.EnableVectorStore {
		return nil // Vector store disabled, no validation needed
	}

	if o.VectorStoreType == "" {
		return fmt.Errorf("vector_store_type is required when vector store is enabled")
	}

	// 验证支持的向量存储类型
	validTypes := []string{"chroma", "pinecone", "weaviate", "qdrant", "milvus"}
	valid := false
	for _, t := range validTypes {
		if o.VectorStoreType == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid vector_store_type: %s (must be one of: chroma, pinecone, weaviate, qdrant, milvus)", o.VectorStoreType)
	}

	if o.EmbeddingProvider == "" {
		return fmt.Errorf("embedding_provider is required when vector store is enabled")
	}

	// 验证 embedding provider
	validProviders := []string{"openai", "local", "huggingface", "cohere"}
	valid = false
	for _, p := range validProviders {
		if o.EmbeddingProvider == p {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid embedding_provider: %s (must be one of: openai, local, huggingface, cohere)", o.EmbeddingProvider)
	}

	return nil
}

// AddFlags 添加命令行参数
func (o *MemoryOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.EnableVectorStore, "memory.enable-vector-store", o.EnableVectorStore, "Enable vector store")
	fs.StringVar(&o.VectorStoreType, "memory.vector-store-type", o.VectorStoreType, "Vector store type (chroma, pinecone, weaviate, qdrant, milvus)")
	fs.StringVar(&o.VectorStorePath, "memory.vector-store-path", o.VectorStorePath, "Vector store path")
	fs.StringVar(&o.EmbeddingModel, "memory.embedding-model", o.EmbeddingModel, "Embedding model")
	fs.StringVar(&o.EmbeddingProvider, "memory.embedding-provider", o.EmbeddingProvider, "Embedding provider (openai, local, huggingface, cohere)")
}

// ApplyTo 将配置应用到目标接口
func (o *MemoryOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	switch v := target.(type) {
	case *[]interface{}:
		*v = append(*v,
			map[string]interface{}{
				"enableVectorStore": o.EnableVectorStore,
				"vectorStoreType":   o.VectorStoreType,
				"vectorStorePath":   o.VectorStorePath,
				"embeddingModel":    o.EmbeddingModel,
				"embeddingProvider": o.EmbeddingProvider,
			},
		)
	}

	return nil
}

// Complete 完成配置初始化
func (o *MemoryOptions) Complete() error {
	if o.VectorStoreType == "" {
		o.VectorStoreType = "chroma"
	}

	if o.VectorStorePath == "" {
		o.VectorStorePath = "./data/vectorstore"
	}

	if o.EmbeddingModel == "" {
		o.EmbeddingModel = "text-embedding-ada-002"
	}

	if o.EmbeddingProvider == "" {
		o.EmbeddingProvider = "openai"
	}

	return nil
}

// WithVectorStoreEnabled 设置是否启用向量存储
func WithVectorStoreEnabled(enabled bool) func(*MemoryOptions) {
	return func(o *MemoryOptions) {
		o.EnableVectorStore = enabled
	}
}

// WithVectorStoreType 设置向量存储类型
func WithVectorStoreType(storeType string) func(*MemoryOptions) {
	return func(o *MemoryOptions) {
		o.VectorStoreType = storeType
	}
}

// WithVectorStorePath 设置向量存储路径
func WithVectorStorePath(path string) func(*MemoryOptions) {
	return func(o *MemoryOptions) {
		o.VectorStorePath = path
	}
}

// WithEmbeddingModel 设置 Embedding 模型
func WithEmbeddingModel(model string) func(*MemoryOptions) {
	return func(o *MemoryOptions) {
		o.EmbeddingModel = model
	}
}

// WithEmbeddingProvider 设置 Embedding 提供商
func WithEmbeddingProvider(provider string) func(*MemoryOptions) {
	return func(o *MemoryOptions) {
		o.EmbeddingProvider = provider
	}
}
