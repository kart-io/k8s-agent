# Common Package 重构迁移指南

## 概述

本次重构的主要目标是将 `common/` 包恢复为纯粹的通用工具库，将包含业务逻辑的部分移至 `pkg/` 目录。

## 重构变化

### 1. 包迁移

以下包已从 `common/` 移至 `pkg/`：

| 原路径 | 新路径 | 说明 |
|--------|--------|------|
| `common/app/` | `pkg/app/` | 应用启动框架（包含业务逻辑） |
| `common/bootstrap/` | `pkg/bootstrap/` | 生命周期管理（项目特定） |
| `common/contextx/` | `pkg/contextx/` | 项目特定的上下文管理 |
| `common/idempotent/` | `pkg/idempotent/` | 幂等性处理（业务逻辑） |
| `common/initializers/` | `pkg/initializers/` | 组件初始化器（业务特定） |

### 2. 配置重构

原有的庞大 `options/` 包已被拆分为更小的模块化配置包：

```
common/config/
├── config.go          # 基础配置接口
├── server/            # 服务器配置
│   └── options.go
├── database/          # 数据库配置
│   └── options.go
├── cache/            # 缓存配置
│   └── options.go
├── middleware/       # 中间件配置
│   └── options.go
└���─ mq/              # 消息队列配置
    └── options.go
```

### 3. 日志统一

- 删除了旧的 `common/logger/` 包
- 统一使用 `github.com/kart-io/logger` 作为日志解决方案

### 4. Cache 包整合

创建了新的 cache 工厂模式，提供统一的缓存创建接口：

```go
import "github.com/kart-io/k8s-agent/common/cache"

// 使用工厂创建缓存
factory := cache.NewFactory(&cacheconfig.Options{
    Type: cacheconfig.TypeRedis,
    RedisAddr: "localhost:6379",
})

cache, err := factory.Create(ctx)
```

## 代码更新指南

### 更新 Import 路径

在你的代码中更新以下 import 路径：

```go
// 旧的 import
import (
    "github.com/kart-io/k8s-agent/common/app"
    "github.com/kart-io/k8s-agent/common/bootstrap"
    "github.com/kart-io/k8s-agent/common/contextx"
    "github.com/kart-io/k8s-agent/common/idempotent"
    "github.com/kart-io/k8s-agent/common/initializers"
)

// 新的 import
import (
    "github.com/kart-io/k8s-agent/pkg/app"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
    "github.com/kart-io/k8s-agent/pkg/contextx"
    "github.com/kart-io/k8s-agent/pkg/idempotent"
    "github.com/kart-io/k8s-agent/pkg/initializers"
)
```

### 使用新的配置包

```go
// 旧的方式
import "github.com/kart-io/k8s-agent/common/options"

opts := options.DefaultServerOptions()

// 新的方式
import "github.com/kart-io/k8s-agent/common/config/server"

opts := server.DefaultOptions()
```

### 日志迁移

```go
// 旧的方式
import "github.com/kart-io/k8s-agent/common/logger"

logger.Init(config)
logger.Info("message")

// 新的方式
import "github.com/kart-io/logger"

log := logger.New(logger.Config{
    Engine: logger.EngineZap,
    Level:  logger.LevelInfo,
})
log.Info("message")
```

## 编译和测试

完成代码更新后，请执行以下步骤验证更改：

```bash
# 1. 更新依赖
go mod tidy

# 2. 测试编译
make build

# 3. 运行单元测试
make test

# 4. 运行集成测试
make test-integration
```

## 回滚方案

如果需要回滚本次重构：

1. 恢复备份的 logger 包：
```bash
mv common/logger.backup common/logger
```

2. 删除 pkg 目录下的新包：
```bash
rm -rf pkg/{app,bootstrap,contextx,idempotent,initializers}
```

3. 恢复原有的 import 路径（运行逆向脚本）

## 未来计划

### Phase 1（已完成）
- ✅ 创建 pkg 目录结构
- ✅ 移动业务逻辑包
- ✅ 重构 options 包
- ✅ 统一日志方案
- ✅ 整合 cache 实现

### Phase 2（计划中）
- [ ] 创建独立的 `github.com/kart-io/go-commons` 库
- [ ] 将真正通用的功能提取到独立库
- [ ] 添加更多单元测试
- [ ] 完善文档和示例

### Phase 3（未来）
- [ ] 性能优化
- [ ] 添加更多通用工具
- [ ] 发布第一个稳定版本

## 联系支持

如有问题，请联系架构团队或在项目 issue 中报告。