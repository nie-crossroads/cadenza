import { createRouter, createWebHistory } from 'vue-router'
import { useSession } from '../composables/useSession'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'cover',
      component: () => import('../views/CoverView.vue'),
    },
    {
      path: '/results',
      name: 'detail',
      component: () => import('../views/HomeView.vue'),
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

// 只探测会话，用来决定要不要显示添加。不根据登录态改路由。
router.beforeEach(async () => {
  const { state, probeSession } = useSession()
  if (!state.checked) await probeSession()
  return true
})

export default router
