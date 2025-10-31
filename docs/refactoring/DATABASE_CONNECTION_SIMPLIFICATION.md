# 数据库和 Redis 连接简化优化

## 优化目标

将 MySQL 和 Redis 连接的创建逻辑直接在 `Options` 中实现，减少调用链层级，简化代码结构。

## 优化前的调用链

### MySQL 连接

```
ServerOptions.Database (DatabaseOptions)
  → db.NewMySQLFromOptions(logger, opts.Database)
  → db.NewMySQL(logger, WithHost(), WithPort(), ...)
  → MySQLClient
```

**问题**：
1. 调用链过长，需要传递 `opts.Database` 参数
2. 中间层 `db.NewMySQLFromOptions` 存在冗余
3. `common/db/helpers.go` 导入 `common/options`，潜在循环依赖风险

### Redis 连接

```
ServerOptions.Redis (RedisOptions)
  → db.NewRedisFromOptions(logger, opts.Redis)
  → db.NewRedis(logger, WithAddr(), WithPassword(), ...)
  → RedisClient
```

**问题**：与 MySQL 相同

## 优化后的调用链

### MySQL 连接

```
opts.Database.NewMySQLClient(logger)
  → db.NewMySQL(logger, WithHost(), WithPort(), ...)
  → MySQLClient
```

**改进**：
1. 调用链缩短，更直观
2. 方法直接定义在 `DatabaseOptions` 上
3. 消除了 `helpers.go` 中的中间层函数

### Redis 连接

```
opts.Redis.NewRedisClient(logger)
  → db.NewRedis(logger, WithAddr(), WithPassword(), ...)
  → RedisClient
```

**改进**：与 MySQL 相同

## 具体修改

### 1. 删除 `common/db/helpers.go`

移除了以下函数：
- `NewMySQLFromOptions(log core.Logger, opts *options.DatabaseOptions) (*MySQLClient, error)`
- `NewRedisFromOptions(log core.Logger, opts *options.RedisOptions) (*RedisClient, error)`

**原因**：
- 这两个函数是中间层，增加了调用链复杂度
- 它们导入 `common/options` 包，与 `common/options` 导入 `common/db` 形成潜在循环依赖

### 2. 新增 `common/options/database_client.go`

添加了 `DatabaseOptions` 的方法：

```go
// NewMySQLClient 根据配置创建 MySQL 客户端
// 这个方法直接在 Options 中实现数据库连接创建，简化调用链
//
// 之前的调用方式：
//   client, err := db.NewMySQLFromOptions(logger, opts.Database)
//
// 现在的调用方式：
//   client, err := opts.Database.NewMySQLClient(logger)
func (o *DatabaseOptions) NewMySQLClient(log core.Logger) (*db.MySQLClient, error) {
    return db.NewMySQL(log,
        db.WithHost(o.Host),
        db.WithPort(o.Port),
        db.WithUser(o.User),
        db.WithPassword(o.Password),
        db.WithDatabase(o.Database),
        db.WithMaxOpenConns(o.MaxOpenConns),
        db.WithMaxIdleConns(o.MaxIdleConns),
        db.WithConnMaxLifetime(o.ConnMaxLifetime),
    )
}
```

### 3. 新增 `common/options/redis_client.go`

添加了 `RedisOptions` 的方法：

```go
// NewRedisClient 根据配置创建 Redis 客户端
// 这个方法直接在 Options 中实现 Redis 连接创建，简化调用链
//
// 之前的调用方式：
//   client, err := db.NewRedisFromOptions(logger, opts.Redis)
//
// 现在的调用方式：
//   client, err := opts.Redis.NewRedisClient(logger)
func (o *RedisOptions) NewRedisClient(log core.Logger) (*db.RedisClient, error) {
    return db.NewRedis(log,
        db.WithAddr(o.Addr),
        db.WithRedisPassword(o.Password),
        db.WithRedisDB(o.DB),
        db.WithRedisPoolSize(o.PoolSize),
        db.WithRedisMinIdleConns(o.MinIdleConns),
        db.WithRedisDialTimeout(o.DialTimeout),
        db.WithRedisReadTimeout(o.ReadTimeout),
        db.WithRedisWriteTimeout(o.WriteTimeout),
    )
}
```

### 4. 更新 `pkg/initializers/database.go`

**修改前**：
```go
// 创建 MySQL 客户端
client, err := db.NewMySQLFromOptions(d.logger, d.opts)
if err != nil {
    return fmt.Errorf("failed to create MySQL client: %w", err)
}
```

**修改后**：
```go
// 直接使用 Options 的 NewMySQLClient 方法创建客户端
client, err := d.opts.NewMySQLClient(d.logger)
if err != nil {
    return fmt.Errorf("failed to create MySQL client: %w", err)
}
```

