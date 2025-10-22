# API 定义说明

本目录包含所有后端服务的 API 定义文件。

## 目录结构

```
apps/
├── README.md              # 本文件
├── auth-service.json      # 认证服务 API
└── agent-manager.json     # Agent 管理 API
```

## API 定义文件格式

每个 API 定义文件遵循 Tyk 的 JSON 格式规范。

### 基础结构

```json
{
  "name": "API 名称",
  "api_id": "唯一标识符",
  "org_id": "组织 ID",
  "enable_jwt": true,
  "proxy": {
    "listen_path": "/gateway/路径/",
    "target_url": "http://backend-service/路径/"
  }
}
```

## 现有 API 说明

### 1. Auth Service API (auth-service.json)

**用途**: 认证和用户管理服务

**配置**:

- **监听路径**: `/api/v1/auth/`
- **目标地址**: `http://auth-service:8090/api/v1/auth/`
- **认证方式**: JWT
- **白名单路径**:
  - `POST /api/v1/auth/login` - 登录(无需认证)
  - `POST /api/v1/auth/register` - 注册(无需认证)

**端点示例**:

```bash
# 登录 (无需 Token)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 获取用户信息 (需要 Token)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/auth/me
```

### 2. Agent Manager API (agent-manager.json)

**用途**: Kubernetes Agent 管理

**配置**:

- **监听路径**: `/api/v1/agents`
- **目标地址**: `http://agent-manager:8081/api/v1/agents`
- **认证方式**: JWT
- **缓存**: 启用(60秒)
- **缓存端点**:
  - `GET /api/v1/agents` - Agent 列表
  - `GET /api/v1/clusters` - 集群列表

**端点示例**:

```bash
# 获取 Agent 列表
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/agents

# 获取特定 Agent
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/agents/{id}
```

## 创建新 API 定义

### 步骤 1: 创建 JSON 文件

在 `apps/` 目录创建新文件,例如 `my-service.json`:

```bash
cd apps/
touch my-service.json
```

### 步骤 2: 编写 API 定义

```json
{
  "name": "My Service API",
  "api_id": "my-service",
  "org_id": "default",
  "enable_jwt": true,
  "jwt_signing_method": "hmac",
  "jwt_identity_base_field": "sub",
  "proxy": {
    "listen_path": "/api/v1/myservice/",
    "target_url": "http://my-service:8080/api/v1/myservice/",
    "disable_strip_slash": true,
    "strip_listen_path": false
  },
  "version_data": {
    "not_versioned": true,
    "versions": {
      "Default": {
        "name": "Default",
        "use_extended_paths": true,
        "extended_paths": {}
      }
    }
  },
  "CORS": {
    "enable": true,
    "allowed_origins": ["*"],
    "allowed_methods": ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
    "allowed_headers": ["Origin", "Accept", "Content-Type", "Authorization"]
  },
  "global_rate_limit": {
    "rate": 1000,
    "per": 60
  },
  "active": true
}
```

### 步骤 3: 验证配置

```bash
# 验证 JSON 格式
cat my-service.json | jq .

# 或使用 Makefile
cd ..
make validate
```

### 步骤 4: 重载 API

```bash
# 方式 1: 使用 Makefile
make reload

# 方式 2: 手动重载
curl -H "x-tyk-authorization: 352d20ee67be67f6340b4c0605b044b7" \
  http://localhost:8080/tyk/reload/group

# 方式 3: 重启 Gateway
docker-compose restart tyk-gateway
```

### 步骤 5: 测试新 API

```bash
# 测试端点
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/myservice/
```

## 高级配置

### 1. 白名单路径(无需认证)

某些端点不需要 JWT 认证:

```json
{
  "version_data": {
    "versions": {
      "Default": {
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
    }
  }
}
```

### 2. 缓存配置

启用 HTTP 缓存以提高性能:

```json
{
  "cache_options": {
    "enable_cache": true,
    "cache_timeout": 60,
    "cache_all_safe_requests": false,
    "cache_response_codes": [200]
  },
  "version_data": {
    "versions": {
      "Default": {
        "extended_paths": {
          "cache": [
            {
              "path": "/api/v1/data",
              "method": "GET",
              "cache_response_codes": [200]
            }
          ]
        }
      }
    }
  }
}
```

