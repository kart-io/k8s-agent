<template>
  <a-layout class="main-layout">
    <!-- 侧边栏 -->
    <a-layout-sider
      v-model:collapsed="collapsed"
      :trigger="null"
      collapsible
      :width="220"
      theme="dark"
    >
      <div class="logo">
        <img src="@/assets/logo.svg" alt="Logo" v-if="!collapsed" />
        <img src="@/assets/logo-mini.svg" alt="Logo" v-else />
        <span v-if="!collapsed">Aetherius</span>
      </div>

      <a-menu
        v-model:selectedKeys="selectedKeys"
        mode="inline"
        theme="dark"
        :items="menuItems"
        @click="handleMenuClick"
      />
    </a-layout-sider>

    <!-- 主内容区 -->
    <a-layout>
      <!-- 顶部导航 -->
      <a-layout-header class="header">
        <div class="header-left">
          <menu-unfold-outlined
            v-if="collapsed"
            class="trigger"
            @click="() => (collapsed = !collapsed)"
          />
          <menu-fold-outlined
            v-else
            class="trigger"
            @click="() => (collapsed = !collapsed)"
          />
          <a-breadcrumb class="breadcrumb">
            <a-breadcrumb-item>
              <home-outlined />
            </a-breadcrumb-item>
            <a-breadcrumb-item v-for="item in breadcrumbs" :key="item">
              {{ item }}
            </a-breadcrumb-item>
          </a-breadcrumb>
        </div>

        <div class="header-right">
          <a-space :size="16">
            <a-badge :count="notifications" :overflow-count="99">
              <bell-outlined style="font-size: 18px; cursor: pointer" />
            </a-badge>

            <a-dropdown>
              <a class="user-dropdown">
                <user-outlined style="font-size: 18px" />
                <span class="user-name">管理员</span>
                <down-outlined />
              </a>
              <template #overlay>
                <a-menu>
                  <a-menu-item key="profile">
                    <user-outlined />
                    个人设置
                  </a-menu-item>
                  <a-menu-divider />
                  <a-menu-item key="logout">
                    <logout-outlined />
                    退出登录
                  </a-menu-item>
                </a-menu>
              </template>
            </a-dropdown>
          </a-space>
        </div>
      </a-layout-header>

      <!-- 内容区 -->
      <a-layout-content class="content">
        <div class="content-wrapper">
          <router-view v-slot="{ Component }">
            <transition name="fade" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </div>
      </a-layout-content>

      <!-- 底部 -->
      <a-layout-footer class="footer">
        Aetherius Agent Manager ©2025 Created by Kart
      </a-layout-footer>
    </a-layout>
  </a-layout>
</template>

<script setup>
import { ref, computed, h, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  DashboardOutlined,
  ClusterOutlined,
  AlertOutlined,
  CodeOutlined,
  DeploymentUnitOutlined,
  BellOutlined,
  HomeOutlined,
  UserOutlined,
  LogoutOutlined,
  DownOutlined
} from '@ant-design/icons-vue'

const router = useRouter()
const route = useRoute()

const collapsed = ref(false)
const selectedKeys = ref(['/dashboard'])
const notifications = ref(5)

// 菜单项
const menuItems = ref([
  {
    key: '/dashboard',
    icon: () => h(DashboardOutlined),
    label: '仪表盘',
    title: '仪表盘'
  },
  {
    key: '/agents',
    icon: () => h(ClusterOutlined),
    label: 'Agent 管理',
    title: 'Agent 管理'
  },
  {
    key: '/events',
    icon: () => h(AlertOutlined),
    label: '事件监控',
    title: '事件监控'
  },
  {
    key: '/commands',
    icon: () => h(CodeOutlined),
    label: '命令执行',
    title: '命令执行'
  },
  {
    key: '/clusters',
    icon: () => h(DeploymentUnitOutlined),
    label: '集群管理',
    title: '集群管理'
  },
  {
    key: '/alerts',
    icon: () => h(BellOutlined),
    label: '告警规则',
    title: '告警规则'
  }
])

// 面包屑
const breadcrumbs = computed(() => {
  const matched = route.matched.filter(r => r.meta && r.meta.title)
  return matched.map(r => r.meta.title)
})

// 监听路由变化
watch(
  () => route.path,
  (path) => {
    selectedKeys.value = [path]
  },
  { immediate: true }
)

// 菜单点击
const handleMenuClick = ({ key }) => {
  router.push(key)
}
</script>

<style lang="scss" scoped>
.main-layout {
  height: 100vh;

  .logo {
    height: 64px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 0 16px;
    background: rgba(255, 255, 255, 0.1);

    img {
      height: 32px;
    }

    span {
      color: #fff;
      font-size: 18px;
      font-weight: 600;
    }
  }

  .header {
    background: #fff;
    padding: 0 24px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);

    .header-left {
      display: flex;
      align-items: center;
      gap: 16px;

      .trigger {
        font-size: 18px;
        cursor: pointer;
        transition: color 0.3s;

        &:hover {
          color: #1890ff;
        }
      }

      .breadcrumb {
        margin: 0;
      }
    }

    .header-right {
      .user-dropdown {
        display: flex;
        align-items: center;
        gap: 8px;
        cursor: pointer;

        .user-name {
          margin: 0 4px;
        }
      }
    }
  }

  .content {
    margin: 16px;
    overflow: auto;

    .content-wrapper {
      min-height: calc(100vh - 180px);
    }
  }

  .footer {
    text-align: center;
    background: #f0f2f5;
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
