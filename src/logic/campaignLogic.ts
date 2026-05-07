/**
 * 战役系统逻辑
 * 处理任务进度、解锁、完成等核心逻辑
 */

import {
  QuestStatus,
  ObjectiveType,
  type Player,
  type PlayerCampaignProgress,
  type QuestObjective,
  type QuestReward,
  type QuestNotification,
  type CampaignQuestConfig,
  type NPC,
  RelationStatus,
  type BuildingType,
  type TechnologyType,
  type ShipType
} from '@/types/game'
import { MAIN_CAMPAIGN, getAllQuests, getQuestById, getQuestsByChapter } from '@/config/campaignConfig'
import { generateId } from '@/utils/id'
import * as resourceLogic from './resourceLogic'

/**
 * 初始化玩家战役进度
 */
export const initializeCampaignProgress = (player: Player): PlayerCampaignProgress => {
  // 如果已有进度，返回现有进度
  if (player.campaignProgress) {
    return player.campaignProgress
  }

  // 创建初始进度，第一个任务默认解锁
  const firstQuest = getQuestsByChapter(1)[0]
  const progress: PlayerCampaignProgress = {
    campaignId: MAIN_CAMPAIGN.id,
    currentChapter: 1,
    currentQuestId: firstQuest?.id,
    questProgress: {},
    completedQuests: [],
    unlockedQuests: firstQuest ? [firstQuest.id] : [],
    branchChoices: {}
  }

  // 初始化第一个任务为可接取状态
  if (firstQuest) {
    progress.questProgress[firstQuest.id] = {
      questId: firstQuest.id,
      status: QuestStatus.Available,
      objectives: {}
    }
    // 初始化目标进度
    const questProgress = progress.questProgress[firstQuest.id]
    if (questProgress) {
      firstQuest.objectives.forEach(obj => {
        questProgress.objectives[obj.id] = {
          current: 0,
          completed: false
        }
      })
    }
  }

  return progress
}

/**
 * 获取任务状态
 */
export const getQuestStatus = (progress: PlayerCampaignProgress, questId: string): QuestStatus => {
  const questProgress = progress.questProgress[questId]
  if (questProgress) {
    return questProgress.status
  }

  // 检查是否已解锁
  if (progress.unlockedQuests.includes(questId)) {
    return QuestStatus.Available
  }

  return QuestStatus.Locked
}

/**
 * 开始任务
 */
export const startQuest = (player: Player, questId: string): { success: boolean; error?: string } => {
  const progress = player.campaignProgress
  if (!progress) {
    return { success: false, error: 'campaign.errors.notInitialized' }
  }

  const quest = getQuestById(questId)
  if (!quest) {
    return { success: false, error: 'campaign.errors.questNotFound' }
  }

  // 检查是否已解锁
  if (!progress.unlockedQuests.includes(questId)) {
    return { success: false, error: 'campaign.errors.questLocked' }
  }

  // 检查是否已完成
  if (progress.completedQuests.includes(questId)) {
    return { success: false, error: 'campaign.errors.questAlreadyCompleted' }
  }

  // 初始化或更新任务进度
  if (!progress.questProgress[questId]) {
    progress.questProgress[questId] = {
      questId,
      status: QuestStatus.Active,
      objectives: {},
      startedAt: Date.now()
    }
  }
  const questProgress = progress.questProgress[questId]
  if (questProgress) {
    questProgress.status = QuestStatus.Active
    questProgress.startedAt = Date.now()
    // 初始化目标进度
    quest.objectives.forEach(obj => {
      if (!questProgress.objectives[obj.id]) {
        questProgress.objectives[obj.id] = {
          current: 0,
          completed: false
        }
      }
    })
  }

  progress.currentQuestId = questId

  return { success: true }
}

/**
 * 检查单个目标的进度
 */
