package mysql

import (
	"context"
	"fmt"
)

// Health checks if the MySQL database is healthy by pinging it.
// This is a convenience method that uses the default timeout.
// For custom timeout, use Health(ctx) method on the Client.
func (c *Client) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.SlowQueryThreshold*2)
	defer cancel()
	return c.Health(ctx)
}

// Stats returns connection pool statistics.
type Stats struct {
	MaxOpenConnections int
	OpenConnections    int
	InUse              int
	Idle               int
}

// Stats returns the current connection pool statistics.
func (c *Client) Stats() (*Stats, error) {
	sqlDB, err := c.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	dbStats := sqlDB.Stats()
	return &Stats{
		MaxOpenConnections: dbStats.MaxOpenConnections,
		OpenConnections:    dbStats.OpenConnections,
		InUse:              dbStats.InUse,
		Idle:               dbStats.Idle,
	}, nil
}
