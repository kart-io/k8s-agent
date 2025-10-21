# Auth Service 登录 401 错误修复报告

**日期**: 2025-10-21
**问题**: 登录接口返回 401 Unauthorized
**状态**: ✅ 已修复

---

## 问题描述

### 症状

```bash
$ curl -X POST http://localhost:5668/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

返回:
{
  "code": 401,
  "message": "Invalid username or password"
}
```

---

## 根本原因

SQL 初始化脚本 `scripts/init-mysql.sql` 中的 admin 用户密码哈希值**不正确**。

### 错误的密码哈希 (旧)

```sql
-- 不完整/错误的 bcrypt hash
INSERT INTO users ... VALUES
('user-admin', 'admin', '$2a$10$rI5JhXWJIkGKhKqWQqQqJ.MkN5JZ5JQqQqJ.MkN5JZ5JQqQqJ.MkN5', ...)
```

这个哈希值看起来不完整，导致密码验证失败。

---

## 解决方案

### 1. 生成正确的密码哈希

使用项目提供的密码哈希工具：

```bash
$ cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/auth-service
$ go run scripts/hash_password.go admin123

输出:
$2a$10$fSA0jhcFVnG..gkMmi5Ypug2YFwjVHCd7rveiwp6XUJGpYlHQ5ZSK
```

### 2. 更新数据库中的密码

```bash
docker exec cluster-mysql mysql -uroot -proot123 -e \
  "USE user_auth;
   UPDATE users
   SET password = '\$2a\$10\$fSA0jhcFVnG..gkMmi5Ypug2YFwjVHCd7rveiwp6XUJGpYlHQ5ZSK'
   WHERE username = 'admin';"
```

### 3. 更新 SQL 脚本

修改了 `scripts/init-mysql.sql` (第 107 行):

```sql
-- 正确的 bcrypt hash for "admin123"
INSERT INTO users (id, username, password, email, real_name, status, created_at, updated_at) VALUES
('user-admin', 'admin', '$2a$10$fSA0jhcFVnG..gkMmi5Ypug2YFwjVHCd7rveiwp6XUJGpYlHQ5ZSK', 'admin@example.com', '超级管理员', 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE password=VALUES(password);  -- ← 更新为也更新密码
```

**关键改动**:
- ✅ 使用正确的 bcrypt 哈希
- ✅ 修改 `ON DUPLICATE KEY UPDATE` 为 `password=VALUES(password)`，确保重新运行时会更新密码

---

## 验证

### 测试登录

```bash
$ curl -X POST "http://localhost:5668/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"admin123\"}"

成功响应:
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "jti": "61917a7a-b7dd-497e-bd2c-e2c4d2f9ff76",
    "expires_at": "2025-10-22T14:30:30+08:00",
    "user": {
      "id": "user-admin",
      "username": "admin",
      "email": "admin@example.com",
      "real_name": "超级管理员",
      "roles": [
        {
          "id": "role-super-admin",
          "name": "超级管理员",
          "code": "super_admin",
          "description": "系统超级管理员，拥有所有权限"
        }
      ]
    }
  }
}
```

✅ **登录成功！** 返回了 JWT token 和用户信息。

---

## 默认管理员账号

### 登录凭据

- **用户名**: `admin`
- **密码**: `admin123`
- **邮箱**: `admin@example.com`

### 用户信息

- **ID**: `user-admin`
- **角色**: 超级管理员 (super_admin)
- **权限**: 所有权限

⚠️ **安全提示**: 首次登录后请立即修改密码！

---

## 密码哈希工具

项目提供了密码哈希生成工具 `scripts/hash_password.go`：

### 使用方法

```bash
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/auth-service

# 生成密码哈希
go run scripts/hash_password.go <你的密码>

# 示例
go run scripts/hash_password.go mypassword123
```

### 工具代码

```go
package main

import (
    "fmt"
    "os"

    "golang.org/x/crypto/bcrypt"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run hash_password.go <password>")
        os.Exit(1)
    }

    password := os.Args[1]
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }

    fmt.Println(string(hash))
}
```

---

## 重新初始化数据库

如果需要重新初始化数据库（会使用修复后的密码哈希）：

```bash
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/auth-service

# 方法 1: 完全重置 (删除并重建)
make db-reset
# 会提示确认: Are you sure? (y/N):
# 输入 y 确认

# 方法 2: 仅重新运行初始化脚本
make init-mysql
```

---

## API 测试示例

