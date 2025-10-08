# Agent Manager UI 微前端架构分析

## 📊 当前应用分析

### 现有结构

```
agent-manager-ui/src/
├── api/                    # API 接口层
│   ├── agent.js           # Agent 相关接口
│   ├── alert.js           # 告警相关接口
│   ├── auth.js            # 认证相关接口
│   ├── cluster.js         # 集群相关接口
│   ├── command.js         # 命令相关接口
│   ├── event.js           # 事件相关接口
│   ├── user.js            # 用户相关接口
│   └── request.js         # Axios 封装
├── views/                  # 页面组件
│   ├── Dashboard.vue      # 仪表盘
│   ├── Login.vue          # 登录页
│   ├── agents/            # Agent 管理
│   │   └── AgentList.vue
│   ├── events/            # 事件监控
│   │   └── EventList.vue
│   ├── commands/          # 命令执行
│   │   └── CommandList.vue
│   ├── clusters/          # 集群管理
│   │   └── ClusterList.vue
│   └── alerts/            # 告警规则
│       └── AlertList.vue
├── layouts/               # 布局组件
│   └── MainLayout.vue     # 主布局
├── store/                 # 状态管理
│   └── user.js            # 用户状态
├── router/                # 路由
│   └── index.js
├── directives/            # 指令
│   └── permission.js      # 权限指令
├── assets/                # 静态资源
└── main.js                # 入口文件
```

### 功能模块划分

根据业务域和后端服务，可以划分为以下模块：

| 功能模块 | 页面组件 | 对应后端服务 | 业务职责 |
|---------|---------|-------------|---------|
| **认证模块** | Login.vue | auth-service | 用户登录、权限验证 |
| **仪表盘** | Dashboard.vue | monitor-service | 系统概览、实时监控 |
| **Agent 管理** | AgentList.vue, EventList.vue, CommandList.vue | agent-manager | Agent 注册、事件、命令 |
| **集群管理** | ClusterList.vue | cluster-service | K8s 集群管理、资源操作 |
| **监控管理** | AlertList.vue | monitor-service | 告警规则、监控配置 |
| **系统管理** | (待添加) UserList, RoleList, PermissionList | auth-service | 用户、角色、权限管理 |

---

## 🎯 微前端拆分目标

### 为什么要微前端？

1. **团队独立开发**
   - 不同团队负责不同子应用
   - 减少代码冲突
   - 提高开发效率

2. **技术栈独立演进**
   - 各子应用可选择不同版本
   - 可逐步升级技术栈
   - 降低技术债务

3. **独立部署发布**
   - 子应用独立构建部署
   - 支持灰度发布
   - 快速回滚
   - 降低发布风险

4. **按需加载**
   - 减小首屏加载体积
   - 提升用户体验
   - 优化性能

5. **业务隔离**
   - 模块边界清晰
   - 故障隔离
   - 便于维护

### 拆分原则

1. **业务域划分** - 按业务功能模块拆分
2. **服务对齐** - 与后端微服务对应
3. **独立性** - 每个子应用可独立运行
4. **复用性** - 公共组件和工具统一管理
5. **渐进式** - 支持逐步迁移

---

## 🏗️ 推荐架构方案

### 方案一: qiankun (推荐)

**技术栈**: qiankun + Vue 3

**优点**:
- ✅ 阿里开源，成熟稳定
- ✅ Vue 3 支持完善
- ✅ 文档详细，社区活跃
- ✅ 支持样式隔离
- ✅ 支持 JS 沙箱
- ✅ 学习成本低

**缺点**:
- ❌ 基于路由劫持，对路由依赖较强
- ❌ 子应用需要导出生命周期函数

**适用场景**:
- Vue 技术栈为主
- 需要样式和 JS 隔离
- 团队对微前端经验较少

### 方案二: Module Federation

**技术栈**: Webpack 5 Module Federation / Vite Module Federation

**优点**:
- ✅ Webpack 5 原生支持
- ✅ 真正的模块共享
- ✅ 运行时动态加载
- ✅ 性能优秀
- ✅ 灵活性强

