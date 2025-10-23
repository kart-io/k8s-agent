# 决策记录：Makefile 命令格式统一

## 背景

项目中存在两套 Makefile 命令格式：

1. **新格式（模块化）**: `make go.build.agent-manager`
2. **旧格式（兼容性）**: `make build-agent-manager`

这导致开发者困惑，不知道应该使用哪个命令。

## 问题

- 用户不清楚两个命令的区别
- 文档中混用两种格式
- 新手不知道选择哪一个
- CI/CD 脚本使用不一致

## 考虑的方案

### 方案 1：保留两种格式（采纳） ✅

**实现**:
- 保持两种格式都可用
- 在帮助信息中明确标注 `[LEGACY]`
- 更新分类名称为 "Legacy Aliases - prefer xxx"
- 在 CLAUDE.md 中说明差异和推荐用法
- 创建详细的命令参考文档

**优点**:
- ✅ 向后兼容，不破坏现有脚本
- ✅ 用户可以自由选择（新手用短格式，专家用长格式）
- ✅ 渐进式迁移，不强制改变习惯
- ✅ 符合 OneX 项目的设计理念

**缺点**:
- ❌ 维护两套别名
- ❌ 可能导致命令格式不一致

### 方案 2：只保留新格式

**实现**:
- 删除所有 `build-%`, `docker-%` 等别名
- 强制使用 `go.build.%`, `docker.build.%` 等

**优点**:
- ✅ 命令格式统一
- ✅ 减少维护成本

**缺点**:
- ❌ 破坏向后兼容性
- ❌ 需要更新所有文档和脚本
- ❌ 用户需要重新学习
- ❌ 违背渐进式改进原则

### 方案 3：只保留旧格式

**实现**:
- 删除 `go.build.%` 等模块化命令
- 统一使用简短格式

**优点**:
- ✅ 命令更短，易于输入

**缺点**:
- ❌ 失去模块化清晰度
- ❌ 与 OneX 设计背道而驰
- ❌ 不利于大型项目扩展

## 决策

**采用方案 1：保留两种格式，明确推荐新格式**

### 具体措施

1. **Makefile 改进**:
   ```makefile
   ##@ Build (Legacy Aliases - prefer go.build.* commands)

   .PHONY: build-%
   build-%: ## [LEGACY] Build specific service - prefer 'make go.build.SERVICE'
       @$(MAKE) go.build.$*
   ```

2. **文档更新**:
   - 在 CLAUDE.md 中添加"Command Format Guide"章节
   - 创建 docs/MAKEFILE_COMMANDS.md 完整参考
   - 在所有示例中优先使用新格式

3. **帮助信息优化**:
   - 分类名称改为 "Legacy Aliases - prefer xxx"
   - 具体命令标注 `[LEGACY]`

4. **推荐实践**:
   - 新代码：使用新格式 `go.build.%`
   - CI/CD：使用新格式
   - 个人开发：可以使用任一格式
   - 旧脚本：无需修改，继续工作

## 影响

### 用户影响
- **现有用户**: 无影响，两种格式都可用
- **新用户**: 有清晰指引，知道推荐使用新格式
- **脚本维护者**: 建议逐步迁移到新格式

### 维护影响
- 需要维护两套别名（额外 ~50 行代码）
- 帮助信息更清晰
- 文档需要说明两种格式

## 迁移计划

### 短期（当前 - 6 个月）
- ✅ 保留所有兼容性别名
- ✅ 在帮助中标注推荐
- ✅ 更新文档使用新格式

### 中期（6 个月 - 2 年）
- 在发布说明中提示旧格式将被废弃
- 在旧格式命令执行时输出警告（可选）
- 继续保持功能正常

### 长期（2 年+）
- 根据社区反馈决定是否移除
- 如果 99% 用户已迁移，可以考虑移除
- 发布 breaking change 版本

## 参考资料

- [OneX Project Makefile](https://github.com/superproj/onex/blob/main/Makefile)
- [GNU Make Pattern Rules](https://www.gnu.org/software/make/manual/html_node/Pattern-Rules.html)
- [Kubernetes Makefile](https://github.com/kubernetes/kubernetes/blob/master/Makefile) - 也使用类似的兼容性策略

## 相关文档

- [MAKEFILE_COMMANDS.md](../MAKEFILE_COMMANDS.md) - 完整命令参考
- [CLAUDE.md](../../CLAUDE.md) - 项目开发指南
- [CONTRIBUTING.md](../../CONTRIBUTING.md) - 贡献指南

---

**决策日期**: 2025-10-23
**决策人**: 项目维护团队
**状态**: 已采纳
**影响**: 低（向后兼容）
