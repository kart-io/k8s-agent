# Gateway Service

API 网关服务，作为所有后端服务的统一入口，提供路由转发、认证授权、限流、CORS 等功能。

支持两种部署模式：
- **Go 网关** - 基于 Gin 的轻量级网关
- **Traefik 网关** - 功能强大的云原生边缘路由器（推荐生产环境使用）

## 功能特性

### 1. 统一入口
- 所有后端服务通过网关统一访问
- 服务路由自动转发
- 请求/响应透明代理

### 2. 认证授权
- JWT Token 验证
- 统一认证拦截
- 用户信息传递

### 3. 限流保护
- 基于 Redis 的分布式限流
- 支持本地限流（无 Redis 时）
- 按用户/IP 限流
- 可配置限流策略

### 4. 跨域支持 (CORS)
- 灵活的跨域配置
- 支持多源配置
- 预检请求处理

### 5. 服务健康检查
- 网关自身健康检查
- 后端服务健康检查
- 自动服务发现

### 6. 监控指标
- 请求统计
- 成功/失败率
- 平均延迟
- 服务级别指标

## 目录结构

```
gateway-service/
├── cmd/server/          # 主程序入口
│   └── main.go
├── internal/
│   ├── handler/         # HTTP 处理器
│   │   ├── health.go    # 健康检查
│   │   └── metrics.go   # 监控指标
│   ├── middleware/      # 中间件
│   │   ├── auth.go      # JWT 认证
│   │   ├── cors.go      # 跨域处理
│   │   └── ratelimit.go # 限流
│   ├── proxy/           # 代理处理
│   │   └── proxy.go     # 请求转发
│   └── router/          # 路由配置
│       └── router.go
├── pkg/types/           # 类型定义
│   └── types.go
├── configs/             # 配置文件
│   └── config.yaml
├── go.mod
├── Makefile
└── README.md
```

## 配置说明

### 服务配置 (configs/config.yaml)

```yaml
server:
  port: 8080              # 网关端口
  mode: debug             # 运行模式: debug, release, test

redis:
  host: localhost         # Redis 地址
  port: 6379

jwt:
  secret: "your-secret"   # JWT 密钥（生产环境必须修改）

rate_limit:
  enabled: true           # 是否启用限流
  requests_per_second: 100
  burst: 200

cors:
  enabled: true
  allow_origins:
    - "http://localhost:3000"

services:                 # 后端服务配置
  auth:
    url: http://localhost:8090
    timeout: 30s
  agent_manager:
    url: http://localhost:8081
    timeout: 30s
```

## API 路由

### 网关自身 API

```
GET  /health                        # 网关健康检查
GET  /api/v1/health/services        # 所有服务健康状态
GET  /api/v1/health/services/:name  # 单个服务健康状态
GET  /metrics                       # 监控指标
```

### 代理路由

所有 `/api/v1/*` 的请求会根据配置转发到对应的后端服务：

**认证服务** (不需要认证)
```
POST /api/v1/auth/login       -> auth-service
POST /api/v1/auth/logout      -> auth-service
GET  /api/v1/auth/me          -> auth-service
GET  /api/v1/auth/menus       -> auth-service
```

**Agent 管理** (需要认证)
```
GET    /api/v1/agents         -> agent-manager
GET    /api/v1/clusters       -> agent-manager
GET    /api/v1/events         -> agent-manager
POST   /api/v1/commands       -> agent-manager
```

**用户管理** (需要认证)
```
GET    /api/v1/users          -> auth-service
POST   /api/v1/users          -> auth-service
PUT    /api/v1/users/:id      -> auth-service
DELETE /api/v1/users/:id      -> auth-service
```

**工作流** (需要认证)
```
GET    /api/v1/workflows      -> orchestrator-service
POST   /api/v1/workflows      -> orchestrator-service
```

## 快速开始

### 方案一：使用 Traefik（推荐）

Traefik 是一个现代化的云原生边缘路由器，提供更强大的功能和更好的性能。

```bash
cd gateway-service/deployments

# 启动完整的服务栈
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f traefik
```

