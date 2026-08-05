import axios from 'axios'
import { pinia, useUserStore } from './stores/user'

const api = axios.create({ baseURL: '/api', timeout: 30000 })

api.interceptors.request.use((cfg) => {
  const store = useUserStore(pinia)
  if (store.token) cfg.headers.Authorization = `Bearer ${store.token}`
  return cfg
})

api.interceptors.response.use(
  (res) => {
    const body = res.data
    if (body && typeof body === 'object' && 'code' in body) {
      if (body.code === 0) return body.data
      // 业务错误（HTTP 200）仅抛出错误信息，不跳转
      return Promise.reject(new Error(body.msg || '请求失败'))
    }
    return res.data
  },
  (err) => {
    // 仅认证失败（HTTP 401）时登出并跳转登录页
    if (err.response && err.response.status === 401) {
      useUserStore(pinia).logout()
      if (location.pathname !== '/login') location.href = '/login'
    }
    return Promise.reject(err)
  }
)

export const login = (username, password) => api.post('/login', { username, password })
export const logout = () => api.post('/logout')
export const fetchUserInfo = () => api.get('/user/info')
export const updateProfile = (data) => api.put('/user/profile', data)
export const fetchVideos = (page, pageSize) => api.get('/video/list', { params: { page, pageSize } })
export const fetchAdjacent = (id) => api.get('/video/adjacent', { params: { id } })

// 用户管理（仅 admin）
export const fetchUsers = () => api.get('/admin/users')
export const createUser = (data) => api.post('/admin/users', data)
export const updateUser = (id, data) => api.put(`/admin/users/${id}`, data)
export const deleteUser = (id) => api.delete(`/admin/users/${id}`)
