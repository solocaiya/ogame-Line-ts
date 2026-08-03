import { computed } from 'vue'
import { useGameStore } from '@/stores/gameStore'
import type { Planet, WaitingQueueItem, BuildQueueItem, Resources } from '@/types/game'

interface PlanetQueueSummary {
  planetId: string
  planetName: string
  isMoon: boolean
  activeCount: number
  waitingCount: number
  queuePaused: boolean
  estimatedCompletionTime: number // timestamp
  items: {
    active: BuildQueueItem[]
    waiting: WaitingQueueItem[]
  }
}

/**
 * 队列汇总 composable
 * 提供跨星球统一队列统计
 */
export function useQueueSummary() {
  const gameStore = useGameStore()

  const planets = computed(() => gameStore.player.planets || [])

  /**
   * 每个星球的队列状态摘要
   */
  const planetSummaries = computed<PlanetQueueSummary[]>(() => {
    return planets.value.map(planet => {
      const active = planet.buildQueue || []
      const waiting = planet.waitingBuildQueue || []
      const lastActive = active.length > 0 ? Math.max(...active.map(i => i.endTime)) : Date.now()
      return {
        planetId: planet.id,
        planetName: planet.name,
        isMoon: planet.isMoon,
        activeCount: active.length,
        waitingCount: waiting.length,
        queuePaused: planet.queuePaused || false,
        estimatedCompletionTime: lastActive,
        items: { active, waiting }
      }
    })
  })

  /**
   * 所有队列的总活跃项目数
   */
  const totalActiveItems = computed(() =>
    planetSummaries.value.reduce((sum, p) => sum + p.activeCount, 0)
  )

  /**
   * 所有队列的总等待项目数
   */
  const totalWaitingItems = computed(() =>
    planetSummaries.value.reduce((sum, p) => sum + p.waitingCount, 0)
  )

  /**
   * 全局最晚完成时间
   */
  const globalEstimatedCompletion = computed(() => {
    const times = planetSummaries.value.map(p => p.estimatedCompletionTime)
    return times.length > 0 ? Math.max(...times) : Date.now()
  })

  /**
   * 按类型分类统计（所有星球合计）
   */
  const categoryStats = computed(() => {
    const stats: Record<string, { active: number; waiting: number }> = {
      building: { active: 0, waiting: 0 },
      ship: { active: 0, waiting: 0 },
      defense: { active: 0, waiting: 0 },
      technology: { active: 0, waiting: 0 },
      demolish: { active: 0, waiting: 0 }
    }
    for (const planet of planets.value) {
      for (const item of planet.buildQueue || []) {
        const key = item.type === 'scrap_ship' ? 'ship' : item.type
        if (stats[key]) stats[key].active++
      }
      for (const item of planet.waitingBuildQueue || []) {
        const key = item.type === 'scrap_ship' ? 'ship' : item.type
        if (stats[key]) stats[key].waiting++
      }
    }
    // 研究队列
    const player = gameStore.player
    for (const item of player.researchQueue || []) {
      stats.technology.active++
    }
    for (const item of player.waitingResearchQueue || []) {
      stats.technology.waiting++
    }
    return stats
  })

  /**
   * 估算等待队列总资源需求
   */
  const waitingQueueResourceEstimate = computed<Resources>(() => {
    const total: Resources = { metal: 0, crystal: 0, deuterium: 0, darkMatter: 0, energy: 0 }
    // 简化估算：遍历所有等待队列项
    for (const planet of planets.value) {
      for (const _item of planet.waitingBuildQueue || []) {
        // 实际成本需要调用 calculateWaitingItemCost，这里用占位
        // 视图层可以调用 waitingQueueLogic.calculateWaitingItemCost 获取精确值
      }
    }
    return total
  })

  /**
   * 获取指定星球的队列信息
   */
  const getPlanetQueue = (planetId: string): PlanetQueueSummary | undefined => {
    return planetSummaries.value.find(p => p.planetId === planetId)
  }

  return {
    planetSummaries,
    totalActiveItems,
    totalWaitingItems,
    globalEstimatedCompletion,
    categoryStats,
    waitingQueueResourceEstimate,
    getPlanetQueue
  }
}
