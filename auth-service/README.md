# Auth Service

认证授权服务，提供用户认证、权限管理和 API 授权功能。

## 功能特性

### 1. 用户认证
- 用户登录/登出
- JWT Token 认证
- 密码加密存储（bcrypt）
- Token 刷新机制

### 2. 角色权限管理（RBAC）
- 用户管理（CRUD）
- 角色管理（CRUD）
- 权限管理（CRUD）
- 用户-角色关联
- 角色-权限关联

### 3. 权限类型
- **菜单权限**: 控制前端菜单显示
- **按钮权限**: 控制页面内按钮显示
- **API权限**: 控制接口访问

### 4. API Key 授权
- API Key 生成和管理
- API Key 认证
- API Key 过期管理

## 数据库表结构

### users (用户表)
```sql
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE,
    real_name VARCHAR(50),
    phone VARCHAR(20),
    avatar VARCHAR(255),
    status INT DEFAULT 1,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### roles (角色表)
```sql
CREATE TABLE roles (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    status INT DEFAULT 1,
    sort INT DEFAULT 0,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### permissions (权限表)
```sql
CREATE TABLE permissions (
    id VARCHAR(36) PRIMARY KEY,
    parent_id VARCHAR(36),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL, -- menu, button, api
    path VARCHAR(255),
    method VARCHAR(10),
    component VARCHAR(255),
    icon VARCHAR(50),
    sort INT DEFAULT 0,
    status INT DEFAULT 1,
    description TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### user_roles (用户角色关联表)
```sql
CREATE TABLE user_roles (
    user_id VARCHAR(36),
    role_id VARCHAR(36),
    PRIMARY KEY (user_id, role_id)
);
```

### role_permissions (角色权限关联表)
```sql
CREATE TABLE role_permissions (
    role_id VARCHAR(36),
    permission_id VARCHAR(36),
    PRIMARY KEY (role_id, permission_id)
);
```

### api_keys (API密钥表)
```sql
CREATE TABLE api_keys (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    key VARCHAR(255) UNIQUE NOT NULL,
    secret VARCHAR(255) NOT NULL,
    user_id VARCHAR(36),
    description TEXT,
    expires_at TIMESTAMP,
    status INT DEFAULT 1,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

## API 接口

### 认证接口

#### 1. 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
    "username": "admin",
    "password": "password123"
}

Response:
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-01-02T00:00:00Z",
    "user": {
        "id": "uuid",
        "username": "admin",
        "email": "admin@example.com",
        "real_name": "Administrator",
        "roles": [...]
    }
}
```

#### 2. 获取用户信息
```http
GET /api/v1/auth/me
Authorization: Bearer <token>

Response:
{
    "id": "uuid",
    "username": "admin",
    "email": "admin@example.com",
    "roles": [...]
}
```

#### 3. 获取用户菜单
```http
GET /api/v1/auth/menus
Authorization: Bearer <token>

Response:
[
    {
        "id": "1",
        "name": "Dashboard",
        "path": "/dashboard",
        "icon": "dashboard",
        "children": [...]
    }
]
```

### 用户管理接口

```http
GET    /api/v1/users          # 获取用户列表
GET    /api/v1/users/:id      # 获取用户详情
POST   /api/v1/users          # 创建用户
PUT    /api/v1/users/:id      # 更新用户
DELETE /api/v1/users/:id      # 删除用户
```

### 角色管理接口

```http
GET    /api/v1/roles          # 获取角色列表
GET    /api/v1/roles/:id      # 获取角色详情
POST   /api/v1/roles          # 创建角色
PUT    /api/v1/roles/:id      # 更新角色
DELETE /api/v1/roles/:id      # 删除角色
POST   /api/v1/roles/:id/permissions  # 分配权限
```

### 权限管理接口

```http
GET    /api/v1/permissions          # 获取权限列表
GET    /api/v1/permissions/:id      # 获取权限详情
POST   /api/v1/permissions          # 创建权限
PUT    /api/v1/permissions/:id      # 更新权限
DELETE /api/v1/permissions/:id      # 删除权限
GET    /api/v1/permissions/tree     # 获取权限树
```

### API Key 管理接口

```http
GET    /api/v1/api-keys        # 获取 API Key 列表
POST   /api/v1/api-keys        # 创建 API Key
DELETE /api/v1/api-keys/:id    # 删除 API Key
```

## 权限校验中间件

### JWT 认证中间件
```go
// 使用方式
router.Use(middleware.JWTAuth())
```

### 权限检查中间件
```go
// 检查是否有特定权限
router.Use(middleware.RequirePermission("user:create"))

