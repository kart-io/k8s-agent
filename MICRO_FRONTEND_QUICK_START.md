# 微前端快速开始指南

> 30分钟将 agent-manager-ui 改造为微前端架构

## 🎯 目标

将现有的单体前端应用拆分成:
- 1 个主应用 (基座)
- 5 个子应用 (业务模块)

## 📋 准备工作

### 1. 备份现有代码

```bash
# 备份当前应用
cp -r agent-manager-ui agent-manager-ui-backup

# 创建微前端项目目录
mkdir micro-frontend-apps
cd micro-frontend-apps
```

### 2. 规划目录结构

```
micro-frontend-apps/
├── main-app/          # 主应用 (端口 3000)
├── dashboard-app/     # 仪表盘 (端口 3001)
├── agent-app/         # Agent管理 (端口 3002)
├── cluster-app/       # 集群管理 (端口 3003)
├── monitor-app/       # 监控管理 (端口 3004)
└── system-app/        # 系统管理 (端口 3005)
```

---

## 🚀 第一步: 创建主应用

### 1. 创建项目

```bash
npm create vite@latest main-app -- --template vue
cd main-app
npm install
```

### 2. 安装依赖

```bash
npm install qiankun
npm install vue-router pinia
npm install ant-design-vue
npm install axios
```

### 3. 创建目录结构

```bash
cd src
mkdir -p micro layouts store api
```

### 4. 复制现有文件

```bash
# 从旧项目复制
cp ../../agent-manager-ui-backup/src/layouts/MainLayout.vue layouts/
cp ../../agent-manager-ui-backup/src/views/Login.vue views/
cp ../../agent-manager-ui-backup/src/store/user.js store/
cp ../../agent-manager-ui-backup/src/api/request.js api/
cp ../../agent-manager-ui-backup/src/api/auth.js api/
cp ../../agent-manager-ui-backup/src/directives/permission.js directives/
```

### 5. 创建子应用配置

创建 `src/micro/apps.js`:

```javascript
export default [
  {
    name: 'dashboard-app',
    entry: '//localhost:3001',
    container: '#micro-app-container',
    activeRule: '/dashboard'
  },
  {
    name: 'agent-app',
    entry: '//localhost:3002',
    container: '#micro-app-container',
    activeRule: '/agents'
  },
  {
    name: 'cluster-app',
    entry: '//localhost:3003',
    container: '#micro-app-container',
    activeRule: '/clusters'
  },
  {
    name: 'monitor-app',
    entry: '//localhost:3004',
    container: '#micro-app-container',
    activeRule: '/monitor'
  },
  {
    name: 'system-app',
    entry: '//localhost:3005',
    container: '#micro-app-container',
    activeRule: '/system'
  }
]
```

### 6. 创建 qiankun 配置

创建 `src/micro/index.js`:

```javascript
import { registerMicroApps, start } from 'qiankun'
import apps from './apps'

export function registerApps() {
  registerMicroApps(apps)
}

export function startQiankun() {
  start({
    prefetch: true,
    sandbox: { experimentalStyleIsolation: true }
  })
}
```

### 7. 修改 main.js

```javascript
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Antd from 'ant-design-vue'
import 'ant-design-vue/dist/reset.css'

import App from './App.vue'
import router from './router'
import { registerApps, startQiankun } from './micro'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(Antd)
app.mount('#app')

// 启动微前端
registerApps()
startQiankun()
```

### 8. 配置路由

`src/router/index.js`:

```javascript
import { createRouter, createWebHistory } from 'vue-router'

export default createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      component: () => import('@/views/Login.vue')
    },
    {
      path: '/',
      component: () => import('@/layouts/MainLayout.vue'),
      redirect: '/dashboard'
    }
  ]
})
```

### 9. 修改布局文件

`src/layouts/MainLayout.vue` - 添加子应用容器:

```vue
<template>
  <a-layout class="main-layout">
    <a-layout-header>...</a-layout-header>
    <a-layout>
      <a-layout-sider>...</a-layout-sider>
      <a-layout-content>
        <!-- 子应用容器 -->
        <div id="micro-app-container"></div>
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>
```

