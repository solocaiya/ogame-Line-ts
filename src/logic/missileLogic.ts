/**
 * 导弹系统逻辑
 * 处理星际导弹攻击、射程计算、拦截等
 */

import type { Planet, MissileAttack, DefenseType, TechnologyType, Position, BuildQueueItem } from '@/types/game'
import { DefenseType as DefenseTypes, BuildingType } from '@/types/game'

/**
 * 计算导弹发射井容量
 */
export const calculateMissileSiloCapacity = (buildings: Partial<Record<BuildingType, number>>): number => {
  const siloLevel = buildings[BuildingType.MissileSilo] || 0
  return siloLevel * 10 // 每级存储10枚导弹
}

/**
 * 计算当前导弹总数
 */
export const calculateCurrentMissileCount = (defense: Partial<Record<DefenseType, number>>): number => {
  const interplanetaryMissiles = defense[DefenseTypes.InterplanetaryMissile] || 0
  const antiBallisticMissiles = defense[DefenseTypes.AntiBallisticMissile] || 0
  return interplanetaryMissiles + antiBallisticMissiles
}

/**
 * 计算建造队列中的导弹总数
 */
export const calculateQueueMissileCount = (buildQueue: BuildQueueItem[]): number => {
  let queueMissileCount = 0

  for (const item of buildQueue) {
    if (item.type === 'defense') {
      const defenseType = item.itemType as DefenseType
      if (defenseType === DefenseTypes.InterplanetaryMissile || defenseType === DefenseTypes.AntiBallisticMissile) {
        queueMissileCount += item.quantity || 0
      }
    }
  }

  return queueMissileCount
}

/**
 * 检查导弹容量限制
 */
export const checkMissileSiloLimit = (
  defenseType: DefenseType,
  currentDefense: Partial<Record<DefenseType, number>>,
  buildings: Partial<Record<BuildingType, number>>,
  quantity: number,
  buildQueue?: BuildQueueItem[]
): boolean => {
  // 只对导弹类型进行检查
  if (defenseType !== DefenseTypes.InterplanetaryMissile && defenseType !== DefenseTypes.AntiBallisticMissile) {
    return true
  }

  const maxCapacity = calculateMissileSiloCapacity(buildings)
  const currentCount = calculateCurrentMissileCount(currentDefense)
  const queueCount = buildQueue ? calculateQueueMissileCount(buildQueue) : 0
  const newCount = currentCount + queueCount + quantity

  return newCount <= maxCapacity
}

/**
 * 计算导弹射程（基于脉冲引擎等级）
 * 射程 = 5 * impulseDriveLevel - 1（系统距离）
 */
export const calculateMissileRange = (impulseDriveLevel: number): number => {
  if (impulseDriveLevel === 0) return 0
  return 5 * impulseDriveLevel - 1
}

/**
 * 计算两个位置之间的系统距离
 */
export const calculateSystemDistance = (from: Position, to: Position): number => {
  // 如果在不同银河系，距离无限大
  if (from.galaxy !== to.galaxy) {
    return Infinity
  }

  // 同一银河系内的系统距离
  return Math.abs(from.system - to.system)
}

/**
 * 检查目标是否在射程内
 */
export const isTargetInRange = (originPosition: Position, targetPosition: Position, impulseDriveLevel: number): boolean => {
  const range = calculateMissileRange(impulseDriveLevel)
  const distance = calculateSystemDistance(originPosition, targetPosition)
  return distance <= range
}

/**
 * 计算导弹飞行时间（秒）
 * 基础飞行时间: 30秒 + 60秒/系统距离
 */
export const calculateMissileFlightTime = (distance: number): number => {
  return 30 + distance * 60
}

/**
 * 创建导弹攻击任务
 */
export const createMissileAttack = (
  playerId: string,
  originPlanet: Planet,
  targetPosition: Position,
  targetPlanetId: string | undefined,
  missileCount: number
): MissileAttack => {
  const now = Date.now()
  const distance = calculateSystemDistance(originPlanet.position, targetPosition)
  const flightTime = calculateMissileFlightTime(distance) * 1000 // 转换为毫秒

  return {
    id: `missile_${now}_${playerId}`,
    playerId,
    originPlanetId: originPlanet.id,
    targetPosition,
    targetPlanetId,
    missileCount,
    launchTime: now,
    arrivalTime: now + flightTime,
    status: 'flying'
  }
}

