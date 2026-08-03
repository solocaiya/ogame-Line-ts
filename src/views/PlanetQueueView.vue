<template>
  <div class="container mx-auto p-4 sm:p-6 space-y-6">
    <h1 class="text-2xl sm:text-3xl font-bold">{{ t('queueManagement.title') }}</h1>

    <!-- 星球选择器 -->
    <Tabs v-model="selectedPlanetId" class="w-full">
      <TabsList class="w-full flex-wrap h-auto" :tab-count="planets.length">
        <TabsTrigger
          v-for="planet in planets"
          :key="planet.id"
          :value="planet.id"
          class="flex items-center gap-1.5"
        >
          <component :is="planet.isMoon ? Moon : Globe" class="h-3.5 w-3.5" />
          <span class="truncate max-w-[100px] sm:max-w-none">{{ planet.name }}</span>
          <Badge v-if="(planet.buildQueue?.length || 0) + (planet.waitingBuildQueue?.length || 0) > 0" variant="secondary" class="text-xs ml-1">
            {{ (planet.buildQueue?.length || 0) + (planet.waitingBuildQueue?.length || 0) }}
          </Badge>
        </TabsTrigger>
      </TabsList>

      <!-- 每个星球的队列内容 -->
      <TabsContent v-for="planet in planets" :key="planet.id" :value="planet.id" class="mt-4 space-y-4">
        <!-- 快捷操作 -->
        <div class="flex items-center justify-between gap-2 flex-wrap">
          <div class="flex items-center gap-2">
            <Button variant="outline" size="sm" @click="togglePause(planet)" class="gap-2">
              <component :is="planet.queuePaused ? Play : Pause" class="h-4 w-4" />
              {{ planet.queuePaused ? t('queueManagement.resume') : t('queueManagement.pause') }}
            </Button>
            <Button
              v-if="planet.waitingBuildQueue && planet.waitingBuildQueue.length > 0"
              variant="outline"
              size="sm"
              class="gap-2 text-destructive"
              @click="cancelAllWaiting(planet.id)"
            >
              <X class="h-4 w-4" />
              {{ t('queueManagement.cancelAll') }}
            </Button>
          </div>
          <div class="text-sm text-muted-foreground">
            {{ t('queueManagement.active') }}: {{ planet.buildQueue?.length || 0 }} /
            {{ t('queueManagement.waiting') }}: {{ planet.waitingBuildQueue?.length || 0 }}
          </div>
        </div>

        <!-- 活跃队列 -->
        <Card>
          <CardHeader>
            <CardTitle class="text-base flex items-center gap-2">
              <Clock class="h-4 w-4" />
              {{ t('queueManagement.activeQueue') }}
            </CardTitle>
          </CardHeader>
          <CardContent class="space-y-3">
            <div v-if="!planet.buildQueue || planet.buildQueue.length === 0" class="text-center py-4 text-muted-foreground text-sm">
              {{ t('queueManagement.emptyActive') }}
            </div>
            <div v-else class="space-y-3">
              <div
                v-for="(item, index) in planet.buildQueue"
                :key="item.id"
                class="p-3 bg-muted/50 rounded-lg border"
              >
                <div class="flex items-center justify-between gap-2 mb-2">
                  <div class="flex items-center gap-2 min-w-0">
                    <Badge variant="outline" class="text-xs shrink-0">{{ getTypeLabel(item.type) }}</Badge>
                    <span class="text-sm font-medium truncate">{{ getItemName(item, planet) }}</span>
                  </div>
                  <span class="text-xs text-muted-foreground shrink-0">
                    {{ getRemainingTime(item.endTime) }}
                  </span>
                </div>
                <!-- 进度条 -->
                <div class="w-full bg-secondary rounded-full h-2">
                  <div
                    class="bg-primary h-2 rounded-full transition-all"
                    :style="{ width: getProgress(item) + '%' }"
                  />
                </div>
                <!-- 取消按钮 -->
                <div v-if="index === 0" class="mt-2 flex justify-end">
                  <Button variant="ghost" size="sm" class="text-xs text-destructive gap-1" @click="cancelActiveItem(planet.id, item.id)">
                    <X class="h-3 w-3" />
                    {{ t('queueManagement.cancel') }}
                  </Button>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- 等待队列 -->
        <Card>
          <CardHeader>
            <CardTitle class="text-base flex items-center gap-2">
              <ListOrdered class="h-4 w-4" />
              {{ t('queueManagement.waitingQueue') }}
            </CardTitle>
          </CardHeader>
          <CardContent class="space-y-2">
            <div v-if="!planet.waitingBuildQueue || planet.waitingBuildQueue.length === 0" class="text-center py-4 text-muted-foreground text-sm">
              {{ t('queueManagement.emptyWaiting') }}
            </div>
            <div v-else class="space-y-2">
              <div
                v-for="(item, index) in planet.waitingBuildQueue"
                :key="item.id"
                draggable="true"
                class="flex items-center gap-2 p-3 bg-muted/30 rounded-lg border hover:bg-muted/50 transition-colors cursor-grab active:cursor-grabbing"
                :class="{ 'opacity-50': dragItemId === item.id, 'border-primary border-2': dragOverIndex === index && dragItemId !== item.id }"
                @dragstart="onDragStart($event, item.id)"
                @dragover.prevent="onDragOver($event, index)"
                @dragend="onDragEnd(planet.id)"
                @drop.prevent="onDrop(planet.id, index)"
              >
                <!-- 拖拽手柄 -->
                <GripVertical class="h-4 w-4 text-muted-foreground/50 shrink-0" />

                <!-- 序号 -->
                <span class="text-xs text-muted-foreground w-6 text-center shrink-0">{{ index + 1 }}</span>

                <!-- 信息 -->
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2">
                    <Badge variant="outline" class="text-xs shrink-0">{{ getTypeLabel(item.type) }}</Badge>
                    <span class="text-sm truncate">{{ getItemName(item, planet) }}</span>
                  </div>
                  <div class="text-xs text-muted-foreground mt-1">
                    {{ getItemCostText(item) }}
                  </div>
                </div>

                <!-- 操作按钮 -->
                <div class="flex items-center gap-1 shrink-0">
                  <Button
                    variant="ghost"
                    size="icon"
                    class="h-7 w-7"
                    :disabled="index === 0"
                    @click="moveUp(planet.id, item.id)"
                  >
                    <ChevronUp class="h-4 w-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    class="h-7 w-7"
                    :disabled="index === planet.waitingBuildQueue!.length - 1"
                    @click="moveDown(planet.id, item.id)"
                  >
                    <ChevronDown class="h-4 w-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    class="h-7 w-7 text-destructive"
                    @click="removeWaitingItem(planet.id, item.id)"
                  >
                    <X class="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>

    <!-- 跨星球概览 -->
    <Card>
      <CardHeader>
        <CardTitle class="text-base flex items-center gap-2">
          <Globe class="h-4 w-4" />
          {{ t('queueManagement.globalOverview') }}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b">
                <th class="text-left py-2 px-2 font-medium text-muted-foreground">{{ t('queueManagement.planet') }}</th>
                <th class="text-center py-2 px-2 font-medium text-muted-foreground">{{ t('queueManagement.active') }}</th>
                <th class="text-center py-2 px-2 font-medium text-muted-foreground">{{ t('queueManagement.waiting') }}</th>
                <th class="text-center py-2 px-2 font-medium text-muted-foreground">{{ t('queueManagement.status') }}</th>
                <th class="text-right py-2 px-2 font-medium text-muted-foreground">{{ t('queueManagement.estimatedTime') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="summary in queueSummary.planetSummaries.value" :key="summary.planetId" class="border-b last:border-0 hover:bg-muted/30">
                <td class="py-2 px-2">
                  <div class="flex items-center gap-1.5">
                    <component :is="summary.isMoon ? Moon : Globe" class="h-3.5 w-3.5 text-muted-foreground" />
                    <span class="truncate max-w-[120px]">{{ summary.planetName }}</span>
                  </div>
                </td>
                <td class="text-center py-2 px-2">{{ summary.activeCount }}</td>
                <td class="text-center py-2 px-2">{{ summary.waitingCount }}</td>
                <td class="text-center py-2 px-2">
                  <Badge :variant="summary.queuePaused ? 'secondary' : 'default'" class="text-xs">
                    {{ summary.queuePaused ? t('queueManagement.paused') : t('queueManagement.running') }}
                  </Badge>
                </td>
                <td class="text-right py-2 px-2 text-xs text-muted-foreground">
                  {{ summary.activeCount > 0 ? formatTimeRemaining(summary.estimatedCompletionTime) : '-' }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 汇总 -->
        <div class="mt-4 pt-4 border-t grid grid-cols-2 sm:grid-cols-3 gap-4 text-center">
          <div>
            <div class="text-xs text-muted-foreground">{{ t('queueManagement.totalActive') }}</div>
            <div class="text-lg font-bold">{{ queueSummary.totalActiveItems.value }}</div>
          </div>
          <div>
            <div class="text-xs text-muted-foreground">{{ t('queueManagement.totalWaiting') }}</div>
            <div class="text-lg font-bold">{{ queueSummary.totalWaitingItems.value }}</div>
          </div>
          <div class="col-span-2 sm:col-span-1">
            <div class="text-xs text-muted-foreground">{{ t('queueManagement.globalEstimate') }}</div>
            <div class="text-lg font-bold">{{ formatTimeRemaining(queueSummary.globalEstimatedCompletion.value) }}</div>
          </div>
        </div>

        <!-- 等待队列资源需求汇总 -->
        <div v-if="queueSummary.totalWaitingItems.value > 0" class="mt-4 pt-4 border-t">
          <div class="text-xs text-muted-foreground text-center mb-2">{{ t('queueManagement.resourceSummary') }}</div>
          <div class="flex justify-center gap-6 text-sm">
            <span class="text-amber-600 dark:text-amber-400">
              {{ t('resources.metal') }}: {{ formatNumber(queueSummary.waitingQueueResourceEstimate.value.metal) }}
            </span>
            <span class="text-cyan-500">
              {{ t('resources.crystal') }}: {{ formatNumber(queueSummary.waitingQueueResourceEstimate.value.crystal) }}
            </span>
            <span class="text-green-500">
              {{ t('resources.deuterium') }}: {{ formatNumber(queueSummary.waitingQueueResourceEstimate.value.deuterium) }}
            </span>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed } from 'vue'
  import { useI18n } from '@/composables/useI18n'
  import { useQueueSummary } from '@/composables/useQueueSummary'
  import { useGameStore } from '@/stores/gameStore'
  import { useGameConfig } from '@/composables/useGameConfig'
  import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
  import { Button } from '@/components/ui/button'
  import { Badge } from '@/components/ui/badge'
  import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
  import { formatNumber } from '@/utils/format'
  import {
    Moon, Globe, Clock, ListOrdered, Play, Pause, X, ChevronUp, ChevronDown, GripVertical
  } from 'lucide-vue-next'
  import type { Planet, BuildQueueItem, WaitingQueueItem } from '@/types/game'
  import { BuildingType, ShipType, DefenseType, TechnologyType } from '@/types/game'
  import * as waitingQueueLogic from '@/logic/waitingQueueLogic'

  const { t } = useI18n()
  const gameStore = useGameStore()
  const queueSummary = useQueueSummary()
  const { SHIPS, DEFENSES, BUILDINGS, TECHNOLOGIES } = useGameConfig()

  const planets = computed(() => gameStore.player.planets)
  const selectedPlanetId = ref(planets.value.length > 0 ? planets.value[0].id : '')

  const getTypeLabel = (type: string) => {
    switch (type) {
      case 'building': return t('queueManagement.building')
      case 'ship': return t('queueManagement.ship')
      case 'defense': return t('queueManagement.defense')
      case 'demolish': return t('queueManagement.demolish')
      case 'scrap_ship': return t('queueManagement.scrapShip')
      default: return type
    }
  }

  const getItemName = (item: BuildQueueItem | WaitingQueueItem, _planet: Planet) => {
    const itemType = item.itemType
    if (item.type === 'building' || item.type === 'demolish') {
      const config = BUILDINGS[itemType as BuildingType]
      return config?.name || itemType
    }
    if (item.type === 'ship' || item.type === 'scrap_ship') {
      const config = SHIPS[itemType as ShipType]
      const qty = item.quantity ? ` ×${item.quantity}` : ''
      return (config?.name || itemType) + qty
    }
    if (item.type === 'defense') {
      const config = DEFENSES[itemType as DefenseType]
      const qty = item.quantity ? ` ×${item.quantity}` : ''
      return (config?.name || itemType) + qty
    }
    if (item.type === 'technology') {
      const config = TECHNOLOGIES[itemType as TechnologyType]
      return config?.name || itemType
    }
    return itemType
  }

  const getItemCostText = (item: WaitingQueueItem) => {
    const cost = waitingQueueLogic.calculateWaitingItemCost(item)
    const parts: string[] = []
    if (cost.metal > 0) parts.push(`${t('resources.metal')}: ${formatNumber(cost.metal)}`)
    if (cost.crystal > 0) parts.push(`${t('resources.crystal')}: ${formatNumber(cost.crystal)}`)
    if (cost.deuterium > 0) parts.push(`${t('resources.deuterium')}: ${formatNumber(cost.deuterium)}`)
    return parts.join(' / ')
  }

  const getRemainingTime = (endTime: number) => {
    const remaining = Math.max(0, endTime - Date.now())
    const seconds = Math.floor(remaining / 1000)
    if (seconds < 60) return `${seconds}s`
    const minutes = Math.floor(seconds / 60)
    if (minutes < 60) return `${minutes}m ${seconds % 60}s`
    const hours = Math.floor(minutes / 60)
    return `${hours}h ${minutes % 60}m`
  }

  const getProgress = (item: BuildQueueItem) => {
    const now = Date.now()
    const total = item.endTime - item.startTime
    const elapsed = now - item.startTime
    return Math.min(100, Math.max(0, (elapsed / total) * 100))
  }

  const formatTimeRemaining = (timestamp: number) => {
    const diff = Math.max(0, timestamp - Date.now())
    const seconds = Math.floor(diff / 1000)
    if (seconds < 60) return `${seconds}s`
    const minutes = Math.floor(seconds / 60)
    if (minutes < 60) return `${minutes}m`
    const hours = Math.floor(minutes / 60)
    if (hours < 24) return `${hours}h ${minutes % 60}m`
    const days = Math.floor(hours / 24)
    return `${days}d ${hours % 24}h`
  }

  const togglePause = (planet: Planet) => {
    gameStore.toggleQueuePaused(planet.id)
  }

  const cancelActiveItem = (planetId: string, itemId: string) => {
    gameStore.cancelBuildQueueItem(planetId, itemId)
  }

  const moveUp = (planetId: string, itemId: string) => {
    const planet = planets.value.find(p => p.id === planetId)
    if (planet) {
      waitingQueueLogic.moveWaitingQueueItemUp(planet, itemId)
    }
  }

  const moveDown = (planetId: string, itemId: string) => {
    const planet = planets.value.find(p => p.id === planetId)
    if (planet) {
      waitingQueueLogic.moveWaitingQueueItemDown(planet, itemId)
    }
  }

  const removeWaitingItem = (planetId: string, itemId: string) => {
    const planet = planets.value.find(p => p.id === planetId)
    if (planet) {
      waitingQueueLogic.removeFromBuildWaitingQueue(planet, itemId)
    }
  }

  const cancelAllWaiting = (planetId: string) => {
    const planet = planets.value.find(p => p.id === planetId)
    if (planet && planet.waitingBuildQueue) {
      const itemIds = planet.waitingBuildQueue.map(item => item.id)
      for (const id of itemIds) {
        waitingQueueLogic.removeFromBuildWaitingQueue(planet, id)
      }
    }
  }

  // 拖拽重排状态
  const dragItemId = ref<string | null>(null)
  const dragOverIndex = ref<number | null>(null)

  const onDragStart = (_event: DragEvent, itemId: string) => {
    dragItemId.value = itemId
  }

  const onDragOver = (_event: DragEvent, index: number) => {
    dragOverIndex.value = index
  }

  const onDragEnd = (planetId: string) => {
    if (dragItemId.value && dragOverIndex.value !== null) {
      const planet = planets.value.find(p => p.id === planetId)
      if (planet && planet.waitingBuildQueue) {
        const fromIndex = planet.waitingBuildQueue.findIndex(item => item.id === dragItemId.value)
        const toIndex = dragOverIndex.value
        if (fromIndex !== -1 && fromIndex !== toIndex) {
          waitingQueueLogic.reorderWaitingQueueItem(planet, dragItemId.value, toIndex)
        }
      }
    }
    dragItemId.value = null
    dragOverIndex.value = null
  }

  const onDrop = (_planetId: string, _index: number) => {
    // drop 事件由 onDragEnd 处理
  }
</script>
