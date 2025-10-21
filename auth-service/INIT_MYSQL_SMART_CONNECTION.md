# init-mysql.sh 智能连接配置

**日期**: 2025-10-21
**版本**: v2.0 - 基于配置文件的智能连接
**状态**: ✅ 已完成

---

## 概述

`init-mysql.sh` 脚本现在具有智能连接逻辑，能够根据配置文件中的数据库主机地址自动选择最佳的连接方式。

---

## 连接逻辑

脚本会读取 `configs/config-dev.yaml` 中的数据库配置，并根据 `database.host` 的值自动选择连接方式：

### 决策流程

```
读取配置文件
   ↓
检查 database.host 值
   ↓
   ├─ localhost / 127.0.0.1
   │    ↓
   │    检测 Docker 容器
   │    ├─ 找到 cluster-mysql → 使用 Docker exec
   │    └─ 未找到容器 → 使用 TCP 协议连接本地 MySQL
   │
   └─ 其他 IP/域名
        ↓
        使用标准 MySQL 客户端 TCP 连接
```

---

## 场景示例

### 场景 1: 本地 Docker MySQL (当前使用)

**配置文件** (`config-dev.yaml`):
```yaml
database:
  host: "localhost"
  port: 3306
  user: "root"
  password: "root123"
  dbname: "user_auth"
```

**脚本行为**:
1. 检测到 `host: localhost`
2. 检测到 Docker 容器 `cluster-mysql` 正在运行
3. 使用 Docker exec 连接

**连接命令**:
```bash
docker exec -i cluster-mysql mysql -u root -proot123
```

**输出**:
```
✓ Detected MySQL in Docker container 'cluster-mysql'
Using Docker exec for connection...
Connection method: docker exec -i cluster-mysql mysql -u root
```

---

### 场景 2: 本地安装的 MySQL

**配置文件**:
```yaml
database:
  host: "localhost"
  port: 3306
  user: "root"
  password: "mypassword"
  dbname: "user_auth"
```

**脚本行为**:
1. 检测到 `host: localhost`
2. 未检测到 Docker 容器 `cluster-mysql`
3. 使用 TCP 协议连接本地 MySQL

**连接命令**:
```bash
mysql -h localhost -P 3306 -u root --protocol=TCP -pmypassword
```

**输出**:
```
Connecting to MySQL via TCP protocol...
Connection method: mysql -h localhost -P 3306 -u root --protocol=TCP
```

---

### 场景 3: 远程 MySQL 服务器

**配置文件**:
```yaml
database:
  host: "192.168.1.100"  # 或域名: "mysql.example.com"
  port: 3306
  user: "remote_user"
  password: "remote_pass"
  dbname: "user_auth"
```

**脚本行为**:
1. 检测到 `host` 不是 localhost
2. 直接使用标准 MySQL 客户端连接

**连接命令**:
```bash
mysql -h 192.168.1.100 -P 3306 -u remote_user -premote_pass
```

**输出**:
```
Connecting to remote MySQL at 192.168.1.100:3306...
Connection method: mysql -h 192.168.1.100 -P 3306 -u remote_user
```

---

### 场景 4: 使用 127.0.0.1

**配置文件**:
```yaml
database:
  host: "127.0.0.1"
  port: 3306
  user: "root"
  password: "root123"
  dbname: "user_auth"
```

**脚本行为**:
与 `localhost` 相同，检测 Docker 容器后选择连接方式。

---

## 配置文件支持

### 支持的配置文件

脚本默认读取 `configs/config-dev.yaml`，但可以通过修改脚本中的 `CONFIG_FILE` 变量来指定其他配置：

```bash
# 脚本中的配置 (第17行)
CONFIG_FILE="$PROJECT_ROOT/configs/config-dev.yaml"
```

### 使用不同的配置文件

如果需要使用其他配置文件，可以修改脚本或创建符号链接：

```bash
# 方法 1: 修改脚本
# 编辑 scripts/init-mysql.sh，将第17行改为:
CONFIG_FILE="$PROJECT_ROOT/configs/config-prod.yaml"

# 方法 2: 创建符号链接 (推荐)
cd configs
ln -sf config-local.yaml config-dev.yaml
```

---

## YAML 解析

脚本支持两种 YAML 解析方式：

### 1. 使用 yq (推荐)

如果系统安装了 `yq`，脚本会使用它进行精确的 YAML 解析：

```bash
DB_HOST=$(yq eval '.database.host' "$CONFIG_FILE")
DB_PORT=$(yq eval '.database.port' "$CONFIG_FILE")
DB_USER=$(yq eval '.database.user' "$CONFIG_FILE")
DB_PASSWORD=$(yq eval '.database.password' "$CONFIG_FILE")
DB_NAME=$(yq eval '.database.dbname' "$CONFIG_FILE")
```

**安装 yq**:
```bash
# macOS
brew install yq

# Linux
wget https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 -O /usr/local/bin/yq
chmod +x /usr/local/bin/yq

# 验证安装
yq --version
```

### 2. 使用 grep/awk (备用方案)

如果 `yq` 未安装，脚本会自动使用 `grep` 和 `awk` 进行简单解析：

```bash
DB_HOST=$(grep -A 10 "^database:" "$CONFIG_FILE" | grep "host:" | awk '{print $2}' | tr -d '"')
DB_PORT=$(grep -A 10 "^database:" "$CONFIG_FILE" | grep "port:" | awk '{print $2}')
# ... 其他字段类似
```

**注意**: 这种方式对简单的 YAML 文件有效，但对复杂的 YAML 结构可能不够准确。

---

## 安全特性

