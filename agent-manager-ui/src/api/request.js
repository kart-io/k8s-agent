import axios from 'axios'
import { message } from 'ant-design-vue'

const request = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// Request interceptor
request.interceptors.request.use(
  (config) => {
    // 添加 token
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor
request.interceptors.response.use(
  (response) => {
    const res = response.data
    return res
  },
  (error) => {
    // 401 未授权，跳转登录
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
      message.error('登录已过期，请重新登录')
    } else {
      message.error(error.response?.data?.error || error.message || 'Request failed')
    }
    return Promise.reject(error)
  }
)

export default request
