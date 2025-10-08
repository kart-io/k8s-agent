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

## 模型管理

### 当前模型架构

Reasoning Service 当前采用**基于规则和机器学习混合**的方法,不依赖深度神经网络模型:

#### 1. 根因分析模型

**分析方法**:
- **规则匹配**: 基于事件 reason、日志关键词、指标阈值的规则引擎
- **模式识别**: 正则表达式和字符串匹配
- **多模态融合**: 综合事件、日志、指标的加权投票

**无需训练**: 规则和模式基于 Kubernetes 最佳实践和故障模式库

**核心组件** (`internal/analyzer/root_cause.py`):
```python
class RootCauseAnalyzer:
    def analyze(self, context: AnalysisContext) -> AnalysisResult:
        # 事件分析
        event_analysis = self._analyze_event(context.event)

        # 日志分析 (关键词 + 正则)
        log_analysis = self._analyze_logs(context.logs)

        # 指标分析 (阈值检测)
        metric_analysis = self._analyze_metrics(context.metrics)

        # 选择最佳结果 (最高置信度)
        return self._select_best_analysis([
            event_analysis,
            log_analysis,
            metric_analysis
        ])
```

**根因检测规则**:
```python
# 事件映射
EVENT_REASON_MAPPING = {
    "OOMKilled": RootCauseType.OOMKiller,
    "CrashLoopBackOff": None,  # 需要进一步分析
    "ImagePullBackOff": RootCauseType.ImagePullError,
    # ... 更多映射
}

# 日志模式 (正则表达式)
LOG_PATTERNS = {
    RootCauseType.OOMKiller: [
        r"out of memory",
        r"OOM killer",
        r"fatal error: runtime: out of memory"
    ],
    RootCauseType.NetworkError: [
        r"connection refused",
        r"dial tcp.*timeout",
        r"no route to host"
    ],
    # ... 更多模式
}

# 指标阈值
METRIC_THRESHOLDS = {
    "memory_usage_percent": {
        "critical": 95,
        "warning": 85
    },
    "cpu_throttling_percent": {
        "critical": 80,
        "warning": 60
    }
}
```

#### 2. 故障预测模型

**预测方法**:
- **阈值预测**: 基于当前指标判断是否接近临界值
- **趋势分析**: 简单线性趋势 (增长速度)
- **异常检测**: Isolation Forest (sklearn)

**Isolation Forest 模型**:

```python
from sklearn.ensemble import IsolationForest

class PredictionEngine:
    def __init__(self):
        # Isolation Forest 用于异常检测
        self.anomaly_detector = IsolationForest(
            contamination=0.1,      # 假设 10% 数据为异常
            random_state=42,
            n_estimators=100
        )
        self.is_trained = False

    def predict_failure(self, request: PredictionRequest) -> PredictionResult:
        # 1. 阈值预测
        threshold_prediction = self._predict_by_threshold(request.metrics)

        # 2. 趋势预测
        trend_prediction = self._predict_by_trend(request.metrics)

        # 3. 异常检测 (如果有历史数据)
        anomaly_prediction = None
        if self.is_trained and request.metrics.history:
            anomaly_prediction = self._predict_by_anomaly(request.metrics)

        # 4. 聚合预测结果
        return self._aggregate_predictions([
            threshold_prediction,
            trend_prediction,
            anomaly_prediction
        ])
```

**模型特点**:
- **无监督学习**: Isolation Forest 不需要标注数据
- **增量训练**: 可以使用新数据持续更新
- **轻量级**: 不需要 GPU,CPU 即可运行

#### 3. 推荐引擎

**推荐方法**:
- **基于规则**: 预定义的修复动作规则库
- **条件匹配**: 根据根因类型和上下文条件选择推荐

