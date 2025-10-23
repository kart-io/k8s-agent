package cache

import "errors"

var (
	// ErrKeyNotFound indicates the key does not exist in cache.
	ErrKeyNotFound = errors.New("cache: key not found")

	// ErrCacheMiss indicates a cache miss occurred.
	ErrCacheMiss = errors.New("cache: miss")

	// ErrInvalidValue indicates the value is invalid.
	ErrInvalidValue = errors.New("cache: invalid value")

	// ErrConnectionFailed indicates cache connection failed.
	ErrConnectionFailed = errors.New("cache: connection failed")

	// ErrOperationTimeout indicates operation timed out.
	ErrOperationTimeout = errors.New("cache: operation timeout")

	// ErrCacheFull indicates cache is full.
	ErrCacheFull = errors.New("cache: cache full")
)
