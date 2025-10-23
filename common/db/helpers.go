package db

import (
	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

// NewMySQLFromOptions creates a MySQL client from DatabaseOptions
// This is a convenience function that converts options to functional options
func NewMySQLFromOptions(log core.Logger, opts *options.DatabaseOptions) (*MySQLClient, error) {
	return NewMySQL(log,
		WithHost(opts.Host),
		WithPort(opts.Port),
		WithUser(opts.User),
		WithPassword(opts.Password),
		WithDatabase(opts.Database),
		WithMaxOpenConns(opts.MaxOpenConns),
		WithMaxIdleConns(opts.MaxIdleConns),
		WithConnMaxLifetime(opts.ConnMaxLifetime),
	)
}

// NewRedisFromOptions creates a Redis client from RedisOptions
// This is a convenience function that converts options to functional options
func NewRedisFromOptions(log core.Logger, opts *options.RedisOptions) (*RedisClient, error) {
	return NewRedis(log,
		WithAddr(opts.Addr),
		WithRedisPassword(opts.Password),
		WithRedisDB(opts.DB),
		WithRedisPoolSize(opts.PoolSize),
		WithRedisMinIdleConns(opts.MinIdleConns),
		WithRedisDialTimeout(opts.DialTimeout),
		WithRedisReadTimeout(opts.ReadTimeout),
		WithRedisWriteTimeout(opts.WriteTimeout),
	)
}