**规则库结构** (`internal/recommender/engine.py`):
```python
RECOMMENDATION_RULES = {
    RootCauseType.OOMKiller: [
        {
            "action": "increase_memory_limit",
            "description": "增加内存限制防止 OOM",
            "confidence": 0.90,
            "risk": "low",
            "conditions": {
                "memory_usage_percent": {"gt": 85}  # 仅当使用率 > 85% 时推荐
            },
            "steps": [
                "分析当前内存使用模式",
                "计算推荐内存限制 (当前 + 50%)",
                "更新 Deployment/StatefulSet 内存限制",
                "kubectl apply -f updated-manifest.yaml",
                "监控 OOM 是否再次发生"
            ],
            "rollback_steps": [
                "恢复到之前的内存限制",
                "kubectl rollout undo deployment/<name>"
            ],
            "estimated_duration": "5 minutes"
        },
        {
            "action": "investigate_memory_leak",
            "description": "调查内存泄漏",
            "confidence": 0.70,
            "risk": "none",
            "steps": [
                "获取内存 profile",
                "分析内存泄漏点",
                "修复应用程序代码"
            ]
        }
    ],
    # ... 其他根因类型的推荐规则
}
```

#### 4. 知识图谱 (Neo4j)

**用途**:
- 存储历史案例
- 相似案例检索 (基于向量相似度或图匹配)
- 案例关联分析

**图结构**:
```cypher
# 案例节点
(:Case {
  id: "case-123",
  title: "OOM in production",
  root_cause: "OOMKiller",
  symptoms: ["OOMKilled", "high memory"],
  solution: "Increased memory to 1Gi",
  cluster_id: "prod"
})

# 症状节点
(:Symptom {name: "OOMKilled"})

# 关系
(:Case)-[:HAS_SYMPTOM]->(:Symptom)
(:Case)-[:SIMILAR_TO {score: 0.85}]->(:Case)
```

**相似度计算**:
```python
def find_similar_cases(self, context: AnalysisContext, limit: int = 5) -> List[CaseStudy]:
    # 基于症状的 Jaccard 相似度
    query = """
    MATCH (c:Case)
    WHERE c.root_cause = $root_cause
    WITH c,
         size([s IN c.symptoms WHERE s IN $symptoms]) AS common,
         size(c.symptoms) + size($symptoms) AS total
    RETURN c,
           1.0 * common / (total - common) AS similarity
    ORDER BY similarity DESC
    LIMIT $limit
    """
```

### 模型训练与更新

#### 1. Isolation Forest 训练

**初始训练**:
```python
# 使用历史指标数据训练
def train_anomaly_detector(self, historical_data: List[MetricSnapshot]):
    """
    Args:
        historical_data: 历史指标快照列表
            [
                {"memory_usage": 70, "cpu_usage": 50, "restart_count": 0},
                {"memory_usage": 85, "cpu_usage": 75, "restart_count": 1},
                ...
            ]
    """
    if len(historical_data) < 20:
        logger.warning("Insufficient data for training (<20 samples)")
        return False

    # 准备特征矩阵
    X = np.array([[
        d["memory_usage"],
        d["cpu_usage"],
        d["restart_count"]
    ] for d in historical_data])

    # 训练模型
    self.anomaly_detector.fit(X)
    self.is_trained = True

    logger.info(f"Anomaly detector trained with {len(historical_data)} samples")
    return True
```

**增量更新**:
```python
# 定期重新训练
async def update_models_periodically():
    while True:
        await asyncio.sleep(3600)  # 每小时

        # 从知识图谱获取最近 7 天的数据
        recent_data = knowledge_graph.get_recent_metrics(days=7)

        if len(recent_data) >= 20:
            predictor.train_anomaly_detector(recent_data)
```

#### 2. 规则更新

**手动更新**:
1. 编辑 `internal/analyzer/root_cause.py` 中的规则
2. 重启服务或热加载 (规划中)

**示例 - 添加新的日志模式**:
```python
# 在 LOG_PATTERNS 中添加新模式
LOG_PATTERNS = {
    RootCauseType.OOMKiller: [
        r"out of memory",
        r"OOM killer",
        r"fatal error: runtime: out of memory",
        r"java\.lang\.OutOfMemoryError"  # 新增 Java OOM 模式
    ],
}
```

