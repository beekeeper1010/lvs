import axios from 'axios'

const api = axios.create({ baseURL: '/api', timeout: 30000 })

api.interceptors.request.use((cfg) => {
  const token = localStorage.getItem('token')
  if (token) cfg.headers.Authorization = `Bearer ${token}`
  return cfg
})

api.interceptors.response.use(
  (res) => {
    const body = res.data
    if (body && typeof body === 'object' && 'code' in body) {
      if (body.code === 0) return body.data
      if (body.code === 1) {
        localStorage.removeItem('token')
        if (location.pathname !== '/login') location.href = '/login'
      }
      return Promise.reject(new Error(body.msg || '请求失败'))
    }
    return res.data
  },
  (err) => Promise.reject(err)
)

export const login = (username, password) => api.post('/login', { username, password })
export const logout = () => api.post('/logout')
export const fetchVideos = (page, pageSize) => api.get('/video/list', { params: { page, pageSize } })
