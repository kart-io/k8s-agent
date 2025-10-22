# Aetherius Common Library - 公共代码重构方案

## 📋 现状分析

### 已有的 Common 库功能

- ✅ `response/` - 统一 API 响应格式
- ✅ `errors/` - 错误码和错误处理
- ✅ `pagination/` - 分页功能
- ✅ `logger/` - 基于 Zap 的日志工具
- ✅ `k8sutils/` - Kubernetes 资源转换工具
- ✅ `validator/` - 数据验证工具
- ✅ `middleware/` - Gin 中间件 (CORS, 限流, JWT, 日志, 恢复)
- ✅ `utils/` - 网络工具

### 各服务中重复的代码

通过分析 `agent-manager`, `orchestrator-service`, `reasoning-service-go`, `auth-service` 等服务，发现以下重复代码模式：

#### 1. 数据库连接 (MySQL/PostgreSQL)
- ✅ 每个服务都有类似的 GORM 连接代码
- ✅ 连接池配置重复
- ✅ 健康检查逻辑重复

**位置**:
- `agent-manager/internal/storage/postgres.go`
- `orchestrator-service/internal/storage/postgres.go`
- `auth-service/internal/storage/postgres.go`

#### 2. Redis 连接
- ✅ Redis 客户端初始化重复
- ✅ 缓存操作封装重复 (Get/Set/Delete/Exists)

**位置**:
- `agent-manager/internal/storage/redis.go`
- `orchestrator-service/internal/storage/redis.go`
- `auth-service/internal/storage/redis.go`

#### 3. NATS 消息队列
- ✅ NATS 连接逻辑重复
- ✅ Pub/Sub 封装重复

**位置**:
- `agent-manager/internal/nats/`
- `orchestrator-service/internal/subscriber/`

#### 4. 配置加载
- ✅ Viper 配置加载逻辑重复
- ✅ 环境变量覆盖逻辑重复
- ✅ 配置验证重复

**位置**:
- `agent-manager/internal/config/`
- `orchestrator-service/internal/config/`
- `reasoning-service-go/internal/config/`
- `auth-service/internal/config/`

#### 5. HTTP 服务器启动
- ✅ Gin 路由器初始化重复
- ✅ 中间件配置重复
- ✅ 优雅关闭逻辑重复

#### 6. 健康检查端点
- ✅ `/health` 端点逻辑重复
- ✅ `/readiness` 端点逻辑重复
- ✅ 依赖健康检查重复

#### 7. 信号处理
- ✅ 优雅关闭逻辑重复
- ✅ 信号捕获重复

---

## 🎯 重构目标

### 1. 扩展 Common 库

在现有 `common/` 基础上，新增以下包：

```
common/
├── response/          # ✅ 已有 - 统一 API 响应
├── errors/            # ✅ 已有 - 错误处理
├── pagination/        # ✅ 已有 - 分页
├── logger/            # ✅ 已有 - 日志
├── k8sutils/          # ✅ 已有 - K8s 工具
├── validator/         # ✅ 已有 - 验证
├── middleware/        # ✅ 已有 - Gin 中间件
├── utils/             # ✅ 已有 - 工具函数
├── db/                # 🆕 数据库连接封装
│   ├── mysql.go       # MySQL/GORM 连接
│   ├── redis.go       # Redis 连接
│   └── neo4j.go       # Neo4j 连接
├── mq/                # 🆕 消息队列封装
│   ├── nats.go        # NATS 客户端
│   └── publisher.go   # 发布者接口
├── config/            # 🆕 配置加载
│   ├── loader.go      # Viper 配置加载器
│   ├── types.go       # 通用配置类型
│   └── validator.go   # 配置验证
├── server/            # 🆕 HTTP 服务器
│   ├── gin.go         # Gin 服务器封装
│   ├── health.go      # 健康检查
│   └── shutdown.go    # 优雅关闭
├── telemetry/         # 🆕 可观测性
│   ├── metrics.go     # Prometheus 指标
│   └── tracing.go     # OpenTelemetry 追踪
└── auth/              # 🆕 认证授权
    ├── jwt.go         # JWT 工具
    └── rbac.go        # RBAC 权限检查
```

---

