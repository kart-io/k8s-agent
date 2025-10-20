# Auth Service 配置文件说明

## 配置文件列表

auth-service 提供了多个环境的配置文件，参考了 cluster-service 的配置格式：

### 1. config.yaml (默认配置)
**用途**: 默认生产环境配置模板
**数据库**: MySQL (localhost:3306)
**特点**:
- mode: `release` (生产模式)
- 日志格式: `json` (便于日志收集)
- 日志级别: `info`
- 超时: 10s

**使用方式**:
```bash
go run cmd/server/main.go
# 或
go run cmd/server/main.go -c configs/config.yaml
```

### 2. config-dev.yaml (远程开发环境)
**用途**: 连接到远程 MySQL 数据库的开发配置
**数据库**: MySQL (dbconn.sealoshzh.site:31675)
**特点**:
- mode: `debug` (调试模式)
- 日志格式: `text` (更易读)
- 日志级别: `debug`
- 超时: 30s (开发环境更长)
- 包含真实的远程数据库和 Redis 配置

**使用方式**:
```bash
go run cmd/server/main.go -c configs/config-dev.yaml
```

### 3. config-local.yaml (本地开发环境) 🆕
**用途**: 本地 Docker 或本地安装的 MySQL/Redis
**数据库**: MySQL (localhost:3306)
**特点**:
- mode: `debug`
- 日志格式: `text`
- 日志级别: `debug`
- 连接数较少 (适合本地资源)
- 用户名/密码: root/root (本地默认)

**使用方式**:
```bash
go run cmd/server/main.go -c configs/config-local.yaml
```

**适用场景**:
- 使用 Docker Compose 启动本地 MySQL
- 本地安装了 MySQL 和 Redis
- 离线开发

### 4. config-prod.yaml (生产环境) 🆕
**用途**: 生产环境配置 (支持环境变量)
**特点**:
- 所有敏感信息通过环境变量配置
- mode: `release`
- 日志格式: `json`
- 支持环境变量默认值

**环境变量**:
```bash
# 必需的环境变量
export DB_HOST=prod-mysql.example.com
export DB_PORT=3306
export DB_USER=auth_user
export DB_PASSWORD=your-secure-password
export DB_NAME=k8s_agent_auth

export REDIS_HOST=prod-redis.example.com
export REDIS_PORT=6379
export REDIS_PASSWORD=your-redis-password

export JWT_SECRET=your-super-secret-jwt-key-min-32-chars

# 可选的环境变量
export LOG_LEVEL=info
export JWT_EXPIRES_HOURS=24
export EMAIL_ENABLED=true
export SMTP_HOST=smtp.gmail.com
export SMTP_PORT=587
export SMTP_USER=your-email@gmail.com
export SMTP_PASSWORD=your-email-password
```

**使用方式**:
```bash
go run cmd/server/main.go -c configs/config-prod.yaml
```

## 配置文件对比

| 配置项 | config.yaml | config-dev.yaml | config-local.yaml | config-prod.yaml |
|--------|-------------|-----------------|-------------------|------------------|
| **用途** | 生产模板 | 远程开发 | 本地开发 | 生产环境 |
| **Server Mode** | release | debug | debug | release |
| **数据库** | localhost:3306 | dbconn.sealoshzh.site:31675 | localhost:3306 | 环境变量 |
| **Redis** | localhost:6379 | dbconn.sealoshzh.site:40210 | localhost:6379 | 环境变量 |
| **日志格式** | json | text | text | json |
| **日志级别** | info | debug | debug | info |
| **超时设置** | 10s | 30s | 30s | 10s |
| **连接池** | 100/10 | 100/10 | 25/5 | 100/10 |

## 参考 cluster-service 的改进

参考了 `k8s-agent/cluster-service/configs/` 的配置格式，进行了以下改进：

### 1. 统一的配置结构
```yaml
server:
  host: "0.0.0.0"
  port: 8090
  mode: "debug"
  read_timeout: 30s    # 新增: 参考 cluster-service
  write_timeout: 30s   # 新增: 参考 cluster-service
```

