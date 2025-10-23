# Common Config Package

k8s-agent 项目统一配置包，采用 Options 模式设计。

## 设计理念

- **独立文件**: 每个配置类型一个独立文件，易于维护
- **Options 模式**: 灵活的配置选项，支持函数式配置
- **验证机制**: 每个配置都有 `Validate()` 方法
- **默认值**: 提供合理的默认配置

## 目录结构

```
config/
├── options.go           # 接口定义
├── server_options.go    # HTTP 服务器配置
├── database_options.go  # 数据库配置
├── redis_options.go     # Redis 配置
├── nats_options.go      # NATS 消息队列配置
├── logging_options.go   # 日志配置
├── jwt_options.go       # JWT 认证配置
├── metrics_options.go   # 指标采集配置
├── cors_options.go      # CORS 跨域配置
├── loader.go           # 配置加载工具
└── README.md           # 本文档
```

## 配置类型

### ServerOptions - HTTP 服务器配置

```go
import "github.com/kart-io/k8s-agent/common/config"

// 创建默认配置
opts := config.NewServerOptions()

// 使用 Options 模式
opts := config.NewServerOptions()
config.WithHost("0.0.0.0")(opts)
config.WithPort(8080)(opts)
config.WithMode("release")(opts)

// 验证配置
if err := opts.Validate(); err != nil {
    log.Fatal(err)
}
```

### DatabaseOptions - 数据库配置

```go
// 创建默认配置
opts := config.NewDatabaseOptions()

// 使用 Options 模式
opts := config.NewDatabaseOptions()
config.WithDBHost("mysql.example.com")(opts)
config.WithDBPort(3306)(opts)
config.WithDBUser("app_user")(opts)
config.WithDBPassword("secure_password")(opts)
config.WithDBName("myapp")(opts)

// 获取 DSN
dsn := opts.DSN()

// 验证配置
if err := opts.Validate(); err != nil {
    log.Fatal(err)
}
```

### RedisOptions - Redis 配置

```go
// 创建默认配置
opts := config.NewRedisOptions()

// 使用 Options 模式
opts := config.NewRedisOptions()
config.WithRedisAddr("redis.example.com:6379")(opts)
config.WithRedisPassword("redis_password")(opts)
config.WithRedisDB(1)(opts)
config.WithRedisPoolSize(20)(opts)

// 验证配置
if err := opts.Validate(); err != nil {
    log.Fatal(err)
}
```

### NATSOptions - NATS 消息队列配置

```go
// 创建默认配置
opts := config.NewNATSOptions()

// 使用 Options 模式
opts := config.NewNATSOptions()
config.WithNATSURL("nats://nats.example.com:4222")(opts)
config.WithNATSClusterID("k8s-agent")(opts)
config.WithNATSEnableJetStream(true)(opts)

// 验证配置
if err := opts.Validate(); err != nil {
    log.Fatal(err)
}
```

### LoggingOptions - 日志配置

```go
// 创建默认配置
opts := config.NewLoggingOptions()

// 使用 Options 模式
opts := config.NewLoggingOptions()
config.WithLogLevel("info")(opts)
config.WithLogFormat("json")(opts)
config.WithLogEngine("zap")(opts)
config.WithLogDevelopment(false)(opts)

// 验证配置
if err := opts.Validate(); err != nil {
    log.Fatal(err)
}
```

### JWTOptions - JWT 认证配置

```go
// 创建默认配置
opts := config.NewJWTOptions()

// 使用 Options 模式
opts := config.NewJWTOptions()
config.WithJWTSecret("your-secret-key")(opts)
config.WithJWTExpiresHours(48)(opts)

// 验证配置
if err := opts.Validate(); err != nil {
    log.Fatal(err)
}
```

### MetricsOptions - 指标采集配置

```go
// 创建默认配置
opts := config.NewMetricsOptions()

// 使用 Options 模式
opts := config.NewMetricsOptions()
config.WithMetricsEnabled(true)(opts)
config.WithMetricsPort(9090)(opts)
config.WithMetricsPath("/metrics")(opts)

// 验证配置
if err := opts.Validate(); err != nil {
    log.Fatal(err)
}
```

### CORSOptions - CORS 跨域配置

```go
// 创建默认配置
opts := config.NewCORSOptions()

// 使用 Options 模式
opts := config.NewCORSOptions()
config.WithCORSEnabled(true)(opts)
config.WithCORSAllowOrigins([]string{"https://example.com"})(opts)
config.WithCORSAllowCredentials(true)(opts)

// 验证配置
if err := opts.Validate(); err != nil {
    log.Fatal(err)
}
```

## 完整示例

### 示例1: 服务配置组合

