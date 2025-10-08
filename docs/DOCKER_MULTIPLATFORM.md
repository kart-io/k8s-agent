# Docker 多平台构建指南

本项目支持多平台 Docker 镜像构建，可以同时为 `linux/amd64` 和 `linux/arm64` 架构构建镜像。

## 前置要求

1. **Docker Buildx**
   ```bash
   # 检查 buildx 是否已安装
   docker buildx version
   
   # 如果未安装，安装 buildx
   docker buildx install
   ```

2. **创建 Builder**
   ```bash
   # 创建并使用新的 builder
   docker buildx create --name k8s-agent-builder --use --bootstrap
   
   # 验证 builder
   docker buildx inspect k8s-agent-builder
   ```

## 构建方式

### 方式一：使用统一脚本

项目提供了统一的多平台构建脚本 `scripts/docker-buildx.sh`：

```bash
# 基本用法
./scripts/docker-buildx.sh <service-name> [options]

# 构建 collect-agent (本地加载)
./scripts/docker-buildx.sh collect-agent -v v1.0.0

# 构建并推送到 registry
./scripts/docker-buildx.sh agent-manager -v v1.1.0 --push

# 只构建 amd64 架构
./scripts/docker-buildx.sh gateway-service -v latest --platforms linux/amd64
```

**支持的服务:**
- `collect-agent`
- `agent-manager`
- `orchestrator-service`
- `gateway-service`
- `auth-service`

**可用选项:**
- `-v, --version VERSION` - 镜像版本 (默认: v1.0.0)
- `-p, --platforms PLATFORMS` - 目标平台 (默认: linux/amd64,linux/arm64)
- `-r, --registry REGISTRY` - Docker registry (默认: docker.io)
- `-n, --namespace NAMESPACE` - Docker namespace (默认: aetherius)
- `--push` - 构建后推送到 registry

### 方式二：使用根目录 Makefile

```bash
# 构建所有服务的多平台镜像 (本地加载)
make docker-buildx VERSION=v1.0.0

# 构建并推送所有服务的多平台镜像
make docker-buildx-push VERSION=v1.0.0
```

### 方式三：使用各服务的 Makefile

每个服务都有独立的 Makefile 支持多平台构建：

```bash
# collect-agent
cd collect-agent
make docker-buildx          # 构建多平台镜像
make docker-buildx-push     # 构建并推送

# agent-manager
cd agent-manager
make docker-buildx
make docker-buildx-push

# orchestrator-service
cd orchestrator-service
make docker-buildx
make docker-buildx-push

# gateway-service
cd gateway-service
make docker-buildx
make docker-buildx-push

# auth-service
cd auth-service
make docker-buildx
make docker-buildx-push
```

## Dockerfile 特性

所有 Dockerfile 都支持以下特性：

### 1. 多平台构建参数

```dockerfile
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -o app .
```

### 2. 多阶段构建

- **构建阶段**: 使用 `golang:1.21-alpine` 编译二进制
- **运行阶段**: 使用 `alpine:latest` 最小化镜像

### 3. 安全特性

- 非 root 用户运行 (appuser)
- 只包含必要的运行时依赖
- 静态链接的二进制文件 (CGO_ENABLED=0)

### 4. 健康检查

所有服务都包含健康检查配置：

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
```

## 支持的平台

默认支持以下平台组合：

- `linux/amd64` - x86_64 架构 (Intel/AMD 64位)
- `linux/arm64` - ARM64 架构 (Apple Silicon, AWS Graviton等)

可以根据需要添加更多平台：

```bash
# 添加 ARM v7 支持
./scripts/docker-buildx.sh collect-agent -v v1.0.0 \
  --platforms linux/amd64,linux/arm64,linux/arm/v7
```

## 镜像命名规范

```
<registry>/<namespace>/<service>:<version>
```

**默认配置:**
- Registry: `docker.io`
- Namespace: `aetherius`

**示例:**
- `docker.io/aetherius/collect-agent:v1.0.0`
- `docker.io/aetherius/agent-manager:latest`

## 常见问题

### 1. buildx 构建卡住

```bash
# 重启 Docker Desktop
# 或者重新创建 builder
docker buildx rm k8s-agent-builder
docker buildx create --name k8s-agent-builder --use --bootstrap
```

### 2. 无法推送镜像

```bash
# 登录 Docker registry
docker login

# 验证凭证
docker info | grep Username
```

### 3. 跨平台构建缓慢

这是正常现象，因为需要使用 QEMU 模拟不同架构。可以：
- 使用本地架构的原生构建器
- 使用云端构建服务
- 配置构建缓存

### 4. --load 和 --push 的区别

- `--load`: 将镜像加载到本地 Docker
  - 限制：只能加载单一平台
  - 用途：本地测试
  
- `--push`: 推送到远程 registry
  - 支持：多平台同时推送
  - 用途：生产部署

## 性能优化

### 1. 使用构建缓存

```bash
# 配置 buildx 使用 registry 缓存
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --cache-from type=registry,ref=aetherius/collect-agent:buildcache \
  --cache-to type=registry,ref=aetherius/collect-agent:buildcache,mode=max \
  -t aetherius/collect-agent:latest \
  --push .
```

### 2. 并行构建

根目录 Makefile 已配置并行构建：

```bash
# 同时构建所有服务
make -j4 docker-buildx
```

## CI/CD 集成

### GitHub Actions 示例

```yaml
name: Build Multi-Platform Images

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2
      
      - name: Login to Docker Hub
        uses: docker/login-action@v2
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}
      
      - name: Build and Push
        run: |
          make docker-buildx-push VERSION=${{ github.ref_name }}
```

## 验证镜像

### 检查镜像平台

```bash
# 查看镜像支持的平台
docker buildx imagetools inspect aetherius/collect-agent:v1.0.0

# 输出示例:
# MediaType: application/vnd.docker.distribution.manifest.list.v2+json
# Digest:    sha256:abc...
#
# Manifests:
#   Name:      aetherius/collect-agent:v1.0.0@sha256:def...
#   MediaType: application/vnd.docker.distribution.manifest.v2+json
#   Platform:  linux/amd64
#
#   Name:      aetherius/collect-agent:v1.0.0@sha256:ghi...
#   MediaType: application/vnd.docker.distribution.manifest.v2+json
#   Platform:  linux/arm64
```

### 在不同平台运行

```bash
# 强制使用特定平台
docker run --platform linux/amd64 aetherius/collect-agent:v1.0.0
docker run --platform linux/arm64 aetherius/collect-agent:v1.0.0
```

## 最佳实践

1. **版本标签**: 总是使用明确的版本号，避免只用 `latest`
2. **测试**: 在推送前先用 `--load` 本地测试
3. **缓存**: 利用 layer 缓存加速构建
4. **安全**: 定期更新基础镜像
5. **大小**: 使用 Alpine 等精简基础镜像
6. **文档**: 记录每个镜像的用途和配置

## 相关资源

- [Docker Buildx 文档](https://docs.docker.com/buildx/working-with-buildx/)
- [多平台构建最佳实践](https://docs.docker.com/build/building/multi-platform/)
- [Dockerfile 参考](https://docs.docker.com/engine/reference/builder/)
