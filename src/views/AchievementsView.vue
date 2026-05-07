<template>
  <div class="container mx-auto p-4 sm:p-6 space-y-4 sm:space-y-6">
    <div class="flex flex-row items-center justify-between gap-4">
      <h1 class="text-2xl sm:text-3xl font-bold">{{ t('achievements.title') }}</h1>
      <div class="flex items-center gap-2">{{ unlockedCount }} / {{ totalCount }} {{ t('achievements.unlocked') }}</div>
    </div>

    <!-- 分类标签 -->
    <Tabs v-model="activeCategory" class="w-full">
      <TabsList class="w-full grid grid-cols-5 h-10">
        <TabsTrigger v-for="category in categories" :key="category.value" :value="category.value" class="text-xs sm:text-sm">
          {{ t(`achievements.categories.${category.value}`) }}
          <Badge v-if="getCategoryUnlockedCount(category.value) > 0" class="ml-1 h-5 px-1.5 text-[10px] bg-primary text-primary-foreground">
            {{ getCategoryUnlockedCount(category.value) }}
          </Badge>
        </TabsTrigger>
      </TabsList>

      <!-- 成就卡片网格 -->
      <TabsContent v-for="category in categories" :key="category.value" :value="category.value" class="mt-4">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          <Card v-for="achievement in getAchievementsByCategory(category.value)" :key="achievement.id" class="relative overflow-hidden">
            <!-- 等级指示条 -->
            <div class="absolute top-0 left-0 right-0 h-1 flex">
              <div v-for="tier in tierOrder" :key="tier" class="flex-1" :class="getTierBarClass(achievement.id, tier)" />
            </div>

            <CardHeader class="pt-4">
              <div class="flex items-start gap-3">
                <div class="p-2 rounded-lg" :class="getIconBgClass(achievement.id)">
                  <component :is="getIcon(achievement.icon)" class="h-6 w-6" :class="getIconClass(achievement.id)" />
                </div>
                <div class="flex-1 min-w-0">
                  <CardTitle class="text-sm sm:text-base flex items-center gap-2">
                    {{ t(`achievements.names.${achievement.id}`) }}
                    <Badge v-if="getCurrentTier(achievement.id)" :class="getTierBadgeClass(getCurrentTier(achievement.id)!)">
                      {{ t(`achievements.tiers.${getCurrentTier(achievement.id)}`) }}
                    </Badge>
                  </CardTitle>
                  <CardDescription class="text-xs mt-1">
                    {{ t(`achievements.descriptions.${achievement.id}`) }}
                  </CardDescription>
                </div>
              </div>
            </CardHeader>

            <CardContent class="space-y-3">
              <!-- 进度条 -->
              <div class="space-y-1">
                <div class="flex justify-between text-xs">
                  <span class="text-muted-foreground">{{ t('achievements.progress') }}</span>
                  <span class="font-medium">
                    {{ formatNumber(getCurrentValue(achievement.id)) }} /
                    {{ formatNumber(getNextTarget(achievement.id) || getCurrentValue(achievement.id)) }}
                  </span>
                </div>
                <Progress :model-value="getProgressPercentage(achievement.id)" class="h-2" />
              </div>

              <!-- 下一等级奖励 -->
              <div v-if="getNextTierConfig(achievement.id)" class="p-2 bg-muted/50 rounded-lg">
                <p class="text-xs text-muted-foreground mb-1">
                  {{ t('achievements.nextTier') }}: {{ t(`achievements.tiers.${getNextTierConfig(achievement.id)!.tier}`) }}
                </p>
                <div class="flex items-center gap-3 text-xs">
                  <div v-if="getNextTierConfig(achievement.id)!.reward.darkMatter" class="flex items-center gap-1">
                    <Sparkles class="h-3 w-3 text-purple-500" />
                    <span>+{{ formatNumber(getNextTierConfig(achievement.id)!.reward.darkMatter!) }}</span>
                  </div>
                  <div v-if="getNextTierConfig(achievement.id)!.reward.points" class="flex items-center gap-1">
                    <Star class="h-3 w-3 text-yellow-500" />
                    <span>+{{ formatNumber(getNextTierConfig(achievement.id)!.reward.points!) }}</span>
                  </div>
                </div>
              </div>

              <!-- 已达最高等级 -->
              <div
                v-else-if="getCurrentTier(achievement.id) === 'diamond'"
                class="p-2 bg-gradient-to-r from-purple-500/10 to-blue-500/10 rounded-lg"
              >
                <p class="text-xs text-center font-medium text-purple-600 dark:text-purple-400">
                  {{ t('achievements.maxTierReached') }}
                </p>
              </div>
            </CardContent>
          </Card>
        </div>
      </TabsContent>
    </Tabs>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue'
  import { useGameStore } from '@/stores/gameStore'
  import { useI18n } from '@/composables/useI18n'
  import { formatNumber } from '@/utils/format'
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
  import { Badge } from '@/components/ui/badge'
  import { Progress } from '@/components/ui/progress'
  import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
  import { AchievementCategory, AchievementTier, type AchievementTierConfig } from '@/types/game'
  import { ACHIEVEMENTS, ACHIEVEMENT_MAP, TIER_ORDER, getNextTier } from '@/config/achievementConfig'
  import { getAchievementProgress } from '@/logic/achievementLogic'
  import {
    Sparkles,
    Star,
    Gem,
    Diamond,
    Droplet,
    Flame,
    Building2,
    FlaskConical,
    Rocket,
    Shield,
    Swords,
    Crown,
    ShieldCheck,
    Bomb,
    Trash2,
    Skull,
    ShieldOff,
    Plane,
    Truck,
    Package,
    Flag,
    Eye,
    ArrowDownToLine,
    Compass,
    Sparkle,
    Recycle,
    Pickaxe,
    Zap,
    Fuel,
    Handshake as HandshakeIcon,
    Angry,
    Gift,
    HeartHandshake,
    Target,
    ScanEye,
    Banknote,
    BadgeDollarSign
  } from 'lucide-vue-next'

  const { t } = useI18n()
  const gameStore = useGameStore()

  const activeCategory = ref<AchievementCategory>(AchievementCategory.Resource)

  const categories = [
    { value: AchievementCategory.Resource },
    { value: AchievementCategory.Building },
    { value: AchievementCategory.Combat },
    { value: AchievementCategory.Mission },
    { value: AchievementCategory.Diplomacy }
  ]

  const tierOrder = TIER_ORDER

  // 图标映射
  const iconMap: Record<string, any> = {
    Gem,
    Diamond,
    Droplet,
    Sparkles,
    Flame,
    Building2,
    FlaskConical,
    Rocket,
    Shield,
    Swords,
    Crown,
    ShieldCheck,
    Bomb,
    Trash2,
    Skull,
    ShieldOff,
    Plane,
    Truck,
    Package,
    Flag,
    Eye,
    ArrowDownToLine,
    Compass,
    Sparkle,
    Recycle,
    Pickaxe,
    Zap,
    Fuel,
    HandshakeIcon,
    Angry,
    Gift,
    HeartHandshake,
    Target,
    ScanEye,
    Banknote,
    BadgeDollarSign
  }

  const getIcon = (iconName: string) => {
    return iconMap[iconName] || Sparkles
  }

  // 获取成就进度
  const getProgress = (achievementId: string) => {
    return gameStore.player.achievements?.[achievementId]
  }

  const getCurrentTier = (achievementId: string) => {
    return getProgress(achievementId)?.currentTier || null
  }

  const getCurrentValue = (achievementId: string) => {
    return getProgress(achievementId)?.currentValue || 0
  }

  const getNextTarget = (achievementId: string) => {
    const config = ACHIEVEMENT_MAP[achievementId]
    if (!config) return null

    const currentTier = getCurrentTier(achievementId)
    const nextTier = getNextTier(currentTier)
    if (!nextTier) return null

    const tierConfig = config.tiers.find(t => t.tier === nextTier)
    return tierConfig?.target ?? null
  }

  const getNextTierConfig = (achievementId: string): AchievementTierConfig | null => {
    const config = ACHIEVEMENT_MAP[achievementId]
    if (!config) return null

    const currentTier = getCurrentTier(achievementId)
    const nextTier = getNextTier(currentTier)
    if (!nextTier) return null

    return config.tiers.find(t => t.tier === nextTier) || null
  }

  const getProgressPercentage = (achievementId: string) => {
    const currentValue = getCurrentValue(achievementId)
    const currentTier = getCurrentTier(achievementId)
    return getAchievementProgress(achievementId, currentValue, currentTier)
  }

  // 按类别获取成就
  const getAchievementsByCategory = (category: AchievementCategory) => {
    return ACHIEVEMENTS.filter(a => a.category === category)
  }

  // 统计
  const unlockedCount = computed(() => {
    if (!gameStore.player.achievements) return 0
    return Object.values(gameStore.player.achievements).filter(p => p.currentTier !== null).length
  })

  const totalCount = computed(() => ACHIEVEMENTS.length)

  const getCategoryUnlockedCount = (category: AchievementCategory) => {
    if (!gameStore.player.achievements) return 0
    const categoryAchievements = ACHIEVEMENTS.filter(a => a.category === category)
    return categoryAchievements.filter(a => {
      const progress = gameStore.player.achievements?.[a.id]
      return progress?.currentTier !== null
    }).length
  }

  // 样式函数
  const getTierBarClass = (achievementId: string, tier: AchievementTier) => {
    const progress = getProgress(achievementId)
    if (!progress) return 'bg-muted'

    const tierUnlock = progress.tierUnlocks[tier]
    if (tierUnlock !== null) {
      // 已解锁
      switch (tier) {
        case AchievementTier.Bronze:
          return 'bg-amber-600'
        case AchievementTier.Silver:
          return 'bg-gray-400'
        case AchievementTier.Gold:
          return 'bg-yellow-500'
        case AchievementTier.Platinum:
          return 'bg-cyan-400'
        case AchievementTier.Diamond:
          return 'bg-purple-500'
      }
    }
    return 'bg-muted'
  }

  const getTierBadgeClass = (tier: AchievementTier) => {
    switch (tier) {
      case AchievementTier.Bronze:
        return 'bg-amber-600 text-white'
      case AchievementTier.Silver:
        return 'bg-gray-400 text-white'
      case AchievementTier.Gold:
        return 'bg-yellow-500 text-black'
      case AchievementTier.Platinum:
        return 'bg-cyan-400 text-black'
      case AchievementTier.Diamond:
        return 'bg-gradient-to-r from-purple-500 to-blue-500 text-white'
    }
  }

  const getIconBgClass = (achievementId: string) => {
    const tier = getCurrentTier(achievementId)
    if (!tier) return 'bg-muted'

    switch (tier) {
      case AchievementTier.Bronze:
        return 'bg-amber-100 dark:bg-amber-900/30'
      case AchievementTier.Silver:
        return 'bg-gray-100 dark:bg-gray-800'
      case AchievementTier.Gold:
        return 'bg-yellow-100 dark:bg-yellow-900/30'
      case AchievementTier.Platinum:
        return 'bg-cyan-100 dark:bg-cyan-900/30'
      case AchievementTier.Diamond:
        return 'bg-purple-100 dark:bg-purple-900/30'
    }
  }

  const getIconClass = (achievementId: string) => {
    const tier = getCurrentTier(achievementId)
    if (!tier) return 'text-muted-foreground'

    switch (tier) {
      case AchievementTier.Bronze:
        return 'text-amber-600'
      case AchievementTier.Silver:
        return 'text-gray-500'
      case AchievementTier.Gold:
        return 'text-yellow-600'
      case AchievementTier.Platinum:
        return 'text-cyan-500'
      case AchievementTier.Diamond:
        return 'text-purple-500'
    }
  }
</script>
