package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Store defines the interface for long-term persistent storage.
//
// Inspired by LangChain's Store, it provides:
//   - Namespace-based organization
//   - Key-value storage with metadata
//   - Search capabilities
//   - Timestamp tracking
//
// Use cases:
//   - User preferences and settings
//   - Historical conversation data
//   - Application-specific persistent data
type Store interface {
	// Put stores a value with the given namespace and key.
	Put(ctx context.Context, namespace []string, key string, value interface{}) error

	// Get retrieves a value by namespace and key.
	Get(ctx context.Context, namespace []string, key string) (*StoreValue, error)

	// Delete removes a value by namespace and key.
	Delete(ctx context.Context, namespace []string, key string) error

	// Search finds values matching the filter within a namespace.
	Search(ctx context.Context, namespace []string, filter map[string]interface{}) ([]*StoreValue, error)

	// List returns all keys within a namespace.
	List(ctx context.Context, namespace []string) ([]string, error)

	// Clear removes all values within a namespace.
	Clear(ctx context.Context, namespace []string) error
}

// StoreValue represents a value stored in the Store.
type StoreValue struct {
	// Value is the stored data
	Value interface{} `json:"value"`

	// Metadata holds additional information about the value
	Metadata map[string]interface{} `json:"metadata"`

	// Created is when the value was first created
	Created time.Time `json:"created"`

	// Updated is when the value was last updated
	Updated time.Time `json:"updated"`

	// Namespace is the hierarchical namespace path
	Namespace []string `json:"namespace"`

	// Key is the unique identifier within the namespace
	Key string `json:"key"`
}

// InMemoryStore is a thread-safe in-memory implementation of Store.
//
// Suitable for:
//   - Development and testing
//   - Small-scale deployments
//   - Ephemeral data that doesn't need persistence
type InMemoryStore struct {
	// data maps namespace path to key-value pairs
	data map[string]map[string]*StoreValue
	mu   sync.RWMutex
}

// NewInMemoryStore creates a new InMemoryStore.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]map[string]*StoreValue),
	}
}

// Put stores a value with the given namespace and key.
func (s *InMemoryStore) Put(ctx context.Context, namespace []string, key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	nsKey := namespaceToKey(namespace)
	if s.data[nsKey] == nil {
		s.data[nsKey] = make(map[string]*StoreValue)
	}

	now := time.Now()
	existing := s.data[nsKey][key]

	storeValue := &StoreValue{
		Value:     value,
		Metadata:  make(map[string]interface{}),
		Updated:   now,
		Namespace: namespace,
		Key:       key,
	}

	if existing != nil {
		storeValue.Created = existing.Created
		// Preserve existing metadata
		for k, v := range existing.Metadata {
			storeValue.Metadata[k] = v
		}
	} else {
		storeValue.Created = now
	}

	s.data[nsKey][key] = storeValue
	return nil
}

// Get retrieves a value by namespace and key.
func (s *InMemoryStore) Get(ctx context.Context, namespace []string, key string) (*StoreValue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nsKey := namespaceToKey(namespace)
	if s.data[nsKey] == nil {
		return nil, fmt.Errorf("namespace not found: %v", namespace)
	}

	value, ok := s.data[nsKey][key]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return value, nil
}

// Delete removes a value by namespace and key.
func (s *InMemoryStore) Delete(ctx context.Context, namespace []string, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	nsKey := namespaceToKey(namespace)
	if s.data[nsKey] == nil {
		return nil
	}

	delete(s.data[nsKey], key)

	// Clean up empty namespace
	if len(s.data[nsKey]) == 0 {
		delete(s.data, nsKey)
	}

	return nil
}

// Search finds values matching the filter within a namespace.
func (s *InMemoryStore) Search(ctx context.Context, namespace []string, filter map[string]interface{}) ([]*StoreValue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nsKey := namespaceToKey(namespace)
	if s.data[nsKey] == nil {
		return []*StoreValue{}, nil
	}

	results := make([]*StoreValue, 0)

	for _, value := range s.data[nsKey] {
		if matchesFilter(value, filter) {
			results = append(results, value)
		}
	}

	return results, nil
}

