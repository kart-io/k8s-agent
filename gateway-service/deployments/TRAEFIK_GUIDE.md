# Traefik 网关部署指南

## 概述

本项目使用 Traefik 作为 API 网关，提供以下功能：

- **反向代理** - 统一入口，路由到各个后端服务
- **负载均衡** - 自动负载均衡
- **SSL/TLS** - HTTPS 支持和自动证书管理
- **限流** - 防止服务过载
- **CORS** - 跨域请求处理
- **健康检查** - 自动服务健康检测
- **监控** - Prometheus 指标导出

## 架构图

```
Internet
    ↓
Traefik (80/443)
    ├→ /api/v1/auth/*        → Auth Service (8090)
    ├→ /api/v1/agents/*      → Agent Manager (8081)
    ├→ /api/v1/clusters/*    → Agent Manager (8081)
    ├→ /api/v1/workflows/*   → Orchestrator (8082)
    ├→ /api/v1/reasoning/*   → Reasoning Service (8083)
    └→ /health               → Gateway Service (8080)
```

## 快速开始

### 1. 使用 Docker Compose 部署

```bash
cd gateway-service/deployments

# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f traefik

# 停止服务
docker-compose down
```

### 2. 访问服务

**Traefik Dashboard**
```
http://localhost:8080
```

**API 访问**
```bash
# 认证服务
curl http://localhost/api/v1/auth/login

# Agent 管理
curl http://localhost/api/v1/agents

# 健康检查
curl http://localhost/health
```

**Prometheus**
```
http://localhost:9090
```

**Grafana**
```
http://localhost:3001
用户名: admin
密码: admin
```

## 配置说明

### Traefik 静态配置 (traefik.yml)

```yaml
# 入口点
entryPoints:
  web:
    address: ":80"
  websecure:
    address: ":443"

# 提供者
providers:
  docker:
    exposedByDefault: false
  file:
    directory: /etc/traefik/dynamic
```

### 中间件配置 (dynamic/middlewares.yml)

**限流**
```yaml
rate-limit:
  rateLimit:
    average: 100      # 每秒平均请求数
    burst: 200        # 突发请求数
```

**CORS**
```yaml
cors-headers:
  headers:
    accessControlAllowOriginList:
      - "http://localhost:3000"
    accessControlAllowMethods:
      - GET
      - POST
      - PUT
      - DELETE
```

### 路由配置 (dynamic/routes.yml)

```yaml
routers:
  auth-service-router:
    rule: "PathPrefix(`/api/v1/auth`)"
    service: auth-service
    middlewares:
      - gateway-chain
```

## 中间件说明

### 1. gateway-chain (标准链)

包含以下中间件：
- CORS 处理
- 响应压缩
- 限流保护
- 自动重试
- 超时控制

### 2. secure-chain (安全链)

额外包含：
- 安全响应头
- HTTPS 重定向
- XSS 防护

### 3. 自定义中间件

**添加 IP 白名单**
```yaml
middlewares:
  ip-whitelist:
    ipWhiteList:
      sourceRange:
        - "192.168.1.0/24"
        - "10.0.0.0/8"
```

**添加基础认证**
```yaml
middlewares:
  basic-auth:
    basicAuth:
      users:
        - "admin:$apr1$H6uskkkW$..."
```

## 路由规则

### 基于路径前缀

```yaml
rule: "PathPrefix(`/api/v1/auth`)"
```

### 基于主机名

```yaml
rule: "Host(`api.example.com`)"
```

### 组合规则

```yaml
rule: "Host(`api.example.com`) && PathPrefix(`/api/v1/auth`)"
```

### 优先级

```yaml
priority: 100  # 数字越大优先级越高
```

## 负载均衡

### 轮询（默认）

```yaml
services:
  auth-service:
    loadBalancer:
      servers:
        - url: "http://auth-service-1:8090"
        - url: "http://auth-service-2:8090"
```

### 加权轮询

