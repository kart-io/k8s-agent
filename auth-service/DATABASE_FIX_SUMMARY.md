# Auth Service 数据库配置修复总结

**日期**: 2025-10-21
**状态**: ✅ 已完成

---

## 问题
auth-service 无法连接 MySQL 数据库

## 原因
1. 配置文件中 MySQL 密码错误 (`root` 应为 `root123`)
2. MySQL 容器中缺少 `user_auth` 数据库

## 解决方案

### 1. 创建数据库 ✅
```bash
docker exec cluster-mysql mysql -uroot -proot123 -e \
  "CREATE DATABASE IF NOT EXISTS user_auth CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

### 2. 更新配置文件 ✅
- `configs/config.yaml`: 密码从 `root` → `root123`
- `configs/config-dev.yaml`: ✅ 已正确
- `configs/config-local.yaml`: ✅ 已正确

---

## MySQL 数据库凭据

**容器**: `cluster-mysql`
**Root 用户**: `root` / `root123`
**数据库**: `user_auth`
**端口**: `3306`

**快速连接**:
```bash
docker exec -it cluster-mysql mysql -uroot -proot123 user_auth
```

---

## 共享 MySQL 容器

| 数据库 | 服务 |
|--------|------|
| `cluster_db` | cluster-service |
| `user_auth` | auth-service |

---

## 启动 auth-service

```bash
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/auth-service

# 构建
make build

# 启动（开发配置）
./bin/auth-service -c configs/config-dev.yaml
```

---

## 修改的文件

1. ✅ `configs/config.yaml` - 密码已更新
2. ✅ `README.md` - 添加数据库配置说明
3. ✅ `DATABASE_FIX_REPORT.md` - 详细修复报告

---

## 详细文档

查看 [DATABASE_FIX_REPORT.md](./DATABASE_FIX_REPORT.md) 获取完整信息。

---

**修复者**: Claude Code
