# 代码重组方案：common/ vs pkg/

## 设计原则

### common/ - 通用基础包 (可独立启动项目)

**定位**：通用、可复用的基础功能库，**任何** Go 项目都可以使用

**特点**：
- 独立的 Go module (`github.com/kart-io/k8s-agent/common`)
- 零业务逻辑，纯技术实现
- 可以被其他项目直接引用
- 功能完整、文档齐全
- 类似于第三方库的定位

**适合放在 common/ 的代码**：
- ✅ HTTP/gRPC 服务器封装
- ✅ 数据库客户端封装 (MySQL, Redis)
- ✅ 消息队列客户端 (NATS, Kafka)
- ✅ 统一的配置管理 (Options 模式)
- ✅ 通用错误处理
- ✅ 日志封装
- ✅ 统一响应格式
- ✅ 分页工具
- ✅ 通用中间件 (CORS, 认证, 限流)
- ✅ 通用验证器
- ✅ 通用工具函数

### pkg/ - 项目专属包 (仅适用于当前项目)

**定位**：针对 **Aetherius (k8s-agent)** 项目的业务逻辑和领域模型

**特点**：
- 包含业务逻辑
- 与项目领域紧密相关
- 不适合被其他项目引用
- 可以引用 `common/` 包
- 位于根目录的 `pkg/` (不是 `internal/pkg/`)

**适合放在 pkg/ 的代码**：
- ✅ Kubernetes 相关的业务逻辑
- ✅ Agent 管理的领域模型
- ✅ 工作流编排的业务规则
- ✅ AI 推理的业务封装
- ✅ 幂等性处理 (业务相关)
- ✅ 项目特定的上下文管理
- ✅ 项目特定的指标定义
- ✅ 引导程序 (bootstrap)

---

## 当前代码分布分析

### common/ 现有内容 (18 个包)

| 包名 | 当前位置 | 建议 | 理由 |
|------|----------|------|------|
| `app/` | common/ | ❌ 移除或移到 pkg/ | 应用启动逻辑，项目特定 |
| `cache/` | common/ | ✅ 保留 | 通用缓存接口 (内存/Redis) |
| `client/` | common/ | ✅ 保留 | 通用客户端封装 |
| `config/` | common/ | ✅ 保留 | Options 模式配置，纯技术实现 |
| `db/` | common/ | ✅ 保留 | 数据库客户端封装 |
| `errors/` | common/ | ✅ 保留 | 通用错误处理 |
| `k8sutils/` | common/ | ⚠️ 部分移动 | 通用 K8s 转换保留，业务逻辑移到 pkg/ |
| `logger/` | common/ | ✅ 保留 | 日志封装 (但应使用 kart-io/logger) |
| `middleware/` | common/ | ✅ 保留 | 通用中间件 |
| `mq/` | common/ | ✅ 保留 | 消息队列客户端 |
| `pagination/` | common/ | ✅ 保留 | 分页工具 |
| `response/` | common/ | ✅ 保留 | 统一响应格式 |
| `server/` | common/ | ✅ 保留 | HTTP/gRPC 服务器封装 |
| `telemetry/` | common/ | ⚠️ 评估 | 如果是通用遥测保留，项目特定移走 |
| `types/` | common/ | ⚠️ 评估 | 通用类型保留，业务类型移到 pkg/ |
| `utils/` | common/ | ⚠️ 评估 | 通用工具保留，业务工具移到 pkg/ |
| `validator/` | common/ | ✅ 保留 | 通用数据验证 |

### internal/pkg/ 现有内容 (4 个包)

| 包名 | 当前位置 | 建议 | 理由 |
|------|----------|------|------|
| `bootstrap/` | internal/pkg/ | ➡️ 移到 pkg/ | 项目启动逻辑，应该在根 pkg/ |
| `contextx/` | internal/pkg/ | ➡️ 移到 pkg/ | 项目特定上下文增强 |
| `idempotent/` | internal/pkg/ | ➡️ 移到 pkg/ | 业务幂等性逻辑 |
| `metrics/` | internal/pkg/ | ➡️ 移到 pkg/ | 项目特定指标 |

---

## 重组方案

### 第一步：创建根级 pkg/ 目录

```bash
mkdir -p /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/pkg
```

### 第二步：移动 internal/pkg/ → pkg/