### 1. 登录获取 Token

```bash
# Linux/macOS
TOKEN=$(curl -s -X POST "http://localhost:5668/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')

echo "Token: $TOKEN"
```

### 2. 使用 Token 访问受保护的 API

```bash
# 获取用户列表
curl -X GET "http://localhost:5668/api/v1/users" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"

# 获取当前用户信息
curl -X GET "http://localhost:5668/api/v1/auth/me" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"
```

### 3. 修改密码

```bash
# 修改当前用户密码
curl -X PUT "http://localhost:5668/api/v1/users/user-admin/password" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "admin123",
    "new_password": "new_secure_password_123"
  }'
```

### 4. 登出

```bash
curl -X POST "http://localhost:5668/api/v1/auth/logout" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"
```

---

## 前端集成

### Vue 3 + Axios 示例

```javascript
// api/auth.js
import axios from 'axios'

const API_BASE = 'http://localhost:5668/api/v1'

export const authAPI = {
  // 登录
  async login(username, password) {
    const response = await axios.post(`${API_BASE}/auth/login`, {
      username,
      password
    })

    if (response.data.code === 0) {
      // 保存 token
      localStorage.setItem('token', response.data.data.token)
      localStorage.setItem('user', JSON.stringify(response.data.data.user))
      return response.data.data
    } else {
      throw new Error(response.data.message)
    }
  },

  // 获取当前用户信息
  async getCurrentUser() {
    const token = localStorage.getItem('token')
    const response = await axios.get(`${API_BASE}/auth/me`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })
    return response.data.data
  },

  // 登出
  async logout() {
    const token = localStorage.getItem('token')
    await axios.post(`${API_BASE}/auth/logout`, {}, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })

    // 清除本地存储
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }
}
```

### 使用示例

```javascript
// 在 Vue 组件中
import { authAPI } from '@/api/auth'

// 登录
try {
  const result = await authAPI.login('admin', 'admin123')
  console.log('登录成功:', result)
  console.log('Token:', result.token)
  console.log('用户信息:', result.user)
} catch (error) {
  console.error('登录失败:', error.message)
}
```

---

## 常见问题

### Q1: 仍然无法登录

**A**: 检查以下几点：

1. **确认密码已更新**:
```bash
docker exec cluster-mysql mysql -uroot -proot123 -e \
  "USE user_auth;
   SELECT username, LEFT(password, 30) as password_hash
   FROM users WHERE username='admin';"
```

应该显示: `$2a$10$fSA0jhcFVnG..gkMmi5Ypug`

2. **确认服务正在运行**:
```bash
curl http://localhost:5668/health
```

3. **检查请求格式**:
```bash
# 正确的格式
curl -X POST "http://localhost:5668/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### Q2: Token 过期了怎么办？

**A**: 重新登录获取新的 token。

Token 默认有效期为 24 小时（在 `configs/config.yaml` 中配置）：

```yaml
jwt:
  expires_hours: 24  # 可以修改这个值
```

### Q3: 如何创建新用户？

**A**: 使用管理员 token 调用创建用户 API：

```bash
curl -X POST "http://localhost:5668/api/v1/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "password": "password123",
    "email": "newuser@example.com",
    "real_name": "新用户",
    "role_ids": ["role-user"]
  }'
```

---

## 修改的文件

1. ✅ `scripts/init-mysql.sql` (第 104-108 行)
   - 更新了 admin 用户的密码哈希
   - 修改了 ON DUPLICATE KEY UPDATE 逻辑

2. ✅ 数据库 `user_auth.users` 表
   - 更新了 admin 用户的密码字段

---

## 相关文档

- `INIT_MYSQL_FIX_REPORT.md` - init-mysql 修复报告
- `DATABASE_FIX_REPORT.md` - 数据库配置修复报告
- `README.md` - auth-service 使用文档

---

## 总结

✅ **问题已修复**: admin 用户密码哈希已更正
✅ **SQL 脚本已更新**: 重新初始化时会使用正确的密码
✅ **登录成功**: 可以使用 admin/admin123 登录
✅ **获取 Token**: 登录后返回 JWT token 和用户信息

**现在可以正常使用 auth-service 的所有功能了！** 🎉

⚠️ **重要提醒**: 首次登录后，请立即修改默认管理员密码以确保安全！

---

**修复时间**: 2025-10-21
**修复者**: Claude Code (AI Assistant)
**状态**: ✅ 完成并验证
