package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresStoreConfig holds configuration for PostgreSQL store
type PostgresStoreConfig struct {
	// DSN is the PostgreSQL Data Source Name
	// Example: "host=localhost user=postgres password=secret dbname=agent port=5432 sslmode=disable"
	DSN string

	// TableName is the name of the table to use for storage
	TableName string

	// MaxIdleConns is the maximum number of idle connections
	MaxIdleConns int

	// MaxOpenConns is the maximum number of open connections
	MaxOpenConns int

	// ConnMaxLifetime is the maximum lifetime of a connection
	ConnMaxLifetime time.Duration

	// LogLevel is the GORM log level
	LogLevel logger.LogLevel

	// AutoMigrate enables automatic table creation
	AutoMigrate bool
}

// DefaultPostgresStoreConfig returns default PostgreSQL configuration
func DefaultPostgresStoreConfig() *PostgresStoreConfig {
	return &PostgresStoreConfig{
		DSN:             "host=localhost user=postgres password=postgres dbname=agent port=5432 sslmode=disable",
		TableName:       "agent_stores",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		LogLevel:        logger.Silent,
		AutoMigrate:     true,
	}
}

// PostgresStore is a PostgreSQL-backed implementation of the Store interface.
//
// Features:
//   - JSONB storage for flexible data types
//   - Efficient indexing on namespace and key
//   - ACID compliance for data integrity
//   - Powerful search with JSONB queries
//   - Connection pooling
//
// Suitable for:
//   - Production deployments
//   - Large-scale data storage
//   - Complex queries
//   - Distributed systems with shared database
type PostgresStore struct {
	db     *gorm.DB
	config *PostgresStoreConfig
}

