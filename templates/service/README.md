# 服务代码模板

本目录包含创建新服务所需的代码模板。

## 使用方法

### 方式 1: 手动复制模板

```bash
# 1. 创建目录结构
mkdir -p cmd/<service>/app/options
mkdir -p internal/<service>/{api,initializers,service,storage}

# 2. 复制模板文件
cp templates/service/main.go.tmpl cmd/<service>/main.go
cp templates/service/options.go.tmpl cmd/<service>/app/options/options.go
cp templates/service/app.go.tmpl cmd/<service>/app/app.go
cp templates/service/config.go.tmpl internal/<service>/config.go
cp templates/service/database.go.tmpl internal/<service>/initializers/database.go
cp templates/service/redis.go.tmpl internal/<service>/initializers/redis.go
cp templates/service/servers.go.tmpl internal/<service>/initializers/servers.go

# 3. 替换占位符
find cmd/<service> internal/<service> -type f -name "*.go" -exec sed -i '' \
  -e 's/{{SERVICE_NAME}}/<service>/g' \
  -e 's/{{SERVICE_NAME_UPPER}}/<SERVICE>/g' \
  -e 's/{{SERVICE_NAME_TITLE}}/<Service>/g' \
  -e 's/{{HEALTH_PORT}}/8090/g' \
  {} +
```

### 方式 2: 使用生成脚本（推荐）

```bash
# 生成完整的服务骨架
./scripts/generate-service.sh <service-name> [health-port]

# 示例
./scripts/generate-service.sh myservice 8094
```

## 模板说明

- `main.go.tmpl` - 服务入口点
- `options.go.tmpl` - Options 配置（启动层）
- `app.go.tmpl` - Application 实现（生命周期管理）
- `config.go.tmpl` - 业务配置（业务层）
- `database.go.tmpl` - 数据库初始化器
- `redis.go.tmpl` - Redis 初始化器
- `servers.go.tmpl` - HTTP 服务器初始化器

## 占位符

模板中使用以下占位符，需要替换为实际值：

- `{{SERVICE_NAME}}` - 服务名称（小写，如 `myservice`）
- `{{SERVICE_NAME_UPPER}}` - 服务名称（大写，如 `MYSERVICE`）
- `{{SERVICE_NAME_TITLE}}` - 服务名称（标题格式，如 `MyService`）
- `{{HEALTH_PORT}}` - 健康检查端口（如 `8090`）

## 端口分配

| 服务 | HTTP 端口 | 健康检查端口 |
|------|-----------|--------------|
| auth | 8080 | 8090 |
| agent-manager | 8080 | 8091 |
| orchestrator | 8081 | 8092 |
| reasoning | 8082 | 8093 |
| <新服务> | ? | 8094+ |

## 参考文档

- [SERVICE_STANDARD_PATTERN.md](../../docs/SERVICE_STANDARD_PATTERN.md) - 服务标准实现模式
- [SERVICE_UNIFICATION_PLAN.md](../../docs/SERVICE_UNIFICATION_PLAN.md) - 服务统一化计划
