import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/store/user'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { title: '登录', requiresAuth: false }
    },
    {
      path: '/',
      component: () => import('@/layouts/MainLayout.vue'),
      redirect: '/dashboard',
      meta: { requiresAuth: true },
      children: [
        {
          path: '/dashboard',
          name: 'Dashboard',
          component: () => import('@/views/Dashboard.vue'),
          meta: { title: '仪表盘', icon: 'DashboardOutlined' }
        },
        {
          path: '/agents',
          name: 'Agents',
          component: () => import('@/views/agents/AgentList.vue'),
          meta: { title: 'Agent 管理', icon: 'ClusterOutlined' }
        },
        {
          path: '/events',
          name: 'Events',
          component: () => import('@/views/events/EventList.vue'),
          meta: { title: '事件监控', icon: 'AlertOutlined' }
        },
        {
          path: '/commands',
          name: 'Commands',
          component: () => import('@/views/commands/CommandList.vue'),
          meta: { title: '命令执行', icon: 'CodeOutlined' }
        },
        {
          path: '/clusters',
          name: 'Clusters',
          component: () => import('@/views/clusters/ClusterList.vue'),
          meta: { title: '集群管理', icon: 'DeploymentUnitOutlined' }
        },
        {
          path: '/alerts',
          name: 'Alerts',
          component: () => import('@/views/alerts/AlertList.vue'),
          meta: { title: '告警规则', icon: 'BellOutlined' }
        }
      ]
    }
  ]
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth !== false)

  if (requiresAuth && !userStore.isLogin) {
    // 需要登录但未登录，跳转到登录页
    next('/login')
  } else if (to.path === '/login' && userStore.isLogin) {
    // 已登录访问登录页，跳转到首页
    next('/')
  } else {
    next()
  }
})

export default router