**缺点**:
- ❌ Vite 支持需要插件
- ❌ 配置复杂
- ❌ 学习曲线陡峭

**适用场景**:
- 使用 Webpack 5
- 需要细粒度的模块共享
- 团队技术能力强

### 方案三: Micro App

**技术栈**: Micro App (京东)

**优点**:
- ✅ 基于 Web Component
- ✅ 零侵入
- ✅ 样式隔离好
- ✅ 性能优秀

**缺点**:
- ❌ 社区相对较小
- ❌ 生态不如 qiankun 完善

### 方案对比

| 特性 | qiankun | Module Federation | Micro App |
|------|---------|-------------------|-----------|
| 成熟度 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| Vue 3 支持 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| 学习成本 | 低 | 高 | 中 |
| 样式隔离 | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| JS 隔离 | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| 性能 | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| 生态 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |

**综合推荐**: **qiankun**

理由:
- Vue 3 项目，qiankun 支持完善
- 学习成本低，上手快
- 社区成熟，文档完善
- 满足业务需求

---

## 📦 拆分方案设计

### 应用架构

```
┌──────────────────────────────────────────────────────────┐
│                     Main App (主应用)                      │
│                   - 基座应用，端口 3000                     │
│                   - 路由调度、布局、登录                    │
├──────────────────────────────────────────────────────────┤
│  全局资源:                                                 │
│  - Vue 3、Ant Design Vue (共享)                          │
│  - 全局状态 (用户信息、权限)                               │
│  - 公共组件、工具函数                                      │
└───┬──────────┬──────────┬──────────┬──────────┬──────────┘
    │          │          │          │          │
    ↓          ↓          ↓          ↓          ↓
┌─────────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐
│Dashboard│ │Agent │ │Cluster│ │Monitor│ │System│
│  App    │ │ App  │ │  App  │ │  App  │ │  App │
│         │ │      │ │       │ │       │ │      │
│ :3001   │ │:3002 │ │ :3003 │ │ :3004 │ │:3005 │
└─────────┘ └──────┘ └───────┘ └───────┘ └──────┘
仪表盘      Agent    集群管理   监控管理   系统管理
            事件
            命令
```

### 具体划分

#### 1. 主应用 (main-app)

**端口**: 3000
**路由**: `/`

**职责**:
- 应用外壳和布局
- 路由管理和子应用注册
- 全局状态管理
- 登录认证
- 公共组件库

**包含内容**:
```
main-app/
├── src/
│   ├── layouts/
│   │   └── MainLayout.vue      # 主布局
│   ├── views/
│   │   └── Login.vue           # 登录页
│   ├── store/
│   │   └── user.js             # 用户状态
│   ├── api/
│   │   ├── request.js          # HTTP 封装
│   │   └── auth.js             # 认证 API
│   ├── micro-apps.js           # 子应用配置
│   ├── router/
│   │   └── index.js            # 主路由
│   └── main.js                 # qiankun 主应用入口
├── public/
└── package.json
```

**路由配置**:
```javascript
{
  path: '/login',
  path: '/dashboard',     // → dashboard-app
  path: '/agents',        // → agent-app
  path: '/clusters',      // → cluster-app
  path: '/monitor',       // → monitor-app
  path: '/system',        // → system-app
}
```

---

#### 2. 仪表盘应用 (dashboard-app)

**端口**: 3001
**路由**: `/dashboard`

**职责**:
- 系统概览
- 实时监控数据展示
- 关键指标展示

**包含内容**:
```
dashboard-app/
├── src/
│   ├── views/
│   │   └── Dashboard.vue
│   ├── api/
│   │   └── dashboard.js
│   ├── router/
│   │   └── index.js
│   ├── main.js            # qiankun 子应用入口
│   └── public-path.js     # 动态 publicPath
└── package.json
```

**对应后端**: monitor-service
**API**:
- `GET /api/v1/dashboard/overview`
- `GET /api/v1/metrics/summary`

---

#### 3. Agent 管理应用 (agent-app)

**端口**: 3002
**路由**: `/agents/*`

