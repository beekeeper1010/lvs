import { createPinia, defineStore } from 'pinia'

// 独立导出 pinia 实例，供路由守卫与 axios 拦截器等非组件环境使用
export const pinia = createPinia()

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    username: localStorage.getItem('username') || '',
    nickname: localStorage.getItem('nickname') || '',
    role: localStorage.getItem('role') || '',
    avatar: localStorage.getItem('avatar') || '',
  }),
  getters: {
    isLoggedIn: (s) => !!s.token,
    isAdmin: (s) => s.role === 'admin',
    displayName: (s) => s.nickname || s.username || '',
    avatarText: (s) => (s.nickname || s.username || '?').charAt(0).toUpperCase(),
    hasAvatar: (s) => !!s.avatar,
  },
  actions: {
    // 登录成功后写入状态与 localStorage
    setLogin(data) {
      this.token = data.token
      this.username = data.username
      this.nickname = data.nickname || data.username
      this.role = data.role || 'user'
      this.avatar = data.avatar || ''
      localStorage.setItem('token', this.token)
      localStorage.setItem('username', this.username)
      localStorage.setItem('nickname', this.nickname)
      localStorage.setItem('role', this.role)
      localStorage.setItem('avatar', this.avatar)
    },
    // 同步服务端最新用户信息（昵称/头像可能被修改）
    setInfo(info) {
      this.username = info.username
      this.nickname = info.nickname || info.username
      this.role = info.role
      this.avatar = info.avatar || ''
      localStorage.setItem('username', this.username)
      localStorage.setItem('nickname', this.nickname)
      localStorage.setItem('role', this.role)
      localStorage.setItem('avatar', this.avatar)
    },
    logout() {
      this.token = ''
      this.username = ''
      this.nickname = ''
      this.role = ''
      this.avatar = ''
      localStorage.removeItem('token')
      localStorage.removeItem('username')
      localStorage.removeItem('nickname')
      localStorage.removeItem('role')
      localStorage.removeItem('avatar')
    },
  },
})
