<template>
  <div class="container mx-auto p-4 sm:p-6 space-y-6">
    <h1 class="text-2xl sm:text-3xl font-bold">{{ t('battleReports.title') }}</h1>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
      <Card>
        <CardContent class="p-4 text-center">
          <div class="text-xs text-muted-foreground mb-1">{{ t('battleReports.totalBattles') }}</div>
          <div class="text-2xl font-bold">{{ stats.totalBattles.value }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="p-4 text-center">
          <div class="text-xs text-muted-foreground mb-1">{{ t('battleReports.winRate') }}</div>
          <div class="text-2xl font-bold text-green-500">{{ stats.winRate.value }}%</div>
          <div class="text-xs text-muted-foreground mt-1">
            {{ stats.wins.value }}{{ t('battleReports.winShort') }} /
            {{ stats.losses.value }}{{ t('battleReports.lossShort') }} /
            {{ stats.draws.value }}{{ t('battleReports.drawShort') }}
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="p-4 text-center">
          <div class="text-xs text-muted-foreground mb-1">{{ t('battleReports.totalPlunder') }}</div>
          <div class="text-lg font-bold text-amber-500">{{ formatNumber(stats.totalPlunder.value.metal + stats.totalPlunder.value.crystal + stats.totalPlunder.value.deuterium) }}</div>
          <div class="text-xs text-muted-foreground mt-1">
            <span class="text-amber-600">{{ formatNumber(stats.totalPlunder.value.metal) }}</span> /
            <span class="text-cyan-500">{{ formatNumber(stats.totalPlunder.value.crystal) }}</span> /
            <span class="text-green-500">{{ formatNumber(stats.totalPlunder.value.deuterium) }}</span>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="p-4 text-center">
          <div class="text-xs text-muted-foreground mb-1">{{ t('battleReports.totalLosses') }}</div>
          <div class="text-lg font-bold text-red-500">{{ formatNumber(stats.totalAttackerLosses.value + stats.totalDefenderLosses.value) }}</div>
          <div class="text-xs text-muted-foreground mt-1">
            {{ t('simulatorView.attacker') }}: {{ formatNumber(stats.totalAttackerLosses.value) }} /
            {{ t('simulatorView.defender') }}: {{ formatNumber(stats.totalDefenderLosses.value) }}
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- 最近战斗趋势 -->
    <Card v-if="stats.recentTrendSummary.value.total > 0">
      <CardContent class="p-4">
        <div class="text-xs text-muted-foreground mb-2">{{ t('battleReports.recentTrend') }}</div>
        <div class="flex items-center gap-3 flex-wrap">
          <!-- 趋势圆点 -->
          <div class="flex gap-1.5">
            <div
              v-for="(result, idx) in stats.recentTrend.value"
              :key="idx"
              class="w-3 h-3 rounded-full"
              :class="{
                'bg-green-500': result === 'win',
                'bg-red-500': result === 'loss',
                'bg-gray-400': result === 'draw'
              }"
              :title="result === 'win' ? t('battleReports.victory') : result === 'loss' ? t('battleReports.defeat') : t('battleReports.draw')"
            />
          </div>
          <!-- 趋势统计 -->
          <div class="flex gap-3 text-xs">
            <span class="text-green-500">{{ stats.recentTrendSummary.value.wins }}{{ t('battleReports.winShort') }}</span>
            <span class="text-red-500">{{ stats.recentTrendSummary.value.losses }}{{ t('battleReports.lossShort') }}</span>
            <span class="text-gray-400">{{ stats.recentTrendSummary.value.draws }}{{ t('battleReports.drawShort') }}</span>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- 筛选栏 -->
    <Card>
      <CardContent class="p-4">
        <div class="flex flex-wrap gap-4">
          <!-- 结果筛选 -->
          <div class="flex items-center gap-2">
            <span class="text-sm text-muted-foreground">{{ t('battleReports.filterResult') }}:</span>
            <div class="flex gap-1">
              <Button
                v-for="opt in resultOptions"
                :key="opt.value"
                :variant="resultFilter === opt.value ? 'default' : 'outline'"
                size="sm"
                @click="resultFilter = opt.value"
                class="text-xs"
              >
                {{ opt.label }}
              </Button>
            </div>
          </div>
          <!-- 时间筛选 -->
          <div class="flex items-center gap-2">
            <span class="text-sm text-muted-foreground">{{ t('battleReports.filterTime') }}:</span>
            <div class="flex gap-1">
              <Button
                v-for="opt in timeOptions"
                :key="opt.value"
                :variant="timeFilter === opt.value ? 'default' : 'outline'"
                size="sm"
                @click="timeFilter = opt.value"
                class="text-xs"
              >
                {{ opt.label }}
              </Button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- 战报列表 -->
    <div v-if="paginatedReports.length === 0" class="border rounded-lg p-8 text-center">
      <Swords class="h-10 w-10 text-muted-foreground mx-auto mb-3" />
      <p class="text-muted-foreground">{{ t('battleReports.noReports') }}</p>
    </div>

    <div v-else class="space-y-2">
      <Card
        v-for="report in paginatedReports"
        :key="report.id"
        class="cursor-pointer hover:shadow-md transition-shadow"
        @click="openReport(report)"
      >
        <CardContent class="p-4">
          <div class="flex items-center gap-3 flex-wrap">
            <!-- 时间 -->
            <span class="text-xs text-muted-foreground shrink-0">{{ formatDate(report.timestamp) }}</span>

            <!-- 结果徽章 -->
            <Badge :variant="getResultBadgeVariant(report)" class="text-xs shrink-0">
              {{ getResultText(report) }}
            </Badge>

            <!-- 对手信息 -->
            <span class="text-sm truncate flex-1 min-w-0">
              {{ getOpponentName(report) }}
            </span>

            <!-- 损失摘要 -->
            <span class="text-xs text-muted-foreground shrink-0 hidden sm:inline">
              {{ t('battleReports.lossShort2') }}: {{ formatNumber(getTotalLoss(report)) }}
            </span>

            <!-- 掠夺摘要 -->
            <span v-if="report.plunder && (report.plunder.metal > 0 || report.plunder.crystal > 0 || report.plunder.deuterium > 0)" class="text-xs text-amber-500 shrink-0 hidden sm:inline">
              +{{ formatNumber(report.plunder.metal + report.plunder.crystal + report.plunder.deuterium) }}
            </span>

            <!-- 回合数 -->
            <span v-if="report.rounds" class="text-xs text-muted-foreground shrink-0">
              {{ report.rounds }}{{ t('battleReports.rounds') }}
            </span>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- 分页 -->
    <div v-if="totalPages > 1" class="flex justify-center">
      <div class="flex items-center gap-2">
        <Button variant="outline" size="sm" :disabled="currentPage <= 1" @click="currentPage--">
          {{ t('battleReports.prevPage') }}
        </Button>
        <span class="text-sm text-muted-foreground">
          {{ currentPage }} / {{ totalPages }}
        </span>
        <Button variant="outline" size="sm" :disabled="currentPage >= totalPages" @click="currentPage++">
          {{ t('battleReports.nextPage') }}
        </Button>
      </div>
    </div>

    <!-- 战报详情弹窗 -->
    <BattleReportDialog :report="selectedReport" :open="dialogOpen" @update:open="dialogOpen = $event" />
  </div>
