# MySQL Connection Troubleshooting Guide

## 问题：MySQL 连接失败

**错误信息**：
```
[mysql] unexpected EOF
Lost connection to MySQL server at 'reading initial communication packet'
```

**原因**：远程 MySQL 服务器 `dbconn.sealosbja.site:33726` 无法访问。

---

## 🚀 快速解决方案

### 方案 1: 使用本地 Docker MySQL（推荐）

#### 自动设置（推荐）

```bash
# 运行自动设置脚本
./setup-mysql.sh

# 脚本会自动：
# 1. 检查 Docker 是否安装
# 2. 创建 MySQL 8.0 容器
# 3. 创建数据库和用户
# 4. 等待 MySQL 就绪
# 5. 显示连接信息
```

#### 手动设置

```bash
# 1. 启动 MySQL 容器
docker run -d \
  --name cluster-mysql \
  -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=root123 \
  -e MYSQL_DATABASE=cluster_db \
  -e MYSQL_USER=cluster_user \
  -e MYSQL_PASSWORD=cluster_pass \
  mysql:8.0

# 2. 等待 MySQL 启动（约 15-30 秒）
docker logs -f cluster-mysql  # Ctrl+C 退出
# 看到 "ready for connections" 就可以了

# 3. 运行服务
make run-local
```

#### 验证 MySQL 运行

```bash
# 检查容器状态
docker ps | grep cluster-mysql

# 查看日志
docker logs cluster-mysql

# 连接到 MySQL
docker exec -it cluster-mysql mysql -ucluster_user -pcluster_pass cluster_db

# 测试连接
mysql -h127.0.0.1 -P3306 -ucluster_user -pcluster_pass -e "SELECT VERSION();"
```

---

### 方案 2: 修复远程数据库连接

如果必须使用远程数据库 `dbconn.sealosbja.site:33726`：

#### 1. 检查网络连接

```bash
# 测试端口是否开放
telnet dbconn.sealosbja.site 33726

# 或使用 nc
nc -zv dbconn.sealosbja.site 33726

# 测试 DNS 解析
nslookup dbconn.sealosbja.site
ping dbconn.sealosbja.site
```

#### 2. 验证数据库类型和凭据

```bash
# 尝试使用 mysql 客户端连接
mysql -h dbconn.sealosbja.site -P 33726 -u postgres -p

# 如果提示输入密码，说明端口可达
# 如果报错 "Can't connect to MySQL server"，说明端口不可达或被防火墙阻止
```

#### 3. 检查配置文件

确保 `configs/config.dev.yaml` 中的数据库配置正确：

```yaml
database:
  host: dbconn.sealosbja.site
  port: 33726
  user: <正确的 MySQL 用户名>
  password: <正确的密码>
  dbname: <正确的数据库名>
  sslmode: disable  # MySQL 不使用此参数，但保留兼容性
```

#### 4. 联系数据库管理员

如果仍然无法连接：
- 确认数据库服务器是否在运行
- 检查防火墙/安全组规则
- 验证你的 IP 地址是否被允许访问
- 确认用户名和密码是否正确

---

### 方案 3: 使用环境变量覆盖配置

```bash
# 临时使用不同的数据库配置
DB_HOST=localhost \
DB_PORT=3306 \
DB_USER=cluster_user \
DB_PASSWORD=cluster_pass \
DB_NAME=cluster_db \
make run-dev
```

---

## 📝 配置文件说明

项目提供了多个配置文件：

| 配置文件 | 用途 | 数据库 |
|---------|------|--------|
| `config.yaml` | 生产/默认配置 | localhost:5432 (需修改) |
| `config.dev.yaml` | 开发环境配置 | dbconn.sealosbja.site:33726 |
| `config-local.yaml` | 本地 Docker 配置 | localhost:3306 ✅ |

### 使用本地配置

```bash
# 使用 config-local.yaml（推荐用于本地开发）
make run-local

# 或直接指定配置文件
go run cmd/server/main.go -config configs/config-local.yaml
```

---

## 🔧 常用命令

### Docker MySQL 管理

```bash
# 启动 MySQL 容器
docker start cluster-mysql

# 停止 MySQL 容器
docker stop cluster-mysql

# 重启 MySQL 容器
docker restart cluster-mysql

# 查看 MySQL 日志
docker logs cluster-mysql
docker logs -f cluster-mysql  # 实时查看

# 删除 MySQL 容器（会丢失所有数据）
docker rm -f cluster-mysql

# 进入 MySQL shell
docker exec -it cluster-mysql mysql -ucluster_user -pcluster_pass cluster_db
```

### 服务运行

```bash
# 使用本地 MySQL 运行
make run-local

# 使用开发环境配置运行
make run-dev

# 使用默认配置运行
make run
```