```yaml
loadBalancer:
  servers:
    - url: "http://auth-service-1:8090"
      weight: 3
    - url: "http://auth-service-2:8090"
      weight: 1
```

## 健康检查

```yaml
services:
  auth-service:
    loadBalancer:
      healthCheck:
        path: /health
        interval: 30s
        timeout: 5s
```

## SSL/TLS 配置

### Let's Encrypt 自动证书

```yaml
certificatesResolvers:
  letsencrypt:
    acme:
      email: admin@example.com
      storage: /letsencrypt/acme.json
      httpChallenge:
        entryPoint: web
```

### 使用路由器

```yaml
routers:
  auth-service-router:
    rule: "Host(`api.example.com`)"
    tls:
      certResolver: letsencrypt
```

## 监控和日志

### 访问日志

```yaml
accessLog:
  filePath: /var/log/traefik/access.log
  format: json
```

### Prometheus 指标

访问 `http://traefik:8080/metrics` 获取指标

常用指标：
- `traefik_service_requests_total` - 总请求数
- `traefik_service_request_duration_seconds` - 请求延迟
- `traefik_service_requests_bytes_total` - 请求字节数

## 生产环境建议

### 1. 安全配置

```yaml
# 禁用不安全的仪表板访问
api:
  dashboard: true
  insecure: false

# 启用 HTTPS
entryPoints:
  web:
    http:
      redirections:
        entryPoint:
          to: websecure
          scheme: https
```

### 2. 限流策略

```yaml
rate-limit:
  rateLimit:
    average: 1000
    burst: 2000
    period: 1s
```

### 3. 超时配置

```yaml
timeout:
  responseHeaderTimeout: 30s
  idleConnTimeout: 90s
  dialTimeout: 30s
```

### 4. 重试策略

```yaml
retry:
  attempts: 3
  initialInterval: 100ms
```

## 故障排查

### 查看 Traefik 日志

```bash
docker-compose logs -f traefik
```

### 查看路由配置

访问 Traefik Dashboard: `http://localhost:8080`

### 测试路由

```bash
# 测试认证路由
curl -v http://localhost/api/v1/auth/login

# 测试健康检查
curl http://localhost/health
```

### 常见问题

**1. 404 Not Found**
- 检查路由规则是否正确
- 确认服务是否运行
- 查看 Traefik Dashboard 中的路由配置

**2. 502 Bad Gateway**
- 检查后端服务是否健康
- 查看服务健康检查状态
- 确认服务地址和端口正确

**3. CORS 错误**
- 检查 CORS 中间件配置
- 确认允许的源列表包含前端地址
- 查看浏览器控制台详细错误

**4. 限流触发**
- 检查限流配置是否过严
- 查看访问日志
- 调整 average 和 burst 参数

## 扩展配置

### 添加新服务

1. 在 `dynamic/routes.yml` 添加路由：

```yaml
routers:
  new-service-router:
    rule: "PathPrefix(`/api/v1/newservice`)"
    service: new-service
    middlewares:
      - gateway-chain

services:
  new-service:
    loadBalancer:
      servers:
        - url: "http://new-service:8084"
      healthCheck:
        path: /health
```

2. 在 `docker-compose.yml` 添加服务：

```yaml
new-service:
  image: your-new-service:latest
  networks:
    - k8s-agent-network
  labels:
    - "traefik.enable=false"
```

### 添加自定义中间件

在 `dynamic/middlewares.yml` 中添加：

```yaml
middlewares:
  custom-header:
    headers:
      customRequestHeaders:
        X-Custom-Header: "custom-value"
```

然后在路由中使用：

```yaml
routers:
  service-router:
    middlewares:
      - custom-header
      - gateway-chain
```

## 参考资源

- [Traefik 官方文档](https://doc.traefik.io/traefik/)
- [Docker Provider](https://doc.traefik.io/traefik/providers/docker/)
- [File Provider](https://doc.traefik.io/traefik/providers/file/)
- [Middlewares](https://doc.traefik.io/traefik/middlewares/overview/)