### 2. 数据库配置规范化
```yaml
database:
  host: "localhost"
  port: 3306
  user: "root"
  password: "root"
  dbname: "k8s_agent_auth"
  charset: "utf8mb4"        # MySQL 特有
  parse_time: true          # MySQL 特有
  max_idle_conns: 10        # 参考 cluster-service
  max_open_conns: 100       # 参考 cluster-service
```

### 3. 多环境配置文件
- ✅ config.yaml (默认/生产模板)
- ✅ config-dev.yaml (开发环境)
- ✅ config-local.yaml (本地开发) - 新增
- ✅ config-prod.yaml (生产环境) - 新增

## 快速开始

### 本地开发 (Docker)

1. 启动 MySQL 和 Redis:
```bash
docker-compose up -d mysql redis
```

2. 使用本地配置启动服务:
```bash
cd auth-service
go run cmd/server/main.go -c configs/config-local.yaml
```

### 远程开发环境

```bash
cd auth-service
go run cmd/server/main.go -c configs/config-dev.yaml
```

### 生产环境部署

1. 设置环境变量:
```bash
export DB_HOST=your-mysql-host
export DB_PASSWORD=your-secure-password
export JWT_SECRET=your-super-secret-key
# ... 其他环境变量
```

2. 启动服务:
```bash
go run cmd/server/main.go -c configs/config-prod.yaml
```

## 配置项说明

### Server 配置
- `host`: 监听地址 (0.0.0.0 表示所有接口)
- `port`: 服务端口 (默认: 8090)
- `mode`: 运行模式 (debug/release)
- `read_timeout`: 读取超时时间
- `write_timeout`: 写入超时时间

### Database 配置
- `host`: MySQL 主机地址
- `port`: MySQL 端口 (默认: 3306)
- `user`: 数据库用户名
- `password`: 数据库密码
- `dbname`: 数据库名称
- `charset`: 字符集 (建议: utf8mb4)
- `parse_time`: 是否解析时间类型
- `max_idle_conns`: 最大空闲连接数
- `max_open_conns`: 最大打开连接数

### Redis 配置
- `host`: Redis 主机地址
- `port`: Redis 端口 (默认: 6379)
- `password`: Redis 密码 (可为空)
- `db`: Redis 数据库编号 (0-15)
- `pool_size`: 连接池大小

### JWT 配置
- `secret`: JWT 签名密钥 (生产环境必须修改)
- `expires_hours`: Token 过期时间 (小时)

### Logging 配置
- `engine`: 日志引擎 (zap/slog)
- `level`: 日志级别 (debug/info/warn/error)
- `format`: 日志格式 (json/text)
- `output`: 输出目标 (stdout/stderr/文件路径)

### Email 配置 (可选)
- `enabled`: 是否启用邮件通知
- `smtp_host`: SMTP 服务器地址
- `smtp_port`: SMTP 端口
- `smtp_user`: SMTP 用户名
- `smtp_password`: SMTP 密码
- `from_address`: 发件人邮箱
- `from_name`: 发件人名称
- `template_dir`: 邮件模板目录

## 安全建议

### 开发环境
- ✅ 可以使用简单密码
- ✅ 可以硬编码配置
- ⚠️ 不要提交真实的生产密码到 Git

### 生产环境
- ❌ 不要在配置文件中硬编码密码
- ✅ 使用环境变量管理敏感信息
- ✅ JWT Secret 至少 32 字符
- ✅ 使用强密码
- ✅ 定期轮换密钥
- ✅ 启用 TLS/SSL

## 故障排查

### 无法连接数据库
```bash
# 检查 MySQL 是否运行
mysql -h localhost -P 3306 -u root -p

# 检查端口是否开放
telnet localhost 3306
```

### 无法连接 Redis
```bash
# 检查 Redis 是否运行
redis-cli -h localhost -p 6379 ping

# 使用密码连接
redis-cli -h localhost -p 6379 -a your-password ping
```

### JWT Token 无效
- 检查 JWT Secret 是否一致
- 检查 Token 是否过期
- 查看服务日志

## 参考资料

- [cluster-service 配置](../cluster-service/configs/)
- [Auth Service README](../README.md)
- [集成文档](../../k8s-agent-web/AUTH_SERVICE_INTEGRATION.md)
