import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import { useAuth } from '../composables/useAuth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/article/:slug',
      name: 'article',
      component: () => import('../views/ArticleView.vue'),
    },
    {
      path: '/about',
      name: 'about',
      component: () => import('../views/AboutView.vue'),
    },
    {
      path: '/projects',
      name: 'projects',
      component: () => import('../views/ProjectsView.vue'),
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
    },
    {
      path: '/admin',
      name: 'admin',
      component: () => import('../views/admin/PostListView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/admin/edit',
      name: 'admin-new',
      component: () => import('../views/admin/PostEditView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/admin/edit/:slug',
      name: 'admin-edit',
      component: () => import('../views/admin/PostEditView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/error',
      name: 'error',
      component: () => import('../views/ErrorView.vue'),
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      redirect: () => ({ name: 'error', query: { code: '404' } }),
    },
  ],
})

// 路由守卫: admin 路由需要登录
router.beforeEach((to, _from, next) => {
  const { isLoggedIn } = useAuth()
  if (to.meta.requiresAuth && !isLoggedIn.value) {
    next({ name: 'login' })
  } else {
    next()
  }
})

export default router
