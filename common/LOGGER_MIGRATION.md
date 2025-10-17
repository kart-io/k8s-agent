# Common Package - Logger 模块更新说明

## 更新时间

2025-10-17

## 更新内容

已将 common 包中的 logger 模块从直接使用 `go.uber.org/zap` 更新为集成 [kart-io/logger](https://github.com/kart-io/logger)。

## 主要变化

### 1. 依赖变更

**之前**:
```go
require go.uber.org/zap v1.26.0
```

**现在**:
```go
require github.com/kart-io/logger v0.0.0
replace github.com/kart-io/logger => ../../logger
```

### 2. 配置结构变化

**之前**:
```go
type Config struct {
    Level        string
    Format       string
    OutputPath   string  // 单个路径
    EnableCaller bool
}
```

**现在**:
```go
type Config struct {
    Engine        string   // 新增：引擎选择（zap/slog）
    Level         string
    Format        string
    OutputPaths   []string // 改为数组，支持多个输出
    Development   bool     // 新增：开发模式
    EnableCaller  bool
    OTLPEndpoint  string   // 新增：OTLP 集成
    InitialFields map[string]interface{} // 新增：初始字段
}
```

### 3. API 变化

**之前（Zap 风格）**:
```go
import "go.uber.org/zap"

logger.Info("message", zap.String("key", "value"))
logger.Error("error", zap.Error(err))
```

**现在（三种风格）**:
```go
// 1. 简单参数风格
logger.Info("message")
logger.Error("error message", err)

// 2. Printf 格式化风格
logger.Infof("user %s logged in", username)
logger.Errorf("failed: %v", err)

// 3. 结构化键值对风格（推荐）
logger.Infow("message", "key", "value")
logger.Errorw("error occurred", "error", err.Error())
```

### 4. 新增功能

#### 引擎切换
支持在 Zap（高性能）和 Slog（标准库）之间切换：

```go
config := &logger.Config{
    Engine: "zap",  // 或 "slog"
    // ...
}
```

#### 初始字段（InitialFields）
自动在每个日志条目中包含固定字段：

```go
config := &logger.Config{
    InitialFields: map[string]interface{}{
        "service": "cluster-service",
        "version": "v1.0.0",
        "region":  "us-east-1",
    },
}
logger.Init(config)

// 所有日志都会自动包含 service、version、region 字段
logger.Infow("processing request", "request_id", "req-123")
```

#### OTLP 集成
支持将日志导出到 OpenTelemetry 收集器：

```go
config := &logger.Config{
    OTLPEndpoint: "http://localhost:4317",
}
```

## 迁移指南

### 旧代码（使用 Zap）

```go
import (
    "github.com/kart-io/k8s-agent/common/logger"
    "go.uber.org/zap"
)

func oldStyle() {
    config := &logger.Config{
        Level:      "info",
        Format:     "json",
        OutputPath: "stdout",
    }
    logger.Init(config)

    logger.Info("cluster created",
        zap.String("cluster_id", clusterID),
        zap.Int("node_count", nodeCount),
    )
}
```

### 新代码（使用 kart-io/logger）

```go
import "github.com/kart-io/k8s-agent/common/logger"

func newStyle() {
    config := &logger.Config{
        Engine:      "zap",           // 新增
        Level:       "info",
        Format:      "json",
        OutputPaths: []string{"stdout"}, // 改为数组
        InitialFields: map[string]interface{}{ // 新增
            "service": "cluster-service",
        },
    }
    logger.Init(config)

    // 推荐使用结构化风格（更简洁）
    logger.Infow("cluster created",
        "cluster_id", clusterID,
        "node_count", nodeCount,
    )
}
```

## 兼容性

### 保持兼容的部分

- 函数名称保持不变：`Init()`, `Info()`, `Error()` 等
- 配置字段名称基本保持一致
- 返回值类型没有变化

### 不兼容的部分

1. **字段传递方式**: 从 `zap.Field` 改为简单的键值对
2. **返回类型**: `Get()` 返回 `core.Logger` 而不是 `*zap.Logger`
3. **配置结构**: 新增字段，部分字段类型改变

## 建议

### 对于新代码

直接使用新的结构化风格（`Infow`, `Errorw` 等）：

```go
logger.Infow("operation completed",
    "duration_ms", duration.Milliseconds(),
    "status", "success",
)
```

### 对于现有代码

逐步迁移：

1. 更新配置结构（添加新字段）
2. 将 `zap.String()` 等替换为直接的键值对
3. 移除 `import "go.uber.org/zap"`

## 优势

1. **统一接口**: 支持多个日志引擎，但使用相同的 API
2. **更简洁**: 不需要 `zap.String()` 等包装函数
3. **更灵活**: 支持三种调用风格，适应不同场景
4. **更强大**: 内置 OTLP 支持、初始字段等高级功能
5. **性能**: Zap 引擎保持零分配的高性能特性

## 示例对比

### 记录请求日志

**旧方式**:
```go
logger.Info("Request completed",
    zap.String("method", method),
    zap.String("path", path),
    zap.Int("status", status),
    zap.Duration("latency", latency),
)
```

**新方式**:
```go
logger.Infow("Request completed",
    "method", method,
    "path", path,
    "status", status,
    "latency_ms", latency.Milliseconds(),
)
```

### 记录错误

**旧方式**:
```go
logger.Error("Database query failed",
    zap.String("table", "users"),
    zap.Error(err),
)
```

**新方式**:
```go
logger.Errorw("Database query failed",
    "table", "users",
    "error", err.Error(),
)
```

## 参考资料

- [kart-io/logger README](https://github.com/kart-io/logger/blob/main/README.md)
- [kart-io/logger 完整文档](https://github.com/kart-io/logger)
- [common/README.md - Logger 章节](./README.md#4-logger---日志工具)
