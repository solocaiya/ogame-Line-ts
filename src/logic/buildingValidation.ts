import type { Planet, Resources, BuildQueueItem, Officer } from '@/types/game'
import { BuildingType, TechnologyType, OfficerType } from '@/types/game'
import * as buildingLogic from './buildingLogic'
import * as resourceLogic from './resourceLogic'
import * as publicLogic from './publicLogic'
import * as officerLogic from './officerLogic'
import { BUILDINGS } from '@/config/gameConfig'

/**
 * 验证建筑升级的所有条件
 */
export const validateBuildingUpgrade = (
  planet: Planet,
  buildingType: BuildingType,
  technologies: Partial<Record<TechnologyType, number>>,
  officers: Record<OfficerType, Officer>
): {
  valid: boolean
  reason?: string
} => {
  const currentLevel = planet.buildings[buildingType] || 0
  const targetLevel = currentLevel + 1
  const cost = buildingLogic.calculateBuildingCost(buildingType, targetLevel)
  const buildingConfig = BUILDINGS[buildingType]

  // 检查队列中是否已存在该建筑的升级或拆除任务
  const existingQueueItem = planet.buildQueue.find(
    item => (item.type === 'building' || item.type === 'demolish') && item.itemType === buildingType
  )
  if (existingQueueItem) {
    return { valid: false, reason: 'errors.buildingAlreadyInQueue' }
  }

  // 检查星球/月球限制
  if (buildingConfig.planetOnly && planet.isMoon) {
    return { valid: false, reason: 'errors.planetOnly' }
  }
  if (buildingConfig.moonOnly && !planet.isMoon) {
    return { valid: false, reason: 'errors.moonOnly' }
  }

  // 计算军官加成
  const bonuses = officerLogic.calculateActiveBonuses(officers, Date.now())

  // 检查建造队列是否已满（只计算建筑类型的队列项）
  const maxQueue = publicLogic.getMaxBuildQueue(planet, bonuses.additionalBuildQueue)
  const buildingQueueCount = planet.buildQueue.filter(item => item.type === 'building' || item.type === 'demolish').length
  if (buildingQueueCount >= maxQueue) {
    return { valid: false, reason: 'errors.buildQueueFull' }
  }

  // 检查空间
  if (!buildingLogic.checkSpaceAvailable(planet, buildingType)) {
    return { valid: false, reason: 'errors.insufficientSpace' }
  }

  // 检查资源
  if (!resourceLogic.checkResourcesAvailable(planet.resources, cost)) {
    return { valid: false, reason: 'errors.insufficientResources' }
  }

  // 检查前置条件
  if (!buildingLogic.checkBuildingRequirements(buildingType, planet, technologies)) {
    return { valid: false, reason: 'errors.requirementsNotMet' }
  }

  return { valid: true }
}

/**
 * 执行建筑升级（扣除资源，添加到队列）
 */
export const executeBuildingUpgrade = (
  planet: Planet,
  buildingType: BuildingType,
  officers: Record<OfficerType, Officer>
): BuildQueueItem => {
  const currentLevel = planet.buildings[buildingType] || 0
  const targetLevel = currentLevel + 1
  const cost = buildingLogic.calculateBuildingCost(buildingType, targetLevel)

  // 计算军官加成
  const bonuses = officerLogic.calculateActiveBonuses(officers, Date.now())

  // 获取机器人工厂和纳米工厂等级
  const roboticsFactoryLevel = planet.buildings[BuildingType.RoboticsFactory] || 0
  const naniteFactoryLevel = planet.buildings[BuildingType.NaniteFactory] || 0

  const time = buildingLogic.calculateBuildingTime(
    buildingType,
    targetLevel,
    bonuses.buildingSpeedBonus,
    roboticsFactoryLevel,
    naniteFactoryLevel
  )

  // 扣除资源
  resourceLogic.deductResources(planet.resources, cost)

  // 创建队列项
  return buildingLogic.createBuildQueueItem(buildingType, targetLevel, time)
}

/**
 * 取消建造并计算返还资源
 */
export const cancelBuildingUpgrade = (_planet: Planet, queueItem: BuildQueueItem): Resources => {
  const cost = buildingLogic.calculateBuildingCost(queueItem.itemType as BuildingType, queueItem.targetLevel || 1)

  return {
    metal: Math.floor(cost.metal * 0.5),
    crystal: Math.floor(cost.crystal * 0.5),
    deuterium: Math.floor(cost.deuterium * 0.5),
    darkMatter: Math.floor(cost.darkMatter * 0.5),
    energy: 0
  }
}

/**
 * 查找队列项
 */
export const findQueueItem = (
  queue: BuildQueueItem[],
  queueId: string
): {
  item: BuildQueueItem | null
  index: number
} => {
  const index = queue.findIndex(q => q.id === queueId)
  if (index === -1) {
    return { item: null, index: -1 }
  }
  return { item: queue[index] || null, index }
}

/**
 * 验证建筑拆除的所有条件
 */
export const validateBuildingDemolish = (
  planet: Planet,
  buildingType: BuildingType,
  officers: Record<OfficerType, Officer>
): {
  valid: boolean
  reason?: string
} => {
  const currentLevel = planet.buildings[buildingType] || 0

  // 检查建筑等级
  if (currentLevel <= 0) {
    return { valid: false, reason: 'errors.buildingLevelZero' }
  }

  // 计算军官加成
  const bonuses = officerLogic.calculateActiveBonuses(officers, Date.now())

  // 检查建造队列是否已满（只计算建筑类型的队列项）
  const maxQueue = publicLogic.getMaxBuildQueue(planet, bonuses.additionalBuildQueue)
  const buildingQueueCount = planet.buildQueue.filter(item => item.type === 'building' || item.type === 'demolish').length
  if (buildingQueueCount >= maxQueue) {
    return { valid: false, reason: 'errors.buildQueueFull' }
  }

  return { valid: true }
}

/**
 * 执行建筑拆除（返还资源，添加到队列）
 */
export const executeBuildingDemolish = (
  planet: Planet,
  buildingType: BuildingType,
  officers: Record<OfficerType, Officer>
): BuildQueueItem => {
  const currentLevel = planet.buildings[buildingType] || 0

  // 计算军官加成
  const bonuses = officerLogic.calculateActiveBonuses(officers, Date.now())

  // 获取机器人工厂和纳米工厂等级
  const roboticsFactoryLevel = planet.buildings[BuildingType.RoboticsFactory] || 0
  const naniteFactoryLevel = planet.buildings[BuildingType.NaniteFactory] || 0

  const demolishTime = buildingLogic.calculateDemolishTime(
    buildingType,
    currentLevel,
    bonuses.buildingSpeedBonus,
    roboticsFactoryLevel,
    naniteFactoryLevel
  )

  // 返还50%资源
  const refund = buildingLogic.calculateDemolishRefund(buildingType, currentLevel)
  planet.resources.metal += refund.metal
  planet.resources.crystal += refund.crystal
  planet.resources.deuterium += refund.deuterium
  planet.resources.darkMatter += refund.darkMatter

  // 创建拆除队列项
  return buildingLogic.createDemolishQueueItem(buildingType, currentLevel, demolishTime)
}
