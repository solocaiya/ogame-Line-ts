<template>
  <div class="container mx-auto p-4 sm:p-6 space-y-6">
    <div class="flex flex-row items-center justify-between gap-4">
      <h1 class="text-2xl sm:text-3xl font-bold">{{ t('checkIn.title') }}</h1>
      <Badge v-if="!canCheckIn" variant="secondary">{{ t('checkIn.claimedToday') }}</Badge>
      <Badge v-else variant="default" class="bg-green-600">{{ t('checkIn.available') }}</Badge>
    </div>

    <!-- 签到进度 -->
    <Card>
      <CardHeader>
        <CardTitle class="text-lg">{{ t('checkIn.progress') }}</CardTitle>
        <CardDescription>
          {{ t('checkIn.progressDesc', { current: progress.current, total: progress.total }) }}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Progress :model-value="(progress.current / progress.total) * 100" class="h-3" />
      </CardContent>
    </Card>

    <!-- 7日签到网格 -->
    <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-7 gap-4">
      <Card
        v-for="(reward, index) in CHECK_IN_REWARDS"
        :key="reward.day"
        class="relative overflow-hidden"
        :class="{
          'ring-2 ring-primary': index + 1 === currentDay && canCheckIn,
          'opacity-60': index + 1 < currentDay || (index + 1 > currentDay && !canCheckIn),
          'bg-gradient-to-br from-primary/10 to-primary/5': index + 1 === currentDay && canCheckIn
        }"
      >
        <!-- 完成标记 -->
        <div v-if="index + 1 < currentDay || (index + 1 === currentDay && !canCheckIn)" class="absolute top-2 right-2">
          <CheckCircle class="h-5 w-5 text-green-500" />
        </div>

        <!-- 当前天数标记 -->
        <div v-if="index + 1 === currentDay && canCheckIn" class="absolute top-2 right-2">
          <div class="h-3 w-3 rounded-full bg-primary animate-pulse" />
        </div>

        <CardHeader class="pb-2">
          <CardTitle class="text-sm flex items-center gap-2">
            <Calendar class="h-4 w-4" />
            {{ t('checkIn.day', { day: reward.day }) }}
          </CardTitle>
        </CardHeader>

        <CardContent class="space-y-2">
          <!-- 资源奖励 -->
          <div class="space-y-1 text-xs">
            <div v-if="reward.resources?.metal" class="flex items-center gap-1">
              <div class="h-2 w-2 rounded-full bg-amber-600" />
              <span>{{ formatNumber(reward.resources.metal) }} {{ t('resources.metal') }}</span>
            </div>
            <div v-if="reward.resources?.crystal" class="flex items-center gap-1">
              <div class="h-2 w-2 rounded-full bg-cyan-500" />
              <span>{{ formatNumber(reward.resources.crystal) }} {{ t('resources.crystal') }}</span>
            </div>
            <div v-if="reward.resources?.deuterium" class="flex items-center gap-1">
              <div class="h-2 w-2 rounded-full bg-green-500" />
              <span>{{ formatNumber(reward.resources.deuterium) }} {{ t('resources.deuterium') }}</span>
            </div>
            <div v-if="reward.resources?.darkMatter" class="flex items-center gap-1">
              <Sparkles class="h-3 w-3 text-purple-500" />
              <span>{{ formatNumber(reward.resources.darkMatter) }} {{ t('resources.darkMatter') }}</span>
            </div>
          </div>

          <!-- 领取按钮 -->
          <Button
            v-if="index + 1 === currentDay && canCheckIn"
            class="w-full mt-2"
            size="sm"
            @click="handleCheckIn"
          >
            <Gift class="mr-2 h-4 w-4" />
            {{ t('checkIn.claim') }}
          </Button>
          <div v-else-if="index + 1 < currentDay || (index + 1 === currentDay && !canCheckIn)" class="text-xs text-center text-muted-foreground mt-2">
            {{ t('checkIn.claimed') }}
          </div>
          <div v-else class="text-xs text-center text-muted-foreground mt-2">
            {{ t('checkIn.locked') }}
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- 说明 -->
    <Card class="bg-muted/50">
      <CardContent class="p-4">
        <h3 class="font-medium mb-2">{{ t('checkIn.rules') }}</h3>
        <ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
          <li>{{ t('checkIn.rule1') }}</li>
          <li>{{ t('checkIn.rule2') }}</li>
          <li>{{ t('checkIn.rule3') }}</li>
        </ul>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import { useGameStore } from '@/stores/gameStore'
  import { useI18n } from '@/composables/useI18n'
  import { formatNumber } from '@/utils/format'
  import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
  import { Badge } from '@/components/ui/badge'
  import { Button } from '@/components/ui/button'
  import { Progress } from '@/components/ui/progress'
  import { Calendar, CheckCircle, Gift, Sparkles } from 'lucide-vue-next'
  import { CHECK_IN_REWARDS } from '@/config/checkInConfig'
  import {
    canCheckInToday,
    getCurrentCheckInDay,
    getCheckInProgress,
    claimCheckIn
  } from '@/logic/checkInLogic'
  import { toast } from 'vue-sonner'

  const gameStore = useGameStore()
  const { t } = useI18n()

  const currentDay = computed(() => getCurrentCheckInDay(gameStore.player))
  const canCheckIn = computed(() => canCheckInToday(gameStore.player))
  const progress = computed(() => getCheckInProgress(gameStore.player))

  const handleCheckIn = () => {
    const planet = gameStore.currentPlanet
    if (!planet) {
      toast.error(t('checkIn.noPlanet'))
      return
    }

    const result = claimCheckIn(gameStore.player, planet)
    if (result.success) {
      toast.success(result.message)
    } else {
      toast.error(result.message)
    }
  }
</script>
