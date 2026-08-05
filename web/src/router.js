import { createRouter, createWebHistory } from 'vue-router'
import { pinia, useUserStore } from './stores/user'
import Login from './views/Login.vue'
import Gallery from './views/Gallery.vue'
import Player from './views/Player.vue'
import AdminUsers from './views/AdminUsers.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/gallery' },
    { path: '/login', component: Login },
    { path: '/gallery', component: Gallery },
    { path: '/play/:id', component: Player },
    { path: '/users', component: AdminUsers, meta: { requiresAdmin: true } },
  ],
})

router.beforeEach((to) => {
  const store = useUserStore(pinia)
  if (to.path !== '/login' && !store.isLoggedIn) return '/login'
  if (to.path === '/login' && store.isLoggedIn) return '/gallery'
  // 用户管理页仅 admin 可访问
  if (to.meta.requiresAdmin && !store.isAdmin) return '/gallery'
})

export default router
