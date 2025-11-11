package filter

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/pkg/query"
)

// UserFilters extracts user filter parameters from request.
type UserFilters struct {
	Username string
	Email    string
	RealName string
	Status   *int
}

// ExtractUserFilters extracts user filters from Gin context.
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

// ApplyUserFilters applies user filters to query builder.
func ApplyUserFilters(qb *query.Builder, filters UserFilters) *query.Builder {
	qb.AddLikeFilter("username", filters.Username)
	qb.AddLikeFilter("email", filters.Email)
	qb.AddLikeFilter("real_name", filters.RealName)
	if filters.Status != nil {
		qb.AddEqualFilter("status", *filters.Status)
	}
	return qb
}

// RoleFilters extracts role filter parameters from request.
type RoleFilters struct {
	Name   string
	Code   string
	Status *int
}

// ExtractRoleFilters extracts role filters from Gin context.
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

// ApplyRoleFilters applies role filters to query builder.
func ApplyRoleFilters(qb *query.Builder, filters RoleFilters) *query.Builder {
	qb.AddLikeFilter("name", filters.Name)
	qb.AddLikeFilter("code", filters.Code)
	if filters.Status != nil {
		qb.AddEqualFilter("status", *filters.Status)
	}
	return qb
}

// PermissionFilters extracts permission filter parameters from request.
type PermissionFilters struct {
	Name   string
	Code   string
	Type   string
	Status *int
}

// ExtractPermissionFilters extracts permission filters from Gin context.
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

// ApplyPermissionFilters applies permission filters to query builder.
func ApplyPermissionFilters(qb *query.Builder, filters PermissionFilters) *query.Builder {
	qb.AddLikeFilter("name", filters.Name)
	qb.AddLikeFilter("code", filters.Code)
	qb.AddEqualFilter("type", filters.Type)
	if filters.Status != nil {
		qb.AddEqualFilter("status", *filters.Status)
	}
	return qb
}
