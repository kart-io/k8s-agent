package cache

import "errors"

var (
	// ErrKeyNotFound indicates the key does not exist in cache.
	ErrKeyNotFound = errors.New("cache: key not found")
)
