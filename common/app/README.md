# Common App Framework

通用应用程序启动框架，提供标准化的应用程序启动、配置管理和生命周期管理。

## 功能特性

- ✅ 统一的配置加载流程（配置文件 + 命令行参数 + 环境变量）
- ✅ 自动配置验证和默认值设置
- ✅ 标准化的应用程序生命周期管理
- ✅ 优雅关闭支持
- ✅ 信号处理（SIGINT, SIGTERM）
- ✅ 日志初始化集成
- ✅ 减少样板代码

## 快速开始

### 方式 1: 简单模式（直接使用 RunFunc）

适用于简单的应用程序，只需要一个运行函数。

```go
package main

import (
    "github.com/kart-io/k8s-agent/common/app"
    "github.com/kart-io/k8s-agent/common/config"
)

type MyOptions struct {
    Server   *config.ServerOptions   `json:"server" mapstructure:"server"`
    Database *config.DatabaseOptions `json:"database" mapstructure:"database"`
}

func (o *MyOptions) Complete() error {
    if err := o.Server.Complete(); err != nil {
        return err
    }
    return o.Database.Complete()
}

func (o *MyOptions) Validate() []error {
    var errs []error
    if err := o.Server.Validate(); err != nil {
        errs = append(errs, err)
    }
    if err := o.Database.Validate(); err != nil {
        errs = append(errs, err)
    }
    return errs
}

func (o *MyOptions) AddFlags(fs *pflag.FlagSet) {
    o.Server.AddFlags(fs)
    o.Database.AddFlags(fs)
}

func main() {
    opts := &MyOptions{
        Server:   config.NewServerOptions(),
        Database: config.NewDatabaseOptions(),
    }

    runFunc := func(opts app.Options) error {
        myOpts := opts.(*MyOptions)
        // 初始化和运行应用程序
        return runMyService(myOpts)
    }

    app.Run(opts, runFunc, app.CommandConfig{
        Use:         "my-service",
        Short:       "My Service",
        Long:        "My service does something awesome",
        EnvPrefix:   "MY_SERVICE",
    })
}
```

### 方式 2: 应用程序模式（实现 Application 接口）

适用于复杂的应用程序，需要完整的生命周期管理。

```go
package main

import (
    "context"
    "github.com/kart-io/k8s-agent/common/app"
    "go.uber.org/zap"
)

type MyApplication struct {
    logger *zap.Logger
    server *Server
}

func (a *MyApplication) Initialize(ctx context.Context, opts app.Options) error {
    myOpts := opts.(*MyOptions)
    // 初始化组件
    return nil
}

func (a *MyApplication) Run(ctx context.Context) error {
    // 运行应用程序
    return a.server.Run(ctx)
}

func (a *MyApplication) Shutdown(ctx context.Context) error {
    // 优雅关闭
    return a.server.Shutdown(ctx)
}

func initLogger(opts app.Options) (*zap.Logger, error) {
    myOpts := opts.(*MyOptions)
    // 初始化日志
    return zap.NewProduction()
}

func main() {
    opts := &MyOptions{
        Server:   config.NewServerOptions(),
        Database: config.NewDatabaseOptions(),
    }

    myApp := &MyApplication{}

    app.RunWithRunner(opts, myApp, initLogger, app.CommandConfig{
        Use:         "my-service",
        Short:       "My Service",
        Long:        "My service does something awesome",
        EnvPrefix:   "MY_SERVICE",
    })
}
```

## 配置优先级

配置值的优先级（从高到低）：

1. **命令行参数** - 最高优先级
2. **环境变量** - 需要设置 `EnvPrefix`
3. **配置文件** - 通过 `-c` 或 `--config` 指定
4. **默认值** - 在 `Complete()` 方法中设置

## Options 接口

所有应用程序配置必须实现 `Options` 接口：

```go
type Options interface {
    // Complete 完成配置初始化，设置默认值
    Complete() error

    // Validate 验证配置的有效性
    Validate() []error

    // AddFlags 添加命令行参数
    AddFlags(fs *pflag.FlagSet)
}
```

### Complete() 方法

负责：
- 设置默认值
- 计算派生值
- 修正无效值
- 确保配置一致性

```go
func (o *MyOptions) Complete() error {
    // 调用子配置的 Complete
    if err := o.Server.Complete(); err != nil {
        return err
    }

    // 设置派生值
    if o.CustomField == "" {
        o.CustomField = fmt.Sprintf("%s:%d", o.Server.Host, o.Server.Port)
    }

    return nil
}
```

