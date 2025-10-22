# Common Library - Options Pattern 使用指南

本文档展示如何使用 Functional Options 模式配置 common 库中的各个组件。

## 1. MySQL 连接

### 基础用法(使用默认配置)

```go
import (
    "github.com/kart-io/k8s-agent/common/db"
    "go.uber.org/zap"
)

// 使用默认配置创建 MySQL 连接
mysql, err := db.NewMySQL(logger)
if err != nil {
    log.Fatal(err)
}
defer mysql.Close()
```

### 自定义配置

```go
// 自定义配置
mysql, err := db.NewMySQL(logger,
    db.WithHost("192.168.1.100"),
    db.WithPort(3306),
    db.WithUser("myuser"),
    db.WithPassword("mypassword"),
    db.WithDatabase("mydb"),
    db.WithMaxOpenConns(200),
    db.WithMaxIdleConns(20),
    db.WithConnMaxLifetime(2*time.Hour),
    db.WithLogLevel("info"),
)
```

### 部分自定义(其他使用默认值)

```go
// 只自定义部分配置,其他使用默认值
mysql, err := db.NewMySQL(logger,
    db.WithHost("prod-db.example.com"),
    db.WithDatabase("production"),
    db.WithLogLevel("error"),
)
```

## 2. Redis 连接

### 基础用法

```go
import "github.com/kart-io/k8s-agent/common/db"

// 使用默认配置
redis, err := db.NewRedis(logger)
```

### 自定义配置

```go
redis, err := db.NewRedis(logger,
    db.WithAddr("redis.example.com:6379"),
    db.WithRedisPassword("secretpassword"),
    db.WithRedisDB(1),
    db.WithRedisPoolSize(20),
    db.WithRedisMinIdleConns(10),
    db.WithRedisDialTimeout(10*time.Second),
    db.WithRedisReadTimeout(5*time.Second),
    db.WithRedisWriteTimeout(5*time.Second),
)
```

### 部分自定义

```go
redis, err := db.NewRedis(logger,
    db.WithAddr("localhost:6380"),
    db.WithRedisDB(2),
)
```

## 3. NATS 消息队列

### 基础用法

```go
import "github.com/kart-io/k8s-agent/common/mq"

// 使用默认配置
nats, err := mq.NewNATS(logger)
```

### 自定义配置

```go
nats, err := mq.NewNATS(logger,
    mq.WithNATSURL("nats://nats.example.com:4222"),
    mq.WithNATSMaxReconnects(20),
    mq.WithNATSReconnectWait(5*time.Second),
    mq.WithNATSReconnectBufSize(2*1024*1024), // 2MB
)
```

## 4. Gin HTTP 服务器

### 基础用法

```go
import "github.com/kart-io/k8s-agent/common/server"

// 使用默认配置
ginServer := server.NewGinServer(logger)
```

### 自定义配置

```go
ginServer := server.NewGinServer(logger,
    server.WithGinHost("0.0.0.0"),
    server.WithGinPort(9090),
    server.WithGinMode("debug"),
    server.WithGinReadTimeout(30*time.Second),
    server.WithGinWriteTimeout(30*time.Second),
    server.WithGinIdleTimeout(120*time.Second),
)
```

## 5. 完整应用示例

```go
package main

import (
    "time"

    "go.uber.org/zap"

    "github.com/kart-io/k8s-agent/common/db"
    "github.com/kart-io/k8s-agent/common/mq"
    "github.com/kart-io/k8s-agent/common/server"
)

func main() {
    // 初始化日志
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // 初始化 MySQL
    mysqlClient, err := db.NewMySQL(logger,
        db.WithHost("localhost"),
        db.WithDatabase("myapp"),
        db.WithUser("root"),
        db.WithPassword("password"),
    )
    if err != nil {
        logger.Fatal("Failed to connect to MySQL", zap.Error(err))
    }
    defer mysqlClient.Close()

    // 初始化 Redis
    redisClient, err := db.NewRedis(logger,
        db.WithAddr("localhost:6379"),
    )
    if err != nil {
        logger.Fatal("Failed to connect to Redis", zap.Error(err))
    }
    defer redisClient.Close()

    // 初始化 NATS
    natsClient, err := mq.NewNATS(logger,
        mq.WithNATSURL("nats://localhost:4222"),
    )
    if err != nil {
        logger.Fatal("Failed to connect to NATS", zap.Error(err))
    }
    defer natsClient.Close()

    // 初始化 HTTP 服务器
    ginServer := server.NewGinServer(logger,
        server.WithGinPort(8080),
        server.WithGinMode("release"),
    )

    // 注册健康检查
    healthCheckers := map[string]server.HealthChecker{
        "mysql": mysqlClient,
        "redis": redisClient,
        "nats":  natsClient,
    }
    server.RegisterHealthEndpoints(ginServer.Engine, healthCheckers)

    // 注册业务路由
    ginServer.Engine.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello World"})
    })

    // 启动服务器 (goroutine)
    go func() {
        if err := ginServer.Run(); err != nil {
            logger.Fatal("Server error", zap.Error(err))
        }
    }()

    // 优雅关闭
    server.GracefulShutdown(logger, 30*time.Second,
        ginServer, mysqlClient, redisClient, natsClient,
    )
}
```

## Options 模式的优势

1. **默认值支持**: 不指定的配置项自动使用合理的默认值
2. **灵活性**: 可以只配置需要修改的参数
3. **可读性**: 配置代码清晰易懂
4. **可扩展性**: 添加新配置项不影响现有代码
5. **类型安全**: 编译时检查参数类型

## 默认配置值

### MySQL 默认配置
- Host: `localhost`
- Port: `3306`
- User: `root`
- Password: ` `(空字符串)
- Database: `test`
- MaxOpenConns: `100`
- MaxIdleConns: `10`
- ConnMaxLifetime: `1小时`
- LogLevel: `silent`

### Redis 默认配置
- Addr: `localhost:6379`
- Password: ``(空字符串)
- DB: `0`
- PoolSize: `10`
- MinIdleConns: `5`
- DialTimeout: `5秒`
- ReadTimeout: `3秒`
- WriteTimeout: `3秒`

### NATS 默认配置
- URL: `nats://localhost:4222`
- MaxReconnects: `10`
- ReconnectWait: `2秒`
- ReconnectBufSize: `1MB`

### Gin Server 默认配置
- Host: `0.0.0.0`
- Port: `8080`
- Mode: `release`
- ReadTimeout: `10秒`
- WriteTimeout: `10秒`
- IdleTimeout: `60秒`
