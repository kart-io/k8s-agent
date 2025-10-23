package config

import "errors"

var (
	// ErrNoLLMProviders is returned when LLM is enabled but no providers are configured
	ErrNoLLMProviders = errors.New("LLM is enabled but no providers configured")

	// ErrInvalidVectorStoreType is returned when an invalid vector store type is specified
	ErrInvalidVectorStoreType = errors.New("vector_store_type is required when vector store is enabled")

	// ErrInvalidMinConfidence is returned when min_confidence is out of range
	ErrInvalidMinConfidence = errors.New("invalid min_confidence: must be between 0 and 1")

	// ErrInvalidMaxRecommendations is returned when max_recommendations is less than 1
	ErrInvalidMaxRecommendations = errors.New("max_recommendations must be at least 1")
)
