# 需求文档: Reasoning Service 重构

## 简介

本需求文档定义了基于 go-llm-proxy (gollm) 和 LangChainGo 重构 Reasoning Service 的功能需求。重构目标是实现统一的 LLM 访问层、标准化的推理链、可扩展的 Agent 框架和智能内存管理,同时保持向后兼容性和系统稳定性。

## 需求

### 需求 1: LLM Proxy 适配器

**用户故事**: 作为开发者,我希望有一个统一的 LLM 访问接口,这样我就可以轻松切换不同的 LLM 提供商而无需修改业务代码。

#### 验收标准

1. WHEN 系统从配置文件加载 LLM 配置 THEN 系统应读取 `llm.enabled` 和 `llm.providers` 数组
2. WHEN 创建 ProxyAdapter THEN 系统应使用 LLMProviderConfig 包含 name、api_key、base_url、model、max_tokens、temperature、timeout 和 priority 字段
3. WHEN 系统初始化提供商 THEN 系统应按 priority 字段排序提供商 (数值越小优先级越高)
4. IF 提供商的 api_key 为空 THEN 系统应尝试从环境变量读取 (如 OPENAI_API_KEY、GEMINI_API_KEY、DEEPSEEK_API_KEY)
5. WHEN 调用 Complete() 方法 THEN 系统应通过 gollm 统一 API 发送请求到优先级最高的可用提供商
6. WHEN 主提供商调用失败 THEN 系统应按 priority 顺序自动故障转移到下一个提供商
7. WHEN 请求超时超过提供商配置的 timeout THEN 系统应取消请求并尝试下一个提供商
8. IF 启用了成本跟踪 THEN 系统应记录每次 LLM 调用的成本、token 使用量、使用的提供商和延迟
9. WHEN 系统启动时 THEN 应输出所有配置的提供商列表和它们的状态 (有 API key 或无 API key)

### 需求 2: LangChain 分析链集成

**用户故事**: 作为系统架构师,我希望使用 LangChainGo 实现可组合的推理链,这样我就可以灵活地调整分析步骤和添加新的分析能力。

#### 验收标准

1. WHEN 创建 AnalysisChain THEN 系统应包含 PreprocessChain、RuleBasedAnalysis、LLMAnalysisChain 和 AggregationChain 四个子链
2. WHEN 执行 AnalysisChain.Run() THEN 系统应按顺序执行预处理、规则匹配、LLM 分析和结果聚合步骤
3. WHEN 规则匹配的置信度 > 0.8 THEN 系统应直接返回规则匹配结果,跳过 LLM 分析步骤以提高性能
4. WHEN 执行 LLMAnalysisChain AND 提供了 hint 参数 THEN 系统应将规则匹配的结果作为上下文传递给 LLM
5. WHEN AggregationChain 聚合多个分析结果 THEN 系统应选择置信度最高的结果作为最终输出
6. IF AnalysisChain 执行失败 THEN 系统应返回包含错误信息的 AnalysisOutput
7. WHEN PreprocessChain 处理输入 THEN 系统应清洗和格式化数据,限制日志大小不超过配置的 MaxContextSize

### 需求 3: 建议生成链

**用户故事**: 作为 SRE 工程师,我希望系统能够生成高质量的问题解决建议,这样我就可以快速定位和修复 K8s 集群问题。

#### 验收标准

1. WHEN 创建 RecommendationChain THEN 系统应包含 LoadRulesChain、LLMEnhancementChain、RiskAssessmentChain 和 PrioritizationChain 四个子链
2. WHEN LoadRulesChain 执行 THEN 系统应从规则引擎加载针对特定根因类型的预定义建议模板
3. IF 配置启用 LLM 增强 AND 存在可用的 LLM 客户端 THEN LLMEnhancementChain 应使用 LLM 增强建议描述
4. WHEN RiskAssessmentChain 评估建议 THEN 系统应为每个建议标注风险级别 (low/medium/high)
5. WHEN PrioritizationChain 排序建议 THEN 系统应根据置信度和风险级别对建议进行排序
6. WHEN 建议数量超过配置的 MaxRecommendations THEN 系统应只返回前 N 个优先级最高的建议
7. IF LLM 增强失败 THEN 系统应降级使用规则引擎的原始建议

### 需求 4: K8s 工具 Agent

**用户故事**: 作为系统运维人员,我希望系统能够自动调用 kubectl 工具获取额外信息,这样系统就能提供更准确的根因分析。

#### 验收标准

