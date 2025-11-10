# Common 目录重组为基础设施层实施计划

## 执行摘要

`common/` 目录当前存在严重的层次混乱问题，混合了基础设施、应用层和业务逻辑。本文档提出将 `common/` 重组为纯基础设施层的具体实施计划。

## 一、现状分析

### 1.1 问题总结

| 问题类型 | 具体表现 | 影响范围 |
|---------|---------|---------|
| **层次混乱** | 基础设施与业务逻辑混合 | 违反清洁架构原则 |
| **职责不清** | config/options/core 功能重复 | 代码冗余，维护困难 |
| **依赖倒置** | response/server 依赖具体框架 | 无法在其他项目复用 |
| **业务泄漏** | errors 包含业务错误码 | 违反基础设施层通用性 |

### 1.2 当前目录结构问题分析

```
common/ (当前 - 混乱状态)
├── cache/              ✅ 纯基础设施
├── config/             ⚠️ 与 options 重复
├── core/               ⚠️ 只有配置加载，归属不当
├── db/                 ✅ 纯基础设施
├── errors/             ❌ 包含业务错误码
├── health/             ❌ 应用层关注点
├── k8sutils/           ❌ 包含业务逻辑
├── loggerutil/         ✅ 纯基础设施
├── metrics/            ⚠️ 需要评估
├── middleware/         ❌ 应用层（JWT、CORS）
├── mq/                 ✅ 纯基础设施
├── options/            ⚠️ 40+ 配置文件，过度设计
├── pagination/         ❌ 应用层功能
├── response/           ❌ 依赖 Gin，应用层
├── serializers/        ✅ 纯基础设施
├── server/             ❌ 依赖具体框架
├── storage/            ✅ 纯基础设施（设计良好）
├── telemetry/          ✅ 纯基础设施
├── utils/              ⚠️ 需要拆分
└── validator/          ✅ 通用验证
```

## 二、重组方案

### 2.1 目标架构

```
common/ (重组后 - 纯基础设施层)
├── cache/              # 缓存抽象（Redis、Memory）
├── database/           # 数据库连接（MySQL、Redis）
├── messaging/          # 消息队列（NATS）
├── storage/            # 存储层抽象（Repository Pattern）
├── logging/            # 日志基础设施
├── telemetry/          # 遥测和监控
├── serialization/      # 序列化工具
└── foundation/         # 基础工具（仅技术性）

pkg/ (业务支撑层)
├── config/             # 配置管理（合并 config + options + core）
├── errors/             # 业务错误定义
├── middleware/         # HTTP 中间件
├── server/             # 服务器实现
├── response/           # API 响应格式
├── health/             # 健康检查
├── pagination/         # 分页逻辑
├── k8s/                # K8s 业务工具
└── utils/              # 业务工具函数
```

### 2.2 迁移映射表

| 原路径 | 新路径 | 操作 | 原因 |
|--------|--------|------|------|
| common/cache/ | common/cache/ | 保留 | 纯基础设施 |
| common/config/ | pkg/config/ | 移动+合并 | 包含业务配置 |
| common/core/ | pkg/config/loader/ | 合并 | 属于配置系统 |
| common/db/ | common/database/ | 重命名 | 更清晰的命名 |
| common/errors/ | pkg/errors/ | 移动 | 包含业务错误 |
| common/health/ | pkg/health/ | 移动 | 应用层关注点 |
| common/k8sutils/ | pkg/k8s/utils/ | 移动 | K8s 业务逻辑 |
| common/loggerutil/ | common/logging/ | 重命名 | 更标准的命名 |
| common/middleware/ | pkg/middleware/ | 移动 | 应用层中间件 |
| common/mq/ | common/messaging/ | 重命名 | 更清晰的命名 |
| common/options/ | pkg/config/options/ | 移动+合并 | 配置系统一部分 |
| common/pagination/ | pkg/pagination/ | 移动 | 应用层功能 |
| common/response/ | pkg/response/ | 移动 | 依赖 Gin 框架 |
| common/serializers/ | common/serialization/ | 重命名 | 更标准的命名 |
| common/server/ | pkg/server/ | 移动 | 依赖具体框架 |
| common/storage/ | common/storage/ | 保留 | 纯基础设施 |
| common/telemetry/ | common/telemetry/ | 保留 | 纯基础设施 |
| common/utils/ | 拆分 | 拆分 | 分为基础和业务 |
| common/validator/ | common/validation/ | 重命名 | 更标准的命名 |

## 三、实施步骤

### 第一阶段：准备工作（1天）

#### Step 1.1: 创建迁移分支
```bash
git checkout -b refactor/common-to-infra
```

#### Step 1.2: 分析依赖影响
```bash
# 生成依赖报告
go list -f '{{.ImportPath}} {{.Imports}}' ./... | grep "common/" > deps_report.txt

# 统计每个包的使用情况
for pkg in cache config core db errors health k8sutils loggerutil middleware mq options pagination response serializers server storage telemetry utils validator; do
  echo "$pkg: $(grep -r "common/$pkg" internal/ cmd/ pkg/ --include="*.go" | wc -l) references"
done
```

#### Step 1.3: 创建新目录结构
```bash
# 在 pkg/ 下创建新目录
mkdir -p pkg/config/options pkg/config/loader
mkdir -p pkg/errors pkg/middleware pkg/server/http pkg/server/grpc
mkdir -p pkg/response pkg/health pkg/pagination
mkdir -p pkg/k8s/utils pkg/utils
```

### 第二阶段：高优先级迁移（2天）

#### Step 2.1: 整合配置系统
```bash
# 1. 合并 core/loader.go 到 pkg/config/
mv common/core/loader.go pkg/config/loader/loader.go

# 2. 移动 options/ 到 pkg/config/
mv common/options/* pkg/config/options/

# 3. 合并 config/ 内容
# 需要手动处理冲突和重复
```

