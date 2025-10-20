# Makefile 使用指南

## 快速开始

查看所有可用命令：
```bash
make help
```

查看项目信息：
```bash
make info
```

## 常用命令

### 🚀 运行服务

#### 本地开发（推荐）
```bash
make run-local
# 或简写
make dev
```
使用 `configs/config-local.yaml` 配置运行。

#### 远程开发环境
```bash
make run-dev
```
使用 `configs/config-dev.yaml` 配置运行。

#### 生产环境
```bash
make run-prod
```
使用 `configs/config-prod.yaml` 配置运行（需要设置环境变量）。

#### 默认配置
```bash
make run
```
使用 `configs/config.yaml` 配置运行。

#### 自定义配置
```bash
make run CONFIG=configs/my-config.yaml
```

### 🔨 构建

编译二进制文件：
```bash
make build
```
生成的文件：`bin/auth-service`

### 🗄️ 数据库管理

#### 创建数据库
```bash
make db-create
```

#### 初始化数据库（运行迁移）
```bash
make init-mysql
# 或
make init-db
```

#### 重置数据库（删除+创建+迁移）
```bash
make db-reset
```
⚠️ **警告**: 这会删除所有数据！

#### 删除数据库
```bash
make db-drop
```
会要求确认。

### 🧪 测试

运行所有测试：
```bash
make test
```

运行测试并生成覆盖率报告：
```bash
make test-coverage
```
生成 `coverage.html` 文件。

运行所有检查（格式化、vet、lint、测试）：
```bash
make check
```

### 🔍 代码质量

格式化代码：
```bash
make fmt
```

运行 linter：
```bash
make lint
```

运行 go vet：
```bash
make vet
```

### 📝 配置查看

查看当前使用的配置：
```bash
make show-config
```

查看开发配置：
```bash
make show-config-dev
```

查看本地配置：
```bash
make show-config-local
```

查看生产配置：
```bash
make show-config-prod
```

### 🐳 Docker 命令

构建 Docker 镜像：
```bash
make docker-build
```

运行 Docker 容器：
```bash
make docker-run
```

构建多平台镜像：
```bash
make docker-buildx
```

推送多平台镜像：
```bash
make docker-buildx-push
```

### 🔄 开发工作流

#### 首次设置并运行
```bash
make quick-start
```
这会：
1. 安装依赖
2. 创建数据库
3. 运行迁移
4. 启动服务（本地配置）

#### 热重载开发（推荐）
```bash
# 首次安装 Air
go install github.com/cosmtrek/air@latest

# 启动热重载
make watch
```
代码修改后自动重启服务。

#### 完整开发流程
```bash
# 1. 拉取最新代码
git pull

# 2. 更新依赖
make deps

# 3. 重置数据库（如需要）
make db-reset

# 4. 启动开发服务器
make dev
# 或使用热重载
make watch
```

### 🧹 清理

清理构建产物：
```bash
make clean
```

### 📦 依赖管理

安装/更新依赖：
```bash
make deps
```

### 🏥 健康检查

检查服务是否运行：
```bash
make health
```

查看 Docker 日志：
```bash
make logs
```

## 完整命令列表

