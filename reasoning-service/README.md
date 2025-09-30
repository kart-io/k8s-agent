# Aetherius Reasoning Service

AI 推理服务 (Layer 4),提供根因分析、故障预测和智能推荐。

---

## 功能特性

### 核心功能

- **根因分析**: 多模态分析 (事件、日志、指标) 识别故障根本原因
- **智能推荐**: 基于规则的修复建议,包含步骤、风险评估和回滚方案
- **故障预测**: 基于趋势和异常检测预测潜在故障
- **知识图谱**: 存储历史案例,提供相似案例检索
- **持续学习**: 从用户反馈中学习,持续改进分析准确性

### 支持的根因类型

- **OOMKiller**: 内存溢出
- **CPUThrottling**: CPU 限流
- **DiskPressure**: 磁盘空间不足
- **NetworkError**: 网络连接问题
- **ConfigError**: 配置错误
- **ImagePullError**: 镜像拉取失败
- **VolumeError**: 存储卷挂载失败
- **ResourceLimit**: 资源配额限制

---

## 架构设计

```plaintext
┌──────────────────────────────────────────────────────────────┐
│             Reasoning Service (Layer 4)                       │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌────────────────┐  ┌──────────────┐  ┌────────────────┐  │
│  │ Root Cause     │  │ Recommender  │  │ Predictor      │  │
│  │ Analyzer       │  │ Engine       │  │ Engine         │  │
│  │                │  │              │  │                │  │
│  │ - Events       │  │ - Rules      │  │ - Trends       │  │
│  │ - Logs         │  │ - Steps      │  │ - Anomalies    │  │
│  │ - Metrics      │  │ - Risk       │  │ - Thresholds   │  │
│  └────────────────┘  └──────────────┘  └────────────────┘  │
│                                                               │
│  ┌────────────────┐  ┌──────────────┐                       │
│  │ Knowledge      │  │ Learning     │                       │
│  │ Graph          │  │ System       │                       │
│  │                │  │              │                       │
│  │ - Cases        │  │ - Feedback   │                       │
│  │ - Patterns     │  │ - Metrics    │                       │
│  │ - Similarity   │  │ - Improve    │                       │
│  └────────────────┘  └──────────────┘                       │
│                                                               │
│  ↑ HTTP API (FastAPI)                                        │
└───┼──────────────────────────────────────────────────────────┘
    │
    │ orchestrator-service calls
```

---

## 快速开始

### 前置要求

- Python 3.11+
- Neo4j 5+ (可选,用于知识图谱)
- pip

### 本地运行

```bash
# 1. 安装依赖
make install
# 或者
pip install -r requirements.txt

# 2. 启动 Neo4j (可选)
docker run -d \
  --name neo4j \
  -p 7687:7687 -p 7474:7474 \
  -e NEO4J_AUTH=neo4j/password \
  neo4j:latest

# 3. 配置文件
cp configs/config.yaml configs/config.local.yaml
# 编辑 config.local.yaml 设置 Neo4j 连接信息

# 4. 运行服务
make run
# 或者
python cmd/server/main.py --config=configs/config.local.yaml
```

### 开发模式

```bash
# 使用 auto-reload
make dev
```

### Docker 运行

```bash
# 构建镜像
make docker-build

# 运行容器
make docker-run
```

---

## API 端点

### 健康检查

```bash
GET /health
```

**响应**:

```json
{
  "status": "healthy",
  "service": "reasoning-service",
  "components": {
    "analyzer": true,
    "recommender": true,
    "knowledge_graph": true,
    "predictor": true,
    "learning_system": true
  }
}
```

### 根因分析

```bash
POST /api/v1/analyze/root-cause
```

**请求**:

```json
{
  "request_id": "req-123",
  "analysis_type": "root_cause",
  "context": {
    "event": {
      "reason": "OOMKilled",
      "message": "Container killed due to OOM"
    },
    "logs": "fatal error: runtime: out of memory\n...",
    "metrics": {
      "memory": {
        "usage_percent": 98
      }
    }
  },
  "options": {
    "min_confidence": 0.7,
    "include_similar_cases": true,
    "max_recommendations": 5
  }
}
```

**响应**:

```json
{
  "request_id": "req-123",
  "status": "completed",
  "result": {
    "root_cause": {
      "type": "OOMKiller",
      "description": "Container was killed due to out of memory (OOM)",
      "confidence": 0.95,
      "evidence": [
        "Event reason: OOMKilled",
        "Found pattern: OOM indicator (2 occurrences)"
      ]
    },
    "recommendations": [
      {
        "action": "increase_memory_limit",
        "description": "Increase container memory limits to prevent OOM kills",
        "confidence": 0.90,
        "risk": "low",
        "impact": "Prevents future OOM kills, may increase cluster resource usage",
        "steps": [
          "Analyze current memory usage patterns",
          "Calculate recommended memory limit (current + 50%)",
          "Update Deployment/StatefulSet memory limits",
          "kubectl apply -f updated-manifest.yaml",
          "Monitor for OOM recurrence"
        ],
        "rollback_steps": [
          "Revert to previous memory limits",
          "kubectl rollout undo deployment/<name>"
        ],
        "estimated_duration": "5 minutes"
      }
    ],
    "confidence": 0.95,
    "evidence": [
      "Event reason: OOMKilled",
      "Memory usage at 98%"
    ],
    "similar_cases": []
  },
  "processing_time": 0.123
}
```

