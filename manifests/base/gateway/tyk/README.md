# Tyk API Gateway 集成指南

本文档提供 Tyk API Gateway 在 k8s-agent 项目中的完整集成指南。

## 目录

- [简介](#简介)
- [架构概览](#架构概览)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [API 定义](#api-定义)
- [认证与授权](#认证与授权)
- [限流与配额](#限流与配额)
- [监控与分析](#监控与分析)
- [故障排查](#故障排查)
- [最佳实践](#最佳实践)

## 简介

Tyk 是一个开源的企业级 API 网关,提供以下核心功能:

- **多协议支持**: REST, GraphQL, TCP, gRPC
- **高性能**: 低延迟,单 CPU 可处理数千 RPS
- **认证方式**: JWT, OIDC, OAuth2, Bearer Token, 基础认证
- **流量控制**: 限流、配额、熔断
- **可扩展性**: 插件架构,支持自定义中间件
- **Kubernetes 原生**: 支持声明式 API 和 Operator

## 架构概览

```
┌─────────────┐
│   客户端     │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│        Tyk Gateway (Port 8080)      │
│  - 路由转发                          │
│  - JWT 认证                          │
│  - 限流/配额                         │
│  - CORS 处理                         │
└──────┬──────────────────────────────┘
       │
       ├──────────────┬──────────────┐
       ▼              ▼              ▼
┌────────────┐  ┌──────────┐  ┌──────────┐
│Auth Service│  │  Agent   │  │  Other   │
│ (Port 8090)│  │ Manager  │  │ Services │
└────────────┘  │(Port 8081)│  └──────────┘
                └──────────┘

┌─────────────┐  ┌──────────────┐  ┌──────────┐
│   Redis     │  │ Tyk Dashboard│  │PostgreSQL│
│  (Session)  │  │ (Port 3000)  │  │(Analytics)│
└─────────────┘  └──────────────┘  └──────────┘
```

## 快速开始

### 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- 至少 2GB 可用内存

### 启动服务

```bash
cd gateway-service/deployments/tyk

# 方式1: 使用启动脚本(推荐)
./start.sh

# 方式2: 手动启动
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f tyk-gateway
```

### 验证安装

```bash
# 检查 Tyk Gateway
curl http://localhost:8080/hello

# 预期输出
{
  "status": "pass",
  "version": "v5.2.0",
  "description": "Tyk GW"
}

# 检查 Tyk Dashboard
curl http://localhost:3000/hello

# 检查 API 路由
curl http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

## 配置说明

### 主配置文件 (tyk.conf)

主要配置项说明:

```json
{
  "listen_port": 8080,              // Gateway 监听端口
  "secret": "your-secret-key",      // API 密钥
  "app_path": "/opt/tyk-gateway/apps/",  // API 定义路径
  "storage": {                      // Redis 配置
    "type": "redis",
    "host": "redis",
    "port": 6379
  },
  "enable_analytics": true,         // 启用分析
  "policies": {                     // 策略配置
    "policy_source": "file",
    "policy_record_name": "/opt/tyk-gateway/policies/policies.json"
  }
}
```

### 环境变量

创建 `.env` 文件(从 `.env.example` 复制):

```bash
cp .env.example .env
```

主要环境变量:

```bash
# Gateway 密钥(生产环境必须修改)
TYK_GW_SECRET=your-production-secret-key

# Redis 连接
TYK_GW_STORAGE_HOST=redis
TYK_GW_STORAGE_PORT=6379

# 后端服务地址
AUTH_SERVICE_URL=http://auth-service:8090
AGENT_MANAGER_URL=http://agent-manager:8081
```

## API 定义

### 创建 API 定义

API 定义文件位于 `apps/` 目录,采用 JSON 格式。

#### Auth Service API (apps/auth-service.json)

```json
{
  "name": "Auth Service API",
  "api_id": "auth-service",
  "enable_jwt": true,
  "jwt_signing_method": "hmac",
  "proxy": {
    "listen_path": "/api/v1/auth/",
    "target_url": "http://auth-service:8090/api/v1/auth/"
  },
  "version_data": {
    "versions": {
      "Default": {
        "extended_paths": {
          "ignored": [
            {
              "path": "/api/v1/auth/login",
              "method_actions": {"POST": {"action": "no_action"}}
            }
          ]
        }
      }
    }
  }
}
```

**关键配置**:

- `listen_path`: Gateway 监听路径
- `target_url`: 后端服务地址
- `enable_jwt`: 启用 JWT 认证
- `ignored`: 不需要认证的路径(如登录接口)

#### Agent Manager API (apps/agent-manager.json)

```json
{
  "name": "Agent Manager API",
  "api_id": "agent-manager",
  "enable_jwt": true,
  "proxy": {
    "listen_path": "/api/v1/agents",
    "target_url": "http://agent-manager:8081/api/v1/agents"
  },
  "cache_options": {
    "enable_cache": true,
    "cache_timeout": 60
  }
}
```

### 重载 API 定义

修改 API 定义后需要重载:

```bash
# 方式1: 热重载(推荐)
curl -H "x-tyk-authorization: your-secret" \
  http://localhost:8080/tyk/reload/group

# 方式2: 重启 Gateway
docker-compose restart tyk-gateway
```

## 认证与授权

### JWT 认证

Tyk 原生支持 JWT 认证,无需额外配置。

#### JWT 配置

```json
{
  "enable_jwt": true,
  "jwt_signing_method": "hmac",
  "jwt_identity_base_field": "sub",
  "jwt_policy_field_name": "pol"
}
```

#### 使用 JWT Token

```bash
# 1. 登录获取 Token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r '.token')

# 2. 使用 Token 访问受保护资源
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/agents
```

### 策略(Policies)

策略定义访问权限和限流规则。

#### 策略文件 (policies/policies.json)

```json
{
  "default": {
    "id": "default",
    "name": "Default Policy",
    "rate": 1000,              // 每分钟 1000 次请求
    "per": 60,
    "quota_max": 10000,        // 每小时 10000 次配额
    "access_rights": {
      "auth-service": {
        "api_id": "auth-service",
        "versions": ["Default"]
      },
      "agent-manager": {
        "api_id": "agent-manager",
        "versions": ["Default"]
      }
    }
  }
}
```

### 白名单路径

某些端点不需要认证(如登录、注册):

```json
{
  "extended_paths": {
    "ignored": [
      {
        "path": "/api/v1/auth/login",
        "method_actions": {
          "POST": {"action": "no_action"}
        }
      }
    ]
  }
}
```

## 限流与配额

### 全局限流

在 API 定义中配置:

```json
{
  "global_rate_limit": {
    "rate": 1000,    // 每分钟请求数
    "per": 60
  }
}
```

### 策略级限流

在策略文件中配置:

```json
{
  "rate": 1000,           // 请求速率
  "per": 60,              // 时间窗口(秒)
  "quota_max": 10000,     // 配额上限
  "quota_renewal_rate": 3600  // 配额重置周期(秒)
}
```

### 端点级限流

针对特定端点:

```json
{
  "extended_paths": {
    "rate_limit": [
      {
        "path": "/api/v1/agents",
        "method": "GET",
        "rate": 100,
        "per": 60
      }
    ]
  }
}
```

## 监控与分析

### Tyk Dashboard

访问 `http://localhost:3000` 打开 Dashboard。

**功能**:

- API 管理界面
- 实时流量监控
- 分析报表
- 密钥管理

### Tyk Pump

Tyk Pump 将分析数据导出到多种后端。

#### 配置 (pump.conf)

```json
{
  "pumps": {
    "postgres": {
      "type": "postgres",
      "meta": {
        "connection_string": "host=postgres port=5432 ..."
      }
    },
    "prometheus": {
      "type": "prometheus",
      "meta": {
        "listen_address": ":9090",
        "path": "/metrics"
      }
    }
  }
}
```

### Prometheus 指标

访问 `http://localhost:9090/metrics` 查看指标:

- `tyk_http_requests_total` - 总请求数
- `tyk_http_latency` - 请求延迟
- `tyk_http_status` - HTTP 状态码分布

### 查询日志

```bash
# Gateway 日志
docker-compose logs -f tyk-gateway

# Pump 日志
docker-compose logs -f tyk-pump

# 所有服务日志
docker-compose logs -f
```

## 故障排查

### 常见问题

#### 1. Gateway 无法启动

**错误**: `Error connecting to Redis`

**解决**:

```bash
# 检查 Redis 状态
docker-compose ps redis

# 查看 Redis 日志
docker-compose logs redis

# 重启 Redis
docker-compose restart redis
```

#### 2. API 返回 404

**原因**: API 定义未加载

**解决**:

```bash
# 检查 API 定义文件
ls -la apps/

# 重载 API
curl -H "x-tyk-authorization: your-secret" \
  http://localhost:8080/tyk/reload/group

# 查看已加载的 API
curl -H "x-tyk-authorization: your-secret" \
  http://localhost:8080/tyk/apis
```

#### 3. JWT 认证失败

**错误**: `Authorization field missing`

**解决**:

```bash
# 检查请求头
curl -v -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/agents

# 验证 Token 格式
echo $TOKEN | cut -d'.' -f2 | base64 -d | jq .

# 检查 JWT 配置
cat apps/auth-service.json | jq '.enable_jwt'
```

#### 4. 后端服务不可达

**错误**: `Upstream connect error`

**解决**:

```bash
# 检查后端服务状态
docker-compose ps auth-service agent-manager

# 测试后端连接
docker-compose exec tyk-gateway \
  curl http://auth-service:8090/health

# 检查网络
docker network inspect tyk_tyk-network
```

### 调试模式

启用详细日志:

```json
{
  "log_level": "debug",
  "use_sentry": false
}
```

重启 Gateway:

```bash
docker-compose restart tyk-gateway
docker-compose logs -f tyk-gateway
```

## 最佳实践

### 安全

1. **修改默认密钥**

   ```bash
   # 生成强密钥
   openssl rand -hex 32

   # 更新 .env 文件
   TYK_GW_SECRET=your-new-secret-key
   ```

2. **启用 HTTPS**

   ```json
   {
     "http_server_options": {
       "use_ssl": true,
       "certificates": [
         {
           "domain_name": "*.example.com",
           "cert_file": "/path/to/cert.pem",
           "key_file": "/path/to/key.pem"
         }
       ]
     }
   }
   ```

3. **IP 白名单**

   ```json
   {
     "enable_ip_whitelisting": true,
     "allowed_ips": ["192.168.1.0/24"]
   }
   ```

### 性能优化

1. **启用缓存**

   ```json
   {
     "cache_options": {
       "enable_cache": true,
       "cache_timeout": 60,
       "cache_all_safe_requests": true
     }
   }
   ```

2. **连接池优化**

   ```json
   {
     "max_idle_connections_per_host": 500,
     "max_idle_connections": 100
   }
   ```

3. **Redis 优化**

   ```json
   {
     "storage": {
       "optimisation_max_idle": 2000,
       "optimisation_max_active": 4000
     }
   }
   ```

### 高可用

1. **多实例部署**

   ```yaml
   tyk-gateway:
     deploy:
       replicas: 3
   ```

2. **Redis 集群**

   ```json
   {
     "storage": {
       "enable_cluster": true,
       "hosts": {
         "redis-1:6379": 1,
         "redis-2:6379": 1,
         "redis-3:6379": 1
       }
     }
   }
   ```

3. **健康检查**

   ```yaml
   healthcheck:
     test: ["CMD", "curl", "-f", "http://localhost:8080/hello"]
     interval: 10s
     timeout: 5s
     retries: 5
   ```

### 监控告警

1. **集成 Prometheus**

   ```yaml
   - job_name: 'tyk'
     static_configs:
       - targets: ['tyk-pump:9090']
   ```

2. **配置 Grafana 看板**

   - 导入 Tyk 官方看板: Dashboard ID `12900`

3. **设置告警规则**

   ```yaml
   - alert: HighErrorRate
     expr: rate(tyk_http_status{code="5xx"}[5m]) > 0.05
     for: 5m
   ```

## 参考资源

- [Tyk 官方文档](https://tyk.io/docs/)
- [Tyk GitHub](https://github.com/TykTechnologies/tyk)
- [Tyk 社区论坛](https://community.tyk.io/)
- [Docker Hub - Tyk](https://hub.docker.com/u/tykio)

## 下一步

- [配置自定义中间件](./middleware/README.md)
- [集成 Kubernetes](./kubernetes/README.md)
- [性能调优指南](./performance-tuning.md)
- [生产部署清单](./production-checklist.md)