```go
package main

import (
    "log"
    "github.com/kart-io/k8s-agent/common/config"
)

type ServiceConfig struct {
    Server   *config.ServerOptions   `yaml:"server"`
    Database *config.DatabaseOptions `yaml:"database"`
    Redis    *config.RedisOptions    `yaml:"redis"`
    NATS     *config.NATSOptions     `yaml:"nats"`
    Logging  *config.LoggingOptions  `yaml:"logging"`
    JWT      *config.JWTOptions      `yaml:"jwt"`
    Metrics  *config.MetricsOptions  `yaml:"metrics"`
    CORS     *config.CORSOptions     `yaml:"cors"`
}

func main() {
    // 创建配置
    cfg := &ServiceConfig{
        Server:   config.NewServerOptions(),
        Database: config.NewDatabaseOptions(),
        Redis:    config.NewRedisOptions(),
        NATS:     config.NewNATSOptions(),
        Logging:  config.NewLoggingOptions(),
        JWT:      config.NewJWTOptions(),
        Metrics:  config.NewMetricsOptions(),
        CORS:     config.NewCORSOptions(),
    }

    // 自定义配置
    config.WithPort(8080)(cfg.Server)
    config.WithDBHost("mysql.example.com")(cfg.Database)
    config.WithDBName("myapp")(cfg.Database)
    config.WithRedisAddr("redis.example.com:6379")(cfg.Redis)
    config.WithNATSURL("nats://nats.example.com:4222")(cfg.NATS)
    config.WithLogLevel("info")(cfg.Logging)
    config.WithJWTSecret("my-secret")(cfg.JWT)

    // 验证所有配置
    if err := cfg.Server.Validate(); err != nil {
        log.Fatalf("server config error: %v", err)
    }
    if err := cfg.Database.Validate(); err != nil {
        log.Fatalf("database config error: %v", err)
    }
    if err := cfg.Redis.Validate(); err != nil {
        log.Fatalf("redis config error: %v", err)
    }
    // ... 验证其他配置

    // 使用配置
    log.Printf("Server will listen on %s:%d", cfg.Server.Host, cfg.Server.Port)
    log.Printf("Database DSN: %s", cfg.Database.DSN())
}
```

### 示例2: 从 YAML 加载

```yaml
# config.yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"
  read_timeout: 10s
  write_timeout: 10s

database:
  host: "localhost"
  port: 3306
  user: "root"
  password: "password"
  database: "myapp"

redis:
  addr: "localhost:6379"
  password: ""
  db: 0

logging:
  level: "info"
  format: "json"
  engine: "zap"
```

```go
package main

import (
    "github.com/kart-io/k8s-agent/common/config"
    "github.com/spf13/viper"
)

type AppConfig struct {
    Server   config.ServerOptions   `mapstructure:"server"`
    Database config.DatabaseOptions `mapstructure:"database"`
    Redis    config.RedisOptions    `mapstructure:"redis"`
    Logging  config.LoggingOptions  `mapstructure:"logging"`
}

func main() {
    viper.SetConfigFile("config.yaml")
    viper.SetConfigType("yaml")

    if err := viper.ReadInConfig(); err != nil {
        panic(err)
    }

    var cfg AppConfig
    if err := viper.Unmarshal(&cfg); err != nil {
        panic(err)
    }

    // 验证配置
    if err := cfg.Server.Validate(); err != nil {
        panic(err)
    }
}
```

### 示例3: 环境变量覆盖

```go
package main

import (
    "os"
    "github.com/kart-io/k8s-agent/common/config"
    "github.com/spf13/viper"
)

func main() {
    // 设置环境变量前缀
    viper.SetEnvPrefix("APP")
    viper.AutomaticEnv()

    // 绑定环境变量
    viper.BindEnv("server.port")      // APP_SERVER_PORT
    viper.BindEnv("database.host")    // APP_DATABASE_HOST
    viper.BindEnv("redis.addr")       // APP_REDIS_ADDR

    // 设置默认值
    viper.SetDefault("server.port", 8080)
    viper.SetDefault("database.host", "localhost")
    viper.SetDefault("redis.addr", "localhost:6379")

    // 从文件加载
    viper.SetConfigFile("config.yaml")
    viper.ReadInConfig()

    var cfg AppConfig
    viper.Unmarshal(&cfg)

    // 环境变量会自动覆盖配置文件的值
}
```

## Options 函数列表

### ServerOptions (7个)
- `WithHost(string)` - 设置服务器地址
- `WithPort(int)` - 设置服务器端口
- `WithMode(string)` - 设置运行模式
- `WithReadTimeout(time.Duration)` - 设置读超时
- `WithWriteTimeout(time.Duration)` - 设置写超时
- `WithIdleTimeout(time.Duration)` - 设置空闲超时
- `WithGracefulStop(time.Duration)` - 设置优雅停机时间

