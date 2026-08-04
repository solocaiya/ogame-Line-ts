/**
 * API 服务 — 与 OGame 后端服务器通信
 */

const API_BASE = import.meta.env.VITE_API_BASE || '/api'

interface AuthResponse {
  access_token: string
  refresh_token: string
  user: UserInfo
}

interface UserInfo {
  id: string
  username: string
  created_at: string
  last_login: string
  is_active: boolean
}

interface SaveGameRequest {
  slot?: string
  gameData: string
  npcData?: string
  universeData?: string
}

interface SaveGameResponse {
  message: string
  id: string
  slot: string
  savedAt: string
}

interface LoadGameResponse {
  gameData: string
  npcData: string
  universeData: string
  savedAt: string
}

interface SaveInfo {
  slot: string
  savedAt: string
}

class ApiService {
  private accessToken: string | null = null
  private refreshToken: string | null = null
  private refreshPromise: Promise<boolean> | null = null // prevents concurrent refresh races
  private onTokenChange: ((access: string | null, refresh: string | null) => void) | null = null

  constructor() {
    // 从 localStorage 恢复 token
    this.accessToken = localStorage.getItem('access_token')
    this.refreshToken = localStorage.getItem('refresh_token')
  }

  /** Register a callback invoked whenever tokens change (used by authStore to stay in sync). */
  setTokenCallback(cb: (access: string | null, refresh: string | null) => void) {
    this.onTokenChange = cb
  }

  private getHeaders(): Record<string, string> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json'
    }
    if (this.accessToken) {
      headers['Authorization'] = `Bearer ${this.accessToken}`
    }
    return headers
  }

  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const url = `${API_BASE}${path}`
    const res = await fetch(url, {
      method,
      headers: this.getHeaders(),
      body: body ? JSON.stringify(body) : undefined
    })

    if (res.status === 401 && this.refreshToken) {
      // Token 过期，尝试刷新
      const refreshed = await this.doRefresh()
      if (refreshed) {
        // 重试原请求
        const retryRes = await fetch(url, {
          method,
          headers: this.getHeaders(),
          body: body ? JSON.stringify(body) : undefined
        })
        if (retryRes.ok) return retryRes.json()
        const err = await retryRes.json().catch(() => ({ error: 'request failed' }))
        throw new Error(err.error || `HTTP ${retryRes.status}`)
      }
      // 刷新失败，清除 token
      this.clearTokens()
    }

    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: 'request failed' }))
      throw new Error(err.error || `HTTP ${res.status}`)
    }

    return res.json()
  }

  private async doRefresh(): Promise<boolean> {
    // Deduplicate concurrent refresh attempts — all callers await the same promise
    if (this.refreshPromise) return this.refreshPromise
    this.refreshPromise = (async () => {
      try {
        const res = await fetch(`${API_BASE}/auth/refresh`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: this.refreshToken })
        })
        if (!res.ok) return false

        const data = await res.json()
        this.setTokens(data.access_token, data.refresh_token)
        return true
      } catch {
        return false
      } finally {
        this.refreshPromise = null
      }
    })()
    return this.refreshPromise
  }

  setTokens(access: string, refresh: string) {
    this.accessToken = access
    this.refreshToken = refresh
    localStorage.setItem('access_token', access)
    localStorage.setItem('refresh_token', refresh)
    this.onTokenChange?.(access, refresh)
  }

  clearTokens() {
    this.accessToken = null
    this.refreshToken = null
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    this.onTokenChange?.(null, null)
  }

  getAccessToken() {
    return this.accessToken
  }

  // --- Auth ---

  async register(username: string, password: string): Promise<AuthResponse> {
    const data = await this.request<AuthResponse>('POST', '/auth/register', { username, password })
    this.setTokens(data.access_token, data.refresh_token)
    return data
  }

  async login(username: string, password: string): Promise<AuthResponse> {
    const data = await this.request<AuthResponse>('POST', '/auth/login', { username, password })
    this.setTokens(data.access_token, data.refresh_token)
    return data
  }

  async getMe(): Promise<UserInfo> {
    return this.request<UserInfo>('GET', '/auth/me')
  }

  // --- Player Save ---

  async saveGame(gameData: string, npcData?: string, universeData?: string, slot = 'default'): Promise<SaveGameResponse> {
    return this.request<SaveGameResponse>('PUT', '/player/save', {
      slot,
      gameData,
      npcData: npcData || '',
      universeData: universeData || ''
    })
  }

  async loadGame(slot = 'default'): Promise<LoadGameResponse> {
    return this.request<LoadGameResponse>('GET', `/player/save?slot=${encodeURIComponent(slot)}`)
  }

  async listSaves(): Promise<SaveInfo[]> {
    return this.request<SaveInfo[]>('GET', '/player/saves')
  }

  async deleteSave(slot = 'default'): Promise<{ message: string }> {
    return this.request<{ message: string }>('DELETE', `/player/save?slot=${encodeURIComponent(slot)}`)
  }

  // --- Game ---

  async initPlayer(): Promise<any> {
    return this.request('POST', '/game/init')
  }

  async getGameState(): Promise<any> {
    return this.request('GET', '/game/state')
  }

  async startBuilding(planetId: string, buildingType: string): Promise<any> {
    return this.request('POST', '/game/building/start', { planetId, buildingType })
  }

  async cancelBuilding(planetId: string): Promise<any> {
    return this.request('POST', '/game/building/cancel', { planetId })
  }

  async startShipProduction(planetId: string, shipType: string, count: number): Promise<any> {
    return this.request('POST', '/game/ship/start', { planetId, shipType, count })
  }

  async cancelShipProduction(planetId: string): Promise<any> {
    return this.request('POST', '/game/ship/cancel', { planetId })
  }

  async startDefenseProduction(planetId: string, defenseType: string, count: number): Promise<any> {
    return this.request('POST', '/game/defense/start', { planetId, defenseType, count })
  }

  async cancelDefenseProduction(planetId: string): Promise<any> {
    return this.request('POST', '/game/defense/cancel', { planetId })
  }

  async startResearch(planetId: string, researchType: string): Promise<any> {
    return this.request('POST', '/game/research/start', { planetId, researchType })
  }

  async cancelResearch(planetId: string): Promise<any> {
    return this.request('POST', '/game/research/cancel', { planetId })
  }

  async sendFleet(params: {
    origin: { galaxy: number; system: number; position: number }
    target: { galaxy: number; system: number; position: number }
    fleet: Record<string, number>
    cargo: { metal: number; crystal: number; deuterium: number }
    missionType: string
  }): Promise<any> {
    return this.request('POST', '/game/fleet/send', params)
  }

  // --- Health ---

  async health(): Promise<{ status: string; time: string }> {
    return this.request<{ status: string; time: string }>('GET', '/health')
  }
}

export const apiService = new ApiService()
export type { UserInfo, SaveGameResponse, LoadGameResponse, SaveInfo }
