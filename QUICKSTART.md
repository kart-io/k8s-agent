# 快速启动指南

本指南帮助您快速启动和测试重构后的 K8s Agent 项目 (前端 + 后端)。

---

## 前置要求

### 后端
- Go 1.21+
- 可访问的 Kubernetes 集群 (可选)

### 前端
- Node.js 16+
- npm 或 yarn

---

## 一、后端快速启动

### 1. 进入项目目录

```bash
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/cluster-service
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 编译项目

```bash
go build -o bin/cluster-service ./cmd/server
```

### 4. 运行服务

```bash
./bin/cluster-service
```

服务将在 `http://localhost:8080` 启动。

### 5. 测试 API

**使用测试脚本**:
```bash
./scripts/test_query_params_api.sh
```

**手动测试**:
```bash
# 测试集群 API
curl -X GET "http://localhost:8080/api/k8s/clusters"

# 测试 Pod API (需要真实的集群和 namespace)
curl -X GET "http://localhost:8080/api/k8s/pods?clusterId=default-cluster&namespace=default"
```

---

## 二、前端快速启动

### 1. 进入项目目录

```bash
cd /home/hellotalk/code/web-k8s-agent-web/apps/web-k8s
```

### 2. 安装依赖

```bash
npm install
```

### 3. 配置环境变量

创建或编辑 `.env.development`:

```env
# API 基础地址
VITE_API_BASE_URL=http://127.0.0.1:8080

# 是否使用 Mock 数据 (推荐先用 Mock 测试)
VITE_USE_K8S_MOCK=true

# 应用配置
VITE_APP_TITLE=K8s Management Platform
VITE_APP_PORT=5670
```

### 4. 启动开发服务器

```bash
npm run dev
```

前端将在 `http://localhost:5670` 启动。

### 5. 访问应用

在浏览器中打开 `http://localhost:5670`

---

## 三、使用 Mock 数据测试 (推荐首选)

### 为什么使用 Mock?

- ✅ 无需配置真实的 Kubernetes 集群
- ✅ 快速验证前端功能
- ✅ 测试各种数据场景
- ✅ 独立于后端进行开发

### 启用 Mock

**方式 1: 环境变量** (推荐)

`.env.development`:
```env
VITE_USE_K8S_MOCK=true
```

**方式 2: 在 API 函数中集成**

编辑 `src/api/k8s/index.ts`,在每个 API 函数中添加:

```typescript
import { mockGetPods } from './mock'

export function getPods(params?: PageParams): Promise<PageResponse<Pod>> {
  // 检查环境变量
  if (import.meta.env.VITE_USE_K8S_MOCK === 'true') {
    return mockGetPods(params || {})
  }

  // 调用真实 API
  const clusterId = getClusterId()
  return http.get('/api/k8s/pods', {
    params: { clusterId, ...params }
  })
}
```

### Mock 数据说明

Mock 系统提供以下测试数据:
- **4 个 Namespace**: default, kube-system, kube-public, production
- **3 个 Pod**: nginx-deployment-xxx, redis-master-0, coredns-xxx
- **2 个 Deployment**: nginx-deployment, api-server
- **2 个 Node**: node-1 (m5.large), node-2 (m5.xlarge)
- **2 个 Service**: nginx-service, kubernetes
- **其他资源**: Event, ConfigMap, Secret 等

---

## 四、前后端联调测试

### 1. 启动后端

```bash
cd k8s-agent/cluster-service
./bin/cluster-service &
```

### 2. 配置前端连接真实后端

`.env.development`:
```env
# 连接真实后端
VITE_API_BASE_URL=http://127.0.0.1:8080

# 禁用 Mock
VITE_USE_K8S_MOCK=false
```

### 3. 启动前端

```bash
cd web-k8s-agent-web/apps/web-k8s
npm run dev
```

### 4. 测试集成

1. 在浏览器打开 `http://localhost:5670`
2. 打开 Chrome DevTools > Network 面板
3. 刷新页面,检查 API 请求:
   - 请求 URL 格式: `/api/k8s/xxx?clusterId=xxx&...`
   - 响应状态: 200 OK
   - 响应数据格式正确

---

## 五、测试集群切换功能

### 1. 使用 ClusterSelector 组件

在任意页面中添加:

```vue
<template>
  <div>
    <ClusterSelector />
    <!-- 您的页面内容 -->
  </div>
</template>

<script setup lang="ts">
import ClusterSelector from '@/components/ClusterSelector.vue'
</script>
```

### 2. 手动切换集群

在浏览器控制台:

```javascript
// 切换到另一个集群
const clusterStore = window.__PINIA__.state.value.cluster
clusterStore.currentClusterId = 'another-cluster-id'

// 检查是否生效
console.log(localStorage.getItem('currentClusterId'))
```

### 3. 验证 API 调用

切换集群后,检查 Network 面板中新的 API 请求是否包含新的 `clusterId` 参数。

---

## 六、常见测试场景

### 场景 1: 测试 Pod 列表

```bash
# 后端 (使用 cURL)
curl -X GET "http://localhost:8080/api/k8s/pods?clusterId=default-cluster&namespace=default&page=1&pageSize=20"

# 前端 (在组件中)
const pods = await getPods({
  namespace: 'default',
  page: 1,
  pageSize: 20
})
```

### 场景 2: 测试 Pod 详情

```bash
# 后端
curl -X GET "http://localhost:8080/api/k8s/pod?clusterId=default-cluster&namespace=default&name=nginx-pod"

# 前端
const pod = await getPod('default', 'nginx-pod')
```