**职责**:
- Agent 列表管理
- 事件监控
- 命令执行

**包含内容**:
```
agent-app/
├── src/
│   ├── views/
│   │   ├── AgentList.vue
│   │   ├── EventList.vue
│   │   └── CommandList.vue
│   ├── api/
│   │   ├── agent.js
│   │   ├── event.js
│   │   └── command.js
│   ├── router/
│   │   └── index.js
│   └── main.js
└── package.json
```

**对应后端**: agent-manager
**路由**:
- `/agents/list` - Agent 列表
- `/agents/events` - 事件监控
- `/agents/commands` - 命令执行

**API**:
- `GET /api/v1/agents`
- `GET /api/v1/events`
- `POST /api/v1/commands`

---

#### 4. 集群管理应用 (cluster-app)

**端口**: 3003
**路由**: `/clusters/*`

**职责**:
- K8s 集群管理
- 集群资源查看
- Pod/Deployment 管理

**包含内容**:
```
cluster-app/
├── src/
│   ├── views/
│   │   ├── ClusterList.vue
│   │   ├── ClusterDetail.vue
│   │   ├── PodList.vue
│   │   └── DeploymentList.vue
│   ├── api/
│   │   └── cluster.js
│   ├── router/
│   │   └── index.js
│   └── main.js
└── package.json
```

**对应后端**: cluster-service
**路由**:
- `/clusters/list` - 集群列表
- `/clusters/:id` - 集群详情
- `/clusters/:id/pods` - Pod 列表
- `/clusters/:id/deployments` - Deployment 列表

**API**:
- `GET /api/v1/clusters`
- `GET /api/v1/clusters/:id/pods`
- `POST /api/v1/clusters/:id/deployments`

---

#### 5. 监控管理应用 (monitor-app)

**端口**: 3004
**路由**: `/monitor/*`

**职责**:
- 监控配置
- 告警规则管理
- 告警历史查看

**包含内容**:
```
monitor-app/
├── src/
│   ├── views/
│   │   ├── AlertList.vue
│   │   ├── AlertCreate.vue
│   │   ├── AlertHistory.vue
│   │   └── MetricsConfig.vue
│   ├── api/
│   │   └── alert.js
│   ├── router/
│   │   └── index.js
│   └── main.js
└── package.json
```

**对应后端**: monitor-service
**路由**:
- `/monitor/alerts` - 告警规则
- `/monitor/history` - 告警历史
- `/monitor/metrics` - 指标配置

**API**:
- `GET /api/v1/alerts`
- `POST /api/v1/alerts`
- `GET /api/v1/alert-history`

---

#### 6. 系统管理应用 (system-app)

**端口**: 3005
**路由**: `/system/*`

**职责**:
- 用户管理
- 角色管理
- 权限管理

**包含内容**:
```
system-app/
├── src/
│   ├── views/
│   │   ├── UserList.vue
│   │   ├── RoleList.vue
│   │   └── PermissionList.vue
│   ├── api/
│   │   ├── user.js
│   │   ├── role.js
│   │   └── permission.js
│   ├── router/
│   │   └── index.js
│   └── main.js
└── package.json
```

**对应后端**: auth-service
**路由**:
- `/system/users` - 用户管理
- `/system/roles` - 角色管理
- `/system/permissions` - 权限管理

**API**:
- `GET /api/v1/users`
- `GET /api/v1/roles`
- `GET /api/v1/permissions`

---

## 🔄 状态共享方案

### 全局状态

**在主应用维护**:
```javascript
// main-app/src/store/index.js
import { createPinia } from 'pinia'

const pinia = createPinia()

// 用户状态
export const useUserStore = defineStore('user', {
  state: () => ({
    token: '',
    userInfo: null,
    permissions: []
  })
})

// 通过 props 传递给子应用
```

### 通信方案

1. **Props 传递** (主应用 → 子应用)
```javascript
// 主应用注册时传递
{
  name: 'agent-app',
  props: {
    userInfo: state.userInfo,
    permissions: state.permissions
  }
}
```

