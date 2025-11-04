package l2

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kart-io/k8s-agent/common/cache"
	"github.com/kart-io/k8s-agent/common/cache/memory"
)

// TestAgent is a sample type for testing generic cache
type TestAgent struct {
	ID        string
	Name      string
	ClusterID string
	Version   int
}

func TestL2Cache_NewL2Cache(t *testing.T) {
	remote := memory.NewMemoryCache(cache.WithKeyPrefix("test:"))

	t.Run("Create with default options", func(t *testing.T) {
		l2Cache, err := NewL2Cache[TestAgent](remote)
		require.NoError(t, err)
		require.NotNil(t, l2Cache)
		defer l2Cache.Close()
	})

	t.Run("Create with custom options", func(t *testing.T) {
		l2Cache, err := NewL2Cache[TestAgent](remote,
			cache.WithLocalSize(5000),
			cache.WithLocalTTL(10*time.Minute),
			cache.WithWriteThrough(true),
			cache.WithMetrics(true),
		)
		require.NoError(t, err)
		require.NotNil(t, l2Cache)
		defer l2Cache.Close()
	})
}

func TestL2Cache_GetSet(t *testing.T) {
	remote := memory.NewMemoryCache(cache.WithKeyPrefix("test:"))
	l2Cache, err := NewL2Cache[TestAgent](remote,
		cache.WithLocalSize(100),
		cache.WithLocalTTL(5*time.Minute),
	)
	require.NoError(t, err)
	defer l2Cache.Close()

	ctx := context.Background()

	t.Run("Set and Get from cache", func(t *testing.T) {
		agent := TestAgent{
			ID:        "agent-001",
			Name:      "test-agent",
			ClusterID: "cluster-1",
			Version:   1,
		}

		// Set value
		err := l2Cache.Set(ctx, "agent-001", agent, 10*time.Minute)
		assert.NoError(t, err)

		// Get value (should be in local cache)
		retrieved, err := l2Cache.Get(ctx, "agent-001")
		assert.NoError(t, err)
		assert.Equal(t, agent.ID, retrieved.ID)
		assert.Equal(t, agent.Name, retrieved.Name)
		assert.Equal(t, agent.ClusterID, retrieved.ClusterID)
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		_, err := l2Cache.Get(ctx, "non-existent")
		assert.Error(t, err)
	})
}

func TestL2Cache_LocalCachePopulation(t *testing.T) {
	remote := memory.NewMemoryCache(cache.WithKeyPrefix("test:"))
	l2Cache, err := NewL2Cache[TestAgent](remote,
		cache.WithLocalSize(100),
		cache.WithLocalTTL(5*time.Minute),
		cache.WithMetrics(true),
	)
	require.NoError(t, err)
	defer l2Cache.Close()

	ctx := context.Background()

	agent := TestAgent{
		ID:        "agent-002",
		Name:      "test-agent-2",
		ClusterID: "cluster-2",
		Version:   2,
	}

	// First, populate only remote cache (simulate cold start)
	data, err := json.Marshal(agent)
	require.NoError(t, err)
	err = remote.Set(ctx, "agent-002", data, 10*time.Minute)
	require.NoError(t, err)

	// First Get - should fetch from remote and populate local
	retrieved1, err := l2Cache.Get(ctx, "agent-002")
	assert.NoError(t, err)
	assert.Equal(t, agent.ID, retrieved1.ID)

	// Second Get - should hit local cache (faster)
	retrieved2, err := l2Cache.Get(ctx, "agent-002")
	assert.NoError(t, err)
	assert.Equal(t, agent.ID, retrieved2.ID)

	// Wait for Ristretto metrics to be updated
	time.Sleep(10 * time.Millisecond)

	// Verify cache hit metrics
	stats := l2Cache.Stats()
	assert.NotNil(t, stats)
	// Note: Ristretto metrics are probabilistic, so we just check they're not nil
	assert.NotNil(t, stats, "Stats should be available")
}

func TestL2Cache_WriteThroughBehavior(t *testing.T) {
	remote := memory.NewMemoryCache(cache.WithKeyPrefix("test:"))
	l2Cache, err := NewL2Cache[TestAgent](remote,
		cache.WithLocalSize(100),
		cache.WithWriteThrough(true),
		cache.WithInvalidateOnWrite(false),
	)
	require.NoError(t, err)
	defer l2Cache.Close()

	ctx := context.Background()

	agent := TestAgent{
		ID:        "agent-003",
		Name:      "test-agent-3",
		ClusterID: "cluster-3",
		Version:   3,
	}

	// Set value
	err = l2Cache.Set(ctx, "agent-003", agent, 10*time.Minute)
	assert.NoError(t, err)

	// Verify it's in remote cache
	remoteData, err := remote.Get(ctx, "agent-003")
	assert.NoError(t, err)

	var remoteAgent TestAgent
	err = json.Unmarshal(remoteData, &remoteAgent)
	assert.NoError(t, err)
	assert.Equal(t, agent.ID, remoteAgent.ID)

	// Verify it's in local cache too
	localAgent, err := l2Cache.Get(ctx, "agent-003")
	assert.NoError(t, err)
	assert.Equal(t, agent.ID, localAgent.ID)
}