### 故障预测

```bash
POST /api/v1/analyze/predict
```

**请求**:

```json
{
  "cluster_id": "prod-cluster",
  "resource_type": "pod",
  "resource_name": "my-app-xyz",
  "metrics": {
    "memory": {
      "usage_percent": 85
    },
    "cpu": {
      "usage_percent": 75,
      "throttling_percent": 60
    },
    "restart_count": 3,
    "history": [
      {
        "timestamp": "2024-01-01T10:00:00Z",
        "memory": {"usage_percent": 70},
        "cpu": {"usage_percent": 65}
      }
    ]
  },
  "time_window": "24h"
}
```

**响应**:

```json
{
  "failure_probability": 0.75,
  "predicted_failure_time": "2024-01-01T16:00:00Z",
  "failure_types": ["OOMKiller", "CPUThrottling"],
  "confidence": 0.8,
  "contributing_factors": [
    "Memory usage approaching limit (85%)",
    "CPU throttling at 60%",
    "Pod restarted 3 times"
  ]
}
```

### 提交反馈

```bash
POST /api/v1/feedback
```

**请求**:

```json
{
  "feedback_id": "fb-123",
  "request_id": "req-123",
  "feedback_type": "diagnosis_accuracy",
  "rating": 5,
  "was_helpful": true,
  "actual_root_cause": "OOMKiller",
  "comments": "Diagnosis was accurate and recommendations worked",
  "submitted_by": "admin"
}
```

### 添加案例

```bash
POST /api/v1/cases
```

**请求**:

```json
{
  "id": "case-123",
  "title": "OOM in production API",
  "description": "API pods experienced OOM kills during traffic spike",
  "symptoms": ["OOMKilled", "high memory usage", "slow response"],
  "root_cause": "OOMKiller",
  "solution": "Increased memory limits from 512Mi to 1Gi",
  "outcome": "No more OOM kills after increase",
  "cluster_id": "prod-cluster"
}
```

### 查找相似案例

```bash
GET /api/v1/cases/similar?event_reason=OOMKilled&limit=5
```

### 准确性指标

```bash
GET /api/v1/metrics/accuracy
GET /api/v1/metrics/accuracy?root_cause_type=OOMKiller
```

**响应**:

```json
{
  "overall": 0.87,
  "by_root_cause": {
    "OOMKiller": {
      "total_diagnoses": 50,
      "correct_diagnoses": 47,
      "accuracy": 0.94,
      "last_updated": "2024-01-01T12:00:00Z"
    }
  }
}
```

### 改进建议

```bash
GET /api/v1/metrics/suggestions
```

---

## 分析流程

### 根因分析流程

```plaintext
1. 接收分析请求
    ↓
2. 多模态分析
    ├─> 事件分析 (Event reason mapping)
    ├─> 日志分析 (Pattern matching + Keywords)
    ├─> 指标分析 (Threshold detection)
    └─> 关联分析 (Cross-validation)
    ↓
3. 选择最佳分析结果
    ↓
4. 生成修复建议
    ↓
5. 查找相似案例
    ↓
6. 返回结果
```

### 推荐生成流程

```plaintext
1. 获取根因类型
    ↓
2. 匹配推荐规则
    ↓
3. 检查条件 (事件、指标等)
    ↓
4. 生成推荐列表
    ↓
5. 按置信度 × 风险权重排序
    ↓
6. 返回 Top N 推荐
```

### 预测流程

```plaintext
1. 接收指标数据
    ↓
2. 多种预测方法
    ├─> 阈值检测
    ├─> 趋势分析
    └─> 异常检测 (Isolation Forest)
    ↓
3. 聚合预测结果
    ↓
4. 计算故障概率和时间
    ↓
5. 返回预测结果
```

---

## 配置说明

### 关键配置项

```yaml
# Neo4j 连接
neo4j:
  uri: "bolt://localhost:7687"
  user: "neo4j"
  password: "password"

# 分析设置
analysis:
  min_confidence: 0.7
  max_recommendations: 5
  include_similar_cases: true

# 预测设置
prediction:
  anomaly_detection:
    contamination: 0.1  # 异常比例

# 学习设置
learning:
  enable_feedback: true
  min_samples_for_accuracy: 5
```

---

## 集成示例

### 从 orchestrator-service 调用

```python
import httpx

# 根因分析
async def analyze_root_cause(context):
    async with httpx.AsyncClient() as client:
        response = await client.post(
            "http://reasoning-service:8082/api/v1/analyze/root-cause",
            json={
                "request_id": "req-123",
                "analysis_type": "root_cause",
                "context": context
            }
        )
        return response.json()

# 故障预测
async def predict_failure(metrics):
    async with httpx.AsyncClient() as client:
        response = await client.post(
            "http://reasoning-service:8082/api/v1/analyze/predict",
            json={
                "cluster_id": "prod",
                "resource_type": "pod",
                "resource_name": "my-app",
                "metrics": metrics
            }
        )
        return response.json()
```