### 测试连接

```bash
# 启动服务后，测试版本端点
curl http://localhost:8082/version
curl http://localhost:8082/version/simple

# 测试健康检查
curl http://localhost:8082/health

# 测试 K8s API（需要配置 kubeconfig）
curl http://localhost:8082/api/k8s/clusters
```

---

## 🐛 常见问题

### Q1: Docker 命令找不到

**问题**：`docker: command not found`

**解决**：
```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y docker.io
sudo systemctl start docker
sudo systemctl enable docker

# 添加当前用户到 docker 组（避免每次使用 sudo）
sudo usermod -aG docker $USER
# 需要重新登录才能生效
```

### Q2: 端口 3306 已被占用

**问题**：`Error starting userland proxy: listen tcp4 0.0.0.0:3306: bind: address already in use`

**解决方案 1**：停止占用端口的 MySQL
```bash
# 查找占用 3306 端口的进程
sudo lsof -i :3306
sudo netstat -tulpn | grep 3306

# 停止 MySQL 服务
sudo systemctl stop mysql
```

**解决方案 2**：使用不同端口
```bash
# 启动 MySQL 到 3307 端口
docker run -d \
  --name cluster-mysql \
  -p 3307:3306 \  # 映射到 3307
  -e MYSQL_ROOT_PASSWORD=root123 \
  -e MYSQL_DATABASE=cluster_db \
  -e MYSQL_USER=cluster_user \
  -e MYSQL_PASSWORD=cluster_pass \
  mysql:8.0

# 修改 config-local.yaml 中的端口
database:
  port: 3307  # 改为 3307
```

### Q3: MySQL 容器启动后立即退出

**问题**：容器启动后立即停止

**诊断**：
```bash
# 查看容器日志
docker logs cluster-mysql

# 检查容器状态
docker ps -a | grep cluster-mysql
```

**常见原因**：
- 端口被占用
- Docker 磁盘空间不足
- 权限问题

**解决**：
```bash
# 清理停止的容器
docker system prune -f

# 检查磁盘空间
df -h

# 重新创建容器
docker rm -f cluster-mysql
./setup-mysql.sh
```

### Q4: "Access denied for user"

**问题**：`Access denied for user 'cluster_user'@'localhost'`

**原因**：用户名或密码错误

**解决**：
```bash
# 使用 root 用户检查
docker exec -it cluster-mysql mysql -uroot -proot123 -e "SELECT User, Host FROM mysql.user;"

# 重新创建用户（如果需要）
docker exec -it cluster-mysql mysql -uroot -proot123 <<EOF
DROP USER IF EXISTS 'cluster_user'@'%';
CREATE USER 'cluster_user'@'%' IDENTIFIED BY 'cluster_pass';
GRANT ALL PRIVILEGES ON cluster_db.* TO 'cluster_user'@'%';
FLUSH PRIVILEGES;
EOF
```

### Q5: 服务启动后无法访问 API

**问题**：服务启动成功，但 curl 请求失败

**诊断**：
```bash
# 检查服务是否在监听
netstat -tulpn | grep 8082
lsof -i :8082

# 检查服务日志
# 日志应该显示 "Cluster service started successfully"
```

**解决**：
```bash
# 确保防火墙允许访问
sudo ufw allow 8082/tcp  # Ubuntu/Debian

# 使用 localhost 而不是 127.0.0.1
curl http://localhost:8082/version

# 检查是否监听在正确的地址
# 服务应该监听 0.0.0.0:8082 而不是 127.0.0.1:8082
```

---

## 📋 完整设置流程

### 首次设置（推荐）

```bash
# 1. 运行自动设置脚本
./setup-mysql.sh

# 2. 启动服务
make run-local

# 3. 测试 API（在另一个终端）
curl http://localhost:8082/version
curl http://localhost:8082/version/simple

# 4. 运行测试脚本
export BASE_URL="http://localhost:8082"
./test-version-api.sh
```

### 日常开发流程

```bash
# 1. 启动 MySQL（如果已停止）
docker start cluster-mysql

# 2. 启动服务
make run-local

# 3. 开发完成后停止
# Ctrl+C 停止服务
docker stop cluster-mysql  # 可选：停止 MySQL
```

---

## 📚 相关文档

- [VERSION_INTEGRATION.md](./VERSION_INTEGRATION.md) - 版本管理集成指南
- [PROJECT_STATUS.md](./PROJECT_STATUS.md) - 项目状态报告
- [README.md](./README.md) - 项目主文档
- [QUICKSTART_TEST.md](./QUICKSTART_TEST.md) - 快速测试指南

---

**最后更新**: 2025-10-17
**作者**: Claude (AI Assistant)
