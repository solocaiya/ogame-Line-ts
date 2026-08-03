// 7日签到配置

export interface CheckInReward {
  day: number
  resources?: {
    metal?: number
    crystal?: number
    deuterium?: number
    darkMatter?: number
  }
  specialItem?: string
}

export const CHECK_IN_REWARDS: CheckInReward[] = [
  {
    day: 1,
    resources: { metal: 5000, crystal: 3000, deuterium: 1000 }
  },
  {
    day: 2,
    resources: { metal: 8000, crystal: 5000, deuterium: 2000 }
  },
  {
    day: 3,
    resources: { metal: 12000, crystal: 8000, deuterium: 3000, darkMatter: 50 }
  },
  {
    day: 4,
    resources: { metal: 15000, crystal: 10000, deuterium: 5000 }
  },
  {
    day: 5,
    resources: { metal: 20000, crystal: 15000, deuterium: 8000, darkMatter: 100 }
  },
  {
    day: 6,
    resources: { metal: 25000, crystal: 18000, deuterium: 10000 }
  },
  {
    day: 7,
    resources: { metal: 30000, crystal: 20000, deuterium: 15000, darkMatter: 200 },
    specialItem: 'weekly_bonus'
  }
]

export const CHECK_IN_CYCLE_DAYS = 7
