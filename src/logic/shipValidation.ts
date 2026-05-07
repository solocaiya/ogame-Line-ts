import type { Planet, Resources, BuildQueueItem, Fleet, Officer } from '@/types/game'
import { ShipType, DefenseType, TechnologyType, OfficerType, BuildingType } from '@/types/game'
import * as shipLogic from './shipLogic'
import * as missileLogic from './missileLogic'
import * as resourceLogic from './resourceLogic'
import * as officerLogic from './officerLogic'
import * as publicLogic from './publicLogic'
import * as fleetStorageLogic from './fleetStorageLogic'

/**
 * 验证舰船建造的所有条件
 */
export const validateShipBuild = (
  planet: Planet,
  shipType: ShipType,
  quantity: number,
  technologies: Partial<Record<TechnologyType, number>>
): {
  valid: boolean
  reason?: string
} => {
  const totalCost = shipLogic.calculateShipCost(shipType, quantity)

  // 检查前置条件
  if (!shipLogic.checkShipRequirements(shipType, planet.buildings, technologies)) {
    return { valid: false, reason: 'errors.requirementsNotMet' }
  }

  // 检查资源
  if (!resourceLogic.checkResourcesAvailable(planet.resources, totalCost)) {
    return { valid: false, reason: 'errors.insufficientResources' }
  }

  // 检查舰队仓储空间
  if (!fleetStorageLogic.hasEnoughFleetStorage(planet, shipType, quantity, technologies as Record<TechnologyType, number>)) {
    return { valid: false, reason: 'errors.insufficientFleetStorage' }
  }

  return { valid: true }
}

/**
 * 执行舰船建造
 */
export const executeShipBuild = (
  planet: Planet,
  shipType: ShipType,
  quantity: number,
  officers: Record<OfficerType, Officer>
): BuildQueueItem => {
  const totalCost = shipLogic.calculateShipCost(shipType, quantity)

  // 计算军官加成
  const bonuses = officerLogic.calculateActiveBonuses(officers, Date.now())

  // 获取机器人工厂和纳米工厂等级
  const roboticsFactoryLevel = planet.buildings[BuildingType.RoboticsFactory] || 0
  const naniteFactoryLevel = planet.buildings[BuildingType.NaniteFactory] || 0

  const buildTime = shipLogic.calculateShipBuildTime(
    shipType,
    quantity,
    bonuses.buildingSpeedBonus,
    roboticsFactoryLevel,
    naniteFactoryLevel
  )

  // 扣除资源
  resourceLogic.deductResources(planet.resources, totalCost)

  // 创建队列项
  return shipLogic.createShipQueueItem(shipType, quantity, buildTime)
}

/**
 * 验证防御建造的所有条件
 */
export const validateDefenseBuild = (
  planet: Planet,
  defenseType: DefenseType,
  quantity: number,
  technologies: Partial<Record<TechnologyType, number>>
): {
  valid: boolean
  reason?: string
} => {
  const totalCost = shipLogic.calculateDefenseCost(defenseType, quantity)

  // 检查前置条件
  if (!shipLogic.checkDefenseRequirements(defenseType, planet.buildings, technologies)) {
    return { valid: false, reason: 'errors.requirementsNotMet' }
  }

  // 检查资源
  if (!resourceLogic.checkResourcesAvailable(planet.resources, totalCost)) {
    return { valid: false, reason: 'errors.insufficientResources' }
  }

  // 护盾罩限制
  if (!shipLogic.checkShieldDomeLimit(defenseType, planet.defense, quantity)) {
    return { valid: false, reason: 'errors.shieldDomeLimit' }
  }

  // 导弹发射井容量限制
  if (!missileLogic.checkMissileSiloLimit(defenseType, planet.defense, planet.buildings, quantity, planet.buildQueue)) {
    return { valid: false, reason: 'errors.missileSiloLimit' }
  }

  return { valid: true }
}

/**
 * 执行防御建造
 */
export const executeDefenseBuild = (
  planet: Planet,
  defenseType: DefenseType,
  quantity: number,
  officers: Record<OfficerType, Officer>
): BuildQueueItem => {
  const totalCost = shipLogic.calculateDefenseCost(defenseType, quantity)

  // 计算军官加成
  const bonuses = officerLogic.calculateActiveBonuses(officers, Date.now())

  // 获取机器人工厂和纳米工厂等级
  const roboticsFactoryLevel = planet.buildings[BuildingType.RoboticsFactory] || 0
  const naniteFactoryLevel = planet.buildings[BuildingType.NaniteFactory] || 0

  const buildTime = shipLogic.calculateDefenseBuildTime(
    defenseType,
    quantity,
    bonuses.buildingSpeedBonus,
    roboticsFactoryLevel,
    naniteFactoryLevel
  )

  // 扣除资源
  resourceLogic.deductResources(planet.resources, totalCost)

  // 创建队列项
  return shipLogic.createDefenseQueueItem(defenseType, quantity, buildTime)
}

