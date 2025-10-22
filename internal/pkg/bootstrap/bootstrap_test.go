package bootstrap_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kart-io/k8s-agent/internal/pkg/bootstrap"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockInitializer struct {
	name     string
	priority int
	initErr  error
	closeErr error
	initCalled bool
	closeCalled bool
}

func (m *mockInitializer) Name() string {
	return m.name
}

func (m *mockInitializer) Priority() int {
	return m.priority
}

func (m *mockInitializer) Initialize(ctx context.Context) error {
	m.initCalled = true
	return m.initErr
}

func (m *mockInitializer) Close(ctx context.Context) error {
	m.closeCalled = true
	return m.closeErr
}

func (m *mockInitializer) HealthCheck(ctx context.Context) error {
	return nil
}

func TestBootstrap(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	t.Run("Initialize in priority order", func(t *testing.T) {
		b := bootstrap.New(logger)

		init1 := &mockInitializer{name: "init1", priority: 300}
		init2 := &mockInitializer{name: "init2", priority: 100}
		init3 := &mockInitializer{name: "init3", priority: 200}

		b.Register(init1)
		b.Register(init2)
		b.Register(init3)

		ctx := context.Background()
		err := b.Initialize(ctx)
		require.NoError(t, err)

		assert.True(t, init1.initCalled)
		assert.True(t, init2.initCalled)
		assert.True(t, init3.initCalled)
		assert.True(t, b.IsInitialized())
	})

	t.Run("Initialize error stops process", func(t *testing.T) {
		b := bootstrap.New(logger)

		init1 := &mockInitializer{name: "init1", priority: 100}
		init2 := &mockInitializer{name: "init2", priority: 200, initErr: errors.New("init failed")}
		init3 := &mockInitializer{name: "init3", priority: 300}

		b.Register(init1)
		b.Register(init2)
		b.Register(init3)

		ctx := context.Background()
		err := b.Initialize(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "init failed")

		assert.True(t, init1.initCalled)
		assert.True(t, init2.initCalled)
		assert.False(t, init3.initCalled)
		assert.False(t, b.IsInitialized())
	})

	t.Run("Shutdown in reverse order", func(t *testing.T) {
		b := bootstrap.New(logger)

		init1 := &mockInitializer{name: "init1", priority: 100}
		init2 := &mockInitializer{name: "init2", priority: 200}
		init3 := &mockInitializer{name: "init3", priority: 300}

		b.Register(init1)
		b.Register(init2)
		b.Register(init3)

		ctx := context.Background()
		err := b.Initialize(ctx)
		require.NoError(t, err)

		err = b.Shutdown(ctx)
		require.NoError(t, err)

		assert.True(t, init1.closeCalled)
		assert.True(t, init2.closeCalled)
		assert.True(t, init3.closeCalled)
	})

	t.Run("HealthCheck", func(t *testing.T) {
		b := bootstrap.New(logger)

		init1 := &mockInitializer{name: "init1", priority: 100}
		b.Register(init1)

		ctx := context.Background()
		err := b.Initialize(ctx)
		require.NoError(t, err)

		err = b.HealthCheck(ctx)
		require.NoError(t, err)
	})
}

func TestFuncInitializer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	t.Run("Function initializer", func(t *testing.T) {
		b := bootstrap.New(logger)

		initCalled := false
		closeCalled := false

		init := bootstrap.NewFuncInitializer(
			"test",
			100,
			func(ctx context.Context) error {
				initCalled = true
				return nil
			},
			func(ctx context.Context) error {
				closeCalled = true
				return nil
			},
		)

		b.Register(init)

		ctx := context.Background()
		err := b.Initialize(ctx)
		require.NoError(t, err)
		assert.True(t, initCalled)

		err = b.Shutdown(ctx)
		require.NoError(t, err)
		assert.True(t, closeCalled)
	})
}

func TestRetryInitializer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	t.Run("Retry on failure", func(t *testing.T) {
		attempts := 0
		mock := &mockInitializer{
			name:     "test",
			priority: 100,
			initErr:  errors.New("temporary failure"),
		}

		// Override Initialize to succeed on 3rd attempt
		oldInit := mock.Initialize
		mock.Initialize = func(ctx context.Context) error {
			attempts++
			if attempts < 3 {
				return oldInit(ctx)
			}
			mock.initErr = nil
			return nil
		}

		retry := bootstrap.NewRetryInitializer(mock, 5, time.Millisecond*10)

		b := bootstrap.New(logger)
		b.Register(retry)

		ctx := context.Background()
		err := b.Initialize(ctx)
		require.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})

	t.Run("Retry exhausted", func(t *testing.T) {
		mock := &mockInitializer{
			name:     "test",
			priority: 100,
			initErr:  errors.New("permanent failure"),
		}

		retry := bootstrap.NewRetryInitializer(mock, 2, time.Millisecond*10)

		b := bootstrap.New(logger)
		b.Register(retry)

		ctx := context.Background()
		err := b.Initialize(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "permanent failure")
	})
}
