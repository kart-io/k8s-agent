# Storage Layer

The `common/storage` package provides a unified infrastructure layer for database and cache operations across all services in the k8s-agent project.

## Overview

This package eliminates code duplication across 5+ services by providing:

- **MySQL Client**: GORM-based client with connection pooling, health checks, and structured logging
- **Redis Client**: go-redis wrapper with connection pooling and lifecycle management
- **Distributed Lock**: Redis-based distributed locking with safe release
- **Rate Limiter**: Token bucket rate limiting using Redis
- **Message Queue**: Redis-based FIFO queue with batch operations
- **Session Manager**: Redis-based session management with user tracking
- **Repository Pattern**: Generic CRUD operations with pagination
- **Context Utilities**: Timeout management for database operations

## Architecture

```
common/storage/
├── context.go              # Context timeout utilities
├── health.go               # Health check interface
├── mysql/                  # MySQL client implementation
│   ├── client.go          # Client with lifecycle management
│   ├── config.go          # Configuration struct
│   ├── gorm_logger.go     # GORM logger adapter
│   └── health.go          # Health check methods
├── redis/                 # Redis client implementation
│   ├── client.go          # Client with lifecycle management
│   ├── lock.go            # Distributed lock
│   ├── ratelimit.go       # Rate limiter
│   ├── queue.go           # Message queue
│   ├── session.go         # Session manager
│   └── health.go          # Health check methods
└── repository/            # Generic repository pattern
    ├── interface.go       # Repository interface
    ├── base.go            # Base CRUD implementation
    └── pagination.go      # Pagination utilities
```

## Quick Start

### MySQL Client

```go
import (
    "github.com/kart-io/k8s-agent/common/storage/mysql"
    "github.com/kart-io/logger"
)

// Create configuration
config := mysql.NewConfig()
config.Host = "localhost"
config.Port = 3306
config.User = "myuser"
config.Password = "mypassword"
config.Database = "mydb"
config.LogLevel = "info"

// Initialize logger
log := logger.New(logger.Config{
    Engine: logger.EngineZap,
    Level:  logger.LevelInfo,
})

// Create client with auto-migration
client, err := mysql.NewClient(config, log,
    mysql.WithAutoMigrate(&User{}, &Post{}),
)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Use GORM DB directly
db := client.DB()
var users []User
db.Find(&users)

// Health check
ctx := context.Background()
if err := client.Health(ctx); err != nil {
    log.Error("Database unhealthy", "error", err)
}
```

### Redis Client

```go
import (
    "github.com/kart-io/k8s-agent/common/storage/redis"
    "github.com/kart-io/k8s-agent/common/options"
    "github.com/kart-io/logger"
)

// Create configuration using options.RedisOptions
redisOpts := options.NewRedisOptions()
redisOpts.Addr = "localhost:6379"
redisOpts.DB = 0

// Validate configuration
if err := redisOpts.Validate(); err != nil {
    log.Fatal(err)
}

// Initialize logger
log := logger.New(logger.Config{
    Engine: logger.EngineZap,
    Level:  logger.LevelInfo,
})

// Create client
client, err := redis.NewClient(redisOpts, log)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Use redis client directly
rdb := client.Client()
rdb.Set(ctx, "key", "value", time.Hour)

// Health check
if err := client.Health(ctx); err != nil {
    log.Error("Redis unhealthy", "error", err)
}
```

## Features

### 1. MySQL Client

#### Configuration

```go
config := mysql.NewConfig()
config.Host = "localhost"          // MySQL host
config.Port = 3306                 // MySQL port
config.User = "root"               // Database user
config.Password = "password"       // Database password
config.Database = "mydb"           // Database name
config.Charset = "utf8mb4"         // Character set
config.MaxOpenConns = 100          // Max open connections
config.MaxIdleConns = 10           // Max idle connections
config.ConnMaxLifetime = time.Hour // Connection lifetime
config.LogLevel = "info"           // Log level: silent, error, warn, info
config.SlowQueryThreshold = 200ms  // Slow query threshold
```

#### GORM Logger

The MySQL client includes a custom GORM logger that integrates with `kart-io/logger`:

