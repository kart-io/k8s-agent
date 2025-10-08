# Web 微前端项目总结

## 项目概述

已成功创建基于 qiankun 的微前端项目，包含 1 个主应用和 5 个子应用。

## 项目结构

```
web/
├── main-app/          # 主应用（基座）- 端口 3000
│   ├── src/
│   │   ├── api/          # API 请求层
│   │   ├── assets/       # 静态资源
│   │   ├── directives/   # 自定义指令（权限控制）
│   │   ├── layouts/      # 布局组件
│   │   ├── micro/        # 微前端配置
│   │   ├── router/       # 路由配置
│   │   ├── store/        # 状态管理（Pinia）
│   │   ├── views/        # 页面组件
│   │   ├── App.vue
│   │   └── main.js
│   ├── vite.config.js
│   └── package.json
│
├── dashboard-app/     # 仪表盘子应用 - 端口 3001
│   ├── src/
│   │   ├── views/
│   │   │   └── Dashboard.vue    # 数据统计、图表展示
│   │   ├── router/
│   │   ├── App.vue
│   │   └── main.js
│   └── package.json
│
├── agent-app/         # Agent 管理子应用 - 端口 3002
│   ├── src/
│   │   ├── api/
│   │   │   └── agent.js         # Agent API
│   │   ├── views/
│   │   │   ├── AgentList.vue    # Agent 列表
│   │   │   ├── EventList.vue    # 事件列表
│   │   │   └── CommandList.vue  # 命令管理
│   │   └── router/
│   └── package.json
│
├── cluster-app/       # 集群管理子应用 - 端口 3003
│   ├── src/
│   │   ├── api/
│   │   │   └── cluster.js       # 集群 API
│   │   ├── views/
│   │   │   ├── ClusterList.vue  # 集群列表
│   │   │   └── ClusterDetail.vue # 集群详情（Pods/Services/Deployments）
│   │   └── router/
│   └── package.json
│
├── monitor-app/       # 监控管理子应用 - 端口 3004
│   ├── src/
│   │   ├── api/
│   │   │   └── monitor.js       # 监控 API
│   │   ├── views/
│   │   │   ├── MetricsList.vue  # 监控指标（CPU/内存/磁盘/网络）
│   │   │   └── AlertsList.vue   # 告警规则管理
│   │   └── router/
│   └── package.json
│
├── system-app/        # 系统管理子应用 - 端口 3005
│   ├── src/
│   │   ├── api/
│   │   │   └── system.js        # 系统 API
│   │   ├── views/
│   │   │   ├── UserList.vue     # 用户管理
│   │   │   ├── RoleList.vue     # 角色管理
│   │   │   └── PermissionList.vue # 权限管理
│   │   └── router/
│   └── package.json
│
├── package.json       # 根 package.json（统一脚本）
├── setup.sh           # 一键安装脚本
├── QUICK_START.md     # 快速启动指南
└── .gitignore
```

## 技术栈

### 核心技术
- **微前端框架**: qiankun 2.x
- **前端框架**: Vue 3.3 + Composition API
- **构建工具**: Vite 4
- **包管理器**: pnpm 8

### UI & 组件
- **UI 组件库**: Ant Design Vue 4
- **图表库**: ECharts 5 + vue-echarts
- **图标**: @ant-design/icons-vue

### 状态 & 路由
- **状态管理**: Pinia 2.1
- **路由**: Vue Router 4
- **HTTP 请求**: Axios 1.5

### 工具库
- **日期处理**: Day.js
- **YAML 解析**: js-yaml (cluster-app)
- **样式**: SCSS

## 主要功能

### 主应用 (main-app)
- ✅ qiankun 微前端集成
- ✅ 统一登录认证（JWT）
- ✅ 路由守卫（自动跳转登录页）
- ✅ 全局状态管理（initGlobalState）
- ✅ 权限指令（v-permission, v-role）
- ✅ 统一布局（顶部导航 + 侧边菜单）
- ✅ 子应用容器（#micro-app-container）
- ✅ API 代理配置（proxy to port 8080）

