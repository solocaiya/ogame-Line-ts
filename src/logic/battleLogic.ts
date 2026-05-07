import type { Fleet, Resources, BattleResult, Officer, TechnologyType } from '@/types/game'
import { DefenseType, OfficerType } from '@/types/game'
import { workerManager } from '@/workers/workerManager'
import { MOON_CONFIG } from '@/config/gameConfig'
import { generateId } from '@/utils/id'

/**
 * 执行战斗模拟
 * 使用 Web Worker 在后台线程中执行计算密集型的战斗模拟
 */
export const simulateBattle = async (
  attackerFleet: Partial<Fleet>,
  defenderFleet: Partial<Fleet>,
  defenderDefense: Partial<Record<DefenseType, number>>,
  defenderResources: Resources,
  _attackerOfficers: Record<OfficerType, Officer>,
  _defenderOfficers: Record<OfficerType, Officer>,
  attackerTechnologies: Record<TechnologyType, number>,
  defenderTechnologies: Record<TechnologyType, number>,
  battleToFinish: boolean = false // 战斗到底模式
): Promise<BattleResult> => {
  // 从科技系统读取实际科技等级
  const attackerWeaponTech = attackerTechnologies['weaponsTechnology'] || 0
  const attackerShieldTech = attackerTechnologies['shieldingTechnology'] || 0
  const attackerArmorTech = attackerTechnologies['armourTechnology'] || 0

  const defenderWeaponTech = defenderTechnologies['weaponsTechnology'] || 0
  const defenderShieldTech = defenderTechnologies['shieldingTechnology'] || 0
  const defenderArmorTech = defenderTechnologies['armourTechnology'] || 0

  // 使用 Worker 执行战斗模拟
  const simulationResult = await workerManager.simulateBattle({
    attacker: {
      ships: attackerFleet,
      weaponTech: attackerWeaponTech,
      shieldTech: attackerShieldTech,
      armorTech: attackerArmorTech
    },
    defender: {
      ships: defenderFleet,
      defense: defenderDefense,
      weaponTech: defenderWeaponTech,
      shieldTech: defenderShieldTech,
      armorTech: defenderArmorTech
    },
    maxRounds: battleToFinish ? 100 : 6 // 战斗到底模式最多100回合，经典模式6回合
  })

  // 计算掠夺（仅攻击方胜利时）
  const plunder =
    simulationResult.winner === 'attacker'
      ? await workerManager.calculatePlunder({
          defenderResources,
          attackerFleet: simulationResult.attackerRemaining
        })
      : { metal: 0, crystal: 0, deuterium: 0, darkMatter: 0, energy: 0 }

  // 计算残骸场
  const debrisField = await workerManager.calculateDebris({
    attackerLosses: simulationResult.attackerLosses,
    defenderLosses: simulationResult.defenderLosses
  })

  // 计算月球生成概率（根据残骸场总量）
  const totalDebris = debrisField.metal + debrisField.crystal
  const moonChance = Math.min(MOON_CONFIG.baseChance + Math.floor(totalDebris / MOON_CONFIG.chancePerDebris), MOON_CONFIG.maxChance) / 100 // 转换为0-1的概率

  // 生成战斗报告
  const battleResult: BattleResult = {
    id: generateId('battle'),
    timestamp: Date.now(),
    attackerId: '',
    defenderId: '',
    attackerPlanetId: '',
    defenderPlanetId: '',
    attackerFleet,
    defenderFleet,
    defenderDefense,
    attackerLosses: simulationResult.attackerLosses,
    defenderLosses: simulationResult.defenderLosses,
    winner: simulationResult.winner,
    plunder,
    debrisField,
    // 新增详细信息
    rounds: simulationResult.rounds,
    attackerRemaining: simulationResult.attackerRemaining,
    defenderRemaining: simulationResult.defenderRemaining,
    roundDetails: simulationResult.roundDetails,
    moonChance
  }

  return battleResult
}

/**
 * 计算防御设施修复（防御有70%概率修复）
 */
export const repairDefense = (
  defenseBeforeBattle: Partial<Record<DefenseType, number>>,
  defenseAfterBattle: Partial<Record<DefenseType, number>>
): Partial<Record<DefenseType, number>> => {
  const repaired: Partial<Record<DefenseType, number>> = { ...defenseAfterBattle }

  Object.keys(defenseBeforeBattle).forEach(defenseType => {
    const before = defenseBeforeBattle[defenseType as DefenseType] || 0
    const after = defenseAfterBattle[defenseType as DefenseType] || 0
    const lost = before - after

    if (lost > 0) {
      // 70%修复概率
      const repairedCount = Math.floor(lost * 0.7)
      repaired[defenseType as DefenseType] = after + repairedCount
    }
  })

  return repaired
}