- **Structured Logging**: All SQL queries logged with structured fields
- **Slow Query Detection**: Queries exceeding threshold logged as warnings
- **Error Filtering**: "record not found" errors can be ignored
- **Performance Metrics**: Elapsed time, rows affected, SQL statement

```go
// Logged fields:
// - elapsed_ms: Query execution time in milliseconds
// - rows: Number of rows affected
// - sql: SQL statement
// - slow_query: Boolean flag for slow queries
// - error: Error message (if any)
```

#### Auto-Migration

```go
client, err := mysql.NewClient(config, log,
    mysql.WithAutoMigrate(&User{}, &Post{}, &Comment{}),
)
```

Or after client creation:

```go
err := client.AutoMigrate(&User{}, &Post{})
```

#### Health Check

```go
// With custom timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
err := client.Health(ctx)

// With default timeout
err := client.HealthCheck()

// Get connection pool statistics
stats, err := client.Stats()
fmt.Printf("Open: %d, Idle: %d, InUse: %d\n",
    stats.OpenConnections, stats.Idle, stats.InUse)
```

### 2. Redis Client

#### Configuration

Redis client uses `github.com/kart-io/k8s-agent/common/options.RedisOptions` for configuration:

```go
import "github.com/kart-io/k8s-agent/common/options"

redisOpts := options.NewRedisOptions()
redisOpts.Addr = "localhost:6379"      // Redis server address
redisOpts.Password = ""                // Redis password
redisOpts.DB = 0                       // Database index (0-15)
redisOpts.PoolSize = 10                // Connection pool size
redisOpts.MinIdleConns = 5             // Min idle connections
redisOpts.DialTimeout = 5 * time.Second  // Connection timeout
redisOpts.ReadTimeout = 3 * time.Second  // Read timeout
redisOpts.WriteTimeout = 3 * time.Second // Write timeout

// Validate configuration
if err := redisOpts.Validate(); err != nil {
    log.Fatal(err)
}
```

#### Health Check

```go
// With custom timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
err := client.Health(ctx)

// With default timeout
err := client.HealthCheck()

// Get pool statistics
stats := client.PoolStats()
fmt.Printf("Hits: %d, Misses: %d, Total: %d\n",
    stats.Hits, stats.Misses, stats.TotalConns)
```

### 3. Distributed Lock

Redis-based distributed lock with safe release using Lua scripts:

```go
import "github.com/kart-io/k8s-agent/common/storage/redis"

// Acquire lock
lock, err := client.AcquireLock(ctx, "my-resource", 10*time.Second)
if err != nil {
    // Lock already held or error
    return err
}
defer lock.Release(ctx)

// ... critical section ...

// Extend lock if needed
err = lock.Extend(ctx, 5*time.Second)
```

**Features**:

- **Unique Lock Value**: Uses UUID to prevent accidental release
- **Atomic Release**: Lua script ensures value check + delete atomicity
- **TTL Support**: Automatic expiration prevents deadlocks
- **Lock Extension**: Extend lock TTL for long-running operations

### 4. Rate Limiter

Token bucket rate limiting using Redis:

```go
import "github.com/kart-io/k8s-agent/common/storage/redis"

// Create rate limiter: 100 requests per minute
limiter := redis.NewRateLimiter(client, 100, time.Minute)

// Check if request is allowed
allowed, err := limiter.Allow(ctx, "user:123")
if err != nil {
    return err
}
if !allowed {
    return errors.New("rate limit exceeded")
}

// Get current count
count, err := limiter.GetCount(ctx, "user:123")

// Get remaining requests
remaining, err := limiter.GetRemaining(ctx, "user:123")

// Reset rate limit (admin operation)
err = limiter.Reset(ctx, "user:123")
```

**Features**:

- **Atomic Operations**: Lua script ensures INCR + EXPIRE atomicity
- **Sliding Window**: Time-based expiration for accurate rate limiting
- **Per-Key Limits**: Different limits for different users/resources
- **Query Operations**: Get count, remaining, and reset

### 5. Message Queue

Redis-based FIFO queue for asynchronous task processing:

