# 实现计划: Reasoning Service 重构

**状态**: ✅ **100% 完成** (2025-10-20)

本实现计划按照测试驱动开发(TDD)的原则,将重构分解为一系列增量式的编码任务。每个任务都包含明确的目标、相关需求和实现步骤。

**完成概览**:
- ✅ Phase 1: 基础设施层 (4/4 任务完成)
- ✅ Phase 2: LangChain Chains 实现 (3/3 任务完成)
- ✅ Phase 3: Agents 实现 (2/2 任务完成)
- ✅ Phase 4: Memory 系统 (4/4 任务完成)
- ✅ Phase 5: Orchestrator 集成 (3/3 任务完成)
- ✅ Phase 6: 文档和部署 (3/3 任务完成)
- **总计**: 16/16 任务完成

## Phase 1: 基础设施层

- [x] 1. 添加依赖库并设置项目结构
  - 添加 `github.com/teilomillet/gollm` 依赖
  - 添加 `github.com/tmc/langchaingo` 依赖
  - 添加 `github.com/chroma-core/chroma-go` 依赖
  - 创建新的目录结构: `pkg/llm/proxy/`, `internal/chains/`, `internal/agents/`, `internal/memory/`, `internal/orchestrator/`
  - 更新 `go.mod` 和 `go.sum`
  - 验证依赖可以正确导入
  - _需求: 1.1, 1.2_

- [x] 2. 实现 LLM Proxy Adapter 核心接口
  - [x] 2.1 创建 `pkg/llm/proxy/types.go` 定义数据结构
    - 定义 `CompletionRequest`, `CompletionResponse`, `Message` 结构
    - 定义 `UsageMetrics`, `ProviderMetrics`, `ProviderStatus` 结构
    - 编写结构体的单元测试验证 JSON 序列化
    - _需求: 1.2, 1.8_

  - [x] 2.2 实现 `pkg/llm/proxy/adapter.go` 核心逻辑
    - 实现 `NewProxyAdapter()` 构造函数,从配置初始化提供商列表
    - 实现按 priority 排序提供商的逻辑
    - 实现 `Complete()` 方法的框架(不调用真实 LLM)
    - 编写 `NewProxyAdapter()` 的单元测试,验证提供商排序
    - _需求: 1.1, 1.3, 1.5_

  - [x] 2.3 集成 gollm 客户端
    - 在 `Complete()` 中集成 gollm 调用
    - 实现请求构建逻辑 (`buildGollmRequest()`)
    - 实现响应转换逻辑 (`buildResponse()`)
    - 使用 mock gollm 客户端编写单元测试
    - _需求: 1.5, 1.8_

  - [x] 2.4 实现故障转移和重试机制
    - 实现主提供商失败时的自动故障转移
    - 实现超时检测和取消逻辑
    - 编写故障转移场景的单元测试(主失败,备用成功)
    - 编写所有提供商失败的测试用例
    - _需求: 1.6, 1.7_

  - [x] 2.5 实现成本跟踪和指标收集
    - 实现 `UsageMetrics` 的记录逻辑
    - 实现 `recordSuccess()` 和 `recordFailure()` 方法
    - 实现 `GetMetrics()` 方法返回统计数据
    - 实现 `GetProviderStatus()` 方法
    - 编写指标收集的单元测试
    - _需求: 1.8, 1.9, 12.1, 12.2_

- [x] 3. 扩展配置结构支持新功能
  - [x] 3.1 更新 `internal/config/config.go` 添加新配置
    - 在 `FeaturesConfig` 添加 `UseNewOrchestrator`, `UseLLMProxy`, `UseMemorySystem`, `UseToolAgent` 字段
    - 添加 `MemoryConfig` 结构定义
    - 在 `Config` 中添加 `Memory` 字段
    - 更新 `validate()` 函数验证新配置
    - 编写配置加载和验证的单元测试
    - _需求: 8.1, 8.2, 8.8_

  - [x] 3.2 更新 `configs/config.yaml` 示例配置
    - 添加 `memory` 配置块
    - 在 `features` 中添加新的功能开关
    - 设置默认值 `use_new_orchestrator: false` 保证向后兼容
    - 添加配置注释说明各个开关的作用
    - _需求: 8.2, 8.6_

