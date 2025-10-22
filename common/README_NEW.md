# Aetherius Common Library - 公共代码库

通用功能包，提供 Aetherius k8s-agent 项目中所有服务共享的核心功能。

## 📦 包结构

```
common/
├── response/          # 统一的 API 响应格式
├── errors/            # 错误码和错误处理
├── pagination/        # 分页功能
├── logger/            # 日志工具（基于 Zap）
├── k8sutils/          # Kubernetes 资源转换工具
├── validator/         # 数据验证工具
├── middleware/        # Gin 中间件
├── utils/             # 工具函数
├── db/                # 🆕 数据库连接封装
│   ├── mysql.go       # MySQL/GORM 连接
│   └── redis.go       # Redis 连接
├── mq/                # 🆕 消息队列封装
│   └── nats.go        # NATS 客户端
├── config/            # 🆕 配置加载
│   └── loader.go      # Viper 配置加载器
└── server/            # 🆕 HTTP 服务器
    ├── gin.go         # Gin 服务器封装
    ├── health.go      # 健康检查
    └── shutdown.go    # 优雅关闭
```

## 🆕 新增功能

### 1. `db/` - 数据库连接

#### MySQL/GORM

```go
import "github.com/kart-io/k8s-agent/common/db"

// MySQL 配置
cfg := db.MySQLConfig{
    Host:            "localhost",
    Port:            3306,
    User:            "root",
    Password:        "password",
    Database:        "mydb",
    MaxOpenConns:    100,
    MaxIdleConns:    10,
    ConnMaxLifetime: time.Hour,
    LogLevel:        "info", // silent, error, warn, info
}

// 创建连接
mysql, err := db.NewMySQL(cfg, logger)
if err != nil {
    log.Fatal(err)
}
defer mysql.Close()

// 使用 GORM
var users []User
mysql.DB.Find(&users)

// 自动迁移
mysql.AutoMigrate(&User{}, &Order{})

// 健康检查
if err := mysql.Health(ctx); err != nil {
    log.Error("MySQL unhealthy", zap.Error(err))
}
```

#### Redis

```go
import "github.com/kart-io/k8s-agent/common/db"

// Redis 配置
cfg := db.RedisConfig{
    Addr:         "localhost:6379",
    Password:     "",
    DB:           0,
    PoolSize:     10,
    MinIdleConns: 5,
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
}

// 创建连接
redis, err := db.NewRedis(cfg, logger)
if err != nil {
    log.Fatal(err)
}
defer redis.Close()

// 设置缓存 (自动 JSON 序列化)
user := User{ID: 1, Name: "Alice"}
redis.Set(ctx, "user:1", user, 10*time.Minute)

// 获取缓存 (自动 JSON 反序列化)
var cachedUser User
redis.Get(ctx, "user:1", &cachedUser)

// 字符串操作
redis.SetString(ctx, "key", "value", time.Hour)
val, _ := redis.GetString(ctx, "key")

// 其他操作
redis.Delete(ctx, "key1", "key2")
exists, _ := redis.Exists(ctx, "key")
redis.Incr(ctx, "counter")
```

### 2. `mq/` - 消息队列

#### NATS

```go
import "github.com/kart-io/k8s-agent/common/mq"

// NATS 配置
cfg := mq.NATSConfig{
    URL:              "nats://localhost:4222",
    MaxReconnects:    10,
    ReconnectWait:    2 * time.Second,
    ReconnectBufSize: 1024 * 1024,
}

// 创建连接
nats, err := mq.NewNATS(cfg, logger)
if err != nil {
    log.Fatal(err)
}
defer nats.Close()

// 发布消息
nats.Publish("events.pod.created", []byte(`{"pod": "nginx"}`))

// 订阅消息
nats.Subscribe("events.*", func(msg *nats.Msg) {
    log.Info("Received", zap.String("subject", msg.Subject))
})

// 队列订阅 (负载均衡)
nats.QueueSubscribe("tasks", "workers", func(msg *nats.Msg) {
    // 处理任务
})

// 请求-响应模式
response, err := nats.Request("query.pods", []byte("list"), 5*time.Second)
```

### 3. `config/` - 配置加载