// storeModel represents the database schema for store values
type storeModel struct {
	ID        uint           `gorm:"primaryKey"`
	Namespace string         `gorm:"index;not null"`
	Key       string         `gorm:"index;not null"`
	Value     datatypes.JSON `gorm:"type:jsonb;not null"`
	Metadata  datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`

	// Composite unique index
	// Note: This is defined in the model but applied via AutoMigrate
}

// TableName returns the table name for the store model
func (storeModel) TableName() string {
	// This will be overridden by the store's config
	return "agent_stores"
}

// NewPostgresStore creates a new PostgreSQL-backed store
func NewPostgresStore(config *PostgresStoreConfig) (*PostgresStore, error) {
	if config == nil {
		config = DefaultPostgresStoreConfig()
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(config.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(config.LogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL database: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	store := &PostgresStore{
		db:     db,
		config: config,
	}

	// Auto-migrate if enabled
	if config.AutoMigrate {
		if err := store.migrate(); err != nil {
			return nil, fmt.Errorf("failed to migrate database: %w", err)
		}
	}

	return store, nil
}

// NewPostgresStoreFromDB creates a PostgresStore from an existing GORM DB
func NewPostgresStoreFromDB(db *gorm.DB, config *PostgresStoreConfig) (*PostgresStore, error) {
	if config == nil {
		config = DefaultPostgresStoreConfig()
	}

	store := &PostgresStore{
		db:     db,
		config: config,
	}

	if config.AutoMigrate {
		if err := store.migrate(); err != nil {
			return nil, fmt.Errorf("failed to migrate database: %w", err)
		}
	}

	return store, nil
}

// migrate creates the table and indexes
func (s *PostgresStore) migrate() error {
	// Set custom table name
	if s.config.TableName != "" {
		model := storeModel{}
		s.db.Table(s.config.TableName).AutoMigrate(&model)
	} else {
		s.db.AutoMigrate(&storeModel{})
	}

	// Create composite unique index
	return s.db.Exec(fmt.Sprintf(
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_%s_namespace_key ON %s (namespace, key)",
		s.config.TableName, s.config.TableName,
	)).Error
}

// Put stores a value with the given namespace and key
func (s *PostgresStore) Put(ctx context.Context, namespace []string, key string, value interface{}) error {
	nsKey := namespaceToKey(namespace)

	// Serialize value to JSON
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Check if exists
	var existing storeModel
	result := s.getDB().Where("namespace = ? AND key = ?", nsKey, key).First(&existing)

	now := time.Now()

	if result.Error == nil {
		// Update existing
		existing.Value = valueJSON
		existing.UpdatedAt = now

		if err := s.getDB().Save(&existing).Error; err != nil {
			return fmt.Errorf("failed to update value: %w", err)
		}
	} else if result.Error == gorm.ErrRecordNotFound {
		// Create new
		model := storeModel{
			Namespace: nsKey,
			Key:       key,
			Value:     valueJSON,
			Metadata:  []byte("{}"),
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := s.getDB().Create(&model).Error; err != nil {
			return fmt.Errorf("failed to create value: %w", err)
		}
	} else {
		return fmt.Errorf("failed to query existing value: %w", result.Error)
	}

	return nil
}

// Get retrieves a value by namespace and key
func (s *PostgresStore) Get(ctx context.Context, namespace []string, key string) (*StoreValue, error) {
	nsKey := namespaceToKey(namespace)

	var model storeModel
	result := s.getDB().Where("namespace = ? AND key = ?", nsKey, key).First(&model)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("key not found: %s", key)
		}
		return nil, fmt.Errorf("failed to get value: %w", result.Error)
	}

	// Deserialize value
	var value interface{}
	if err := json.Unmarshal(model.Value, &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	// Deserialize metadata
	var metadata map[string]interface{}
	if len(model.Metadata) > 0 {
		if err := json.Unmarshal(model.Metadata, &metadata); err != nil {
			metadata = make(map[string]interface{})
		}
	} else {
		metadata = make(map[string]interface{})
	}

	return &StoreValue{
		Value:     value,
		Metadata:  metadata,
		Created:   model.CreatedAt,
		Updated:   model.UpdatedAt,
		Namespace: namespace,
		Key:       key,
	}, nil
}

// Delete removes a value by namespace and key
func (s *PostgresStore) Delete(ctx context.Context, namespace []string, key string) error {
	nsKey := namespaceToKey(namespace)

	result := s.getDB().Where("namespace = ? AND key = ?", nsKey, key).Delete(&storeModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete value: %w", result.Error)
	}

	return nil
}

// Search finds values matching the filter within a namespace
func (s *PostgresStore) Search(ctx context.Context, namespace []string, filter map[string]interface{}) ([]*StoreValue, error) {
	nsKey := namespaceToKey(namespace)

	query := s.getDB().Where("namespace = ?", nsKey)

	// Apply metadata filters using JSONB queries
	for key, value := range filter {
		// Use JSONB contains query
		filterJSON := map[string]interface{}{key: value}
		filterBytes, err := json.Marshal(filterJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal filter: %w", err)
		}

		query = query.Where("metadata @> ?", filterBytes)
	}

	var models []storeModel
	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to search values: %w", err)
	}

	// Convert models to StoreValues
	results := make([]*StoreValue, 0, len(models))
	for _, model := range models {
		var value interface{}
		if err := json.Unmarshal(model.Value, &value); err != nil {
			continue // Skip invalid values
		}

		var metadata map[string]interface{}
		if len(model.Metadata) > 0 {
			json.Unmarshal(model.Metadata, &metadata)
		} else {
			metadata = make(map[string]interface{})
		}

		results = append(results, &StoreValue{
			Value:     value,
			Metadata:  metadata,
			Created:   model.CreatedAt,
			Updated:   model.UpdatedAt,
			Namespace: namespace,
			Key:       model.Key,
		})
	}

	return results, nil
}

// List returns all keys within a namespace
func (s *PostgresStore) List(ctx context.Context, namespace []string) ([]string, error) {
	nsKey := namespaceToKey(namespace)

	var keys []string
	err := s.getDB().
		Model(&storeModel{}).
		Where("namespace = ?", nsKey).
		Pluck("key", &keys).
		Error
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	return keys, nil
}

// Clear removes all values within a namespace
func (s *PostgresStore) Clear(ctx context.Context, namespace []string) error {
	nsKey := namespaceToKey(namespace)

	result := s.getDB().Where("namespace = ?", nsKey).Delete(&storeModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to clear namespace: %w", result.Error)
	}

	return nil
}

// Close closes the database connection
func (s *PostgresStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// getDB returns the DB instance with custom table name if configured
func (s *PostgresStore) getDB() *gorm.DB {
	if s.config.TableName != "" {
		return s.db.Table(s.config.TableName)
	}
	return s.db
}

// Ping tests the connection to PostgreSQL
func (s *PostgresStore) Ping(ctx context.Context) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// Size returns the total number of stored values
func (s *PostgresStore) Size(ctx context.Context) (int64, error) {
	var count int64
	err := s.getDB().Model(&storeModel{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count values: %w", err)
	}
	return count, nil
}
