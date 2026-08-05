import { createPinia, defineStore } from 'pinia'

// 独立导出 pinia 实例，供路由守卫与 axios 拦截器等非组件环境使用
export const pinia = createPinia()

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    username: localStorage.getItem('username') || '',
    nickname: localStorage.getItem('nickname') || '',
    role: localStorage.getItem('role') || '',
  }),
  getters: {
    isLoggedIn: (s) => !!s.token,
    isAdmin: (s) => s.role === 'admin',
    displayName: (s) => s.nickname || s.username || '',
    avatarText: (s) => (s.nickname || s.username || '?').charAt(0).toUpperCase(),
  },
  actions: {
    // 登录成功后写入状态与 localStorage
    setLogin(data) {
      this.token = data.token
      this.username = data.username
      this.nickname = data.nickname || data.username
      this.role = data.role || 'user'
      localStorage.setItem('token', this.token)
      localStorage.setItem('username', this.username)
      localStorage.setItem('nickname', this.nickname)
      localStorage.setItem('role', this.role)
    },
    // 同步服务端最新用户信息（昵称可能被修改）
    setInfo(info) {
      this.username = info.username
      this.nickname = info.nickname || info.username
      this.role = info.role
      localStorage.setItem('username', this.username)
      localStorage.setItem('nickname', this.nickname)
      localStorage.setItem('role', this.role)
    },
    logout() {
      this.token = ''
      this.username = ''
      this.nickname = ''
      this.role = ''
      localStorage.removeItem('token')
      localStorage.removeItem('username')
      localStorage.removeItem('nickname')
      localStorage.removeItem('role')
    },
  },
})
