package factory

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kart-io/k8s-agent/common/cache"
	rediscache "github.com/kart-io/k8s-agent/common/cache/redis"
)

func TestNew_MemoryCache(t *testing.T) {
	config := &Config{
		Type: TypeMemory,
		Options: &cache.Options{
			KeyPrefix:         "test:",
			DefaultExpiration: 5 * time.Minute,
		},
	}

	cache, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, cache)
	defer cache.Close()

	// Test basic operations
	ctx := context.Background()
	err = cache.Set(ctx, "key1", []byte("value1"), time.Minute)
	assert.NoError(t, err)

	value, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, []byte("value1"), value)
}

func TestNew_RedisCache(t *testing.T) {
	// Setup miniredis
	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{
		Addr: mr.Addr(),
	})

	config := &Config{
		Type:        TypeRedis,
		RedisClient: client,
		Options: &cache.Options{
			KeyPrefix:         "test:",
			DefaultExpiration: 5 * time.Minute,
		},
	}

	cache, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, cache)
	defer cache.Close()

	// Test basic operations
	ctx := context.Background()
	err = cache.Set(ctx, "key1", []byte("value1"), time.Minute)
	assert.NoError(t, err)

	value, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, []byte("value1"), value)
}

func TestNew_L2Cache(t *testing.T) {
	// Setup miniredis
	mr := miniredis.RunT(t)
	redisClient := goredis.NewClient(&goredis.Options{
		Addr: mr.Addr(),
	})

	// Create remote cache (Redis)
	remoteCache := rediscache.NewRedisCache(redisClient,
		cache.WithKeyPrefix("remote:"),
		cache.WithDefaultExpiration(10*time.Minute),
	)

	config := &Config{
		Type:     TypeL2,
		L2Remote: remoteCache,
		L2Options: &cache.L2Options{
			LocalSize:     1000,
			LocalCounters: 10000,
			LocalTTL:      5 * time.Minute,
			LocalCost:     1,
			WriteThrough:  true,
		},
	}

	cache, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, cache)
	defer cache.Close()

	// Test basic operations
	ctx := context.Background()
	err = cache.Set(ctx, "key1", []byte("value1"), time.Minute)
	assert.NoError(t, err)

	value, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, []byte("value1"), value)
}

func TestNew_NilConfig(t *testing.T) {
	cache, err := New(nil)
	assert.Error(t, err)
	assert.Nil(t, cache)
	assert.Contains(t, err.Error(), "cache config cannot be nil")
}

func TestNew_UnsupportedType(t *testing.T) {
	config := &Config{
		Type: CacheType("unsupported"),
	}

	cache, err := New(config)
	assert.Error(t, err)
	assert.Nil(t, cache)
	assert.Contains(t, err.Error(), "unsupported cache type")
}

func TestNew_RedisCache_MissingClient(t *testing.T) {
	config := &Config{
		Type:        TypeRedis,
		RedisClient: nil,
	}

	cache, err := New(config)
	assert.Error(t, err)
	assert.Nil(t, cache)
	assert.Contains(t, err.Error(), "redis client is required")
}

func TestNew_L2Cache_MissingRemote(t *testing.T) {
	config := &Config{
		Type:     TypeL2,
		L2Remote: nil,
	}

	cache, err := New(config)
	assert.Error(t, err)
	assert.Nil(t, cache)
	assert.Contains(t, err.Error(), "remote cache is required")
}

func TestBuilder_Memory(t *testing.T) {
	cache, err := NewBuilder().
		Memory().
		WithPrefix("test:").
		WithExpiration(5 * time.Minute).
		Build()

	require.NoError(t, err)
	require.NotNil(t, cache)
	defer cache.Close()

	// Verify it works
	ctx := context.Background()
	err = cache.Set(ctx, "key1", []byte("value1"), time.Minute)
	assert.NoError(t, err)

	value, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, []byte("value1"), value)
}

func TestBuilder_Redis(t *testing.T) {
	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{
		Addr: mr.Addr(),
	})

	cache, err := NewBuilder().
		Redis(client).
		WithPrefix("test:").
		WithExpiration(5 * time.Minute).
		Build()

	require.NoError(t, err)
	require.NotNil(t, cache)
	defer cache.Close()

	// Verify it works
	ctx := context.Background()
	err = cache.Set(ctx, "key1", []byte("value1"), time.Minute)
	assert.NoError(t, err)

	value, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, []byte("value1"), value)
}

func TestBuilder_L2(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := goredis.NewClient(&goredis.Options{
		Addr: mr.Addr(),
	})

	remoteCache := rediscache.NewRedisCache(redisClient, cache.WithKeyPrefix("remote:"))

	cache, err := NewBuilder().
		L2(remoteCache).
		WithL2LocalSize(1000).
		WithL2LocalTTL(5 * time.Minute).
		Build()

	require.NoError(t, err)
	require.NotNil(t, cache)
	defer cache.Close()

	// Verify it works
	ctx := context.Background()
	err = cache.Set(ctx, "key1", []byte("value1"), time.Minute)
	assert.NoError(t, err)

	value, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, []byte("value1"), value)
}

func TestBuilder_MustBuild_Success(t *testing.T) {
	assert.NotPanics(t, func() {
		cache := NewBuilder().Memory().MustBuild()
		defer cache.Close()
		assert.NotNil(t, cache)
	})
}

func TestBuilder_MustBuild_Panic(t *testing.T) {
	assert.Panics(t, func() {
		// L2 without remote should panic
		NewBuilder().L2(nil).MustBuild()
	})
}

func TestBuilder_ChainedCalls(t *testing.T) {
	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{
		Addr: mr.Addr(),
	})

	// Test that builder can be reused for different cache types
	builder := NewBuilder()

	cache1, err := builder.Memory().WithPrefix("mem:").Build()
	require.NoError(t, err)
	defer cache1.Close()

	cache2, err := builder.Redis(client).WithPrefix("redis:").Build()
	require.NoError(t, err)
	defer cache2.Close()

	// Verify both work independently
	ctx := context.Background()

	err = cache1.Set(ctx, "key1", []byte("mem-value"), time.Minute)
	assert.NoError(t, err)

	err = cache2.Set(ctx, "key1", []byte("redis-value"), time.Minute)
	assert.NoError(t, err)

	val1, _ := cache1.Get(ctx, "key1")
	val2, _ := cache2.Get(ctx, "key1")

	assert.Equal(t, []byte("mem-value"), val1)
	assert.Equal(t, []byte("redis-value"), val2)
}

func TestNew_DefaultOptionsCreated(t *testing.T) {
	// When Options is nil, should create default options
	config := &Config{
		Type:    TypeMemory,
		Options: nil, // Explicitly nil
	}

	cache, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, cache)
	defer cache.Close()

	// Should still work with default options
	ctx := context.Background()
	err = cache.Set(ctx, "key1", []byte("value1"), time.Minute)
	assert.NoError(t, err)
}

func TestNew_L2_DefaultOptionsCreated(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := goredis.NewClient(&goredis.Options{
		Addr: mr.Addr(),
	})

	remoteCache := rediscache.NewRedisCache(redisClient)

	// L2Options nil should use defaults
	config := &Config{
		Type:      TypeL2,
		L2Remote:  remoteCache,
		L2Options: nil, // Should create default
	}

	cache, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, cache)
	defer cache.Close()

	// Should work with default L2 options
	ctx := context.Background()
	err = cache.Set(ctx, "key1", []byte("value1"), time.Minute)
	assert.NoError(t, err)
}
