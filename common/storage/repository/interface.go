package repository

import "context"

// Repository defines a generic repository interface using Go generics.
// T is the entity type (e.g., User, Product, etc.)
// ID is the identifier type (typically string or int)
type Repository[T any, ID comparable] interface {
	// Create creates a new entity in the repository.
	Create(ctx context.Context, entity *T) error

	// Get retrieves an entity by its ID.
	// Returns ErrNotFound if the entity doesn't exist.
	Get(ctx context.Context, id ID) (*T, error)

	// List retrieves multiple entities based on the provided options.
	// Returns a slice of entities and the total count (for pagination).
	List(ctx context.Context, opts *ListOptions) ([]*T, int64, error)

	// Update updates an existing entity.
	// Returns ErrNotFound if the entity doesn't exist.
	Update(ctx context.Context, entity *T) error

	// Delete deletes an entity by its ID.
	// Returns ErrNotFound if the entity doesn't exist.
	Delete(ctx context.Context, id ID) error

	// Exists checks if an entity exists by its ID.
	Exists(ctx context.Context, id ID) (bool, error)
}

// ListOptions contains options for listing entities.
type ListOptions struct {
	// Page is the page number (1-based, default: 1)
	Page int
	// PageSize is the number of items per page (default: 20, max: 100)
	PageSize int
	// Sort is the field to sort by (e.g., "created_at", "name")
	Sort string
	// Order is the sort order ("asc" or "desc", default: "asc")
	Order string
	// Filters is a map of field names to filter values
	// Example: {"status": "active", "role": "admin"}
	Filters map[string]interface{}
}

// NewListOptions creates a new ListOptions with default values.
func NewListOptions() *ListOptions {
	return &ListOptions{
		Page:     1,
		PageSize: 20,
		Sort:     "id",
		Order:    "asc",
		Filters:  make(map[string]interface{}),
	}
}

// Validate validates and normalizes list options.
func (o *ListOptions) Validate() {
	if o.Page < 1 {
		o.Page = 1
	}
	if o.PageSize < 1 {
		o.PageSize = 20
	}
	if o.PageSize > 100 {
		o.PageSize = 100
	}
	if o.Sort == "" {
		o.Sort = "id"
	}
	if o.Order != "asc" && o.Order != "desc" {
		o.Order = "asc"
	}
	if o.Filters == nil {
		o.Filters = make(map[string]interface{})
	}
}

// Offset calculates the offset for pagination.
func (o *ListOptions) Offset() int {
	return (o.Page - 1) * o.PageSize
}

// Limit returns the page size (for clarity in queries).
func (o *ListOptions) Limit() int {
	return o.PageSize
}
