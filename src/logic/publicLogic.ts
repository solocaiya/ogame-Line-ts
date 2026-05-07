/**
 * 公共业务逻辑模块
 * 提供跨模块共享的通用业务逻辑功能
 */

import { BuildingType, TechnologyType, ShipType, DefenseType } from '@/types/game'
import type { Planet, Resources, Officer, BuildingConfig, TechnologyConfig, Player } from '@/types/game'
import { OfficerType } from '@/types/game'
import * as officerLogic from '@/logic/officerLogic'
import * as resourceLogic from '@/logic/resourceLogic'
import { scaleResources } from '@/utils/speed'
import { BUILDINGS, TECHNOLOGIES, SHIPS, DEFENSES } from '@/config/gameConfig'

/**
 * 获取特定等级的升级条件
 * 合并基础 requirements 和等级门槛 levelRequirements
 * @param config 建筑或科技配置
 * @param targetLevel 目标等级
 * @returns 合并后的前置条件
 */
export const getLevelRequirements = (
  config: BuildingConfig | TechnologyConfig,
  targetLevel: number
): Partial<Record<BuildingType | TechnologyType, number>> => {
  const requirements: Partial<Record<BuildingType | TechnologyType, number>> = {}

  // 1. 添加基础 requirements（如果存在）
  if (config.requirements) {
    Object.assign(requirements, config.requirements)
  }

  // 2. 添加等级门槛 requirements（如果存在）
  if (config.levelRequirements) {
    // 找出所有小于等于目标等级的门槛
    const applicableLevels = Object.keys(config.levelRequirements)
      .map(Number)
      .filter(level => level <= targetLevel)
      .sort((a, b) => a - b)

    // 依次合并所有适用的等级要求（后面的覆盖前面的）
    for (const level of applicableLevels) {
      const levelReqs = config.levelRequirements[level]
      if (levelReqs) {
        // 合并要求，取最大值
        for (const [key, value] of Object.entries(levelReqs)) {
          const currentValue = requirements[key as BuildingType | TechnologyType] || 0
          requirements[key as BuildingType | TechnologyType] = Math.max(currentValue, value)
        }
      }
    }
  }

  return requirements
}

/**
 * 检查建造/研发前置条件是否满足
 * @param planet 星球对象
 * @param technologies 已研究的科技等级
 * @param requirements 前置条件要求（建筑等级或科技等级）
 * @returns 是否满足前置条件
 */
export const checkRequirements = (
  planet: Planet | undefined,
  technologies: Partial<Record<TechnologyType, number>>,
  requirements?: Partial<Record<BuildingType | TechnologyType, number>>
): boolean => {
  // 如果星球不存在或没有前置条件，默认返回 true
  if (!planet || !requirements) return true

  // 检查所有前置条件
  for (const [key, requiredLevel] of Object.entries(requirements)) {
    // 检查是否为建筑类型
    if (Object.values(BuildingType).includes(key as BuildingType)) {
      const requiredBuildingType = key as BuildingType
      const requiredBuildingConfig = BUILDINGS[requiredBuildingType]

      // 如果当前是月球，且所需建筑是星球专属建筑（planetOnly），则跳过此前置条件
      // 这允许在月球上建造机器人工厂等建筑，即使它们的前置条件是无法在月球建造的矿场
      if (planet.isMoon && requiredBuildingConfig?.planetOnly) {
        continue
      }

      const currentLevel = planet.buildings[requiredBuildingType] || 0
      if (currentLevel < requiredLevel) {
        return false
      }
    }
    // 检查是否为科技类型
    else if (Object.values(TechnologyType).includes(key as TechnologyType)) {
      const currentLevel = technologies[key as TechnologyType] || 0
      if (currentLevel < requiredLevel) {
        return false
      }
    }
  }

  return true
}

/**
 * 计算星球的资源产量（包含军官加成和科技加成）
 * @param planet 星球对象
 * @param officers 玩家的军官对象
 * @param resourceSpeed 游戏速度
 * @param techBonuses 科技加成（可选，矿物研究/晶体研究/燃料研究）
 * @returns 每小时各类资源的产量
 */
export const getResourceProduction = (
  planet: Planet,
  officers: Record<OfficerType, Officer>,
  resourceSpeed: number = 1,
  techBonuses?: {
    mineralResearchLevel?: number
    crystalResearchLevel?: number
    fuelResearchLevel?: number
  }
): Resources => {
  // 计算当前激活的军官加成
  const bonuses = officerLogic.calculateActiveBonuses(officers, Date.now())
  // 根据建筑等级和军官加成计算资源产量
  const base = resourceLogic.calculateResourceProduction(planet, bonuses, techBonuses)
  return scaleResources(base, resourceSpeed)
}

/**
 * 计算星球的资源存储容量（包含军官加成）
 * @param planet 星球对象
 * @param officers 玩家的军官对象
 * @returns 各类资源的最大存储容量
 */
export const getResourceCapacity = (planet: Planet, officers: Record<OfficerType, Officer>): Resources => {
  // 计算当前激活的军官加成
  const bonuses = officerLogic.calculateActiveBonuses(officers, Date.now())
  // 根据仓库建筑等级和军官加成计算存储容量
  return resourceLogic.calculateResourceCapacity(planet, bonuses.storageCapacityBonus)
}

/**
 * 计算最大建造队列数量
 * @param planet 星球对象
 * @param additionalBuildQueue 军官提供的额外队列数量
 * @returns 最大建造队列数量（基础3个 + 纳米工厂等级 + 军官加成，最多10个）
 */
