# init-mysql.sh 脚本修复报告

**日期**: 2025-10-21
**问题**: `make init-mysql` 命令无法运行
**状态**: ✅ 已修复

---

## 问题分析

### 原始错误

```bash
$ make init-mysql
ERROR 2002 (HY000): Can't connect to local MySQL server through socket '/tmp/mysql.sock' (2)
✗ Failed to initialize database
```

### 根本原因

`init-mysql.sh` 脚本尝试使用 **Unix socket** (`/tmp/mysql.sock`) 连接本地 MySQL，但实际的 MySQL 运行在 **Docker 容器** 中，无法通过 socket 连接。

---

## 解决方案

### 修改脚本支持 Docker 容器

更新了 `scripts/init-mysql.sh`，添加了 Docker 检测逻辑：

**修改前**:
```bash
# 直接使用 mysql 命令连接
MYSQL_CMD="mysql -h $DB_HOST -P $DB_PORT -u $DB_USER"
if [ -n "$DB_PASSWORD" ]; then
    MYSQL_CMD="$MYSQL_CMD -p$DB_PASSWORD"
fi
```

**修改后**:
```bash
# 检测 MySQL 是否在 Docker 容器中运行
if docker ps | grep -q cluster-mysql; then
    echo "✓ Detected MySQL running in Docker container 'cluster-mysql'"
    echo "Using Docker exec to connect to MySQL..."

    # 使用 Docker exec 连接
    if [ -n "$DB_PASSWORD" ]; then
        MYSQL_CMD="docker exec -i cluster-mysql mysql -u $DB_USER -p$DB_PASSWORD"
    else
        MYSQL_CMD="docker exec -i cluster-mysql mysql -u $DB_USER"
    fi
else
    # 本地 MySQL，使用 TCP 协议
    MYSQL_CMD="mysql -h $DB_HOST -P $DB_PORT -u $DB_USER --protocol=TCP"
    if [ -n "$DB_PASSWORD" ]; then
        MYSQL_CMD="$MYSQL_CMD -p$DB_PASSWORD"
    fi
fi
```

---

## 修复效果

### 测试命令

```bash
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/auth-service
make init-mysql
```

### 成功输出

```
Initializing MySQL database from config.yaml...
=== Auth Service MySQL Initialization ===
Warning: yq not found. Attempting to parse config.yaml with grep/awk...
Database Configuration:
  Host: localhost
  Port: 3306
  User: root
  Database: user_auth

✓ Detected MySQL running in Docker container 'cluster-mysql'
Using Docker exec to connect to MySQL...
Executing initialization script...
✓ Database initialized successfully!

Default Admin Account:
  Username: admin
  Password: admin123
  Email: admin@example.com

⚠️  Please change the default password after first login!
✅ MySQL database initialized successfully
```

---

## 数据库初始化内容

### 创建的表

| 表名 | 说明 |
|------|------|
| `users` | 用户信息表 |
| `roles` | 角色表 |
| `permissions` | 权限表 |
| `user_roles` | 用户角色关联表 |
| `role_permissions` | 角色权限关联表 |
| `api_keys` | API 密钥表 |

### 默认管理员账号

**重要**: 首次登录后请立即更改密码！

- **用户名**: `admin`
- **密码**: `admin123`
- **邮箱**: `admin@example.com`
- **角色**: 超级管理员 (拥有所有权限)

### 默认角色

| 角色 | 说明 |
|------|------|
| `super_admin` | 超级管理员 (所有权限) |
| `cluster_admin` | 集群管理员 (集群管理权限) |
| `developer` | 开发者 (只读权限) |

### 默认权限

| 权限 | 说明 |
|------|------|
| `cluster:read` | 读取集群信息 |
| `cluster:write` | 修改集群配置 |
| `cluster:delete` | 删除集群 |
| `user:read` | 读取用户信息 |
| `user:write` | 修改用户信息 |
| `user:delete` | 删除用户 |
| `role:read` | 读取角色信息 |
| `role:write` | 修改角色配置 |
| `role:delete` | 删除角色 |

---

## 使用方法

### 初始化数据库

```bash
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/auth-service

# 方法 1: 使用 Make
make init-mysql

# 方法 2: 直接运行脚本
bash scripts/init-mysql.sh
```

### 重置数据库

⚠️ **警告**: 这会删除所有数据！

```bash
make db-reset
```