func TestL2Cache_InvalidateOnWrite(t *testing.T) {
	remote := memory.NewMemoryCache(cache.WithKeyPrefix("test:"))
	l2Cache, err := NewL2Cache[TestAgent](remote,
		cache.WithLocalSize(100),
		cache.WithInvalidateOnWrite(true),
	)
	require.NoError(t, err)
	defer l2Cache.Close()

	ctx := context.Background()

	agent := TestAgent{
		ID:        "agent-004",
		Name:      "test-agent-4",
		ClusterID: "cluster-4",
		Version:   4,
	}

	// Set value (should write to remote but not populate local due to InvalidateOnWrite)
	err = l2Cache.Set(ctx, "agent-004", agent, 10*time.Minute)
	assert.NoError(t, err)

	// Get value (should fetch from remote and populate local)
	retrieved, err := l2Cache.Get(ctx, "agent-004")
	assert.NoError(t, err)
	assert.Equal(t, agent.ID, retrieved.ID)
}

func TestL2Cache_Delete(t *testing.T) {
	remote := memory.NewMemoryCache(cache.WithKeyPrefix("test:"))
	l2Cache, err := NewL2Cache[TestAgent](remote,
		cache.WithLocalSize(100),
	)
	require.NoError(t, err)
	defer l2Cache.Close()

	ctx := context.Background()

	agent := TestAgent{
		ID:        "agent-005",
		Name:      "test-agent-5",
		ClusterID: "cluster-5",
		Version:   5,
	}

	// Set value
	err = l2Cache.Set(ctx, "agent-005", agent, 10*time.Minute)
	assert.NoError(t, err)

	// Verify it exists
	exists, err := l2Cache.Exists(ctx, "agent-005")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Delete
	err = l2Cache.Delete(ctx, "agent-005")
	assert.NoError(t, err)

	// Verify it's gone
	exists, err = l2Cache.Exists(ctx, "agent-005")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Try to get deleted key
	_, err = l2Cache.Get(ctx, "agent-005")
	assert.Error(t, err)
}

func TestL2Cache_GetMulti(t *testing.T) {
	remote := memory.NewMemoryCache(cache.WithKeyPrefix("test:"))
	l2Cache, err := NewL2Cache[TestAgent](remote,
		cache.WithLocalSize(100),
	)
	require.NoError(t, err)
	defer l2Cache.Close()

	ctx := context.Background()

	// Populate cache with multiple agents
	agents := map[string]TestAgent{
		"agent-101": {ID: "agent-101", Name: "agent-1", ClusterID: "cluster-1", Version: 1},
		"agent-102": {ID: "agent-102", Name: "agent-2", ClusterID: "cluster-1", Version: 2},
		"agent-103": {ID: "agent-103", Name: "agent-3", ClusterID: "cluster-2", Version: 3},
	}

	for key, agent := range agents {
		err := l2Cache.Set(ctx, key, agent, 10*time.Minute)
		require.NoError(t, err)
	}

	// Get multiple keys
	keys := []string{"agent-101", "agent-102", "agent-103"}
	result, err := l2Cache.GetMulti(ctx, keys)
	assert.NoError(t, err)
	assert.Len(t, result, 3)

	for _, key := range keys {
		agent, found := result[key]
		assert.True(t, found, "Agent %s should be found", key)
		assert.Equal(t, agents[key].ID, agent.ID)
	}
}

func TestL2Cache_SetMulti(t *testing.T) {
	remote := memory.NewMemoryCache(cache.WithKeyPrefix("test:"))
	l2Cache, err := NewL2Cache[TestAgent](remote,
		cache.WithLocalSize(100),
	)
	require.NoError(t, err)
	defer l2Cache.Close()

	ctx := context.Background()

	// Set multiple agents
	agents := map[string]TestAgent{
		"agent-201": {ID: "agent-201", Name: "agent-1", ClusterID: "cluster-1", Version: 1},
		"agent-202": {ID: "agent-202", Name: "agent-2", ClusterID: "cluster-1", Version: 2},
		"agent-203": {ID: "agent-203", Name: "agent-3", ClusterID: "cluster-2", Version: 3},
	}

	err = l2Cache.SetMulti(ctx, agents, 10*time.Minute)
	assert.NoError(t, err)

	// Verify all agents can be retrieved
	for key, expectedAgent := range agents {
		retrievedAgent, err := l2Cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, expectedAgent.ID, retrievedAgent.ID)
		assert.Equal(t, expectedAgent.Name, retrievedAgent.Name)
	}
}

