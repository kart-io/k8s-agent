package filter

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// Filter represents a query filter
type Filter struct {
	Field    string
	Operator string
	Value    interface{}
}

// QueryBuilder helps build SQL WHERE clauses with filters
type QueryBuilder struct {
	filters []Filter
	args    []interface{}
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		filters: make([]Filter, 0),
		args:    make([]interface{}, 0),
	}
}

// AddFilter adds a filter to the builder
func (qb *QueryBuilder) AddFilter(field, operator string, value interface{}) *QueryBuilder {
	if value == nil || value == "" {
		return qb
	}
	qb.filters = append(qb.filters, Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return qb
}

// AddEqualFilter adds an equality filter
func (qb *QueryBuilder) AddEqualFilter(field string, value interface{}) *QueryBuilder {
	return qb.AddFilter(field, "=", value)
}

// AddLikeFilter adds a LIKE filter (case-insensitive)
func (qb *QueryBuilder) AddLikeFilter(field string, value string) *QueryBuilder {
	if value == "" {
		return qb
	}
	return qb.AddFilter(field, "ILIKE", "%"+value+"%")
}

// AddInFilter adds an IN filter
func (qb *QueryBuilder) AddInFilter(field string, values []string) *QueryBuilder {
	if len(values) == 0 {
		return qb
	}
	return qb.AddFilter(field, "IN", values)
}

// AddRangeFilter adds range filters (>=, <=)
func (qb *QueryBuilder) AddRangeFilter(field string, min, max interface{}) *QueryBuilder {
	if min != nil && min != "" {
		qb.AddFilter(field, ">=", min)
	}
	if max != nil && max != "" {
		qb.AddFilter(field, "<=", max)
	}
	return qb
}

// Build builds the WHERE clause and returns it with arguments
func (qb *QueryBuilder) Build() (string, []interface{}) {
	if len(qb.filters) == 0 {
		return "", nil
	}

	var conditions []string
	argIndex := 1

	for _, filter := range qb.filters {
		switch filter.Operator {
		case "IN":
			values, ok := filter.Value.([]string)
			if !ok || len(values) == 0 {
				continue
			}
			placeholders := make([]string, len(values))
			for i, v := range values {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				qb.args = append(qb.args, v)
				argIndex++
			}
			conditions = append(conditions, fmt.Sprintf("%s IN (%s)", filter.Field, strings.Join(placeholders, ", ")))
		default:
			conditions = append(conditions, fmt.Sprintf("%s %s $%d", filter.Field, filter.Operator, argIndex))
			qb.args = append(qb.args, filter.Value)
			argIndex++
		}
	}

	return strings.Join(conditions, " AND "), qb.args
}

// UserFilters extracts user filter parameters from request
type UserFilters struct {
	Username string
	Email    string
	RealName string
	Status   *int
}

// ExtractUserFilters extracts user filters from Gin context
func ExtractUserFilters(c *gin.Context) UserFilters {
	var status *int
	if statusStr := c.Query("status"); statusStr != "" {
		switch statusStr {
		case "0":
			s := 0
			status = &s
		case "1":
			s := 1
			status = &s
		}
	}

	return UserFilters{
		Username: c.Query("username"),
		Email:    c.Query("email"),
		RealName: c.Query("real_name"),
		Status:   status,
	}
}

// ApplyUserFilters applies user filters to query builder
func ApplyUserFilters(qb *QueryBuilder, filters UserFilters) *QueryBuilder {
	qb.AddLikeFilter("username", filters.Username)
	qb.AddLikeFilter("email", filters.Email)
	qb.AddLikeFilter("real_name", filters.RealName)
	if filters.Status != nil {
		qb.AddEqualFilter("status", *filters.Status)
	}
	return qb
}

// RoleFilters extracts role filter parameters from request
type RoleFilters struct {
	Name   string
	Code   string
	Status *int
}

// ExtractRoleFilters extracts role filters from Gin context
func ExtractRoleFilters(c *gin.Context) RoleFilters {
	var status *int
	if statusStr := c.Query("status"); statusStr != "" {
		switch statusStr {
		case "0":
			s := 0
			status = &s
		case "1":
			s := 1
			status = &s
		}
	}

	return RoleFilters{
		Name:   c.Query("name"),
		Code:   c.Query("code"),
		Status: status,
	}
}

// ApplyRoleFilters applies role filters to query builder
func ApplyRoleFilters(qb *QueryBuilder, filters RoleFilters) *QueryBuilder {
	qb.AddLikeFilter("name", filters.Name)
	qb.AddLikeFilter("code", filters.Code)
	if filters.Status != nil {
		qb.AddEqualFilter("status", *filters.Status)
	}
	return qb
}

// PermissionFilters extracts permission filter parameters from request
type PermissionFilters struct {
	Name   string
	Code   string
	Type   string
	Status *int
}

// ExtractPermissionFilters extracts permission filters from Gin context
func ExtractPermissionFilters(c *gin.Context) PermissionFilters {
	var status *int
	if statusStr := c.Query("status"); statusStr != "" {
		switch statusStr {
		case "0":
			s := 0
			status = &s
		case "1":
			s := 1
			status = &s
		}
	}

	return PermissionFilters{
		Name:   c.Query("name"),
		Code:   c.Query("code"),
		Type:   c.Query("type"),
		Status: status,
	}
}

// ApplyPermissionFilters applies permission filters to query builder
func ApplyPermissionFilters(qb *QueryBuilder, filters PermissionFilters) *QueryBuilder {
	qb.AddLikeFilter("name", filters.Name)
	qb.AddLikeFilter("code", filters.Code)
	qb.AddEqualFilter("type", filters.Type)
	if filters.Status != nil {
		qb.AddEqualFilter("status", *filters.Status)
	}
	return qb
}
