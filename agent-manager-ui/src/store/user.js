import { defineStore } from 'pinia'
import { login, logout, getUserInfo, getUserMenus } from '@/api/auth'
import router from '@/router'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    userInfo: null,
    menus: [],
    permissions: []
  }),

  getters: {
    isLogin: (state) => !!state.token,
    username: (state) => state.userInfo?.username || '',
    avatar: (state) => state.userInfo?.avatar || '',
    roles: (state) => state.userInfo?.roles || []
  },

  actions: {
    // 用户登录
    async login(username, password) {
      try {
        const res = await login(username, password)
        this.token = res.token
        this.userInfo = res.user

        // 保存 token 到 localStorage
        localStorage.setItem('token', res.token)

        return res
      } catch (error) {
        throw error
      }
    },

    // 获取用户信息
    async fetchUserInfo() {
      try {
        const userInfo = await getUserInfo()
        this.userInfo = userInfo
        return userInfo
      } catch (error) {
        throw error
      }
    },

    // 获取用户菜单
    async fetchUserMenus() {
      try {
        const menus = await getUserMenus()
        this.menus = menus

        // 提取所有权限
        const permissions = []
        const extractPermissions = (items) => {
          items.forEach(item => {
            if (item.permissions) {
              permissions.push(...item.permissions)
            }
            if (item.children) {
              extractPermissions(item.children)
            }
          })
        }
        extractPermissions(menus)
        this.permissions = permissions

        return menus
      } catch (error) {
        throw error
      }
    },

    // 用户登出
    async logout() {
      try {
        await logout()
      } catch (error) {
        console.error('Logout error:', error)
      } finally {
        this.token = ''
        this.userInfo = null
        this.menus = []
        this.permissions = []
        localStorage.removeItem('token')
        router.push('/login')
      }
    },

    // 检查权限
    hasPermission(permission) {
      if (!permission) return true
      return this.permissions.includes(permission)
    },

    // 检查角色
    hasRole(role) {
      if (!role) return true
      return this.roles.some(r => r.code === role)
    }
  }
})
