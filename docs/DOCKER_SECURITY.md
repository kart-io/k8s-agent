# Docker 镜像安全最佳实践

本文档描述了项目中实施的 Docker 镜像安全措施和最佳实践。

## 实施的安全措施

### 1. 使用特定版本的基础镜像

❌ **不推荐**：
```dockerfile
FROM alpine:latest
```

✅ **推荐**：
```dockerfile
FROM alpine:3.20
```

**原因**：
- `latest` 标签不稳定，可能引入未知的安全问题
- 特定版本便于追踪和审计
- 可以针对特定版本应用安全补丁

### 2. 定期更新所有包

所有 Dockerfile 都包含：
```dockerfile
RUN apk add --no-cache --upgrade ca-certificates tzdata && \
    apk upgrade --no-cache && \
    rm -rf /var/cache/apk/* /tmp/*
```

这确保：
- 安装最新的安全补丁
- 清理包管理器缓存减小镜像大小
- 移除临时文件

### 3. 非 root 用户运行

所有服务都使用非 root 用户运行：

```dockerfile
# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# 切换到非 root 用户
USER appuser
```

**安全优势**：
- 限制容器内的权限
- 防止权限提升攻击
- 符合最小权限原则

### 4. 静态编译的二进制文件

```dockerfile
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o collect-agent \
    ./main.go
```

**优势**：
- 无需外部动态库依赖
- 减少攻击面
- 更小的镜像体积

### 5. 多阶段构建

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder
...

# Runtime stage
FROM alpine:3.20
COPY --from=builder /app/service .
```

**安全优势**：
- 最终镜像不包含构建工具
- 减小镜像体积
- 降低攻击面

### 6. 健康检查

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1
```

**优势**：
- 自动检测服务健康状态
- Kubernetes 可以自动重启不健康的容器
- 提高服务可用性

## 扫描漏洞

### 使用 Docker Scout（推荐）

```bash
# 扫描镜像
docker scout cves docker.io/aetherius/collect-agent:v1.0.0

# 快速查看
docker scout quickview docker.io/aetherius/collect-agent:v1.0.0

# 生成报告
docker scout cves --format sarif --output report.json docker.io/aetherius/collect-agent:v1.0.0
```

### 使用 Trivy

```bash
# 安装 Trivy
brew install aquasecurity/trivy/trivy

# 扫描镜像
trivy image docker.io/aetherius/collect-agent:v1.0.0

# 只显示高危和严重漏洞
trivy image --severity HIGH,CRITICAL docker.io/aetherius/collect-agent:v1.0.0

# 生成 JSON 报告
trivy image --format json --output report.json docker.io/aetherius/collect-agent:v1.0.0
```

### 使用 Grype

```bash
# 安装 Grype
brew install grype

# 扫描镜像
grype docker.io/aetherius/collect-agent:v1.0.0

# 只显示高危漏洞
grype docker.io/aetherius/collect-agent:v1.0.0 --fail-on high
```

## 镜像签名与验证

### 使用 Docker Content Trust

```bash
# 启用 Docker Content Trust
export DOCKER_CONTENT_TRUST=1

# 推送签名的镜像
docker push docker.io/aetherius/collect-agent:v1.0.0

# 验证镜像签名
docker trust inspect docker.io/aetherius/collect-agent:v1.0.0
```

### 使用 Cosign (Sigstore)

```bash
# 安装 Cosign
brew install cosign

# 生成密钥对
cosign generate-key-pair

# 签名镜像
cosign sign --key cosign.key docker.io/aetherius/collect-agent:v1.0.0

# 验证镜像
cosign verify --key cosign.pub docker.io/aetherius/collect-agent:v1.0.0
```

## 运行时安全

### 1. 使用只读根文件系统

在 Kubernetes 中：
```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: collect-agent
    image: docker.io/aetherius/collect-agent:v1.0.0
    securityContext:
      readOnlyRootFilesystem: true
      allowPrivilegeEscalation: false
      runAsNonRoot: true
      runAsUser: 65534
```

### 2. 删除不必要的 Capabilities

```yaml
securityContext:
  capabilities:
    drop:
      - ALL
    add:
      - NET_BIND_SERVICE  # 仅在需要时添加
```

### 3. 使用网络策略

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: collect-agent-netpol
spec:
  podSelector:
    matchLabels:
      app: collect-agent
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: agent-manager
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: nats
```

## 供应链安全

### 1. 使用 SBOM (Software Bill of Materials)

```bash
# 使用 Syft 生成 SBOM
syft docker.io/aetherius/collect-agent:v1.0.0 -o json > sbom.json

# 使用 Docker Scout 生成 SBOM
docker scout sbom docker.io/aetherius/collect-agent:v1.0.0 > sbom.spdx.json
```

### 2. 依赖项验证

```bash
# 检查 Go 依赖的漏洞
go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth

# 使用 govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## CI/CD 集成

### GitHub Actions 示例