### 10. 启动主应用

```bash
npm run dev
# 访问 http://localhost:3000
```

✅ **主应用完成!**

---

## 🎨 第二步: 创建第一个子应用 (Dashboard)

### 1. 创建项目

```bash
cd ..
npm create vite@latest dashboard-app -- --template vue
cd dashboard-app
npm install
```

### 2. 安装依赖

```bash
npm install vite-plugin-qiankun -D
npm install vue-router
npm install ant-design-vue
npm install axios
```

### 3. 复制文件

```bash
# 从旧项目复制 Dashboard 相关文件
cp ../../agent-manager-ui-backup/src/views/Dashboard.vue src/views/
```

### 4. 创建 public-path.js

```javascript
// src/public-path.js
if (window.__POWERED_BY_QIANKUN__) {
  __webpack_public_path__ = window.__INJECTED_PUBLIC_PATH_BY_QIANKUN__
}
```

### 5. 修改 main.js

```javascript
import './public-path'
import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import Antd from 'ant-design-vue'
import App from './App.vue'

let app = null
let router = null

function render(props = {}) {
  const { container } = props

  router = createRouter({
    history: createWebHistory(window.__POWERED_BY_QIANKUN__ ? '/dashboard' : '/'),
    routes: [
      {
        path: '/',
        component: () => import('./views/Dashboard.vue')
      }
    ]
  })

  app = createApp(App)
  app.use(router)
  app.use(Antd)

  const el = container ? container.querySelector('#app') : '#app'
  app.mount(el)
}

if (!window.__POWERED_BY_QIANKUN__) {
  render()
}

export async function bootstrap() {}
export async function mount(props) {
  render(props)
}
export async function unmount() {
  app?.unmount()
}
```

### 6. 配置 vite.config.js

```javascript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import qiankun from 'vite-plugin-qiankun'

export default defineConfig({
  plugins: [
    vue(),
    qiankun('dashboard-app', { useDevMode: true })
  ],
  server: {
    port: 3001,
    cors: true,
    headers: {
      'Access-Control-Allow-Origin': '*'
    }
  }
})
```

### 7. 启动子应用

```bash
npm run dev
# 访问 http://localhost:3001 (独立运行)
```

### 8. 在主应用中查看

访问主应用 `http://localhost:3000/dashboard`

✅ **第一个子应用完成!**

---

## 📦 第三步: 批量创建其他子应用

使用相同的方式创建其他子应用:

### 1. Agent 应用 (端口 3002)

```bash
cd ..
npm create vite@latest agent-app -- --template vue
cd agent-app

# 复制文件
cp ../../agent-manager-ui-backup/src/views/agents/* src/views/
cp ../../agent-manager-ui-backup/src/views/events/* src/views/
cp ../../agent-manager-ui-backup/src/views/commands/* src/views/
cp ../../agent-manager-ui-backup/src/api/agent.js src/api/
cp ../../agent-manager-ui-backup/src/api/event.js src/api/
cp ../../agent-manager-ui-backup/src/api/command.js src/api/

# 安装依赖和配置 (同 dashboard-app)
npm install
npm install vite-plugin-qiankun -D vue-router ant-design-vue axios
```

修改路由:

```javascript
// src/router/index.js
export default [
  {
    path: '/list',
    component: () => import('@/views/AgentList.vue')
  },
  {
    path: '/events',
    component: () => import('@/views/EventList.vue')
  },
  {
    path: '/commands',
    component: () => import('@/views/CommandList.vue')
  }
]
```

### 2. Cluster 应用 (端口 3003)

```bash
cd ..
npm create vite@latest cluster-app -- --template vue
cd cluster-app

# 复制文件
cp ../../agent-manager-ui-backup/src/views/clusters/* src/views/
cp ../../agent-manager-ui-backup/src/api/cluster.js src/api/

# 配置同上
```

### 3. Monitor 应用 (端口 3004)

```bash
cd ..
npm create vite@latest monitor-app -- --template vue
cd monitor-app

# 复制文件
cp ../../agent-manager-ui-backup/src/views/alerts/* src/views/
cp ../../agent-manager-ui-backup/src/api/alert.js src/api/

# 配置同上
```

