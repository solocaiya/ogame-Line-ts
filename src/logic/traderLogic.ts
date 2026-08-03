// 资源交易所逻辑

import type { Player, Planet, Resources } from '@/types/game'

// 交易汇率配置（暗物质 → 基础资源）
export const TRADER_RATES = {
  // 1 暗物质 = X 金属
  metal: 100,
  // 1 暗物质 = X 晶体
  crystal: 60,
  // 1 暗物质 = X 重氢
  deuterium: 30
} as const

// 交易手续费比例（0.05 = 5%）
export const TRADER_FEE_RATE = 0.05

// 单次最大交易量（暗物质单位）
export const TRADER_MAX_DARK_MATTER = 10000

// 交易记录
export interface TradeRecord {
  id: string
  timestamp: number
  darkMatterSpent: number
  resourceType: 'metal' | 'crystal' | 'deuterium'
  resourceGained: number
  fee: number
}

/**
 * 计算交易：花费暗物质获得基础资源
 * @param darkMatterAmount 花费的暗物质数量
 * @param resourceType 目标资源类型
 * @returns 交易详情
 */
export const calculateTrade = (
  darkMatterAmount: number,
  resourceType: 'metal' | 'crystal' | 'deuterium'
): { resourceGained: number; fee: number; netResource: number } => {
  const rate = TRADER_RATES[resourceType]
  const grossResource = darkMatterAmount * rate
  const fee = Math.floor(grossResource * TRADER_FEE_RATE)
  const netResource = grossResource - fee
  return { resourceGained: grossResource, fee, netResource }
}

/**
 * 执行交易
 */
export const executeTrade = (
  player: Player,
  planet: Planet,
  darkMatterAmount: number,
  resourceType: 'metal' | 'crystal' | 'deuterium'
): { success: boolean; message: string; record?: TradeRecord } => {
  // 验证交易量
  if (darkMatterAmount <= 0 || darkMatterAmount > TRADER_MAX_DARK_MATTER) {
    return { success: false, message: '无效的交易数量' }
  }

  // 验证暗物质是否足够
  if (planet.resources.darkMatter < darkMatterAmount) {
    return { success: false, message: '暗物质不足' }
  }

  const { resourceGained, fee, netResource } = calculateTrade(darkMatterAmount, resourceType)

  // 执行资源交换
  planet.resources.darkMatter -= darkMatterAmount
  planet.resources[resourceType] += netResource

  // 记录交易
  const record: TradeRecord = {
    id: `trade_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`,
    timestamp: Date.now(),
    darkMatterSpent: darkMatterAmount,
    resourceType,
    resourceGained: netResource,
    fee
  }

  if (!player.tradeHistory) {
    player.tradeHistory = []
  }
  player.tradeHistory.unshift(record)
  // 只保留最近50条记录
  if (player.tradeHistory.length > 50) {
    player.tradeHistory = player.tradeHistory.slice(0, 50)
  }

  const resourceName = resourceType === 'metal' ? '金属' : resourceType === 'crystal' ? '晶体' : '重氢'
  return {
    success: true,
    message: `交易成功！花费 ${darkMatterAmount} 暗物质，获得 ${netResource.toLocaleString()} ${resourceName}（手续费 ${fee.toLocaleString()}）`,
    record
  }
}
