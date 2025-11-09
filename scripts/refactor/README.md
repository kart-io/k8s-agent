# 代码重构工具集

本目录包含用于代码规范统一的自动化脚本。

## 脚本列表

### 1. check-architecture.sh

**用途**：检查整个项目的架构一致性

**使用方法**：

```bash
./scripts/refactor/check-architecture.sh
```

**输出示例**：

```
======================================
Architecture Consistency Check
======================================

✓ agent-manager        : options/
✓ agent-manager        : initializers/
✓ agent-manager        : Application interface
✓ agent-manager        : New logger

✗ reasoning            : options/
✗ reasoning            : initializers/
✗ reasoning            : Application interface
✗ reasoning            : New logger

======================================
Summary
======================================
Total checks: 32
Passed: 16
Failed: 16
Pass rate: 50.0%
```

### 2. migrate-logger.sh

**用途**：批量迁移服务的日志系统到 kart-io/logger

**使用方法**：

```bash
./scripts/refactor/migrate-logger.sh <service-name>
```

**示例**：

```bash
# 迁移 reasoning 服务
./scripts/refactor/migrate-logger.sh reasoning

# 迁移 collect-agent 服务
./scripts/refactor/migrate-logger.sh collect-agent
```

**功能**：
- 自动备份原文件
- 批量替换导入语句
- 替换类型声明
- 编译验证
- 失败时提供回滚指令

**输出示例**：

```
======================================
Migrating reasoning to kart-io/logger
======================================

1. Creating backup at internal/reasoning.logger-migration-backup-20251030-120000
✓ Backup created

2. Finding files to migrate...
Found 8 files to migrate

3. Replacing imports...
✓ Imports replaced

4. Compiling to verify changes...
✓ Compilation successful!

Migration completed successfully!

Next steps:
1. Review the changes carefully
2. Update logger initialization code if needed
3. Run tests: go test ./internal/reasoning/... -v
4. If everything works, delete backup: rm -rf internal/reasoning.logger-migration-backup-20251030-120000
```

### 3. verify-service.sh

**用途**：验证服务是否符合标准架构

**使用方法**：

```bash
./scripts/refactor/verify-service.sh <service-name>
```

**示例**：

```bash
# 验证 agent-manager 服务
./scripts/refactor/verify-service.sh agent-manager

# 验证 reasoning 服务
./scripts/refactor/verify-service.sh reasoning
```

**检查项目**：
1. 目录结构（options/, initializers/）
2. Application 接口实现
3. 日志系统
4. Bootstrap 框架使用
5. 编译成功
6. 测试通过
7. 代码质量（Linter）

**输出示例**：

```
======================================
Verifying agent-manager Service
======================================

=== 1. Directory Structure ===
✓ cmd/agent-manager/app/options/ exists
✓ internal/agent-manager/initializers/ exists
✓ cmd/agent-manager/app/app.go exists

=== 2. Application Interface ===
✓ Uses commonapp.RunWithRunner()
✓ Implements Initialize() method
✓ Implements Run() method
✓ Implements Shutdown() method

=== 3. Logger System ===
✓ Uses kart-io/logger
✓ Not using old common/logger

=== 4. Bootstrap Framework ===
✓ Uses bootstrap.Bootstrap
✓ Registers initializers

=== 5. Compilation ===
✓ Service compiles successfully

=== 6. Tests ===
✓ Tests pass

=== 7. Code Quality ===
✓ Linter passes

======================================
Verification Summary
======================================
Passed:  14
Failed:  0
Warnings: 0

✓ Service meets all critical standards!
```

## 使用工作流

### 完整重构流程

```bash
# 1. 检查当前架构状态
./scripts/refactor/check-architecture.sh

# 2. 迁移 reasoning 服务的日志
./scripts/refactor/migrate-logger.sh reasoning

# 3. 验证迁移结果
./scripts/refactor/verify-service.sh reasoning

# 4. 迁移 collect-agent 服务
./scripts/refactor/migrate-logger.sh collect-agent
./scripts/refactor/verify-service.sh collect-agent

# 5. 最终检查
./scripts/refactor/check-architecture.sh
```

### 单服务重构流程

```bash
SERVICE=reasoning

# 1. 迁移日志系统
./scripts/refactor/migrate-logger.sh $SERVICE

# 2. 手动创建 options 和 initializers（参考 REFACTORING_PLAN.md）
# ...

# 3. 验证服务
./scripts/refactor/verify-service.sh $SERVICE

# 4. 运行测试
make test-$SERVICE

# 5. 启动服务验证功能
make run-$SERVICE
```

## 安全特性

所有脚本都包含以下安全特性：

1. **自动备份**：修改前自动备份原文件
2. **编译验证**：修改后立即编译验证
3. **回滚指令**：失败时提供明确的回滚命令
4. **错误处理**：使用 `set -e` 遇错即停

## 相关文档

- [CODE_STANDARDIZATION.md](../docs/CODE_STANDARDIZATION.md) - 代码规范统一方案
- [REFACTORING_PLAN.md](../docs/REFACTORING_PLAN.md) - 详细重构执行计划
- [CODE_REORGANIZATION.md](../docs/CODE_REORGANIZATION.md) - 代码重组方案

## 故障排除

### 问题：脚本无法执行

```bash
# 确保脚本有可执行权限
chmod +x scripts/refactor/*.sh
```

### 问题：迁移后编译失败

```bash
# 查看详细错误日志
cat /tmp/build-<service>.log

# 回滚到备份
SERVICE=reasoning
BACKUP=$(ls -td internal/${SERVICE}.logger-migration-backup-* | head -1)
rm -rf internal/$SERVICE
mv $BACKUP internal/$SERVICE
```

### 问题：测试失败

```bash
# 查看详细测试日志
cat /tmp/test-<service>-verify.log

# 单独运行失败的测试
go test ./internal/<service>/path/to/failing/test -v
```

## 贡献

如需添加新的重构脚本：

1. 遵循现有脚本的命名规范
2. 包含详细的使用说明和示例
3. 添加错误处理和备份机制
4. 更新此 README