```go
import "github.com/kart-io/k8s-agent/common/config"

// 定义配置结构
type AppConfig struct {
    Server struct {
        Host string `mapstructure:"host"`
        Port int    `mapstructure:"port"`
    } `mapstructure:"server"`
    Database db.MySQLConfig   `mapstructure:"database"`
    Redis    db.RedisConfig   `mapstructure:"redis"`
    NATS     mq.NATSConfig    `mapstructure:"nats"`
}

// 加载配置文件
var cfg AppConfig
if err := config.LoadConfig("configs/config.yaml", &cfg); err != nil {
    log.Fatal(err)
}

// 带默认值加载
defaults := map[string]interface{}{
    "server.port": 8080,
    "database.max_open_conns": 100,
}
config.LoadConfigWithDefaults("configs/config.yaml", &cfg, defaults)

// 仅从环境变量加载
config.LoadConfigFromEnv("APP", &cfg)
```

**配置文件示例 (config.yaml)**:

```yaml
server:
  host: 0.0.0.0
  port: 8080
  mode: release

database:
  host: localhost
  port: 3306
  user: root
  password: password
  database: mydb
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 1h
  log_level: silent

redis:
  addr: localhost:6379
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 5

nats:
  url: nats://localhost:4222
  max_reconnects: 10
  reconnect_wait: 2s
```

### 4. `server/` - HTTP 服务器

#### Gin 服务器

```go
import "github.com/kart-io/k8s-agent/common/server"

// 服务器配置
cfg := server.GinServerConfig{
    Host:         "0.0.0.0",
    Port:         8080,
    Mode:         "release", // debug, release, test
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  60 * time.Second,
}

// 创建服务器
srv := server.NewGinServer(cfg, logger)

// 注册路由
srv.Engine.GET("/api/v1/users", GetUsers)
srv.Engine.POST("/api/v1/users", CreateUser)

// 启动服务器 (goroutine)
go func() {
    if err := srv.Run(); err != nil {
        log.Fatal(err)
    }
}()
```

#### 健康检查

```go
// 注册健康检查端点
healthCheckers := map[string]server.HealthChecker{
    "mysql": mysqlClient,
    "redis": redisClient,
    "nats":  natsClient,
}
server.RegisterHealthEndpoints(srv.Engine, healthCheckers)

// 访问端点:
// GET /health       - 简单存活检查
// GET /healthz      - K8s 风格存活检查
// GET /liveness     - 不检查依赖
// GET /readiness    - 检查所有依赖
```

**响应示例**:

```json
{
  "status": "ready",
  "checks": {
    "mysql": { "status": "healthy" },
    "redis": { "status": "healthy" },
    "nats": { "status": "healthy" }
  }
}
```

#### 优雅关闭

```go
// 优雅关闭 (自动监听 SIGINT/SIGTERM)
server.GracefulShutdown(logger, 30*time.Second,
    srv,          // HTTP 服务器
    mysqlClient,  // MySQL 连接
    redisClient,  // Redis 连接
    natsClient,   // NATS 连接
)

// 或者自定义关闭逻辑
sig := server.WaitForShutdown(logger)
// 执行自定义关闭逻辑
```

## 🚀 完整使用示例

```go
package main

import (
    "flag"
    "time"

    "go.uber.org/zap"

    "github.com/kart-io/k8s-agent/common/config"
    "github.com/kart-io/k8s-agent/common/db"
    "github.com/kart-io/k8s-agent/common/logger"
    "github.com/kart-io/k8s-agent/common/mq"
    "github.com/kart-io/k8s-agent/common/server"
)

// AppConfig 应用配置
type AppConfig struct {
    Server   server.GinServerConfig `mapstructure:"server"`
    Database db.MySQLConfig         `mapstructure:"database"`
    Redis    db.RedisConfig         `mapstructure:"redis"`
    NATS     mq.NATSConfig          `mapstructure:"nats"`
    Log      logger.Config          `mapstructure:"log"`
}

func main() {
    // 解析命令行参数
    configPath := flag.String("config", "configs/config.yaml", "config file path")
    flag.Parse()

    // 加载配置
    var cfg AppConfig
    if err := config.LoadConfig(*configPath, &cfg); err != nil {
        panic(err)
    }

    // 初始化日志
    if err := logger.Init(&cfg.Log); err != nil {
        panic(err)
    }
    defer logger.Sync()
    log := logger.GetLogger()

    // 初始化数据库
    mysqlClient, err := db.NewMySQL(cfg.Database, log)
    if err != nil {
        log.Fatal("Failed to connect to MySQL", zap.Error(err))
    }
    defer mysqlClient.Close()

    // 初始化 Redis
    redisClient, err := db.NewRedis(cfg.Redis, log)
    if err != nil {
        log.Fatal("Failed to connect to Redis", zap.Error(err))
    }
    defer redisClient.Close()

    // 初始化 NATS
    natsClient, err := mq.NewNATS(cfg.NATS, log)
    if err != nil {
        log.Fatal("Failed to connect to NATS", zap.Error(err))
    }
    defer natsClient.Close()

    // 初始化 HTTP 服务器
    ginServer := server.NewGinServer(cfg.Server, log)

    // 注册健康检查
    healthCheckers := map[string]server.HealthChecker{
        "mysql": mysqlClient,
        "redis": redisClient,
        "nats":  natsClient,
    }
    server.RegisterHealthEndpoints(ginServer.Engine, healthCheckers)

    // 注册业务路由
    RegisterRoutes(ginServer.Engine, mysqlClient, redisClient, natsClient, log)

    // 启动服务器 (goroutine)
    go func() {
        if err := ginServer.Run(); err != nil {
            log.Fatal("Server error", zap.Error(err))
        }
    }()

    // 优雅关闭
    server.GracefulShutdown(log, 30*time.Second,
        ginServer, mysqlClient, redisClient, natsClient,
    )
}

func RegisterRoutes(engine *gin.Engine, mysql *db.MySQLClient, redis *db.RedisClient, nats *mq.NATSClient, log *zap.Logger) {
    v1 := engine.Group("/api/v1")
    {
        v1.GET("/users", func(c *gin.Context) {
            // 业务逻辑
        })
    }
}
```