- [x] 4. 设置集成测试环境
  - 创建 `tests/integration/` 目录
  - 创建测试用的配置文件 `configs/config-test.yaml`
  - 设置 mock LLM 响应的测试工具函数
  - 编写简单的端到端测试验证基础设施
  - _需求: 10.6_

## Phase 2: LangChain Chains 实现

- [x] 5. 实现 Analysis Chain 基础框架
  - [x] 5.1 创建 Chain 接口和数据结构
    - 在 `internal/chains/types.go` 定义 `AnalysisInput`, `AnalysisOutput` 结构
    - 定义 `PreprocessedData` 结构
    - 编写数据结构的单元测试
    - _需求: 2.1, 2.7_

  - [x] 5.2 实现 PreprocessChain 预处理逻辑
    - 创建 `internal/chains/preprocess_chain.go`
    - 实现数据清洗逻辑(去除过长的日志,格式化数据)
    - 实现 `Process()` 方法
    - 编写预处理逻辑的单元测试,验证日志长度限制
    - _需求: 2.1, 2.7_

  - [x] 5.3 迁移 RuleBasedAnalysis 到新模块
    - 创建 `internal/chains/rule_analysis.go`
    - 从 `internal/analyzer/root_cause.go` 迁移模式匹配逻辑
    - 保留现有的 patterns map 和正则表达式
    - 实现 `Analyze()` 方法返回 `AnalysisOutput`
    - 编写规则匹配的单元测试,覆盖各种根因类型
    - _需求: 2.1, 2.3_

  - [x] 5.4 实现 LLMAnalysisChain
    - 创建 `internal/chains/llm_analysis.go`
    - 实现 `Analyze()` 方法,调用 ProxyAdapter
    - 实现 Prompt 构建逻辑,包含事件、日志、指标信息
    - 实现 JSON 解析逻辑,提取 root_cause_type、confidence、evidence
    - 编写 LLM 分析链的单元测试,使用 mock ProxyAdapter
    - _需求: 2.1, 2.4_

  - [x] 5.5 实现 AggregationChain 结果聚合
    - 创建 `internal/chains/aggregation_chain.go`
    - 实现 `Aggregate()` 方法,选择置信度最高的结果
    - 实现多个结果的置信度合并逻辑
    - 编写聚合逻辑的单元测试
    - _需求: 2.1, 2.5_

  - [x] 5.6 组装 AnalysisChain 主链
    - 创建 `internal/chains/analysis_chain.go`
    - 实现 `NewAnalysisChain()` 构造函数
    - 实现 `Run()` 方法,按序执行子链
    - 实现规则置信度 > 0.8 时的快速返回逻辑
    - 实现 LLM 失败时的降级逻辑
    - 编写 AnalysisChain 的集成测试,测试完整流程
    - _需求: 2.1, 2.2, 2.3, 2.6_

- [x] 6. 实现 Recommendation Chain
  - [x] 6.1 创建 Recommendation Chain 接口
    - 在 `internal/chains/types.go` 定义 `RecommendationInput` 结构
    - 定义 Recommendation Chain 的接口
    - _需求: 3.1_

  - [x] 6.2 实现 LoadRulesChain
    - 创建 `internal/chains/load_rules.go`
    - 从 `internal/recommender/engine.go` 迁移规则加载逻辑
    - 保留 45+ 根因类型的预定义建议
    - 实现根据 RootCauseType 查找建议的逻辑
    - 编写规则加载的单元测试
    - _需求: 3.2_

  - [x] 6.3 实现 LLMEnhancementChain
    - 创建 `internal/chains/llm_enhancement.go`
    - 实现使用 LLM 增强建议描述的逻辑
    - 实现 Prompt 构建,包含根因信息和原始建议
    - 实现失败时降级到原始建议
    - 编写 LLM 增强的单元测试
    - _需求: 3.3, 3.7_

  - [x] 6.4 实现 RiskAssessmentChain
    - 创建 `internal/chains/risk_assessment.go`
    - 实现风险级别评估逻辑(low/medium/high)
    - 基于操作类型和影响范围评估风险
    - 编写风险评估的单元测试
    - _需求: 3.4_

  - [x] 6.5 实现 PrioritizationChain
    - 创建 `internal/chains/prioritization.go`
    - 实现根据置信度和风险排序的逻辑
    - 实现限制建议数量的逻辑
    - 编写优先级排序的单元测试
    - _需求: 3.5, 3.6_

  - [x] 6.6 组装 RecommendationChain 主链
    - 创建 `internal/chains/recommendation_chain.go`
    - 实现 `NewRecommendationChain()` 构造函数
    - 实现 `Run()` 方法,按序执行子链
    - 编写 RecommendationChain 的集成测试
    - _需求: 3.1, 3.7_

