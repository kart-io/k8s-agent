// Package query provides generic SQL query building utilities.
//
// This package is independent of business logic and can be used by any service
// that needs to build dynamic SQL WHERE clauses with filters.
//
// Key features:
//   - Fluent API for building WHERE clauses
//   - Support for common operators (=, IN, LIKE, >=, <=)
//   - SQL injection protection via parameterized queries
//   - Automatic handling of nil/empty values
//   - PostgreSQL-style placeholders ($1, $2, etc.)
//
// Example usage:
//
//	qb := query.NewBuilder()
//	qb.AddLikeFilter("name", "john")
//	qb.AddEqualFilter("status", 1)
//	qb.AddRangeFilter("created_at", startTime, endTime)
//
//	whereClause, args := qb.Build()
//	// whereClause: "name ILIKE $1 AND status = $2 AND created_at >= $3 AND created_at <= $4"
//	// args: []interface{}{"%john%", 1, startTime, endTime}
//
//	db.Where(whereClause, args...).Find(&results)
package query