```bash
# 移动内容到根级 pkg/
mv internal/pkg/bootstrap     pkg/
mv internal/pkg/contextx      pkg/
mv internal/pkg/idempotent    pkg/
mv internal/pkg/metrics       pkg/

# 删除空的 internal/pkg/
rmdir internal/pkg/
```

### 第三步：从 common/ 提取业务相关代码到 pkg/

#### 3.1 创建 pkg/ 业务包结构

```bash
mkdir -p pkg/k8s              # K8s 业务逻辑
mkdir -p pkg/agent            # Agent 相关领域模型
mkdir -p pkg/workflow         # 工作流业务逻辑
mkdir -p pkg/diagnosis        # 诊断相关业务
```

#### 3.2 评估并移动 common/ 中的业务代码

需要详细审查以下包，将业务逻辑部分移到 pkg/:

**common/app/**
- 应用启动、初始化逻辑 → 移到 `pkg/app/`
- 或者集成到 `pkg/bootstrap/`

**common/k8sutils/**
- 纯 K8s 资源转换 → 保留在 `common/k8sutils/`
- 业务相关的 K8s 逻辑 → 移到 `pkg/k8s/`

**common/types/**
- HTTP 请求/响应通用类型 → 保留
- Agent、Workflow 等业务类型 → 移到 `pkg/types/`

**common/utils/**
- 字符串、时间等通用工具 → 保留
- 业务相关工具 → 移到 `pkg/utils/`

**common/telemetry/**
- 通用 OpenTelemetry 封装 → 保留
- 项目特定遥测逻辑 → 移到 `pkg/telemetry/`

### 第四步：清理 common/logger/

当前 `common/logger/` 应该被废弃，统一使用 `github.com/kart-io/logger`：

```bash
# 1. 检查谁还在使用 common/logger/
grep -r "github.com/kart-io/k8s-agent/common/logger" --include="*.go"

# 2. 替换为 github.com/kart-io/logger

# 3. 删除 common/logger/
rm -rf common/logger/
```

---

## 重组后的目录结构

```
k8s-agent/
├── common/                          # 通用基础包 (独立 module)
│   ├── cache/                       # ✅ 通用缓存
│   ├── client/                      # ✅ 通用客户端
│   ├── config/                      # ✅ Options 配置
│   ├── db/                          # ✅ 数据库封装
│   ├── errors/                      # ✅ 错误处理
│   ├── k8sutils/                    # ✅ K8s 资源转换 (仅通用部分)
│   ├── middleware/                  # ✅ 通用中间件
│   ├── mq/                          # ✅ 消息队列
│   ├── pagination/                  # ✅ 分页
│   ├── response/                    # ✅ 响应格式
│   ├── server/                      # ✅ 服务器封装
│   ├── telemetry/                   # ✅ 遥测封装 (仅通用部分)
│   ├── types/                       # ✅ 通用类型
│   ├── utils/                       # ✅ 通用工具
│   ├── validator/                   # ✅ 验证器
│   └── go.mod                       # 独立模块
│
├── pkg/                             # 项目专属包 (业务逻辑)
│   ├── bootstrap/                   # 项目启动引导
│   ├── contextx/                    # 项目上下文增强
│   ├── idempotent/                  # 业务幂等性
│   ├── metrics/                     # 项目指标
│   ├── k8s/                         # K8s 业务逻辑
│   ├── agent/                       # Agent 领域模型
│   ├── workflow/                    # 工作流业务
│   ├── diagnosis/                   # 诊断业务
│   ├── types/                       # 业务类型定义
│   ├── telemetry/                   # 项目遥测逻辑
│   └── utils/                       # 业务工具
│
├── internal/                        # 服务实现 (私有)
│   ├── agent-manager/
│   ├── orchestrator/
│   ├── reasoning/
│   └── ...
│
├── cmd/                             # 服务入口
│   ├── agent-manager/
│   ├── orchestrator/
│   └── ...
│
└── api/proto/                       # Protocol Buffers
```

---

## 判断标准

### 放在 common/ 的标准

问自己这些问题：

1. ✅ **这个代码可以被任何 Go 项目使用吗？**
   - 例如：Redis 客户端封装、日志工具、分页工具

2. ✅ **这个代码完全不包含业务逻辑吗？**
   - 纯技术实现，零业务规则

3. ✅ **这个代码可以作为独立库发布吗？**
   - 可以开源给社区使用

4. ✅ **如果我换一个项目，这个代码还有用吗？**
   - 例如：从 k8s-agent 换到电商项目，还能用吗？

**如果 4 个问题都是 YES → 放在 common/**

### 放在 pkg/ 的标准

1. ✅ **这个代码包含业务逻辑吗？**
   - 例如：工作流编排规则、Agent 心跳逻辑

2. ✅ **这个代码是 Aetherius 项目特有的吗？**
   - 例如：诊断策略、K8s 故障分析

3. ✅ **这个代码依赖项目的领域模型吗？**
   - 例如：Agent、Workflow、Strategy 等概念

4. ✅ **其他项目不太可能需要这个代码吗？**
   - 例如：K8s 故障诊断逻辑

**如果任何一个问题是 YES → 放在 pkg/**

---

## 实施步骤

### Phase 1: 准备工作 (1 天)

- [ ] 创建 `pkg/` 目录结构
- [ ] 审查 `common/` 和 `internal/pkg/` 所有代码
- [ ] 制定详细的迁移清单

### Phase 2: 代码迁移 (2-3 天)

- [ ] 移动 `internal/pkg/` → `pkg/`
- [ ] 从 `common/` 提取业务代码到 `pkg/`
- [ ] 更新所有 import 路径
- [ ] 运行测试确保没有破坏

### Phase 3: 清理优化 (1 天)

- [ ] 删除 `common/logger/`，统一使用 `kart-io/logger`
- [ ] 清理未使用的代码
- [ ] 更新文档

### Phase 4: 验证 (1 天)

- [ ] 运行所有测试
- [ ] 构建所有服务
- [ ] 更新 CLAUDE.md

---

## 迁移清单示例

### 从 common/ 移到 pkg/

```bash
# common/app/ → pkg/app/
mv common/app pkg/

# common/k8sutils/ 中的业务逻辑 → pkg/k8s/
# (需要手动拆分)

# common/types/ 中的业务类型 → pkg/types/
# (需要手动拆分)
```

### 更新 import 路径

**之前**：
```go
import "github.com/kart-io/k8s-agent/internal/pkg/bootstrap"
```

**之后**：
```go
import "github.com/kart-io/k8s-agent/pkg/bootstrap"
```

---

## 收益

### 对 common/ 的收益

1. **更清晰的定位**：纯技术库，无业务污染
2. **更容易复用**：可以被其他项目直接引用
3. **更容易维护**：代码职责单一
4. **可以独立开源**：作为独立库发布

### 对 pkg/ 的收益

1. **业务逻辑集中**：更容易理解项目
2. **更好的封装**：业务代码和通用代码分离
3. **更灵活**：可以随意修改，不影响 common/
4. **更符合 Go 惯例**：标准的 pkg/ 目录

### 对项目的收益

1. **代码结构更清晰**：common/ 通用, pkg/ 业务, internal/ 实现
2. **新人更容易上手**：职责明确
3. **更好的可测试性**：依赖关系清晰
4. **更容易重构**：边界明确

---

## 注意事项

### ⚠️ 不要过度移动

- 如果代码确实是通用的，不要强行移到 pkg/
- 宁可保守（留在 common/），也不要激进

### ⚠️ 保持向后兼容

- 迁移过程中保留旧的 import 路径（临时）
- 逐步弃用，给服务时间适配

### ⚠️ 测试充分

- 每次移动后立即运行测试
- 确保没有循环依赖
- 确保构建成功

### ⚠️ 文档同步更新

- 更新 README.md
- 更新 CLAUDE.md
- 更新代码注释中的 import 示例

---

## 参考资料

### Go 项目布局标准

- [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
  - `/pkg` - 外部应用程序可以使用的库代码
  - `/internal` - 私有应用程序和库代码

### 其他项目示例

- **Kubernetes**: `pkg/` 用于可复用组件，`internal/` 用于私有实现
- **Prometheus**: `pkg/` 用于工具库
- **Istio**: 清晰的 `pkg/` 和 `internal/` 划分

---

**总结**：
- `common/` = 任何项目都能用的技术库
- `pkg/` = 仅 Aetherius 项目使用的业务逻辑
- `internal/` = 服务私有实现，不对外暴露
