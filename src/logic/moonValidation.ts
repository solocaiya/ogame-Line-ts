import type { Planet, Resources } from '@/types/game'
import * as planetLogic from './planetLogic'

/**
 * 检查是否可以生成月球
 */
export const canCreateMoon = (
  planets: Planet[],
  position: { galaxy: number; system: number; position: number },
  debrisField: Resources
): {
  canCreate: boolean
  reason?: string
  chance?: number
} => {
  // 检查该位置是否已有月球
  const existingMoon = planets.find(
    p =>
      p.position.galaxy === position.galaxy &&
      p.position.system === position.system &&
      p.position.position === position.position &&
      p.isMoon
  )

  if (existingMoon) {
    return { canCreate: false, reason: 'errors.moonExists' }
  }

  const chance = planetLogic.calculateMoonChance(debrisField)
  if (chance === 0) {
    return { canCreate: false, reason: 'errors.insufficientDebris', chance }
  }

  return { canCreate: true, chance }
}

/**
 * 计算月球生成概率并判断是否生成
 */
export const shouldGenerateMoon = (chance: number): boolean => {
  const random = Math.random() * 100
  return random <= chance
}

/**
 * 查找母星
 */
export const findParentPlanet = (planets: Planet[], position: { galaxy: number; system: number; position: number }): Planet | null => {
  return (
    planets.find(
      p =>
        p.position.galaxy === position.galaxy &&
        p.position.system === position.system &&
        p.position.position === position.position &&
        !p.isMoon
    ) || null
  )
}
