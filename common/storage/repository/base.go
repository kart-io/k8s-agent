package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var (
	// ErrNotFound is returned when an entity is not found.
	ErrNotFound = errors.New("entity not found")
	// ErrAlreadyExists is returned when trying to create an entity that already exists.
	ErrAlreadyExists = errors.New("entity already exists")
)

// BaseRepository provides a base CRUD implementation using GORM.
// It can be embedded in service-specific repositories to inherit common functionality.
//
// Example usage:
//
//	type UserRepository struct {
//	    *repository.BaseRepository[User, string]
//	}
//
//	func NewUserRepository(db *gorm.DB) *UserRepository {
//	    return &UserRepository{
//	        BaseRepository: repository.NewBaseRepository[User, string](db),
//	    }
//	}
type BaseRepository[T any, ID comparable] struct {
	db *gorm.DB
}

// NewBaseRepository creates a new base repository.
func NewBaseRepository[T any, ID comparable](db *gorm.DB) *BaseRepository[T, ID] {
	return &BaseRepository[T, ID]{
		db: db,
	}
}

// Create creates a new entity in the database.
func (r *BaseRepository[T, ID]) Create(ctx context.Context, entity *T) error {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return fmt.Errorf("failed to create entity: %w", result.Error)
	}
	return nil
}

// Get retrieves an entity by its ID.
func (r *BaseRepository[T, ID]) Get(ctx context.Context, id ID) (*T, error) {
	var entity T
	result := r.db.WithContext(ctx).First(&entity, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get entity: %w", result.Error)
	}
	return &entity, nil
}

// List retrieves multiple entities based on options.
func (r *BaseRepository[T, ID]) List(ctx context.Context, opts *ListOptions) ([]*T, int64, error) {
	if opts == nil {
		opts = NewListOptions()
	}
	opts.Validate()

	var entities []*T
	var total int64

	// Build query
	query := r.db.WithContext(ctx).Model(new(T))

	// Apply filters
	for field, value := range opts.Filters {
		query = query.Where(fmt.Sprintf("%s = ?", field), value)
	}

	// Count total (before pagination)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count entities: %w", err)
	}

	// Apply sorting
	orderClause := fmt.Sprintf("%s %s", opts.Sort, opts.Order)
	query = query.Order(orderClause)

	// Apply pagination
	query = query.Offset(opts.Offset()).Limit(opts.Limit())

	// Execute query
	if err := query.Find(&entities).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list entities: %w", err)
	}

	return entities, total, nil
}

// Update updates an existing entity.
func (r *BaseRepository[T, ID]) Update(ctx context.Context, entity *T) error {
	result := r.db.WithContext(ctx).Save(entity)
	if result.Error != nil {
		return fmt.Errorf("failed to update entity: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes an entity by its ID.
func (r *BaseRepository[T, ID]) Delete(ctx context.Context, id ID) error {
	result := r.db.WithContext(ctx).Delete(new(T), "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete entity: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Exists checks if an entity exists by its ID.
func (r *BaseRepository[T, ID]) Exists(ctx context.Context, id ID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to check existence: %w", result.Error)
	}
	return count > 0, nil
}

// DB returns the underlying GORM DB instance for custom queries.
func (r *BaseRepository[T, ID]) DB() *gorm.DB {
	return r.db
}
