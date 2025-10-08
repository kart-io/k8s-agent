import request from './request'

/**
 * 用户登录
 * @param {string} username 用户名
 * @param {string} password 密码
 */
export function login(username, password) {
  return request({
    url: '/auth/login',
    method: 'post',
    data: {
      username,
      password
    }
  })
}

/**
 * 用户登出
 */
export function logout() {
  return request({
    url: '/auth/logout',
    method: 'post'
  })
}

/**
 * 获取当前用户信息
 */
export function getUserInfo() {
  return request({
    url: '/auth/me',
    method: 'get'
  })
}

/**
 * 获取用户菜单
 */
export function getUserMenus() {
  return request({
    url: '/auth/menus',
    method: 'get'
  })
}

/**
 * 刷新 Token
 */
export function refreshToken() {
  return request({
    url: '/auth/refresh',
    method: 'post'
  })
}

/**
 * 修改密码
 */
export function changePassword(oldPassword, newPassword) {
  return request({
    url: '/auth/password',
    method: 'put',
    data: {
      old_password: oldPassword,
      new_password: newPassword
    }
  })
}
