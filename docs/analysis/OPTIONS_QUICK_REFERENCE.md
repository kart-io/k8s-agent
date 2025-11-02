# Options 包快速参考指南

## 当前 Options 文件清单

### 配置文件总览（17个主要选项）

```
server_options.go          HTTP服务器配置
database_options.go        MySQL数据库配置  
redis_options.go           Redis缓存配置
nats_options.go            NATS消息队列配置
logging_options.go         日志配置
jwt_options.go             JWT认证配置
cors_options.go            CORS跨域配置
grpc_options.go            gRPC服务器配置
metrics_options.go         Prometheus指标配置
agent_options.go           Agent配置
llm_options.go             LLM AI模型配置
email_options.go           邮件配置
memory_options.go          向量存储配置
learning_options.go        学习功能配置
analysis_options.go        分析功能配置
performance_options.go     性能配置
prediction_options.go      预测功能配置
```

### 支持文件

```
options.go                 接口定义
loader.go                  配置加载工具（5个函数）
helpers.go                 辅助函数
database_client.go         数据库客户端适配器
redis_client.go            Redis客户端适配器
validation/                验证函数集合
```

---

## Options 标准模式模板

当添加新选项时，使用这个模板：

```go
package options

import (
	"time"
	"github.com/spf13/pflag"
	"github.com/kart-io/k8s-agent/common/options/validation"
)

// YourOptions 功能描述
type YourOptions struct {
	Field1 string        `mapstructure:"field1" yaml:"field1" json:"field1"`
	Field2 int           `mapstructure:"field2" yaml:"field2" json:"field2"`
	Field3 time.Duration `mapstructure:"field3" yaml:"field3" json:"field3"`
}

// NewYourOptions 创建默认配置
func NewYourOptions() *YourOptions {
	return &YourOptions{
		Field1: "default",
		Field2: 100,
		Field3: 30 * time.Second,
	}
}

// Validate 验证配置
func (o *YourOptions) Validate() error {
	if err := validation.ValidateRequired(o.Field1, "your.field1"); err != nil {
		return err
	}
	if err := validation.ValidatePositiveInt(o.Field2, "your.field2"); err != nil {
		return err
	}
	return nil
}

// AddFlags 添加命令行参数
func (o *YourOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Field1, "your.field1", o.Field1, "Description")
	fs.IntVar(&o.Field2, "your.field2", o.Field2, "Description")
	fs.DurationVar(&o.Field3, "your.field3", o.Field3, "Description")
}

// Complete 完成配置初始化
func (o *YourOptions) Complete() error {
	if o.Field1 == "" {
		o.Field1 = "default"
	}
	if o.Field2 <= 0 {
		o.Field2 = 100
	}
	if o.Field3 <= 0 {
		o.Field3 = 30 * time.Second
	}
	return nil
}

// With* 函数族（4-5个）
func WithField1(value string) func(*YourOptions) {
	return func(o *YourOptions) {
		o.Field1 = value
	}
}

func WithField2(value int) func(*YourOptions) {
	return func(o *YourOptions) {
		o.Field2 = value
	}
}

func WithField3(value time.Duration) func(*YourOptions) {
	return func(o *YourOptions) {
		o.Field3 = value
	}
}
```

---

## 标签约定

所有字段必须有三层标签：

```go
type Example struct {
	MyField string `mapstructure:"my_field" yaml:"my_field" json:"my_field"`
}
```

- `mapstructure`: Viper解析配置文件时使用
- `yaml`: YAML序列化
- `json`: JSON序列化
- 使用下划线连接: `read_timeout`, `max_open_conns`

---

## 命名规范

| 模式 | 示例 | 说明 |
|------|------|------|
| 结构体 | `ServerOptions` | `{功能}Options` |
| 构造函数 | `NewServerOptions()` | `New{结构体}()` |
| With函数 | `WithPort(8080)` | `With{字段名}()` |
| 前缀With | `WithDBPort(3306)` | 复杂字段用前缀 |
| 标签 | `mapstructure:"port"` | 全小写，下划线连接 |

---

## 验证器函数速查表

```go
// 必填验证
validation.ValidateRequired(value, "field_name")

// 端口验证 (1-65535)
validation.ValidatePort(port, "service_name")

// 地址验证
validation.ValidateAddr(addr, "service_name")

// 正整数验证
validation.ValidatePositiveInt(value, "field_name")

// 非负整数验证
validation.ValidateNonNegativeInt(value, "field_name")

// 枚举验证
validation.ValidateEnum(value, "field_name", []string{"opt1", "opt2"})

// 连接池验证
validation.ValidateConnectionPool(maxOpen, maxIdle, "service_name")

// 超时验证
validation.ValidateTimeouts(dial, read, write, "service_name")

// Redis数据库验证
validation.ValidateRedisDB(db)
```

---

## 配置加载方式

### 方式1：标准加载（推荐）

