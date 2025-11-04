package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/kart-io/k8s-agent/internal/auth/crypto"
	"github.com/kart-io/k8s-agent/internal/auth/model"
	"github.com/kart-io/k8s-agent/internal/auth/storage"
	"github.com/kart-io/k8s-agent/internal/auth/types"
)

// APIKeyService handles API key management business logic
type APIKeyService struct {
	db *storage.PostgresDB
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(db *storage.PostgresDB) *APIKeyService {
	return &APIKeyService{db: db}
}

// List retrieves API keys for a user (secrets are masked)
func (s *APIKeyService) List(userID string) ([]types.APIKey, error) {
	var modelKeys []model.APIKey
	err := s.db.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&modelKeys).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}

	// Convert model.APIKey to types.APIKey
	keys := make([]types.APIKey, len(modelKeys))
	for i, k := range modelKeys {
		keys[i] = types.APIKey{
			ID:          k.ID,
			Name:        k.Name,
			Key:         k.Key,
			Secret:      "****", // Mask secret
			UserID:      k.UserID,
			Description: k.Description,
			ExpiresAt:   k.ExpiresAt,
			Status:      k.Status,
			LastUsedAt:  k.LastUsedAt,
			CreatedAt:   k.CreatedAt,
			UpdatedAt:   k.UpdatedAt,
		}
	}

	return keys, nil
}

// Create creates a new API key and returns it with the plain secret (shown only once)
func (s *APIKeyService) Create(userID string, req *types.APIKeyCreateRequest) (*types.APIKeyWithSecret, error) {
	// Generate API key and secret
	keyID := uuid.New().String()
	apiKey, err := generateKey("ak")
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	secretPlain, err := generateKey("sk")
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	// Hash secret before storage
	secretHashed, err := crypto.HashPassword(secretPlain)
	if err != nil {
		return nil, fmt.Errorf("failed to hash secret: %w", err)
	}

	// Create API key model
	modelKey := &model.APIKey{
		ID:          keyID,
		Name:        req.Name,
		Key:         apiKey,
		Secret:      secretHashed,
		UserID:      userID,
		Description: req.Description,
		Status:      1,
	}

	if !req.ExpiresAt.IsZero() {
		modelKey.ExpiresAt = &req.ExpiresAt
	}

	// Insert API key
	if err := s.db.DB.Create(modelKey).Error; err != nil {
		return nil, fmt.Errorf("failed to insert API key: %w", err)
	}

	// Get the created key
	var createdKey model.APIKey
	if err := s.db.DB.Where("id = ?", keyID).First(&createdKey).Error; err != nil {
		return nil, fmt.Errorf("failed to query created API key: %w", err)
	}

	// Convert to types.APIKey
	apiKeyType := types.APIKey{
		ID:          createdKey.ID,
		Name:        createdKey.Name,
		Key:         createdKey.Key,
		Secret:      "****",
		UserID:      createdKey.UserID,
		Description: createdKey.Description,
		ExpiresAt:   createdKey.ExpiresAt,
		Status:      createdKey.Status,
		LastUsedAt:  createdKey.LastUsedAt,
		CreatedAt:   createdKey.CreatedAt,
		UpdatedAt:   createdKey.UpdatedAt,
	}

	// Return with plain secret (shown only once)
	return &types.APIKeyWithSecret{
		APIKey:      apiKeyType,
		SecretPlain: secretPlain,
	}, nil
}

// Delete deletes an API key
func (s *APIKeyService) Delete(id, userID string) error {
	result := s.db.DB.Delete(&model.APIKey{}, "id = ? AND user_id = ?", id, userID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete API key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}

	return nil
}

// Validate validates an API key and secret
func (s *APIKeyService) Validate(key, secret string) (*types.APIKey, error) {
	var modelKey model.APIKey
	err := s.db.DB.Where("key = ? AND status = ?", key, 1).First(&modelKey).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("invalid API key")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query API key: %w", err)
	}

	// Check expiration
	if modelKey.ExpiresAt != nil && time.Now().After(*modelKey.ExpiresAt) {
		return nil, fmt.Errorf("API key has expired")
	}

	// Verify secret
	if err := crypto.CheckPassword(modelKey.Secret, secret); err != nil {
		return nil, fmt.Errorf("invalid API secret")
	}

	// Update last_used_at
	now := time.Now()
	if err := s.db.DB.Model(&model.APIKey{}).Where("id = ?", modelKey.ID).Update("last_used_at", now).Error; err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to update last_used_at: %v\n", err)
	}

	// Convert to types.APIKey
	apiKey := &types.APIKey{
		ID:          modelKey.ID,
		Name:        modelKey.Name,
		Key:         modelKey.Key,
		Secret:      "****", // Mask secret
		UserID:      modelKey.UserID,
		Description: modelKey.Description,
		ExpiresAt:   modelKey.ExpiresAt,
		Status:      modelKey.Status,
		LastUsedAt:  modelKey.LastUsedAt,
		CreatedAt:   modelKey.CreatedAt,
		UpdatedAt:   modelKey.UpdatedAt,
	}

	return apiKey, nil
}

// CleanupExpired removes expired API keys (background task)
func (s *APIKeyService) CleanupExpired() (int, error) {
	result := s.db.DB.Where("expires_at < ?", time.Now()).Delete(&model.APIKey{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup expired keys: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}

// generateKey generates a random key with the given prefix
func generateKey(prefix string) (string, error) {
	bytes := make([]byte, 16) // 32 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(bytes), nil
}
