# 部署配置文档 (Deployment Configuration)

## 文档信息

- **版本**: v1.6
- **最后更新**: 2025-09-28
- **状态**: 正式版
- **所属系统**: Aetherius AI Agent
- **文档类型**: 部署配置指南

## 1. 概述

### 1.1 部署目标

本文档提供 Aetherius AI Agent 在 Kubernetes 环境中的完整部署指南,包括:

- **前置条件检查**: 环境要求和资源验证
- **分步部署流程**: 从基础设施到核心服务的完整部署
- **配置管理**: 环境配置、密钥管理和参数调优
- **多环境支持**: 开发、测试、生产环境的差异化配置
- **部署验证**: 健康检查、集成测试和监控配置

### 1.2 部署架构

Aetherius 采用微服务架构,核心组件包括:

```
Ingress 层 (流量入口)
    ↓
应用服务层 (Orchestrator, Reasoning, Execution, Report, Dashboard)
    ↓
存储与依赖层 (PostgreSQL, Redis, Vault, Vector DB)
    ↓
持久化存储层 (PVC/PV)
```

### 1.3 支持的环境

| 环境 | 用途 | 规模 | 高可用 |
|------|------|------|--------|
| **Development** | 本地开发测试 | 单节点 | 否 |
| **Staging** | 预生产验证 | 2-3节点 | 部分 |
| **Production** | 生产环境 | 多节点/多可用区 | 是 |

## 2. 前置条件 (Prerequisites)

### 2.1 环境要求

#### 2.1.1 Kubernetes 集群要求

```yaml
最低要求:
  kubernetes_version: ">= v1.20"
  node_count: 3
  node_spec:
    cpu: 4 cores
    memory: 16GB
    storage: 100GB SSD

推荐配置:
  kubernetes_version: ">= v1.26"
  node_count: 6+
  node_spec:
    cpu: 8 cores
    memory: 32GB
    storage: 200GB SSD
```

#### 2.1.2 依赖服务版本

| 组件 | 最低版本 | 推荐版本 |
|------|----------|----------|
| PostgreSQL | v12 | v14+ |
| Redis | v5 | v7+ |
| HashiCorp Vault | v1.8 | v1.12+ |
| Vector Database | - | Weaviate v1.20+/Qdrant v1.3+ |
| Prometheus | v2.30 | v2.45+ |
| Grafana | v8.0 | v10.0+ |

#### 2.1.3 必需的 Kubernetes 功能

```bash
# 验证功能可用性
kubectl api-versions | grep networking.k8s.io/v1       # NetworkPolicy
kubectl api-versions | grep autoscaling/v2              # HPA
kubectl api-versions | grep batch/v1                    # CronJob
kubectl api-versions | grep storage.k8s.io/v1           # StorageClass
```

### 2.2 前置条件检查脚本

```bash
#!/bin/bash
# aetherius-pre-deploy-check.sh

set -e

echo "=== Aetherius 部署前置条件检查 ==="

check_kubernetes() {
    echo "检查 Kubernetes 版本..."
    KUBE_VERSION=$(kubectl version --short 2>/dev/null | grep "Server Version" | awk '{print $3}')
    if [ -z "$KUBE_VERSION" ]; then
        echo "❌ 无法连接到 Kubernetes 集群"
        exit 1
    fi
    echo "✅ Kubernetes 版本: $KUBE_VERSION"
}

check_resources() {
    echo "检查集群资源..."
    NODE_COUNT=$(kubectl get nodes --no-headers | wc -l)
    echo "节点数量: $NODE_COUNT"

    if [ "$NODE_COUNT" -lt 3 ]; then
        echo "⚠️  警告: 生产环境建议至少 3 个节点"
    else
        echo "✅ 节点数量满足要求"
    fi

    echo "节点资源概览:"
    kubectl top nodes 2>/dev/null || echo "⚠️  metrics-server 未安装,无法显示资源使用情况"
}

check_storage_class() {
    echo "检查存储类..."
    STORAGE_CLASSES=$(kubectl get storageclass --no-headers | wc -l)
    if [ "$STORAGE_CLASSES" -eq 0 ]; then
        echo "❌ 未找到 StorageClass,请先配置动态存储供应"
        exit 1
    fi
    echo "✅ 可用的 StorageClass:"
    kubectl get storageclass
}

check_rbac() {
    echo "验证 RBAC 权限..."
    CHECKS=(
        "create:pods"
        "create:services"
        "create:configmaps"
        "create:secrets"
        "create:deployments"
        "create:statefulsets"
    )

    for check in "${CHECKS[@]}"; do
        verb=$(echo $check | cut -d: -f1)
        resource=$(echo $check | cut -d: -f2)

        if kubectl auth can-i $verb $resource --namespace=aetherius 2>/dev/null; then
            echo "✅ 权限验证通过: $verb $resource"
        else
            echo "❌ 权限不足: $verb $resource"
            exit 1
        fi
    done
}

create_namespace() {
    echo "创建/验证命名空间..."
    if kubectl get namespace aetherius &>/dev/null; then
        echo "✅ 命名空间 aetherius 已存在"
    else
        kubectl create namespace aetherius
        kubectl label namespace aetherius name=aetherius app.kubernetes.io/name=aetherius
        echo "✅ 命名空间 aetherius 已创建"
    fi
}

check_external_connectivity() {
    echo "检查外部服务连通性..."
    ENDPOINTS=(
        "https://api.openai.com"
        "https://api.anthropic.com"
    )

    for endpoint in "${ENDPOINTS[@]}"; do
        if curl -s --connect-timeout 5 -o /dev/null "$endpoint"; then
            echo "✅ 连通性验证通过: $endpoint"
        else
            echo "⚠️  无法连接: $endpoint (请检查网络策略和防火墙)"
        fi
    done
}

main() {
    check_kubernetes
    check_resources
    check_storage_class
    check_rbac
    create_namespace
    check_external_connectivity

    echo ""
    echo "=== 前置条件检查完成 ==="
    echo "✅ 环境已就绪,可以开始部署"
}

main
```