### Validate() 方法

负责验证配置的有效性：

```go
func (o *MyOptions) Validate() []error {
    var errs []error

    // 验证子配置
    if err := o.Server.Validate(); err != nil {
        errs = append(errs, err)
    }

    // 自定义验证
    if o.CustomField == "" {
        errs = append(errs, fmt.Errorf("custom_field is required"))
    }

    return errs
}
```

### AddFlags() 方法

负责注册命令行参数：

```go
func (o *MyOptions) AddFlags(fs *pflag.FlagSet) {
    // 添加子配置的参数
    o.Server.AddFlags(fs)
    o.Database.AddFlags(fs)

    // 添加自定义参数
    fs.String("custom.field", o.CustomField, "Custom field description")
}
```

## Application 接口

复杂应用程序可以实现 `Application` 接口：

```go
type Application interface {
    // Initialize 初始化应用程序
    Initialize(ctx context.Context, opts Options) error

    // Run 运行应用程序
    Run(ctx context.Context) error

    // Shutdown 优雅关闭应用程序
    Shutdown(ctx context.Context) error
}
```

## CommandConfig 配置

```go
type CommandConfig struct {
    // Use 命令使用说明（必需）
    Use string

    // Short 命令简短描述（必需）
    Short string

    // Long 命令详细描述（可选）
    Long string

    // EnvPrefix 环境变量前缀（可选，如 "MY_SERVICE"）
    EnvPrefix string

    // ConfigFileFlag 配置文件参数名称（默认 "config"）
    ConfigFileFlag string

    // ConfigFileShort 配置文件参数短名称（默认 "c"）
    ConfigFileShort string
}
```

## 使用示例

### 启动应用程序

```bash
# 使用默认配置
./my-service

# 使用配置文件
./my-service -c config.yaml

# 使用命令行参数
./my-service --server.host=0.0.0.0 --server.port=8080

# 使用环境变量
export MY_SERVICE_SERVER_HOST=0.0.0.0
export MY_SERVICE_SERVER_PORT=8080
./my-service

# 混合使用（命令行优先级最高）
./my-service -c config.yaml --server.port=9090
```

### 查看帮助

```bash
./my-service --help
```

## 完整示例

参考 `agent-manager` 服务的实现：

```bash
agent-manager/
├── cmd/
│   ├── app/
│   │   ├── app.go        # 使用 common/app 框架
│   │   └── server.go     # 实现 Application 接口
│   └── server/
│       └── main.go       # 调用 app.Run()
├── internal/
│   └── config/
│       └── options.go    # 实现 Options 接口
```

## 优势

相比传统方式，使用此框架可以：

1. **减少样板代码** - 从 ~50 行减少到 ~10 行
2. **统一启动流程** - 所有服务使用相同的启动模式
3. **简化维护** - 配置管理逻辑集中在一处
4. **提高一致性** - 所有服务遵循相同的配置约定
5. **更好的测试性** - 配置逻辑独立，易于测试

## 迁移指南

### 从旧的启动方式迁移

**旧代码：**
```go
cmd := &cobra.Command{
    Use:   "my-service",
    Short: "My Service",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 50+ 行配置加载、验证、初始化代码
        if cfgFile := viper.GetString("config"); cfgFile != "" {
            viper.SetConfigFile(cfgFile)
            if err := viper.ReadInConfig(); err != nil {
                return err
            }
        }
        // ... 更多代码
    },
}
```

**新代码：**
```go
app.Run(opts, runFunc, app.CommandConfig{
    Use:       "my-service",
    Short:     "My Service",
    EnvPrefix: "MY_SERVICE",
})
```

节省了 90% 的样板代码！

## 最佳实践

1. **使用组合** - Options 结构组合 common/config 中的标准选项
2. **实现接口** - 确保所有配置实现 Options 接口
3. **调用子方法** - 在 Complete/Validate 中调用子配置的对应方法
4. **设置前缀** - 为环境变量设置合理的前缀
5. **添加注释** - 为配置字段添加清晰的注释

## 相关文档

- [Common Config](../config/README.md) - 标准配置选项
- [Options Pattern](../OPTIONS_PATTERN.md) - 配置选项模式
- [Agent Manager](../../agent-manager/README.md) - 使用示例