</template>

<script setup lang="ts">
  import { ref, computed } from 'vue'
  import { useI18n } from '@/composables/useI18n'
  import { useBattleStats } from '@/composables/useBattleStats'
  import { useGameStore } from '@/stores/gameStore'
  import { useUniverseStore } from '@/stores/universeStore'
  import { Card, CardContent } from '@/components/ui/card'
  import { Button } from '@/components/ui/button'
  import { Badge } from '@/components/ui/badge'
  import BattleReportDialog from '@/components/dialogs/BattleReportDialog.vue'
  import { formatNumber, formatDate } from '@/utils/format'
  import { Swords } from 'lucide-vue-next'
  import type { BattleResult } from '@/types/game'

  const { t } = useI18n()
  const stats = useBattleStats()
  const gameStore = useGameStore()
  const universeStore = useUniverseStore()

  const resultFilter = ref<'all' | 'win' | 'loss' | 'draw'>('all')
  const timeFilter = ref<'all' | 'today' | 'week' | 'month'>('all')
  const currentPage = ref(1)
  const pageSize = 15

  const selectedReport = ref<BattleResult | null>(null)
  const dialogOpen = ref(false)

  const resultOptions = computed(() => [
    { value: 'all' as const, label: t('battleReports.all') },
    { value: 'win' as const, label: t('battleReports.victory') },
    { value: 'loss' as const, label: t('battleReports.defeat') },
    { value: 'draw' as const, label: t('battleReports.draw') }
  ])

  const timeOptions = computed(() => [
    { value: 'all' as const, label: t('battleReports.allTime') },
    { value: 'today' as const, label: t('battleReports.today') },
    { value: 'week' as const, label: t('battleReports.thisWeek') },
    { value: 'month' as const, label: t('battleReports.thisMonth') }
  ])

  const filteredReports = computed(() => {
    return stats.getFilteredReports(resultFilter.value, timeFilter.value)
      .sort((a, b) => b.timestamp - a.timestamp)
  })

  const totalPages = computed(() => Math.max(1, Math.ceil(filteredReports.value.length / pageSize)))

  const paginatedReports = computed(() => {
    const start = (currentPage.value - 1) * pageSize
    return filteredReports.value.slice(start, start + pageSize)
  })

  const openReport = (report: BattleResult) => {
    selectedReport.value = report
    dialogOpen.value = true
    // 标记为已读
    report.read = true
  }

  const getResultBadgeVariant = (report: BattleResult) => {
    const isAttacker = gameStore.player.planets.some(p => p.id === report.attackerPlanetId)
    if (report.winner === 'draw') return 'secondary'
    const isWin = (isAttacker && report.winner === 'attacker') || (!isAttacker && report.winner === 'defender')
    return isWin ? 'default' : 'destructive'
  }

  const getResultText = (report: BattleResult) => {
    const isAttacker = gameStore.player.planets.some(p => p.id === report.attackerPlanetId)
    if (report.winner === 'draw') return t('battleReports.draw')
    const isWin = (isAttacker && report.winner === 'attacker') || (!isAttacker && report.winner === 'defender')
    return isWin ? t('battleReports.victory') : t('battleReports.defeat')
  }

  const getOpponentName = (report: BattleResult) => {
    const isAttacker = gameStore.player.planets.some(p => p.id === report.attackerPlanetId)
    const opponentPlanetId = isAttacker ? report.defenderPlanetId : report.attackerPlanetId
    // 从玩家星球中查找
    const playerPlanet = gameStore.player.planets.find(p => p.id === opponentPlanetId)
    if (playerPlanet) return playerPlanet.name
    // 从宇宙中查找
    const universePlanet = Object.values(universeStore.planets).find(p => p.id === opponentPlanetId)
    if (universePlanet) return universePlanet.name
    return opponentPlanetId || t('battleReports.unknownOpponent')
  }

  const getTotalLoss = (report: BattleResult) => {
    const isAttacker = gameStore.player.planets.some(p => p.id === report.attackerPlanetId)
    if (isAttacker) {
      return Object.values(report.attackerLosses).reduce((sum, c) => sum + c, 0)
    }
    const fleetLoss = Object.values(report.defenderLosses.fleet || {}).reduce((sum, c) => sum + c, 0)
    const defenseLoss = Object.values(report.defenderLosses.defense || {}).reduce((sum, c) => sum + c, 0)
    return fleetLoss + defenseLoss
  }
</script>