```yaml
name: Security Scan

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Build image
        run: docker build -t test-image .
      
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: test-image
          format: 'sarif'
          output: 'trivy-results.sarif'
          severity: 'CRITICAL,HIGH'
      
      - name: Upload Trivy results to GitHub Security
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'
```

## 定期维护

### 1. 定期更新基础镜像

```bash
# 检查可用的 Alpine 版本
docker pull alpine:3.20
docker pull alpine:3.21  # 新版本发布时

# 更新 Dockerfile
sed -i 's/alpine:3.20/alpine:3.21/g' */Dockerfile

# 重新构建
make docker-build
```

### 2. 监控 CVE 数据库

订阅安全通知：
- [Alpine Security](https://alpinelinux.org/security/)
- [Go Security](https://pkg.go.dev/vuln/)
- [NIST NVD](https://nvd.nist.gov/)

### 3. 自动化安全扫描

设置定期扫描（例如每周一次）：

```bash
#!/bin/bash
# scan-all-images.sh

IMAGES=(
  "docker.io/aetherius/collect-agent:v1.0.0"
  "docker.io/aetherius/agent-manager:v1.0.0"
  "docker.io/aetherius/orchestrator-service:v1.0.0"
  "docker.io/aetherius/gateway-service:v1.0.0"
  "docker.io/aetherius/auth-service:v1.0.0"
)

for image in "${IMAGES[@]}"; do
  echo "Scanning $image..."
  trivy image --severity HIGH,CRITICAL "$image"
done
```

## 应急响应

### 当发现高危漏洞时：

1. **评估影响**
   - 确定受影响的服务
   - 评估漏洞的可利用性
   - 检查是否有已知的利用

2. **修复步骤**
   ```bash
   # 更新依赖
   cd collect-agent
   go get -u ./...
   go mod tidy
   
   # 重新构建
   make docker-build
   
   # 重新扫描
   trivy image docker.io/aetherius/collect-agent:v1.0.1
   
   # 推送修复版本
   docker push docker.io/aetherius/collect-agent:v1.0.1
   ```

3. **部署修复**
   ```bash
   # 更新 Kubernetes 部署
   kubectl set image deployment/collect-agent \
     collect-agent=docker.io/aetherius/collect-agent:v1.0.1
   
   # 验证部署
   kubectl rollout status deployment/collect-agent
   ```

## 参考资源

- [Docker Security Best Practices](https://docs.docker.com/develop/security-best-practices/)
- [OWASP Docker Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Docker_Security_Cheat_Sheet.html)
- [CIS Docker Benchmark](https://www.cisecurity.org/benchmark/docker)
- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/security-checklist/)

## Scratch 镜像说明

### 为什么使用 Scratch？

所有服务现在都使用 `FROM scratch` 作为运行时镜像：

**优势：**
- ✅ **零 OS 漏洞**：没有任何操作系统包，因此没有 CVE
- ✅ **极小体积**：最终镜像仅包含二进制文件
- ✅ **最小攻击面**：没有 shell、包管理器或其他工具
- ✅ **最佳安全性**：无法通过容器执行任何额外命令

**限制：**
- ❌ 无法使用 `docker exec` 进入容器（没有 shell）
- ❌ 无法在 Dockerfile 中使用 `HEALTHCHECK`（没有 curl/wget）
- ❌ 必须使用静态编译的二进制文件

### 健康检查的替代方案

由于 scratch 镜像没有 shell 或工具，健康检查必须在 Kubernetes 层面配置：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: collect-agent
spec:
  template:
    spec:
      containers:
      - name: collect-agent
        image: docker.io/aetherius/collect-agent:v1.0.0
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
```

### 调试 Scratch 容器

由于无法 exec 进入容器，使用以下方法调试：

#### 1. 查看日志
```bash
kubectl logs -f pod/collect-agent-xxx
```

#### 2. 使用 ephemeral 容器（Kubernetes 1.23+）
```bash
kubectl debug -it pod/collect-agent-xxx --image=busybox --target=collect-agent
```

#### 3. 临时切换到 debug 镜像
为开发环境创建一个带工具的版本：

```dockerfile
# Dockerfile.debug
FROM alpine:3.20

COPY --from=builder /app/collect-agent /collect-agent
RUN apk add --no-cache curl netcat-openbsd

ENTRYPOINT ["/collect-agent"]
```

### 切换回 Alpine（如果需要）

如果需要 shell 或调试工具，可以切换回 Alpine：

```dockerfile
# 替换 FROM scratch 为
FROM alpine:3.20

# 添加必要的工具
RUN apk add --no-cache ca-certificates curl && \
    apk upgrade --no-cache && \
    rm -rf /var/cache/apk/*
```

但这会引入 OS 漏洞，只建议在开发环境使用。

### 最佳实践

1. **生产环境**：使用 `scratch` 获得最佳安全性
2. **开发环境**：使用 `alpine` 或 `distroless/debug` 便于调试
3. **CI/CD**：在部署前扫描两个版本的镜像
4. **监控**：依赖 Kubernetes 的健康检查和日志