/**
 * 验证导弹发射条件
 */
export const validateMissileLaunch = (
  originPlanet: Planet,
  targetPosition: Position,
  missileCount: number,
  technologies: Partial<Record<TechnologyType, number>>
): {
  valid: boolean
  reason?: string
} => {
  // 检查是否有足够的星际导弹
  const availableMissiles = originPlanet.defense[DefenseTypes.InterplanetaryMissile] || 0
  if (availableMissiles < missileCount) {
    return { valid: false, reason: 'errors.insufficientMissiles' }
  }

  // 检查发射数量
  if (missileCount <= 0) {
    return { valid: false, reason: 'errors.invalidMissileCount' }
  }

  // 检查射程
  const impulseDriveLevel = technologies['impulseDrive'] || 0
  if (!isTargetInRange(originPlanet.position, targetPosition, impulseDriveLevel)) {
    return { valid: false, reason: 'errors.targetOutOfRange' }
  }

  // 不能攻击自己的星球
  if (
    originPlanet.position.galaxy === targetPosition.galaxy &&
    originPlanet.position.system === targetPosition.system &&
    originPlanet.position.position === targetPosition.position
  ) {
    return { valid: false, reason: 'errors.cannotAttackOwnPlanet' }
  }

  return { valid: true }
}

/**
 * 执行导弹发射（扣除导弹）
 */
export const executeMissileLaunch = (planet: Planet, missileCount: number): void => {
  const currentMissiles = planet.defense[DefenseTypes.InterplanetaryMissile] || 0
  planet.defense[DefenseTypes.InterplanetaryMissile] = currentMissiles - missileCount
}

/**
 * 计算导弹攻击结果（考虑拦截）
 * @returns 实际命中的导弹数量
 */
export const calculateMissileImpact = (
  attackingMissiles: number,
  defenderPlanet: Planet
): {
  missileHits: number
  missileIntercepted: number
  defenseLosses: Partial<Record<DefenseType, number>>
} => {
  const antiBallisticMissiles = defenderPlanet.defense[DefenseTypes.AntiBallisticMissile] || 0

  // 反弹道导弹拦截（1:1）
  const intercepted = Math.min(attackingMissiles, antiBallisticMissiles)
  const missileHits = attackingMissiles - intercepted

  // 计算防御损失
  const defenseLosses: Partial<Record<DefenseType, number>> = {}

  // 消耗的反弹道导弹
  if (intercepted > 0) {
    defenseLosses[DefenseTypes.AntiBallisticMissile] = intercepted
  }

  // 如果有导弹命中，随机摧毁防御设施
  if (missileHits > 0) {
    const defenseTypes = Object.keys(defenderPlanet.defense) as DefenseType[]
    const availableDefenses = defenseTypes.filter(type => {
      // 不能摧毁护盾罩和行星护盾
      if (type === DefenseTypes.SmallShieldDome || type === DefenseTypes.LargeShieldDome || type === DefenseTypes.PlanetaryShield) {
        return false
      }
      return (defenderPlanet.defense[type] || 0) > 0
    })

    // 每枚导弹可以摧毁一个防御设施
    for (let i = 0; i < missileHits && availableDefenses.length > 0; i++) {
      const randomIndex = Math.floor(Math.random() * availableDefenses.length)
      const targetDefense = availableDefenses[randomIndex]

      if (targetDefense) {
        if (!defenseLosses[targetDefense]) {
          defenseLosses[targetDefense] = 0
        }
        defenseLosses[targetDefense]!++

        // 如果该类型防御全部摧毁，从可用列表中移除
        const remaining = (defenderPlanet.defense[targetDefense] || 0) - (defenseLosses[targetDefense] || 0)
        if (remaining <= 0) {
          availableDefenses.splice(randomIndex, 1)
        }
      }
    }
  }

  return {
    missileHits,
    missileIntercepted: intercepted,
    defenseLosses
  }
}

/**
 * 应用导弹攻击结果到星球
 */
export const applyMissileAttackResult = (planet: Planet, defenseLosses: Partial<Record<DefenseType, number>>): void => {
  for (const [defenseType, lossCount] of Object.entries(defenseLosses)) {
    const currentCount = planet.defense[defenseType as DefenseType] || 0
    planet.defense[defenseType as DefenseType] = Math.max(0, currentCount - lossCount)
  }
}
