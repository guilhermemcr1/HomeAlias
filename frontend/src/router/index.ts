import { createRouter, createWebHistory } from 'vue-router'
import { installAuthRedirect, useAuth } from '../composables/useAuth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('../pages/LoginPage.vue'), meta: { public: true, title: 'Entrar' } },
    { path: '/convite', component: () => import('../pages/AcceptInvitePage.vue'), meta: { public: true, title: 'Aceitar convite' } },
    { path: '/', component: () => import('../pages/DashboardPage.vue'), meta: { title: 'Painel' } },
    { path: '/connections', component: () => import('../pages/ConnectionsPage.vue'), meta: { title: 'Conexões' } },
    { path: '/hosts', component: () => import('../pages/HostsPage.vue'), meta: { title: 'Hosts' } },
    { path: '/tokens', component: () => import('../pages/TokenEmitPage.vue'), meta: { title: 'Tokens' } },
    { path: '/history', component: () => import('../pages/HistoryPage.vue'), meta: { title: 'Histórico' } },
    { path: '/alerts', component: () => import('../pages/AlertsPage.vue'), meta: { title: 'Alertas' } },
    { path: '/admin/users', component: () => import('../pages/AdminUsersPage.vue'), meta: { admin: true, title: 'Usuários' } },
    { path: '/admin/audit', component: () => import('../pages/AuditPage.vue'), meta: { admin: true, title: 'Auditoria' } },
    { path: '/conta', component: () => import('../pages/AccountPage.vue'), meta: { title: 'Minha conta' } },
    { path: '/ajuda', component: () => import('../pages/HelpPage.vue'), meta: { title: 'Ajuda' } },
    { path: '/:pathMatch(.*)*', component: () => import('../pages/NotFoundPage.vue'), meta: { title: 'Página não encontrada' } },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuth()
  await auth.ensure()
  const logged = !!auth.user.value
  if (to.meta.public) {
    return logged && to.path === '/login' ? { path: '/' } : true
  }
  if (!logged) {
    return { path: '/login', query: to.fullPath === '/' ? {} : { next: to.fullPath } }
  }
  if (to.meta.admin && !auth.isAdmin.value) return { path: '/' }
  return true
})

router.afterEach((to) => {
  document.title = `${to.meta.title ?? 'HomeAlias'} · HomeAlias`
})

installAuthRedirect(() => {
  const current = router.currentRoute.value
  if (!current.meta.public) {
    router.push({ path: '/login', query: current.fullPath === '/' ? {} : { next: current.fullPath } })
  }
})

export default router
