package query

import (
	"fmt"
	"strings"
)

// Filter represents a query filter.
type Filter struct {
	Field    string
	Operator string
	Value    interface{}
}

// Builder helps build SQL WHERE clauses with filters.
type Builder struct {
	filters []Filter
	args    []interface{}
}

// NewBuilder creates a new query builder.
func NewBuilder() *Builder {
	return &Builder{
		filters: make([]Filter, 0),
		args:    make([]interface{}, 0),
	}
}

// AddFilter adds a filter to the builder.
func (qb *Builder) AddFilter(field, operator string, value interface{}) *Builder {
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

// AddEqualFilter adds an equality filter.
func (qb *Builder) AddEqualFilter(field string, value interface{}) *Builder {
	return qb.AddFilter(field, "=", value)
}

// AddLikeFilter adds a LIKE filter (case-insensitive).
func (qb *Builder) AddLikeFilter(field string, value string) *Builder {
	if value == "" {
		return qb
	}
	return qb.AddFilter(field, "ILIKE", "%"+value+"%")
}

// AddInFilter adds an IN filter.
func (qb *Builder) AddInFilter(field string, values []string) *Builder {
	if len(values) == 0 {
		return qb
	}
	return qb.AddFilter(field, "IN", values)
}

// AddRangeFilter adds range filters (>=, <=).
func (qb *Builder) AddRangeFilter(field string, min, max interface{}) *Builder {
	if min != nil && min != "" {
		qb.AddFilter(field, ">=", min)
	}
	if max != nil && max != "" {
		qb.AddFilter(field, "<=", max)
	}
	return qb
}

// Build builds the WHERE clause and returns it with arguments.
func (qb *Builder) Build() (string, []interface{}) {
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