```go
import "github.com/kart-io/k8s-agent/common/storage/redis"

// Create queue
queue := redis.NewQueue(client, "tasks")

// Push message to queue
task := map[string]interface{}{
    "id":   "task-123",
    "type": "email",
    "data": "...",
}
err := queue.Push(ctx, task)

// Pop message (blocking, waits up to timeout)
data, err := queue.Pop(ctx, 30*time.Second)
if err != nil {
    return err
}

// Unmarshal message
var task Task
json.Unmarshal(data, &task)

// Pop message (non-blocking)
data, err := queue.PopNonBlocking(ctx)
if err == redis.Nil {
    // Queue is empty
    return nil
}

// Get queue length
length, err := queue.Length(ctx)

// Peek at queue items without removing
items, err := queue.Peek(ctx, 10)  // View first 10 items

// Batch push
tasks := []interface{}{task1, task2, task3}
err := queue.PushBatch(ctx, tasks)

// Clear queue
err := queue.Clear(ctx)
```

**Features**:

- **FIFO Order**: First-in-first-out message delivery
- **Blocking/Non-blocking**: Support both blocking and non-blocking pop
- **Batch Operations**: Efficient batch push using Redis Pipeline
- **JSON Serialization**: Automatic JSON encoding/decoding
- **Peek Operation**: View queue items without removing them

**Use Cases**:
- Background task processing
- Message buffering between services
- Job queue for worker pools
- Event streaming

### 6. Session Manager

Redis-based session management with user tracking:

```go
import "github.com/kart-io/k8s-agent/common/storage/redis"

// Create session manager
sessionMgr := redis.NewSessionManager(client, "myapp")

// Session data structure
type SessionData struct {
    UserID    string
    Username  string
    Roles     []string
    LoginTime time.Time
}

// Create session
sessionData := SessionData{
    UserID:    "user-123",
    Username:  "john",
    Roles:     []string{"admin", "user"},
    LoginTime: time.Now(),
}
err := sessionMgr.Set(ctx, sessionID, sessionData, 1*time.Hour)

// Get session
var data SessionData
err := sessionMgr.Get(ctx, sessionID, &data)
if err != nil {
    // Session not found or expired
    return err
}

// Update session (keeps original TTL)
data.Roles = append(data.Roles, "moderator")
err := sessionMgr.Update(ctx, sessionID, data)

// Refresh session TTL
err := sessionMgr.Refresh(ctx, sessionID, 2*time.Hour)

// Check if session exists
exists, err := sessionMgr.Exists(ctx, sessionID)

// Get remaining TTL
ttl, err := sessionMgr.GetTTL(ctx, sessionID)

// Track user sessions (for multi-device support)
err := sessionMgr.AddUserSession(ctx, userID, sessionID, 24*time.Hour)

// Get all user sessions
sessions, err := sessionMgr.GetUserSessions(ctx, userID)

// Count active sessions
count, err := sessionMgr.CountUserSessions(ctx, userID)

// Delete specific session
err := sessionMgr.Delete(ctx, sessionID)

// Force logout (delete all user sessions)
err := sessionMgr.DeleteAllUserSessions(ctx, userID)
```

**Features**:

- **JSON Storage**: Automatic JSON serialization/deserialization
- **TTL Management**: Session expiration with refresh capability
- **User Tracking**: Track multiple sessions per user
- **Batch Operations**: Efficient multi-session deletion using Pipeline
- **Force Logout**: Delete all user sessions at once

**Use Cases**:
- User authentication sessions
- Shopping cart data
- Multi-device login management
- Session-based rate limiting
- Temporary data storage with expiration

### 7. Repository Pattern

Generic CRUD operations with Go generics:

```go
import (
    "github.com/kart-io/k8s-agent/common/storage/repository"
)

// Define your entity
type User struct {
    ID        string    `gorm:"primarykey"`
    Name      string
    Email     string
    CreatedAt time.Time
}

// Create repository
type UserRepository struct {
    *repository.BaseRepository[User, string]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{
        BaseRepository: repository.NewBaseRepository[User, string](db),
    }
}

// Use repository
repo := NewUserRepository(client.DB())

// Create
user := &User{ID: "1", Name: "John", Email: "john@example.com"}
err := repo.Create(ctx, user)

// Get by ID
user, err := repo.Get(ctx, "1")

// List with pagination
opts := repository.NewListOptions()
opts.Page = 1
opts.PageSize = 20
opts.Sort = "created_at"
opts.Order = "desc"
opts.Filters = map[string]interface{}{
    "status": "active",
}

users, total, err := repo.List(ctx, opts)

// Update
user.Name = "John Doe"
err = repo.Update(ctx, user)

// Delete
err = repo.Delete(ctx, "1")

// Check existence
exists, err := repo.Exists(ctx, "1")
```