**基于反馈的自动调整** (规划中):
```python
# 学习系统根据反馈自动调整规则权重
def adjust_rule_weights(self, feedback: List[Feedback]):
    for fb in feedback:
        if not fb.was_helpful:
            # 降低错误规则的权重
            rule = self.get_rule(fb.root_cause_predicted)
            rule.confidence *= 0.9
```

#### 3. 知识图谱更新

**添加案例**:
```bash
# API 调用添加新案例
POST /api/v1/cases
{
  "id": "case-new",
  "title": "新的故障案例",
  "root_cause": "NetworkError",
  "symptoms": ["connection timeout", "dns resolution failed"],
  "solution": "Updated DNS configuration",
  "cluster_id": "prod"
}
```

**批量导入**:
```python
# 从 CSV 或 JSON 批量导入案例
def import_cases_from_file(filepath: str):
    with open(filepath) as f:
        cases = json.load(f)

    for case_data in cases:
        case = CaseStudy(**case_data)
        knowledge_graph.add_case_study(case)

    logger.info(f"Imported {len(cases)} cases")
```

### 模型性能监控

#### 1. 准确率追踪

```bash
# 查看整体准确率
GET /api/v1/metrics/accuracy

# 响应
{
  "overall": 0.87,
  "by_root_cause": {
    "OOMKiller": {
      "total_diagnoses": 50,
      "correct_diagnoses": 47,
      "accuracy": 0.94
    },
    "NetworkError": {
      "total_diagnoses": 30,
      "correct_diagnoses": 24,
      "accuracy": 0.80
    }
  }
}
```

#### 2. 改进建议

```bash
# 获取改进建议
GET /api/v1/metrics/suggestions

# 响应
{
  "suggestions": [
    {
      "type": "low_accuracy",
      "root_cause": "NetworkError",
      "current_accuracy": 0.80,
      "message": "NetworkError 诊断准确率较低 (80%), 建议添加更多检测模式或案例"
    },
    {
      "type": "insufficient_samples",
      "root_cause": "DiskPressure",
      "sample_count": 3,
      "message": "DiskPressure 样本数不足 (3), 建议收集更多反馈数据"
    }
  ]
}
```

#### 3. 响应时间监控

```python
# 在 server.py 中自动记录
@app.middleware("http")
async def log_performance(request: Request, call_next):
    start_time = time.time()
    response = await call_next(request)
    duration = time.time() - start_time

    logger.info(
        f"{request.method} {request.url.path}",
        extra={"duration": duration}
    )
    return response
```

### 模型导出与备份

#### 1. Isolation Forest 模型导出

```python
import joblib

# 导出模型
def export_model(filepath: str = "models/isolation_forest.pkl"):
    if not self.is_trained:
        logger.warning("Model not trained, nothing to export")
        return False

    joblib.dump(self.anomaly_detector, filepath)
    logger.info(f"Model exported to {filepath}")
    return True

# 加载模型
def load_model(filepath: str = "models/isolation_forest.pkl"):
    try:
        self.anomaly_detector = joblib.load(filepath)
        self.is_trained = True
        logger.info(f"Model loaded from {filepath}")
        return True
    except Exception as e:
        logger.error(f"Failed to load model: {e}")
        return False
```

#### 2. 知识图谱备份

```bash
# 使用 Neo4j dump
neo4j-admin database dump neo4j --to=/backups/neo4j-$(date +%Y%m%d).dump

# 恢复
neo4j-admin database load neo4j --from=/backups/neo4j-20251001.dump
```

#### 3. 规则配置版本控制

```bash
# 将规则配置提交到 Git
git add internal/analyzer/root_cause.py
git add internal/recommender/engine.py
git commit -m "Update detection rules for NetworkError"
git push
```

### 未来模型规划

#### 1. 大语言模型 (LLM) 集成

**用途**:
- 日志分析和总结
- 自然语言描述生成
- 智能问答

