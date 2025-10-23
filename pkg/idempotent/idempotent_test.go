package idempotent_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kart-io/k8s-agent/pkg/idempotent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdempotentHandler(t *testing.T) {
	ctx := context.Background()

	t.Run("First execution succeeds", func(t *testing.T) {
		store := idempotent.NewMemoryStore()
		handler := idempotent.NewHandler(store, time.Hour, time.Minute)

		callCount := 0
		fn := func(ctx context.Context) ([]byte, error) {
			callCount++
			return []byte("result"), nil
		}

		result, err := handler.Execute(ctx, "test-key-1", fn)
		require.NoError(t, err)
		assert.Equal(t, []byte("result"), result)
		assert.Equal(t, 1, callCount)
	})

	t.Run("Duplicate execution returns cached result", func(t *testing.T) {
		store := idempotent.NewMemoryStore()
		handler := idempotent.NewHandler(store, time.Hour, time.Minute)

		callCount := 0
		fn := func(ctx context.Context) ([]byte, error) {
			callCount++
			return []byte("result"), nil
		}

		// First execution
		result1, err := handler.Execute(ctx, "test-key-2", fn)
		require.NoError(t, err)
		assert.Equal(t, []byte("result"), result1)

		// Second execution (should return cached)
		result2, err := handler.Execute(ctx, "test-key-2", fn)
		require.NoError(t, err)
		assert.Equal(t, []byte("result"), result2)

		// Function should only be called once
		assert.Equal(t, 1, callCount)
	})

	t.Run("Failed execution is recorded", func(t *testing.T) {
		store := idempotent.NewMemoryStore()
		handler := idempotent.NewHandler(store, time.Hour, time.Minute)

		expectedErr := errors.New("operation failed")
		fn := func(ctx context.Context) ([]byte, error) {
			return nil, expectedErr
		}

		// First execution fails
		_, err := handler.Execute(ctx, "test-key-3", fn)
		require.Error(t, err)

		// Check record status
		record, err := handler.Check(ctx, "test-key-3")
		require.NoError(t, err)
		assert.Equal(t, idempotent.StatusFailed, record.Status)
		assert.Equal(t, "operation failed", record.Error)
	})

	t.Run("Concurrent duplicate requests rejected", func(t *testing.T) {
		store := idempotent.NewMemoryStore()
		handler := idempotent.NewHandler(store, time.Hour, time.Minute)

		fn := func(ctx context.Context) ([]byte, error) {
			time.Sleep(time.Millisecond * 100)
			return []byte("result"), nil
		}

		// Start first execution
		errChan := make(chan error, 2)
		go func() {
			_, err := handler.Execute(ctx, "test-key-4", fn)
			errChan <- err
		}()

		// Wait a bit then try duplicate
		time.Sleep(time.Millisecond * 10)
		go func() {
			_, err := handler.Execute(ctx, "test-key-4", fn)
			errChan <- err
		}()

		// Collect results
		err1 := <-errChan
		err2 := <-errChan

		// One should succeed, one should be rejected
		if err1 == nil {
			assert.ErrorIs(t, err2, idempotent.ErrDuplicateRequest)
		} else {
			assert.ErrorIs(t, err1, idempotent.ErrDuplicateRequest)
			assert.NoError(t, err2)
		}
	})

	t.Run("Delete record", func(t *testing.T) {
		store := idempotent.NewMemoryStore()
		handler := idempotent.NewHandler(store, time.Hour, time.Minute)

		fn := func(ctx context.Context) ([]byte, error) {
			return []byte("result"), nil
		}

		// Execute
		_, err := handler.Execute(ctx, "test-key-5", fn)
		require.NoError(t, err)

		// Delete record
		err = handler.Delete(ctx, "test-key-5")
		require.NoError(t, err)

		// Check record is gone
		_, err = handler.Check(ctx, "test-key-5")
		assert.Error(t, err)
	})
}

func TestGenerateKey(t *testing.T) {
	t.Run("Generate key from strings", func(t *testing.T) {
		key1 := idempotent.GenerateKey("prefix", "data1", "data2")
		key2 := idempotent.GenerateKey("prefix", "data1", "data2")
		key3 := idempotent.GenerateKey("prefix", "data1", "data3")

		// Same input should generate same key
		assert.Equal(t, key1, key2)

		// Different input should generate different key
		assert.NotEqual(t, key1, key3)

		// Key should have prefix
		assert.Contains(t, key1, "prefix:")
	})

	t.Run("Generate key from bytes", func(t *testing.T) {
		data := []byte("test data")
		key1 := idempotent.GenerateKeyFromBytes("prefix", data)
		key2 := idempotent.GenerateKeyFromBytes("prefix", data)

		// Same input should generate same key
		assert.Equal(t, key1, key2)

		// Key should have prefix
		assert.Contains(t, key1, "prefix:")
	})
}

func TestMemoryStore(t *testing.T) {
	ctx := context.Background()

	t.Run("Store and retrieve record", func(t *testing.T) {
		store := idempotent.NewMemoryStore()

		record := &idempotent.Record{
			Key:       "test-key",
			Status:    idempotent.StatusCompleted,
			Response:  []byte("response"),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			ExpiresAt: time.Now().Add(time.Hour),
		}

		err := store.Set(ctx, record)
		require.NoError(t, err)

		retrieved, err := store.Get(ctx, "test-key")
		require.NoError(t, err)
		assert.Equal(t, record.Key, retrieved.Key)
		assert.Equal(t, record.Status, retrieved.Status)
		assert.Equal(t, record.Response, retrieved.Response)
	})

	t.Run("Lock acquisition", func(t *testing.T) {
		store := idempotent.NewMemoryStore()

		// First acquire should succeed
		acquired, err := store.Acquire(ctx, "lock-key", time.Second)
		require.NoError(t, err)
		assert.True(t, acquired)

		// Second acquire should fail
		acquired, err = store.Acquire(ctx, "lock-key", time.Second)
		require.NoError(t, err)
		assert.False(t, acquired)

		// Release lock
		err = store.Release(ctx, "lock-key")
		require.NoError(t, err)

		// Acquire should succeed again
		acquired, err = store.Acquire(ctx, "lock-key", time.Second)
		require.NoError(t, err)
		assert.True(t, acquired)
	})

	t.Run("Expired record", func(t *testing.T) {
		store := idempotent.NewMemoryStore()

		record := &idempotent.Record{
			Key:       "expired-key",
			Status:    idempotent.StatusCompleted,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			ExpiresAt: time.Now().Add(time.Millisecond * 50),
		}

		err := store.Set(ctx, record)
		require.NoError(t, err)

		// Wait for expiration
		time.Sleep(time.Millisecond * 100)

		// Should return expired error
		_, err = store.Get(ctx, "expired-key")
		assert.ErrorIs(t, err, idempotent.ErrKeyExpired)
	})
}
