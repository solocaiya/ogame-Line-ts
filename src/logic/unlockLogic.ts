/**
 * 解锁检查逻辑
 * 用于检测建筑/科技完成后新解锁的内容
 */

import { BuildingType, TechnologyType, type Planet } from '@/types/game'
import { BUILDINGS, TECHNOLOGIES } from '@/config/gameConfig'
import * as buildingLogic from './buildingLogic'
import * as researchLogic from './researchLogic'

export interface UnlockedItem {
  type: 'building' | 'technology'
  id: BuildingType | TechnologyType
  name: string
}

/**
 * 检查哪些建筑在完成建筑/科技后被新解锁
 */
export const checkNewlyUnlockedBuildings = (
  planet: Planet,
  technologies: Partial<Record<TechnologyType, number>>,
  previousBuildings: Partial<Record<BuildingType, number>>,
  previousTechnologies: Partial<Record<TechnologyType, number>>
): UnlockedItem[] => {
  const newlyUnlocked: UnlockedItem[] = []

  // 遍历所有建筑类型
  for (const buildingType of Object.values(BuildingType)) {
    const config = BUILDINGS[buildingType]
    if (!config) continue

    // 跳过已经建造过的建筑（等级 > 0）
    if ((planet.buildings[buildingType] || 0) > 0) continue

    // 检查之前是否已解锁
    const wasUnlockedBefore = checkBuildingRequirementsWith(buildingType, previousBuildings, previousTechnologies)

    // 检查现在是否解锁
    const isUnlockedNow = buildingLogic.checkBuildingRequirements(buildingType, planet, technologies)

    // 如果之前未解锁，现在解锁了，则是新解锁
    if (!wasUnlockedBefore && isUnlockedNow) {
      newlyUnlocked.push({
        type: 'building',
        id: buildingType,
        name: config.name
      })
    }
  }

  return newlyUnlocked
}

/**
 * 检查哪些科技在完成建筑/科技后被新解锁
 */
export const checkNewlyUnlockedTechnologies = (
  buildings: Partial<Record<BuildingType, number>>,
  technologies: Partial<Record<TechnologyType, number>>,
  previousBuildings: Partial<Record<BuildingType, number>>,
  previousTechnologies: Partial<Record<TechnologyType, number>>
): UnlockedItem[] => {
  const newlyUnlocked: UnlockedItem[] = []

  // 遍历所有科技类型
  for (const techType of Object.values(TechnologyType)) {
    const config = TECHNOLOGIES[techType]
    if (!config) continue

    // 跳过已经研究过的科技（等级 > 0）
    if ((technologies[techType] || 0) > 0) continue

    // 检查之前是否已解锁
    const wasUnlockedBefore = researchLogic.checkTechnologyRequirements(techType, previousBuildings, previousTechnologies)

    // 检查现在是否解锁
    const isUnlockedNow = researchLogic.checkTechnologyRequirements(techType, buildings, technologies)

    // 如果之前未解锁，现在解锁了，则是新解锁
    if (!wasUnlockedBefore && isUnlockedNow) {
      newlyUnlocked.push({
        type: 'technology',
        id: techType,
        name: config.name
      })
    }
  }

  return newlyUnlocked
}

/**
 * 使用指定的建筑和科技等级检查建筑需求
 * （用于与之前状态比较）
 */
const checkBuildingRequirementsWith = (
  buildingType: BuildingType,
  buildings: Partial<Record<BuildingType, number>>,
  technologies: Partial<Record<TechnologyType, number>>
): boolean => {
  const config = BUILDINGS[buildingType]
  const requirements = (config as any).requirements
  if (!requirements) return true

  for (const [key, level] of Object.entries(requirements)) {
    const requiredLevel = level as number
    if (Object.values(BuildingType).includes(key as BuildingType)) {
      if ((buildings[key as BuildingType] || 0) < requiredLevel) {
        return false
      }
    } else if (Object.values(TechnologyType).includes(key as TechnologyType)) {
      if ((technologies[key as TechnologyType] || 0) < requiredLevel) {
        return false
      }
    }
  }
  return true
}

/**
 * 检查所有新解锁的内容（建筑 + 科技）
 */
export const checkAllNewlyUnlocked = (
  planet: Planet,
  technologies: Partial<Record<TechnologyType, number>>,
  previousBuildings: Partial<Record<BuildingType, number>>,
  previousTechnologies: Partial<Record<TechnologyType, number>>
): UnlockedItem[] => {
  const unlockedBuildings = checkNewlyUnlockedBuildings(planet, technologies, previousBuildings, previousTechnologies)

  const unlockedTechnologies = checkNewlyUnlockedTechnologies(planet.buildings, technologies, previousBuildings, previousTechnologies)

  return [...unlockedBuildings, ...unlockedTechnologies]
}