### 3. 请求转换

修改请求头或路径:

```json
{
  "version_data": {
    "versions": {
      "Default": {
        "extended_paths": {
          "transform_headers": [
            {
              "path": "/api/v1/",
              "method": "GET",
              "add_headers": {
                "X-Custom-Header": "value"
              }
            }
          ],
          "url_rewrites": [
            {
              "path": "/api/v1/old-path",
              "method": "GET",
              "match_pattern": "/api/v1/old-path",
              "rewrite_to": "/api/v1/new-path"
            }
          ]
        }
      }
    }
  }
}
```

### 4. 熔断器

防止级联故障:

```json
{
  "version_data": {
    "versions": {
      "Default": {
        "extended_paths": {
          "circuit_breakers": [
            {
              "path": "/api/v1/",
              "method": "GET",
              "threshold_percent": 0.5,
              "samples": 5,
              "return_to_service_after": 60
            }
          ]
        }
      }
    }
  }
}
```

### 5. 请求大小限制

限制请求体大小:

```json
{
  "version_data": {
    "versions": {
      "Default": {
        "global_size_limit": 5242880,
        "extended_paths": {
          "size_limits": [
            {
              "path": "/api/v1/upload",
              "method": "POST",
              "size_limit": 10485760
            }
          ]
        }
      }
    }
  }
}
```

### 6. IP 白名单/黑名单

```json
{
  "enable_ip_whitelisting": true,
  "allowed_ips": [
    "192.168.1.0/24",
    "10.0.0.0/8"
  ],
  "enable_ip_blacklisting": false,
  "blacklisted_ips": []
}
```

### 7. 自定义响应

模拟端点或返回自定义响应:

```json
{
  "version_data": {
    "versions": {
      "Default": {
        "extended_paths": {
          "virtual": [
            {
              "path": "/api/v1/mock",
              "method": "GET",
              "response_function_name": "mockResponse",
              "function_source_type": "blob",
              "function_source_uri": "function mockResponse(request, session, config) { return TykJsResponse({ Body: JSON.stringify({message: 'Mock response'}), Code: 200 }, session.meta_data) }"
            }
          ]
        }
      }
    }
  }
}
```

## 常用命令

```bash
# 验证所有 API 定义
for file in *.json; do
  echo "Validating $file..."
  cat $file | jq . > /dev/null && echo "✓" || echo "✗"
done

# 列出所有 API
curl -H "x-tyk-authorization: 352d20ee67be67f6340b4c0605b044b7" \
  http://localhost:8080/tyk/apis | jq '.[] | {name, api_id, listen_path}'

# 获取特定 API 详情
curl -H "x-tyk-authorization: 352d20ee67be67f6340b4c0605b044b7" \
  http://localhost:8080/tyk/apis/auth-service | jq .

# 删除 API (需要使用 Gateway API)
curl -X DELETE \
  -H "x-tyk-authorization: 352d20ee67be67f6340b4c0605b044b7" \
  http://localhost:8080/tyk/apis/my-service
```

## 调试

启用详细日志:

```json
{
  "do_not_track": false,
  "enable_detailed_recording": true
}
```

查看 API 请求日志:

```bash
# 查看 Gateway 日志
docker-compose logs -f tyk-gateway | grep "my-service"

# 查看 Pump 日志(分析数据)
docker-compose logs -f tyk-pump
```

## 参考资源

- [Tyk API Definition](https://tyk.io/docs/tyk-apis/tyk-gateway-api/api-definition-objects/)
- [Extended Paths](https://tyk.io/docs/advanced-configuration/transform-traffic/)
- [Plugins & Middleware](https://tyk.io/docs/plugins/)
- [Authentication Methods](https://tyk.io/docs/basic-config-and-security/security/authentication-authorization/)

## 最佳实践

1. **使用有意义的 api_id**: 使用服务名作为 api_id
2. **设置合理的限流**: 根据服务容量设置限流规则
3. **启用缓存**: 对不频繁变化的数据启用缓存
4. **版本管理**: 使用 URL 版本(`/api/v1/`)而非 Header 版本
5. **CORS 配置**: 生产环境不要使用 `"*"`,指定具体域名
6. **监控**: 启用 `enable_analytics` 收集指标
7. **文档**: 每个 API 添加标签(`tags`)方便管理