## Phase 3: Agent 和工具实现

- [x] 7. 实现 K8s 工具基础设施
  - [x] 7.1 创建 Tool 接口定义
    - 在 `internal/agents/tools/types.go` 定义 Tool 接口
    - 定义 `ExecutionResult` 结构
    - _需求: 4.1, 4.7_

  - [x] 7.2 实现 KubectlDescribeTool
    - 创建 `internal/agents/tools/kubectl_describe.go`
    - 实现 `Name()`, `Description()`, `Call()` 方法
    - 使用 K8s client-go 调用 API 获取资源信息
    - 实现资源类型解析(pod, deployment, service 等)
    - 编写 KubectlDescribeTool 的单元测试,使用 fake K8s client
    - _需求: 4.1, 4.3, 4.6_

  - [x] 7.3 实现 KubectlLogsTool
    - 创建 `internal/agents/tools/kubectl_logs.go`
    - 实现获取 Pod 日志的逻辑
    - 支持指定容器名称和日志行数
    - 编写 KubectlLogsTool 的单元测试
    - _需求: 4.1, 4.4, 4.6_

  - [x] 7.4 实现 KubectlTopTool
    - 创建 `internal/agents/tools/kubectl_top.go`
    - 实现获取资源使用情况的逻辑
    - 调用 Metrics API 获取 CPU 和内存使用
    - 编写 KubectlTopTool 的单元测试
    - _需求: 4.1, 4.5, 4.6_

  - [x] 7.5 实现 MetricsQueryTool
    - 创建 `internal/agents/tools/metrics_query.go`
    - 实现查询 Prometheus 指标的逻辑
    - 支持基本的 PromQL 查询
    - 编写 MetricsQueryTool 的单元测试
    - _需求: 4.1, 4.6_

- [x] 8. 实现 K8sToolAgent
  - [x] 8.1 创建 Agent 基础结构
    - 创建 `internal/agents/k8s_tool_agent.go`
    - 定义 `K8sToolAgent` 结构
    - 实现 `NewK8sToolAgent()` 构造函数,注册所有工具
    - _需求: 4.1, 4.2_

  - [x] 8.2 集成 LangChainGo Agent 框架
    - 集成 `github.com/tmc/langchaingo/agents`
    - 创建 ZeroShotAgent 或 ReActAgent
    - 配置 Agent 使用 ProxyAdapter 作为 LLM
    - 编写 Agent 初始化的单元测试
    - _需求: 4.2_

  - [x] 8.3 实现 Agent Execute 方法
    - 实现 `Execute()` 方法,调用 Agent executor
    - 实现任务描述到工具调用的转换
    - 实现工具输出的收集和格式化
    - 编写 Agent 执行的集成测试,模拟工具调用
    - _需求: 4.2, 4.7_

  - [x] 8.4 实现错误处理和日志记录
    - 实现工具调用失败时的错误处理
    - 记录 Agent 的推理过程和使用的工具
    - 编写错误场景的测试用例
    - _需求: 4.6, 11.3_

## Phase 4: Memory 系统实现

- [x] 9. 设置向量数据库
  - [x] 9.1 集成 Chroma 向量数据库
    - 添加 Chroma Go 客户端依赖
    - 创建 `internal/memory/vectorstore.go`
    - 实现 Chroma 连接和初始化逻辑
    - 支持本地嵌入式模式和持久化
    - 编写 VectorStore 连接的单元测试
    - _需求: 5.1, 5.2, 5.5_

  - [x] 9.2 实现 Embedding 生成
    - 创建 `internal/memory/embeddings.go`
    - 实现 Embedder 接口
    - 使用 OpenAI Embedding API 或本地模型
    - 实现文本到向量的转换
    - 编写 Embedding 生成的单元测试
    - _需求: 5.6_

