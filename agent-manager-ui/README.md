# Agent Manager UI

Aetherius Agent Manager 的 Web 管理界面。

## 技术栈

- **Vue 3** - 渐进式 JavaScript 框架
- **Ant Design Vue 4** - 企业级 UI 组件库
- **VXETable** - 强大的表格组件
- **Vite** - 下一代前端构建工具
- **Pinia** - Vue 状态管理
- **Axios** - HTTP 客户端

## 功能特性

### 1. 仪表盘 (Dashboard)
- 实时显示系统关键指标
- Agent 在线/离线统计
- 今日事件数量
- 执行中命令数量
- 最近事件列表
- Agent 状态列表
- 自动刷新（30秒）

### 2. Agent 管理
- Agent 列表展示（使用 VXETable）
- 支持搜索和过滤
- 查看 Agent 详细信息
- 删除 Agent
- 自动刷新（30秒）
- 分页支持

### 3. 事件监控
- 实时事件流展示
- 按严重程度过滤（Critical, Warning, Info, Normal）
- 按事件类型过滤
- 资源名称搜索
- 事件详情查看
- 自动刷新（10秒）
- 分页支持

### 4. 命令执行
- 命令列表管理
- 创建新命令
- 执行待执行命令
- 命令状态跟踪（待执行、执行中、已完成、失败、超时）
- 查看命令详情和执行结果
- 自动刷新（5秒）
- 分页支持

### 5. 集群管理
- 集群列表展示
- 添加/编辑集群
- 删除集群
- 集群详细信息
- 集群状态监控
- 分页支持

### 6. 告警规则
- 告警规则列表
- 创建/编辑告警规则
- 启用/禁用规则
- 删除规则
- 支持多种通知渠道（邮件、Webhook、Slack）
- 自定义告警条件
- 分页支持

## 快速开始

### 前置要求

- Node.js >= 16
- npm 或 yarn
- agent-manager 后端服务运行在 `localhost:8080`

### 安装依赖

```bash
npm install
```

### 开发模式

```bash
npm run dev
```

访问 http://localhost:3000

### 生产构建

```bash
npm run build
```

构建产物将生成在 `dist/` 目录。

### 预览生产构建

```bash
npm run preview
```

## 项目结构

```
agent-manager-ui/
├── public/              # 静态资源
├── src/
│   ├── api/            # API 接口封装
│   │   ├── request.js  # Axios 实例配置
│   │   ├── agent.js    # Agent 相关接口
│   │   ├── event.js    # Event 相关接口
│   │   ├── command.js  # Command 相关接口
│   │   ├── cluster.js  # Cluster 相关接口
│   │   └── alert.js    # Alert 相关接口
│   ├── assets/         # 资源文件
│   │   └── styles/     # 样式文件
│   ├── layouts/        # 布局组件
│   │   └── MainLayout.vue
│   ├── router/         # 路由配置
│   │   └── index.js
│   ├── views/          # 页面组件
│   │   ├── Dashboard.vue
│   │   ├── agents/
│   │   │   └── AgentList.vue
│   │   ├── events/
│   │   │   └── EventList.vue
│   │   ├── commands/
│   │   │   └── CommandList.vue
│   │   ├── clusters/
│   │   │   └── ClusterList.vue
│   │   └── alerts/
│   │       └── AlertList.vue
│   ├── App.vue         # 根组件
│   └── main.js         # 入口文件
├── index.html
├── vite.config.js      # Vite 配置
└── package.json
```

## API 代理配置

开发环境下，所有 `/api` 请求会被代理到 `http://localhost:8080`。

如需修改后端地址，编辑 `vite.config.js`:

```javascript
server: {
  port: 3000,
  proxy: {
    '/api': {
      target: 'http://your-backend-url',
      changeOrigin: true
    }
  }
}
```

## 环境变量

可以创建 `.env.development` 和 `.env.production` 文件来配置不同环境的变量。

## 主要特性

### 响应式设计
- 自适应布局
- 支持侧边栏折叠
- 移动端友好

### 实时更新
- Dashboard: 30秒自动刷新
- Agent 列表: 30秒自动刷新
- Event 列表: 10秒自动刷新
- Command 列表: 5秒自动刷新

### 数据展示
- 使用 VXETable 提供强大的表格功能
- 支持排序、筛选
- 固定列
- 分页

### 用户体验
- 面包屑导航
- 路由过渡动画
- 加载状态提示
- 操作反馈提示

## 开发指南

### 添加新页面

1. 在 `src/views/` 创建页面组件
2. 在 `src/router/index.js` 添加路由
3. 在 `src/layouts/MainLayout.vue` 添加菜单项

### 添加新 API

1. 在 `src/api/` 创建对应的 API 文件
2. 使用统一的 request 实例
3. 导出 API 函数

## 浏览器兼容性

- Chrome >= 90
- Firefox >= 88
- Safari >= 14
- Edge >= 90

## License

MIT
