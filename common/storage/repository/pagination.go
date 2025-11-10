package repository

// PaginationInfo contains pagination metadata.
type PaginationInfo struct {
	// CurrentPage is the current page number (1-based)
	CurrentPage int `json:"current_page"`
	// PageSize is the number of items per page
	PageSize int `json:"page_size"`
	// TotalPages is the total number of pages
	TotalPages int `json:"total_pages"`
	// TotalItems is the total number of items across all pages
	TotalItems int64 `json:"total_items"`
	// HasNext indicates if there is a next page
	HasNext bool `json:"has_next"`
	// HasPrev indicates if there is a previous page
	HasPrev bool `json:"has_prev"`
}

// NewPaginationInfo creates pagination info from list options and total count.
func NewPaginationInfo(opts *ListOptions, total int64) *PaginationInfo {
	if opts == nil {
		opts = NewListOptions()
	}
	opts.Validate()

	totalPages := int(total) / opts.PageSize
	if int(total)%opts.PageSize > 0 {
		totalPages++
	}

	return &PaginationInfo{
		CurrentPage: opts.Page,
		PageSize:    opts.PageSize,
		TotalPages:  totalPages,
		TotalItems:  total,
		HasNext:     opts.Page < totalPages,
		HasPrev:     opts.Page > 1,
	}
}

// PagedResponse wraps entities with pagination metadata.
type PagedResponse[T any] struct {
	// Items is the list of entities for the current page
	Items []*T `json:"items"`
	// Pagination contains pagination metadata
	Pagination *PaginationInfo `json:"pagination"`
}

// NewPagedResponse creates a paged response from entities and pagination info.
func NewPagedResponse[T any](items []*T, pagination *PaginationInfo) *PagedResponse[T] {
	if items == nil {
		items = []*T{}
	}
	return &PagedResponse[T]{
		Items:      items,
		Pagination: pagination,
	}
}