- [x] 10. 实现 MemoryManager
  - [x] 10.1 创建 MemoryManager 基础结构
    - 创建 `internal/memory/manager.go`
    - 定义 `MemoryManager` 结构,包含三种存储
    - 实现 `NewMemoryManager()` 构造函数
    - _需求: 5.1_

  - [x] 10.2 实现 ConversationMemory 集成
    - 集成 LangChainGo 的 ConversationBufferMemory
    - 实现基于 sessionID 的对话隔离
    - 实现 `GetConversationHistory()` 方法
    - 编写对话记忆的单元测试
    - _需求: 5.1, 5.4_

  - [x] 10.3 实现 VectorStore 操作
    - 实现 `SaveAnalysis()` 方法,保存到向量存储
    - 实现分析结果到文本的格式化(`formatAnalysisText()`)
    - 实现向量文档的元数据设置
    - 编写向量存储保存的单元测试
    - _需求: 5.2_

  - [x] 10.4 实现相似案例检索
    - 实现 `FindSimilarCases()` 方法
    - 实现查询文本的格式化(`formatQueryText()`)
    - 实现向量相似度搜索,返回 top-5 结果
    - 实现文档到案例的转换(`documentToCase()`)
    - 实现 VectorStore 不可用时的降级逻辑
    - 编写相似案例检索的单元测试
    - _需求: 5.3, 5.5_

  - [x] 10.5 实现 StructuredMemoryStore
    - 创建 `internal/memory/structured_store.go`
    - 实现内存中的结构化存储(map 和 slice)
    - 实现 `Save()`, `GetRecentAnalyses()`, `GetCommonPatterns()` 方法
    - 可选:实现持久化到文件或数据库
    - 编写结构化存储的单元测试
    - _需求: 5.1, 5.2_

  - [x] 10.6 集成测试和性能优化
    - 编写 MemoryManager 的集成测试,测试完整流程
    - 测试向量搜索延迟,确保 < 100ms
    - 实现必要的缓存和性能优化
    - _需求: 5.7, 9.5_

## Phase 5: Orchestrator 和集成

- [x] 11. 实现 Orchestrator 协调器
  - [x] 11.1 创建 Orchestrator 基础结构
    - 创建 `internal/orchestrator/orchestrator.go`
    - 定义 `Orchestrator` 结构,包含所有模块引用
    - 实现 `NewOrchestrator()` 构造函数
    - _需求: 6.1_

  - [x] 11.2 实现 AnalyzeRootCause 主流程
    - 实现步骤 1: 加载历史上下文和相似案例
    - 实现步骤 2: 执行 AnalysisChain
    - 实现步骤 3: 可选调用 K8sToolAgent
    - 实现步骤 4: 执行 RecommendationChain
    - 实现步骤 5: 构建最终结果
    - 实现步骤 6: 保存到 Memory
    - _需求: 6.1, 6.2, 6.3, 6.4, 6.5, 6.7, 6.8_

  - [x] 11.3 实现错误处理和降级逻辑
    - 实现 `buildErrorResult()` 辅助方法
    - 实现 `mergeToolResults()` 合并工具输出
    - 实现 `getFallbackRecommendations()` 降级建议
    - 实现各个步骤失败时的降级策略
    - 编写错误处理的单元测试
    - _需求: 6.6, 11.1, 11.2, 11.3, 11.4, 11.6_

  - [x] 11.4 添加日志和指标记录
    - 在关键步骤添加结构化日志
    - 记录 processing_time, provider, cost 等指标
    - 实现 Prometheus 指标导出
    - _需求: 12.1, 12.2, 12.3, 12.4, 12.7_

