package idempotent

import (
	"context"
	"sync"
	"time"
)

// MemoryStore implements Store interface using in-memory storage.
type MemoryStore struct {
	records map[string]*Record
	locks   map[string]time.Time
	mu      sync.RWMutex
}

// NewMemoryStore creates a new in-memory idempotency store.
func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		records: make(map[string]*Record),
		locks:   make(map[string]time.Time),
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

// cleanup periodically removes expired records and locks.
func (m *MemoryStore) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()

		// Clean up expired records
		for key, record := range m.records {
			if !record.ExpiresAt.IsZero() && now.After(record.ExpiresAt) {
				delete(m.records, key)
			}
		}

		// Clean up expired locks
		for key, expiry := range m.locks {
			if now.After(expiry) {
				delete(m.locks, key)
			}
		}

		m.mu.Unlock()
	}
}

// Get retrieves a record by key.
func (m *MemoryStore) Get(ctx context.Context, key string) (*Record, error) {
	m.mu.RLock()
	record, exists := m.records[key]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrInvalidKey
	}

	// Check if expired
	if !record.ExpiresAt.IsZero() && time.Now().After(record.ExpiresAt) {
		m.mu.Lock()
		delete(m.records, key)
		m.mu.Unlock()
		return nil, ErrKeyExpired
	}

	// Return a copy
	recordCopy := *record
	if record.Response != nil {
		recordCopy.Response = make([]byte, len(record.Response))
		copy(recordCopy.Response, record.Response)
	}

	return &recordCopy, nil
}

// Set stores a record.
func (m *MemoryStore) Set(ctx context.Context, record *Record) error {
	if record.Key == "" {
		return ErrInvalidKey
	}

	// Make a copy
	recordCopy := *record
	if record.Response != nil {
		recordCopy.Response = make([]byte, len(record.Response))
		copy(recordCopy.Response, record.Response)
	}

	m.mu.Lock()
	m.records[record.Key] = &recordCopy
	m.mu.Unlock()

	return nil
}

// Delete removes a record.
func (m *MemoryStore) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	delete(m.records, key)
	m.mu.Unlock()

	return nil
}

// Acquire attempts to acquire a lock for processing.
func (m *MemoryStore) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	// Check if lock exists and is not expired
	if expiry, exists := m.locks[key]; exists {
		if now.Before(expiry) {
			return false, nil
		}
	}

	// Acquire lock
	m.locks[key] = now.Add(ttl)
	return true, nil
}

// Release releases a lock.
func (m *MemoryStore) Release(ctx context.Context, key string) error {
	m.mu.Lock()
	delete(m.locks, key)
	m.mu.Unlock()

	return nil
}

// Clear removes all records and locks (useful for testing).
func (m *MemoryStore) Clear() {
	m.mu.Lock()
	m.records = make(map[string]*Record)
	m.locks = make(map[string]time.Time)
	m.mu.Unlock()
}