export const getMaxBuildQueue = (planet: Planet, additionalBuildQueue: number = 0): number => {
  const naniteFactoryLevel = planet.buildings[BuildingType.NaniteFactory] || 0
  return Math.min(3 + naniteFactoryLevel + additionalBuildQueue, 10)
}

/**
 * 计算最大研究队列数量
 * @param technologies 已研究的科技等级
 * @returns 最大研究队列数量（基础3个 + 计算机技术等级，最多10个）
 */
export const getMaxResearchQueue = (technologies: Partial<Record<TechnologyType, number>>): number => {
  const computerTechLevel = technologies[TechnologyType.ComputerTechnology] || 0
  return Math.min(3 + computerTechLevel, 10)
}

/**
 * 计算最大舰队任务数量
 * @param additionalFleetSlots 军官提供的额外槽位数量
 * @param computerTechnologyLevel 计算机技术等级
 * @returns 最大舰队任务数量（基础3个 + 计算机技术等级 + 军官加成，最多20个）
 */
export const getMaxFleetMissions = (additionalFleetSlots: number = 0, computerTechnologyLevel: number = 0): number => {
  return Math.min(3 + computerTechnologyLevel + additionalFleetSlots, 20)
}

/**
 * 计算建筑的总成本（从等级1到目标等级的累计成本）
 * @param buildingType 建筑类型
 * @param level 目标等级
 * @returns 总资源成本（金属+水晶+重氢）
 */
export const calculateBuildingTotalCost = (buildingType: BuildingType, level: number): number => {
  if (level <= 0) return 0

  const config = BUILDINGS[buildingType]
  if (!config) return 0

  let totalCost = 0
  const { baseCost, costMultiplier } = config

  // 累加从等级1到目标等级的所有成本
  for (let i = 1; i <= level; i++) {
    const levelCost = {
      metal: Math.floor(baseCost.metal * Math.pow(costMultiplier, i - 1)),
      crystal: Math.floor(baseCost.crystal * Math.pow(costMultiplier, i - 1)),
      deuterium: Math.floor(baseCost.deuterium * Math.pow(costMultiplier, i - 1))
    }
    totalCost += levelCost.metal + levelCost.crystal + levelCost.deuterium
  }

  return totalCost
}

/**
 * 计算科技的总成本（从等级1到目标等级的累计成本）
 * @param techType 科技类型
 * @param level 目标等级
 * @returns 总资源成本（金属+水晶+重氢）
 */
export const calculateTechnologyTotalCost = (techType: TechnologyType, level: number): number => {
  if (level <= 0) return 0

  const config = TECHNOLOGIES[techType]
  if (!config) return 0

  let totalCost = 0
  const { baseCost, costMultiplier } = config

  // 累加从等级1到目标等级的所有成本
  for (let i = 1; i <= level; i++) {
    const levelCost = {
      metal: Math.floor(baseCost.metal * Math.pow(costMultiplier, i - 1)),
      crystal: Math.floor(baseCost.crystal * Math.pow(costMultiplier, i - 1)),
      deuterium: Math.floor(baseCost.deuterium * Math.pow(costMultiplier, i - 1))
    }
    totalCost += levelCost.metal + levelCost.crystal + levelCost.deuterium
  }

  return totalCost
}

/**
 * 计算单个舰船的成本
 * @param shipType 舰船类型
 * @returns 单个舰船的资源成本（金属+水晶+重氢）
 */
export const calculateShipUnitCost = (shipType: ShipType): number => {
  const config = SHIPS[shipType]
  if (!config) return 0

  return config.cost.metal + config.cost.crystal + config.cost.deuterium
}

/**
 * 计算单个防御的成本
 * @param defenseType 防御类型
 * @returns 单个防御的资源成本（金属+水晶+重氢）
 */
export const calculateDefenseUnitCost = (defenseType: DefenseType): number => {
  const config = DEFENSES[defenseType]
  if (!config) return 0

  return config.cost.metal + config.cost.crystal + config.cost.deuterium
}

/**
 * 计算玩家的总积分
 * 积分规则：(建筑成本 + 科技成本 + 舰队成本 + 防御成本) / 1000
 * @param player 玩家对象
 * @returns 玩家总积分
 */
export const calculatePlayerPoints = (player: Player): number => {
  let totalCost = 0

  // 1. 计算所有星球的建筑成本
  player.planets.forEach(planet => {
    Object.entries(planet.buildings).forEach(([buildingType, level]) => {
      totalCost += calculateBuildingTotalCost(buildingType as BuildingType, level)
    })
  })

  // 2. 计算科技成本
  Object.entries(player.technologies).forEach(([techType, level]) => {
    totalCost += calculateTechnologyTotalCost(techType as TechnologyType, level)
  })

  // 3. 计算所有星球的舰队成本
  player.planets.forEach(planet => {
    Object.entries(planet.fleet).forEach(([shipType, count]) => {
      totalCost += calculateShipUnitCost(shipType as ShipType) * count
    })
  })

  // 4. 计算所有星球的防御成本
  player.planets.forEach(planet => {
    Object.entries(planet.defense).forEach(([defenseType, count]) => {
      totalCost += calculateDefenseUnitCost(defenseType as DefenseType) * count
    })
  })

  // 每1000资源 = 1积分（基础积分）
  const basePoints = Math.floor(totalCost / 1000)

  // 加上奖励积分（战役、成就等奖励的积分）
  const bonusPoints = player.bonusPoints || 0

  return basePoints + bonusPoints
}