1. WHEN 创建 K8sToolAgent THEN 系统应注册 KubectlDescribeTool、KubectlLogsTool、KubectlTopTool 和 MetricsQueryTool 工具
2. WHEN Agent 执行任务 THEN LLM 应根据任务描述决定需要调用哪些工具
3. WHEN 调用 KubectlDescribeTool THEN 系统应通过 K8s API 获取指定资源的详细信息
4. WHEN 调用 KubectlLogsTool THEN 系统应通过 K8s API 获取指定 Pod 的日志
5. WHEN 调用 KubectlTopTool THEN 系统应通过 K8s API 获取资源使用情况
6. IF K8s API 调用失败 THEN 系统应返回错误信息但不中断整个分析流程
7. WHEN Agent 执行完成 THEN 系统应返回包含工具输出、使用的工具列表和推理过程的 ExecutionResult

### 需求 5: 记忆管理系统

**用户故事**: 作为产品经理,我希望系统能够记住历史分析结果和相似案例,这样系统就能不断学习和提高分析准确率。

#### 验收标准

1. WHEN 创建 MemoryManager THEN 系统应初始化 ConversationMemory、VectorStore 和 StructuredMemory 三种存储
2. WHEN 保存分析结果 THEN 系统应同时保存到对话记忆、向量存储和结构化存储
3. WHEN 查找相似案例 THEN 系统应使用向量相似度搜索返回最相似的 5 个历史案例
4. WHEN 获取对话历史 THEN 系统应根据 sessionID 返回该会话的所有历史消息
5. IF VectorStore 未配置或不可用 THEN 系统应降级使用 StructuredMemory 进行相似案例检索
6. WHEN 生成 embedding THEN 系统应使用配置的 embedding 模型将文本转换为向量
7. WHEN 向量搜索延迟超过 100ms THEN 系统应记录性能警告日志

### 需求 6: 协调器 (Orchestrator)

**用户故事**: 作为系统集成者,我希望有一个协调器统一管理所有模块的调用流程,这样系统就能有序地执行分析、工具调用、建议生成和记忆保存。

#### 验收标准

1. WHEN Orchestrator 接收分析请求 THEN 系统应先从 MemoryManager 加载历史上下文和相似案例
2. WHEN 执行 AnalysisChain THEN 系统应将历史上下文和相似案例作为输入传递
3. IF AnalysisOutput 标记 NeedsMoreInfo = true THEN 系统应调用 K8sToolAgent 获取额外信息
4. WHEN K8sToolAgent 返回结果 THEN 系统应将工具输出合并到 AnalysisOutput
5. WHEN 分析完成 THEN 系统应调用 RecommendationChain 生成建议
6. IF RecommendationChain 失败 THEN 系统应降级使用规则引擎的预定义建议
7. WHEN 构建最终结果 THEN 系统应包含 root_cause、recommendations、confidence、evidence 和 similar_cases
8. WHEN 返回结果前 THEN 系统应调用 MemoryManager 保存分析结果

### 需求 7: API 兼容性

**用户故事**: 作为 API 使用者,我希望重构后的系统保持现有 API 不变,这样我就不需要修改现有的客户端代码。

#### 验收标准

1. WHEN 客户端调用 POST /api/v1/analyze/root-cause THEN 系统应返回与现有实现相同格式的 AnalysisResult
2. WHEN 客户端调用 POST /api/v1/analyze/k8s-event THEN 系统应返回与现有实现相同格式的响应
3. WHEN 客户端调用 GET /health THEN 系统应返回健康检查响应
4. IF 响应中添加新字段 (如 similar_cases、llm_provider) THEN 这些字段应是可选的,不影响现有客户端
5. WHEN 现有集成测试运行 THEN 所有测试应通过,无需修改测试代码
6. WHEN 请求格式不变 THEN 系统应兼容现有的请求格式和参数

### 需求 8: 配置管理

**用户故事**: 作为运维工程师,我希望通过配置文件灵活控制新旧功能的切换,这样我就可以安全地进行灰度发布和快速回滚。

#### 验收标准

1. WHEN 系统加载配置 THEN 应支持现有的配置结构 (server、llm、analysis、prediction、learning、performance、logging、features)
2. WHEN 在 features 配置块添加新的功能开关 THEN 系统应支持 use_new_orchestrator、use_llm_proxy、use_memory_system、use_tool_agent 等开关
3. WHEN features.use_new_orchestrator = true THEN 系统应使用新的 Orchestrator 实现
4. WHEN features.use_new_orchestrator = false THEN 系统应使用旧的 Analyzer 实现
5. WHEN features.use_llm_proxy = true THEN 系统应使用 ProxyAdapter 替代旧的 LLM 客户端
6. IF 旧配置文件不包含新的 features 开关 THEN 系统应使用默认值 (use_new_orchestrator=false 保证向后兼容)
7. WHEN 环境变量设置 (如 OPENAI_API_KEY) THEN 系统应覆盖配置文件中的 api_key 值
8. WHEN 配置文件包含无效值 THEN 系统应在启动时验证并返回清晰的错误信息

