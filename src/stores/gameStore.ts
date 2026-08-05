import { defineStore } from 'pinia'
import type {
  Planet,
  Player,
  BuildQueueItem,
  FleetMission,
  BattleResult,
  SpyReport,
  Officer,
  SpiedNotification,
  NPCActivityNotification,
  IncomingFleetAlert,
  MissileAttack,
  AchievementStats,
  AchievementProgress,
  WebDAVConfig
} from '@/types/game'
import { TechnologyType, OfficerType } from '@/types/game'
import { initializeAchievementStats, initializeAchievements } from '@/logic/achievementLogic'
import type { Locale } from '@/locales'
import pkg from '../../package.json'
import { encryptData, decryptData } from '@/utils/crypto'
import { apiService } from '@/services/apiService'

export const useGameStore = defineStore('game', {
  state: () => ({
    gameTime: Date.now(),
    isPaused: false,
    gameSpeed: 1,
    battleToFinish: true, // 锁定100回合模式（前后端双战斗系统结果不一致，暂时禁用6回合切换）
    player: {
      id: 'player1',
      name: '',
      planets: [] as Planet[],
      technologies: {} as Record<TechnologyType, number>,
      officers: {} as Record<OfficerType, Officer>,
      researchQueue: [] as BuildQueueItem[],
      waitingResearchQueue: [],
      fleetMissions: [] as FleetMission[],
      missileAttacks: [] as MissileAttack[],
      battleReports: [] as BattleResult[],
      spyReports: [] as SpyReport[],
      spiedNotifications: [] as SpiedNotification[],
      npcActivityNotifications: [] as NPCActivityNotification[],
      missionReports: [],
      incomingFleetAlerts: [] as IncomingFleetAlert[],
      giftNotifications: [],
      giftRejectedNotifications: [],
      points: 0,
      bonusPoints: 0,
      isGMEnabled: false, // 明确设置 GM 模式默认为 false
      lastVersionCheckTime: 0, // 最后一次检查版本的时间戳，默认为0
      achievementStats: initializeAchievementStats() as AchievementStats,
      achievements: initializeAchievements() as Record<string, AchievementProgress>
    } as Player,
    currentPlanetId: '',
    isDark: '',
    locale: 'zh-CN' as Locale,
    notificationSettings: {
      browser: false,
      inApp: true,
      suppressInFocus: false,
      types: {
        construction: true,
        research: true,
        unlock: true
      }
    },
    webdavConfig: null as WebDAVConfig | null,
    // Server sync state
    _lastSyncTime: 0,
    _pendingSync: false,
    _optimisticQueue: null as Promise<unknown> | null // serializes optimistic updates
  }),
  actions: {
    async requestBrowserPermission(): Promise<boolean> {
      if (!('Notification' in window)) return false

      if (Notification.permission === 'granted') return true

      const permission = await Notification.requestPermission()
      return permission === 'granted'
    },
    toggleQueuePaused(planetId: string) {
      const planet = this.player.planets.find(p => p.id === planetId)
      if (planet) {
        planet.queuePaused = !planet.queuePaused
      }
    },
    cancelBuildQueueItem(planetId: string, itemId: string) {
      const planet = this.player.planets.find(p => p.id === planetId)
      if (planet && planet.buildQueue.length > 0) {
        // 只能取消第一个（当前正在建造的）
        const firstItem = planet.buildQueue[0]
        if (firstItem.id === itemId) {
          planet.buildQueue.shift()
        }
      }
    },

    // --- Server Sync (cache layer pattern) ---

    /**
     * Fetch authoritative state from server and merge into local store.
     * Local state is treated as a cache — server is the source of truth.
     */
    async syncFromServer(): Promise<boolean> {
      if (this._pendingSync) return false
      this._pendingSync = true
      try {
        const response = await apiService.getGameState()
        if (response && response.player) {
          // Merge server player state into local store
          const serverPlayer = response.player
          // Preserve local-only fields that server doesn't manage
          const localOnly = {
            achievementStats: this.player.achievementStats,
            achievements: this.player.achievements,
            battleReports: this.player.battleReports,
            spyReports: this.player.spyReports,
            spiedNotifications: this.player.spiedNotifications,
            npcActivityNotifications: this.player.npcActivityNotifications,
            incomingFleetAlerts: this.player.incomingFleetAlerts,
            giftNotifications: this.player.giftNotifications,
            giftRejectedNotifications: this.player.giftRejectedNotifications,
            missileAttacks: this.player.missileAttacks,
            missionReports: this.player.missionReports
          }
          // Update player from server
          Object.assign(this.player, serverPlayer, localOnly)
          this._lastSyncTime = Date.now()
          return true
        }
        return false
      } catch (e) {
        console.warn('[GameStore] syncFromServer failed:', e)
        return false
      } finally {
        this._pendingSync = false
      }
    },

    /**
     * Apply a partial update from WebSocket event.
     * Only updates the specific fields mentioned in the event.
     */
    applyServerEvent(eventType: string, data: any) {
      switch (eventType) {
        case 'buildingComplete': {
          const planet = this.player.planets.find(p => p.id === data.planetId)
          if (planet) {
            planet.buildings = planet.buildings || {}
            planet.buildings[data.building] = data.level
            // Remove from build queue if present
            const idx = planet.buildQueue?.findIndex(q => q.type === data.building)
            if (idx !== undefined && idx >= 0) {
              planet.buildQueue.splice(idx, 1)
            }
          }
          break
        }
        case 'researchComplete': {
          this.player.technologies = this.player.technologies || {}
          this.player.technologies[data.type] = data.level
          break
        }
        case 'shipComplete': {
          const planet = this.player.planets.find(p => p.id === data.planetId)
          if (planet) {
            planet.ships = planet.ships || {}
            planet.ships[data.type] = (planet.ships[data.type] || 0) + data.count
            // Remove from ship queue if present
            const idx = planet.shipQueue?.findIndex(q => q.type === data.type)
            if (idx !== undefined && idx >= 0) {
              planet.shipQueue.splice(idx, 1)
            }
          }
          break
        }
        case 'defenseComplete': {
          const planet = this.player.planets.find(p => p.id === data.planetId)
          if (planet) {
            planet.defenses = planet.defenses || {}
            planet.defenses[data.type] = (planet.defenses[data.type] || 0) + data.count
            // Remove from defense queue if present
            const idx = planet.defenseQueue?.findIndex(q => q.type === data.type)
            if (idx !== undefined && idx >= 0) {
              planet.defenseQueue.splice(idx, 1)
            }
          }
          break
        }
        case 'fleetReturned':
        case 'fleetArrived':
          // Trigger full sync for fleet changes (complex state)
          this.syncFromServer()
          break
        case 'battleResult':
          // Trigger full sync after battle (ships/defenses changed)
          this.syncFromServer()
          break
      }
    },

    /**
     * Optimistic update pattern (serialized — prevents concurrent clobbering):
     * 1. Wait for any prior optimistic update to finish
     * 2. Snapshot current state
     * 3. Apply local change immediately
     * 4. Call server API
     * 5. On failure, rollback to snapshot
     */
    optimisticUpdate<T>(
      applyLocal: () => void,
      serverCall: () => Promise<T>,
      onError?: (error: Error) => void
    ): Promise<T | null> {
      const run = async (): Promise<T | null> => {
        // Snapshot for rollback
        const snapshot = JSON.parse(JSON.stringify(this.player))

        // Apply immediately (optimistic)
        applyLocal()

        try {
          const result = await serverCall()
          return result
        } catch (error) {
          // Rollback
          console.warn('[GameStore] Optimistic update failed, rolling back:', error)
          Object.assign(this.player, snapshot)
          if (onError && error instanceof Error) {
            onError(error)
          }
          return null
        }
      }

      // Chain onto the queue so updates serialize — no interleaving
      const prev = this._optimisticQueue ?? Promise.resolve()
      const next = prev.then(() => run(), () => run())
      this._optimisticQueue = next.then(() => {}, () => {}) // swallow for queue
      return next
    }
  },
  getters: {
    currentPlanet(): Planet | undefined {
      return this.player.planets.find(p => p.id === this.currentPlanetId)
    },
    getMoonForPlanet(): (planetId: string) => Planet | undefined {
      return (planetId: string) => {
        return this.player.planets.find(p => p.parentPlanetId === planetId && p.isMoon)
      }
    }
  },
  persist: {
    key: pkg.name,
    storage: localStorage,
    serializer: {
      serialize: state => {
        // Strip transient flags — _pendingSync must not survive a crash/reload
        const { _pendingSync, ...rest } = state
        void _pendingSync
        return encryptData(rest)
      },
      deserialize: value => {
        const data = decryptData(value)
        // Ensure _pendingSync always starts as false after load
        return { ...data, _pendingSync: false }
      }
    }
  }
})