### 4. System 应用 (端口 3005)

```bash
cd ..
npm create vite@latest system-app -- --template vue
cd system-app

# 新建用户管理等页面
# 复制 API
cp ../../agent-manager-ui-backup/src/api/user.js src/api/

# 配置同上
```

---

## 🎬 第四步: 启动所有应用

### 1. 创建统一启动脚本

在项目根目录创建 `start-all.sh`:

```bash
#!/bin/bash

# 启动主应用
cd main-app
npm run dev &

# 启动子应用
cd ../dashboard-app
npm run dev &

cd ../agent-app
npm run dev &

cd ../cluster-app
npm run dev &

cd ../monitor-app
npm run dev &

cd ../system-app
npm run dev &

echo "所有应用已启动"
echo "主应用: http://localhost:3000"
```

### 2. 使用 concurrently (推荐)

安装:

```bash
npm install -g concurrently
```

创建 `package.json`:

```json
{
  "scripts": {
    "dev:all": "concurrently \"npm:dev:*\"",
    "dev:main": "cd main-app && npm run dev",
    "dev:dashboard": "cd dashboard-app && npm run dev",
    "dev:agent": "cd agent-app && npm run dev",
    "dev:cluster": "cd cluster-app && npm run dev",
    "dev:monitor": "cd monitor-app && npm run dev",
    "dev:system": "cd system-app && npm run dev"
  }
}
```

启动:

```bash
npm run dev:all
```

---

## ✅ 验证

### 1. 访问主应用

打开浏览器: `http://localhost:3000`

### 2. 测试各个模块

- `/dashboard` - 仪表盘
- `/agents/list` - Agent 列表
- `/agents/events` - 事件监控
- `/agents/commands` - 命令执行
- `/clusters` - 集群管理
- `/monitor/alerts` - 告警规则
- `/system/users` - 用户管理

### 3. 检查子应用加载

打开浏览器控制台，应该看到:

```
[主应用] 注册子应用
[dashboard-app] bootstrap
[dashboard-app] mount
```

---

## 🐛 常见问题

### 1. 子应用加载失败

**问题**: 控制台报错 CORS

**解决**:
```javascript
// 子应用 vite.config.js
server: {
  cors: true,
  headers: {
    'Access-Control-Allow-Origin': '*'
  }
}
```

### 2. 样式冲突

**解决**: 使用 CSS 命名空间

```vue
<style scoped>
.dashboard-app {
  /* 所有样式都在这个命名空间下 */
}
</style>
```

### 3. 路由不生效

**检查**:
- 主应用路由是否正确
- 子应用 base 是否配置
- activeRule 是否匹配

---

## 📊 性能优化

### 1. 预加载

```javascript
// 主应用
import { prefetchApps } from 'qiankun'

prefetchApps([
  { name: 'dashboard-app', entry: '//localhost:3001' },
  { name: 'agent-app', entry: '//localhost:3002' }
])
```

### 2. 共享依赖

```javascript
// 主应用 index.html
<script src="https://cdn.jsdelivr.net/npm/vue@3"></script>
<script src="https://cdn.jsdelivr.net/npm/ant-design-vue@4"></script>

// 子应用 vite.config.js
build: {
  rollupOptions: {
    external: ['vue', 'ant-design-vue']
  }
}
```

---

## 🎉 完成!

你已经成功将单体前端应用改造为微前端架构!

### 下一步

1. **优化**: 调整样式、优化性能
2. **测试**: 测试各个功能模块
3. **部署**: 配置生产环境部署
4. **文档**: 编写团队开发文档

### 参考文档

- [完整架构分析](MICRO_FRONTEND_ANALYSIS.md)
- [qiankun 实施指南](MICRO_FRONTEND_QIANKUN_GUIDE.md)

---

## 📞 需要帮助?

遇到问题可以:
1. 查看 qiankun 官方文档
2. 检查控制台错误信息
3. 对比示例代码

祝你成功! 🚀