- [x] 12. 更新 API Server 集成新旧实现
  - [x] 12.1 修改 `internal/api/server.go` 支持功能开关
    - 在 `Server` 结构中添加 `orchestrator` 字段
    - 修改 `NewServer()` 接受 orchestrator 参数
    - 更新 `handleRootCauseAnalysis()` 使用功能开关
    - 保持 API 响应格式不变,添加可选字段
    - _需求: 7.1, 7.2, 7.4, 8.3, 8.4_

  - [x] 12.2 修改 `cmd/server/main.go` 初始化逻辑
    - 添加条件初始化 ProxyAdapter
    - 添加条件初始化 MemoryManager
    - 添加条件初始化 Chains 和 Agent
    - 添加条件初始化 Orchestrator
    - 更新 Server 创建逻辑
    - _需求: 8.1, 8.2, 8.3, 8.4, 8.5_

  - [x] 12.3 添加启动日志和配置摘要
    - 输出启用的功能开关
    - 输出配置的 LLM 提供商列表
    - 输出 Memory 系统状态
    - _需求: 1.9, 12.6_

- [x] 13. 端到端集成测试
  - [x] 13.1 编写完整分析流程的集成测试
    - 测试使用新 Orchestrator 的完整流程
    - 测试各种根因类型的分析准确性
    - 测试 LLM 调用和故障转移
    - 测试 Memory 系统的保存和检索
    - _需求: 10.6_

  - [x] 13.2 编写 API 兼容性测试
    - 使用现有的集成测试套件
    - 验证响应格式与旧实现一致
    - 验证新增字段不影响现有客户端
    - _需求: 7.5, 7.6_

  - [x] 13.3 编写性能基准测试
    - 测量 P50 和 P99 延迟
    - 对比新旧实现的性能
    - 测试并发场景(100 个并发请求)
    - _需求: 9.1, 9.2, 9.4, 9.7_

  - [x] 13.4 编写降级和错误场景测试
    - 测试 LLM 全部失败的降级
    - 测试 VectorStore 不可用的降级
    - 测试 Agent 调用失败的处理
    - _需求: 11.1, 11.2, 11.3, 11.4_

## Phase 6: 清理和文档

- [x] 14. 代码清理和优化
  - [x] 14.1 清理旧的 LLM 客户端实现(可选)
    - 保留旧实现作为 fallback
    - 添加废弃注释(Deprecated)
    - _需求: NFR-1_

  - [x] 14.2 代码优化和重构
    - 优化 LLM 调用频率,添加缓存
    - 优化 VectorStore 查询性能
    - 重构重复代码
    - _需求: NFR-1, NFR-2_

  - [x] 14.3 提高测试覆盖率
    - 补充缺失的单元测试
    - 确保代码覆盖率 ≥ 80%
    - _需求: 10.1_

- [x] 15. 更新文档
  - 更新 `README.md` 说明新功能
  - 更新 API 文档,描述新增字段
  - 创建配置指南,说明功能开关
  - 创建部署指南,说明依赖服务(Chroma)
  - 添加代码示例和使用教程
  - _需求: NFR-4_

- [x] 16. 部署准备
  - 更新 `Dockerfile` 包含新的依赖
  - 更新 K8s 配置文件
  - 配置 Prometheus 监控和告警
  - 创建部署检查清单
  - _需求: NFR-2, NFR-3_

## 验收标准

完成所有任务后,系统应满足以下验收标准:

- ✅ 所有单元测试通过,代码覆盖率 ≥ 80%
- ✅ 所有集成测试通过,包括现有测试套件
- ✅ P50 延迟 ≤ 当前实现,P99 延迟 ≤ 当前实现 * 1.2
- ✅ LLM 调用成功率 ≥ 95%
- ✅ 故障转移时间 < 1s
- ✅ 向量搜索延迟 < 100ms
- ✅ API 兼容性 100%,现有客户端无需修改
- ✅ 配置向后兼容,旧配置文件继续工作
- ✅ 功能开关正常工作,支持平滑迁移
- ✅ 文档完整,包括 API 文档、配置指南和部署指南

## 注意事项

1. **渐进式实施**: 每完成一个 Phase,进行充分测试后再继续下一个
2. **功能开关**: 默认关闭新功能,通过配置逐步启用
3. **向后兼容**: 确保旧实现继续工作,新旧可以共存
4. **测试驱动**: 先写测试,再写实现
5. **性能监控**: 持续监控性能指标,确保不降低性能
6. **文档同步**: 代码变更时同步更新文档
