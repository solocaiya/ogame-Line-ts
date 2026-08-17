// 数据统计逻辑

import type { Player, Planet } from '@/types/game'
import * as resourceLogic from './resourceLogic'
import * as officerLogic from './officerLogic'

export interface GameStatistics {
  // 基本信息
  totalPlayTime: number // 总游戏时间（秒）
  accountCreated: number // 账号创建时间

  // 资源统计
  totalMetalProduced: number
  totalCrystalProduced: number
  totalDeuteriumProduced: number
  totalDarkMatterEarned: number

  // 建筑统计
  totalBuildingsBuilt: number
  totalBuildingLevel: number

  // 科技统计
  totalResearchCompleted: number
  totalResearchLevel: number

  // 舰队统计
  totalShipsBuilt: number
  totalFleetDispatches: number
  totalBattlesWon: number
  totalBattlesLost: number

  // 交易统计
  totalTrades: number
  totalDarkMatterSpent: number

  // 签到统计
  totalCheckInDays: number
  currentStreak: number

  // 星球统计
  totalPlanets: number
  totalMoons: number

  // 排行榜
  totalScore: number
}

/**
 * 计算游戏统计数据
 */
export const calculateStatistics = (player: Player, planets: Planet[]): GameStatistics => {
  // 建筑统计
  let totalBuildingsBuilt = 0
  let totalBuildingLevel = 0
  const buildingKeys = [
    'metalMine', 'crystalMine', 'deuteriumSynthesizer', 'solarPlant', 'fusionPlant',
    'researchLab', 'robotFactory', 'nanoFactory', 'shipyard', 'metalStorage',
    'crystalStorage', 'deuteriumTank', 'missileSilo', 'terraformer', 'spaceDock'
  ]

  // 科技统计
  let totalResearchCompleted = 0
  let totalResearchLevel = 0
  const researchKeys = [
    'energy', 'laser', 'ion', 'hyperspace', 'plasma', 'combustionDrive',
    'impulseDrive', 'hyperspaceDrive', 'espionage', 'computer', 'astrophysics',
    'intergalacticResearchNetwork', 'graviton', 'weapons', 'shielding', 'armour'
  ]

  // 遍历所有星球统计建筑
  for (const planet of planets) {
    for (const key of buildingKeys) {
      const level = (planet.buildings as Record<string, number>)?.[key] ?? 0
      if (level > 0) {
        totalBuildingsBuilt += level
        totalBuildingLevel += level
      }
    }
  }

  // 统计科技
  for (const key of researchKeys) {
    const level = (player.research as Record<string, number>)?.[key] ?? 0
    if (level > 0) {
      totalResearchCompleted += level
      totalResearchLevel += level
    }
  }

  // 舰队统计
  let totalShipsBuilt = 0
  const shipKeys = [
    'lightFighter', 'heavyFighter', 'cruiser', 'battleship', 'smallCargo',
    'largeCargo', 'colonyShip', 'recycler', 'espionageProbe', 'bomber',
    'destroyer', 'deathstar', 'battlecruiser'
  ]
  for (const planet of planets) {
    for (const key of shipKeys) {
      totalShipsBuilt += (planet.fleet as Record<string, number>)?.[key] ?? 0
    }
  }

  // 交易统计
  const tradeHistory = player.tradeHistory || []
  const totalTrades = tradeHistory.length
  const totalDarkMatterSpent = tradeHistory.reduce((sum, r) => sum + r.darkMatterSpent, 0)

  // 星球统计
  const totalPlanets = planets.filter(p => !p.isMoon).length
  const totalMoons = planets.filter(p => p.isMoon).length

  // 签到统计
  const checkInData = player.checkInData
  const totalCheckInDays = checkInData?.checkedDays?.length ?? 0
  const currentStreak = checkInData?.currentStreak ?? 0

  // 总游戏时间
  const totalPlayTime = player.totalPlayTime ?? 0
  const accountCreated = player.createdAt ?? Date.now()

  // 资源产出估算：用当前每分钟产量 × 账号存在时长（分钟）
  // player.statistics 从未被后端填充，所以从当前产量反推累计产出
  const accountAgeMinutes = Math.max(1, Math.floor((Date.now() - accountCreated) / 60000))

  // 从所有星球的当前资源产量汇总（每分钟）
  // 使用 resourceLogic.calculateResourceProduction 计算每颗星球的每小时产量
  let totalMetalPerMin = 0
  let totalCrystalPerMin = 0
  let totalDeuteriumPerMin = 0
  const now = Date.now()
  const bonuses = officerLogic.calculateActiveBonuses(player.officers, now)
  for (const planet of planets) {
    const hourlyProd = resourceLogic.calculateResourceProduction(planet, {
      resourceProductionBonus: bonuses.resourceProductionBonus,
      darkMatterProductionBonus: bonuses.darkMatterProductionBonus,
      energyProductionBonus: bonuses.energyProductionBonus
    })
    totalMetalPerMin += (hourlyProd.metal ?? 0) / 60
    totalCrystalPerMin += (hourlyProd.crystal ?? 0) / 60
    totalDeuteriumPerMin += (hourlyProd.deuterium ?? 0) / 60
  }

  // 估算累计产出 = 当前产量/分钟 × 账号时长（取较大值，避免新号全0）
  const estimatedMetalProduced = Math.max(
    player.statistics?.totalMetalProduced ?? 0,
    Math.floor(totalMetalPerMin * accountAgeMinutes)
  )
  const estimatedCrystalProduced = Math.max(
    player.statistics?.totalCrystalProduced ?? 0,
    Math.floor(totalCrystalPerMin * accountAgeMinutes)
  )
  const estimatedDeuteriumProduced = Math.max(
    player.statistics?.totalDeuteriumProduced ?? 0,
    Math.floor(totalDeuteriumPerMin * accountAgeMinutes)
  )

  // 暗物质估算：从签到天数 + 暗物质收集器等级粗略估算
  const estimatedDarkMatter = Math.max(
    player.statistics?.totalDarkMatterEarned ?? 0,
    totalCheckInDays * 50 // 签到每天约50暗物质
  )

  // 舰队派出/战斗统计：无历史数据源，保持为0（后续可通过事件追踪补充）
  const totalFleetDispatches = player.statistics?.totalFleetDispatches ?? 0
  const totalBattlesWon = player.statistics?.totalBattlesWon ?? 0
  const totalBattlesLost = player.statistics?.totalBattlesLost ?? 0

  // 计算总分
  const totalScore = calculateScore(player, planets, totalBuildingLevel, totalResearchLevel, totalShipsBuilt, estimatedDarkMatter)

  return {
    totalPlayTime,
    accountCreated,
    totalMetalProduced: estimatedMetalProduced,
    totalCrystalProduced: estimatedCrystalProduced,
    totalDeuteriumProduced: estimatedDeuteriumProduced,
    totalDarkMatterEarned: estimatedDarkMatter,
    totalBuildingsBuilt,
    totalBuildingLevel,
    totalResearchCompleted,
    totalResearchLevel,
    totalShipsBuilt,
    totalFleetDispatches,
    totalBattlesWon,
    totalBattlesLost,
    totalTrades,
    totalDarkMatterSpent,
    totalCheckInDays,
    currentStreak,
    totalPlanets,
    totalMoons,
    totalScore
  }
}

/**
 * 计算总分
 */
const calculateScore = (
  player: Player,
  planets: Planet[],
  totalBuildingLevel: number,
  totalResearchLevel: number,
  totalShipsBuilt: number,
  estimatedDarkMatter: number = 0
): number => {
  // 建筑分：每级10分
  const buildingScore = totalBuildingLevel * 10
  // 科技分：每级20分
  const researchScore = totalResearchLevel * 20
  // 舰队分
  const fleetScore = totalShipsBuilt * 5
  // 暗物质分（使用估算值，因为 player.statistics 从未被后端填充）
  const darkMatterScore = estimatedDarkMatter * 2
  // 签到分
  const checkInScore = (player.checkInData?.checkedDays?.length ?? 0) * 50

  return buildingScore + researchScore + fleetScore + darkMatterScore + checkInScore
}

/**
 * 格式化游戏时间
 */
export const formatPlayTime = (seconds: number): string => {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)

  if (days > 0) return `${days}天 ${hours}小时`
  if (hours > 0) return `${hours}小时 ${minutes}分钟`
  return `${minutes}分钟`
}
