import type { Resources, BuildQueueItem, Planet } from '@/types/game'
import { TechnologyType, BuildingType } from '@/types/game'
import { TECHNOLOGIES } from '@/config/gameConfig'
import * as pointsLogic from './pointsLogic'

/**
 * 计算有效研究实验室等级（考虑星际研究网络）
 * 星际研究网络允许连接多个星球的研究实验室，取等级最高的N个实验室等级之和
 * @param planets 玩家的所有星球
 * @param currentPlanetId 当前星球ID（研究所在的星球）
 * @param intergalacticResearchNetworkLevel 星际研究网络等级
 * @returns 有效的研究实验室等级总和
 */
export const calculateEffectiveLabLevel = (
  planets: Planet[],
  currentPlanetId: string,
  intergalacticResearchNetworkLevel: number
): number => {
  // 收集所有星球的研究实验室等级
  const labLevels: { planetId: string; level: number }[] = []

  for (const planet of planets) {
    const labLevel = planet.buildings[BuildingType.ResearchLab] || 0
    if (labLevel > 0) {
      labLevels.push({ planetId: planet.id, level: labLevel })
    }
  }

  // 如果没有星际研究网络，只返回当前星球的实验室等级
  if (intergalacticResearchNetworkLevel === 0) {
    const currentPlanet = planets.find(p => p.id === currentPlanetId)
    return currentPlanet?.buildings[BuildingType.ResearchLab] || 0
  }

  // 按等级降序排序
  labLevels.sort((a, b) => b.level - a.level)

  // 可连接的实验室数量 = 1 + 星际研究网络等级
  // 等级1可连接2个实验室，等级2可连接3个，以此类推
  const maxLabs = 1 + intergalacticResearchNetworkLevel

  // 取前N个实验室的等级之和
  let totalLevel = 0
  const count = Math.min(maxLabs, labLevels.length)
  for (let i = 0; i < count; i++) {
    const lab = labLevels[i]
    if (lab) {
      totalLevel += lab.level
    }
  }

  return totalLevel
}

// 用于生成唯一ID的计数器
let researchQueueIdCounter = 0

/**
 * 计算科技研究成本
 */
export const calculateTechnologyCost = (techType: TechnologyType, targetLevel: number): Resources => {
  const config = TECHNOLOGIES[techType]
  const multiplier = Math.pow(config.costMultiplier, targetLevel - 1)
  return {
    metal: Math.floor(config.baseCost.metal * multiplier),
    crystal: Math.floor(config.baseCost.crystal * multiplier),
    deuterium: Math.floor(config.baseCost.deuterium * multiplier),
    darkMatter: Math.floor(config.baseCost.darkMatter * multiplier),
    energy: 0
  }
}

/**
 * 计算科技研究时间
 * 使用 2moons 公式（调整版）：
 * 1. 成本系数 = Σ (资源^0.3 / 0.003)
 * 2. 时间(秒) = 成本系数 / ((1 + 研究实验室等级) × 能源加成 × 游戏速度)
 * @param techType 科技类型
 * @param currentLevel 当前等级（要研究的目标等级 = currentLevel + 1）
 * @param researchSpeedBonus 军官等提供的研究速度加成百分比
 * @param researchLabLevel 研究实验室等级
 * @param energyTechLevel 能源技术等级（提供额外加速）
 * @param gameSpeed 游戏速度（默认1）
 * @param universityLevel 大学等级（每级减少研究时间8%）
 */
export const calculateTechnologyTime = (
  techType: TechnologyType,
  currentLevel: number,
  researchSpeedBonus: number = 0,
  researchLabLevel: number = 1,
  energyTechLevel: number = 0,
  gameSpeed: number = 1,
  universityLevel: number = 0
): number => {
  // 目标等级 = 当前等级 + 1
  const targetLevel = currentLevel + 1

  // 计算该等级的成本
  const cost = calculateTechnologyCost(techType, targetLevel)

  // 2moons 公式：成本系数 = Σ (资源^0.3 / 0.003)
  let elementCost = 0
  if (cost.metal > 0) elementCost += Math.pow(cost.metal, 0.3) / 0.003
  if (cost.crystal > 0) elementCost += Math.pow(cost.crystal, 0.3) / 0.003
  if (cost.deuterium > 0) elementCost += Math.pow(cost.deuterium, 0.3) / 0.003

  // 研究实验室等级至少为1，防止除以0
  const labLevel = Math.max(1, researchLabLevel)

  // 能源技术提供额外加速（每级5%）
  const energyBonus = 1 + energyTechLevel * 0.05

  // 简化公式：时间(秒) = 成本系数 / (实验室加成 × 能源加成 × 游戏速度)
  const timeInSeconds = elementCost / ((1 + labLevel) * energyBonus * gameSpeed)

  // 军官等的百分比加成
  const speedMultiplier = 1 - researchSpeedBonus / 100

  // 大学加成：每级减少研究时间8%（最高10级=57%减少，因为是乘法叠加）
  // 使用乘法叠加：(1 - 0.08)^level
  const universityMultiplier = Math.pow(0.92, universityLevel)

  // 确保最小时间为5秒
  return Math.max(5, Math.floor(timeInSeconds * speedMultiplier * universityMultiplier))
}

/**
 * 检查科技研究条件
 */
export const checkTechnologyRequirements = (
  techType: TechnologyType,
  buildings: Partial<Record<BuildingType, number>>,
  technologies: Partial<Record<TechnologyType, number>>
): boolean => {
  const config = TECHNOLOGIES[techType]
  if (!config.requirements) return true

  for (const [key, level] of Object.entries(config.requirements)) {
    if (Object.values(BuildingType).includes(key as BuildingType)) {
      if ((buildings[key as BuildingType] || 0) < level) {
        return false
      }
    } else if (Object.values(TechnologyType).includes(key as TechnologyType)) {
      if ((technologies[key as TechnologyType] || 0) < level) {
        return false
      }
    }
  }
  return true
}

/**
 * 创建研究队列项
 */
export const createResearchQueueItem = (techType: TechnologyType, targetLevel: number, researchTime: number): BuildQueueItem => {
  const now = Date.now()
  researchQueueIdCounter++
  return {
    id: `research_${now}_${researchQueueIdCounter}`,
    type: 'technology',
    itemType: techType,
    targetLevel,
    startTime: now,
    endTime: now + researchTime * 1000
  }
}

/**
 * 处理研究完成
 */
export const completeResearchQueue = (
  researchQueue: BuildQueueItem[],
  technologies: Partial<Record<TechnologyType, number>>,
  now: number,
  onPointsEarned?: (points: number, type: 'technology', itemType: string, level: number) => void,
  onCompleted?: (type: 'technology', itemType: string, level: number) => void
): BuildQueueItem[] => {
  return researchQueue.filter(item => {
    if (now >= item.endTime) {
      // 研究完成
      const oldLevel = technologies[item.itemType as TechnologyType] || 0
      // 研究完成时，等级+1（而不是直接使用targetLevel，保持一致性）
      const newLevel = oldLevel + 1
      technologies[item.itemType as TechnologyType] = newLevel

      // 计算并累积积分
      if (onPointsEarned && newLevel > oldLevel) {
        const points = pointsLogic.calculateTechnologyPoints(item.itemType as TechnologyType, oldLevel, newLevel)
        onPointsEarned(points, 'technology', item.itemType, newLevel)
      }

      // 通知完成
      if (onCompleted) {
        onCompleted('technology', item.itemType, newLevel)
      }

      return false
    }
    return true
  })
}
