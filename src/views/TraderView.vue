<template>
  <div class="container mx-auto p-4 sm:p-6 space-y-6">
    <h1 class="text-2xl sm:text-3xl font-bold">{{ t('trader.title') }}</h1>

    <!-- 汇率说明 -->
    <Card>
      <CardHeader>
        <CardTitle class="text-lg">{{ t('trader.exchangeRates') }}</CardTitle>
        <CardDescription>{{ t('trader.exchangeRatesDesc') }}</CardDescription>
      </CardHeader>
      <CardContent>
        <div class="grid grid-cols-3 gap-4 text-center">
          <div class="p-3 bg-amber-50 dark:bg-amber-950/30 rounded-lg">
            <div class="text-xs text-muted-foreground">{{ t('resources.metal') }}</div>
            <div class="text-lg font-bold text-amber-600">1 : {{ TRADER_RATES.metal }}</div>
          </div>
          <div class="p-3 bg-cyan-50 dark:bg-cyan-950/30 rounded-lg">
            <div class="text-xs text-muted-foreground">{{ t('resources.crystal') }}</div>
            <div class="text-lg font-bold text-cyan-500">1 : {{ TRADER_RATES.crystal }}</div>
          </div>
          <div class="p-3 bg-green-50 dark:bg-green-950/30 rounded-lg">
            <div class="text-xs text-muted-foreground">{{ t('resources.deuterium') }}</div>
            <div class="text-lg font-bold text-green-500">1 : {{ TRADER_RATES.deuterium }}</div>
          </div>
        </div>
        <p class="text-xs text-muted-foreground mt-3 text-center">
          {{ t('trader.feeNotice', { fee: TRADER_FEE_RATE * 100 }) }}
        </p>
      </CardContent>
    </Card>

    <!-- 交易面板 -->
    <Card>
      <CardHeader>
        <CardTitle class="text-lg">{{ t('trader.tradePanel') }}</CardTitle>
        <CardDescription>
          {{ t('trader.currentDarkMatter') }}:
          <span class="font-bold text-purple-500">{{ formatNumber(planet?.resources.darkMatter ?? 0) }}</span>
        </CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <!-- 暗物质输入 -->
        <div class="space-y-2">
          <Label>{{ t('trader.darkMatterAmount') }}</Label>
          <div class="flex gap-2">
            <Input
              v-model.number="darkMatterInput"
              type="number"
              :min="0"
              :max="TRADER_MAX_DARK_MATTER"
              :placeholder="t('trader.enterAmount')"
              class="flex-1"
            />
            <Button variant="outline" size="sm" @click="darkMatterInput = 100">100</Button>
            <Button variant="outline" size="sm" @click="darkMatterInput = 1000">1K</Button>
            <Button variant="outline" size="sm" @click="darkMatterInput = 10000">10K</Button>
          </div>
          <p class="text-xs text-muted-foreground">
            {{ t('trader.maxTrade') }}: {{ formatNumber(TRADER_MAX_DARK_MATTER) }}
          </p>
        </div>

        <!-- 预览 -->
        <div v-if="darkMatterInput > 0" class="p-4 bg-muted rounded-lg space-y-2">
          <p class="text-sm font-medium">{{ t('trader.preview') }}</p>
          <div class="grid grid-cols-3 gap-3 text-sm">
            <div v-for="res in resourceTypes" :key="res.key" class="text-center">
              <div class="text-xs text-muted-foreground">{{ t(`resources.${res.key}`) }}</div>
              <div class="font-bold" :class="res.color">
                +{{ formatNumber(getTradePreview(res.key).netResource) }}
              </div>
              <div class="text-xs text-muted-foreground">
                {{ t('trader.fee') }}: {{ formatNumber(getTradePreview(res.key).fee) }}
              </div>
            </div>
          </div>
        </div>

        <!-- 交易按钮 -->
        <div class="grid grid-cols-3 gap-3">
          <Button
            v-for="res in resourceTypes"
            :key="res.key"
            :disabled="!canTrade(res.key)"
            @click="handleTrade(res.key)"
            class="w-full"
          >
            <ArrowLeftRight class="mr-2 h-4 w-4" />
            {{ t(`resources.${res.key}`) }}
          </Button>
        </div>
      </CardContent>
    </Card>

    <!-- 交易历史 -->
    <Card>
      <CardHeader>
        <CardTitle class="text-lg">{{ t('trader.history') }}</CardTitle>
      </CardHeader>
      <CardContent>
        <div v-if="!tradeHistory.length" class="text-center text-muted-foreground py-8">
          {{ t('trader.noHistory') }}
        </div>
        <div v-else class="space-y-2 max-h-80 overflow-y-auto">
          <div
            v-for="record in tradeHistory"
            :key="record.id"
            class="flex items-center justify-between p-3 bg-muted/50 rounded-lg text-sm"
          >
            <div class="flex items-center gap-3">
              <div class="h-2 w-2 rounded-full" :class="getResourceColor(record.resourceType)" />
              <div>
                <span class="font-medium">-{{ formatNumber(record.darkMatterSpent) }} {{ t('resources.darkMatter') }}</span>
                <span class="text-muted-foreground mx-2">→</span>
                <span class="font-medium">+{{ formatNumber(record.resourceGained) }} {{ t(`resources.${record.resourceType}`) }}</span>
              </div>
            </div>
            <div class="text-xs text-muted-foreground">
              {{ formatTime(record.timestamp) }}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed } from 'vue'
  import { useGameStore } from '@/stores/gameStore'
  import { useI18n } from '@/composables/useI18n'
  import { formatNumber } from '@/utils/format'
  import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
  import { Button } from '@/components/ui/button'
  import { Input } from '@/components/ui/input'
  import { Label } from '@/components/ui/label'
  import { ArrowLeftRight } from 'lucide-vue-next'
  import { TRADER_RATES, TRADER_FEE_RATE, TRADER_MAX_DARK_MATTER, calculateTrade, executeTrade } from '@/logic/traderLogic'
  import { playSound, SoundType } from '@/logic/soundManager'
  import { toast } from 'vue-sonner'

  const gameStore = useGameStore()
  const { t } = useI18n()

  const darkMatterInput = ref(100)

  const planet = computed(() => gameStore.currentPlanet)
  const tradeHistory = computed(() => gameStore.player.tradeHistory || [])

  const resourceTypes = [
    { key: 'metal' as const, color: 'text-amber-600' },
    { key: 'crystal' as const, color: 'text-cyan-500' },
    { key: 'deuterium' as const, color: 'text-green-500' }
  ]

  const getTradePreview = (resourceType: 'metal' | 'crystal' | 'deuterium') => {
    return calculateTrade(darkMatterInput.value, resourceType)
  }

  const canTrade = (resourceType: 'metal' | 'crystal' | 'deuterium') => {
    if (!planet.value) return false
    if (darkMatterInput.value <= 0) return false
    if (darkMatterInput.value > TRADER_MAX_DARK_MATTER) return false
    if (planet.value.resources.darkMatter < darkMatterInput.value) return false
    return true
  }

  const getResourceColor = (type: string) => {
    if (type === 'metal') return 'bg-amber-600'
    if (type === 'crystal') return 'bg-cyan-500'
    return 'bg-green-500'
  }

  const handleTrade = (resourceType: 'metal' | 'crystal' | 'deuterium') => {
    if (!planet.value) {
      toast.error(t('trader.noPlanet'))
      return
    }

    const result = executeTrade(gameStore.player, planet.value, darkMatterInput.value, resourceType)
    if (result.success) {
      toast.success(result.message)
      playSound(SoundType.TradeComplete, { enabled: gameStore.player.soundEnabled !== false, volume: gameStore.player.soundVolume ?? 0.7 })
    } else {
      toast.error(result.message)
    }
  }

  const formatTime = (ts: number) => {
    const d = new Date(ts)
    return `${d.getMonth() + 1}/${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  }
</script>
