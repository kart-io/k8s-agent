# K8s Agent - 生产部署指南

## 目录

- [构建准备](#构建准备)
- [构建应用](#构建应用)
- [Docker 部署](#docker-部署)
- [Kubernetes 部署](#kubernetes-部署)
- [生产配置](#生产配置)
- [监控和告警](#监控和告警)
- [故障排查](#故障排查)

## 构建准备

### 系统要求

- Go 1.21+ (用于编译)
- Docker 20.10+ (用于容器化)
- Kubernetes 1.24+ (用于 K8s 部署)
- PostgreSQL/MySQL 8.0+ (数据库)

### 源码获取

```bash
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/cluster-service
```

## 构建应用

### 1. 本地构建

```bash
# 标准构建
go build -o bin/cluster-service cmd/server/main.go

# 生产构建（优化大小）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w" \
  -o bin/cluster-service \
  cmd/server/main.go

# 查看构建结果
ls -lh bin/cluster-service
# 预期输出：约 50-60MB
```

### 2. 使用 Makefile

创建 `Makefile`:

```makefile
# Makefile
.PHONY: build clean test run docker-build docker-push

VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags="$(LDFLAGS)" \
		-o bin/cluster-service \
		cmd/server/main.go

clean:
	rm -rf bin/

test:
	go test -v -cover ./...

run:
	go run cmd/server/main.go -config configs/config-dev.yaml

docker-build:
	docker build -t k8s-agent/cluster-service:$(VERSION) .

docker-push:
	docker push k8s-agent/cluster-service:$(VERSION)
```

使用：

```bash
make build
make test
```

## Docker 部署

### 1. 创建 Dockerfile

在 `cluster-service/` 目录创建 `Dockerfile`:

```dockerfile
# 多阶段构建
FROM golang:1.21-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git make

WORKDIR /build

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./
COPY ../common/go.mod ../common/go.sum ../common/
COPY ../../logger/ ../../logger/

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .
COPY ../common ../common

# 构建
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o cluster-service \
    cmd/server/main.go

# 运行阶段
FROM alpine:3.18

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/cluster-service .
COPY configs/ configs/

# 设置权限
RUN chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动命令
CMD ["./cluster-service", "-config", "configs/config.yaml"]
```

### 2. 构建镜像

```bash
# 构建镜像
docker build -t k8s-agent/cluster-service:v1.0.0 .

# 查看镜像大小
docker images k8s-agent/cluster-service

# 运行镜像测试
docker run -d \
  --name cluster-service \
  -p 8080:8080 \
  -e DB_HOST=postgres \
  -e DB_PASSWORD=secret \
  k8s-agent/cluster-service:v1.0.0

# 查看日志
docker logs -f cluster-service

# 测试健康检查
curl http://localhost:8080/health
```

### 3. Docker Compose 部署

创建 `docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:13-alpine
    environment:
      POSTGRES_DB: cluster_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${DB_PASSWORD:-postgres}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  cluster-service:
    image: k8s-agent/cluster-service:v1.0.0
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: ${DB_PASSWORD:-postgres}
      DB_NAME: cluster_db
      SERVER_PORT: 8080
      LOG_LEVEL: info
      LOG_FORMAT: json
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped

volumes:
  postgres_data:
```

启动：

```bash
# 启动所有服务
docker-compose up -d

# 查看状态
docker-compose ps

# 查看日志
docker-compose logs -f cluster-service

# 停止服务
docker-compose down
```

## Kubernetes 部署

### 1. 创建配置文件

**ConfigMap** - `k8s/configmap.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-service-config
  namespace: kart-io
data:
  config.yaml: |
    server:
      port: 8080
      mode: release
      read_timeout: 30s
      write_timeout: 30s

    database:
      host: postgres-service
      port: 5432
      user: postgres
      dbname: cluster_db
      sslmode: disable
      max_open_conns: 100
      max_idle_conns: 10

    logging:
      level: info
      format: json

    jwt:
      secret: ${JWT_SECRET}
```

**Secret** - `k8s/secret.yaml`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cluster-service-secret
  namespace: kart-io
type: Opaque
stringData:
  DB_PASSWORD: "your-db-password"
  JWT_SECRET: "your-jwt-secret-key"
```

**Deployment** - `k8s/deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cluster-service
  namespace: kart-io
  labels:
    app: cluster-service
    version: v1.0.0
spec:
  replicas: 3
  selector:
    matchLabels:
      app: cluster-service
  template:
    metadata:
      labels:
        app: cluster-service
        version: v1.0.0
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: cluster-service
      containers:
      - name: cluster-service
        image: k8s-agent/cluster-service:v1.0.0
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
        env:
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: cluster-service-secret
              key: DB_PASSWORD
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: cluster-service-secret
              key: JWT_SECRET
        - name: SERVICE_VERSION
          value: "v1.0.0"
        volumeMounts:
        - name: config
          mountPath: /app/configs
          readOnly: true
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        securityContext:
          runAsNonRoot: true
          runAsUser: 1000
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
      volumes:
      - name: config
        configMap:
          name: cluster-service-config
```

**Service** - `k8s/service.yaml`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: cluster-service
  namespace: kart-io
  labels:
    app: cluster-service
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: cluster-service
```

**ServiceAccount** - `k8s/serviceaccount.yaml`:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cluster-service
  namespace: kart-io

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-service-role
rules:
- apiGroups: [""]
  resources: ["pods", "namespaces", "nodes", "services"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "statefulsets", "daemonsets"]
  verbs: ["get", "list", "watch", "create", "update", "delete", "patch"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cluster-service-binding
subjects:
- kind: ServiceAccount
  name: cluster-service
  namespace: kart-io
roleRef:
  kind: ClusterRole
  name: cluster-service-role
  apiGroup: rbac.authorization.k8s.io
```

### 2. 部署到 Kubernetes

```bash
# 创建命名空间
kubectl create namespace kart-io

# 应用配置
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/serviceaccount.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

# 查看部署状态
kubectl -n kart-io get pods
kubectl -n kart-io get svc

# 查看日志
kubectl -n kart-io logs -f deployment/cluster-service

# 端口转发测试
kubectl -n kart-io port-forward svc/cluster-service 8080:80

# 测试
curl http://localhost:8080/health
```

### 3. Ingress 配置

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: cluster-service
  namespace: kart-io
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - api.k8s-agent.example.com
    secretName: cluster-service-tls
  rules:
  - host: api.k8s-agent.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: cluster-service
            port:
              number: 80
```

## 生产配置

### 1. 数据库优化

```yaml
database:
  max_open_conns: 100    # 根据负载调整
  max_idle_conns: 10     # 通常为 max_open_conns 的 10%
  conn_max_lifetime: 1h  # 连接最大生命周期
  conn_max_idle_time: 10m # 空闲连接超时
```

### 2. 日志配置

```yaml
logging:
  level: info           # 生产环境使用 info 或 warn
  format: json          # 生产环境必须用 json
  otlp_endpoint: "http://otel-collector:4317"  # 可选
```

### 3. 性能调优

```yaml
server:
  read_timeout: 30s     # 读超时
  write_timeout: 30s    # 写超时
  max_header_bytes: 1048576  # 1MB

# Rate limiting (代码中配置)
# middleware.RateLimitByIP(1000, 2000) // 每秒1000请求
```

## 监控和告警

### 1. Prometheus 指标

（待实现）服务将暴露 `/metrics` 端点。

### 2. 日志收集

使用 OTLP 或 Fluentd/Fluent Bit：

```yaml
# fluent-bit configmap
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config
data:
  fluent-bit.conf: |
    [SERVICE]
        Flush        5
        Daemon       Off
        Log_Level    info

    [INPUT]
        Name              tail
        Path              /var/log/containers/cluster-service*.log
        Parser            json
        Tag               kube.*

    [OUTPUT]
        Name   es
        Match  kube.*
        Host   elasticsearch
        Port   9200
```

### 3. 健康检查

```bash
# Kubernetes 健康检查已配置
# livenessProbe: /health
# readinessProbe: /health

# 外部监控
curl https://api.k8s-agent.example.com/health
```

## 故障排查

### 1. Pod 无法启动

```bash
# 查看 Pod 状态
kubectl -n kart-io describe pod <pod-name>

# 查看日志
kubectl -n kart-io logs <pod-name>

# 常见问题：
# - 配置文件错误：检查 ConfigMap
# - 数据库连接失败：检查 Secret 和网络
# - 权限问题：检查 ServiceAccount
```

### 2. 性能问题

```bash
# 查看资源使用
kubectl -n kart-io top pods

# 查看限流日志
kubectl -n kart-io logs deployment/cluster-service | grep "rate limit"

# 调整副本数
kubectl -n kart-io scale deployment cluster-service --replicas=5
```

### 3. 数据库连接池耗尽

```bash
# 增加连接池大小
# 编辑 ConfigMap
kubectl -n kart-io edit configmap cluster-service-config

# 重启服务
kubectl -n kart-io rollout restart deployment cluster-service
```

## 版本升级

### 滚动升级

```bash
# 构建新版本镜像
docker build -t k8s-agent/cluster-service:v1.1.0 .
docker push k8s-agent/cluster-service:v1.1.0

# 更新 Deployment
kubectl -n kart-io set image deployment/cluster-service \
  cluster-service=k8s-agent/cluster-service:v1.1.0

# 查看滚动升级状态
kubectl -n kart-io rollout status deployment/cluster-service

# 回滚（如果需要）
kubectl -n kart-io rollout undo deployment/cluster-service
```

### 金丝雀发布

```yaml
# 创建金丝雀 Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cluster-service-canary
spec:
  replicas: 1  # 只部署 1 个副本
  selector:
    matchLabels:
      app: cluster-service
      track: canary
  template:
    metadata:
      labels:
        app: cluster-service
        track: canary
        version: v1.1.0
    spec:
      # ... 使用新版本镜像
```

## 备份和恢复

### 数据库备份

```bash
# 备份
kubectl -n kart-io exec -it postgres-0 -- \
  pg_dump -U postgres cluster_db > backup.sql

# 恢复
kubectl -n kart-io exec -i postgres-0 -- \
  psql -U postgres cluster_db < backup.sql
```

### 配置备份

```bash
# 备份所有配置
kubectl -n kart-io get configmap,secret,deployment,service -o yaml > backup.yaml

# 恢复
kubectl apply -f backup.yaml
```

## 安全加固

1. **使用非 root 用户运行**（已配置）
2. **启用 HTTPS**（通过 Ingress + cert-manager）
3. **定期更新依赖**：`go get -u ./...`
4. **扫描镜像漏洞**：`trivy image k8s-agent/cluster-service:v1.0.0`
5. **限制 RBAC 权限**：最小权限原则
6. **启用网络策略**：限制 Pod 间通信

## 相关文档

- [K8S_API_IMPLEMENTATION.md](./K8S_API_IMPLEMENTATION.md) - 完整实现文档
- [API_QUICKSTART.md](./API_QUICKSTART.md) - API 快速指南
- [QUICKSTART.md](./QUICKSTART.md) - 开发快速启动
