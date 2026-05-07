import type { Player, Resources, BuildingType, TechnologyType, ShipType, DefenseType } from '@/types/game'
import { SHIPS } from '@/config/gameConfig'
import { DEFENSES } from '@/config/gameConfig'
import * as buildingLogic from './buildingLogic'
import * as researchLogic from './researchLogic'

/**
 * 计算资源总和（仅计算金属+晶体+重氢）
 * 根据游戏规则：不包括暗物质和能量
 */
export const calculateResourceCost = (resources: Resources): number => {
  return resources.metal + resources.crystal + resources.deuterium
}

/**
 * 将资源总和转换为积分
 * 规则：每1000资源 = 1分
 */
export const calculatePointsFromResources = (resourceCost: number): number => {
  return Math.floor(resourceCost / 1000)
}

/**
 * 为玩家添加积分
 */
export const addPoints = (player: Player, points: number): void => {
  player.points += points
}

/**
 * 计算建筑升级到指定等级所获得的积分
 * @param buildingType 建筑类型
 * @param fromLevel 起始等级（不包括）
 * @param toLevel 目标等级（包括）
 * @returns 积分数
 */
export const calculateBuildingPoints = (buildingType: BuildingType, fromLevel: number, toLevel: number): number => {
  let totalPoints = 0
  for (let level = fromLevel + 1; level <= toLevel; level++) {
    const cost = buildingLogic.calculateBuildingCost(buildingType, level)
    const resourceCost = calculateResourceCost(cost)
    totalPoints += calculatePointsFromResources(resourceCost)
  }
  return totalPoints
}

/**
 * 计算科技研究到指定等级所获得的积分
 * @param technologyType 科技类型
 * @param fromLevel 起始等级（不包括）
 * @param toLevel 目标等级（包括）
 * @returns 积分数
 */
export const calculateTechnologyPoints = (technologyType: TechnologyType, fromLevel: number, toLevel: number): number => {
  let totalPoints = 0
  for (let level = fromLevel + 1; level <= toLevel; level++) {
    const cost = researchLogic.calculateTechnologyCost(technologyType, level)
    const resourceCost = calculateResourceCost(cost)
    totalPoints += calculatePointsFromResources(resourceCost)
  }
  return totalPoints
}

/**
 * 计算舰船建造所获得的积分
 * @param shipType 舰船类型
 * @param quantity 数量
 * @returns 积分数
 */
export const calculateShipPoints = (shipType: ShipType, quantity: number): number => {
  const config = SHIPS[shipType]
  const resourceCost = calculateResourceCost(config.cost)
  const pointsPerShip = calculatePointsFromResources(resourceCost)
  return pointsPerShip * quantity
}

/**
 * 计算防御建造所获得的积分
 * @param defenseType 防御类型
 * @param quantity 数量
 * @returns 积分数
 */
export const calculateDefensePoints = (defenseType: DefenseType, quantity: number): number => {
  const config = DEFENSES[defenseType]
  const resourceCost = calculateResourceCost(config.cost)
  const pointsPerDefense = calculatePointsFromResources(resourceCost)
  return pointsPerDefense * quantity
}

/**
 * 计算玩家当前的总积分（用于数据迁移或重新计算）
 * 会计算所有建筑、科技、舰船、防御的累积积分
 */
export const calculateTotalPlayerPoints = (player: Player): number => {
  let totalPoints = 0

  // 计算所有星球的建筑积分
  for (const planet of player.planets) {
    for (const [buildingType, level] of Object.entries(planet.buildings)) {
      if (level > 0) {
        totalPoints += calculateBuildingPoints(buildingType as BuildingType, 0, level)
      }
    }

    // 计算所有星球的舰船积分
    for (const [shipType, quantity] of Object.entries(planet.fleet)) {
      if (quantity > 0) {
        totalPoints += calculateShipPoints(shipType as ShipType, quantity)
      }
    }

    // 计算所有星球的防御积分
    for (const [defenseType, quantity] of Object.entries(planet.defense)) {
      if (quantity > 0) {
        totalPoints += calculateDefensePoints(defenseType as DefenseType, quantity)
      }
    }
  }

  // 计算所有科技的积分
  for (const [technologyType, level] of Object.entries(player.technologies)) {
    if (level > 0) {
      totalPoints += calculateTechnologyPoints(technologyType as TechnologyType, 0, level)
    }
  }

  // 计算正在飞行的舰队积分（舰队已经建造，所以也计入积分）
  for (const mission of player.fleetMissions) {
    for (const [shipType, quantity] of Object.entries(mission.fleet)) {
      if (quantity && quantity > 0) {
        totalPoints += calculateShipPoints(shipType as ShipType, quantity)
      }
    }
  }

  return totalPoints
}