**架构**:
```plaintext
┌─────────────────────────────────────────┐
│  Reasoning Service                       │
│                                          │
│  ┌────────────────┐   ┌──────────────┐ │
│  │ Rule-based     │   │ LLM Engine   │ │
│  │ Analyzer       │   │              │ │
│  │                │   │ - Log Parse  │ │
│  │ (Current)      │   │ - Summary    │ │
│  └────────────────┘   │ - NLG        │ │
│          │             └──────────────┘ │
│          │                     │         │
│          └─────────┬───────────┘         │
│                    │                     │
│              ┌─────▼──────┐             │
│              │  Ensemble  │             │
│              │  Fusion    │             │
│              └────────────┘             │
└─────────────────────────────────────────┘
```

**集成方案**:
- OpenAI API
- 本地 LLaMA/Mistral
- Azure OpenAI Service

#### 2. 深度学习模型

**用途**:
- 时序预测 (LSTM/Transformer)
- 多维异常检测
- 日志序列分析

**模型选择**:
- **时序预测**: Prophet, LSTM
- **异常检测**: Autoencoder, VAE
- **日志分析**: BERT, RoBERTa

#### 3. 联邦学习

**多集群学习**:
- 各集群本地训练
- 中央服务器聚合模型
- 保护数据隐私

### 模型管理最佳实践

#### 1. 数据收集

**收集关键数据**:
- 所有分析请求的输入和输出
- 用户反馈 (准确率、有用性)
- 故障案例和解决方案
- 指标时序数据

**数据存储**:
```yaml
# PostgreSQL 或 InfluxDB
analysis_requests:
  - request_id
  - timestamp
  - context (JSON)
  - result (JSON)
  - feedback_rating

metric_history:
  - timestamp
  - resource_id
  - metric_name
  - value
```

#### 2. 模型版本控制

**版本命名**:
```plaintext
models/
├── isolation_forest_v1.0.0.pkl
├── isolation_forest_v1.1.0.pkl
└── metadata.json
```

**元数据记录**:
```json
{
  "model_name": "isolation_forest",
  "version": "1.1.0",
  "trained_at": "2025-10-01T10:00:00Z",
  "training_samples": 5000,
  "features": ["memory_usage", "cpu_usage", "restart_count"],
  "hyperparameters": {
    "contamination": 0.1,
    "n_estimators": 100
  },
  "performance": {
    "accuracy": 0.87,
    "precision": 0.85,
    "recall": 0.90
  }
}
```

#### 3. A/B 测试

**对比测试新规则**:
```python
# 在部分请求上使用新规则
if request_id % 10 < 2:  # 20% 流量
    result = new_analyzer.analyze(context)
else:
    result = current_analyzer.analyze(context)

# 记录结果用于对比
log_ab_test_result(request_id, "new" if use_new else "current", result)
```

#### 4. 模型监控告警

**设置告警**:
```yaml
# Prometheus 告警规则
groups:
  - name: reasoning_service
    rules:
      - alert: LowAccuracy
        expr: reasoning_accuracy < 0.7
        for: 1h
        annotations:
          summary: "分析准确率过低 ({{ $value }})"

      - alert: HighLatency
        expr: reasoning_latency_p95 > 5
        for: 5m
        annotations:
          summary: "分析延迟过高 ({{ $value }}s)"
```

---

## 路线图

- [ ] 集成大语言模型 (LLM) 进行日志分析
- [ ] 深度学习模型训练 (LSTM, Transformer)
- [ ] 更多根因检测模式
- [ ] 自动化修复执行
- [ ] 多集群联邦学习
- [ ] 实时流式分析
- [ ] 模型 A/B 测试框架
- [ ] Web UI 界面
- [ ] 模型可解释性增强

---

## 许可证

MIT License

---

## 相关文档

- [系统架构](../../docs/architecture/SYSTEM_ARCHITECTURE.md)
- [orchestrator-service](../orchestrator-service/README.md)
- [agent-manager](../agent-manager/README.md)