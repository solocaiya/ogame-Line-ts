<template>
  <div class="container mx-auto p-4 sm:p-6 space-y-6">
    <h1 class="text-2xl sm:text-3xl font-bold">{{ t('statistics.title') }}</h1>

    <!-- 总分卡片 -->
    <Card class="bg-gradient-to-r from-primary/10 to-purple-500/10 border-primary/20">
      <CardContent class="p-6 text-center">
        <div class="text-sm text-muted-foreground mb-1">{{ t('statistics.totalScore') }}</div>
        <div class="text-4xl font-bold text-primary">{{ formatNumber(stats.totalScore) }}</div>
        <div class="text-xs text-muted-foreground mt-2">
          {{ t('statistics.playTime') }}: {{ formatPlayTime(stats.totalPlayTime) }}
        </div>
      </CardContent>
    </Card>

    <!-- 资源统计 -->
    <Card>
      <CardHeader>
        <CardTitle class="text-lg">{{ t('statistics.resources') }}</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <div class="text-center p-3 bg-amber-50 dark:bg-amber-950/30 rounded-lg">
            <div class="text-xs text-muted-foreground">{{ t('resources.metal') }}</div>
            <div class="text-lg font-bold text-amber-600">{{ formatNumber(stats.totalMetalProduced) }}</div>
          </div>
          <div class="text-center p-3 bg-cyan-50 dark:bg-cyan-950/30 rounded-lg">
            <div class="text-xs text-muted-foreground">{{ t('resources.crystal') }}</div>
            <div class="text-lg font-bold text-cyan-500">{{ formatNumber(stats.totalCrystalProduced) }}</div>
          </div>
          <div class="text-center p-3 bg-green-50 dark:bg-green-950/30 rounded-lg">
            <div class="text-xs text-muted-foreground">{{ t('resources.deuterium') }}</div>
            <div class="text-lg font-bold text-green-500">{{ formatNumber(stats.totalDeuteriumProduced) }}</div>
          </div>
          <div class="text-center p-3 bg-purple-50 dark:bg-purple-950/30 rounded-lg">
            <div class="text-xs text-muted-foreground">{{ t('resources.darkMatter') }}</div>
            <div class="text-lg font-bold text-purple-500">{{ formatNumber(stats.totalDarkMatterEarned) }}</div>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- 建筑与科技 -->
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <Card>
        <CardHeader>
          <CardTitle class="text-lg">{{ t('statistics.buildings') }}</CardTitle>
        </CardHeader>
        <CardContent class="space-y-3">
          <div class="flex justify-between">
            <span class="text-sm text-muted-foreground">{{ t('statistics.totalBuildings') }}</span>
            <span class="font-bold">{{ stats.totalBuildingsBuilt }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-sm text-muted-foreground">{{ t('statistics.totalBuildingLevels') }}</span>
            <span class="font-bold">{{ stats.totalBuildingLevel }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-sm text-muted-foreground">{{ t('statistics.buildingScore') }}</span>
            <span class="font-bold text-amber-600">{{ stats.totalBuildingLevel * 10 }}</span>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle class="text-lg">{{ t('statistics.research') }}</CardTitle>
        </CardHeader>
        <CardContent class="space-y-3">
          <div class="flex justify-between">
            <span class="text-sm text-muted-foreground">{{ t('statistics.totalResearch') }}</span>
            <span class="font-bold">{{ stats.totalResearchCompleted }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-sm text-muted-foreground">{{ t('statistics.totalResearchLevels') }}</span>
            <span class="font-bold">{{ stats.totalResearchLevel }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-sm text-muted-foreground">{{ t('statistics.researchScore') }}</span>
            <span class="font-bold text-cyan-500">{{ stats.totalResearchLevel * 20 }}</span>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- 舰队与战斗 -->
    <Card>
      <CardHeader>
        <CardTitle class="text-lg">{{ t('statistics.fleetAndCombat') }}</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <div class="text-center">
            <div class="text-xs text-muted-foreground">{{ t('statistics.totalShips') }}</div>
            <div class="text-xl font-bold">{{ stats.totalShipsBuilt }}</div>
          </div>
          <div class="text-center">
            <div class="text-xs text-muted-foreground">{{ t('statistics.fleetDispatches') }}</div>
            <div class="text-xl font-bold">{{ stats.totalFleetDispatches }}</div>
          </div>
          <div class="text-center">
            <div class="text-xs text-muted-foreground">{{ t('statistics.battlesWon') }}</div>
            <div class="text-xl font-bold text-green-500">{{ stats.totalBattlesWon }}</div>
          </div>
          <div class="text-center">
            <div class="text-xs text-muted-foreground">{{ t('statistics.battlesLost') }}</div>
            <div class="text-xl font-bold text-red-500">{{ stats.totalBattlesLost }}</div>
          </div>
        </div>
        <div v-if="stats.totalBattlesWon + stats.totalBattlesLost > 0" class="mt-4">
          <div class="flex justify-between text-xs text-muted-foreground mb-1">
            <span>{{ t('statistics.winRate') }}</span>
            <span>{{ winRate }}%</span>
          </div>
          <div class="h-2 bg-muted rounded-full overflow-hidden">
            <div class="h-full bg-green-500 rounded-full" :style="{ width: `${winRate}%` }" />
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- 其他统计 -->
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <Card>
        <CardHeader>
          <CardTitle class="text-lg">{{ t('statistics.trade') }}</CardTitle>
        </CardHeader>
        <CardContent class="space-y-3">
          <div class="flex justify-between">
            <span class="text-sm text-muted-foreground">{{ t('statistics.totalTrades') }}</span>
            <span class="font-bold">{{ stats.totalTrades }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-sm text-muted-foreground">{{ t('statistics.darkMatterSpent') }}</span>
            <span class="font-bold text-purple-500">{{ formatNumber(stats.totalDarkMatterSpent) }}</span>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle class="text-lg">{{ t('statistics.checkIn') }}</CardTitle>
        </CardHeader>
        <CardContent class="space-y-3">
          <div class="flex justify-between">
            <span class="text-sm text-muted-foreground">{{ t('statistics.totalCheckInDays') }}</span>
            <span class="font-bold">{{ stats.totalCheckInDays }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-sm text-muted-foreground">{{ t('statistics.currentStreak') }}</span>
            <span class="font-bold text-orange-500">{{ stats.currentStreak }}</span>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- 星球统计 -->
    <Card>
      <CardHeader>
        <CardTitle class="text-lg">{{ t('statistics.empire') }}</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 text-center">
          <div>
            <div class="text-xs text-muted-foreground">{{ t('statistics.planets') }}</div>
            <div class="text-xl font-bold">{{ stats.totalPlanets }}</div>
          </div>
          <div>
            <div class="text-xs text-muted-foreground">{{ t('statistics.moons') }}</div>
            <div class="text-xl font-bold">{{ stats.totalMoons }}</div>
          </div>
          <div>
            <div class="text-xs text-muted-foreground">{{ t('statistics.accountAge') }}</div>
            <div class="text-xl font-bold">{{ accountAgeDays }}</div>
          </div>
          <div>
            <div class="text-xs text-muted-foreground">{{ t('statistics.bookmarks') }}</div>
            <div class="text-xl font-bold">{{ bookmarkCount }}</div>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import { useGameStore } from '@/stores/gameStore'
  import { useI18n } from '@/composables/useI18n'
  import { formatNumber } from '@/utils/format'
  import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
  import { calculateStatistics, formatPlayTime } from '@/logic/statisticsLogic'

  const gameStore = useGameStore()
  const { t } = useI18n()

  const planets = computed(() => gameStore.planets || [])

  const stats = computed(() => calculateStatistics(gameStore.player, planets.value))

  const winRate = computed(() => {
    const total = stats.value.totalBattlesWon + stats.value.totalBattlesLost
    if (total === 0) return 0
    return Math.round((stats.value.totalBattlesWon / total) * 100)
  })

  const accountAgeDays = computed(() => {
    const created = stats.value.accountCreated
    if (!created) return 0
    return Math.floor((Date.now() - created) / 86400000)
  })

  const bookmarkCount = computed(() => gameStore.player.bookmarks?.length ?? 0)
</script>
