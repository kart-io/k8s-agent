# init-mysql 修复总结

**日期**: 2025-10-21
**状态**: ✅ 已完成

---

## 问题
```bash
$ make init-mysql
ERROR 2002 (HY000): Can't connect to local MySQL server through socket '/tmp/mysql.sock' (2)
```

## 原因
脚本尝试通过 Unix socket 连接本地 MySQL，但 MySQL 运行在 Docker 容器中。

## 解决方案 ✅

修改 `scripts/init-mysql.sh`，添加 Docker 检测：

```bash
# 自动检测 Docker 容器
if docker ps | grep -q cluster-mysql; then
    # 使用 Docker exec 连接
    MYSQL_CMD="docker exec -i cluster-mysql mysql -u $DB_USER -p$DB_PASSWORD"
else
    # 使用 TCP 连接本地 MySQL
    MYSQL_CMD="mysql -h $DB_HOST -P $DB_PORT -u $DB_USER --protocol=TCP -p$DB_PASSWORD"
fi
```

---

## 现在可以使用 ✅

```bash
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/auth-service

# 初始化数据库
make init-mysql

# 或完整启动流程
make quick-start
```

---

## 初始化内容

### 创建的表
- `users` - 用户表
- `roles` - 角色表
- `permissions` - 权限表
- `user_roles` - 用户角色关联
- `role_permissions` - 角色权限关联
- `api_keys` - API 密钥表

### 默认管理员账号
- **用户名**: `admin`
- **密码**: `admin123`
- **邮箱**: `admin@example.com`

⚠️ **请立即修改默认密码！**

---

## 验证

```bash
# 检查表
docker exec cluster-mysql mysql -uroot -proot123 -e "USE user_auth; SHOW TABLES;"

# 查看管理员
docker exec cluster-mysql mysql -uroot -proot123 -e "USE user_auth; SELECT username, email FROM users WHERE username='admin';"
```

---

## 相关命令

```bash
make init-mysql    # 初始化数据库
make db-create     # 仅创建数据库
make db-reset      # 重置数据库 (drop + create + init)
make run-dev       # 启动服务 (开发配置)
make quick-start   # 完整启动流程
```

---

## 详细文档

查看 [INIT_MYSQL_FIX_REPORT.md](./INIT_MYSQL_FIX_REPORT.md) 获取完整信息。

---

**修复者**: Claude Code
**文件**: `scripts/init-mysql.sh`
