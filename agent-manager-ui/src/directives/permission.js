import { useUserStore } from '@/store/user'

/**
 * 权限指令
 * 使用方式: v-permission="'user:create'"
 */
export const permission = {
  mounted(el, binding) {
    const { value } = binding
    const userStore = useUserStore()

    if (value) {
      const hasPermission = userStore.hasPermission(value)
      if (!hasPermission) {
        el.style.display = 'none'
        // 或者直接移除元素
        // el.parentNode && el.parentNode.removeChild(el)
      }
    }
  }
}

/**
 * 角色指令
 * 使用方式: v-role="'admin'"
 */
export const role = {
  mounted(el, binding) {
    const { value } = binding
    const userStore = useUserStore()

    if (value) {
      const hasRole = userStore.hasRole(value)
      if (!hasRole) {
        el.style.display = 'none'
      }
    }
  }
}

// 注册全局指令
export default {
  install(app) {
    app.directive('permission', permission)
    app.directive('role', role)
  }
}