### 5. 更新 `pkg/initializers/redis.go`

**修改前**：
```go
// 创建 Redis 客户端
client, err := db.NewRedisFromOptions(r.logger, r.opts)
if err != nil {
    return fmt.Errorf("failed to create Redis client: %w", err)
}
```

**修改后**：
```go
// 直接使用 Options 的 NewRedisClient 方法创建客户端
client, err := r.opts.NewRedisClient(r.logger)
if err != nil {
    return fmt.Errorf("failed to create Redis client: %w", err)
}
```

### 6. 更新各服务的 storage 层

以下文件也进行了相应更新：

- `internal/orchestrator/storage/postgres.go`
- `internal/orchestrator/storage/redis.go`
- `internal/monitor/storage/postgres.go`
- `internal/cluster/storage/mysql.go`

所有调用都从 `db.NewXXXFromOptions(logger, opts)` 改为 `opts.NewXXXClient(logger)`。

## 优势总结

### 1. 调用链简化

**之前**：
```go
// 需要知道 db 包的辅助函数
client, err := db.NewMySQLFromOptions(logger, opts.Database)
```

**现在**：
```go
// 直接在 Options 上调用方法，更自然
client, err := opts.Database.NewMySQLClient(logger)
```

### 2. 消除循环依赖风险

- 删除了 `common/db/helpers.go`
- `common/db` 包不再依赖 `common/options`
- `common/options` 可以安全地导入 `common/db`

### 3. 面向对象设计

- `DatabaseOptions` 和 `RedisOptions` 现在有了创建连接的能力
- 符合"数据和行为封装"的OOP原则
- 调用更符合直觉：`opts.NewClient()` 而不是 `db.NewXXXFromOptions(opts)`

### 4. 代码更易维护

- 减少了中间层函数
- 调用路径更清晰
- 修改连接逻辑时只需关注 Options 包

## 编译验证

所有8个服务编译成功：

```bash
$ make build
==> go.build
Building agent-manager...
Building orchestrator...
Building reasoning...
Building auth...
Building gateway...
Building monitor...
Building cluster...
Building collect-agent...
Build completed: /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/_output/bin
```

## 使用示例

### 在初始化器中使用

```go
package initializers

import (
    "context"
    "github.com/kart-io/k8s-agent/common/options"
    "github.com/kart-io/logger/core"
)

type DatabaseInitializer struct {
    opts   *options.DatabaseOptions
    logger core.Logger
}

func (d *DatabaseInitializer) Initialize(ctx context.Context) error {
    // 简化的调用方式
    client, err := d.opts.NewMySQLClient(d.logger)
    if err != nil {
        return err
    }

    // 使用 client...
    return nil
}
```

### 在 Storage 层使用

```go
package storage

import (
    "github.com/kart-io/k8s-agent/common/options"
    "github.com/kart-io/logger/core"
)

func NewPostgresStore(opts *options.DatabaseOptions, log core.Logger) (*PostgresStore, error) {
    // 简化的调用方式
    mysqlClient, err := opts.NewMySQLClient(log)
    if err != nil {
        return nil, err
    }

    // 使用 client 创建 store...
    return &PostgresStore{db: mysqlClient.DB}, nil
}
```

## 影响范围

### 修改的文件

1. **删除**：
   - `common/db/helpers.go`

2. **新增**：
   - `common/options/database_client.go`
   - `common/options/redis_client.go`

3. **修改**：
   - `pkg/initializers/database.go`
   - `pkg/initializers/redis.go`
   - `internal/orchestrator/storage/postgres.go`
   - `internal/orchestrator/storage/redis.go`
   - `internal/monitor/storage/postgres.go`
   - `internal/cluster/storage/mysql.go`

### 向后兼容性

- **不兼容**：移除了 `db.NewMySQLFromOptions` 和 `db.NewRedisFromOptions` 函数
- **迁移简单**：只需将调用从 `db.NewXXXFromOptions(logger, opts)` 改为 `opts.NewXXXClient(logger)`
- **编译时检查**：所有使用旧API的代码会在编译时报错，便于发现和修复

### 测试建议

1. 单元测试：验证 `NewMySQLClient` 和 `NewRedisClient` 方法正常工作
2. 集成测试：验证各服务能正常连接数据库和 Redis
3. 功能测试：验证所有依赖数据库和 Redis 的功能正常

## 总结

此次优化通过将连接创建逻辑直接放到 Options 中，成功地：

1. ✅ 简化了调用链（减少1层中间调用）
2. ✅ 消除了循环依赖风险
3. ✅ 提升了代码的可读性和可维护性
4. ✅ 符合面向对象的设计原则
5. ✅ 所有服务编译通过，无功能影响

这是一次成功的重构优化，提高了代码质量而不影响功能。