### 从 Go 服务调用

```go
type AnalysisRequest struct {
    RequestID    string                 `json:"request_id"`
    AnalysisType string                 `json:"analysis_type"`
    Context      map[string]interface{} `json:"context"`
}

func analyzeRootCause(ctx context.Context, analysisCtx map[string]interface{}) (*AnalysisResponse, error) {
    req := AnalysisRequest{
        RequestID:    generateRequestID(),
        AnalysisType: "root_cause",
        Context:      analysisCtx,
    }

    body, _ := json.Marshal(req)
    resp, err := http.Post(
        "http://reasoning-service:8082/api/v1/analyze/root-cause",
        "application/json",
        bytes.NewBuffer(body),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result AnalysisResponse
    json.NewDecoder(resp.Body).Decode(&result)
    return &result, nil
}
```

---

## 开发指南

### 项目结构

```plaintext
reasoning-service/
├── cmd/
│   └── server/
│       └── main.py           # 主程序入口
├── internal/
│   ├── analyzer/
│   │   └── root_cause.py     # 根因分析器
│   ├── recommender/
│   │   └── engine.py         # 推荐引擎
│   ├── predictor/
│   │   └── engine.py         # 预测引擎
│   ├── knowledge/
│   │   └── graph.py          # 知识图谱
│   ├── learning/
│   │   └── system.py         # 学习系统
│   └── api/
│       └── server.py         # FastAPI 服务器
├── pkg/
│   └── types.py              # 类型定义
├── configs/
│   └── config.yaml           # 配置文件
├── tests/                    # 测试
├── requirements.txt          # Python 依赖
├── Dockerfile
├── Makefile
└── README.md
```

### 添加新的根因类型

1. 在 `pkg/types.py` 添加枚举:

```python
class RootCauseType(str, Enum):
    NEW_TYPE = "NewType"
```

2. 在 `root_cause.py` 添加检测逻辑:

```python
# 在 _analyze_logs 或 _analyze_metrics 中添加检测模式
```

3. 在 `engine.py` 添加推荐规则:

```python
RootCauseType.NEW_TYPE: [
    {
        "action": "fix_new_issue",
        "description": "...",
        "confidence": 0.9,
        "risk": "low",
        "steps": [...]
    }
]
```

### 添加新的分析方法

```python
# 在 RootCauseAnalyzer 类中
def _analyze_custom(self, data: Dict) -> Optional[Tuple]:
    # 实现自定义分析逻辑
    return (root_cause, confidence, evidence)

# 在 analyze 方法中调用
if custom_data:
    custom_analysis = self._analyze_custom(custom_data)
    if custom_analysis:
        analyses.append(custom_analysis)
```

---

## 测试

### 运行测试

```bash
make test
```

### 示例测试

```python
import pytest
from internal.analyzer.root_cause import RootCauseAnalyzer
from pkg.types import AnalysisContext

def test_oom_detection():
    analyzer = RootCauseAnalyzer()
    context = AnalysisContext(
        event={"reason": "OOMKilled"},
        logs="fatal error: out of memory"
    )
    result = analyzer.analyze(context)
    assert result.root_cause.type == "OOMKiller"
    assert result.confidence >= 0.9
```

---

## 监控和调试

### 日志

```bash
# 查看实时日志
tail -f logs/reasoning-service.log

# 查看分析日志
grep "Analyzing root cause" logs/reasoning-service.log

# 查看错误
grep "ERROR" logs/reasoning-service.log
```

### 健康检查

```bash
curl http://localhost:8082/health
```

### 性能监控

- 使用 FastAPI 内置的 `/docs` 查看 API 文档和测试接口
- 监控响应时间和准确率
- 定期查看学习系统的准确性指标

---

## 故障排查

### 问题 1: Neo4j 连接失败

**检查**:

- Neo4j 是否运行
- 连接配置是否正确
- 网络连通性

**解决**: 服务会自动降级到内存存储模式

### 问题 2: 分析准确率低

**检查**:

- 查看准确性指标: `GET /api/v1/metrics/accuracy`
- 查看改进建议: `GET /api/v1/metrics/suggestions`

**解决**:

- 添加更多历史案例
- 调整检测模式和权重
- 收集用户反馈

### 问题 3: 预测不准确

**检查**:

- 指标数据是否完整
- 历史数据是否足够

**解决**:

- 使用更长的时间窗口
- 训练异常检测模型
- 调整阈值

---

## 路线图

- [ ] 集成大语言模型 (LLM) 进行日志分析
- [ ] 深度学习模型训练
- [ ] 更多根因检测模式
- [ ] 自动化修复执行
- [ ] 多集群分析
- [ ] 实时流式分析
- [ ] Web UI 界面

---

## 许可证

MIT License

---

## 相关文档

- [系统架构](../../docs/architecture/SYSTEM_ARCHITECTURE.md)
- [orchestrator-service](../orchestrator-service/README.md)
- [agent-manager](../agent-manager/README.md)