# Tyk Gateway 快速开始

5 分钟快速部署和测试 Tyk API Gateway。

## 一键启动

```bash
cd gateway-service/deployments/tyk
./start.sh
```

## 手动启动步骤

### 1. 准备配置

```bash
# 复制环境变量文件
cp .env.example .env

# (可选) 修改密钥
vi .env
```

### 2. 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps
```

### 3. 验证安装

```bash
# 测试 Tyk Gateway
curl http://localhost:8080/hello

# 测试 Tyk Dashboard
curl http://localhost:3000/hello
```

## 测试 API 路由

### 1. 测试公开端点(无需认证)

```bash
# 登录接口
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

**预期输出**:

```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "username": "admin"
    }
  }
}
```

### 2. 测试受保护端点(需要认证)

```bash
# 保存 Token
TOKEN="your-jwt-token-from-login"

# 获取 Agent 列表
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/agents

# 获取集群列表
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/clusters
```

### 3. 测试限流

```bash
# 快速发送多次请求测试限流
for i in {1..100}; do
  curl -s -o /dev/null -w "%{http_code}\n" \
    -H "Authorization: Bearer $TOKEN" \
    http://localhost:8080/api/v1/agents
done
```

**预期**: 前 1000 次请求返回 `200`,超过后返回 `429 Too Many Requests`。

## 访问 Dashboard

1. 浏览器打开 `http://localhost:3000`
2. 首次访问需要创建管理员账号
3. 查看 API 列表、流量统计、密钥管理

## 查看监控指标

```bash
# Prometheus 指标
curl http://localhost:9090/metrics

# 过滤 Tyk 相关指标
curl http://localhost:9090/metrics | grep tyk_
```

## 常用命令

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看 Gateway 日志
docker-compose logs -f tyk-gateway

# 重启 Gateway
docker-compose restart tyk-gateway

# 停止所有服务
docker-compose down

# 停止并清理数据
docker-compose down -v
```

## 故障排查

### Gateway 启动失败

```bash
# 查看日志
docker-compose logs tyk-gateway

# 检查配置文件
cat tyk.conf | jq .

# 测试 Redis 连接
docker-compose exec redis redis-cli ping
```

### API 返回 404

```bash
# 检查 API 定义
ls -la apps/

# 重载 API
curl -H "x-tyk-authorization: 352d20ee67be67f6340b4c0605b044b7" \
  http://localhost:8080/tyk/reload/group
```

### JWT 认证失败

```bash
# 检查 Token 格式
echo $TOKEN | cut -d'.' -f2 | base64 -d | jq .

# 检查 Authorization 头
curl -v -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/agents
```

## 下一步

- 阅读完整文档: [README.md](./README.md)
- 配置自定义 API: [apps/README.md](./apps/README.md)
- 生产部署: [PRODUCTION.md](./PRODUCTION.md)

## 帮助

遇到问题? 查看:

- [故障排查指南](./README.md#故障排查)
- [常见问题](./FAQ.md)
- [Tyk 官方文档](https://tyk.io/docs/)
