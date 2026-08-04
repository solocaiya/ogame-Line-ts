import { createRouter, createWebHashHistory } from 'vue-router'
import { useGameStore } from '@/stores/gameStore'
import { useAuthStore } from '@/stores/authStore'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue'), meta: { public: true } },
    { path: '/', name: 'home', component: () => import('@/views/HomeView.vue') },
    { path: '/overview', name: 'overview', component: () => import('@/views/OverviewView.vue') },
    { path: '/buildings', name: 'buildings', component: () => import('@/views/BuildingsView.vue') },
    { path: '/research', name: 'research', component: () => import('@/views/ResearchView.vue') },
    { path: '/shipyard', name: 'shipyard', component: () => import('@/views/ShipyardView.vue') },
    { path: '/defense', name: 'defense', component: () => import('@/views/DefenseView.vue') },
    { path: '/fleet', name: 'fleet', component: () => import('@/views/FleetView.vue') },
    { path: '/officers', name: 'officers', component: () => import('@/views/OfficersView.vue') },
    { path: '/battle-simulator', name: 'battle-simulator', component: () => import('@/views/BattleSimulatorView.vue') },
    { path: '/messages', name: 'messages', component: () => import('@/views/MessagesView.vue') },
    { path: '/galaxy', name: 'galaxy', component: () => import('@/views/GalaxyView.vue') },
    { path: '/diplomacy', name: 'diplomacy', component: () => import('@/views/DiplomacyView.vue') },
    { path: '/achievements', name: 'achievements', component: () => import('@/views/AchievementsView.vue') },
    { path: '/checkin', name: 'checkin', component: () => import('@/views/CheckInView.vue') },
    { path: '/campaign', name: 'campaign', component: () => import('@/views/CampaignView.vue') },
    { path: '/ranking', name: 'ranking', component: () => import('@/views/RankingView.vue') },
    { path: '/trader', name: 'trader', component: () => import('@/views/TraderView.vue') },
    { path: '/bookmarks', name: 'bookmarks', component: () => import('@/views/BookmarkView.vue') },
    { path: '/statistics', name: 'statistics', component: () => import('@/views/StatisticsView.vue') },
    { path: '/battle-reports', name: 'battle-reports', component: () => import('@/views/BattleReportsView.vue') },
    { path: '/planet-queue', name: 'planet-queue', component: () => import('@/views/PlanetQueueView.vue') },
    { path: '/settings', name: 'settings', component: () => import('@/views/SettingsView.vue') },
    { path: '/gm', name: 'gm', component: () => import('@/views/GMView.vue') },
    { path: '/:pathMatch(.*)*', name: 'not-found', component: () => import('@/views/NotFoundView.vue') }
  ]
})

// 路由守卫：检查登录状态和隐私协议同意状态
router.beforeEach((to, _from, next) => {
  const gameStore = useGameStore()
  const authStore = useAuthStore()

  // 公开页面（登录等）始终可访问
  if (to.meta.public) {
    next()
    return
  }

  // 是否已认证（登录或游客模式）
  // 用 accessToken 判断而非 isLoggedIn，避免 fetchUser 未完成时的竞态条件
  const isAuthenticated = !!authStore.accessToken
  const isGuest = localStorage.getItem('guest_mode') === 'true'
  const hasAccess = isAuthenticated || isGuest

  if (!hasAccess) {
    // 首页允许访问（用于显示登录入口）
    if (to.path === '/') {
      next()
      return
    }
    // 其他页面重定向到登录页
    next('/login')
    return
  }

  // 已同意隐私协议
  if (gameStore.player.privacyAgreed) {
    // 已同意但访问首页，重定向到总览页
    if (to.path === '/') {
      next('/overview')
      return
    }
    // 正常访问其他页面
    next()
    return
  }

  // 未同意隐私协议 → 去登录页处理（登录成功后弹出隐私协议）
  next('/login')
})

export default router
