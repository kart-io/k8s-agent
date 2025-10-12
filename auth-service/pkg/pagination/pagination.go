package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/auth-service/pkg/types"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// GetPaginationParams extracts and validates pagination parameters from request
func GetPaginationParams(c *gin.Context) types.PaginationParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(DefaultPage)))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(DefaultPageSize)))
	sort := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")

	// Validate page
	if page < 1 {
		page = DefaultPage
	}

	// Validate page size
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	// Validate order
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	return types.PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Order:    order,
	}
}

// CalculateOffset calculates the database offset from page and pageSize
func CalculateOffset(page, pageSize int) int {
	return (page - 1) * pageSize
}

// CalculateTotalPages calculates total pages from total records and page size
func CalculateTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	return totalPages
}

// BuildPaginatedResponse builds a paginated response
func BuildPaginatedResponse(items interface{}, total int64, params types.PaginationParams) types.PaginatedResponse {
	return types.PaginatedResponse{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(total, params.PageSize),
	}
}

// GetLimitOffset returns LIMIT and OFFSET values for SQL queries
func GetLimitOffset(params types.PaginationParams) (limit, offset int) {
	return params.PageSize, CalculateOffset(params.Page, params.PageSize)
}

// BuildOrderBy builds ORDER BY clause for SQL queries
func BuildOrderBy(params types.PaginationParams, allowedFields map[string]string) string {
	// Check if sort field is allowed
	field, ok := allowedFields[params.Sort]
	if !ok {
		// Default to created_at if field not allowed
		field = "created_at"
	}

	order := "DESC"
	if params.Order == "asc" {
		order = "ASC"
	}

	return field + " " + order
}