### Dashboard App
- ✅ Agent 状态统计（总数/在线/离线/待处理事件）
- ✅ Agent 状态分布饼图
- ✅ 事件类型统计柱状图
- ✅ 事件趋势折线图
- ✅ 最近事件列表

### Agent App
- ✅ Agent 列表管理（查看/编辑/删除）
- ✅ Agent 详情查看
- ✅ 事件列表（筛选/详情）
- ✅ 命令发送和管理
- ✅ 命令状态跟踪

### Cluster App
- ✅ 集群列表管理（增删改查）
- ✅ KubeConfig 配置
- ✅ 集群详情页
- ✅ Pod 列表查看
- ✅ Service 列表查看
- ✅ Deployment 列表查看
- ✅ 多 Tab 切换

### Monitor App
- ✅ 系统指标概览（CPU/内存/磁盘/网络）
- ✅ 指标历史趋势图表
- ✅ 实时数据自动刷新（30秒）
- ✅ 时间范围选择（1h/6h/24h/7d）
- ✅ 告警规则管理（CRUD）
- ✅ 告警级别配置（严重/警告/信息）
- ✅ 多种通知方式（邮件/短信/Webhook）

### System App
- ✅ 用户管理（CRUD）
- ✅ 角色管理（CRUD）
- ✅ 权限管理（CRUD）
- ✅ 用户角色分配
- ✅ 角色权限分配
- ✅ 用户状态管理（正常/禁用）

## 路由配置

| 路由 | 应用 | 端口 | 说明 |
|------|------|------|------|
| / | main-app | 3000 | 首页（重定向到 /dashboard） |
| /login | main-app | 3000 | 登录页 |
| /dashboard | dashboard-app | 3001 | 仪表盘 |
| /agents/list | agent-app | 3002 | Agent 列表 |
| /agents/events | agent-app | 3002 | 事件列表 |
| /agents/commands | agent-app | 3002 | 命令管理 |
| /clusters/list | cluster-app | 3003 | 集群列表 |
| /clusters/:id | cluster-app | 3003 | 集群详情 |
| /monitor/metrics | monitor-app | 3004 | 监控指标 |
| /monitor/alerts | monitor-app | 3004 | 告警规则 |
| /system/users | system-app | 3005 | 用户管理 |
| /system/roles | system-app | 3005 | 角色管理 |
| /system/permissions | system-app | 3005 | 权限管理 |

## 微前端特性

### 1. 应用隔离
- **JS 隔离**: qiankun 自动实现 JS 沙箱
- **样式隔离**: experimentalStyleIsolation 开启
- **路由隔离**: 每个子应用独立的路由配置

### 2. 状态共享
```javascript
// 主应用
const actions = initGlobalState({
  user: null,
  token: ''
})

// 子应用接收
props.onGlobalStateChange((state, prev) => {
  console.log('state changed', state)
})
```

### 3. 通信机制
- **Props 传递**: 主应用向子应用传递数据
- **全局状态**: initGlobalState 实现应用间通信
- **自定义事件**: window.postMessage

### 4. 生命周期
- **bootstrap**: 应用首次加载时调用
- **mount**: 应用挂载时调用
- **unmount**: 应用卸载时调用
- **update**: 应用更新时调用

## 快速启动

### 1. 安装依赖
```bash
cd web/
./setup.sh
# 或者
pnpm run install:all
```

### 2. 启动所有应用
```bash
pnpm dev
```

### 3. 单独启动某个应用
```bash
pnpm dev:main       # 主应用
pnpm dev:dashboard  # Dashboard
pnpm dev:agent      # Agent
pnpm dev:cluster    # Cluster
pnpm dev:monitor    # Monitor
pnpm dev:system     # System
```

### 4. 构建生产版本
```bash
pnpm build          # 构建所有
pnpm build:main     # 构建主应用
# ...
```