访问地址：
- API 入口: `http://localhost` (端口 80)
- Traefik Dashboard: `http://localhost:8080`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3001` (admin/admin)

详细使用请查看 [Traefik 部署指南](deployments/TRAEFIK_GUIDE.md)

### 方案二：使用 Go 网关

适合开发环境快速测试。

#### 1. 初始化依赖

```bash
cd gateway-service
make deps
```

#### 2. 配置服务

编辑 `configs/config.yaml`，配置后端服务地址：

```yaml
services:
  auth:
    url: http://localhost:8090    # auth-service 地址
  agent_manager:
    url: http://localhost:8081    # agent-manager 地址
  orchestrator:
    url: http://localhost:8082    # orchestrator-service 地址
```

#### 3. 运行服务

```bash
make run
```

或编译后运行：

```bash
make build
./bin/gateway-service
```

## 使用示例

### 1. 通过网关登录

```bash
# 原来直接调用 auth-service
curl -X POST http://localhost:8090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 现在通过网关调用
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### 2. 使用 Token 访问受保护资源

```bash
# 获取 agent 列表
curl -X GET http://localhost:8080/api/v1/agents \
  -H "Authorization: Bearer <your-token>"
```

### 3. 检查服务健康状态

```bash
# 网关健康检查
curl http://localhost:8080/health

# 所有后端服务健康检查
curl http://localhost:8080/api/v1/health/services

# 单个服务健康检查
curl http://localhost:8080/api/v1/health/services/auth
```

### 4. 查看监控指标

```bash
curl http://localhost:8080/metrics
```

## 前端集成

### 修改 API 基础地址

在前端项目中，将所有 API 请求地址改为网关地址：

```javascript
// 原来
const API_BASE_URL = 'http://localhost:8081/api/v1';

// 修改为网关地址
const API_BASE_URL = 'http://localhost:8080/api/v1';
```

### Axios 配置示例

```javascript
import axios from 'axios';

const request = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  timeout: 30000,
});

// 请求拦截器 - 添加 token
request.interceptors.request.use(config => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 响应拦截器
request.interceptors.response.use(
  response => response.data,
  error => {
    if (error.response?.status === 401) {
      // 未认证，跳转登录
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default request;
```

## 架构优势

### 1. 统一入口
- 所有请求通过网关，便于管理和监控
- 前端只需要知道网关地址

### 2. 安全性
- 统一的认证授权
- 后端服务不直接暴露
- 支持限流防护

### 3. 灵活性
- 后端服务地址变更，只需修改网关配置
- 支持服务版本控制
- 支持灰度发布

### 4. 可扩展
- 易于添加新的中间件
- 支持插件化扩展
- 支持服务治理

## 部署建议

### 开发环境

```bash
# 本地运行，使用默认配置
make run
```

### 生产环境

1. 修改配置文件

```yaml
server:
  mode: release
  port: 8080

jwt:
  secret: "your-production-secret-key"  # 使用强密钥

rate_limit:
  enabled: true
  requests_per_second: 1000
  burst: 2000

redis:
  host: redis.production.com
  password: "redis-password"
```

2. 编译运行

```bash
make build
./bin/gateway-service
```

3. 使用 systemd 管理

```ini
[Unit]
Description=K8s Agent Gateway Service
After=network.target

[Service]
Type=simple
User=gateway
WorkingDirectory=/opt/gateway-service
ExecStart=/opt/gateway-service/bin/gateway-service
Restart=always

[Install]
WantedBy=multi-user.target
```

## 监控和日志

### 日志查看

```bash
# 查看实时日志
tail -f logs/gateway.log
```

### 监控指标

访问 `http://localhost:8080/metrics` 查看：

- 总请求数
- 成功/失败率
- 平均延迟
- 运行时间

## 常见问题

### 1. CORS 错误

确保在 `config.yaml` 中配置了正确的前端地址：

```yaml
cors:
  allow_origins:
    - "http://localhost:3000"  # 添加你的前端地址
```

### 2. 503 服务不可用

检查后端服务是否正常运行，查看健康检查：

```bash
curl http://localhost:8080/api/v1/health/services
```

### 3. 401 未授权

确保请求头中包含有效的 JWT Token：

```
Authorization: Bearer <token>
```

## 开发计划

- [x] 基础路由转发
- [x] JWT 认证
- [x] CORS 支持
- [x] 限流保护
- [x] 健康检查
- [x] 监控指标
- [ ] 请求日志记录
- [ ] API 版本管理
- [ ] 服务熔断
- [ ] 负载均衡
- [ ] WebSocket 支持
- [ ] gRPC 网关