export const checkObjectiveProgress = (player: Player, objective: QuestObjective, npcs: NPC[]): number => {
  switch (objective.type) {
    case ObjectiveType.BuildBuilding: {
      // 检查玩家所有星球上该建筑的最高等级
      const targetLevel = objective.targetSecondary as number
      const buildingType = objective.target as BuildingType
      const hasBuilding = player.planets.some(planet => {
        const level = planet.buildings[buildingType] || 0
        return level >= targetLevel
      })
      return hasBuilding ? 1 : 0
    }

    case ObjectiveType.ResearchTech: {
      const targetLevel = objective.targetSecondary as number
      const techType = objective.target as TechnologyType
      const currentLevel = player.technologies[techType] || 0
      return currentLevel >= targetLevel ? 1 : 0
    }

    case ObjectiveType.ProduceShips: {
      // 统计所有星球上该类型舰船的总数
      const shipType = objective.target as ShipType
      let totalShips = 0
      player.planets.forEach(planet => {
        totalShips += planet.fleet[shipType] || 0
      })
      return Math.min(totalShips, objective.required)
    }

    case ObjectiveType.AccumulateResources: {
      // 统计所有星球的资源总量
      let totalResources = 0
      player.planets.forEach(planet => {
        totalResources += planet.resources.metal + planet.resources.crystal + planet.resources.deuterium
      })
      return Math.min(totalResources, objective.required)
    }

    case ObjectiveType.Colonize: {
      // 统计殖民地数量（不包括母星）
      const colonyCount = player.planets.length - 1
      return Math.min(colonyCount, objective.required)
    }

    case ObjectiveType.SpyTarget: {
      // 统计侦查报告数量
      const target = objective.target
      if (target === 'any') {
        return Math.min(player.spyReports.length, objective.required)
      } else if (target === 'hostile') {
        // 统计敌对NPC的侦查报告
        const hostileNpcIds = npcs
          .filter(npc => {
            const relation = npc.relations?.[player.id]
            return relation?.status === RelationStatus.Hostile
          })
          .map(npc => npc.id)
        const hostileReports = player.spyReports.filter(report => hostileNpcIds.includes(report.targetPlayerId || ''))
        return Math.min(hostileReports.length, objective.required)
      }
      return 0
    }

    case ObjectiveType.SendGift: {
      // 从成就统计中获取送礼次数
      return Math.min(player.achievementStats?.giftsSent || 0, objective.required)
    }

    case ObjectiveType.Expedition: {
      // 从成就统计中获取探险次数（使用 expeditionsTotal）
      return Math.min(player.achievementStats?.expeditionsTotal || 0, objective.required)
    }

    case ObjectiveType.WinBattles: {
      const target = objective.target
      if (target === 'attack') {
        return Math.min(player.achievementStats?.attacksWon || 0, objective.required)
      } else if (target === 'defense') {
        // 使用 defensesSuccessful 代替 defensesWon
        return Math.min(player.achievementStats?.defensesSuccessful || 0, objective.required)
      }
      const totalWins = (player.achievementStats?.attacksWon || 0) + (player.achievementStats?.defensesSuccessful || 0)
      return Math.min(totalWins, objective.required)
    }

    case ObjectiveType.RecycleDebris: {
      return Math.min(player.achievementStats?.recyclingMissions || 0, objective.required)
    }

    case ObjectiveType.ReachRelation: {
      // 检查是否有任何NPC达到指定关系等级
      const targetStatus = objective.target as string
      const hasRelation = npcs.some(npc => {
        const relation = npc.relations?.[player.id]
        return relation?.status === targetStatus
      })
      return hasRelation ? 1 : 0
    }

    case ObjectiveType.FormAlliance: {
      // 检查是否有盟友（使用 Friendly 作为最高关系等级）
      const hasAlliance = npcs.some(npc => {
        const relation = npc.relations?.[player.id]
        return relation?.status === RelationStatus.Friendly && (relation.reputation || 0) >= 80
      })
      return hasAlliance ? 1 : 0
    }

    case ObjectiveType.DefeatNPC: {
      // 检查攻击NPC获胜次数
      return Math.min(player.achievementStats?.attacksWon || 0, objective.required)
    }

    case ObjectiveType.DestroyPlanet: {
      return Math.min(player.achievementStats?.planetDestructions || 0, objective.required)
    }

    default:
      return 0
  }
}

/**
 * 更新任务进度
 */
