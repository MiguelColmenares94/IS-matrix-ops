import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../domains/auth/auth.store'

const routes = [
  { path: '/login', component: () => import('../domains/auth/LoginView.vue') },
  { path: '/', component: () => import('../domains/matrix/MatrixView.vue'), meta: { requiresAuth: true } },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isAuthenticated) return '/login'
  if (to.path === '/login' && auth.isAuthenticated) return '/'
})

export default router
