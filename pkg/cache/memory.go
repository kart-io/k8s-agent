package cache

import (
	"context"
	"sync"
	"time"
)

// MemoryCache implements Cache interface using in-memory storage.
type MemoryCache struct {
	data    map[string]*memoryItem
	mu      sync.RWMutex
	options *Options
}

type memoryItem struct {
	value      []byte
	expiration time.Time
}

// NewMemoryCache creates a new in-memory cache instance.
func NewMemoryCache(opts ...Option) Cache {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	cache := &MemoryCache{
		data:    make(map[string]*memoryItem),
		options: options,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// cleanup periodically removes expired items.
func (m *MemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for key, item := range m.data {
			if !item.expiration.IsZero() && now.After(item.expiration) {
				delete(m.data, key)
			}
		}
		m.mu.Unlock()
	}
}

// buildKey constructs the full cache key with prefix.
func (m *MemoryCache) buildKey(key string) string {
	if m.options.KeyPrefix == "" {
		return key
	}
	return m.options.KeyPrefix + ":" + key
}

// Get retrieves a value from cache by key.
func (m *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	fullKey := m.buildKey(key)

	m.mu.RLock()
	item, exists := m.data[fullKey]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrKeyNotFound
	}

	// Check expiration
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		m.mu.Lock()
		delete(m.data, fullKey)
		m.mu.Unlock()
		return nil, ErrKeyNotFound
	}

	// Return a copy to prevent external modification
	result := make([]byte, len(item.value))
	copy(result, item.value)
	return result, nil
}

// Set stores a value in cache with expiration.
func (m *MemoryCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	fullKey := m.buildKey(key)

	if expiration == 0 {
		expiration = m.options.DefaultExpiration
	}

	// Make a copy to prevent external modification
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)

	item := &memoryItem{
		value: valueCopy,
	}

	if expiration > 0 {
		item.expiration = time.Now().Add(expiration)
	}

	m.mu.Lock()
	m.data[fullKey] = item
	m.mu.Unlock()

	return nil
}

// Delete removes a value from cache.
func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	fullKey := m.buildKey(key)

	m.mu.Lock()
	delete(m.data, fullKey)
	m.mu.Unlock()

	return nil
}

// Exists checks if a key exists in cache.
func (m *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := m.buildKey(key)

	m.mu.RLock()
	item, exists := m.data[fullKey]
	m.mu.RUnlock()

	if !exists {
		return false, nil
	}

	// Check expiration
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		m.mu.Lock()
		delete(m.data, fullKey)
		m.mu.Unlock()
		return false, nil
	}

	return true, nil
}

// Expire sets expiration time for a key.
func (m *MemoryCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	fullKey := m.buildKey(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.data[fullKey]
	if !exists {
		return ErrKeyNotFound
	}

	if expiration > 0 {
		item.expiration = time.Now().Add(expiration)
	} else {
		item.expiration = time.Time{}
	}

	return nil
}

// GetWithTTL retrieves value and remaining TTL.
func (m *MemoryCache) GetWithTTL(ctx context.Context, key string) ([]byte, time.Duration, error) {
	fullKey := m.buildKey(key)

	m.mu.RLock()
	item, exists := m.data[fullKey]
	m.mu.RUnlock()

	if !exists {
		return nil, 0, ErrKeyNotFound
	}

	// Check expiration
	now := time.Now()
	if !item.expiration.IsZero() && now.After(item.expiration) {
		m.mu.Lock()
		delete(m.data, fullKey)
		m.mu.Unlock()
		return nil, 0, ErrKeyNotFound
	}

	// Calculate TTL
	var ttl time.Duration
	if !item.expiration.IsZero() {
		ttl = item.expiration.Sub(now)
	} else {
		ttl = -1 // No expiration
	}

	// Return a copy
	result := make([]byte, len(item.value))
	copy(result, item.value)

	return result, ttl, nil
}

// MGet retrieves multiple values by keys.
func (m *MemoryCache) MGet(ctx context.Context, keys ...string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return make(map[string][]byte), nil
	}

	result := make(map[string][]byte, len(keys))
	now := time.Now()

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, key := range keys {
		fullKey := m.buildKey(key)
		item, exists := m.data[fullKey]
		if !exists {
			continue
		}

		// Check expiration
		if !item.expiration.IsZero() && now.After(item.expiration) {
			continue
		}

		// Copy value
		valueCopy := make([]byte, len(item.value))
		copy(valueCopy, item.value)
		result[key] = valueCopy
	}

	return result, nil
}

// MSet stores multiple key-value pairs.
func (m *MemoryCache) MSet(ctx context.Context, items map[string][]byte, expiration time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	if expiration == 0 {
		expiration = m.options.DefaultExpiration
	}

	var exp time.Time
	if expiration > 0 {
		exp = time.Now().Add(expiration)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for key, value := range items {
		fullKey := m.buildKey(key)

		// Make a copy
		valueCopy := make([]byte, len(value))
		copy(valueCopy, value)

		m.data[fullKey] = &memoryItem{
			value:      valueCopy,
			expiration: exp,
		}
	}

	return nil
}

// Increment atomically increments a counter.
func (m *MemoryCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	fullKey := m.buildKey(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.data[fullKey]
	var currentValue int64

	if exists {
		// Parse existing value as int64
		if len(item.value) == 8 {
			currentValue = int64(item.value[0]) | int64(item.value[1])<<8 |
				int64(item.value[2])<<16 | int64(item.value[3])<<24 |
				int64(item.value[4])<<32 | int64(item.value[5])<<40 |
				int64(item.value[6])<<48 | int64(item.value[7])<<56
		}
	}

	newValue := currentValue + delta

	// Store as bytes
	valueBytes := make([]byte, 8)
	valueBytes[0] = byte(newValue)
	valueBytes[1] = byte(newValue >> 8)
	valueBytes[2] = byte(newValue >> 16)
	valueBytes[3] = byte(newValue >> 24)
	valueBytes[4] = byte(newValue >> 32)
	valueBytes[5] = byte(newValue >> 40)
	valueBytes[6] = byte(newValue >> 48)
	valueBytes[7] = byte(newValue >> 56)

	if exists {
		item.value = valueBytes
	} else {
		m.data[fullKey] = &memoryItem{
			value: valueBytes,
		}
	}

	return newValue, nil
}

// Decrement atomically decrements a counter.
func (m *MemoryCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return m.Increment(ctx, key, -delta)
}

// Clear removes all keys from cache.
func (m *MemoryCache) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If prefix is set, only clear keys with that prefix
	if m.options.KeyPrefix != "" {
		prefix := m.options.KeyPrefix + ":"
		for key := range m.data {
			if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
				delete(m.data, key)
			}
		}
	} else {
		// Clear all keys
		m.data = make(map[string]*memoryItem)
	}

	return nil
}

// Close closes the cache (no-op for memory cache).
func (m *MemoryCache) Close() error {
	return nil
}
