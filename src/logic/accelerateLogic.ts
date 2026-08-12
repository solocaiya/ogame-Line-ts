/**
 * 加速逻辑 — 暗物质加速建造/研究/造船/舰队旅行
 *
 * 核心公式：costPerHour DM = 60 分钟（1 小时），不足 1 小时向上取整，最低 costPerHour DM
 * 不同类型加速消耗不同（与后端 shopConfig.go 保持一致）
 */

/** 每小时加速消耗的暗物质，按类型区分（与后端 AccelerateCost* 常量一致） */
export const ACCELERATE_COST_BUILDING = 200
export const ACCELERATE_COST_RESEARCH = 250
export const ACCELERATE_COST_FLEET_BUILD = 300
/** 舰队旅行加速: 固定 500 DM/次，不按小时计（与后端 AccelerateCostFleetTravel 一致） */
export const ACCELERATE_COST_FLEET_TRAVEL = 500

/** @deprecated 使用按类型区分的常量 */
export const ACCELERATE_COST_PER_HOUR = ACCELERATE_COST_BUILDING

/**
 * 计算加速剩余时间所需的暗物质数量
 * @param remainingMs 剩余时间（毫秒）
 * @param costPerHour 每小时消耗的暗物质（默认建筑加速）
 * @returns 所需暗物质数量
 */
export function calculateAccelerateCost(remainingMs: number, costPerHour: number = ACCELERATE_COST_BUILDING): number {
  if (remainingMs <= 0) return 0
  const remainingMinutes = remainingMs / 60000
  const hours = Math.ceil(remainingMinutes / 60)
  return Math.max(1, hours) * costPerHour
}

/**
 * 计算舰队旅行加速费用（固定 500 DM/次，不按时间计）
 * 与后端 accelerate.go fleet-travel 逻辑一致
 */
export function calculateFleetTravelCost(): number {
  return ACCELERATE_COST_FLEET_TRAVEL
}

/**
 * 计算加速后的新结束时间
 * @param currentEndTime 当前结束时间（unix ms）
 * @param skipMs 要跳过的毫秒数
 * @returns 新的结束时间（unix ms）
 */
export function calculateNewEndTime(currentEndTime: number, skipMs: number): number {
  return currentEndTime - skipMs
}

/**
 * 计算跳过指定分钟数对应的毫秒数
 * @param minutes 要跳过的分钟数
 * @returns 毫秒数
 */
export function minutesToMs(minutes: number): number {
  return minutes * 60 * 1000
}

/**
 * 格式化剩余时间为人类可读字符串
 * @param remainingMs 剩余毫秒
 * @returns 格式化字符串，如 "2h 30m" 或 "45m"
 */
export function formatRemainingTime(remainingMs: number): string {
  if (remainingMs <= 0) return '0m'
  const totalMinutes = Math.ceil(remainingMs / 60000)
  const hours = Math.floor(totalMinutes / 60)
  const minutes = totalMinutes % 60
  if (hours > 0) {
    return minutes > 0 ? `${hours}h ${minutes}m` : `${hours}h`
  }
  return `${minutes}m`
}

/**
 * 计算加速可节省的时间（用于显示）
 * @param skipMinutes 用户选择跳过的分钟数
 * @param remainingMs 实际剩余时间（毫秒）
 * @returns 实际跳过的毫秒数（不超过剩余时间）
 */
export function calculateActualSkip(skipMinutes: number, remainingMs: number): number {
  const skipMs = minutesToMs(skipMinutes)
  return Math.min(skipMs, remainingMs)
}
