package l2

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kart-io/k8s-agent/common/cache"
	"github.com/kart-io/k8s-agent/common/cache/memory"
	"github.com/kart-io/k8s-agent/common/serializers"
)

func TestL2Cache_WithJSONSerializer(t *testing.T) {
	remote := memory.NewMemoryCache(cache.WithKeyPrefix("test:"))
	l2Cache, err := NewL2Cache[TestAgent](remote,
		cache.WithLocalSize(100),
		cache.WithSerializer(serializers.NewJSONSerializer()),
	)
	require.NoError(t, err)
	defer l2Cache.Close()

	ctx := context.Background()

	agent := TestAgent{
		ID:        "json-test",
		Name:      "JSON Agent",
		ClusterID: "cluster-json",
		Version:   1,
	}

	// Set and get
	err = l2Cache.Set(ctx, "agent:1", agent, time.Minute)
	assert.NoError(t, err)

	retrieved, err := l2Cache.Get(ctx, "agent:1")
	assert.NoError(t, err)
	assert.Equal(t, agent.ID, retrieved.ID)
	assert.Equal(t, agent.Name, retrieved.Name)
}

func TestL2Cache_WithMsgpackSerializer(t *testing.T) {
	remote := memory.NewMemoryCache(cache.WithKeyPrefix("test:"))
	l2Cache, err := NewL2Cache[TestAgent](remote,
		cache.WithLocalSize(100),
		cache.WithSerializer(serializers.NewMsgpackSerializer()),
	)
	require.NoError(t, err)
	defer l2Cache.Close()

	ctx := context.Background()

	agent := TestAgent{
		ID:        "msgpack-test",
		Name:      "Msgpack Agent",
		ClusterID: "cluster-msgpack",
		Version:   2,
	}

	// Set and get
	err = l2Cache.Set(ctx, "agent:2", agent, time.Minute)
	assert.NoError(t, err)

	retrieved, err := l2Cache.Get(ctx, "agent:2")
	assert.NoError(t, err)
	assert.Equal(t, agent.ID, retrieved.ID)
	assert.Equal(t, agent.Name, retrieved.Name)
}

func BenchmarkL2Cache_Serializers(b *testing.B) {
	testAgent := TestAgent{
		ID:        "benchmark-agent",
		Name:      "Performance Test Agent",
		ClusterID: "cluster-perf",
		Version:   100,
	}

	b.Run("JSON_Serializer", func(b *testing.B) {
		remote := memory.NewMemoryCache()
		l2Cache, _ := NewL2Cache[TestAgent](remote,
			cache.WithSerializer(serializers.NewJSONSerializer()),
		)
		defer l2Cache.Close()

		ctx := context.Background()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = l2Cache.Set(ctx, "agent:bench", testAgent, time.Minute)
			_, _ = l2Cache.Get(ctx, "agent:bench")
		}
	})

	b.Run("Msgpack_Serializer", func(b *testing.B) {
		remote := memory.NewMemoryCache()
		l2Cache, _ := NewL2Cache[TestAgent](remote,
			cache.WithSerializer(serializers.NewMsgpackSerializer()),
		)
		defer l2Cache.Close()

		ctx := context.Background()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = l2Cache.Set(ctx, "agent:bench", testAgent, time.Minute)
			_, _ = l2Cache.Get(ctx, "agent:bench")
		}
	})
}
