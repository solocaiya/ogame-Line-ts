// 数据统计逻辑

import type { Player, Planet } from '@/types/game'

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
      totalShipsBuilt += (planet.ships as Record<string, number>)?.[key] ?? 0
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

  // 计算总分
  const totalScore = calculateScore(player, planets, totalBuildingLevel, totalResearchLevel, totalShipsBuilt)

  return {
    totalPlayTime,
    accountCreated,
    totalMetalProduced: player.statistics?.totalMetalProduced ?? 0,
    totalCrystalProduced: player.statistics?.totalCrystalProduced ?? 0,
    totalDeuteriumProduced: player.statistics?.totalDeuteriumProduced ?? 0,
    totalDarkMatterEarned: player.statistics?.totalDarkMatterEarned ?? 0,
    totalBuildingsBuilt,
    totalBuildingLevel,
    totalResearchCompleted,
    totalResearchLevel,
    totalShipsBuilt,
    totalFleetDispatches: player.statistics?.totalFleetDispatches ?? 0,
    totalBattlesWon: player.statistics?.totalBattlesWon ?? 0,
    totalBattlesLost: player.statistics?.totalBattlesLost ?? 0,
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
  totalShipsBuilt: number
): number => {
  // 建筑分：每级10分
  const buildingScore = totalBuildingLevel * 10
  // 科技分：每级20分
  const researchScore = totalResearchLevel * 20
  // 舰队分
  const fleetScore = totalShipsBuilt * 5
  // 暗物质分
  const darkMatterScore = (player.statistics?.totalDarkMatterEarned ?? 0) * 2
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
