import { computed } from 'vue'
import { useGameStore } from '@/stores/gameStore'
import type { BattleResult, Resources } from '@/types/game'

/**
 * 战斗统计 composable
 * 从 gameStore.player.battleReports 计算统计数据
 */
export function useBattleStats() {
  const gameStore = useGameStore()

  const reports = computed(() => gameStore.player.battleReports || [])

  const totalBattles = computed(() => reports.value.length)

  const wins = computed(() => reports.value.filter(r => r.winner === 'attacker').length)
  const losses = computed(() => reports.value.filter(r => r.winner === 'defender').length)
  const draws = computed(() => reports.value.filter(r => r.winner === 'draw').length)

  const winRate = computed(() => {
    if (totalBattles.value === 0) return 0
    return Math.round((wins.value / totalBattles.value) * 100)
  })

  const totalPlunder = computed<Resources>(() => {
    const result: Resources = { metal: 0, crystal: 0, deuterium: 0, darkMatter: 0, energy: 0 }
    for (const r of reports.value) {
      result.metal += r.plunder?.metal || 0
      result.crystal += r.plunder?.crystal || 0
      result.deuterium += r.plunder?.deuterium || 0
    }
    return result
  })

  const totalAttackerLosses = computed(() => {
    let total = 0
    for (const r of reports.value) {
      total += calculateFleetLossValue(r.attackerLosses)
    }
    return total
  })

  const totalDefenderLosses = computed(() => {
    let total = 0
    for (const r of reports.value) {
      total += calculateFleetLossValue(r.defenderLosses.fleet)
      total += calculateDefenseLossValue(r.defenderLosses.defense)
    }
    return total
  })

  /**
   * 按结果筛选
   */
  const filterByResult = (result: 'all' | 'win' | 'loss' | 'draw') => {
    switch (result) {
      case 'win':
        return reports.value.filter(r => r.winner === 'attacker')
      case 'loss':
        return reports.value.filter(r => r.winner === 'defender')
      case 'draw':
        return reports.value.filter(r => r.winner === 'draw')
      default:
        return reports.value
    }
  }

  /**
   * 按时间范围筛选
   */
  const filterByTimeRange = (range: 'all' | 'today' | 'week' | 'month') => {
    const now = Date.now()
    const msPerDay = 86400000
    let cutoff = 0
    switch (range) {
      case 'today':
        cutoff = now - msPerDay
        break
      case 'week':
        cutoff = now - msPerDay * 7
        break
      case 'month':
        cutoff = now - msPerDay * 30
        break
    }
    return reports.value.filter(r => r.timestamp >= cutoff)
  }

  /**
   * 组合筛选
   */
  const getFilteredReports = (
    result: 'all' | 'win' | 'loss' | 'draw',
    timeRange: 'all' | 'today' | 'week' | 'month'
  ): BattleResult[] => {
    let filtered = filterByResult(result)
    const now = Date.now()
    const msPerDay = 86400000
    if (timeRange !== 'all') {
      const cutoff = timeRange === 'today' ? now - msPerDay
        : timeRange === 'week' ? now - msPerDay * 7
        : now - msPerDay * 30
      filtered = filtered.filter(r => r.timestamp >= cutoff)
    }
    return filtered
  }

  return {
    reports,
    totalBattles,
    wins,
    losses,
    draws,
    winRate,
    totalPlunder,
    totalAttackerLosses,
    totalDefenderLosses,
    filterByResult,
    filterByTimeRange,
    getFilteredReports
  }
}

// 简化损失计算（用于统计，不需要精确到每个舰船的配置）
const SHIP_BASE_VALUES: Record<string, number> = {
  lightFighter: 3000, heavyFighter: 6000, cruiser: 20000,
  battleship: 45000, smallCargo: 4000, largeCargo: 12000,
  colonyShip: 10000, recycler: 10000, espionageProbe: 1000,
  bomber: 75000, destroyer: 60000, deathstar: 5000000, battlecruiser: 30000,
  interceptor: 30000, heavyInterceptor: 55000, reaper: 85000, pathfinder: 12000
}

const DEFENSE_BASE_VALUES: Record<string, number> = {
  rocketLauncher: 2000, laserLight: 1500, laserHeavy: 6000,
  ionCannon: 8000, gaussCannon: 20000, plasmaTurret: 50000,
  shieldDomeSmall: 10000, shieldDomeLarge: 50000,
  antiBallisticMissile: 8000, interplanetaryMissile: 12500
}

function calculateFleetLossValue(losses: Partial<Record<string, number>>): number {
  let total = 0
  for (const [type, count] of Object.entries(losses)) {
    total += (SHIP_BASE_VALUES[type] || 5000) * (count || 0)
  }
  return total
}

function calculateDefenseLossValue(losses: Partial<Record<string, number>>): number {
  let total = 0
  for (const [type, count] of Object.entries(losses)) {
    total += (DEFENSE_BASE_VALUES[type] || 3000) * (count || 0)
  }
  return total
}
