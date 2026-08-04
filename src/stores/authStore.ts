import { defineStore } from 'pinia'
import { apiService, type UserInfo } from '@/services/apiService'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null as UserInfo | null,
    accessToken: localStorage.getItem('access_token'),
    refreshToken: localStorage.getItem('refresh_token'),
    loading: false,
    error: null as string | null
  }),

  getters: {
    isLoggedIn: (state) => !!state.accessToken && !!state.user,
    isGuest: (state) => state.user?.is_guest ?? false,
    username: (state) => state.user?.username || ''
  },

  actions: {
    init() {
      // Keep authStore tokens in sync when apiService refreshes or clears them
      apiService.setTokenCallback((access, refresh) => {
        this.accessToken = access
        this.refreshToken = refresh
      })
    },

    async register(username: string, password: string) {
      this.loading = true
      this.error = null
      try {
        const res = await apiService.register(username, password)
        this.user = res.user
        this.accessToken = res.access_token
        this.refreshToken = res.refresh_token
      } catch (e: any) {
        this.error = e.message || '注册失败'
        throw e
      } finally {
        this.loading = false
      }
    },

    async login(username: string, password: string) {
      this.loading = true
      this.error = null
      try {
        const res = await apiService.login(username, password)
        this.user = res.user
        this.accessToken = res.access_token
        this.refreshToken = res.refresh_token
      } catch (e: any) {
        this.error = e.message || '登录失败'
        throw e
      } finally {
        this.loading = false
      }
    },

    async fetchUser() {
      if (!this.accessToken) return
      try {
        this.user = await apiService.getMe()
      } catch (e: any) {
        // 区分 401（token 无效）和网络错误（如离线、超时）
        // 只有 token 无效才登出，网络抖动保持 session
        if (e?.message?.includes('401') || e?.status === 401 || e?.response?.status === 401) {
          this.logout()
        }
        // 网络错误不清 session，让用户继续操作
      }
    },

    async enterGuest(deviceId: string) {
      this.loading = true
      this.error = null
      try {
        const res = await apiService.guest(deviceId)
        this.user = res.user
        this.accessToken = res.access_token
        this.refreshToken = res.refresh_token
      } catch (e: any) {
        this.error = e.message || '游客登录失败'
        throw e
      } finally {
        this.loading = false
      }
    },

    async bindAccount(username: string, password: string) {
      this.loading = true
      this.error = null
      try {
        const res = await apiService.bindAccount(username, password)
        this.user = res.user
        this.accessToken = res.access_token
        this.refreshToken = res.refresh_token
      } catch (e: any) {
        this.error = e.message || '绑定账号失败'
        throw e
      } finally {
        this.loading = false
      }
    },

    logout() {
      this.user = null
      this.accessToken = null
      this.refreshToken = null
      apiService.clearTokens()
    },

    clearError() {
      this.error = null
    }
  }
})
