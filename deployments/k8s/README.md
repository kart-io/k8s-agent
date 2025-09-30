# Aetherius Kubernetes 部署

在 Kubernetes 集群中部署 Aetherius 智能运维平台。

---

## 前置要求

- Kubernetes 1.23+
- kubectl 配置正确
- 至少 3 个 worker 节点 (推荐)
- 集群至少 8 CPU 核心和 16GB 内存可用
- StorageClass 支持 (用于持久化存储)

---

## 快速部署

### 1. 创建命名空间

```bash
kubectl apply -f namespace.yaml
```

### 2. 部署依赖服务

```bash
kubectl apply -f dependencies.yaml
```

等待所有依赖服务就绪:

```bash
kubectl -n aetherius wait --for=condition=ready pod -l app=postgres --timeout=300s
kubectl -n aetherius wait --for=condition=ready pod -l app=redis --timeout=300s
kubectl -n aetherius wait --for=condition=ready pod -l app=nats --timeout=300s
kubectl -n aetherius wait --for=condition=ready pod -l app=neo4j --timeout=300s
```

### 3. 部署应用服务

```bash
kubectl apply -f agent-manager.yaml
kubectl apply -f orchestrator-service.yaml
kubectl apply -f reasoning-service.yaml
```

### 4. 验证部署

```bash
# 检查所有 Pod 状态
kubectl -n aetherius get pods

# 检查服务
kubectl -n aetherius get svc

# 查看日志
kubectl -n aetherius logs -l app=agent-manager --tail=50
```

---

## 访问服务

### 方法 1: Port Forward (开发/测试)

```bash
# Agent Manager
kubectl -n aetherius port-forward svc/agent-manager 8080:8080

# Orchestrator Service
kubectl -n aetherius port-forward svc/orchestrator-service 8081:8081

# Reasoning Service
kubectl -n aetherius port-forward svc/reasoning-service 8082:8082

# Neo4j Browser
kubectl -n aetherius port-forward svc/aetherius-neo4j 7474:7474

# NATS Monitoring
kubectl -n aetherius port-forward svc/aetherius-nats 8222:8222
```

### 方法 2: Ingress (生产环境)

创建 `ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: aetherius-ingress
  namespace: aetherius
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - api.aetherius.example.com
    secretName: aetherius-tls
  rules:
  - host: api.aetherius.example.com
    http:
      paths:
      - path: /api/v1/agents
        pathType: Prefix
        backend:
          service:
            name: agent-manager
            port:
              number: 8080
      - path: /api/v1/workflows
        pathType: Prefix
        backend:
          service:
            name: orchestrator-service
            port:
              number: 8081
      - path: /api/v1/analyze
        pathType: Prefix
        backend:
          service:
            name: reasoning-service
            port:
              number: 8082
```

应用:

```bash
kubectl apply -f ingress.yaml
```

### 方法 3: LoadBalancer (云环境)

修改 Service 类型:

```bash
kubectl -n aetherius patch svc agent-manager -p '{"spec": {"type": "LoadBalancer"}}'
```

获取外部 IP:

```bash
kubectl -n aetherius get svc agent-manager
```

---

## 部署 Collect Agent

在被监控的集群中部署 collect-agent:

### 1. 创建配置

```bash
kubectl create namespace aetherius-agent

# 创建 ConfigMap
kubectl -n aetherius-agent create configmap collect-agent-config \
  --from-literal=cluster-id=prod-cluster \
  --from-literal=central-endpoint=nats://<EXTERNAL-IP>:4222
```