## 📦 新增包设计

### 1. `common/db/` - 数据库连接

#### `mysql.go` - MySQL/GORM 封装

```go
package db

import (
    "context"
    "fmt"
    "time"

    "go.uber.org/zap"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// MySQLConfig MySQL 配置
type MySQLConfig struct {
    Host            string
    Port            int
    User            string
    Password        string
    Database        string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    LogLevel        string // silent, error, warn, info
}

// MySQLClient MySQL 客户端
type MySQLClient struct {
    DB     *gorm.DB
    logger *zap.Logger
}

// NewMySQL 创建 MySQL 连接
func NewMySQL(cfg MySQLConfig, log *zap.Logger) (*MySQLClient, error) {
    dsn := fmt.Sprintf(
        "%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
    )

    // GORM 日志级别
    gormLogLevel := logger.Silent
    switch cfg.LogLevel {
    case "error":
        gormLogLevel = logger.Error
    case "warn":
        gormLogLevel = logger.Warn
    case "info":
        gormLogLevel = logger.Info
    }

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(gormLogLevel),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
    }

    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get database instance: %w", err)
    }

    // 连接池配置
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
    sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

    // 测试连接
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := sqlDB.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping MySQL: %w", err)
    }

    log.Info("MySQL connected",
        zap.String("host", cfg.Host),
        zap.String("database", cfg.Database),
    )

    return &MySQLClient{
        DB:     db,
        logger: log,
    }, nil
}

// Health 检查数据库健康
func (c *MySQLClient) Health(ctx context.Context) error {
    sqlDB, err := c.DB.DB()
    if err != nil {
        return err
    }
    return sqlDB.PingContext(ctx)
}

// Close 关闭连接
func (c *MySQLClient) Close() error {
    sqlDB, err := c.DB.DB()
    if err != nil {
        return err
    }
    return sqlDB.Close()
}
```

#### `redis.go` - Redis 封装

```go
package db

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
)

// RedisConfig Redis 配置
type RedisConfig struct {
    Addr         string
    Password     string
    DB           int
    PoolSize     int
    MinIdleConns int
    DialTimeout  time.Duration
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
}

// RedisClient Redis 客户端
type RedisClient struct {
    Client *redis.Client
    logger *zap.Logger
}

// NewRedis 创建 Redis 连接
func NewRedis(cfg RedisConfig, log *zap.Logger) (*RedisClient, error) {
    client := redis.NewClient(&redis.Options{
        Addr:         cfg.Addr,
        Password:     cfg.Password,
        DB:           cfg.DB,
        PoolSize:     cfg.PoolSize,
        MinIdleConns: cfg.MinIdleConns,
        DialTimeout:  cfg.DialTimeout,
        ReadTimeout:  cfg.ReadTimeout,
        WriteTimeout: cfg.WriteTimeout,
    })

    // 测试连接
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("failed to connect to Redis: %w", err)
    }

    log.Info("Redis connected", zap.String("addr", cfg.Addr))

    return &RedisClient{
        Client: client,
        logger: log,
    }, nil
}

// Get 获取缓存 (支持 JSON 反序列化)
func (c *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
    val, err := c.Client.Get(ctx, key).Result()
    if err != nil {
        return err
    }
    return json.Unmarshal([]byte(val), dest)
}

// Set 设置缓存 (支持 JSON 序列化)
func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return c.Client.Set(ctx, key, data, expiration).Err()
}

// Delete 删除缓存
func (c *RedisClient) Delete(ctx context.Context, keys ...string) error {
    return c.Client.Del(ctx, keys...).Err()
}

// Exists 检查 key 是否存在
func (c *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
    count, err := c.Client.Exists(ctx, key).Result()
    return count > 0, err
}

// Health 检查 Redis 健康
func (c *RedisClient) Health(ctx context.Context) error {
    return c.Client.Ping(ctx).Err()
}

// Close 关闭连接
func (c *RedisClient) Close() error {
    return c.Client.Close()
}
```

---

### 2. `common/mq/` - 消息队列

#### `nats.go` - NATS 封装