**Features**:

- **Generic Types**: Works with any entity type
- **Pagination**: Built-in pagination support
- **Filtering**: Dynamic filter support
- **Sorting**: Configurable sort field and order
- **Error Handling**: Standardized errors (ErrNotFound, ErrAlreadyExists)

### 8. Pagination

Create paginated responses with metadata:

```go
import "github.com/kart-io/k8s-agent/common/storage/repository"

// Get entities with pagination
opts := repository.NewListOptions()
opts.Page = 2
opts.PageSize = 20

entities, total, err := repo.List(ctx, opts)

// Create pagination info
pagination := repository.NewPaginationInfo(opts, total)

// Create paged response
response := repository.NewPagedResponse(entities, pagination)

// Response includes:
// {
//   "items": [...],
//   "pagination": {
//     "current_page": 2,
//     "page_size": 20,
//     "total_pages": 5,
//     "total_items": 100,
//     "has_next": true,
//     "has_prev": true
//   }
// }
```

### 9. Context Utilities

Timeout management for database operations:

```go
import "github.com/kart-io/k8s-agent/common/storage"

// Add default timeout (5s) if context doesn't have deadline
ctx, cancel := storage.WithDefaultTimeout(context.Background())
defer cancel()

// Add custom timeout if context doesn't have deadline
ctx, cancel := storage.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// If ctx already has deadline, it's returned unchanged
```

## Migration Guide

### From `common/db/mysql.go`

**Before**:

```go
import "github.com/kart-io/k8s-agent/common/db"

client, err := db.NewMySQL(log,
    db.WithHost("localhost"),
    db.WithPort(3306),
    db.WithUser("root"),
    db.WithPassword("password"),
    db.WithDatabase("mydb"),
)
```

**After**:

```go
import "github.com/kart-io/k8s-agent/common/storage/mysql"

config := mysql.NewConfig()
config.Host = "localhost"
config.Port = 3306
config.User = "root"
config.Password = "password"
config.Database = "mydb"

client, err := mysql.NewClient(config, log)
```

### From Service-Specific Storage

**Before** (internal/agent-manager/storage/redis.go):

```go
type RedisStore struct {
    client *redis.Client
    logger core.Logger
}

func NewRedisStore(config *options.RedisOptions, log core.Logger) (*RedisStore, error) {
    client, err := config.ConnectRedis(log)
    // ... custom initialization
}

func (s *RedisStore) AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
    // ... custom lock implementation
}
```

**After**:

```go
import (
    "github.com/kart-io/k8s-agent/common/storage/redis"
    "github.com/kart-io/k8s-agent/common/options"
)

// Create Redis options
redisOpts := options.NewRedisOptions()
redisOpts.Addr = "localhost:6379"

// Validate configuration
if err := redisOpts.Validate(); err != nil {
    log.Fatal(err)
}

// Use common storage client
client, err := redis.NewClient(redisOpts, log)

// Use common lock implementation
lock, err := client.AcquireLock(ctx, "my-resource", 10*time.Second)
defer lock.Release(ctx)
```

## Best Practices

### 1. Connection Management

```go
// Create client once at application startup
client, err := mysql.NewClient(config, log)
if err != nil {
    log.Fatal(err)
}

// Always defer Close() for cleanup
defer client.Close()

// Share the client across your application
// (GORM DB is safe for concurrent use)
```

### 2. Context Timeouts

```go
// Always use context with timeout for database operations
ctx, cancel := storage.WithDefaultTimeout(context.Background())
defer cancel()

result := db.WithContext(ctx).Find(&users)
```

### 3. Health Checks

```go
// Register health check in your HTTP server
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()

    if err := mysqlClient.Health(ctx); err != nil {
        http.Error(w, "MySQL unhealthy", http.StatusServiceUnavailable)
        return
    }

    if err := redisClient.Health(ctx); err != nil {
        http.Error(w, "Redis unhealthy", http.StatusServiceUnavailable)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
})
```

