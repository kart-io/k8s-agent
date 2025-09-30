# Aetherius 配置示例

本目录包含 Aetherius 各个组件的完整配置示例。

---

## 目录结构

```text
configs/
├── agent-manager/          # Agent Manager 配置
│   └── config.yaml
├── orchestrator-service/   # Orchestrator Service 配置
│   └── config.yaml
├── reasoning-service/      # Reasoning Service 配置
│   └── config.yaml
├── collect-agent/          # Collect Agent 配置
│   └── config.yaml
└── README.md               # 本文件
```

---

## 配置文件说明

### Agent Manager

**文件**: `agent-manager/config.yaml`

**关键配置项**:

- **服务端口**: 8080
- **数据库**: PostgreSQL (agent_manager)
- **Redis**: 数据库 0
- **NATS**: 连接中央消息总线
- **Agent 管理**: 注册、心跳、命令执行
- **事件处理**: 缓冲、批处理、过滤

**用途**:
- Agent 注册和管理
- 事件收集和分发
- 命令分发和结果收集

### Orchestrator Service

**文件**: `orchestrator-service/config.yaml`

**关键配置项**:

- **服务端口**: 8081
- **数据库**: PostgreSQL (orchestrator)
- **Redis**: 数据库 1
- **工作流引擎**: 定义加载、执行管理
- **步骤执行器**: 诊断、AI 分析、修复、通知
- **策略引擎**: 自动化策略评估

**用途**:
- 工作流编排和执行
- 自动化故障诊断
- 修复任务协调

### Reasoning Service

**文件**: `reasoning-service/config.yaml`

**关键配置项**:

- **服务端口**: 8082
- **根因分析器**: 多模态分析、置信度阈值
- **推荐引擎**: 风险评估、步骤生成
- **知识图谱**: Neo4j 连接、相似度计算
- **故障预测**: 阈值、趋势、异常检测
- **学习系统**: 反馈处理、模型更新

**用途**:
- 根因分析
- 故障预测
- 修复建议生成
- 持续学习

### Collect Agent

**文件**: `collect-agent/config.yaml`

**关键配置项**:

- **集群配置**: 集群 ID、名称、标签
- **事件监控**: 过滤规则、增强配置
- **指标收集**: Pod/Node/Container 指标
- **日志收集**: 按需或连续收集
- **NATS**: 连接中央消息总线
- **命令执行**: 工具配置、超时设置

**用途**:
- Kubernetes 事件监控
- 指标和日志收集
- 边缘命令执行

---

## 使用方法

### 1. 复制配置文件

```bash
# 复制到服务目录
cp examples/configs/agent-manager/config.yaml agent-manager/config.yaml
cp examples/configs/orchestrator-service/config.yaml orchestrator-service/config.yaml
cp examples/configs/reasoning-service/config.yaml reasoning-service/config.yaml
cp examples/configs/collect-agent/config.yaml collect-agent/config.yaml
```

### 2. 根据环境修改配置

#### 开发环境

```yaml
# 数据库配置
database:
  host: localhost
  port: 5432

# 日志配置
logging:
  level: debug
  format: text

# 特性开关
features:
  experimental: true
```

#### 生产环境

```yaml
# 数据库配置
database:
  host: postgres.production.svc.cluster.local
  port: 5432
  max_open_conns: 100

# 日志配置
logging:
  level: info
  format: json

# 安全配置
security:
  tls:
    enabled: true
  auth:
    enabled: true

# 特性开关
features:
  experimental: false
```

### 3. 使用环境变量覆盖

大多数配置项都支持通过环境变量覆盖:

```bash
# Agent Manager
export SERVER_PORT=8080
export DATABASE_HOST=postgres
export DATABASE_PASSWORD=your-password
export REDIS_PASSWORD=your-redis-password
export NATS_URL=nats://nats:4222

# Orchestrator Service
export SERVER_PORT=8081
export DATABASE_DATABASE=orchestrator
export WORKFLOW_MAX_CONCURRENT_EXECUTIONS=100

# Reasoning Service
export SERVER_PORT=8082
export KNOWLEDGE_GRAPH_BACKEND=neo4j
export NEO4J_URI=bolt://neo4j:7687
export NEO4J_PASSWORD=your-neo4j-password

# Collect Agent
export AGENT_CLUSTER_NAME=my-cluster
export NATS_URL=nats://nats.aetherius.svc.cluster.local:4222
```

### 4. 使用 Kubernetes ConfigMap

```bash
# 创建 ConfigMap
kubectl create configmap agent-manager-config \
  --from-file=config.yaml=examples/configs/agent-manager/config.yaml \
  -n aetherius

kubectl create configmap orchestrator-config \
  --from-file=config.yaml=examples/configs/orchestrator-service/config.yaml \
  -n aetherius

kubectl create configmap reasoning-config \
  --from-file=config.yaml=examples/configs/reasoning-service/config.yaml \
  -n aetherius

kubectl create configmap collect-agent-config \
  --from-file=config.yaml=examples/configs/collect-agent/config.yaml \
  -n kube-system
```

在 Deployment 中引用:

```yaml
spec:
  containers:
    - name: agent-manager
      volumeMounts:
        - name: config
          mountPath: /etc/aetherius/config.yaml
          subPath: config.yaml
  volumes:
    - name: config
      configMap:
        name: agent-manager-config
```