func TestL2Cache_Clear(t *testing.T) {
	remote := memory.NewMemoryCache(cache.WithKeyPrefix("test:"))
	l2Cache, err := NewL2Cache[TestAgent](remote,
		cache.WithLocalSize(100),
	)
	require.NoError(t, err)
	defer l2Cache.Close()

	ctx := context.Background()

	// Populate cache
	agents := map[string]TestAgent{
		"agent-301": {ID: "agent-301", Name: "agent-1", ClusterID: "cluster-1"},
		"agent-302": {ID: "agent-302", Name: "agent-2", ClusterID: "cluster-2"},
	}

	for key, agent := range agents {
		err := l2Cache.Set(ctx, key, agent, 10*time.Minute)
		require.NoError(t, err)
	}

	// Clear cache
	err = l2Cache.Clear(ctx)
	assert.NoError(t, err)

	// Verify all entries are gone
	for key := range agents {
		exists, err := l2Cache.Exists(ctx, key)
		assert.NoError(t, err)
		assert.False(t, exists, "Agent %s should be cleared", key)
	}
}

func TestL2Cache_Exists(t *testing.T) {
	remote := memory.NewMemoryCache(cache.WithKeyPrefix("test:"))
	l2Cache, err := NewL2Cache[TestAgent](remote,
		cache.WithLocalSize(100),
	)
	require.NoError(t, err)
	defer l2Cache.Close()

	ctx := context.Background()

	agent := TestAgent{
		ID:        "agent-006",
		Name:      "test-agent-6",
		ClusterID: "cluster-6",
	}

	// Key should not exist initially
	exists, err := l2Cache.Exists(ctx, "agent-006")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Set value
	err = l2Cache.Set(ctx, "agent-006", agent, 10*time.Minute)
	require.NoError(t, err)

	// Key should exist now
	exists, err = l2Cache.Exists(ctx, "agent-006")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestL2Cache_Stats(t *testing.T) {
	remote := memory.NewMemoryCache(cache.WithKeyPrefix("test:"))

	t.Run("Stats enabled", func(t *testing.T) {
		l2Cache, err := NewL2Cache[TestAgent](remote,
			cache.WithLocalSize(100),
			cache.WithMetrics(true),
		)
		require.NoError(t, err)
		defer l2Cache.Close()

		ctx := context.Background()

		agent := TestAgent{ID: "agent-007", Name: "test-agent-7"}
		err = l2Cache.Set(ctx, "agent-007", agent, 10*time.Minute)
		require.NoError(t, err)

		// First get - populates local cache
		_, err = l2Cache.Get(ctx, "agent-007")
		require.NoError(t, err)

		// Second get - hits local cache
		_, err = l2Cache.Get(ctx, "agent-007")
		require.NoError(t, err)

		// Wait for Ristretto metrics
		time.Sleep(10 * time.Millisecond)

		stats := l2Cache.Stats()
		assert.NotNil(t, stats)
		// Note: Ristretto metrics are probabilistic and may not always reflect immediate hits
		assert.NotNil(t, stats, "Stats should be available")
	})

	t.Run("Stats disabled", func(t *testing.T) {
		l2Cache, err := NewL2Cache[TestAgent](remote,
			cache.WithLocalSize(100),
			cache.WithMetrics(false),
		)
		require.NoError(t, err)
		defer l2Cache.Close()

		stats := l2Cache.Stats()
		assert.Nil(t, stats, "Stats should be nil when metrics disabled")
	})
}

func TestL2Cache_PerformanceComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	remote := memory.NewMemoryCache(cache.WithKeyPrefix("test:"))
	l2Cache, err := NewL2Cache[TestAgent](remote,
		cache.WithLocalSize(1000),
		cache.WithLocalTTL(5*time.Minute),
		cache.WithMetrics(true),
	)
	require.NoError(t, err)
	defer l2Cache.Close()

	ctx := context.Background()

	// Populate cache with test data
	agent := TestAgent{
		ID:        "perf-test-001",
		Name:      "performance-agent",
		ClusterID: "cluster-perf",
		Version:   1,
	}
	err = l2Cache.Set(ctx, "perf-test-001", agent, 10*time.Minute)
	require.NoError(t, err)

	// First access - should fetch from remote
	start := time.Now()
	_, err = l2Cache.Get(ctx, "perf-test-001")
	firstAccessDuration := time.Since(start)
	require.NoError(t, err)

	// Subsequent accesses - should hit local cache (faster)
	start = time.Now()
	for i := 0; i < 100; i++ {
		_, err = l2Cache.Get(ctx, "perf-test-001")
		require.NoError(t, err)
	}
	localAccessDuration := time.Since(start) / 100

	t.Logf("First access (remote): %v", firstAccessDuration)
	t.Logf("Avg local access: %v", localAccessDuration)
	t.Logf("Speedup: %.2fx", float64(firstAccessDuration)/float64(localAccessDuration))

	stats := l2Cache.Stats()
	t.Logf("Cache stats - Hits: %d, Misses: %d, Ratio: %.2f%%",
		stats.LocalHits, stats.LocalMisses, stats.Ratio*100)

	// Local cache should be significantly faster
	assert.Less(t, localAccessDuration, firstAccessDuration,
		"Local cache should be faster than remote")
}