### 2.3 准备配置文件

在部署前,准备以下配置文件:

```bash
# 创建部署配置目录
mkdir -p aetherius-deploy
cd aetherius-deploy

# 创建密钥配置文件 (注意: 生产环境应使用 Vault 或 Sealed Secrets)
cat > secrets.env << 'EOF'
# 数据库密码
POSTGRES_PASSWORD=your-secure-postgres-password

# Redis 密码
REDIS_PASSWORD=your-secure-redis-password

# AI 服务 API 密钥
OPENAI_API_KEY=sk-your-openai-api-key
ANTHROPIC_API_KEY=sk-ant-your-anthropic-api-key

# Vault 配置
VAULT_TOKEN=your-vault-root-token
EOF

chmod 600 secrets.env
```

## 3. 基础设施部署

### 3.1 创建基础资源

#### 3.1.1 命名空间和 RBAC

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: aetherius
  labels:
    name: aetherius
    app.kubernetes.io/name: aetherius
    app.kubernetes.io/version: v1.6
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: aetherius-service-account
  namespace: aetherius
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: aetherius-cluster-role
rules:
- apiGroups: [""]
  resources: ["pods", "services", "endpoints", "events"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets", "statefulsets"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["batch"]
  resources: ["jobs", "cronjobs"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["metrics.k8s.io"]
  resources: ["pods", "nodes"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: aetherius-cluster-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: aetherius-cluster-role
subjects:
- kind: ServiceAccount
  name: aetherius-service-account
  namespace: aetherius
EOF
```

#### 3.1.2 网络策略

```bash
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: aetherius-network-policy
  namespace: aetherius
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: aetherius
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring
    ports:
    - protocol: TCP
      port: 8080
    - protocol: TCP
      port: 9090
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: TCP
      port: 443
  - to:
    - podSelector:
        matchLabels:
          app: postgresql
    ports:
    - protocol: TCP
      port: 5432
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
  - ports:
    - protocol: TCP
      port: 443
EOF
```

### 3.2 部署 PostgreSQL

#### 3.2.1 创建密钥

```bash
# 从环境文件加载密钥
source secrets.env

kubectl create secret generic postgresql-secret \
  --from-literal=password="$POSTGRES_PASSWORD" \
  --from-literal=username="aetherius" \
  --from-literal=database="aetherius" \
  --namespace=aetherius
```

#### 3.2.2 部署 PostgreSQL StatefulSet

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: postgresql
  namespace: aetherius
spec:
  selector:
    app: postgresql
  ports:
  - port: 5432
    targetPort: 5432
  clusterIP: None
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgresql
  namespace: aetherius
spec:
  serviceName: postgresql
  replicas: 1
  selector:
    matchLabels:
      app: postgresql
  template:
    metadata:
      labels:
        app: postgresql
    spec:
      containers:
      - name: postgresql
        image: postgres:14-alpine
        env:
        - name: POSTGRES_DB
          valueFrom:
            secretKeyRef:
              name: postgresql-secret
              key: database
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: postgresql-secret
              key: username
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgresql-secret
              key: password
        - name: PGDATA
          value: /var/lib/postgresql/data/pgdata
        ports:
        - containerPort: 5432
          name: postgresql
        volumeMounts:
        - name: postgresql-storage
          mountPath: /var/lib/postgresql/data
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
        livenessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - aetherius
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - aetherius
          initialDelaySeconds: 5
          periodSeconds: 5
  volumeClaimTemplates:
  - metadata:
      name: postgresql-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 100Gi
      storageClassName: standard
EOF
```

#### 3.2.3 数据库初始化

```bash
# 等待 PostgreSQL 就绪
kubectl wait --for=condition=ready pod -l app=postgresql --timeout=300s -n aetherius

# 执行数据库初始化作业
kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: aetherius-db-migration
  namespace: aetherius
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
      - name: db-migration
        image: migrate/migrate:latest
        env:
        - name: DATABASE_URL
          value: "postgres://aetherius:\$(POSTGRES_PASSWORD)@postgresql:5432/aetherius?sslmode=disable"
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgresql-secret
              key: password
        command:
        - sh
        - -c
        - |
          cat > /migrations/001_initial_schema.up.sql << 'SQL'
          -- 创建诊断任务表
          CREATE TABLE IF NOT EXISTS diagnostic_tasks (
              id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
              alert_id VARCHAR(255) NOT NULL,
              cluster_id VARCHAR(255) NOT NULL,
              namespace VARCHAR(255),
              priority INTEGER NOT NULL,
              status VARCHAR(50) NOT NULL,
              created_at TIMESTAMP NOT NULL DEFAULT NOW(),
              started_at TIMESTAMP,
              completed_at TIMESTAMP,
              context JSONB,
              steps JSONB,
              result JSONB,
              metadata JSONB
          );

          CREATE INDEX idx_tasks_status ON diagnostic_tasks(status);
          CREATE INDEX idx_tasks_cluster_id ON diagnostic_tasks(cluster_id);
          CREATE INDEX idx_tasks_created_at ON diagnostic_tasks(created_at DESC);

          -- 创建知识库表
          CREATE TABLE IF NOT EXISTS knowledge_base (
              id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
              title VARCHAR(255) NOT NULL,
              content TEXT NOT NULL,
              category VARCHAR(100),
              tags TEXT[],
              embedding FLOAT8[],
              metadata JSONB,
              created_at TIMESTAMP NOT NULL DEFAULT NOW(),
              updated_at TIMESTAMP NOT NULL DEFAULT NOW()
          );

          CREATE INDEX idx_kb_category ON knowledge_base(category);
          CREATE INDEX idx_kb_tags ON knowledge_base USING GIN(tags);
          SQL

          migrate -path /migrations -database "$DATABASE_URL" up
      serviceAccountName: aetherius-service-account
EOF

# 等待数据库迁移完成
kubectl wait --for=condition=complete job/aetherius-db-migration --timeout=300s -n aetherius
```

### 3.3 部署 Redis

#### 3.3.1 创建 Redis 密钥

```bash
kubectl create secret generic redis-secret \
  --from-literal=password="$REDIS_PASSWORD" \
  --namespace=aetherius
```

#### 3.3.2 部署 Redis StatefulSet

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: aetherius
spec:
  selector:
    app: redis
  ports:
  - port: 6379
    targetPort: 6379
  clusterIP: None
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis
  namespace: aetherius
spec:
  serviceName: redis
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        command:
        - redis-server
        - --requirepass
        - \$(REDIS_PASSWORD)
        - --appendonly
        - "yes"
        - --maxmemory
        - "512mb"
        - --maxmemory-policy
        - allkeys-lru
        env:
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-secret
              key: password
        ports:
        - containerPort: 6379
          name: redis
        volumeMounts:
        - name: redis-storage
          mountPath: /data
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          exec:
            command:
            - redis-cli
            - -a
            - \$(REDIS_PASSWORD)
            - ping
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - redis-cli
            - -a
            - \$(REDIS_PASSWORD)
            - ping
          initialDelaySeconds: 5
          periodSeconds: 5
  volumeClaimTemplates:
  - metadata:
      name: redis-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 50Gi
      storageClassName: standard
EOF
```

### 3.4 部署 HashiCorp Vault

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: vault
  namespace: aetherius
spec:
  selector:
    app: vault
  ports:
  - port: 8200
    targetPort: 8200
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: vault
  namespace: aetherius
spec:
  serviceName: vault
  replicas: 1
  selector:
    matchLabels:
      app: vault
  template:
    metadata:
      labels:
        app: vault
    spec:
      containers:
      - name: vault
        image: vault:1.12
        env:
        - name: VAULT_DEV_ROOT_TOKEN_ID
          valueFrom:
            secretKeyRef:
              name: vault-secret
              key: root-token
        - name: VAULT_DEV_LISTEN_ADDRESS
          value: "0.0.0.0:8200"
        ports:
        - containerPort: 8200
          name: vault
        volumeMounts:
        - name: vault-storage
          mountPath: /vault/data
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        securityContext:
          capabilities:
            add:
            - IPC_LOCK
  volumeClaimTemplates:
  - metadata:
      name: vault-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 20Gi
EOF

# 创建 Vault 密钥
kubectl create secret generic vault-secret \
  --from-literal=root-token="$VAULT_TOKEN" \
  --namespace=aetherius
```

### 3.5 部署向量数据库 (Weaviate)

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: weaviate
  namespace: aetherius
spec:
  selector:
    app: weaviate
  ports:
  - port: 8080
    targetPort: 8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: weaviate
  namespace: aetherius
spec:
  replicas: 2
  selector:
    matchLabels:
      app: weaviate
  template:
    metadata:
      labels:
        app: weaviate
    spec:
      containers:
      - name: weaviate
        image: semitechnologies/weaviate:1.20.0
        env:
        - name: QUERY_DEFAULTS_LIMIT
          value: "25"
        - name: AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED
          value: "true"
        - name: PERSISTENCE_DATA_PATH
          value: "/var/lib/weaviate"
        - name: DEFAULT_VECTORIZER_MODULE
          value: "none"
        - name: ENABLE_MODULES
          value: ""
        - name: CLUSTER_HOSTNAME
          value: "weaviate"
        ports:
        - containerPort: 8080
          name: http
        volumeMounts:
        - name: weaviate-storage
          mountPath: /var/lib/weaviate
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
      volumes:
      - name: weaviate-storage
        persistentVolumeClaim:
          claimName: weaviate-pvc
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: weaviate-pvc
  namespace: aetherius
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 200Gi
  storageClassName: standard
EOF
```

## 4. 核心服务部署

### 4.1 创建配置 ConfigMap

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: aetherius-config
  namespace: aetherius
data:
  config.yaml: |
    server:
      port: 8080
      host: "0.0.0.0"
      read_timeout: 30s
      write_timeout: 30s
      shutdown_timeout: 10s

    database:
      host: postgresql
      port: 5432
      database: aetherius
      username: aetherius
      max_connections: 50
      idle_connections: 10
      connection_lifetime: 1h
      connection_timeout: 5s

    redis:
      host: redis
      port: 6379
      db: 0
      pool_size: 20
      idle_timeout: 5m
      dial_timeout: 5s

    vault:
      address: http://vault:8200
      mount_path: secret
      token_renewal_interval: 1h

    vector_db:
      type: weaviate
      endpoint: http://weaviate:8080
      timeout: 30s

    task_queue:
      max_concurrent_tasks: 50
      task_timeout: 10m
      retry_attempts: 3
      retry_delay: 30s
      priority_weights:
        p0: 100
        p1: 50
        p2: 10
        p3: 1

    ai_service:
      provider: openai
      model: gpt-4
      max_tokens: 4000
      temperature: 0.1
      timeout: 60s
      max_retries: 3

    security:
      enable_rbac: true
      token_expiry: 24h
      session_timeout: 8h
      allowed_namespaces: []

    monitoring:
      enable_metrics: true
      metrics_port: 9090
      health_check_interval: 30s
      enable_tracing: true
      tracing_endpoint: http://jaeger:14268/api/traces

    logging:
      level: info
      format: json
      output: stdout
      enable_caller: true
      enable_stacktrace: true
EOF
```

### 4.2 创建应用密钥

```bash
kubectl create secret generic aetherius-secrets \
  --from-literal=database-url="postgres://aetherius:$POSTGRES_PASSWORD@postgresql:5432/aetherius?sslmode=disable" \
  --from-literal=redis-url="redis://:$REDIS_PASSWORD@redis:6379/0" \
  --from-literal=openai-api-key="$OPENAI_API_KEY" \
  --from-literal=anthropic-api-key="$ANTHROPIC_API_KEY" \
  --from-literal=vault-token="$VAULT_TOKEN" \
  --namespace=aetherius
```

### 4.3 部署 Orchestrator 服务

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: aetherius-orchestrator
  namespace: aetherius
  labels:
    app: aetherius-orchestrator
spec:
  selector:
    app: aetherius-orchestrator
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aetherius-orchestrator
  namespace: aetherius
spec:
  replicas: 3
  selector:
    matchLabels:
      app: aetherius-orchestrator
  template:
    metadata:
      labels:
        app: aetherius-orchestrator
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: aetherius-service-account
      containers:
      - name: orchestrator
        image: aetherius/orchestrator:v1.6
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: CONFIG_PATH
          value: /etc/aetherius/config.yaml
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: aetherius-secrets
              key: database-url
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: aetherius-secrets
              key: redis-url
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: aetherius-secrets
              key: openai-api-key
        - name: VAULT_TOKEN
          valueFrom:
            secretKeyRef:
              name: aetherius-secrets
              key: vault-token
        volumeMounts:
        - name: config
          mountPath: /etc/aetherius
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: aetherius-config
EOF
```

### 4.4 部署 Reasoning 服务

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: aetherius-reasoning
  namespace: aetherius
spec:
  selector:
    app: aetherius-reasoning
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aetherius-reasoning
  namespace: aetherius
spec:
  replicas: 2
  selector:
    matchLabels:
      app: aetherius-reasoning
  template:
    metadata:
      labels:
        app: aetherius-reasoning
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
    spec:
      serviceAccountName: aetherius-service-account
      containers:
      - name: reasoning
        image: aetherius/reasoning:v1.6
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: CONFIG_PATH
          value: /etc/aetherius/config.yaml
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: aetherius-secrets
              key: openai-api-key
        volumeMounts:
        - name: config
          mountPath: /etc/aetherius
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
      volumes:
      - name: config
        configMap:
          name: aetherius-config
EOF
```

### 4.5 部署 Execution 服务

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: aetherius-execution
  namespace: aetherius
spec:
  selector:
    app: aetherius-execution
  ports:
  - name: http
    port: 80
    targetPort: 8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aetherius-execution
  namespace: aetherius
spec:
  replicas: 2
  selector:
    matchLabels:
      app: aetherius-execution
  template:
    metadata:
      labels:
        app: aetherius-execution
    spec:
      serviceAccountName: aetherius-service-account
      containers:
      - name: execution
        image: aetherius/execution:v1.6
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: CONFIG_PATH
          value: /etc/aetherius/config.yaml
        - name: VAULT_TOKEN
          valueFrom:
            secretKeyRef:
              name: aetherius-secrets
              key: vault-token
        volumeMounts:
        - name: config
          mountPath: /etc/aetherius
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: aetherius-config
EOF
```

### 4.6 部署 Report 和 Dashboard 服务

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: aetherius-report
  namespace: aetherius
spec:
  selector:
    app: aetherius-report
  ports:
  - port: 80
    targetPort: 8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aetherius-report
  namespace: aetherius
spec:
  replicas: 2
  selector:
    matchLabels:
      app: aetherius-report
  template:
    metadata:
      labels:
        app: aetherius-report
    spec:
      serviceAccountName: aetherius-service-account
      containers:
      - name: report
        image: aetherius/report:v1.6
        ports:
        - containerPort: 8080
        env:
        - name: CONFIG_PATH
          value: /etc/aetherius/config.yaml
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: aetherius-secrets
              key: database-url
        volumeMounts:
        - name: config
          mountPath: /etc/aetherius
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: aetherius-config
---
apiVersion: v1
kind: Service
metadata:
  name: aetherius-dashboard
  namespace: aetherius
spec:
  selector:
    app: aetherius-dashboard
  ports:
  - port: 80
    targetPort: 3000
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aetherius-dashboard
  namespace: aetherius
spec:
  replicas: 2
  selector:
    matchLabels:
      app: aetherius-dashboard
  template:
    metadata:
      labels:
        app: aetherius-dashboard
    spec:
      containers:
      - name: dashboard
        image: aetherius/dashboard:v1.6
        ports:
        - containerPort: 3000
        env:
        - name: API_ENDPOINT
          value: http://aetherius-orchestrator
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
EOF
```

## 5. Ingress 配置

### 5.1 部署 Ingress 资源

```bash
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: aetherius-ingress
  namespace: aetherius
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - aetherius.example.com
    secretName: aetherius-tls
  rules:
  - host: aetherius.example.com
    http:
      paths:
      - path: /api/v1
        pathType: Prefix
        backend:
          service:
            name: aetherius-orchestrator
            port:
              number: 80
      - path: /webhook
        pathType: Prefix
        backend:
          service:
            name: aetherius-orchestrator
            port:
              number: 80
      - path: /
        pathType: Prefix
        backend:
          service:
            name: aetherius-dashboard
            port:
              number: 80
EOF
```

## 6. 部署验证

### 6.1 健康检查脚本

```bash
#!/bin/bash
# health-check.sh

NAMESPACE="aetherius"

echo "=== Aetherius 健康检查 ==="

check_pods() {
    echo "检查 Pod 状态..."
    kubectl get pods -n $NAMESPACE -o wide

    NOT_READY=$(kubectl get pods -n $NAMESPACE -o json | jq -r '.items[] | select(.status.phase!="Running") | .metadata.name')
    if [ -n "$NOT_READY" ]; then
        echo "⚠️  以下 Pod 未就绪: $NOT_READY"
        return 1
    fi
    echo "✅ 所有 Pod 运行正常"
}

check_services() {
    echo "检查服务端点..."
    SERVICES=(
        "aetherius-orchestrator"
        "aetherius-reasoning"
        "aetherius-execution"
        "postgresql"
        "redis"
        "weaviate"
    )

    for svc in "${SERVICES[@]}"; do
        ENDPOINTS=$(kubectl get endpoints -n $NAMESPACE $svc -o json | jq -r '.subsets[].addresses | length')
        if [ "$ENDPOINTS" -gt 0 ]; then
            echo "✅ $svc: $ENDPOINTS 个端点就绪"
        else
            echo "❌ $svc: 无可用端点"
            return 1
        fi
    done
}

check_health_endpoints() {
    echo "检查健康接口..."

    kubectl port-forward -n $NAMESPACE svc/aetherius-orchestrator 8080:80 &
    PF_PID=$!
    sleep 5

    if curl -s -f http://localhost:8080/health/live > /dev/null; then
        echo "✅ Liveness 检查通过"
    else
        echo "❌ Liveness 检查失败"
        kill $PF_PID
        return 1
    fi

    if curl -s -f http://localhost:8080/health/ready > /dev/null; then
        echo "✅ Readiness 检查通过"
    else
        echo "❌ Readiness 检查失败"
        kill $PF_PID
        return 1
    fi

    kill $PF_PID
}

main() {
    check_pods || exit 1
    check_services || exit 1
    check_health_endpoints || exit 1

    echo ""
    echo "=== 健康检查完成 ==="
    echo "✅ 系统部署成功且运行正常"
}

main
```

### 6.2 集成测试脚本

```bash
#!/bin/bash
# integration-test.sh

NAMESPACE="aetherius"
API_ENDPOINT="http://localhost:8080"

echo "=== Aetherius 集成测试 ==="

setup_port_forward() {
    kubectl port-forward -n $NAMESPACE svc/aetherius-orchestrator 8080:80 &
    PF_PID=$!
    sleep 5
}

test_database_connectivity() {
    echo "测试数据库连接..."
    RESPONSE=$(curl -s "$API_ENDPOINT/api/v1/health/database")
    if echo "$RESPONSE" | jq -e '.status == "healthy"' > /dev/null; then
        echo "✅ 数据库连接正常"
        return 0
    else
        echo "❌ 数据库连接失败: $RESPONSE"
        return 1
    fi
}

test_redis_connectivity() {
    echo "测试 Redis 连接..."
    RESPONSE=$(curl -s "$API_ENDPOINT/api/v1/health/redis")
    if echo "$RESPONSE" | jq -e '.status == "healthy"' > /dev/null; then
        echo "✅ Redis 连接正常"
        return 0
    else
        echo "❌ Redis 连接失败: $RESPONSE"
        return 1
    fi
}

test_alert_webhook() {
    echo "测试告警 Webhook..."
    TEST_ALERT='{
      "receiver": "aetherius-webhook",
      "status": "firing",
      "alerts": [{
        "status": "firing",
        "labels": {
          "alertname": "IntegrationTestAlert",
          "severity": "warning",
          "cluster_id": "test-cluster",
          "namespace": "default"
        },
        "annotations": {
          "description": "Integration test alert",
          "summary": "Test"
        }
      }]
    }'

    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
      -H "Content-Type: application/json" \
      -d "$TEST_ALERT" \
      "$API_ENDPOINT/api/v1/webhook/alertmanager")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" -eq 200 ] || [ "$HTTP_CODE" -eq 202 ]; then
        echo "✅ Webhook 接收成功: $BODY"
        return 0
    else
        echo "❌ Webhook 接收失败 (HTTP $HTTP_CODE): $BODY"
        return 1
    fi
}

cleanup() {
    if [ -n "$PF_PID" ]; then
        kill $PF_PID 2>/dev/null
    fi
}

main() {
    trap cleanup EXIT

    setup_port_forward

    test_database_connectivity || exit 1
    test_redis_connectivity || exit 1
    test_alert_webhook || exit 1

    echo ""
    echo "=== 集成测试完成 ==="
    echo "✅ 所有测试通过"
}

main
```

## 7. 多环境配置

### 7.1 开发环境配置

```yaml
# values-dev.yaml
global:
  environment: development

orchestrator:
  replicaCount: 1
  resources:
    requests:
      memory: "256Mi"
      cpu: "100m"
    limits:
      memory: "512Mi"
      cpu: "250m"

postgresql:
  persistence:
    size: 10Gi

redis:
  persistence:
    size: 5Gi
```

### 7.2 测试环境配置

```yaml
# values-staging.yaml
global:
  environment: staging

orchestrator:
  replicaCount: 2
  resources:
    requests:
      memory: "512Mi"
      cpu: "250m"
    limits:
      memory: "1Gi"
      cpu: "500m"

postgresql:
  persistence:
    size: 50Gi
  backup:
    enabled: true
    schedule: "0 2 * * *"

redis:
  persistence:
    size: 20Gi
```

### 7.3 生产环境配置

```yaml
# values-prod.yaml
global:
  environment: production

orchestrator:
  replicaCount: 3
  resources:
    requests:
      memory: "512Mi"
      cpu: "250m"
    limits:
      memory: "1Gi"
      cpu: "500m"
  autoscaling:
    enabled: true
    minReplicas: 3
    maxReplicas: 20
    targetCPUUtilizationPercentage: 70
    targetMemoryUtilizationPercentage: 80

postgresql:
  replicaCount: 3
  persistence:
    size: 100Gi
  backup:
    enabled: true
    schedule: "0 */6 * * *"
    retention: 30d

redis:
  replicaCount: 3
  persistence:
    size: 50Gi

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  tls:
    enabled: true
```

## 8. Helm Chart 部署

### 8.1 Helm Chart 结构

```
aetherius/
├── Chart.yaml
├── values.yaml
├── templates/
│   ├── namespace.yaml
│   ├── rbac.yaml
│   ├── configmap.yaml
│   ├── secrets.yaml
│   ├── postgresql/
│   │   ├── statefulset.yaml
│   │   └── service.yaml
│   ├── redis/
│   │   ├── statefulset.yaml
│   │   └── service.yaml
│   ├── orchestrator/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── hpa.yaml
│   ├── reasoning/
│   ├── execution/
│   └── ingress.yaml
└── charts/
```

### 8.2 使用 Helm 部署

```bash
# 添加 Helm 仓库 (假设)
helm repo add aetherius https://charts.aetherius.io
helm repo update

# 安装到开发环境
helm install aetherius aetherius/aetherius \
  --namespace aetherius \
  --create-namespace \
  --values values-dev.yaml

# 安装到生产环境
helm install aetherius aetherius/aetherius \
  --namespace aetherius \
  --create-namespace \
  --values values-prod.yaml \
  --set global.imageTag=v1.6 \
  --set postgresql.auth.password=$POSTGRES_PASSWORD \
  --set redis.auth.password=$REDIS_PASSWORD

# 升级部署
helm upgrade aetherius aetherius/aetherius \
  --namespace aetherius \
  --values values-prod.yaml \
  --set global.imageTag=v1.7

# 回滚到上一个版本
helm rollback aetherius -n aetherius
```

## 9. 配置管理最佳实践

### 9.1 密钥管理

#### 9.1.1 使用 External Secrets Operator

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: vault-backend
  namespace: aetherius
spec:
  provider:
    vault:
      server: "http://vault:8200"
      path: "secret"
      version: "v2"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "aetherius"
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: aetherius-secrets
  namespace: aetherius
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: aetherius-secrets
    creationPolicy: Owner
  data:
  - secretKey: database-url
    remoteRef:
      key: aetherius/database
      property: url
  - secretKey: openai-api-key
    remoteRef:
      key: aetherius/ai
      property: openai_key
```

### 9.2 配置版本控制

```bash
# 将配置存储在 Git 仓库
git init aetherius-config
cd aetherius-config

# 创建配置文件结构
mkdir -p environments/{dev,staging,prod}

# 使用 GitOps 工具 (如 ArgoCD) 管理部署
kubectl apply -f - <<EOF
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: aetherius
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/your-org/aetherius-config
    targetRevision: HEAD
    path: environments/prod
  destination:
    server: https://kubernetes.default.svc
    namespace: aetherius
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
EOF
```

## 10. 故障排查指南

### 10.1 常见问题

#### 问题 1: Pod 无法启动

```bash
# 检查 Pod 状态和事件
kubectl describe pod <pod-name> -n aetherius

# 查看日志
kubectl logs <pod-name> -n aetherius

# 常见原因:
# - 镜像拉取失败: 检查镜像名称和拉取策略
# - 资源不足: 检查节点资源
# - 配置错误: 验证 ConfigMap 和 Secret
```

#### 问题 2: 数据库连接失败

```bash
# 验证 PostgreSQL 运行状态
kubectl exec -it postgresql-0 -n aetherius -- psql -U aetherius -d aetherius -c "SELECT 1;"

# 检查网络策略
kubectl get networkpolicies -n aetherius

# 验证密钥
kubectl get secret postgresql-secret -n aetherius -o jsonpath='{.data.password}' | base64 -d
```

#### 问题 3: Ingress 无法访问

```bash
# 检查 Ingress 控制器
kubectl get pods -n ingress-nginx

# 查看 Ingress 规则
kubectl describe ingress aetherius-ingress -n aetherius

# 测试服务可达性
kubectl port-forward -n aetherius svc/aetherius-orchestrator 8080:80
curl http://localhost:8080/health
```

### 10.2 调试工具

```bash
# 进入 Pod 调试
kubectl exec -it <pod-name> -n aetherius -- /bin/sh

# 临时运行调试 Pod
kubectl run debug-pod --rm -it --image=nicolaka/netshoot -n aetherius -- /bin/bash

# 检查 DNS 解析
kubectl run -it --rm debug --image=busybox --restart=Never -n aetherius -- nslookup postgresql

# 查看资源使用
kubectl top pods -n aetherius
kubectl top nodes
```

## 11. 升级和回滚

### 11.1 滚动升级策略

```yaml
# 在 Deployment 中配置滚动升级策略
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
```

### 11.2 蓝绿部署

```bash
# 部署新版本 (绿环境)
kubectl apply -f deployment-v2.yaml

# 切换流量到新版本
kubectl patch service aetherius-orchestrator -p '{"spec":{"selector":{"version":"v2"}}}'

# 验证成功后删除旧版本
kubectl delete deployment aetherius-orchestrator-v1
```

### 11.3 金丝雀发布

```yaml
# 使用 Flagger 进行金丝雀发布
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: aetherius-orchestrator
  namespace: aetherius
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: aetherius-orchestrator
  progressDeadlineSeconds: 600
  service:
    port: 80
  analysis:
    interval: 1m
    threshold: 5
    maxWeight: 50
    stepWeight: 10
    metrics:
    - name: request-success-rate
      thresholdRange:
        min: 99
    - name: request-duration
      thresholdRange:
        max: 500
```

## 12. 卸载和清理

### 12.1 使用 kubectl 卸载

```bash
#!/bin/bash
# uninstall.sh

NAMESPACE="aetherius"

echo "=== 开始卸载 Aetherius ==="

# 删除应用资源
kubectl delete deployment --all -n $NAMESPACE
kubectl delete statefulset --all -n $NAMESPACE
kubectl delete service --all -n $NAMESPACE
kubectl delete ingress --all -n $NAMESPACE

# 删除配置
kubectl delete configmap --all -n $NAMESPACE
kubectl delete secret --all -n $NAMESPACE

# 删除 RBAC
kubectl delete clusterrolebinding aetherius-cluster-role-binding
kubectl delete clusterrole aetherius-cluster-role
kubectl delete serviceaccount aetherius-service-account -n $NAMESPACE

# 删除持久化数据 (慎重!)
read -p "是否删除持久化数据? (yes/no): " CONFIRM
if [ "$CONFIRM" = "yes" ]; then
    kubectl delete pvc --all -n $NAMESPACE
    echo "⚠️  持久化数据已删除"
fi

# 删除命名空间
kubectl delete namespace $NAMESPACE

echo "=== 卸载完成 ==="
```

### 12.2 使用 Helm 卸载

```bash
# 卸载 Helm release
helm uninstall aetherius -n aetherius

# 清理命名空间
kubectl delete namespace aetherius
```

## 附录

### A. 完整部署脚本

```bash
#!/bin/bash
# deploy-aetherius.sh - 一键部署脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NAMESPACE="aetherius"

echo "=== Aetherius 自动部署脚本 ==="

# 加载配置
if [ -f "$SCRIPT_DIR/secrets.env" ]; then
    source "$SCRIPT_DIR/secrets.env"
else
    echo "❌ secrets.env 文件不存在"
    exit 1
fi

# 执行前置检查
bash "$SCRIPT_DIR/aetherius-pre-deploy-check.sh" || exit 1

# 部署基础设施
echo "部署基础设施..."
kubectl apply -f "$SCRIPT_DIR/manifests/01-namespace.yaml"
kubectl apply -f "$SCRIPT_DIR/manifests/02-rbac.yaml"
kubectl apply -f "$SCRIPT_DIR/manifests/03-networkpolicy.yaml"

# 创建密钥
echo "创建密钥..."
kubectl create secret generic postgresql-secret \
  --from-literal=password="$POSTGRES_PASSWORD" \
  --namespace=$NAMESPACE \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret generic redis-secret \
  --from-literal=password="$REDIS_PASSWORD" \
  --namespace=$NAMESPACE \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret generic aetherius-secrets \
  --from-literal=openai-api-key="$OPENAI_API_KEY" \
  --namespace=$NAMESPACE \
  --dry-run=client -o yaml | kubectl apply -f -

# 部署依赖服务
echo "部署依赖服务..."
kubectl apply -f "$SCRIPT_DIR/manifests/04-postgresql.yaml"
kubectl apply -f "$SCRIPT_DIR/manifests/05-redis.yaml"
kubectl apply -f "$SCRIPT_DIR/manifests/06-weaviate.yaml"

# 等待依赖服务就绪
echo "等待依赖服务就绪..."
kubectl wait --for=condition=ready pod -l app=postgresql --timeout=300s -n $NAMESPACE
kubectl wait --for=condition=ready pod -l app=redis --timeout=300s -n $NAMESPACE

# 数据库初始化
echo "初始化数据库..."
kubectl apply -f "$SCRIPT_DIR/manifests/07-db-migration.yaml"
kubectl wait --for=condition=complete job/aetherius-db-migration --timeout=300s -n $NAMESPACE

# 部署核心服务
echo "部署核心服务..."
kubectl apply -f "$SCRIPT_DIR/manifests/08-config.yaml"
kubectl apply -f "$SCRIPT_DIR/manifests/09-orchestrator.yaml"
kubectl apply -f "$SCRIPT_DIR/manifests/10-reasoning.yaml"
kubectl apply -f "$SCRIPT_DIR/manifests/11-execution.yaml"
kubectl apply -f "$SCRIPT_DIR/manifests/12-report.yaml"
kubectl apply -f "$SCRIPT_DIR/manifests/13-dashboard.yaml"

# 部署 Ingress
echo "部署 Ingress..."
kubectl apply -f "$SCRIPT_DIR/manifests/14-ingress.yaml"

# 等待核心服务就绪
echo "等待核心服务就绪..."
kubectl wait --for=condition=available deployment --all --timeout=300s -n $NAMESPACE

# 执行健康检查
echo "执行健康检查..."
bash "$SCRIPT_DIR/health-check.sh"

echo ""
echo "=== 部署完成 ==="
echo "✅ Aetherius 已成功部署到命名空间: $NAMESPACE"
echo ""
echo "访问地址:"
echo "- Dashboard: http://$(kubectl get ingress aetherius-ingress -n $NAMESPACE -o jsonpath='{.spec.rules[0].host}')"
echo "- API: http://$(kubectl get ingress aetherius-ingress -n $NAMESPACE -o jsonpath='{.spec.rules[0].host}')/api/v1"
echo ""
echo "查看 Pod 状态: kubectl get pods -n $NAMESPACE"
echo "查看服务日志: kubectl logs -f deployment/aetherius-orchestrator -n $NAMESPACE"
```

### B. 参考资源

- **Kubernetes 文档**: <https://kubernetes.io/docs>
- **Helm 文档**: <https://helm.sh/docs>
- **PostgreSQL 文档**: <https://www.postgresql.org/docs>
- **Redis 文档**: <https://redis.io/documentation>
- **HashiCorp Vault**: <https://www.vaultproject.io/docs>
- **Weaviate 文档**: <https://weaviate.io/developers/weaviate>

### C. 相关文档

- [架构设计文档](./02_architecture.md) - 系统架构详细设计
- [数据模型文档](./03_data_models.md) - 核心数据模型定义
- [需求规格说明](../REQUIREMENTS.md) - 完整需求索引
- [运维安全文档](./05_operations.md) - 运维和安全指南 (待创建)