```go
package mq

import (
    "context"
    "fmt"
    "time"

    "github.com/nats-io/nats.go"
    "go.uber.org/zap"
)

// NATSConfig NATS 配置
type NATSConfig struct {
    URL             string
    MaxReconnects   int
    ReconnectWait   time.Duration
    ReconnectBufSize int
}

// NATSClient NATS 客户端
type NATSClient struct {
    Conn   *nats.Conn
    logger *zap.Logger
}

// NewNATS 创建 NATS 连接
func NewNATS(cfg NATSConfig, log *zap.Logger) (*NATSClient, error) {
    opts := []nats.Option{
        nats.MaxReconnects(cfg.MaxReconnects),
        nats.ReconnectWait(cfg.ReconnectWait),
        nats.ReconnectBufSize(cfg.ReconnectBufSize),
        nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
            if err != nil {
                log.Error("NATS disconnected", zap.Error(err))
            }
        }),
        nats.ReconnectHandler(func(nc *nats.Conn) {
            log.Info("NATS reconnected", zap.String("url", nc.ConnectedUrl()))
        }),
    }

    conn, err := nats.Connect(cfg.URL, opts...)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to NATS: %w", err)
    }

    log.Info("NATS connected", zap.String("url", cfg.URL))

    return &NATSClient{
        Conn:   conn,
        logger: log,
    }, nil
}

// Publish 发布消息
func (c *NATSClient) Publish(subject string, data []byte) error {
    return c.Conn.Publish(subject, data)
}

// Subscribe 订阅消息
func (c *NATSClient) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
    return c.Conn.Subscribe(subject, handler)
}

// Health 检查 NATS 健康
func (c *NATSClient) Health(ctx context.Context) error {
    if !c.Conn.IsConnected() {
        return fmt.Errorf("NATS not connected")
    }
    return nil
}

// Close 关闭连接
func (c *NATSClient) Close() error {
    c.Conn.Close()
    return nil
}
```

---

### 3. `common/config/` - 配置加载

#### `loader.go` - 配置加载器

```go
package config

import (
    "fmt"
    "strings"

    "github.com/spf13/viper"
)

// LoadConfig 加载配置文件
func LoadConfig(configPath string, cfg interface{}) error {
    v := viper.New()

    // 设置配置文件路径
    v.SetConfigFile(configPath)

    // 环境变量支持
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    // 读取配置文件
    if err := v.ReadInConfig(); err != nil {
        return fmt.Errorf("failed to read config file: %w", err)
    }

    // 解析到结构体
    if err := v.Unmarshal(cfg); err != nil {
        return fmt.Errorf("failed to unmarshal config: %w", err)
    }

    return nil
}

// LoadConfigWithDefaults 加载配置 (支持默认值)
func LoadConfigWithDefaults(configPath string, cfg interface{}, defaults map[string]interface{}) error {
    v := viper.New()

    // 设置默认值
    for key, value := range defaults {
        v.SetDefault(key, value)
    }

    v.SetConfigFile(configPath)
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    if err := v.ReadInConfig(); err != nil {
        return fmt.Errorf("failed to read config file: %w", err)
    }

    if err := v.Unmarshal(cfg); err != nil {
        return fmt.Errorf("failed to unmarshal config: %w", err)
    }

    return nil
}
```

---

### 4. `common/server/` - HTTP 服务器

#### `gin.go` - Gin 服务器封装

```go
package server

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    "github.com/kart-io/k8s-agent/common/middleware"
)

// GinServerConfig Gin 服务器配置
type GinServerConfig struct {
    Host         string
    Port         int
    Mode         string // debug, release, test
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    IdleTimeout  time.Duration
}

// GinServer Gin 服务器
type GinServer struct {
    Engine *gin.Engine
    Server *http.Server
    logger *zap.Logger
}

// NewGinServer 创建 Gin 服务器
func NewGinServer(cfg GinServerConfig, log *zap.Logger) *GinServer {
    // 设置 Gin 模式
    gin.SetMode(cfg.Mode)

    engine := gin.New()

    // 默认中间件
    engine.Use(middleware.Recovery())
    engine.Use(middleware.RequestID())
    engine.Use(middleware.RequestLogger())
    engine.Use(middleware.CORS())

    server := &http.Server{
        Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        Handler:      engine,
        ReadTimeout:  cfg.ReadTimeout,
        WriteTimeout: cfg.WriteTimeout,
        IdleTimeout:  cfg.IdleTimeout,
    }

    return &GinServer{
        Engine: engine,
        Server: server,
        logger: log,
    }
}

// Run 启动服务器
func (s *GinServer) Run() error {
    s.logger.Info("Starting HTTP server", zap.String("addr", s.Server.Addr))
    if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        return fmt.Errorf("failed to start server: %w", err)
    }
    return nil
}

// Shutdown 优雅关闭
func (s *GinServer) Shutdown(ctx context.Context) error {
    s.logger.Info("Shutting down HTTP server")
    return s.Server.Shutdown(ctx)
}
```

