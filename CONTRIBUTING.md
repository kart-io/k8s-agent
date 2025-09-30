# 贡献指南

感谢您对 Aetherius 项目的关注！我们欢迎各种形式的贡献。

---

## 目录

- [行为准则](#行为准则)
- [如何贡献](#如何贡献)
- [开发流程](#开发流程)
- [代码规范](#代码规范)
- [提交规范](#提交规范)
- [Pull Request 流程](#pull-request-流程)
- [测试](#测试)
- [文档](#文档)

---

## 行为准则

本项目采用 [Contributor Covenant](https://www.contributor-covenant.org/) 行为准则。参与本项目即表示您同意遵守该准则。

### 我们的承诺

- 尊重不同的观点和经验
- 优雅地接受建设性批评
- 关注什么对社区最有利
- 对其他社区成员表现出同理心

---

## 如何贡献

### 报告 Bug

在提交 Bug 之前:

1. 检查 [Issues](https://github.com/kart-io/k8s-agent/issues) 是否已有相同问题
2. 使用最新版本复现问题
3. 收集相关信息 (日志、版本、环境等)

提交 Bug 时请包含:

- **清晰的标题**: 简洁描述问题
- **详细描述**: 问题的详细信息
- **复现步骤**: 如何复现问题
- **预期行为**: 您期望发生什么
- **实际行为**: 实际发生了什么
- **环境信息**: 版本、操作系统、K8s 版本等
- **日志**: 相关的错误日志
- **截图**: 如果适用

示例:

```markdown
### Bug 描述
Agent Manager 无法连接到 NATS

### 复现步骤
1. 启动 Docker Compose
2. 查看 agent-manager 日志
3. 看到连接错误

### 预期行为
应该成功连接到 NATS

### 实际行为
连接超时: "dial tcp: i/o timeout"

### 环境
- Aetherius 版本: v1.0.0
- Docker 版本: 20.10.21
- 操作系统: Ubuntu 22.04
```

---

### 建议新功能

提交功能建议时请包含:

- **功能描述**: 详细描述建议的功能
- **使用场景**: 为什么需要这个功能
- **预期效果**: 功能应该如何工作
- **替代方案**: 是否有其他实现方式
- **附加信息**: 相关参考、示例等

---

### 贡献代码

我们欢迎以下类型的代码贡献:

- **Bug 修复**: 修复已知的 Bug
- **新功能**: 实现新的功能
- **性能优化**: 提升性能
- **代码重构**: 改进代码质量
- **测试**: 添加或改进测试
- **文档**: 改进文档

---

## 开发流程

### 1. Fork 仓库

点击页面右上角的 "Fork" 按钮。

### 2. 克隆您的 Fork

```bash
git clone https://github.com/YOUR_USERNAME/k8s-agent.git
cd k8s-agent
```

### 3. 添加上游仓库

```bash
git remote add upstream https://github.com/kart-io/k8s-agent.git
```

### 4. 创建分支

```bash
git checkout -b feature/my-new-feature
# 或
git checkout -b fix/bug-description
```

分支命名规范:

- `feature/`: 新功能
- `fix/`: Bug 修复
- `docs/`: 文档更新
- `refactor/`: 代码重构
- `test/`: 测试相关
- `chore/`: 构建、配置等

### 5. 开发

#### 设置开发环境

```bash
# 启动依赖服务
cd deployments/docker-compose
docker-compose up -d postgres redis nats neo4j

# 运行服务
cd ../../agent-manager
make run
```

#### 进行更改

- 遵循代码规范
- 添加必要的测试
- 更新相关文档
- 保持提交原子性

### 6. 测试

```bash
# 运行测试
make test

# 运行 lint
make lint

# 运行格式化
make format
```

### 7. 提交更改

```bash
git add .
git commit -m "feat: add new feature"
```

### 8. 同步上游

```bash
git fetch upstream
git rebase upstream/main
```

### 9. 推送到您的 Fork

```bash
git push origin feature/my-new-feature
```

### 10. 创建 Pull Request

访问 GitHub 页面,点击 "Create Pull Request"。

---

## 代码规范

### Go 代码规范

遵循 [Effective Go](https://golang.org/doc/effective_go.html) 和 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)。

#### 命名规范

```go
// 包名: 小写,简短,无下划线
package agent

// 导出的函数、类型: PascalCase
func RegisterAgent() {}
type AgentManager struct {}

// 未导出的: camelCase
func parseConfig() {}
var agentCache map[string]*Agent

// 常量: PascalCase 或 ALL_CAPS
const (
    DefaultTimeout = 30 * time.Second
    MAX_RETRIES    = 3
)

// 接口: 单方法接口以 -er 结尾
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

#### 错误处理

```go
// 优先使用 errors.New 或 fmt.Errorf
if err != nil {
    return fmt.Errorf("failed to register agent: %w", err)
}

// 自定义错误类型
type ValidationError struct {
    Field string
    Err   error
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %v", e.Field, e.Err)
}
```

#### 注释

```go
// RegisterAgent 注册一个新的 Agent 到中央控制平面
//
// 参数:
//   agent: Agent 信息
//
// 返回:
//   error: 注册失败时返回错误
func RegisterAgent(agent *Agent) error {
    // 实现...
}
```

---

### Python 代码规范

遵循 [PEP 8](https://www.python.org/dev/peps/pep-0008/)。

#### 命名规范

```python
# 模块名: 小写,下划线分隔
import root_cause_analyzer

# 类名: PascalCase
class RootCauseAnalyzer:
    pass

# 函数、变量: 小写,下划线分隔
def analyze_logs(log_data):
    max_retries = 3

# 常量: 大写,下划线分隔
MAX_CONFIDENCE = 0.95
DEFAULT_TIMEOUT = 30
```

#### 类型注解

```python
from typing import List, Dict, Optional

def analyze(
    context: AnalysisContext,
    options: Optional[Dict] = None
) -> AnalysisResult:
    """分析故障根因

    Args:
        context: 分析上下文
        options: 可选配置

    Returns:
        分析结果

    Raises:
        ValueError: 参数无效时抛出
    """
    pass
```

#### 文档字符串

```python
class RecommendationEngine:
    """推荐引擎

    生成基于规则的修复建议。

    Attributes:
        rules: 推荐规则列表
        confidence_threshold: 置信度阈值
    """

    def recommend(self, root_cause: RootCause) -> List[Recommendation]:
        """生成推荐

        Args:
            root_cause: 根因信息

        Returns:
            推荐列表,按置信度排序
        """
        pass
```

---

## 提交规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范。

### 格式

```text
<type>(<scope>): <subject>

<body>

<footer>
```

### Type

- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式 (不影响功能)
- `refactor`: 重构
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建、配置等
- `ci`: CI/CD 相关

### Scope

可选,表示影响范围:

- `agent`: collect-agent
- `manager`: agent-manager
- `orchestrator`: orchestrator-service
- `reasoning`: reasoning-service
- `api`: API 相关
- `docs`: 文档
- `deps`: 依赖

### 示例

```bash
# 新功能
git commit -m "feat(reasoning): add failure prediction engine"

# Bug 修复
git commit -m "fix(manager): resolve NATS connection timeout"

# 文档
git commit -m "docs(api): update API reference for v1.1"

# 多行提交
git commit -m "feat(orchestrator): add parallel step execution

- Implement parallel executor
- Add concurrent step runner
- Update workflow engine

Closes #123"
```

---

## Pull Request 流程

### 1. PR 标题

遵循提交规范:

```text
feat(reasoning): add failure prediction engine
fix(manager): resolve NATS connection timeout
```

### 2. PR 描述

使用模板:

```markdown
## 概述
简要描述此 PR 的目的

## 变更类型
- [ ] Bug 修复
- [ ] 新功能
- [ ] 文档更新
- [ ] 代码重构
- [ ] 性能优化

## 相关 Issue
Closes #123

## 变更内容
- 添加了 X 功能
- 修复了 Y 问题
- 优化了 Z 性能

## 测试
- [ ] 添加了单元测试
- [ ] 手动测试通过
- [ ] 所有测试通过

## 截图
如果适用,添加截图

## 检查清单
- [ ] 代码遵循项目规范
- [ ] 添加了必要的测试
- [ ] 更新了相关文档
- [ ] 所有测试通过
- [ ] 没有引入新的警告
```

### 3. 代码审查

- 响应审查意见
- 进行必要的修改
- 保持讨论友好和建设性

### 4. 合并

一旦 PR 被批准且所有检查通过:

- Maintainer 会合并 PR
- 您的贡献会出现在 Contributors 列表中

---

## 测试

### Go 测试

```bash
# 运行所有测试
make test

# 运行特定包测试
go test ./internal/agent/...

# 带覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### 编写测试

```go
func TestRegisterAgent(t *testing.T) {
    // 准备
    registry := NewRegistry()
    agent := &Agent{
        ID:        "test-agent",
        ClusterID: "test-cluster",
    }

    // 执行
    err := registry.RegisterAgent(context.Background(), agent)

    // 断言
    if err != nil {
        t.Errorf("RegisterAgent() error = %v", err)
    }
}
```

---

### Python 测试

```bash
# 运行测试
pytest tests/

# 带覆盖率
pytest --cov=internal tests/

# 生成 HTML 报告
pytest --cov=internal --cov-report=html tests/
```

#### 编写测试

```python
def test_root_cause_analysis():
    # 准备
    analyzer = RootCauseAnalyzer()
    context = AnalysisContext(
        event={"reason": "OOMKilled"},
        logs="out of memory"
    )

    # 执行
    result = analyzer.analyze(context)

    # 断言
    assert result.root_cause.type == "OOMKiller"
    assert result.confidence >= 0.8
```

---

## 文档

### 更新文档

在添加或修改功能时,请更新相关文档:

- **README.md**: 项目介绍、快速开始
- **API 文档**: API 接口变更
- **架构文档**: 架构变更
- **组件 README**: 组件功能变更

### 文档规范

- 使用 Markdown 格式
- 保持清晰简洁
- 添加代码示例
- 包含截图 (如果适用)
- 更新目录

---

## 获取帮助

如有任何问题:

- 查看 [文档](docs/)
- 搜索 [Issues](https://github.com/kart-io/k8s-agent/issues)
- 在 [Discussions](https://github.com/kart-io/k8s-agent/discussions) 提问
- 加入 [Slack](https://aetherius-slack.example.com)

---

## 许可证

贡献代码即表示您同意根据项目的 [MIT License](LICENSE) 授权您的贡献。

---

## 致谢

感谢所有贡献者！您的贡献让 Aetherius 更加强大。

[![Contributors](https://contrib.rocks/image?repo=kart-io/k8s-agent)](https://github.com/kart-io/k8s-agent/graphs/contributors)