### 4. Error Handling

```go
import "github.com/kart-io/k8s-agent/common/storage/repository"

user, err := repo.Get(ctx, id)
if err != nil {
    if errors.Is(err, repository.ErrNotFound) {
        return nil, errors.New("user not found")
    }
    return nil, fmt.Errorf("failed to get user: %w", err)
}
```

### 5. Distributed Locking

```go
// Always use defer to ensure lock is released
lock, err := client.AcquireLock(ctx, "critical-resource", 10*time.Second)
if err != nil {
    if strings.Contains(err.Error(), "already held") {
        return errors.New("operation in progress")
    }
    return err
}
defer lock.Release(ctx)

// ... critical section ...
```

### 6. Rate Limiting

```go
// Use user/IP-based keys for rate limiting
key := fmt.Sprintf("user:%s", userID)
allowed, err := limiter.Allow(ctx, key)
if err != nil {
    return err
}
if !allowed {
    return errors.New("rate limit exceeded, please try again later")
}
```

### 7. Message Queue

```go
// Use separate queues for different task types
highPriorityQueue := redis.NewQueue(client, "tasks:high")
normalQueue := redis.NewQueue(client, "tasks:normal")

// Always handle Pop errors properly
data, err := queue.Pop(ctx, 30*time.Second)
if err != nil {
    if err == redis.Nil || strings.Contains(err.Error(), "timeout") {
        // Queue empty or timeout - not an error
        return nil
    }
    return err
}

// Use batch operations for efficiency
tasks := collectTasks()
if err := queue.PushBatch(ctx, tasks); err != nil {
    return err
}
```

### 8. Session Management

```go
// Always set appropriate TTL for sessions
sessionTTL := 1 * time.Hour
err := sessionMgr.Set(ctx, sessionID, data, sessionTTL)

// Track user sessions for multi-device support
// Use longer TTL for user session tracking
userSessionsTTL := 24 * time.Hour
err := sessionMgr.AddUserSession(ctx, userID, sessionID, userSessionsTTL)

// Implement session refresh on activity
if exists, _ := sessionMgr.Exists(ctx, sessionID); exists {
    sessionMgr.Refresh(ctx, sessionID, sessionTTL)
}

// Clean up both session and user tracking on logout
sessionMgr.Delete(ctx, sessionID)
sessionMgr.DeleteUserSession(ctx, userID, sessionID)

// Force logout: delete all user sessions
sessionMgr.DeleteAllUserSessions(ctx, userID)
```

## Testing

The storage layer includes comprehensive unit tests using:

- **MySQL**: `github.com/DATA-DOG/go-sqlmock` for mocking
- **Redis**: `github.com/alicebob/miniredis/v2` for in-memory Redis
- **Assertions**: `github.com/stretchr/testify` for assertions

Run tests:

```bash
cd common/storage
go test -v -cover ./...
```

## Performance Considerations

### MySQL

- **Connection Pooling**: Configure `MaxOpenConns` and `MaxIdleConns` based on workload
- **Connection Lifetime**: Set `ConnMaxLifetime` to prevent stale connections
- **Slow Query Threshold**: Monitor slow queries with appropriate threshold (default: 200ms)

### Redis

- **Connection Pooling**: Adjust `PoolSize` and `MinIdleConns` for optimal performance
- **Timeouts**: Set appropriate `DialTimeout`, `ReadTimeout`, `WriteTimeout`
- **Pipeline**: Use Redis pipeline for batch operations (via `client.Client()`)

### Rate Limiting

- **Lua Scripts**: Atomic operations prevent race conditions
- **Key Design**: Use hierarchical keys for flexible rate limiting
- **Window Size**: Balance between accuracy and Redis memory usage

## Dependencies

- `gorm.io/gorm` v1.25.0 - GORM ORM
- `gorm.io/driver/mysql` v1.5.0 - MySQL driver
- `github.com/redis/go-redis/v9` v9.0.0 - Redis client
- `github.com/kart-io/logger` v0.1.0 - Structured logging
- `github.com/google/uuid` v1.6.0 - UUID generation for locks

## License

This package is part of the k8s-agent project.