### 2. 部署 Agent

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: collect-agent
  namespace: aetherius-agent

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: collect-agent
rules:
- apiGroups: [""]
  resources: ["events", "pods", "nodes", "namespaces", "services", "persistentvolumeclaims"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments", "statefulsets", "daemonsets", "replicasets"]
  verbs: ["get", "list", "watch"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: collect-agent
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: collect-agent
subjects:
- kind: ServiceAccount
  name: collect-agent
  namespace: aetherius-agent

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: collect-agent
  namespace: aetherius-agent
spec:
  replicas: 1
  selector:
    matchLabels:
      app: collect-agent
  template:
    metadata:
      labels:
        app: collect-agent
    spec:
      serviceAccountName: collect-agent
      containers:
      - name: agent
        image: aetherius/collect-agent:latest
        env:
        - name: CLUSTER_ID
          valueFrom:
            configMapKeyRef:
              name: collect-agent-config
              key: cluster-id
        - name: CENTRAL_ENDPOINT
          valueFrom:
            configMapKeyRef:
              name: collect-agent-config
              key: central-endpoint
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
EOF
```

---

## 配置说明

### 修改数据库密码

```bash
# 生成新密码
NEW_PASSWORD=$(openssl rand -base64 32)

# 更新 Secret
kubectl -n aetherius patch secret agent-manager-secrets \
  -p "{\"stringData\":{\"database-password\":\"$NEW_PASSWORD\"}}"

# 重启相关服务
kubectl -n aetherius rollout restart deployment agent-manager
kubectl -n aetherius rollout restart statefulset postgres
```

### 修改副本数

```bash
# Agent Manager
kubectl -n aetherius scale deployment agent-manager --replicas=5

# Orchestrator
kubectl -n aetherius scale deployment orchestrator-service --replicas=3

# Reasoning Service
kubectl -n aetherius scale deployment reasoning-service --replicas=3
```

### 资源限制调整

编辑对应的 YAML 文件,修改 `resources` 部分:

```yaml
resources:
  requests:
    cpu: 1000m
    memory: 2Gi
  limits:
    cpu: 4000m
    memory: 8Gi
```

重新应用:

```bash
kubectl apply -f agent-manager.yaml
```

---

## 监控和日志

### 查看日志

```bash
# 所有 Agent Manager 日志
kubectl -n aetherius logs -l app=agent-manager --tail=100 -f

# 特定 Pod 日志
kubectl -n aetherius logs <pod-name> -f

# 查看错误日志
kubectl -n aetherius logs -l app=agent-manager | grep ERROR
```

### 资源使用

```bash
# Pod 资源使用
kubectl -n aetherius top pods

# 节点资源使用
kubectl top nodes
```

### 事件

```bash
# 查看命名空间事件
kubectl -n aetherius get events --sort-by='.lastTimestamp'

# 查看特定资源事件
kubectl -n aetherius describe pod <pod-name>
```

---

## 数据备份

### PostgreSQL 备份

```bash
# 备份所有数据库
kubectl -n aetherius exec -it postgres-0 -- \
  pg_dumpall -U aetherius > aetherius-backup-$(date +%Y%m%d).sql

# 恢复
kubectl -n aetherius exec -i postgres-0 -- \
  psql -U aetherius < aetherius-backup-20250930.sql
```

### Neo4j 备份

```bash
# 创建备份
kubectl -n aetherius exec -it neo4j-0 -- \
  neo4j-admin dump --database=neo4j --to=/tmp/neo4j-backup.dump

# 复制到本地
kubectl -n aetherius cp neo4j-0:/tmp/neo4j-backup.dump ./neo4j-backup.dump

# 恢复
kubectl -n aetherius cp ./neo4j-backup.dump neo4j-0:/tmp/
kubectl -n aetherius exec -it neo4j-0 -- \
  neo4j-admin load --from=/tmp/neo4j-backup.dump --database=neo4j --force
```

### Redis 备份

```bash
# 触发保存
kubectl -n aetherius exec -it <redis-pod> -- redis-cli -a redis_pass SAVE

# 复制 RDB 文件
kubectl -n aetherius cp <redis-pod>:/data/dump.rdb ./redis-backup.rdb
```

---

## 升级

### 滚动更新

```bash
# 更新镜像
kubectl -n aetherius set image deployment/agent-manager \
  agent-manager=aetherius/agent-manager:v1.1.0

# 查看更新状态
kubectl -n aetherius rollout status deployment/agent-manager

# 回滚
kubectl -n aetherius rollout undo deployment/agent-manager
```

### 零停机升级策略

在 Deployment 中配置:

```yaml
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
```

---

## 故障排查

### Pod 无法启动

```bash
# 查看 Pod 状态
kubectl -n aetherius describe pod <pod-name>

# 查看日志
kubectl -n aetherius logs <pod-name> --previous

# 常见原因:
# 1. 镜像拉取失败 - 检查镜像名称和权限
# 2. 资源不足 - 检查节点资源
# 3. 配置错误 - 检查 ConfigMap 和 Secret
```

### 服务无法连接

```bash
# 测试服务连通性
kubectl -n aetherius run test --image=busybox -it --rm -- \
  wget -O- http://agent-manager:8080/health

# 检查 Service 和 Endpoints
kubectl -n aetherius get svc
kubectl -n aetherius get endpoints
```

### 数据库连接失败

```bash
# 测试数据库连接
kubectl -n aetherius run psql --image=postgres:14 -it --rm -- \
  psql -h aetherius-postgres -U aetherius -d aetherius

# 检查 PostgreSQL 日志
kubectl -n aetherius logs postgres-0
```

### NATS 连接问题

```bash
# 查看 NATS 监控
kubectl -n aetherius port-forward svc/aetherius-nats 8222:8222
# 访问 http://localhost:8222

# 检查连接数
curl http://localhost:8222/connz
```

---

## 安全加固

### 启用 TLS

1. 安装 cert-manager:

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
```

2. 创建 Issuer:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
```

### 网络策略

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: aetherius-network-policy
  namespace: aetherius
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: aetherius
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: aetherius
  - to:
    - namespaceSelector: {}
      podSelector:
        matchLabels:
          k8s-app: kube-dns
    ports:
    - port: 53
      protocol: UDP
```

---

## 清理

### 删除所有资源

```bash
kubectl delete namespace aetherius
```

### 删除 PersistentVolumes

```bash
kubectl get pv | grep aetherius | awk '{print $1}' | xargs kubectl delete pv
```

---

## 生产环境最佳实践

1. **高可用部署**
   - 至少 3 个副本
   - 配置 PodDisruptionBudget
   - 使用 AntiAffinity 分散到不同节点

2. **资源管理**
   - 设置合理的 requests 和 limits
   - 启用 HPA 自动扩缩容
   - 监控资源使用情况

3. **数据持久化**
   - 使用 StatefulSet 部署有状态服务
   - 配置存储类和持久卷
   - 定期备份数据

4. **安全性**
   - 使用 RBAC 最小权限
   - 启用 TLS 加密
   - 定期更新密码
   - 配置网络策略

5. **监控告警**
   - 集成 Prometheus 监控
   - 配置 Grafana 仪表板
   - 设置关键指标告警
   - 配置日志聚合

---

## 相关文档

- [系统架构](../../docs/architecture/SYSTEM_ARCHITECTURE.md)
- [Docker Compose 部署](../docker-compose/README.md)
- [Collect Agent 部署](../../collect-agent/README.md)