// List returns all keys within a namespace.
func (s *InMemoryStore) List(ctx context.Context, namespace []string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nsKey := namespaceToKey(namespace)
	if s.data[nsKey] == nil {
		return []string{}, nil
	}

	keys := make([]string, 0, len(s.data[nsKey]))
	for key := range s.data[nsKey] {
		keys = append(keys, key)
	}

	return keys, nil
}

// Clear removes all values within a namespace.
func (s *InMemoryStore) Clear(ctx context.Context, namespace []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	nsKey := namespaceToKey(namespace)
	delete(s.data, nsKey)

	return nil
}

// Size returns the total number of values across all namespaces.
func (s *InMemoryStore) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := 0
	for _, ns := range s.data {
		total += len(ns)
	}
	return total
}

// Namespaces returns all namespace keys currently in the store.
func (s *InMemoryStore) Namespaces() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	namespaces := make([]string, 0, len(s.data))
	for ns := range s.data {
		namespaces = append(namespaces, ns)
	}
	return namespaces
}

// namespaceToKey converts a namespace slice to a string key.
func namespaceToKey(namespace []string) string {
	if len(namespace) == 0 {
		return "/"
	}
	return "/" + joinNamespace(namespace)
}

// joinNamespace joins namespace components with "/".
func joinNamespace(namespace []string) string {
	if len(namespace) == 0 {
		return ""
	}
	result := namespace[0]
	for i := 1; i < len(namespace); i++ {
		result += "/" + namespace[i]
	}
	return result
}

// matchesFilter checks if a StoreValue matches the given filter.
func matchesFilter(value *StoreValue, filter map[string]interface{}) bool {
	if len(filter) == 0 {
		return true
	}

	for key, filterValue := range filter {
		metaValue, ok := value.Metadata[key]
		if !ok {
			return false
		}
		if metaValue != filterValue {
			return false
		}
	}

	return true
}

// StoreWithMetadata wraps a Store to automatically add metadata to all operations.
type StoreWithMetadata struct {
	store    Store
	metadata map[string]interface{}
}

// NewStoreWithMetadata creates a Store that adds metadata to all Put operations.
func NewStoreWithMetadata(store Store, metadata map[string]interface{}) *StoreWithMetadata {
	return &StoreWithMetadata{
		store:    store,
		metadata: metadata,
	}
}

// Put stores a value and adds the configured metadata.
func (s *StoreWithMetadata) Put(ctx context.Context, namespace []string, key string, value interface{}) error {
	// First put the value
	err := s.store.Put(ctx, namespace, key, value)
	if err != nil {
		return err
	}

	// Then update with metadata
	storeValue, err := s.store.Get(ctx, namespace, key)
	if err != nil {
		return err
	}

	// Add configured metadata
	for k, v := range s.metadata {
		storeValue.Metadata[k] = v
	}

	return s.store.Put(ctx, namespace, key, storeValue.Value)
}

// Get retrieves a value by namespace and key.
func (s *StoreWithMetadata) Get(ctx context.Context, namespace []string, key string) (*StoreValue, error) {
	return s.store.Get(ctx, namespace, key)
}

// Delete removes a value by namespace and key.
func (s *StoreWithMetadata) Delete(ctx context.Context, namespace []string, key string) error {
	return s.store.Delete(ctx, namespace, key)
}

// Search finds values matching the filter within a namespace.
func (s *StoreWithMetadata) Search(ctx context.Context, namespace []string, filter map[string]interface{}) ([]*StoreValue, error) {
	return s.store.Search(ctx, namespace, filter)
}

// List returns all keys within a namespace.
func (s *StoreWithMetadata) List(ctx context.Context, namespace []string) ([]string, error) {
	return s.store.List(ctx, namespace)
}

// Clear removes all values within a namespace.
func (s *StoreWithMetadata) Clear(ctx context.Context, namespace []string) error {
	return s.store.Clear(ctx, namespace)
}
