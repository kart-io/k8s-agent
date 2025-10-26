# 端口使用说明

## 📌 两种端口说明

agent-manager 服务使用 **两个不同的端口**：

### 1. **业务服务端口** (`Server.Port`)

**用途**: 提供业务 API 服务
- **字段**: `opts.Server.Port`
- **默认值**: `8080`
- **配置方式**:
  - 配置文件: `server.port: 8080`
  - 环境变量: `AGENT_MANAGER_SERVER_PORT=8080`
  - 命令行: `--server.port=8080`

**提供的服务**:
```bash
# 业务 API 端点
http://localhost:8080/api/v1/agents         # Agent 管理 API
http://localhost:8080/api/v1/commands       # 命令管理 API
http://localhost:8080/api/v1/events         # 事件管理 API
http://localhost:8080/metrics               # Prometheus 指标
```

### 2. **健康检查端口** (`Health.Port`)

**用途**: 提供独立的健康检查服务
- **字段**: `opts.Health.Port`
- **默认值**: `8091`
- **配置方式**:
  - 配置文件: `health.port: 8091`
  - 环境变量: `AGENT_MANAGER_HEALTH_PORT=8091`
  - 命令行: `--health.port=8091`

**提供的服务**:
```bash
# 健康检查端点（独立端口）
http://localhost:8091/health                # 健康检查
http://localhost:8091/ready                 # 就绪检查
http://localhost:8091/live                  # 存活检查
```

---

## 🔍 为什么使用两个端口？

### 原因 1: 安全隔离

```yaml
# Kubernetes 部署示例
spec:
  containers:
  - name: agent-manager
    ports:
    - containerPort: 8080  # 业务端口（可能需要认证）
      name: http
    - containerPort: 8091  # 健康检查端口（无需认证）
      name: health

    livenessProbe:
      httpGet:
        path: /health
        port: 8091         # ✅ 使用独立端口，无需认证

    readinessProbe:
      httpGet:
        path: /ready
        port: 8091         # ✅ 使用独立端口，避免业务端口过载
```

### 原因 2: 故障隔离

- **业务端口崩溃** → 健康检查端口仍可用 → Kubernetes 可以检测到并重启
- **健康检查端口崩溃** → 业务端口仍可用 → 不影响正在处理的请求

### 原因 3: 负载隔离

- 健康检查请求（频繁）不会占用业务端口的资源
- 避免健康检查影响业务性能

---

## 📊 当前实现验证

### agent-manager 配置

```go
// cmd/agent-manager/app/options/options.go
func NewServerOptions() *ServerOptions {
    healthOpts := commonoptions.NewHealthOptions()
    healthOpts.Port = 8091  // ✅ 健康检查端口

    return &ServerOptions{
        Server: commonoptions.NewServerOptions(),  // Port = 8080（业务端口）
        Health: healthOpts,                         // Port = 8091（健康检查端口）
    }
}

// GetHealthPort 返回健康检查端口
func (o *ServerOptions) GetHealthPort() int {
    if o.Health != nil {
        return o.Health.Port  // ✅ 返回 8091
    }
    return 8091  // 默认端口
}
```

### app.go 使用

```go
// cmd/agent-manager/app/app.go

// 1. 业务 HTTP 服务器使用 Server.Port (8080)
a.httpInit = initializers.NewHTTPServerInitializer(
    a.opts,  // 内部使用 opts.Server.Port = 8080
    a.logger,
    // ...
)

// 2. 健康检查服务器使用 Health.Port (8091)
healthPort := a.opts.GetHealthPort()  // ✅ 返回 8091
healthAddr := fmt.Sprintf(":%d", healthPort)  // ":8091"
a.healthInit = pkginitializers.NewHealthCheckInitializer(healthAddr, a.logger)
```

---

## ✅ 端口分配表

| 服务 | 业务端口 (Server.Port) | 健康检查端口 (Health.Port) |
|------|------------------------|----------------------------|
| auth | 8080 | 8090 |
| agent-manager | 8080 | 8091 |
| orchestrator | 8081 | 8092 |
| reasoning | 8082 | 8093 |

---

## 🧪 测试验证

```bash
# 1. 启动 agent-manager
make run-agent-manager

# 2. 测试业务端口（8080）
curl http://localhost:8080/api/v1/agents
# 返回: {"code":0,"data":[],"message":"success"}

# 3. 测试健康检查端口（8091）
curl http://localhost:8091/health
# 返回: {"status":"ok"}

# 4. 查看日志确认
# 你应该看到:
# - HTTP Server listening on :8080  （业务端口）
# - Health Check Server listening on :8091  （健康检查端口）
```

---

## 📝 配置示例

### configs/config.yaml

```yaml
server:
  host: "0.0.0.0"
  port: 8080          # ← 业务服务端口
  mode: "release"

health:
  port: 8091          # ← 健康检查端口
  enabled: true
```

### 环境变量

```bash
# 业务端口
export AGENT_MANAGER_SERVER_PORT=8080

# 健康检查端口
export AGENT_MANAGER_HEALTH_PORT=8091
```

### 命令行参数

```bash
./agent-manager \
  --server.port=8080 \      # 业务端口
  --health.port=8091        # 健康检查端口
```

---

## 🔧 常见问题

### Q1: 为什么不使用同一个端口？

**A**:
- ✅ **安全**: 健康检查端口不需要认证，业务端口需要
- ✅ **隔离**: 避免健康检查请求影响业务性能
- ✅ **可靠**: 业务端口出问题时，健康检查仍可用

### Q2: GetHealthPort() 返回的是哪个端口？

**A**: 返回 **健康检查端口** (`Health.Port = 8091`)
- ❌ 不是业务端口 (`Server.Port = 8080`)

### Q3: 如何修改端口？

**A**: 三种方式任选其一：
```bash
# 方式1: 配置文件
vim configs/config.yaml
# server.port: 9090
# health.port: 9091

# 方式2: 环境变量
export AGENT_MANAGER_SERVER_PORT=9090
export AGENT_MANAGER_HEALTH_PORT=9091

# 方式3: 命令行
./agent-manager --server.port=9090 --health.port=9091
```

---

## ✨ 总结

**GetHealthPort()** 的作用:
```
GetHealthPort()
  ↓
返回 Health.Port (8091)
  ↓
用于创建独立的健康检查服务器
  ↓
监听在 :8091 端口
  ↓
提供 /health, /ready, /live 端点
```

**不是用于业务服务器！业务服务器使用 Server.Port (8080)**

---

**创建日期**: 2025-01-25
**文档版本**: v1.0