export const updateQuestProgress = (
  player: Player,
  questId: string,
  npcs: NPC[]
): {
  updated: boolean
  completedObjectives: string[]
  questCompleted: boolean
} => {
  const result = {
    updated: false,
    completedObjectives: [] as string[],
    questCompleted: false
  }

  const progress = player.campaignProgress
  if (!progress) return result

  const questProgress = progress.questProgress[questId]
  if (!questProgress || questProgress.status !== QuestStatus.Active) return result

  const quest = getQuestById(questId)
  if (!quest) return result

  // 检查每个目标的进度
  quest.objectives.forEach(objective => {
    const objProgress = questProgress.objectives[objective.id]
    if (!objProgress || objProgress.completed) return

    const newProgress = checkObjectiveProgress(player, objective, npcs)
    if (newProgress !== objProgress.current) {
      objProgress.current = newProgress
      result.updated = true

      // 检查目标是否完成
      if (newProgress >= objective.required && !objProgress.completed) {
        objProgress.completed = true
        result.completedObjectives.push(objective.id)
      }
    }
  })

  // 检查任务是否全部完成
  const allCompleted = quest.objectives.every(obj => questProgress.objectives[obj.id]?.completed)

  if (allCompleted && questProgress.status === QuestStatus.Active) {
    questProgress.status = QuestStatus.Completed
    questProgress.completedAt = Date.now()
    result.questCompleted = true
  }

  return result
}

/**
 * 领取任务奖励
 */
export const claimQuestRewards = (
  player: Player,
  questId: string
): {
  success: boolean
  rewards?: QuestReward
  error?: string
} => {
  const progress = player.campaignProgress
  if (!progress) {
    return { success: false, error: 'campaign.errors.notInitialized' }
  }

  const questProgress = progress.questProgress[questId]
  if (!questProgress) {
    return { success: false, error: 'campaign.errors.questNotFound' }
  }

  if (questProgress.status !== QuestStatus.Completed) {
    return { success: false, error: 'campaign.errors.questNotCompleted' }
  }

  if (questProgress.rewardsClaimed) {
    return { success: false, error: 'campaign.errors.rewardsAlreadyClaimed' }
  }

  const quest = getQuestById(questId)
  if (!quest) {
    return { success: false, error: 'campaign.errors.questNotFound' }
  }

  // 发放奖励
  const rewards = quest.rewards
  const currentPlanet = player.planets[0] // 默认发放到第一个星球

  if (rewards.resources && currentPlanet) {
    resourceLogic.addResources(currentPlanet.resources, rewards.resources)
  }

  if (rewards.darkMatter && currentPlanet) {
    currentPlanet.resources.darkMatter += rewards.darkMatter
  }

  if (rewards.points) {
    // 奖励积分存入 bonusPoints，不会被 calculatePlayerPoints 覆盖
    player.bonusPoints = (player.bonusPoints || 0) + rewards.points
  }

  if (rewards.ships && currentPlanet) {
    Object.entries(rewards.ships).forEach(([shipType, count]) => {
      const ship = shipType as ShipType
      currentPlanet.fleet[ship] = (currentPlanet.fleet[ship] || 0) + count
    })
  }

  // 标记奖励已领取
  questProgress.rewardsClaimed = true

  // 将任务添加到已完成列表
  if (!progress.completedQuests.includes(questId)) {
    progress.completedQuests.push(questId)
  }

  // 解锁后续任务
  unlockNextQuests(player, questId)

  return { success: true, rewards }
}

/**
 * 解锁后续任务
 */
export const unlockNextQuests = (player: Player, _completedQuestId: string): string[] => {
  const progress = player.campaignProgress
  if (!progress) return []

  const unlockedQuests: string[] = []
  const allQuests = getAllQuests()

  allQuests.forEach(quest => {
    // 跳过已解锁的任务
    if (progress.unlockedQuests.includes(quest.id)) return

    // 检查前置任务是否完成
    if (quest.requiredQuestIds && quest.requiredQuestIds.length > 0) {
      const allRequiredCompleted = quest.requiredQuestIds.every(reqId => progress.completedQuests.includes(reqId))

      if (allRequiredCompleted) {
        progress.unlockedQuests.push(quest.id)
        unlockedQuests.push(quest.id)

        // 初始化任务进度
        progress.questProgress[quest.id] = {
          questId: quest.id,
          status: QuestStatus.Available,
          objectives: {}
        }
        const questProgress = progress.questProgress[quest.id]
        if (questProgress) {
          quest.objectives.forEach(obj => {
            questProgress.objectives[obj.id] = {
              current: 0,
              completed: false
            }
          })
        }

        // 更新当前章节
        if (quest.chapter > progress.currentChapter) {
          progress.currentChapter = quest.chapter
        }
      }
    }
  })

  return unlockedQuests
}

/**
 * 计算战役总进度百分比
 */
export const calculateCampaignProgress = (progress: PlayerCampaignProgress): number => {
  const totalQuests = getAllQuests().length
  const completedCount = progress.completedQuests.length
  return Math.round((completedCount / totalQuests) * 100)
}