## 📝 迁移指南

### 从旧代码迁移到新 common 库

#### 之前 (每个服务重复实现)

```go
// agent-manager/internal/storage/postgres.go
type PostgresStore struct {
    db *gorm.DB
}

func NewPostgresStore(config Config) (*PostgresStore, error) {
    dsn := fmt.Sprintf("...")
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    // ... 100+ 行连接池配置
}

// orchestrator-service/internal/storage/postgres.go
type PostgresStore struct {
    db *gorm.DB
}

func NewPostgresStore(config Config) (*PostgresStore, error) {
    dsn := fmt.Sprintf("...")
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    // ... 重复的 100+ 行代码
}
```

#### 之后 (使用 common 库)

```go
// agent-manager/internal/storage/storage.go
import "github.com/kart-io/k8s-agent/common/db"

// 直接使用
mysql, err := db.NewMySQL(cfg.Database, log)
```

### 收益对比

| 功能 | 旧代码行数 | 新代码行数 | 节省 |
|------|-----------|-----------|------|
| MySQL 连接 | ~150 行 × 4 服务 | 3 行 | **99%** |
| Redis 连接 | ~120 行 × 4 服务 | 3 行 | **99%** |
| NATS 连接 | ~100 行 × 3 服务 | 3 行 | **99%** |
| HTTP 服务器 | ~200 行 × 4 服务 | 10 行 | **99%** |
| 健康检查 | ~100 行 × 4 服务 | 2 行 | **99%** |
| 优雅关闭 | ~80 行 × 4 服务 | 2 行 | **99%** |

## 📖 原有功能文档

详细的原有功能文档请参考之前的 [README.md](README.md.backup)，包括：

- `response/` - API 响应格式
- `errors/` - 错误处理
- `pagination/` - 分页
- `logger/` - 日志
- `k8sutils/` - K8s 工具
- `validator/` - 验证
- `middleware/` - 中间件

## 🔧 依赖管理

在服务的 `go.mod` 中引用 common 包：

```go
require (
    github.com/kart-io/k8s-agent/common v0.0.0
)

replace github.com/kart-io/k8s-agent/common => ../common
```

然后运行:

```bash
go mod tidy
```

## ✅ 最佳实践

1. **统一初始化顺序**: 日志 → 配置 → 数据库 → 消息队列 → HTTP 服务器
2. **使用健康检查**: 所有依赖都应实现 `HealthChecker` 接口
3. **优雅关闭**: 使用 `GracefulShutdown` 确保资源正确释放
4. **结构化日志**: 所有操作都使用 `zap.Logger` 记录
5. **配置外部化**: 敏感信息通过环境变量或配置文件管理

## 📊 测试

```bash
# 运行所有测试
cd common && go test ./...

# 运行特定包测试
go test ./db
go test ./mq
go test ./server
```

## 🤝 贡献

添加新功能时请遵循：

1. 功能必须是多个服务共享的
2. 提供清晰的文档和示例
3. 包含完整的单元测试
4. 更新本 README

## 版本

- **v0.1.0** - 初始版本 (response, errors, pagination, logger, k8sutils, validator, middleware)
- **v0.2.0** - 新增 db, mq, config, server 包

## 许可证

MIT License
