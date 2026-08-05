import { createRouter, createWebHistory } from 'vue-router'
import Login from './views/Login.vue'
import Gallery from './views/Gallery.vue'
import Player from './views/Player.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/gallery' },
    { path: '/login', component: Login },
    { path: '/gallery', component: Gallery },
    { path: '/play/:id', component: Player },
  ],
})

router.beforeEach((to) => {
  const token = localStorage.getItem('token')
  if (to.path !== '/login' && !token) return '/login'
  if (to.path === '/login' && token) return '/gallery'
})

export default router
