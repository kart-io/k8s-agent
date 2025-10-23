package contextx_test

import (
	"context"
	"testing"
	"time"

	"github.com/kart-io/k8s-agent/pkg/contextx"
	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	t.Run("RequestID", func(t *testing.T) {
		ctx := context.Background()

		// Set request ID
		ctx = contextx.WithRequestID(ctx, "req-123")

		// Get request ID
		requestID := contextx.GetRequestID(ctx)
		assert.Equal(t, "req-123", requestID)
	})

	t.Run("GetOrCreateRequestID", func(t *testing.T) {
		ctx := context.Background()

		// Create new request ID
		ctx, requestID := contextx.GetOrCreateRequestID(ctx)
		assert.NotEmpty(t, requestID)

		// Get existing request ID
		ctx2, requestID2 := contextx.GetOrCreateRequestID(ctx)
		assert.Equal(t, requestID, requestID2)
		assert.Equal(t, ctx, ctx2)
	})

	t.Run("UserID and Username", func(t *testing.T) {
		ctx := context.Background()

		ctx = contextx.WithUserID(ctx, "user-456")
		ctx = contextx.WithUsername(ctx, "john.doe")

		assert.Equal(t, "user-456", contextx.GetUserID(ctx))
		assert.Equal(t, "john.doe", contextx.GetUsername(ctx))
	})

	t.Run("TraceID and SpanID", func(t *testing.T) {
		ctx := context.Background()

		ctx = contextx.WithTraceID(ctx, "trace-789")
		ctx = contextx.WithSpanID(ctx, "span-012")

		assert.Equal(t, "trace-789", contextx.GetTraceID(ctx))
		assert.Equal(t, "span-012", contextx.GetSpanID(ctx))
	})

	t.Run("TenantID", func(t *testing.T) {
		ctx := context.Background()
		ctx = contextx.WithTenantID(ctx, "tenant-abc")

		assert.Equal(t, "tenant-abc", contextx.GetTenantID(ctx))
	})

	t.Run("Client information", func(t *testing.T) {
		ctx := context.Background()

		ctx = contextx.WithClientIP(ctx, "192.168.1.1")
		ctx = contextx.WithUserAgent(ctx, "Mozilla/5.0")
		ctx = contextx.WithRealIP(ctx, "10.0.0.1")

		assert.Equal(t, "192.168.1.1", contextx.GetClientIP(ctx))
		assert.Equal(t, "Mozilla/5.0", contextx.GetUserAgent(ctx))
		assert.Equal(t, "10.0.0.1", contextx.GetRealIP(ctx))
	})

	t.Run("SessionID", func(t *testing.T) {
		ctx := context.Background()
		ctx = contextx.WithSessionID(ctx, "session-xyz")

		assert.Equal(t, "session-xyz", contextx.GetSessionID(ctx))
	})

	t.Run("ExtractInfo", func(t *testing.T) {
		ctx := context.Background()
		ctx = contextx.WithRequestID(ctx, "req-123")
		ctx = contextx.WithUserID(ctx, "user-456")
		ctx = contextx.WithUsername(ctx, "john.doe")
		ctx = contextx.WithTraceID(ctx, "trace-789")

		info := contextx.ExtractInfo(ctx)
		assert.Equal(t, "req-123", info.RequestID)
		assert.Equal(t, "user-456", info.UserID)
		assert.Equal(t, "john.doe", info.Username)
		assert.Equal(t, "trace-789", info.TraceID)
	})

	t.Run("ApplyInfo", func(t *testing.T) {
		info := &contextx.ContextInfo{
			RequestID: "req-123",
			UserID:    "user-456",
			Username:  "john.doe",
			TraceID:   "trace-789",
		}

		ctx := contextx.ApplyInfo(context.Background(), info)

		assert.Equal(t, "req-123", contextx.GetRequestID(ctx))
		assert.Equal(t, "user-456", contextx.GetUserID(ctx))
		assert.Equal(t, "john.doe", contextx.GetUsername(ctx))
		assert.Equal(t, "trace-789", contextx.GetTraceID(ctx))
	})

	t.Run("CopyContext", func(t *testing.T) {
		ctx := context.Background()
		ctx = contextx.WithRequestID(ctx, "req-123")
		ctx = contextx.WithUserID(ctx, "user-456")

		newCtx := contextx.CopyContext(ctx)

		assert.Equal(t, "req-123", contextx.GetRequestID(newCtx))
		assert.Equal(t, "user-456", contextx.GetUserID(newCtx))
	})
}

func TestTimeout(t *testing.T) {
	t.Run("WithTimeout", func(t *testing.T) {
		opts := contextx.DefaultTimeoutOptions()

		ctx, cancel := opts.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		timeout := contextx.GetTimeout(ctx)
		assert.True(t, timeout > 0 && timeout <= time.Second*5)
	})

	t.Run("WithTimeout - enforce minimum", func(t *testing.T) {
		opts := contextx.DefaultTimeoutOptions()

		ctx, cancel := opts.WithTimeout(context.Background(), time.Millisecond*100)
		defer cancel()

		timeout := contextx.GetTimeout(ctx)
		assert.True(t, timeout >= time.Second)
	})

	t.Run("WithTimeout - enforce maximum", func(t *testing.T) {
		opts := contextx.DefaultTimeoutOptions()

		ctx, cancel := opts.WithTimeout(context.Background(), time.Hour)
		defer cancel()

		timeout := contextx.GetTimeout(ctx)
		assert.True(t, timeout <= time.Minute*5)
	})

	t.Run("HasTimeout", func(t *testing.T) {
		ctx := context.Background()
		assert.False(t, contextx.HasTimeout(ctx))

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		assert.True(t, contextx.HasTimeout(ctx))
	})

	t.Run("IsExpired", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
		defer cancel()

		assert.False(t, contextx.IsExpired(ctx))

		time.Sleep(time.Millisecond * 20)
		assert.True(t, contextx.IsExpired(ctx))
	})

	t.Run("WaitWithTimeout - success", func(t *testing.T) {
		ctx := context.Background()

		err := contextx.WaitWithTimeout(ctx, time.Second, func() error {
			time.Sleep(time.Millisecond * 10)
			return nil
		})

		assert.NoError(t, err)
	})

	t.Run("WaitWithTimeout - timeout", func(t *testing.T) {
		ctx := context.Background()

		err := contextx.WaitWithTimeout(ctx, time.Millisecond*10, func() error {
			time.Sleep(time.Second)
			return nil
		})

		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
	})

	t.Run("Detach", func(t *testing.T) {
		ctx := context.Background()
		ctx = contextx.WithRequestID(ctx, "req-123")
		ctx, cancel := context.WithCancel(ctx)

		// Cancel parent
		cancel()

		// Detached context should not be cancelled
		detached := contextx.Detach(ctx)
		assert.False(t, contextx.IsExpired(detached))
		assert.Equal(t, "req-123", contextx.GetRequestID(detached))
	})

	t.Run("DetachWithTimeout", func(t *testing.T) {
		ctx := context.Background()
		ctx = contextx.WithRequestID(ctx, "req-123")
		ctx, cancel := context.WithCancel(ctx)

		// Cancel parent
		cancel()

		// Detached context should have its own timeout
		detached, detachedCancel := contextx.DetachWithTimeout(ctx, time.Second)
		defer detachedCancel()

		assert.False(t, contextx.IsExpired(detached))
		assert.Equal(t, "req-123", contextx.GetRequestID(detached))
		assert.True(t, contextx.HasTimeout(detached))
	})
}