2. **全局事件总线** (子应用 → 主应用)
```javascript
// 使用 qiankun 的 initGlobalState
import { initGlobalState } from 'qiankun'

const actions = initGlobalState({
  user: null,
  token: ''
})

// 监听变化
actions.onGlobalStateChange((state, prev) => {
  console.log('状态变化', state, prev)
})

// 修改状态
actions.setGlobalState({ token: 'xxx' })
```

3. **LocalStorage** (简单场景)
```javascript
// 适用于 token 等简单数据
localStorage.setItem('token', token)
```

---

## 🎨 样式隔离

### CSS 命名空间

每个子应用使用前缀:
```scss
// agent-app
.agent-app {
  &__header { }
  &__content { }
}

// cluster-app
.cluster-app {
  &__header { }
  &__content { }
}
```

### CSS Modules

```vue
<style module>
.container {
  /* 自动转换为唯一类名 */
}
</style>
```

### qiankun 沙箱

qiankun 默认提供样式隔离:
```javascript
registerMicroApps([
  {
    name: 'agent-app',
    sandbox: {
      strictStyleIsolation: true,  // 严格样式隔离
      experimentalStyleIsolation: true  // 实验性隔离
    }
  }
])
```

---

## 📊 性能优化

### 1. 公共依赖提取

**主应用配置**:
```javascript
// vite.config.js
export default {
  build: {
    rollupOptions: {
      external: ['vue', 'ant-design-vue', 'axios']
    }
  }
}
```

**子应用使用**:
```javascript
// 从 window 获取共享依赖
const Vue = window.Vue
const antd = window.antd
```

### 2. 预加载

```javascript
import { prefetchApps } from 'qiankun'

// 预加载子应用
prefetchApps([
  { name: 'agent-app', entry: '//localhost:3002' },
  { name: 'cluster-app', entry: '//localhost:3003' }
])
```

### 3. 按需加载

只在路由激活时加载子应用:
```javascript
// 延迟加载
loadMicroApp({
  name: 'agent-app',
  entry: '//localhost:3002',
  container: '#micro-app-container'
})
```

---

## 🚀 开发体验

### 本地开发

**启动所有应用**:
```bash
# 主应用
cd main-app && npm run dev       # :3000

# 子应用
cd dashboard-app && npm run dev  # :3001
cd agent-app && npm run dev      # :3002
cd cluster-app && npm run dev    # :3003
cd monitor-app && npm run dev    # :3004
cd system-app && npm run dev     # :3005
```

**独立开发**:
每个子应用可以独立运行和调试:
```bash
cd agent-app
npm run dev:standalone  # 独立模式
```

### 调试

1. **主应用调试** - 浏览器开发工具
2. **子应用调试** - Vue DevTools
3. **通信调试** - qiankun DevTools

---

## 📈 迁移策略

### 渐进式迁移

**阶段1**: 搭建基础架构
- ✅ 创建主应用
- ✅ 集成 qiankun
- ✅ 配置路由

**阶段2**: 迁移第一个子应用
- ✅ 选择 dashboard-app 作为试点
- ✅ 验证架构可行性
- ✅ 建立最佳实践

**阶段3**: 逐步迁移其他模块
- ✅ agent-app
- ✅ cluster-app
- ✅ monitor-app
- ✅ system-app

**阶段4**: 优化和完善
- ✅ 性能优化
- ✅ 监控告警
- ✅ 文档完善

### 兼容策略

保留单体应用作为备份,直到微前端架构稳定。

---

## 📝 下一步

查看详细的实施指南:
- [技术方案对比](MICRO_FRONTEND_TECH_COMPARISON.md)
- [qiankun 实施指南](MICRO_FRONTEND_QIANKUN_GUIDE.md)
- [示例代码](MICRO_FRONTEND_EXAMPLES.md)

---

## 🤝 总结

通过微前端架构:
- ✅ 实现团队独立开发
- ✅ 支持技术栈独立演进
- ✅ 提升部署效率
- ✅ 优化用户体验
- ✅ 降低维护成本

**推荐使用 qiankun** 进行微前端改造，分 4 个阶段渐进式迁移。