### 需求 9: 性能要求

**用户故事**: 作为系统管理员,我希望重构后的系统性能不低于现有实现,这样系统就能满足生产环境的性能要求。

#### 验收标准

1. WHEN 执行根因分析请求 THEN P50 延迟应 ≤ 当前实现的 P50 延迟
2. WHEN 执行根因分析请求 THEN P99 延迟应 ≤ 当前实现的 P99 延迟 * 1.2
3. WHEN 调用 LLM 提供商 THEN 成功率应 ≥ 95%
4. WHEN 主提供商失败需要故障转移 THEN 转移时间应 < 1s
5. WHEN 执行向量相似度搜索 THEN 搜索延迟应 < 100ms
6. IF 启用缓存 THEN 缓存命中的请求延迟应 < 50ms
7. WHEN 并发处理 100 个请求 THEN 系统应保持稳定,无明显性能下降

### 需求 10: 测试覆盖率

**用户故事**: 作为质量工程师,我希望代码有充分的测试覆盖,这样系统就能保证高质量和稳定性。

#### 验收标准

1. WHEN 运行单元测试 THEN 代码覆盖率应 ≥ 80%
2. WHEN 测试 LLM Proxy Adapter THEN 应包含成功调用、失败重试、故障转移和超时的测试用例
3. WHEN 测试 AnalysisChain THEN 应包含规则匹配、LLM 分析、结果聚合和错误处理的测试用例
4. WHEN 测试 K8sToolAgent THEN 应包含工具调用、Agent 推理和错误处理的测试用例
5. WHEN 测试 MemoryManager THEN 应包含保存、检索和相似度搜索的测试用例
6. WHEN 运行集成测试 THEN 应测试端到端的分析流程,包括所有模块的集成
7. IF 测试失败 THEN CI/CD 流程应阻止代码合并

### 需求 11: 错误处理和降级

**用户故事**: 作为系统可靠性工程师,我希望系统有完善的错误处理和降级机制,这样即使部分模块失败,系统也能继续提供服务。

#### 验收标准

1. WHEN LLM Proxy 调用失败 AND 所有备用提供商都不可用 THEN 系统应降级使用规则引擎
2. WHEN VectorStore 不可用 THEN 系统应降级使用 StructuredMemory 或跳过相似案例检索
3. WHEN K8sToolAgent 调用失败 THEN 系统应继续使用现有数据完成分析
4. WHEN RecommendationChain 失败 THEN 系统应返回规则引擎的预定义建议
5. IF 任何子系统失败 THEN 系统应记录详细的错误日志包含堆栈跟踪
6. WHEN 检测到错误 THEN 系统应返回包含错误信息的响应,而不是返回 500 错误
7. WHEN 错误率超过阈值 THEN 系统应发送告警通知

### 需求 12: 监控和可观测性

**用户故事**: 作为运维工程师,我希望系统提供详细的监控指标和日志,这样我就可以及时发现和诊断问题。

#### 验收标准

1. WHEN 系统处理请求 THEN 应记录请求 ID、处理时间、使用的 LLM 提供商和成本
2. WHEN LLM 调用发生 THEN 应记录提供商、模型、token 使用量、延迟和成本
3. WHEN 故障转移发生 THEN 应记录原提供商、失败原因和目标提供商
4. WHEN 执行 Chain 或 Agent THEN 应记录执行的步骤和中间结果
5. IF 启用调试模式 THEN 系统应记录详细的请求和响应内容
6. WHEN 系统启动 THEN 应输出配置摘要,包括启用的功能和提供商列表
7. WHEN 导出指标 THEN 应提供 Prometheus 格式的指标端点

## 非功能性需求

### NFR-1: 可维护性

系统应使用模块化架构,新增 LLM 提供商或分析能力时只需添加新模块,无需修改现有代码。

### NFR-2: 可扩展性

系统应支持水平扩展,多个实例可以共享 VectorStore 和 StructuredMemory。

### NFR-3: 安全性

系统应安全存储 API 密钥,不在日志中输出敏感信息,支持通过环境变量注入密钥。

### NFR-4: 文档完整性

系统应提供完整的 API 文档、架构文档、部署文档和示例代码。