#### `health.go` - 健康检查

```go
package server

import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
)

// HealthChecker 健康检查接口
type HealthChecker interface {
    Health(ctx context.Context) error
}

// RegisterHealthEndpoints 注册健康检查端点
func RegisterHealthEndpoints(engine *gin.Engine, checkers map[string]HealthChecker) {
    engine.GET("/health", healthHandler(checkers))
    engine.GET("/readiness", readinessHandler(checkers))
}

// healthHandler 健康检查处理器
func healthHandler(checkers map[string]HealthChecker) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "ok",
        })
    }
}

// readinessHandler 就绪检查处理器
func readinessHandler(checkers map[string]HealthChecker) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()
        results := make(map[string]string)
        allHealthy := true

        for name, checker := range checkers {
            if err := checker.Health(ctx); err != nil {
                results[name] = err.Error()
                allHealthy = false
            } else {
                results[name] = "ok"
            }
        }

        if allHealthy {
            c.JSON(http.StatusOK, gin.H{
                "status": "ready",
                "checks": results,
            })
        } else {
            c.JSON(http.StatusServiceUnavailable, gin.H{
                "status": "not ready",
                "checks": results,
            })
        }
    }
}
```

#### `shutdown.go` - 优雅关闭

```go
package server

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"

    "go.uber.org/zap"
)

// ShutdownHandler 关闭处理器
type ShutdownHandler interface {
    Shutdown(ctx context.Context) error
}

// GracefulShutdown 优雅关闭 (监听信号)
func GracefulShutdown(logger *zap.Logger, timeout time.Duration, handlers ...ShutdownHandler) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    <-quit
    logger.Info("Shutting down gracefully...")

    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    for _, handler := range handlers {
        if err := handler.Shutdown(ctx); err != nil {
            logger.Error("Shutdown error", zap.Error(err))
        }
    }

    logger.Info("Server exited")
}
```

---

### 5. `common/telemetry/` - 可观测性

#### `metrics.go` - Prometheus 指标

```go
package telemetry

import (
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    // HTTPRequestsTotal HTTP 请求总数
    HTTPRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    // HTTPRequestDuration HTTP 请求耗时
    HTTPRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
)

// RegisterMetrics 注册 Prometheus 指标
func RegisterMetrics() {
    prometheus.MustRegister(HTTPRequestsTotal)
    prometheus.MustRegister(HTTPRequestDuration)
}

// PrometheusHandler 返回 Prometheus metrics handler
func PrometheusHandler() gin.HandlerFunc {
    h := promhttp.Handler()
    return func(c *gin.Context) {
        h.ServeHTTP(c.Writer, c.Request)
    }
}
```

---

## 🔧 使用示例

### 完整的服务启动代码