### 1. 密码保护

脚本在显示连接信息时会隐藏密码：

```bash
# 显示连接方法时去除密码部分
echo "Connection method: ${MYSQL_CMD% -p*}"

# 输出示例:
# Connection method: mysql -h localhost -P 3306 -u root --protocol=TCP
# (密码 -proot123 被隐藏)
```

### 2. 错误处理

```bash
# 设置严格模式
set -e  # 遇到错误立即退出

# Docker 检测时忽略错误
if docker ps 2>/dev/null | grep -q cluster-mysql; then
    # ... 使用 Docker
fi
```

---

## 环境要求

### 必需的工具

| 工具 | 必需性 | 用途 |
|------|--------|------|
| `bash` | ✅ 必需 | 运行脚本 |
| `mysql` | 条件必需* | 连接非 Docker MySQL |
| `docker` | 条件必需* | 连接 Docker MySQL |
| `yq` | 🟡 可选 | YAML 解析 (推荐) |
| `grep`/`awk` | ✅ 必需 | 备用 YAML 解析 |

*根据配置文件中的 `database.host` 决定

### 环境检查

脚本会自动检测所需工具：

```bash
# 检查 yq
if ! command -v yq &> /dev/null; then
    echo "Warning: yq not found. Using grep/awk fallback..."
fi

# 检查配置文件
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Error: Config file not found at $CONFIG_FILE"
    exit 1
fi
```

---

## 故障排查

### 问题 1: 找不到配置文件

**错误**:
```
Error: Config file not found at /path/to/config-dev.yaml
```

**解决**:
```bash
# 检查文件是否存在
ls -la configs/

# 确保在正确的目录
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/auth-service

# 运行脚本
bash scripts/init-mysql.sh
```

### 问题 2: 连接失败 (Docker 方式)

**错误**:
```
ERROR 2002 (HY000): Can't connect to MySQL server
```

**解决**:
```bash
# 检查 Docker 容器是否运行
docker ps | grep cluster-mysql

# 如果没有运行，启动它
docker start cluster-mysql

# 验证可以连接
docker exec cluster-mysql mysql -uroot -proot123 -e "SELECT 1;"
```

### 问题 3: 连接失败 (TCP 方式)

**错误**:
```
ERROR 2003 (HY000): Can't connect to MySQL server on 'localhost'
```

**解决**:
```bash
# 检查 MySQL 服务是否运行
sudo systemctl status mysql

# 检查端口是否监听
netstat -tulnp | grep 3306

# 尝试手动连接测试
mysql -h localhost -P 3306 -u root -p --protocol=TCP
```

### 问题 4: 远程连接失败

**错误**:
```
ERROR 2003 (HY000): Can't connect to MySQL server on '192.168.1.100'
```

**解决**:
```bash
# 测试网络连通性
ping 192.168.1.100

# 测试端口是否开放
telnet 192.168.1.100 3306
# 或
nc -zv 192.168.1.100 3306

# 检查防火墙规则
# 检查远程 MySQL 是否允许远程连接
# 检查用户权限: GRANT ALL ON *.* TO 'user'@'%';
```

---

## 使用示例

### 示例 1: 标准用法 (Docker MySQL)

```bash
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/auth-service

# 确保 Docker MySQL 正在运行
docker ps | grep cluster-mysql

# 运行初始化
make init-mysql
```

### 示例 2: 切换到本地 MySQL

```bash
# 1. 修改配置保持 localhost
# configs/config-dev.yaml 保持不变

# 2. 停止 Docker MySQL (如果需要使用本地安装的 MySQL)
docker stop cluster-mysql

# 3. 确保本地 MySQL 运行
sudo systemctl start mysql

# 4. 运行初始化
make init-mysql
# 脚本会自动检测并使用 TCP 连接
```

### 示例 3: 使用远程 MySQL

```bash
# 1. 修改配置文件
# configs/config-dev.yaml:
#   database:
#     host: "mysql.example.com"
#     port: 3306
#     user: "remote_user"
#     password: "remote_pass"

# 2. 运行初始化
make init-mysql
# 脚本会自动使用远程连接
```

---

## 优势

### 1. 灵活性

✅ 自动适配 Docker、本地、远程 MySQL
✅ 无需修改脚本即可切换环境
✅ 基于配置文件，不依赖硬编码

### 2. 易用性

✅ 单一命令 `make init-mysql`
✅ 自动检测最佳连接方式
✅ 清晰的输出信息

### 3. 可靠性

✅ 错误处理和验证
✅ 支持多种 YAML 解析方式
✅ 安全的密码处理

---

## 未来改进

可能的增强功能：

1. [ ] 支持通过环境变量覆盖配置
2. [ ] 支持 SSL/TLS 连接
3. [ ] 添加连接超时配置
4. [ ] 支持 SSH 隧道连接
5. [ ] 添加连接测试命令

---

## 相关文件

- `scripts/init-mysql.sh` - 主脚本
- `configs/config-dev.yaml` - 开发配置
- `configs/config-local.yaml` - 本地配置
- `configs/config-prod.yaml` - 生产配置
- `Makefile` - Make 目标定义

---

## 总结

脚本现在完全基于配置文件工作，能够：

✅ 自动检测并选择最佳连接方式
✅ 支持 Docker、本地、远程 MySQL
✅ 提供清晰的连接信息输出
✅ 安全地处理密码信息
✅ 支持多种 YAML 解析方式

无需修改脚本，只需更改配置文件即可适配不同的 MySQL 环境！

---

**版本**: v2.0
**最后更新**: 2025-10-21
**作者**: Claude Code (AI Assistant)