/**
 * 计算章节进度百分比
 */
export const calculateChapterProgress = (progress: PlayerCampaignProgress, chapterNumber: number): number => {
  const chapterQuests = getQuestsByChapter(chapterNumber)
  const completedCount = chapterQuests.filter(quest => progress.completedQuests.includes(quest.id)).length
  return Math.round((completedCount / chapterQuests.length) * 100)
}

/**
 * 计算单个任务的进度百分比
 */
export const calculateQuestProgress = (progress: PlayerCampaignProgress, questId: string): number => {
  const questProgress = progress.questProgress[questId]
  if (!questProgress) return 0

  const quest = getQuestById(questId)
  if (!quest) return 0

  // 计算所有目标的平均进度
  const objectives = quest.objectives
  if (objectives.length === 0) return 0

  let totalProgress = 0
  objectives.forEach(obj => {
    const objProgress = questProgress.objectives[obj.id]
    if (objProgress) {
      if (objProgress.completed) {
        totalProgress += 100
      } else {
        totalProgress += Math.min((objProgress.current / obj.required) * 100, 100)
      }
    }
  })

  return Math.round(totalProgress / objectives.length)
}

/**
 * 获取当前可进行的任务列表
 */
export const getAvailableQuests = (progress: PlayerCampaignProgress): CampaignQuestConfig[] => {
  return getAllQuests().filter(quest => {
    const status = getQuestStatus(progress, quest.id)
    return status === QuestStatus.Available || status === QuestStatus.Active
  })
}

/**
 * 获取当前激活的任务
 */
export const getActiveQuest = (progress: PlayerCampaignProgress): CampaignQuestConfig | undefined => {
  if (!progress.currentQuestId) return undefined
  return getQuestById(progress.currentQuestId)
}

/**
 * 创建任务通知
 */
export const createQuestNotification = (
  questId: string,
  eventType: QuestNotification['eventType'],
  messageKey: string,
  messageParams?: Record<string, string | number>,
  rewards?: QuestReward
): QuestNotification => {
  const quest = getQuestById(questId)
  return {
    id: generateId('quest_notification'),
    timestamp: Date.now(),
    questId,
    questTitleKey: quest?.titleKey || '',
    eventType,
    messageKey,
    messageParams,
    rewards,
    read: false
  }
}

/**
 * 添加任务通知到玩家
 */
export const addQuestNotification = (player: Player, notification: QuestNotification): void => {
  if (!player.questNotifications) {
    player.questNotifications = []
  }
  player.questNotifications.unshift(notification)

  // 限制通知数量
  if (player.questNotifications.length > 50) {
    player.questNotifications = player.questNotifications.slice(0, 50)
  }
}

/**
 * 标记通知为已读
 */
export const markQuestNotificationRead = (player: Player, notificationId: string): void => {
  const notification = player.questNotifications?.find(n => n.id === notificationId)
  if (notification) {
    notification.read = true
  }
}

/**
 * 获取未读通知数量
 */
export const getUnreadQuestNotificationCount = (player: Player): number => {
  return player.questNotifications?.filter(n => !n.read).length || 0
}

/**
 * 检查并更新所有激活任务的进度
 */
export const checkAllActiveQuestsProgress = (player: Player, npcs: NPC[]): QuestNotification[] => {
  const notifications: QuestNotification[] = []
  const progress = player.campaignProgress

  if (!progress) return notifications

  // 遍历所有激活的任务
  Object.entries(progress.questProgress).forEach(([questId, questProgress]) => {
    if (questProgress.status !== QuestStatus.Active) return

    const result = updateQuestProgress(player, questId, npcs)

    // 为完成的目标创建通知
    result.completedObjectives.forEach(objId => {
      const quest = getQuestById(questId)
      const objective = quest?.objectives.find(o => o.id === objId)
      if (objective) {
        notifications.push(
          createQuestNotification(questId, 'objective_completed', 'campaign.notifications.objectiveCompleted', {
            objective: objective.descriptionKey
          })
        )
      }
    })

    // 任务完成通知
    if (result.questCompleted) {
      const quest = getQuestById(questId)
      notifications.push(createQuestNotification(questId, 'quest_completed', 'campaign.notifications.questCompleted', {}, quest?.rewards))
    }
  })

  // 添加通知到玩家
  notifications.forEach(notification => {
    addQuestNotification(player, notification)
  })

  return notifications
}