### 场景 3: 测试 Pod 日志

```bash
# 后端
curl -X GET "http://localhost:8080/api/k8s/pod/logs?clusterId=default-cluster&namespace=default&name=nginx-pod&container=nginx&tailLines=100"

# 前端
const logs = await getPodLogs('default', 'nginx-pod', 'nginx', 100)
```

### 场景 4: 测试 Deployment 扩缩容

```bash
# 后端
curl -X PUT "http://localhost:8080/api/k8s/deployment/scale?clusterId=default-cluster&namespace=default&name=nginx-deployment" \
  -H "Content-Type: application/json" \
  -d '{"replicas": 5}'

# 前端
const deployment = await scaleDeployment('default', 'nginx-deployment', 5)
```

---

## 七、故障排查

### 问题 1: 前端无法连接后端

**症状**: Network 面板显示 `ERR_CONNECTION_REFUSED`

**解决方案**:
1. 检查后端是否运行: `curl http://localhost:8080/api/k8s/clusters`
2. 检查 `.env.development` 中的 `VITE_API_BASE_URL`
3. 检查防火墙设置

### 问题 2: API 返回 400 Bad Request

**症状**: 响应消息 "Invalid query parameters"

**解决方案**:
1. 检查 clusterId 是否传递
2. 检查必需参数是否缺失 (namespace, name 等)
3. 查看浏览器控制台错误消息
4. 检查 Network 面板中的请求 URL

### 问题 3: 集群 ID 始终是 default-cluster

**症状**: 无法切换集群

**解决方案**:
1. 检查 `src/stores/cluster.ts` 是否正确导入
2. 检查 Pinia 是否正确初始化
3. 在控制台检查: `window.__PINIA__.state.value.cluster`
4. 清除 localStorage: `localStorage.clear()`

### 问题 4: Mock 数据不生效

**症状**: 即使设置了 `VITE_USE_K8S_MOCK=true` 仍然调用真实 API

**解决方案**:
1. 确认在 API 函数中添加了 Mock 检查代码
2. 重启开发服务器: `npm run dev`
3. 检查环境变量: `console.log(import.meta.env.VITE_USE_K8S_MOCK)`

### 问题 5: TypeScript 类型错误

**症状**: IDE 显示类型错误

**解决方案**:
1. 重新安装依赖: `npm install`
2. 重启 IDE 的 TypeScript 服务器
3. 检查 `src/api/k8s/types.ts` 是否正确

---

## 八、性能测试

### 1. 后端性能

```bash
# 使用 Apache Bench 测试
ab -n 1000 -c 10 "http://localhost:8080/api/k8s/clusters"

# 使用 hey 测试
hey -n 1000 -c 10 "http://localhost:8080/api/k8s/clusters"
```

### 2. 前端性能

1. 打开 Chrome DevTools > Performance
2. 点击 Record
3. 进行一些操作 (列表加载、切换集群等)
4. 停止 Record,分析性能瓶颈

### 3. 网络性能

在 Network 面板:
- 检查 API 响应时间
- 检查数据大小
- 检查并发请求数量

---

## 九、生产部署准备

### 1. 前端生产构建

```bash
cd web-k8s-agent-web/apps/web-k8s

# 配置生产环境变量
cat > .env.production << EOF
VITE_API_BASE_URL=https://api.your-domain.com
VITE_USE_K8S_MOCK=false
EOF

# 构建
npm run build

# 构建产物在 dist/ 目录
ls -lh dist/
```

### 2. 后端生产构建

```bash
cd k8s-agent/cluster-service

# 使用版本信息构建
go build -ldflags "
  -X 'main.Version=v1.0.0'
  -X 'main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)'
  -X 'main.GitCommit=$(git rev-parse HEAD)'
" -o bin/cluster-service ./cmd/server

# 检查二进制
./bin/cluster-service --version
```

### 3. Docker 部署 (可选)

**后端 Dockerfile**:
```dockerfile
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o cluster-service ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/cluster-service .
EXPOSE 8080
CMD ["./cluster-service"]
```

**前端 Dockerfile**:
```dockerfile
FROM node:18 AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

---

## 十、下一步

### 立即可做

- [x] ✅ 启动后端服务
- [x] ✅ 启动前端服务 (使用 Mock)
- [ ] ⏳ 测试前端 UI 功能
- [ ] ⏳ 集成 Mock 到所有 API 函数
- [ ] ⏳ 前后端联调测试

### 生产前必做

- [ ] 完整的集成测试
- [ ] 性能测试和优化
- [ ] 安全审计
- [ ] 文档完善
- [ ] 监控和日志配置

### 可选优化

- [ ] 添加单元测试
- [ ] 实现 API 缓存
- [ ] 添加国际化支持
- [ ] 优化打包大小
- [ ] 实现 PWA 支持

---

## 参考文档

- **完整重构总结**: `k8s-agent/COMPLETE_REFACTORING_SUMMARY.md`
- **后端迁移指南**: `k8s-agent/docs/API_MIGRATION_QUERY_PARAMS.md`
- **前端重构总结**: `web-k8s-agent-web/apps/web-k8s/FRONTEND_REFACTORING_SUMMARY.md`
- **前端集成指南**: `web-k8s-agent-web/apps/web-k8s/INTEGRATION_GUIDE.md`
- **API 配置文档**: `web-k8s-agent-web/apps/web-k8s/API_CONFIG.md`

---

**版本**: 1.0.0
**最后更新**: 2025-10-21
**作者**: Claude Code (AI Assistant)

祝您使用愉快! 🚀
