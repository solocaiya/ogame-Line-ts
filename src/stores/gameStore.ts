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

export const useGameStore = defineStore('game', {
  state: () => ({
    gameTime: Date.now(),
    isPaused: false,
    gameSpeed: 1,
    battleToFinish: true, // 战斗到底模式：false=经典模式(6回合平局)，true=战斗到底(最多100回合)
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
    webdavConfig: null as WebDAVConfig | null
  }),
  actions: {
    async requestBrowserPermission(): Promise<boolean> {
      if (!('Notification' in window)) return false

      if (Notification.permission === 'granted') return true

      const permission = await Notification.requestPermission()
      return permission === 'granted'
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
      serialize: state => encryptData(state),
      deserialize: value => decryptData(value)
    }
  }
})