/**
 * 验证舰队派遣的所有条件
 */
export const validateFleetDispatch = (
  planet: Planet,
  fleet: Partial<Fleet>,
  cargo: Resources,
  officers: Record<OfficerType, Officer>,
  currentFleetMissions: number = 0,
  technologies: Partial<Record<TechnologyType, number>> = {}
): {
  valid: boolean
  reason?: string
  fuelNeeded?: number
} => {
  // 计算军官加成
  const bonuses = officerLogic.calculateActiveBonuses(officers, Date.now())

  // 检查舰队任务槽位是否已满
  const computerTechLevel = technologies[TechnologyType.ComputerTechnology] || 0
  const maxFleetMissions = publicLogic.getMaxFleetMissions(bonuses.additionalFleetSlots, computerTechLevel)
  if (currentFleetMissions >= maxFleetMissions) {
    return { valid: false, reason: 'errors.fleetMissionsFull' }
  }

  // 检查舰队是否足够
  if (!shipLogic.checkFleetAvailable(planet.fleet, fleet)) {
    return { valid: false, reason: 'errors.insufficientFleet' }
  }

  // 检查是否有足够的重氢作为燃料（包含货物重量影响）
  const fuelNeeded = shipLogic.calculateFleetFuelConsumption(fleet, bonuses.fuelConsumptionReduction, cargo)
  if (planet.resources.deuterium < fuelNeeded) {
    return { valid: false, reason: 'errors.insufficientFuel', fuelNeeded }
  }

  return { valid: true, fuelNeeded }
}

/**
 * 执行舰队派遣（扣除舰队和燃料）
 */
export const executeFleetDispatch = (
  planet: Planet,
  fleet: Partial<Fleet>,
  fuelNeeded: number,
  shouldDeductCargo: boolean,
  cargo: Resources
): void => {
  // 扣除舰队
  shipLogic.deductFleet(planet.fleet, fleet)

  // 扣除燃料
  planet.resources.deuterium -= fuelNeeded

  // 扣除运输的资源
  if (shouldDeductCargo) {
    resourceLogic.deductResources(planet.resources, cargo)
  }
}

/**
 * 验证舰船拆除的所有条件
 */
export const validateShipScrap = (
  planet: Planet,
  shipType: ShipType,
  quantity: number
): {
  valid: boolean
  reason?: string
} => {
  // 检查数量是否有效
  if (quantity <= 0) {
    return { valid: false, reason: 'errors.invalidQuantity' }
  }

  // 检查舰船数量是否足够
  const available = planet.fleet[shipType] || 0
  if (available < quantity) {
    return { valid: false, reason: 'errors.insufficientShips' }
  }

  return { valid: true }
}

/**
 * 执行舰船拆除
 */
export const executeShipScrap = (
  planet: Planet,
  shipType: ShipType,
  quantity: number,
  officers: Record<OfficerType, Officer>
): BuildQueueItem => {
  // 计算军官加成
  const bonuses = officerLogic.calculateActiveBonuses(officers, Date.now())

  // 获取机器人工厂和纳米工厂等级
  const roboticsFactoryLevel = planet.buildings[BuildingType.RoboticsFactory] || 0
  const naniteFactoryLevel = planet.buildings[BuildingType.NaniteFactory] || 0

  // 计算拆除时间
  const scrapTime = shipLogic.calculateShipScrapTime(
    shipType,
    quantity,
    bonuses.buildingSpeedBonus,
    roboticsFactoryLevel,
    naniteFactoryLevel
  )

  // 计算返还资源
  const refund = shipLogic.calculateShipScrapRefund(shipType, quantity)

  // 扣除舰船（立即扣除，避免拆除过程中派遣）
  planet.fleet[shipType] = (planet.fleet[shipType] || 0) - quantity

  // 返还资源
  resourceLogic.addResources(planet.resources, refund)

  // 创建队列项
  return shipLogic.createShipScrapQueueItem(shipType, quantity, scrapTime)
}
