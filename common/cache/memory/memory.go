package memory

import (
	"context"
	"encoding/binary"
	"sync"
	"time"

	"github.com/kart-io/k8s-agent/common/cache"
)

// MemoryCache implements Cache interface using in-memory storage.
type MemoryCache struct {
	data    map[string]*memoryItem
	mu      sync.RWMutex
	options *cache.Options
	done    chan struct{} // Signal to stop cleanup goroutine
	closed  bool          // Track if cache is closed
	closeMu sync.Mutex    // Mutex for Close operation
}

type memoryItem struct {
	value      []byte
	expiration time.Time
}

// NewMemoryCache creates a new in-memory cache instance.
func NewMemoryCache(opts ...cache.Option) cache.Cache {
	options := cache.DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	c := &MemoryCache{
		data:    make(map[string]*memoryItem),
		options: options,
		done:    make(chan struct{}),
	}

	// Start cleanup goroutine
	go c.cleanup()

	return c
}

// cleanup periodically removes expired items.
func (m *MemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for key, item := range m.data {
				if !item.expiration.IsZero() && now.After(item.expiration) {
					delete(m.data, key)
				}
			}
			m.mu.Unlock()
		case <-m.done:
			return // Gracefully stop cleanup goroutine
		}
	}
}

// buildKey constructs the full cache key with prefix.
func (m *MemoryCache) buildKey(key string) string {
	return cache.BuildKey(m.options.KeyPrefix, key)
}

// Get retrieves a value from cache by key.
func (m *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	fullKey := m.buildKey(key)

	m.mu.RLock()
	item, exists := m.data[fullKey]
	m.mu.RUnlock()

	if !exists {
		return nil, cache.ErrKeyNotFound
	}

	// Check expiration
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		m.mu.Lock()
		delete(m.data, fullKey)
		m.mu.Unlock()
		return nil, cache.ErrKeyNotFound
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
		return cache.ErrKeyNotFound
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
		return nil, 0, cache.ErrKeyNotFound
	}

	// Check expiration
	now := time.Now()
	if !item.expiration.IsZero() && now.After(item.expiration) {
		m.mu.Lock()
		delete(m.data, fullKey)
		m.mu.Unlock()
		return nil, 0, cache.ErrKeyNotFound
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
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

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
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

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

	if exists && len(item.value) == 8 {
		// Parse existing value as int64 using standard library
		currentValue = int64(binary.LittleEndian.Uint64(item.value))
	}

	newValue := currentValue + delta

	// Store as bytes using standard library
	valueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueBytes, uint64(newValue))

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
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// If prefix is set, only clear keys with that prefix
	if m.options.KeyPrefix != "" {
		prefix := m.options.KeyPrefix + ":"
		for key := range m.data {
			// Periodically check context for large datasets
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

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

// Close closes the cache and stops the cleanup goroutine.
// It is safe to call Close multiple times.
func (m *MemoryCache) Close() error {
	m.closeMu.Lock()
	defer m.closeMu.Unlock()

	if m.closed {
		return nil // Already closed
	}

	m.closed = true
	close(m.done)
	return nil
}