## 开发指南

### 添加新的子应用

1. 创建新应用目录和基础结构
2. 安装 `vite-plugin-qiankun`
3. 配置 `vite.config.js`
4. 创建 `src/public-path.js`
5. 修改 `src/main.js` 添加 qiankun 生命周期
6. 在主应用的 `src/micro/apps.js` 注册
7. 在主应用的侧边菜单添加入口

### 添加全局状态

```javascript
// main-app/src/micro/index.js
const actions = initGlobalState({
  newState: 'value'
})

// 子应用使用
export async function mount(props) {
  props.onGlobalStateChange((state) => {
    console.log(state.newState)
  })
}
```

### API 请求配置

所有子应用的 API 请求都通过主应用的 proxy 代理到 Gateway（port 8080）：

```javascript
// vite.config.js
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true
    }
  }
}
```

## 与后端集成

### API Gateway
- 端口: 8080
- 提供统一的 API 入口
- JWT 认证
- 请求路由到各个微服务

### 后端服务
- auth-service (8090): 认证授权
- agent-manager (8000): Agent 管理
- monitor-service (8081): 监控管理
- cluster-service (8082): 集群管理

### 认证流程
1. 用户在 `/login` 登录
2. 调用 `/api/auth/login` 获取 JWT token
3. Token 存储到 localStorage
4. 后续请求通过 Axios 拦截器自动添加 token
5. 路由守卫检查登录状态

## 权限控制

### 路由级别
```javascript
// router/index.js
{
  path: '/admin',
  meta: { requiresAuth: true, roles: ['admin'] }
}
```

### 组件级别
```vue
<template>
  <a-button v-permission="'agent:write'">删除</a-button>
  <div v-role="'admin'">管理员专属内容</div>
</template>
```

### API 级别
后端通过 JWT 中的 roles 和 permissions 进行权限验证

## 最佳实践

1. **独立开发**: 每个子应用可独立开发和测试
2. **样式隔离**: 避免全局样式污染
3. **状态管理**: 优先使用子应用内部状态，必要时使用全局状态
4. **懒加载**: 子应用按需加载，提高首屏性能
5. **错误处理**: 使用 qiankun 的错误处理机制
6. **版本管理**: 主应用和子应用独立版本控制

## 已实现功能清单

- ✅ 项目结构搭建
- ✅ 主应用创建（qiankun 集成）
- ✅ 5 个子应用创建
- ✅ 统一登录认证
- ✅ 路由配置
- ✅ 权限控制
- ✅ 全局状态管理
- ✅ API 代理配置
- ✅ 统一构建脚本
- ✅ 快速启动指南
- ✅ 一键安装脚本

## 待优化项

1. **单元测试**: 添加 Vitest 单元测试
2. **E2E 测试**: 添加 Playwright E2E 测试
3. **CI/CD**: 配置 GitHub Actions
4. **Docker 支持**: 添加 Dockerfile 和 docker-compose
5. **性能优化**: 代码分割、预加载优化
6. **错误监控**: 集成 Sentry 等错误监控工具
7. **国际化**: 添加 i18n 支持
8. **主题切换**: 支持亮色/暗色主题

## 相关文档

- [快速启动指南](web/QUICK_START.md)
- [微前端架构分析](MICRO_FRONTEND_ANALYSIS.md)
- [qiankun 实施指南](MICRO_FRONTEND_QIANKUN_GUIDE.md)
- [后端服务架构](BACKEND_MANAGEMENT_ARCHITECTURE.md)
- [后端快速启动](QUICK_START_BACKEND.md)

## 总结

Web 微前端项目已完全搭建完成，包含：
- 1 个主应用（基座）
- 5 个功能完整的子应用
- 完整的认证和权限系统
- 统一的开发和构建流程
- 详细的文档和快速启动指南

所有应用均可独立开发和运行，也可通过主应用进行集成。项目采用现代化技术栈，遵循最佳实践，具有良好的可扩展性和维护性。