// 检查是否有特定角色
router.Use(middleware.RequireRole("admin"))
```

### API Key 认证中间件
```go
// 用于 API 调用
router.Use(middleware.APIKeyAuth())
```

## 初始化数据

系统会自动创建以下初始数据：

1. **超级管理员**
   - 用户名: admin
   - 密码: admin123
   - 角色: 超级管理员

2. **默认角色**
   - 超级管理员 (super_admin)
   - 管理员 (admin)
   - 普通用户 (user)

3. **默认权限**
   - 系统管理相关权限
   - 用户管理相关权限
   - 角色管理相关权限
   - 权限管理相关权限

## 运行服务

### 使用 Docker Compose (推荐)

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f auth-service

# 停止服务
docker-compose down
```

服务地址:
- Auth Service: http://localhost:8080
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

### 手动运行

```bash
# 初始化依赖
go mod tidy

# 初始化 MySQL 数据库
make init-db

# 运行服务
go run cmd/server/main.go

# 或使用 Makefile
make run
```

## 配置说明

### 配置文件

编辑 `configs/config.yaml` 文件：

```yaml
server:
  port: 8090        # 服务端口

database:
  host: localhost
  port: 3306        # MySQL 端口
  user: root
  password: root
  dbname: k8s_agent_auth
  charset: utf8mb4
  parse_time: true

redis:
  host: localhost
  port: 6379

jwt:
  secret: "your-secret-key"  # JWT 密钥，生产环境必须修改
  expires_hours: 24          # Token 过期时间（小时）
```

### 环境变量配置

JWT 配置支持通过环境变量覆盖配置文件中的值，这在容器化部署时特别有用：

```bash
# JWT 密钥（推荐在生产环境使用环境变量）
export JWT_SECRET="your-super-secret-jwt-key-change-in-production"

# Token 过期时间（小时）
export JWT_EXPIRES_HOURS=24

# 运行服务
go run cmd/server/main.go -c configs/config.yaml
```

### Docker 部署时设置 JWT 密钥

使用 docker run：
```bash
docker run -d \
  -e JWT_SECRET="your-super-secret-jwt-key" \
  -e JWT_EXPIRES_HOURS=24 \
  -p 8090:8090 \
  auth-service:latest
```

使用 docker-compose（已在 docker-compose.yml 中配置）：
```yaml
services:
  auth-service:
    environment:
      - JWT_SECRET=your-super-secret-jwt-key-change-in-production
      - JWT_EXPIRES_HOURS=24
```

### Kubernetes 部署时设置 JWT 密钥

使用 Secret：
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: auth-service-secret
type: Opaque
stringData:
  jwt-secret: "your-super-secret-jwt-key"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
spec:
  template:
    spec:
      containers:
      - name: auth-service
        image: auth-service:latest
        env:
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: auth-service-secret
              key: jwt-secret
        - name: JWT_EXPIRES_HOURS
          value: "24"
```

## 前端集成

### 1. 登录
```javascript
// 登录
const login = async (username, password) => {
  const response = await axios.post('/api/v1/auth/login', {
    username,
    password
  });
  // 保存 token
  localStorage.setItem('token', response.data.token);
  return response.data.user;
};
```

### 2. 请求拦截器
```javascript
// axios 拦截器添加 token
axios.interceptors.request.use(config => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
```

### 3. 获取菜单
```javascript
const getMenus = async () => {
  const response = await axios.get('/api/v1/auth/menus');
  return response.data;
};
```

### 4. 权限指令
```vue
<!-- Vue 示例 -->
<button v-permission="'user:create'">创建用户</button>
<button v-permission="'user:delete'">删除用户</button>
```

## 安全建议

1. **生产环境**:
   - 修改 JWT secret
   - 使用 HTTPS
   - 设置合理的 Token 过期时间
   - 启用 Redis 存储 Token 黑名单

2. **密码策略**:
   - 要求强密码
   - 定期修改密码
   - 密码加密存储

3. **API Key**:
   - 设置过期时间
   - 定期轮换
   - 记录使用日志

## 开发计划

- [x] 基础认证功能
- [x] RBAC 权限模型
- [x] API Key 支持
- [ ] OAuth 2.0 支持
- [ ] 单点登录（SSO）
- [ ] 多因素认证（MFA）
- [ ] 审计日志
