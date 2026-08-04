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
      } catch {
        // /auth/me 失败意味着 token 无效或网络异常，统一登出
        // apiService 在 refresh 失败时已清除 token，这里只需清理 user 状态
        this.logout()
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
