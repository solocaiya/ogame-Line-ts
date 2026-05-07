import type { Planet, Resources, BuildQueueItem, Officer } from '@/types/game'
import { TechnologyType, OfficerType, BuildingType } from '@/types/game'
import * as researchLogic from './researchLogic'
import * as resourceLogic from './resourceLogic'
import * as publicLogic from './publicLogic'
import * as officerLogic from './officerLogic'

/**
 * 验证科技研究的所有条件
 */
export const validateTechnologyResearch = (
  planet: Planet,
  techType: TechnologyType,
  technologies: Partial<Record<TechnologyType, number>>,
  researchQueue: BuildQueueItem[]
): {
  valid: boolean
  reason?: string
} => {
  const currentLevel = technologies[techType] || 0
  const targetLevel = currentLevel + 1
  const cost = researchLogic.calculateTechnologyCost(techType, targetLevel)

  // 检查队列中是否已存在该科技的研究任务
  const existingQueueItem = researchQueue.find(item => item.type === 'technology' && item.itemType === techType)
  if (existingQueueItem) {
    return { valid: false, reason: 'errors.technologyAlreadyInQueue' }
  }

  // 检查研究队列是否已满
  const maxQueue = publicLogic.getMaxResearchQueue(technologies)
  if (researchQueue.length >= maxQueue) {
    return { valid: false, reason: 'errors.researchQueueFull' }
  }

  // 检查前置条件
  if (!researchLogic.checkTechnologyRequirements(techType, planet.buildings, technologies)) {
    return { valid: false, reason: 'errors.requirementsNotMet' }
  }

  // 检查资源
  if (!resourceLogic.checkResourcesAvailable(planet.resources, cost)) {
    return { valid: false, reason: 'errors.insufficientResources' }
  }

  return { valid: true }
}

/**
 * 执行科技研究（扣除资源，创建队列项）
 */
export const executeTechnologyResearch = (
  planet: Planet,
  techType: TechnologyType,
  currentLevel: number,
  officers: Record<OfficerType, Officer>,
  technologies: Partial<Record<TechnologyType, number>>,
  allPlanets?: Planet[]
): { queueItem: BuildQueueItem } => {
  const targetLevel = currentLevel + 1
  const cost = researchLogic.calculateTechnologyCost(techType, targetLevel)

  // 计算军官加成
  const bonuses = officerLogic.calculateActiveBonuses(officers, Date.now())

  // 获取星际研究网络等级
  const intergalacticResearchNetworkLevel = technologies[TechnologyType.IntergalacticResearchNetwork] || 0

  // 计算有效研究实验室等级（考虑星际研究网络）
  let researchLabLevel: number
  if (allPlanets && intergalacticResearchNetworkLevel > 0) {
    researchLabLevel = researchLogic.calculateEffectiveLabLevel(allPlanets, planet.id, intergalacticResearchNetworkLevel)
  } else {
    researchLabLevel = planet.buildings[BuildingType.ResearchLab] || 1
  }

  const energyTechLevel = technologies[TechnologyType.EnergyTechnology] || 0
  // 获取大学等级（加速研究）
  const universityLevel = planet.buildings[BuildingType.University] || 0

  const time = researchLogic.calculateTechnologyTime(
    techType,
    currentLevel,
    bonuses.researchSpeedBonus,
    researchLabLevel,
    energyTechLevel,
    1,
    universityLevel
  )

  // 扣除资源
  resourceLogic.deductResources(planet.resources, cost)

  // 创建队列项
  const queueItem = researchLogic.createResearchQueueItem(techType, targetLevel, time)

  return { queueItem }
}

/**
 * 取消研究并计算返还资源
 */
export const cancelTechnologyResearch = (queueItem: BuildQueueItem): Resources => {
  const cost = researchLogic.calculateTechnologyCost(queueItem.itemType as TechnologyType, queueItem.targetLevel || 1)

  return {
    metal: Math.floor(cost.metal * 0.5),
    crystal: Math.floor(cost.crystal * 0.5),
    deuterium: Math.floor(cost.deuterium * 0.5),
    darkMatter: Math.floor(cost.darkMatter * 0.5),
    energy: 0
  }
}