```go
package main

import (
    "context"
    "flag"
    "time"

    "go.uber.org/zap"

    "github.com/kart-io/k8s-agent/common/config"
    "github.com/kart-io/k8s-agent/common/db"
    "github.com/kart-io/k8s-agent/common/logger"
    "github.com/kart-io/k8s-agent/common/mq"
    "github.com/kart-io/k8s-agent/common/server"
    "github.com/kart-io/k8s-agent/common/telemetry"
)

// Config 服务配置
type Config struct {
    Server server.GinServerConfig
    MySQL  db.MySQLConfig
    Redis  db.RedisConfig
    NATS   mq.NATSConfig
    Log    logger.Config
}

func main() {
    // 解析命令行参数
    configPath := flag.String("config", "configs/config.yaml", "config file path")
    flag.Parse()

    // 加载配置
    var cfg Config
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
    mysqlClient, err := db.NewMySQL(cfg.MySQL, log)
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

    // 注册 Prometheus 指标
    telemetry.RegisterMetrics()
    ginServer.Engine.GET("/metrics", telemetry.PrometheusHandler())

    // 注册业务路由
    RegisterRoutes(ginServer.Engine, mysqlClient, redisClient, natsClient)

    // 启动服务器 (goroutine)
    go func() {
        if err := ginServer.Run(); err != nil {
            log.Fatal("Server error", zap.Error(err))
        }
    }()

    // 优雅关闭
    server.GracefulShutdown(log, 30*time.Second, ginServer, mysqlClient, redisClient, natsClient)
}

func RegisterRoutes(engine *gin.Engine, mysql *db.MySQLClient, redis *db.RedisClient, nats *mq.NATSClient) {
    // 注册业务路由
    v1 := engine.Group("/api/v1")
    {
        v1.GET("/example", func(c *gin.Context) {
            // 业务逻辑
        })
    }
}
```

---

## 📊 重构收益

### 代码复用

| 功能 | 原代码行数 (每服务) | 新代码行数 (共享) | 节省 |
|------|---------------------|-------------------|------|
| MySQL 连接 | ~150 行 × 4 = 600 | 150 行 | 75% |
| Redis 连接 | ~120 行 × 4 = 480 | 120 行 | 75% |
| NATS 连接 | ~100 行 × 3 = 300 | 100 行 | 67% |
| 配置加载 | ~80 行 × 4 = 320 | 80 行 | 75% |
| HTTP 服务器 | ~200 行 × 4 = 800 | 200 行 | 75% |
| 健康检查 | ~100 行 × 4 = 400 | 100 行 | 75% |
| **总计** | **~2900 行** | **~750 行** | **74%** |

### 维护性提升

- ✅ **统一接口**: 所有服务使用相同的数据库/缓存接口
- ✅ **统一配置**: 配置结构和加载逻辑一致
- ✅ **统一监控**: Prometheus 指标命名和结构一致
- ✅ **bug 修复更快**: 修复一次，所有服务受益

### 新服务开发

- ✅ **快速启动**: 新服务只需 ~50 行代码即可启动
- ✅ **最佳实践内置**: 健康检查、优雅关闭、日志、监控自动配置
- ✅ **降低学习成本**: 开发者只需学习一套接口

---

## 🚀 实施计划

### Phase 1: 实现新包 (1-2 天)

1. ✅ 实现 `common/db/` (MySQL, Redis, Neo4j)
2. ✅ 实现 `common/mq/` (NATS)
3. ✅ 实现 `common/config/` (配置加载)
4. ✅ 实现 `common/server/` (HTTP 服务器、健康检查、优雅关闭)
5. ✅ 实现 `common/telemetry/` (Prometheus 指标)

### Phase 2: 重构服务 (2-3 天)

1. ✅ 重构 `agent-manager` 使用新 common 包
2. ✅ 重构 `orchestrator-service`
3. ✅ 重构 `reasoning-service-go`
4. ✅ 重构 `auth-service`

### Phase 3: 测试与验证 (1 天)

1. ✅ 单元测试所有新包
2. ✅ 集成测试所有服务
3. ✅ 性能测试确保无回退

### Phase 4: 文档与示例 (1 天)

1. ✅ 更新 `common/README.md`
2. ✅ 添加使用示例
3. ✅ 添加迁移指南

---

## ✅ 验收标准

1. ✅ 所有新包有完整的单元测试
2. ✅ 所有服务成功使用新 common 包
3. ✅ 代码行数减少 > 70%
4. ✅ 服务启动时间无明显增加
5. ✅ 文档完善，包含使用示例
6. ✅ CI/CD 测试全部通过

---

## 📝 注意事项

1. **向后兼容**: 保留现有 common 包功能，只添加不删除
2. **渐进式迁移**: 一个服务一个服务地迁移，降低风险
3. **性能测试**: 确保新封装不引入性能开销
4. **错误处理**: 所有连接操作都要有超时和重试机制
5. **日志统一**: 所有操作都使用结构化日志记录
