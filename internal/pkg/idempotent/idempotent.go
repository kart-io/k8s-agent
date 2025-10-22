// Package idempotent provides idempotency support for distributed operations.
package idempotent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrDuplicateRequest indicates a duplicate request was detected.
	ErrDuplicateRequest = errors.New("duplicate request detected")

	// ErrKeyExpired indicates the idempotency key has expired.
	ErrKeyExpired = errors.New("idempotency key expired")

	// ErrInvalidKey indicates an invalid idempotency key.
	ErrInvalidKey = errors.New("invalid idempotency key")
)

// Status represents the status of an idempotent operation.
type Status string

const (
	// StatusProcessing indicates operation is in progress.
	StatusProcessing Status = "processing"

	// StatusCompleted indicates operation completed successfully.
	StatusCompleted Status = "completed"

	// StatusFailed indicates operation failed.
	StatusFailed Status = "failed"
)

// Record represents an idempotency record.
type Record struct {
	// Key is the unique idempotency key
	Key string

	// Status is the current status
	Status Status

	// Response is the cached response (for completed operations)
	Response []byte

	// Error is the error message (for failed operations)
	Error string

	// CreatedAt is when the record was created
	CreatedAt time.Time

	// UpdatedAt is when the record was last updated
	UpdatedAt time.Time

	// ExpiresAt is when the record expires
	ExpiresAt time.Time
}

// Store defines the interface for storing idempotency records.
type Store interface {
	// Get retrieves a record by key.
	Get(ctx context.Context, key string) (*Record, error)

	// Set stores a record.
	Set(ctx context.Context, record *Record) error

	// Delete removes a record.
	Delete(ctx context.Context, key string) error

	// Acquire attempts to acquire a lock for processing.
	Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error)

	// Release releases a lock.
	Release(ctx context.Context, key string) error
}

// Handler manages idempotent operations.
type Handler struct {
	store      Store
	defaultTTL time.Duration
	lockTTL    time.Duration
}

// NewHandler creates a new idempotency handler.
func NewHandler(store Store, defaultTTL, lockTTL time.Duration) *Handler {
	if defaultTTL == 0 {
		defaultTTL = time.Hour * 24 // 24 hours default
	}
	if lockTTL == 0 {
		lockTTL = time.Minute * 5 // 5 minutes lock
	}

	return &Handler{
		store:      store,
		defaultTTL: defaultTTL,
		lockTTL:    lockTTL,
	}
}

// Execute executes an operation idempotently.
func (h *Handler) Execute(ctx context.Context, key string, fn func(context.Context) ([]byte, error)) ([]byte, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}

	// Check if operation already exists
	record, err := h.store.Get(ctx, key)
	if err == nil {
		// Record exists, check status
		switch record.Status {
		case StatusCompleted:
			// Operation already completed, return cached response
			return record.Response, nil

		case StatusFailed:
			// Operation previously failed, return error
			if record.Error != "" {
				return nil, fmt.Errorf("previous operation failed: %s", record.Error)
			}
			return nil, errors.New("previous operation failed")

		case StatusProcessing:
			// Operation in progress, reject duplicate
			return nil, ErrDuplicateRequest
		}
	}

	// Try to acquire lock
	acquired, err := h.store.Acquire(ctx, key, h.lockTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !acquired {
		return nil, ErrDuplicateRequest
	}

	// Ensure lock is released
	defer h.store.Release(ctx, key)

	// Create processing record
	now := time.Now()
	record = &Record{
		Key:       key,
		Status:    StatusProcessing,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(h.defaultTTL),
	}

	if err := h.store.Set(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create processing record: %w", err)
	}

	// Execute operation
	response, execErr := fn(ctx)

	// Update record with result
	now = time.Now()
	if execErr != nil {
		record.Status = StatusFailed
		record.Error = execErr.Error()
		record.UpdatedAt = now
	} else {
		record.Status = StatusCompleted
		record.Response = response
		record.UpdatedAt = now
	}

	if err := h.store.Set(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to update record: %w", err)
	}

	return response, execErr
}

// Check checks the status of an idempotent operation.
func (h *Handler) Check(ctx context.Context, key string) (*Record, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}

	return h.store.Get(ctx, key)
}

// Delete deletes an idempotency record.
func (h *Handler) Delete(ctx context.Context, key string) error {
	if key == "" {
		return ErrInvalidKey
	}

	return h.store.Delete(ctx, key)
}

// GenerateKey generates an idempotency key from request data.
func GenerateKey(prefix string, data ...string) string {
	hasher := sha256.New()

	hasher.Write([]byte(prefix))
	for _, d := range data {
		hasher.Write([]byte(d))
	}

	hash := hasher.Sum(nil)
	return prefix + ":" + hex.EncodeToString(hash)
}

// GenerateKeyFromBytes generates an idempotency key from byte data.
func GenerateKeyFromBytes(prefix string, data []byte) string {
	hasher := sha256.New()
	hasher.Write([]byte(prefix))
	hasher.Write(data)

	hash := hasher.Sum(nil)
	return prefix + ":" + hex.EncodeToString(hash)
}
