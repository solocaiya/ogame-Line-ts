/**
 * 舰队仓储逻辑模块
 * 处理舰队仓储容量计算和使用量统计
 */

import type { Planet, Fleet } from '@/types/game'
import { ShipType, BuildingType, TechnologyType } from '@/types/game'
import { SHIPS, FLEET_STORAGE_CONFIG, BUILDINGS, TECHNOLOGIES } from '@/config/gameConfig'

/**
 * 计算舰队当前使用的仓储量
 * @param fleet 舰队对象
 * @returns 当前使用的仓储量
 */
export const calculateFleetStorageUsage = (fleet: Fleet): number => {
  let totalUsage = 0

  for (const shipType of Object.values(ShipType)) {
    const shipCount = fleet[shipType] || 0
    const shipConfig = SHIPS[shipType]
    totalUsage += shipCount * shipConfig.storageUsage
  }

  return totalUsage
}

/**
 * 计算星球的最大舰队仓储容量
 * @param planet 星球对象
 * @param technologies 玩家的科技等级
 * @returns 最大舰队仓储容量
 */
export const calculateMaxFleetStorage = (planet: Planet, technologies: Record<TechnologyType, number>): number => {
  // 1. 基础仓储
  let maxStorage = FLEET_STORAGE_CONFIG.baseStorage

  // 2. 造船厂建筑加成（每个星球独立）
  const shipyardLevel = planet.buildings[BuildingType.Shipyard] || 0
  const shipyardBonus = BUILDINGS[BuildingType.Shipyard].fleetStorageBonus || 0
  maxStorage += shipyardLevel * shipyardBonus

  // 3. 机库建筑加成（每个星球独立）
  const hangarLevel = planet.buildings[BuildingType.Hangar] || 0
  const hangarBonus = BUILDINGS[BuildingType.Hangar].fleetStorageBonus || 0
  maxStorage += hangarLevel * hangarBonus

  // 4. 计算机技术全局加成
  const computerTechLevel = technologies[TechnologyType.ComputerTechnology] || 0
  const computerTechBonus = TECHNOLOGIES[TechnologyType.ComputerTechnology].fleetStorageBonus || 0
  maxStorage += computerTechLevel * computerTechBonus

  return maxStorage
}

/**
 * 计算建造队列中的舰船仓储使用量
 * @param buildQueue 建造队列
 * @returns 队列中舰船的仓储使用量
 */
export const calculateQueueFleetStorageUsage = (buildQueue: Array<{ type: string; itemType: string; quantity?: number }>): number => {
  let queueUsage = 0

  for (const item of buildQueue) {
    if (item.type === 'ship') {
      const shipType = item.itemType as ShipType
      const quantity = item.quantity || 0
      const shipConfig = SHIPS[shipType]
      queueUsage += quantity * shipConfig.storageUsage
    }
  }

  return queueUsage
}

/**
 * 检查是否有足够的舰队仓储空间建造新舰船
 * @param planet 星球对象
 * @param shipType 要建造的舰船类型
 * @param quantity 要建造的数量
 * @param technologies 玩家的科技等级
 * @returns 是否有足够的空间
 */
export const hasEnoughFleetStorage = (
  planet: Planet,
  shipType: ShipType,
  quantity: number,
  technologies: Record<TechnologyType, number>
): boolean => {
  const currentUsage = calculateFleetStorageUsage(planet.fleet)
  const queueUsage = calculateQueueFleetStorageUsage(planet.buildQueue)
  const maxStorage = calculateMaxFleetStorage(planet, technologies)
  const newShipUsage = SHIPS[shipType].storageUsage * quantity

  return currentUsage + queueUsage + newShipUsage <= maxStorage
}

/**
 * 计算当前可以建造的最大舰船数量（基于仓储限制）
 * @param planet 星球对象
 * @param shipType 要建造的舰船类型
 * @param technologies 玩家的科技等级
 * @returns 最大可建造数量
 */
export const getMaxBuildableShips = (planet: Planet, shipType: ShipType, technologies: Record<TechnologyType, number>): number => {
  const currentUsage = calculateFleetStorageUsage(planet.fleet)
  const maxStorage = calculateMaxFleetStorage(planet, technologies)
  const availableStorage = maxStorage - currentUsage
  const shipStorageUsage = SHIPS[shipType].storageUsage

  if (shipStorageUsage === 0) return Number.MAX_SAFE_INTEGER

  // 如果当前已经超限（舰队返回等情况），则不允许建造新舰船
  if (availableStorage <= 0) return 0

  return Math.floor(availableStorage / shipStorageUsage)
}

/**
 * 安全地添加舰船到星球（会检查舰队仓储容量上限）
 * @param planet 星球对象
 * @param fleet 要添加的舰船
 * @param technologies 玩家的科技等级
 * @returns 实际添加的舰船数量和溢出的舰船数量
 */
export const addFleetSafely = (
  planet: Planet,
  fleet: Partial<Record<ShipType, number>>,
  technologies: Record<TechnologyType, number>
): { added: Partial<Record<ShipType, number>>; overflow: Partial<Record<ShipType, number>> } => {
  const maxStorage = calculateMaxFleetStorage(planet, technologies)
  let currentUsage = calculateFleetStorageUsage(planet.fleet)

  const added: Partial<Record<ShipType, number>> = {}
  const overflow: Partial<Record<ShipType, number>> = {}

  for (const [shipType, count] of Object.entries(fleet)) {
    if (count <= 0) continue

    const ship = shipType as ShipType
    const shipStorageUsage = SHIPS[ship].storageUsage

    // 计算可以添加多少艘（不超过容量上限）
    const spaceAvailable = Math.max(0, maxStorage - currentUsage)
    const maxCanAdd = shipStorageUsage > 0 ? Math.floor(spaceAvailable / shipStorageUsage) : count
    const actuallyAdded = Math.min(count, maxCanAdd)
    const overflowed = count - actuallyAdded
    if (actuallyAdded > 0) {
      planet.fleet[ship] = (planet.fleet[ship] || 0) + actuallyAdded
      added[ship] = actuallyAdded
      currentUsage += actuallyAdded * shipStorageUsage
    }

    if (overflowed > 0) {
      overflow[ship] = overflowed
    }
  }

  return { added, overflow }
}