### DatabaseOptions (10个)
- `WithDBHost(string)` - 设置数据库地址
- `WithDBPort(int)` - 设置数据库端口
- `WithDBUser(string)` - 设置用户名
- `WithDBPassword(string)` - 设置密码
- `WithDBName(string)` - 设置数据库名
- `WithDBSSLMode(string)` - 设置 SSL 模式
- `WithDBMaxOpenConns(int)` - 设置最大打开连接数
- `WithDBMaxIdleConns(int)` - 设置最大空闲连接数
- `WithDBConnMaxLifetime(time.Duration)` - 设置连接最大生命周期
- `WithDBAutoMigrate(bool)` - 设置自动迁移

### RedisOptions (8个)
- `WithRedisAddr(string)` - 设置 Redis 地址
- `WithRedisPassword(string)` - 设置密码
- `WithRedisDB(int)` - 设置数据库索引
- `WithRedisPoolSize(int)` - 设置连接池大小
- `WithRedisMinIdleConns(int)` - 设置最小空闲连接数
- `WithRedisDialTimeout(time.Duration)` - 设置连接超时
- `WithRedisReadTimeout(time.Duration)` - 设置读超时
- `WithRedisWriteTimeout(time.Duration)` - 设置写超时

### NATSOptions (8个)
- `WithNATSURL(string)` - 设置 NATS 地址
- `WithNATSClusterID(string)` - 设置集群 ID
- `WithNATSMaxReconnect(int)` - 设置最大重连次数
- `WithNATSReconnectWait(time.Duration)` - 设置重连等待时间
- `WithNATSPingInterval(time.Duration)` - 设置 Ping 间隔
- `WithNATSMaxPingsOut(int)` - 设置最大未响应 Ping 数
- `WithNATSEnableJetStream(bool)` - 启用 JetStream
- `WithNATSReconnectBufSize(int)` - 设置重连缓冲区大小

### LoggingOptions (8个)
- `WithLogLevel(string)` - 设置日志级别
- `WithLogFormat(string)` - 设置日志格式
- `WithLogOutput([]string)` - 设置输出路径
- `WithLogEngine(string)` - 设置日志引擎
- `WithLogDevelopment(bool)` - 设置开发模式
- `WithLogEnableCaller(bool)` - 启用调用者信息
- `WithLogOTLPEndpoint(string)` - 设置 OTLP 端点
- `WithLogOTLPInsecure(bool)` - 设置 OTLP 不安全连接

### JWTOptions (2个)
- `WithJWTSecret(string)` - 设置 JWT 密钥
- `WithJWTExpiresHours(int)` - 设置过期时间

### MetricsOptions (5个)
- `WithMetricsEnabled(bool)` - 启用指标
- `WithMetricsPath(string)` - 设置指标路径
- `WithMetricsPort(int)` - 设置指标端口
- `WithMetricsRetentionDays(int)` - 设置保留天数
- `WithMetricsAggregationInterval(time.Duration)` - 设置聚合间隔

### CORSOptions (7个)
- `WithCORSEnabled(bool)` - 启用 CORS
- `WithCORSAllowOrigins([]string)` - 设置允许的来源
- `WithCORSAllowMethods([]string)` - 设置允许的方法
- `WithCORSAllowHeaders([]string)` - 设置允许的头
- `WithCORSExposeHeaders([]string)` - 设置暴露的头
- `WithCORSAllowCredentials(bool)` - 允许凭证
- `WithCORSMaxAge(int)` - 设置缓存时间

**总计**: 55 个配置函数

## 默认值

每个配置类型都提供了合理的默认值：

| 配置类型 | 主要默认值 |
|---------|-----------|
| ServerOptions | port: 8080, mode: release |
| DatabaseOptions | host: localhost, port: 3306 |
| RedisOptions | addr: localhost:6379, db: 0 |
| NATSOptions | url: nats://localhost:4222 |
| LoggingOptions | level: info, format: json |
| JWTOptions | expires_hours: 24 |
| MetricsOptions | port: 9090, path: /metrics |
| CORSOptions | allow_origins: ["*"] |

## 最佳实践

1. **验证配置**: 始终调用 `Validate()` 方法验证配置
2. **使用默认值**: 从 `NewXxxOptions()` 开始，只修改需要的值
3. **环境变量**: 生产环境使用环境变量覆盖敏感信息
4. **配置分离**: 不同环境使用不同的配置文件

## 迁移指南

从旧配置迁移到新的 Options 模式：

```go
// 旧代码
type Config struct {
    Server ServerConfig `yaml:"server"`
}

// 新代码
type Config struct {
    Server config.ServerOptions `yaml:"server"`
}
```

## 参考

- 设计参考: [onexstack/options](https://github.com/onexstack/onexstack/tree/master/pkg/options)
- Options 模式: Functional Options Pattern in Go