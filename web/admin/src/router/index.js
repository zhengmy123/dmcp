import { createRouter, createWebHistory } from 'vue-router'
import NProgress from 'nprogress'
import { useAuthStore } from '@/stores/auth'
import { JWT_TOKEN_KEY } from '@/types'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/pages/LoginPage.vue'),
    meta: { title: '登录', requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/pages/DashboardPage.vue'),
        meta: { title: '仪表盘', requiresAuth: true }
      },
      {
        path: 'tokens',
        name: 'Tokens',
        component: () => import('@/pages/TokensPage.vue'),
        meta: { title: 'Token 管理', requiresAuth: true }
      },
      {
        path: 'tools',
        name: 'Tools',
        component: () => import('@/pages/ToolsPage.vue'),
        meta: { title: '工具定义', requiresAuth: true }
      },
      {
        path: 'services',
        name: 'Services',
        component: () => import('@/pages/ServicesPage.vue'),
        meta: { title: 'HTTP 服务', requiresAuth: true }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/pages/SettingsPage.vue'),
        meta: { title: '系统设置', requiresAuth: true }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  NProgress.start()
  document.title = `${to.meta.title || 'MCP Admin'} - MCP Server`

  // 检查是否需要认证
  if (to.meta.requiresAuth !== false) {
    const jwtToken = localStorage.getItem(JWT_TOKEN_KEY)
    if (!jwtToken) {
      // 未登录，跳转到登录页
      next({ name: 'Login' })
      return
    }
  }

  // 如果已登录且访问登录页，跳转到首页
  if (to.name === 'Login') {
    const jwtToken = localStorage.getItem(JWT_TOKEN_KEY)
    if (jwtToken) {
      next({ name: 'Dashboard' })
      return
    }
  }

  next()
})

router.afterEach(() => {
  NProgress.done()
})

export default router
