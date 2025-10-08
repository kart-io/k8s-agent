# qiankun 微前端实施指南

> 基于 qiankun 2.x + Vue 3 的微前端架构实施详细指南

## 📋 目录

- [快速开始](#快速开始)
- [主应用改造](#主应用改造)
- [子应用改造](#子应用改造)
- [状态共享](#状态共享)
- [样式隔离](#样式隔离)
- [部署配置](#部署配置)
- [最佳实践](#最佳实践)

---

## 🚀 快速开始

### 前置要求

- Node.js 16+
- Vue 3.x
- Vite 4.x

### 整体架构

```
micro-frontend/
├── main-app/          # 主应用 (基座)
├── dashboard-app/     # 仪表盘子应用
├── agent-app/         # Agent 管理子应用
├── cluster-app/       # 集群管理子应用
├── monitor-app/       # 监控管理子应用
├── system-app/        # 系统管理子应用
└── shared/            # 共享资源 (可选)
    ├── components/    # 公共组件
    └── utils/         # 工具函数
```

---

## 🏗️ 主应用改造

### 1. 创建主应用

```bash
npm create vite@latest main-app -- --template vue
cd main-app
npm install
```

### 2. 安装 qiankun

```bash
npm install qiankun
```

### 3. 主应用目录结构

```
main-app/
├── public/
├── src/
│   ├── assets/
│   ├── layouts/
│   │   └── MainLayout.vue       # 主布局
│   ├── views/
│   │   └── Login.vue            # 登录页
│   ├── store/
│   │   └── user.js              # 用户状态
│   ├── api/
│   │   ├── request.js
│   │   └── auth.js
│   ├── micro/
│   │   ├── apps.js              # 子应用配置
│   │   └── index.js             # qiankun 配置
│   ├── router/
│   │   └── index.js
│   ├── App.vue
│   └── main.js
├── index.html
├── vite.config.js
└── package.json
```

### 4. 配置子应用列表

创建 `src/micro/apps.js`:

```javascript
/**
 * 子应用配置
 */
const apps = [
  {
    name: 'dashboard-app',
    entry: import.meta.env.DEV
      ? '//localhost:3001'
      : '/dashboard',
    container: '#micro-app-container',
    activeRule: '/dashboard',
    props: {
      // 传递给子应用的数据
    }
  },
  {
    name: 'agent-app',
    entry: import.meta.env.DEV
      ? '//localhost:3002'
      : '/agent',
    container: '#micro-app-container',
    activeRule: '/agents',
  },
  {
    name: 'cluster-app',
    entry: import.meta.env.DEV
      ? '//localhost:3003'
      : '/cluster',
    container: '#micro-app-container',
    activeRule: '/clusters',
  },
  {
    name: 'monitor-app',
    entry: import.meta.env.DEV
      ? '//localhost:3004'
      : '/monitor',
    container: '#micro-app-container',
    activeRule: '/monitor',
  },
  {
    name: 'system-app',
    entry: import.meta.env.DEV
      ? '//localhost:3005'
      : '/system',
    container: '#micro-app-container',
    activeRule: '/system',
  }
]

export default apps
```

### 5. 初始化 qiankun

创建 `src/micro/index.js`:

```javascript
import { registerMicroApps, start, initGlobalState, addGlobalUncaughtErrorHandler } from 'qiankun'
import apps from './apps'
import { useUserStore } from '@/store/user'

// 初始化全局状态
const initialState = {
  user: null,
  token: ''
}

const actions = initGlobalState(initialState)

// 监听全局状态变化
actions.onGlobalStateChange((state, prev) => {
  console.log('[主应用] 状态变化', state, prev)
})

/**
 * 注册子应用
 */
export function registerApps() {
  const userStore = useUserStore()

  registerMicroApps(
    apps.map(app => ({
      ...app,
      props: {
        // 传递用户信息和 actions 给子应用
        userInfo: userStore.userInfo,
        token: userStore.token,
        actions
      },
      loader: (loading) => {
        console.log(`[主应用] 子应用 ${app.name} loading: ${loading}`)
      }
    })),
    {
      beforeLoad: [
        app => {
          console.log('[主应用] before load', app.name)
        }
      ],
      beforeMount: [
        app => {
          console.log('[主应用] before mount', app.name)
        }
      ],
      afterMount: [
        app => {
          console.log('[主应用] after mount', app.name)
        }
      ],
      afterUnmount: [
        app => {
          console.log('[主应用] after unmount', app.name)
        }
      ]
    }
  )
}

/**
 * 启动 qiankun
 */
export function startQiankun() {
  start({
    prefetch: true,  // 预加载
    sandbox: {
      strictStyleIsolation: false,  // 严格样式隔离
      experimentalStyleIsolation: true  // 实验性样式隔离
    },
    singular: false  // 是否单实例
  })
}

/**
 * 全局错误处理
 */
addGlobalUncaughtErrorHandler((event) => {
  console.error('[主应用] 全局错误', event)
  const { message } = event
  // 可以上报错误
})

export { actions }
```

### 6. 修改主应用入口

修改 `src/main.js`:

```javascript
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Antd from 'ant-design-vue'
import 'ant-design-vue/dist/reset.css'

import App from './App.vue'
import router from './router'
import { registerApps, startQiankun } from './micro'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(Antd)

app.mount('#app')

// 注册并启动微前端
registerApps()
startQiankun()
```

### 7. 主应用布局

修改 `src/layouts/MainLayout.vue`:

```vue
<template>
  <a-layout class="main-layout">
    <!-- 顶部导航 -->
    <a-layout-header class="header">
      <div class="logo">K8s Agent 管理平台</div>
      <div class="user-info">
        <a-dropdown>
          <a-avatar :src="userStore.userInfo?.avatar" />
          <template #overlay>
            <a-menu>
              <a-menu-item @click="handleLogout">退出登录</a-menu-item>
            </a-menu>
          </template>
        </a-dropdown>
      </div>
    </a-layout-header>

    <a-layout>
      <!-- 侧边栏 -->
      <a-layout-sider v-model:collapsed="collapsed" collapsible>
        <a-menu
          v-model:selectedKeys="selectedKeys"
          mode="inline"
          theme="dark"
          @click="handleMenuClick"
        >
          <a-menu-item key="/dashboard">
            <DashboardOutlined />
            <span>仪表盘</span>
          </a-menu-item>

          <a-menu-item key="/agents">
            <ClusterOutlined />
            <span>Agent 管理</span>
          </a-menu-item>

          <a-menu-item key="/clusters">
            <DeploymentUnitOutlined />
            <span>集群管理</span>
          </a-menu-item>

          <a-menu-item key="/monitor">
            <MonitorOutlined />
            <span>监控管理</span>
          </a-menu-item>

          <a-menu-item key="/system">
            <SettingOutlined />
            <span>系统管理</span>
          </a-menu-item>
        </a-menu>
      </a-layout-sider>

      <!-- 内容区域 - 子应用挂载点 -->
      <a-layout-content class="content">
        <!-- 主应用路由 -->
        <router-view v-if="isMainRoute" />

        <!-- 子应用容器 -->
        <div
          id="micro-app-container"
          v-show="!isMainRoute"
          class="micro-app-container"
        />
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/store/user'
import {
  DashboardOutlined,
  ClusterOutlined,
  DeploymentUnitOutlined,
  MonitorOutlined,
  SettingOutlined
} from '@ant-design/icons-vue'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const collapsed = ref(false)
const selectedKeys = ref([route.path])

// 是否是主应用路由 (非子应用路由)
const isMainRoute = computed(() => {
  return route.path === '/' || route.path === '/login'
})

// 监听路由变化
watch(() => route.path, (path) => {
  selectedKeys.value = [path]
})

const handleMenuClick = ({ key }) => {
  router.push(key)
}

const handleLogout = () => {
  userStore.logout()
}
</script>

<style scoped lang="scss">
.main-layout {
  height: 100vh;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #001529;
  padding: 0 20px;
  color: white;
}

.logo {
  font-size: 18px;
  font-weight: bold;
}

.content {
  margin: 16px;
  padding: 24px;
  background: white;
  min-height: 280px;
}

.micro-app-container {
  min-height: 100%;
}
</style>
```

### 8. 主应用路由

修改 `src/router/index.js`:

```javascript
import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/store/user'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/',
      component: () => import('@/layouts/MainLayout.vue'),
      redirect: '/dashboard',
      meta: { requiresAuth: true },
      children: [
        // 子应用路由会被 qiankun 劫持
        // 这里不需要配置子应用的具体路由
      ]
    }
  ]
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth !== false)

  if (requiresAuth && !userStore.isLogin) {
    next('/login')
  } else if (to.path === '/login' && userStore.isLogin) {
    next('/')
  } else {
    next()
  }
})

export default router
```

### 9. Vite 配置

修改 `vite.config.js`:

```javascript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 3000,
    cors: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',  // API 网关
        changeOrigin: true
      }
    }
  },
  build: {
    rollupOptions: {
      output: {
        // 主应用不需要 UMD 格式
        format: 'es'
      }
    }
  }
})
```

---

## 📦 子应用改造

### 1. 创建子应用

以 `dashboard-app` 为例:

```bash
npm create vite@latest dashboard-app -- --template vue
cd dashboard-app
npm install
```

### 2. 子应用目录结构

```
dashboard-app/
├── public/
├── src/
│   ├── views/
│   │   └── Dashboard.vue
│   ├── api/
│   │   └── dashboard.js
│   ├── router/
│   │   └── index.js
│   ├── App.vue
│   ├── main.js
│   └── public-path.js      # 重要: 动态 publicPath
├── vite.config.js
└── package.json
```

### 3. 配置动态 publicPath

创建 `src/public-path.js`:

```javascript
if (window.__POWERED_BY_QIANKUN__) {
  // 动态设置 publicPath
  __webpack_public_path__ = window.__INJECTED_PUBLIC_PATH_BY_QIANKUN__
}
```

### 4. 改造子应用入口

修改 `src/main.js`:

```javascript
import './public-path'
import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import Antd from 'ant-design-vue'
import 'ant-design-vue/dist/reset.css'

import App from './App.vue'
import routes from './router'

let app = null
let router = null
let history = null

/**
 * 渲染函数
 */
function render(props = {}) {
  const { container } = props

  // 创建路由
  history = createWebHistory(
    window.__POWERED_BY_QIANKUN__ ? '/dashboard' : '/'
  )
  router = createRouter({
    history,
    routes
  })

  // 创建应用
  app = createApp(App)
  app.use(router)
  app.use(Antd)

  // 挂载
  const containerEl = container
    ? container.querySelector('#app')
    : document.querySelector('#app')
  app.mount(containerEl)
}

/**
 * 独立运行
 */
if (!window.__POWERED_BY_QIANKUN__) {
  render()
}

/**
 * qiankun 生命周期 - bootstrap
 */
export async function bootstrap() {
  console.log('[dashboard-app] bootstrap')
}

/**
 * qiankun 生命周期 - mount
 */
export async function mount(props) {
  console.log('[dashboard-app] mount', props)
  // 接收主应用传递的数据
  const { userInfo, token, actions } = props

  // 监听全局状态变化
  actions?.onGlobalStateChange?.((state, prev) => {
    console.log('[dashboard-app] 状态变化', state, prev)
  })

  render(props)
}

/**
 * qiankun 生命周期 - unmount
 */
export async function unmount() {
  console.log('[dashboard-app] unmount')
  app?.unmount()
  app = null
  router = null
  history?.destroy?.()
}

/**
 * 可选生命周期 - update
 */
export async function update(props) {
  console.log('[dashboard-app] update', props)
}
```

### 5. 子应用路由

修改 `src/router/index.js`:

```javascript
const routes = [
  {
    path: '/',
    name: 'Dashboard',
    component: () => import('@/views/Dashboard.vue')
  }
]

export default routes
```

### 6. Vite 配置 (重要)

修改 `vite.config.js`:

```javascript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import qiankun from 'vite-plugin-qiankun'

export default defineConfig({
  plugins: [
    vue(),
    qiankun('dashboard-app', {
      useDevMode: true
    })
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 3001,
    cors: true,
    headers: {
      'Access-Control-Allow-Origin': '*'
    }
  },
  base: '/dashboard',  // 生产环境子应用路径
  build: {
    target: 'esnext',
    lib: {
      entry: resolve(__dirname, 'src/main.js'),
      name: 'dashboard-app',
      formats: ['es', 'umd'],
      fileName: (format) => `dashboard-app.${format}.js`
    },
    rollupOptions: {
      external: ['vue', 'ant-design-vue'],  // 外部化依赖
      output: {
        globals: {
          vue: 'Vue',
          'ant-design-vue': 'antd'
        }
      }
    }
  }
})
```

### 7. 安装 vite-plugin-qiankun

```bash
npm install vite-plugin-qiankun -D
```

### 8. package.json

```json
{
  "name": "dashboard-app",
  "version": "1.0.0",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "vue": "^3.3.4",
    "vue-router": "^4.2.4",
    "ant-design-vue": "^4.0.0",
    "axios": "^1.5.0"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^4.3.4",
    "vite": "^4.4.9",
    "vite-plugin-qiankun": "^1.0.15"
  }
}
```

---

## 🔄 状态共享

### 方式1: initGlobalState (推荐)

**主应用**:
```javascript
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

**子应用**:
```javascript
export async function mount(props) {
  const { actions } = props

  // 监听变化
  actions.onGlobalStateChange((state, prev) => {
    console.log('子应用收到状态变化', state, prev)
  })

  // 修改状态
  actions.setGlobalState({ user: { name: 'xxx' } })
}
```

### 方式2: Props 传递

**主应用**:
```javascript
registerMicroApps([
  {
    name: 'dashboard-app',
    props: {
      userInfo: userStore.userInfo,
      permissions: userStore.permissions,
      onUserChange: (user) => {
        userStore.setUser(user)
      }
    }
  }
])
```

**子应用**:
```javascript
export async function mount(props) {
  const { userInfo, permissions, onUserChange } = props
  console.log('接收到主应用数据', userInfo, permissions)

  // 通知主应用
  onUserChange({ name: 'updated' })
}
```

---

## 🎨 样式隔离

### 方式1: CSS 命名空间

```scss
// dashboard-app
.dashboard-app {
  &__header {
    background: #fff;
  }

  &__content {
    padding: 20px;
  }
}
```

### 方式2: CSS Modules

```vue
<template>
  <div :class="$style.container">
    Content
  </div>
</template>

<style module>
.container {
  padding: 20px;
}
</style>
```

### 方式3: Shadow DOM

```javascript
start({
  sandbox: {
    strictStyleIsolation: true  // 开启 Shadow DOM
  }
})
```

---

## 🚢 部署配置

### Nginx 配置

```nginx
server {
  listen 80;
  server_name example.com;

  # 主应用
  location / {
    root /usr/share/nginx/html/main-app;
    try_files $uri $uri/ /index.html;
  }

  # 子应用
  location /dashboard {
    root /usr/share/nginx/html;
    try_files $uri $uri/ /dashboard/index.html;
  }

  location /agent {
    root /usr/share/nginx/html;
    try_files $uri $uri/ /agent/index.html;
  }

  location /cluster {
    root /usr/share/nginx/html;
    try_files $uri $uri/ /cluster/index.html;
  }

  # API 代理
  location /api {
    proxy_pass http://localhost:8080;
  }
}
```

### Docker 部署

```dockerfile
# 主应用
FROM nginx:alpine
COPY dist /usr/share/nginx/html/main-app
COPY nginx.conf /etc/nginx/conf.d/default.conf
```

---

## ✅ 最佳实践

### 1. 子应用独立可运行

每个子应用都应该可以独立开发和运行:
```javascript
if (!window.__POWERED_BY_QIANKUN__) {
  render()  // 独立运行
}
```

### 2. 样式前缀

使用统一的样式前缀避免冲突:
```scss
.dashboard-app-* { }
.agent-app-* { }
```

### 3. 公共依赖外部化

将 Vue、Ant Design 等大型库外部化，减小子应用体积。

### 4. 错误边界

```javascript
addGlobalUncaughtErrorHandler((event) => {
  console.error('微前端错误', event)
  // 上报到监控系统
})
```

### 5. 性能优化

- 启用预加载: `prefetch: true`
- 使用 CDN 加载公共依赖
- 子应用按需加载

---

## 🐛 常见问题

### 1. 子应用加载失败

检查:
- CORS 配置
- 子应用 entry 地址是否正确
- 生命周期函数是否正确导出

### 2. 样式冲突

使用:
- CSS Modules
- 样式命名空间
- Shadow DOM

### 3. 路由不生效

确保:
- 主应用路由使用 `history` 模式
- 子应用 base 配置正确

---

## 📚 参考资源

- [qiankun 官方文档](https://qiankun.umijs.org/)
- [vite-plugin-qiankun](https://github.com/tengmaoqing/vite-plugin-qiankun)
- [微前端实践案例](https://micro-frontends.org/)

---

## 🎉 总结

通过 qiankun 实现微前端架构，可以:
- ✅ 团队独立开发
- ✅ 技术栈独立演进
- ✅ 独立部署发布
- ✅ 提升用户体验

下一步: 查看[示例代码](MICRO_FRONTEND_EXAMPLES.md)进行实践。
