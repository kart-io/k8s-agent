# Cache Package

A unified caching interface with multiple backend support for the Aetherius platform.

## Features

- **Unified Interface**: Single API for multiple cache backends
- **Redis Backend**: Production-ready Redis implementation
- **Memory Backend**: In-memory cache for testing and development
- **Compression**: Optional data compression for large values
- **Key Prefixing**: Namespace isolation with key prefixes
- **Batch Operations**: Efficient multi-get and multi-set operations
- **Atomic Counters**: Thread-safe increment/decrement operations
- **TTL Support**: Flexible expiration time management

## Installation

```go
import "github.com/kart-io/k8s-agent/pkg/cache"
```

## Usage

### Memory Cache (Development/Testing)

```go
package main

import (
    "context"
    "time"

    "github.com/kart-io/k8s-agent/pkg/cache"
)

func main() {
    ctx := context.Background()

    // Create memory cache with options
    c := cache.NewMemoryCache(
        cache.WithKeyPrefix("myapp"),
        cache.WithDefaultExpiration(time.Hour),
    )
    defer c.Close()

    // Set value
    err := c.Set(ctx, "user:123", []byte(`{"name":"John"}`), time.Minute*30)

    // Get value
    data, err := c.Get(ctx, "user:123")

    // Check existence
    exists, err := c.Exists(ctx, "user:123")

    // Delete value
    err = c.Delete(ctx, "user:123")
}
```

### Redis Cache (Production)

```go
package main

import (
    "context"
    "time"

    "github.com/kart-io/k8s-agent/pkg/cache"
    "github.com/redis/go-redis/v9"
)

func main() {
    ctx := context.Background()

    // Create Redis client
    redisClient := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })

    // Create Redis cache with options
    c := cache.NewRedisCache(
        redisClient,
        cache.WithKeyPrefix("myapp"),
        cache.WithDefaultExpiration(time.Hour),
        cache.WithCompression(1024), // Compress values > 1KB
    )
    defer c.Close()

    // Same API as memory cache
    err := c.Set(ctx, "key", []byte("value"), time.Minute*5)
    data, err := c.Get(ctx, "key")
}
```

### Batch Operations

```go
// Multi-set
items := map[string][]byte{
    "key1": []byte("value1"),
    "key2": []byte("value2"),
    "key3": []byte("value3"),
}
err := c.MSet(ctx, items, time.Hour)

// Multi-get
results, err := c.MGet(ctx, "key1", "key2", "key3")
for key, value := range results {
    fmt.Printf("%s: %s\n", key, value)
}
```

### Atomic Counters

```go
// Increment counter
newValue, err := c.Increment(ctx, "page_views", 1)

// Decrement counter
newValue, err := c.Decrement(ctx, "remaining_credits", 5)
```

### TTL Management

```go
// Get value with TTL
data, ttl, err := c.GetWithTTL(ctx, "session:abc")
fmt.Printf("TTL remaining: %v\n", ttl)

// Update expiration
err = c.Expire(ctx, "session:abc", time.Hour)
```

## Configuration Options

```go
type Options struct {
    // Prefix for all cache keys
    KeyPrefix string

    // Default expiration time
    DefaultExpiration time.Duration

    // Maximum number of retries
    MaxRetries int

    // Retry delay
    RetryDelay time.Duration

    // Enable compression for large values
    EnableCompression bool

    // Compression threshold in bytes
    CompressionThreshold int
}
```

### Available Options

- `WithKeyPrefix(prefix string)` - Set key prefix for namespace isolation
- `WithDefaultExpiration(duration)` - Set default expiration time
- `WithMaxRetries(retries int)` - Set maximum retry attempts (Redis only)
- `WithRetryDelay(delay)` - Set retry delay (Redis only)
- `WithCompression(threshold int)` - Enable compression for values above threshold

## Error Handling

```go
data, err := c.Get(ctx, "key")
if err != nil {
    if errors.Is(err, cache.ErrKeyNotFound) {
        // Handle cache miss
    } else {
        // Handle other errors
    }
}
```

### Available Errors

- `ErrKeyNotFound` - Key does not exist
- `ErrCacheMiss` - Cache miss occurred
- `ErrInvalidValue` - Invalid value
- `ErrConnectionFailed` - Connection failed (Redis)
- `ErrOperationTimeout` - Operation timeout
- `ErrCacheFull` - Cache is full

## Best Practices

1. **Use Key Prefixes**: Always use key prefixes to avoid collisions
   ```go
   c := cache.NewMemoryCache(cache.WithKeyPrefix("service1"))
   ```

2. **Set Appropriate TTLs**: Always set expiration times to prevent memory leaks
   ```go
   c.Set(ctx, "key", value, time.Hour*24)
   ```

3. **Handle Cache Misses**: Always handle `ErrKeyNotFound` gracefully
   ```go
   data, err := c.Get(ctx, "key")
   if errors.Is(err, cache.ErrKeyNotFound) {
       // Load from database
   }
   ```

4. **Use Batch Operations**: Use MGet/MSet for multiple keys
   ```go
   results, _ := c.MGet(ctx, "key1", "key2", "key3")
   ```

5. **Enable Compression**: Enable compression for large values
   ```go
   c := cache.NewRedisCache(client, cache.WithCompression(1024))
   ```

## Testing

Run tests with:

```bash
go test -v ./pkg/cache/...
```

## Performance

### Memory Cache
- Get: O(1)
- Set: O(1)
- Delete: O(1)
- No network overhead

### Redis Cache
- Get: O(1)
- Set: O(1)
- Delete: O(1)
- Network latency applies
- Compression overhead for large values

## Thread Safety

Both implementations are thread-safe and can be used concurrently from multiple goroutines.
