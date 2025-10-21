# README.md 更新日志

**日期**: 2025-10-21
**更新内容**: 添加 MySQL 数据库凭据信息

---

## 更新摘要

在 `README.md` 中添加了详细的 MySQL 数据库配置和凭据信息，帮助开发者快速设置本地开发环境。

## 具体更改

### 1. 配置说明部分 (第 161-239 行)

**新增内容**:
- ✅ MySQL 数据库凭据详细说明
- ✅ Root 用户和应用用户凭据
- ✅ 数据库连接信息
- ✅ 快速连接命令示例
- ✅ MySQL 容器管理命令
- ✅ 更新的配置文件示例 (MySQL 替代 PostgreSQL)
- ✅ 相关文档链接

**关键信息**:

**Root 用户**:
- 用户名: `root`
- 密码: `root123`

**应用用户** (推荐):
- 用户名: `cluster_user`
- 密码: `cluster_pass`

**数据库信息**:
- 数据库名: `cluster_db`
- 主机: `localhost`
- 端口: `3306`
- 容器名: `cluster-mysql`

### 2. 快速开始部分 (第 269-304 行)

**更新内容**:
- ✅ 在步骤 0 添加 MySQL 数据库设置说明
- ✅ 强调首次运行必须执行 `./setup-mysql.sh`
- ✅ 添加等待 MySQL 就绪的提示
- ✅ 更新启动命令为 `make run-local` (使用本地 MySQL 配置)
- ✅ 添加 QUICKSTART_GUIDE.md 链接

### 3. 依赖服务部分 (第 403-408 行)

**更新内容**:
- ✅ 将 PostgreSQL 更新为 MySQL 8.0
- ✅ 说明本地开发使用 Docker 容器
- ✅ 添加 setup-mysql.sh 脚本引用
- ✅ 添加凭据参考链接

## 文档结构改进

### 添加的文档链接

1. **[MYSQL_MIGRATION_REPORT.md](./MYSQL_MIGRATION_REPORT.md)** - PostgreSQL 到 MySQL 的迁移报告
2. **[MYSQL_TROUBLESHOOTING.md](./MYSQL_TROUBLESHOOTING.md)** - MySQL 故障排查指南
3. **[QUICKSTART_GUIDE.md](./QUICKSTART_GUIDE.md)** - 完整快速启动指南

### 代码示例

**连接 MySQL**:
```bash
# 使用应用用户连接
docker exec -it cluster-mysql mysql -ucluster_user -pcluster_pass cluster_db

# 使用 root 用户连接
docker exec -it cluster-mysql mysql -uroot -proot123

# 从主机连接
mysql -h127.0.0.1 -P3306 -ucluster_user -pcluster_pass cluster_db
```

**管理命令**:
```bash
# 启动 MySQL
docker start cluster-mysql

# 停止 MySQL
docker stop cluster-mysql

# 查看日志
docker logs cluster-mysql

# 重新设置数据库
./setup-mysql.sh
```

**配置文件示例** (configs/config.yaml):
```yaml
database:
  driver: mysql                    # 使用 MySQL
  host: localhost
  port: 3306                       # MySQL 端口
  user: cluster_user               # MySQL 用户
  password: cluster_pass           # MySQL 密码
  dbname: cluster_db              # 数据库名
  sslmode: disable
```

## 用户体验改进

### 改进前
- ❌ 没有明确的数据库凭据信息
- ❌ 配置示例仍然显示 PostgreSQL
- ❌ 没有 MySQL 设置步骤
- ❌ 缺少快速连接命令

### 改进后
- ✅ 清晰的数据库凭据说明
- ✅ 完整的 MySQL 配置示例
- ✅ 详细的设置步骤
- ✅ 即用的连接命令
- ✅ 完整的管理命令参考
- ✅ 相关文档链接

## 开发者工作流

更新后的 README.md 提供了清晰的开发者工作流:

```bash
# 步骤 0: 设置数据库 (首次)
./setup-mysql.sh

# 步骤 1: 构建
make build

# 步骤 2: 启动 (使用本地 MySQL)
make run-local

# 步骤 3: 测试
curl http://localhost:8082/version
```

## 相关文件

**修改的文件**:
1. `/home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/cluster-service/README.md`

**相关文档**:
1. `setup-mysql.sh` - MySQL 设置脚本
2. `MYSQL_MIGRATION_REPORT.md` - 迁移报告
3. `MYSQL_TROUBLESHOOTING.md` - 故障排查
4. `QUICKSTART_GUIDE.md` - 快速启动指南

## 验证

更新后的文档已经过验证:
- ✅ 所有凭据信息与 `setup-mysql.sh` 脚本一致
- ✅ 所有命令已测试可用
- ✅ 配置示例格式正确
- ✅ 文档链接有效

## 后续建议

1. ✅ README.md 已更新完成
2. ⏳ 考虑在 QUICKSTART_GUIDE.md 中添加更多截图
3. ⏳ 考虑创建视频教程演示 MySQL 设置流程
4. ⏳ 添加常见问题 FAQ 部分

---

**更新者**: Claude Code (AI Assistant)
**状态**: ✅ 完成
**版本**: v1.0.0