这个命令会：
1. 删除 `user_auth` 数据库
2. 重新创建数据库
3. 运行初始化脚本

### 手动操作

```bash
# 创建数据库
make db-create

# 删除数据库 (需要确认)
make db-drop

# 完整重置 (删除 → 创建 → 初始化)
make db-reset
```

---

## 验证

### 检查数据库连接

```bash
docker exec -it cluster-mysql mysql -uroot -proot123 user_auth
```

### 查看表结构

```sql
-- 在 MySQL 中执行
USE user_auth;
SHOW TABLES;

-- 查看用户表
DESC users;

-- 查看管理员账号
SELECT id, username, email, is_active FROM users WHERE username = 'admin';
```

### 查看角色和权限

```sql
-- 查看所有角色
SELECT * FROM roles;

-- 查看所有权限
SELECT * FROM permissions;

-- 查看角色权限关联
SELECT r.name as role, p.name as permission
FROM role_permissions rp
JOIN roles r ON rp.role_id = r.id
JOIN permissions p ON rp.permission_id = p.id;
```

---

## 脚本特性

### 自动检测环境

脚本会自动检测 MySQL 运行环境：

1. **Docker 容器**: 使用 `docker exec` 连接
2. **本地 MySQL**: 使用 `mysql --protocol=TCP` 连接

### YAML 配置解析

脚本支持两种方式解析 YAML：

1. **使用 yq**: 如果安装了 `yq` (推荐)
2. **使用 grep/awk**: 简单解析 (备用方案)

### 配置文件

脚本默认读取 `configs/config-dev.yaml`：

```yaml
database:
  host: localhost
  port: 3306
  user: root
  password: root123  # ✅ 正确的密码
  dbname: user_auth
```

---

## 常见问题

### Q1: yq 未安装警告

**提示**:
```
Warning: yq not found. Attempting to parse config.yaml with grep/awk...
```

**A**: 这只是警告，不影响功能。脚本会使用备用的 grep/awk 方案解析配置。

**可选安装 yq**:
```bash
# macOS
brew install yq

# Linux
wget https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 -O /usr/local/bin/yq
chmod +x /usr/local/bin/yq
```

### Q2: 数据库已存在的表

**A**: 脚本使用 `CREATE TABLE IF NOT EXISTS`，如果表已存在会跳过创建，不会报错。

### Q3: 密码警告

**提示**:
```
mysql: [Warning] Using a password on the command line interface can be insecure.
```

**A**: 这是 MySQL 的安全警告，可以忽略。在生产环境建议使用 `.my.cnf` 配置文件。

### Q4: 想要重新初始化数据库

**A**: 使用 `make db-reset` 完全重置数据库：
```bash
make db-reset
# 会提示确认: Are you sure? (y/N):
# 输入 y 确认
```

---

## 相关文件

### 修改的文件

1. ✅ `scripts/init-mysql.sh` - 添加 Docker 检测逻辑

### 相关文件

2. `scripts/init-mysql.sql` - SQL 初始化脚本
3. `Makefile` - Make 目标定义
4. `configs/config-dev.yaml` - 数据库配置

---

## Makefile 相关命令

```bash
# 数据库初始化
make init-mysql     # 初始化 MySQL 数据库
make init-db        # 别名，等同于 init-mysql

# 数据库管理
make db-create      # 创建数据库
make db-drop        # 删除数据库 (需要确认)
make db-reset       # 重置数据库 (drop + create + init)

# 完整启动流程
make quick-start    # deps + db-create + init-mysql + run-local

# 查看配置
make show-config-dev   # 显示开发配置
make info             # 显示项目信息
```

---

## 下一步

### 1. 启动 auth-service

```bash
# 使用开发配置启动
make run-dev

# 或使用本地配置
make run-local
```

### 2. 测试 API

```bash
# 测试健康检查
curl http://localhost:8090/health

# 测试登录 (使用默认管理员账号)
curl -X POST http://localhost:8090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

### 3. 修改默认密码

首次登录后，立即修改管理员密码！

---

## 总结

✅ **问题已解决**: `make init-mysql` 现在可以正常工作
✅ **Docker 支持**: 自动检测并使用 Docker 容器中的 MySQL
✅ **数据库已初始化**: 所有表和默认数据已创建
✅ **管理员账号**: admin / admin123 (请立即修改密码)

---

**修复时间**: 2025-10-21
**修复者**: Claude Code (AI Assistant)
**状态**: ✅ 完成并验证