---

## 配置优先级

配置加载优先级(从高到低):

1. **命令行参数**
2. **环境变量**
3. **配置文件**
4. **默认值**

示例:

```bash
# 命令行参数优先级最高
./agent-manager --config=/etc/config.yaml --port=9090

# 环境变量次之
SERVER_PORT=8080 ./agent-manager --config=/etc/config.yaml

# 配置文件中的值
server:
  port: 8080

# 代码中的默认值
const defaultPort = 8080
```

---

## 配置验证

### 验证配置语法

```bash
# YAML 语法检查
yamllint config.yaml

# 使用服务内置验证
./agent-manager --config=config.yaml --validate
./orchestrator-service --config=config.yaml --validate
```

### 验证配置完整性

```bash
# 检查必需字段
grep -E "host:|port:|password:" config.yaml

# 检查敏感信息
grep -i "password\|secret\|token" config.yaml
```

---

## 配置最佳实践

### 1. 敏感信息管理

**不要在配置文件中硬编码敏感信息**:

```yaml
# ❌ 不好的做法
database:
  password: my-secret-password

# ✅ 好的做法 - 使用环境变量
database:
  password: ${DB_PASSWORD}
```

**使用 Kubernetes Secrets**:

```bash
# 创建 Secret
kubectl create secret generic db-credentials \
  --from-literal=password=my-secret-password \
  -n aetherius

# 在 Pod 中引用
env:
  - name: DATABASE_PASSWORD
    valueFrom:
      secretKeyRef:
        name: db-credentials
        key: password
```

### 2. 环境分离

为不同环境使用不同的配置:

```text
configs/
├── dev/
│   ├── agent-manager.yaml
│   └── orchestrator.yaml
├── staging/
│   ├── agent-manager.yaml
│   └── orchestrator.yaml
└── prod/
    ├── agent-manager.yaml
    └── orchestrator.yaml
```

### 3. 配置模板化

使用模板引擎(如 Helm):

```yaml
# values.yaml
agentManager:
  replicas: 3
  database:
    host: postgres
    port: 5432

# templates/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: agent-manager-config
data:
  config.yaml: |
    database:
      host: {{ .Values.agentManager.database.host }}
      port: {{ .Values.agentManager.database.port }}
```

### 4. 配置版本控制

```bash
# 为配置文件打标签
git tag -a config-v1.0.0 -m "Initial configuration"
git push origin config-v1.0.0

# 回滚配置
git checkout config-v1.0.0 -- config.yaml
```

### 5. 配置审计

```bash
# 记录配置变更
git log --follow -- config.yaml

# 对比配置差异
git diff v1.0.0 v2.0.0 -- config.yaml
```

---

## 常见配置场景

### 场景 1: 高可用部署

```yaml
# Agent Manager
database:
  max_open_conns: 100
  max_idle_conns: 20

redis:
  pool_size: 20
  min_idle_conns: 10

agent:
  heartbeat:
    timeout: 90
    max_failures: 3

# Orchestrator Service
workflow:
  max_concurrent_executions: 200

executor:
  workers: 50
  queue_size: 2000
```

### 场景 2: 资源受限环境

```yaml
# Collect Agent
performance:
  workers: 5
  queue_size: 500
  batch:
    size: 25
    timeout: 2000

resources:
  memory_limit: 128
  cpu_limit: 0.5

cache:
  memory:
    max_entries: 1000
```

### 场景 3: 调试和开发

```yaml
# 所有服务
logging:
  level: debug
  format: text
  enable_caller: true
  enable_stacktrace: true

metrics:
  enabled: true

health:
  enabled: true

features:
  experimental: true
```

### 场景 4: 生产环境

```yaml
# 所有服务
logging:
  level: info
  format: json
  output: stdout

security:
  tls:
    enabled: true
  auth:
    enabled: true

rate_limit:
  enabled: true
  global_rps: 1000

tracing:
  enabled: true
  sampling_rate: 0.1
```

---

## 故障排查

### 配置加载失败

```bash
# 检查配置文件是否存在
ls -l config.yaml

# 检查文件权限
chmod 644 config.yaml

# 验证 YAML 语法
yamllint config.yaml
```

### 服务无法启动

```bash
# 检查日志
docker logs aetherius-agent-manager
kubectl logs -n aetherius agent-manager-xxx

# 验证数据库连接
psql -h localhost -U aetherius -d agent_manager

# 验证 Redis 连接
redis-cli -h localhost -p 6379 -a redis_pass ping

# 验证 NATS 连接
nats-server --check nats://localhost:4222
```

### 配置未生效

```bash
# 检查配置优先级
echo $SERVER_PORT  # 环境变量

# 检查 ConfigMap
kubectl describe configmap agent-manager-config -n aetherius

# 检查 Pod 挂载
kubectl exec -n aetherius agent-manager-xxx -- cat /etc/aetherius/config.yaml
```

---

## 参考文档

- [Agent Manager 文档](../../agent-manager/README.md)
- [Orchestrator Service 文档](../../orchestrator-service/README.md)
- [Reasoning Service 文档](../../reasoning-service/README.md)
- [Collect Agent 文档](../../collect-agent/README.md)
- [部署指南](../../docs/DEPLOYMENT.md)