```go
type AppConfig struct {
	Server options.ServerOptions `yaml:"server"`
	NATS   options.NATSOptions   `yaml:"nats"`
}

func (c *AppConfig) Complete() error {
	if err := c.Server.Complete(); err != nil {
		return err
	}
	if err := c.NATS.Complete(); err != nil {
		return err
	}
	return nil
}

func (c *AppConfig) Validate() []error {
	var errs []error
	if err := c.Server.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.NATS.Validate(); err != nil {
		errs = append(errs, err)
	}
	return errs
}

// 加载配置
cfg := &AppConfig{
	Server: options.NewServerOptions(),
	NATS:   options.NewNATSOptions(),
}

if err := options.LoadOptions(cfg, "config.yaml", nil); err != nil {
	log.Fatal(err)
}
```

### 方式2：带回调加载

用于需要特殊处理的场景（如环境变量覆盖）：

```go
if err := options.LoadOptionsWithCallback(
	cfg,
	"config.yaml",
	nil,
	func(v *viper.Viper, cfg options.LoadableConfig) error {
		// 自定义逻辑，如处理 LLM 环境变量
		return nil
	},
); err != nil {
	log.Fatal(err)
}
```

### 方式3：环境变量加载

```go
cfg := &AppConfig{}
if err := options.LoadConfigFromEnv("MYAPP", cfg); err != nil {
	log.Fatal(err)
}
// 环境变量: MYAPP_SERVER_PORT, MYAPP_NATS_URL 等
```

---

## 常见使用模式

### YAML配置文件

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"
  read_timeout: 10s
  write_timeout: 10s
  graceful_stop: 30s

database:
  host: "localhost"
  port: 3306
  user: "app"
  password: "${DB_PASSWORD}"
  database: "myapp"
  max_open_conns: 100
  max_idle_conns: 10

redis:
  addr: "localhost:6379"
  password: ""
  db: 0
  pool_size: 10

nats:
  url: "nats://localhost:4222"
  cluster_id: "k8s-agent"
  enable_jetstream: true

logging:
  level: "info"
  format: "json"
  engine: "zap"
```

### 环境变量覆盖

```bash
export SERVER_PORT=9090
export DATABASE_HOST=db.example.com
export REDIS_ADDR=redis.example.com:6379
export NATS_URL=nats://nats.example.com:4222
export LOG_LEVEL=debug
```

### 代码配置

```go
opts := options.NewServerOptions()
options.WithHost("127.0.0.1")(opts)
options.WithPort(9090)(opts)
options.WithMode("debug")(opts)

if err := opts.Validate(); err != nil {
	log.Fatal(err)
}
```

---

## 关键类型说明

### LoadableConfig 接口

```go
type LoadableConfig interface {
	Complete() error
	Validate() []error  // 返回错误切片
}
```

所有配置结构体在服务级应实现此接口。

### Options 接口

```go
type Options interface {
	Validate() error
}
```

个别选项实现此接口。

---

## 端口分配规范

系统内各服务默认端口（可通过配置修改）：

```
8080    HTTP 服务（Agent Manager）
8081    HTTP 服务（Orchestrator）
8082    HTTP 服务（Reasoning Service）
8083    HTTP 服务（Auth Service）
9090    gRPC 服务端口
9091    健康检查端口
6379    Redis
3306    MySQL
4222    NATS
7474    Neo4j HTTP
7687    Neo4j Bolt
```

---

## 常见问题

### Q: 如何添加新的配置选项？

A: 复制选项模板文件（如 `server_options.go`），修改结构体名称、字段、默认值和验证逻辑。

### Q: With函数命名有规则吗？

A: 是的。简短字段: `WithPort()`，复杂字段: `WithDBPort()`，多词段: `WithReadTimeout()`

### Q: 标签顺序重要吗？

A: 不重要，但建议按 `mapstructure`, `yaml`, `json` 顺序排列以保持一致性。

### Q: 验证失败会怎样？

A: LoadOptions 会返回错误，应用启动失败。这是intentional设计。

### Q: 可以有嵌套的Options吗？

A: 可以，但需要在 Complete() 中递归调用嵌套选项的 Complete() 方法。

---

## 相关文件位置

- 完整分析: `/docs/analysis/COMMON_OPTIONS_ANALYSIS.md`
- Options包: `/common/options/`
- 初始化器: `/common/initializers/`
- 健康检查: `/common/app/health.go`
- Bootstrap框架: `/pkg/bootstrap/`

---

## 最佳实践检查清单

添加新选项时：

- [ ] 创建 `xxx_options.go` 文件
- [ ] 实现 `NewXxxOptions()` 构造函数
- [ ] 实现 `Validate()` 方法
- [ ] 实现 `AddFlags()` 方法
- [ ] 实现 `Complete()` 方法
- [ ] 添加 4-5 个 `With*()` 函数
- [ ] 字段有三层标签 (`mapstructure`, `yaml`, `json`)
- [ ] 使用统一的标签命名（下划线）
- [ ] 添加合理的默认值
- [ ] 为命令行参数提供文档字符串
- [ ] 在服务的 AppConfig 结构体中引入

