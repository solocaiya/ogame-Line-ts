// 7日签到逻辑

import type { Player, Planet } from '@/types/game'
import { CHECK_IN_REWARDS, CHECK_IN_CYCLE_DAYS } from '@/config/checkInConfig'

/**
 * 获取今天的日期字符串（YYYY-MM-DD）
 */
export const getTodayDateString = (): string => {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`
}

/**
 * 判断今天是否已签到
 */
export const canCheckInToday = (player: Player): boolean => {
  if (!player.lastCheckInDate) return true
  return player.lastCheckInDate !== getTodayDateString()
}

/**
 * 获取当前是签到周期的第几天（1-7）
 */
export const getCurrentCheckInDay = (player: Player): number => {
  const history = player.checkInHistory || []
  const dayInCycle = (history.length % CHECK_IN_CYCLE_DAYS) + 1
  return dayInCycle
}

/**
 * 获取当前签到周期的进度（已签到天数/7）
 */
export const getCheckInProgress = (player: Player): { current: number; total: number } => {
  const history = player.checkInHistory || []
  const current = history.length % CHECK_IN_CYCLE_DAYS
  return { current, total: CHECK_IN_CYCLE_DAYS }
}

/**
 * 判断是否完成了一个完整的7日周期
 */
export const isSevenDayCycleComplete = (player: Player): boolean => {
  const history = player.checkInHistory || []
  return history.length > 0 && history.length % CHECK_IN_CYCLE_DAYS === 0
}

/**
 * 领取签到奖励
 */
export const claimCheckIn = (player: Player, planet: Planet): { success: boolean; reward: any; message: string } => {
  if (!canCheckInToday(player)) {
    return { success: false, reward: null, message: '今天已经签到过了' }
  }

  const currentDay = getCurrentCheckInDay(player)
  const reward = CHECK_IN_REWARDS[currentDay - 1]

  if (!reward) {
    return { success: false, reward: null, message: '签到奖励配置错误' }
  }

  // 添加资源到当前星球
  if (reward.resources) {
    if (reward.resources.metal) planet.resources.metal += reward.resources.metal
    if (reward.resources.crystal) planet.resources.crystal += reward.resources.crystal
    if (reward.resources.deuterium) planet.resources.deuterium += reward.resources.deuterium
    if (reward.resources.darkMatter) planet.resources.darkMatter += reward.resources.darkMatter
  }

  // 更新签到历史
  if (!player.checkInHistory) {
    player.checkInHistory = []
  }
  player.checkInHistory.push(Date.now())
  player.lastCheckInDate = getTodayDateString()

  return {
    success: true,
    reward,
    message: `签到成功！获得第${currentDay}天奖励`
  }
}

/**
 * 判断指定天数是否已领取（基于签到历史长度）
 */
export const isDayClaimed = (player: Player, day: number): boolean => {
  const history = player.checkInHistory || []
  return history.length >= day
}

/**
 * 重置签到周期（完成7天后自动重置）
 */
export const resetCheckInCycle = (player: Player): void => {
  player.checkInHistory = []
  player.lastCheckInDate = undefined
}