| 命令 | 说明 |
|------|------|
| **运行命令** | |
| `make run` | 使用默认配置运行 |
| `make run-dev` | 使用开发配置运行 |
| `make run-local` | 使用本地配置运行 |
| `make run-prod` | 使用生产配置运行 |
| `make dev` | `run-local` 的别名 |
| `make watch` | 热重载开发模式 |
| **构建命令** | |
| `make build` | 构建二进制文件 |
| **数据库命令** | |
| `make init-db` | 初始化数据库 |
| `make init-mysql` | 初始化 MySQL 数据库 |
| `make db-create` | 创建数据库 |
| `make db-drop` | 删除数据库 |
| `make db-reset` | 重置数据库 |
| **测试命令** | |
| `make test` | 运行测试 |
| `make test-coverage` | 运行测试并生成覆盖率 |
| `make vet` | 运行 go vet |
| `make check` | 运行所有检查 |
| **代码质量** | |
| `make fmt` | 格式化代码 |
| `make lint` | 运行 linter |
| **配置命令** | |
| `make show-config` | 显示当前配置 |
| `make show-config-dev` | 显示开发配置 |
| `make show-config-local` | 显示本地配置 |
| `make show-config-prod` | 显示生产配置 |
| **Docker 命令** | |
| `make docker-build` | 构建 Docker 镜像 |
| `make docker-run` | 运行 Docker 容器 |
| `make docker-buildx` | 构建多平台镜像 |
| `make docker-buildx-push` | 推送多平台镜像 |
| **工具命令** | |
| `make deps` | 管理依赖 |
| `make clean` | 清理构建产物 |
| `make health` | 健康检查 |
| `make logs` | 查看 Docker 日志 |
| `make info` | 显示项目信息 |
| `make help` | 显示帮助信息 |
| **快捷命令** | |
| `make quick-start` | 快速开始（首次） |

## 环境变量

某些命令支持环境变量配置：

```bash
# 自定义配置文件
CONFIG=configs/my-config.yaml make run

# Docker 镜像配置
DOCKER_REGISTRY=myregistry make docker-build
IMAGE_NAME=my-auth-service make docker-build
IMAGE_TAG=v1.0.0 make docker-build
ENV=staging make docker-buildx-push-env

# 多平台构建
PLATFORMS=linux/amd64,linux/arm64 make docker-buildx
```

## 使用示例

### 场景 1: 首次开发设置

```bash
# 1. 克隆仓库
git clone <repo>
cd k8s-agent/auth-service

# 2. 快速启动
make quick-start
```

### 场景 2: 日常开发

```bash
# 启动热重载开发服务器
make watch

# 在另一个终端运行测试
make test

# 检查代码质量
make check
```

### 场景 3: 准备发布

```bash
# 运行所有检查
make check

# 构建
make build

# 构建并推送 Docker 镜像
make docker-buildx-push IMAGE_TAG=v1.2.3
```

### 场景 4: 数据库重置

```bash
# 重置数据库（会确认）
make db-reset

# 或手动步骤
make db-drop
make db-create
make init-mysql
```

### 场景 5: 生产部署

```bash
# 设置环境变量
export DB_HOST=prod-mysql.example.com
export DB_PASSWORD=secure-password
export JWT_SECRET=your-super-secret-key

# 运行
make run-prod

# 或构建镜像
make docker-buildx-push-env ENV=prod
```

## 故障排查

### 命令失败？

1. **数据库连接失败**:
   ```bash
   # 检查 MySQL 是否运行
   mysql -h localhost -u root -p

   # 查看配置
   make show-config-local
   ```

2. **Air 未找到**:
   ```bash
   # 安装 Air
   go install github.com/cosmtrek/air@latest
   ```

3. **golangci-lint 未找到**:
   ```bash
   # 安装 golangci-lint
   brew install golangci-lint
   # 或
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

4. **端口已占用**:
   ```bash
   # 查找占用端口的进程
   lsof -i :8090

   # 终止进程
   kill -9 <PID>
   ```

## 最佳实践

1. **开发时使用热重载**:
   ```bash
   make watch
   ```

2. **提交前运行检查**:
   ```bash
   make check
   ```

3. **定期更新依赖**:
   ```bash
   make deps
   ```

4. **使用正确的配置文件**:
   - 本地开发: `make run-local`
   - 远程开发: `make run-dev`
   - 生产环境: `make run-prod`

5. **Docker 构建时使用多平台**:
   ```bash
   make docker-buildx-push
   ```

## 参考资料

- [Makefile 语法](https://makefiletutorial.com/)
- [Air 文档](https://github.com/cosmtrek/air)
- [golangci-lint 文档](https://golangci-lint.run/)
- [Docker Buildx 文档](https://docs.docker.com/buildx/working-with-buildx/)