#### Step 2.2: 迁移错误系统
```bash
# 移动 errors 包
mv common/errors/* pkg/errors/

# 更新 imports
find . -name "*.go" -exec sed -i 's|"github.com/kart-io/k8s-agent/common/errors"|"github.com/kart-io/k8s-agent/pkg/errors"|g' {} \;
```

#### Step 2.3: 迁移响应和服务器包
```bash
# 移动 response
mv common/response/* pkg/response/

# 移动 server
mv common/server/* pkg/server/

# 更新 imports
find . -name "*.go" -exec sed -i 's|common/response|pkg/response|g' {} \;
find . -name "*.go" -exec sed -i 's|common/server|pkg/server|g' {} \;
```

### 第三阶段：中优先级迁移（2天）

#### Step 3.1: 迁移应用层包
```bash
# 移动 middleware
mv common/middleware/* pkg/middleware/

# 移动 health
mv common/health/* pkg/health/

# 移动 pagination
mv common/pagination/* pkg/pagination/

# 移动 k8sutils
mv common/k8sutils/* pkg/k8s/utils/
```

#### Step 3.2: 重命名基础设施包
```bash
# 重命名以提高清晰度
mv common/db common/database
mv common/mq common/messaging
mv common/loggerutil common/logging
mv common/serializers common/serialization
mv common/validator common/validation
```

#### Step 3.3: 拆分 utils
```bash
# 分析 utils 内容
# 基础工具留在 common/foundation/
# 业务工具移到 pkg/utils/
```

### 第四阶段：测试和验证（2天）

#### Step 4.1: 编译测试
```bash
# 确保所有服务能够编译
make build

# 运行单元测试
make test

# 运行集成测试
make test-integration
```

#### Step 4.2: Import 路径更新脚本
```bash
#!/bin/bash
# update_imports.sh

# 批量更新 import 路径
declare -A PATH_MAPPING=(
  ["common/config"]="pkg/config"
  ["common/core"]="pkg/config/loader"
  ["common/db"]="common/database"
  ["common/errors"]="pkg/errors"
  ["common/health"]="pkg/health"
  ["common/k8sutils"]="pkg/k8s/utils"
  ["common/loggerutil"]="common/logging"
  ["common/middleware"]="pkg/middleware"
  ["common/mq"]="common/messaging"
  ["common/options"]="pkg/config/options"
  ["common/pagination"]="pkg/pagination"
  ["common/response"]="pkg/response"
  ["common/serializers"]="common/serialization"
  ["common/server"]="pkg/server"
  ["common/validator"]="common/validation"
)

for old_path in "${!PATH_MAPPING[@]}"; do
  new_path="${PATH_MAPPING[$old_path]}"
  find . -name "*.go" -exec sed -i "s|github.com/kart-io/k8s-agent/$old_path|github.com/kart-io/k8s-agent/$new_path|g" {} \;
done
```

### 第五阶段：文档更新（1天）

#### Step 5.1: 更新 CLAUDE.md
- 更新 common/ 和 pkg/ 的描述
- 说明新的代码组织原则

#### Step 5.2: 更新 README
- 更新项目结构说明
- 更新开发指南

#### Step 5.3: 创建迁移指南
- 为其他开发者提供迁移说明
- 列出所有路径变更

## 四、风险和缓解措施

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 大量代码需要更新 imports | 高 | 使用自动化脚本批量更新 |
| 可能破坏现有功能 | 高 | 完整的测试覆盖，分阶段迁移 |
| 合并冲突 | 中 | 及时与主分支同步，小步提交 |
| 向后兼容性 | 中 | 提供兼容层或迁移期双路径支持 |

## 五、验收标准

### 5.1 代码组织
- [ ] common/ 只包含纯基础设施代码
- [ ] 无业务逻辑在 common/ 中
- [ ] pkg/ 包含所有应用层代码
- [ ] 清晰的层次边界

### 5.2 功能验证
- [ ] 所有服务正常编译
- [ ] 单元测试全部通过
- [ ] 集成测试全部通过
- [ ] 无性能回归

### 5.3 文档完整
- [ ] CLAUDE.md 已更新
- [ ] README 已更新
- [ ] 迁移指南已创建
- [ ] API 文档已更新

## 六、时间线

| 阶段 | 工作内容 | 预计时间 | 负责人 |
|------|---------|---------|--------|
| 准备 | 依赖分析、目录创建 | 1天 | - |
| 阶段一 | 高优先级迁移 | 2天 | - |
| 阶段二 | 中优先级迁移 | 2天 | - |
| 阶段三 | 测试验证 | 2天 | - |
| 阶段四 | 文档更新 | 1天 | - |
| **总计** | - | **8天** | - |

## 七、回滚计划

如果重组过程中出现严重问题：

1. 保留原分支作为备份
2. 可以分模块回滚（每个包独立）
3. 提供兼容层支持渐进式迁移

## 八、长期收益

1. **架构清晰**：明确的层次边界，符合清洁架构原则
2. **可维护性**：代码组织合理，易于理解和修改
3. **可复用性**：基础设施层可以被其他项目使用
4. **可测试性**：清晰的依赖关系，易于单元测试
5. **团队协作**：明确的代码归属，减少冲突

## 九、下一步行动

1. [ ] 评审本重组计划
2. [ ] 创建迁移分支
3. [ ] 开始第一阶段实施
4. [ ] 每日同步进度
5. [ ] 完成后进行代码评审

---

**文档版本**: v1.0
**创建日期**: 2024-11-10
**作者**: Claude
**状态**: 待评审