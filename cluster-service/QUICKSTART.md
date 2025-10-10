# Cluster Service 快速启动指南

## 目录
- [前置要求](#前置要求)
- [配置文件](#配置文件)
- [启动方式](#启动方式)
- [数据库初始化](#数据库初始化)
- [API 测试](#api-测试)

## 前置要求

### 1. 系统依赖
- Go 1.21+
- PostgreSQL 13+
- (可选) Docker & Docker Compose

### 2. PostgreSQL 安装

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
```

**macOS:**
```bash
brew install postgresql@13
brew services start postgresql@13
```

**Docker:**
```bash
docker run -d \
  --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:13-alpine
```

## 配置文件

### 默认配置文件
项目提供了三个配置文件：

1. **configs/config.yaml** - 默认配置
2. **configs/config.dev.yaml** - 开发环境配置
3. **configs/config.prod.yaml** - 生产环境配置

### 配置说明

```yaml
server:
  port: 8082              # 服务端口
  mode: debug             # 运行模式: debug, release, test
  read_timeout: 10s       # 读超时
  write_timeout: 10s      # 写超时

database:
  host: localhost         # 数据库地址
  port: 5432             # 数据库端口
  user: postgres         # 用户名
  password: postgres     # 密码
  dbname: cluster_db     # 数据库名
  sslmode: disable       # SSL 模式
  max_open_conns: 25     # 最大连接数
  max_idle_conns: 5      # 最大空闲连接数

jwt:
  secret: your-secret-key # JWT 密钥

logging:
  level: info            # 日志级别: debug, info, warn, error
  format: json           # 日志格式: json, text
```

## 启动方式

### 方式 1: 快速启动脚本 (推荐开发环境)

```bash
# 自动检查依赖、创建数据库、初始化表结构并启动服务
./scripts/start-dev.sh
```

脚本会自动：
- ✓ 检查 PostgreSQL 是否运行
- ✓ 创建数据库 (cluster_dev)
- ✓ 初始化表结构
- ✓ 编译服务
- ✓ 启动服务

### 方式 2: 手动启动

#### 步骤 1: 创建数据库
```bash
createdb -U postgres cluster_db
```

#### 步骤 2: 初始化表结构
```bash
psql -U postgres -d cluster_db -f scripts/init-db.sql
```

#### 步骤 3: 编译服务
```bash
go build -o server ./cmd/server
```

#### 步骤 4: 启动服务
```bash
# 使用默认配置
./server

# 使用开发配置
./server -config configs/config.dev.yaml

# 使用生产配置
./server -config configs/config.prod.yaml
```

### 方式 3: Docker Compose (推荐测试环境)

```bash
cd scripts
docker-compose up -d
```

这会启动：
- PostgreSQL 容器
- Cluster Service 容器

查看日志：
```bash
docker-compose logs -f cluster-service
```

停止服务：
```bash
docker-compose down
```

### 方式 4: Make 命令

```bash
# 构建
make build

# 运行
make run

# 测试
make test

# 清理
make clean
```

## 数据库初始化

### 自动初始化
使用 `start-dev.sh` 脚本会自动初始化数据库。

### 手动初始化

```bash
# 连接到 PostgreSQL
psql -U postgres

# 创建数据库
CREATE DATABASE cluster_db;

# 运行初始化脚本
\c cluster_db
\i scripts/init-db.sql
```

### 验证表结构

```bash
psql -U postgres -d cluster_db -c "\dt"
```

应该看到 `clusters` 表。

## API 测试

### 1. 健康检查

```bash
curl http://localhost:8082/health
```

预期响应：
```json
{
  "status": "ok",
  "service": "cluster-service"
}
```

### 2. 添加集群

```bash
curl -X POST http://localhost:8082/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "id": "cluster-001",
    "name": "Development Cluster",
    "description": "Local development cluster",
    "endpoint": "https://localhost:6443",
    "region": "local",
    "provider": "minikube",
    "kubeconfig": "..."
  }'
```

### 3. 获取集群健康状态

```bash
curl http://localhost:8082/api/v1/clusters/cluster-001/health
```

### 4. 获取 Pod 列表

```bash
# 获取 default 命名空间的 Pods
curl http://localhost:8082/api/v1/clusters/cluster-001/namespaces/default/pods

# 获取指定命名空间的 Pods
curl http://localhost:8082/api/v1/clusters/cluster-001/namespaces/kube-system/pods
```

## 环境变量

可以使用环境变量覆盖配置：

```bash
# 复制示例文件
cp .env.example .env

# 编辑 .env
vim .env

# 使用环境变量启动
export $(cat .env | xargs)
./server
```

## 常见问题

### 1. 无法连接 PostgreSQL

**错误:** `connection refused`

**解决:**
```bash
# 检查 PostgreSQL 是否运行
sudo systemctl status postgresql

# 启动 PostgreSQL
sudo systemctl start postgresql
```

### 2. 数据库连接权限问题

**错误:** `password authentication failed`

**解决:**
编辑 PostgreSQL 配置允许本地连接：
```bash
sudo vim /etc/postgresql/13/main/pg_hba.conf
```

添加：
```
local   all   postgres   trust
host    all   all        127.0.0.1/32   trust
```

重启 PostgreSQL：
```bash
sudo systemctl restart postgresql
```

### 3. 端口已被占用

**错误:** `bind: address already in use`

**解决:**
修改 `configs/config.yaml` 中的 `server.port`：
```yaml
server:
  port: 8083  # 改为其他端口
```

### 4. 连接 Kubernetes 集群失败

确保：
- Kubeconfig 文件有效
- 集群网络可访问
- 认证凭据未过期

## 开发建议

### 1. 开发模式
使用 `config.dev.yaml` 启用详细日志：
```yaml
logging:
  level: debug
  format: text
```

### 2. 热重载
使用 `air` 工具实现热重载：
```bash
go install github.com/cosmtrek/air@latest
air
```

### 3. 日志查看
```bash
# 实时查看日志
tail -f logs/cluster-service.log

# 使用 jq 格式化 JSON 日志
tail -f logs/cluster-service.log | jq .
```

## 生产部署

### 1. 安全配置

- ✓ 使用强 JWT 密钥 (至少 32 字符)
- ✓ 启用数据库 SSL: `sslmode: require`
- ✓ 使用环境变量管理敏感信息
- ✓ 设置合理的连接池大小
- ✓ 配置防火墙规则

### 2. 性能优化

```yaml
database:
  max_open_conns: 50
  max_idle_conns: 10

server:
  mode: release
  read_timeout: 10s
  write_timeout: 10s
```

### 3. 监控

- 启用 Prometheus metrics
- 配置健康检查
- 设置日志收集 (ELK/Loki)

## 下一步

- 阅读 [README.md](README.md) 了解完整功能
- 查看 [API 文档](docs/api.md)
- 参考 [开发指南](docs/development.md)

## 支持

如有问题，请提交 Issue 或联系开发团队。
