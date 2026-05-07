<template>
  <div v-if="planet" class="container mx-auto p-4 sm:p-6 space-y-4 sm:space-y-6">
    <!-- 星球信息 -->
    <div class="text-center">
      <h1 class="text-2xl sm:text-3xl font-bold mb-1 sm:mb-2 flex items-center justify-center gap-2">
        {{ planet.name }}
        <Badge v-if="planet.isMoon" variant="secondary">{{ t('planet.moon') }}</Badge>
      </h1>
      <p class="text-xs sm:text-sm text-muted-foreground">
        {{ t('planet.position') }}: [{{ planet.position.galaxy }}:{{ planet.position.system }}:{{ planet.position.position }}]
      </p>
      <!-- 温度信息 -->
      <p v-if="planet.temperature && !planet.isMoon" class="text-xs sm:text-sm text-muted-foreground">
        {{ t('planet.temperature') }}: {{ planet.temperature.min }}°C {{ t('common.to') }} {{ planet.temperature.max }}°C
      </p>
      <!-- 月球信息 -->
      <div v-if="!planet.isMoon && moon" class="mt-2">
        <Button @click="switchToMoon" variant="outline" size="sm">
          {{ t('planet.switchToMoon') }}
        </Button>
      </div>
      <div v-if="planet.isMoon" class="mt-2">
        <Button @click="switchToParentPlanet" variant="outline" size="sm">{{ t('planet.backToPlanet') }}</Button>
      </div>
      <!-- 放弃殖民地按钮 -->
      <div v-if="canShowAbandonButton" class="mt-4">
        <Button @click="showAbandonDialog = true" variant="destructive" size="sm">
          {{ t('planet.abandonColony') }}
        </Button>
      </div>
    </div>

    <!-- 放弃殖民地确认对话框 -->
    <AlertDialog :open="showAbandonDialog" @update:open="showAbandonDialog = $event">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{{ t('planet.confirmAbandon') }}</AlertDialogTitle>
          <AlertDialogDescription class="whitespace-pre-line">
            {{ t('planet.abandonWarning', { name: planet.name }) }}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{{ t('common.cancel') }}</AlertDialogCancel>
          <AlertDialogAction @click="handleAbandonColony" class="bg-destructive text-destructive-foreground hover:bg-destructive/90">
            {{ t('planet.confirmAbandonButton') }}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

    <!-- 错误提示对话框 -->
    <AlertDialog :open="showErrorDialog" @update:open="showErrorDialog = $event">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{{ t('planet.abandonFailed') }}</AlertDialogTitle>
          <AlertDialogDescription>
            {{ errorMessage }}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogAction>{{ t('common.confirm') }}</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

    <!-- 资源管理 -->
    <Card>
      <CardHeader>
        <CardTitle>{{ t('overview.resourceOverview') }}</CardTitle>
      </CardHeader>
      <CardContent>
        <Tabs default-value="overview" class="w-full">
          <TabsList class="grid w-full grid-cols-3">
            <TabsTrigger value="overview">{{ t('overview.tabOverview') }}</TabsTrigger>
            <TabsTrigger value="production">{{ t('overview.tabProduction') }}</TabsTrigger>
            <TabsTrigger value="consumption">{{ t('overview.tabConsumption') }}</TabsTrigger>
          </TabsList>
          <!-- 概览标签页 -->
          <TabsContent value="overview" class="mt-4">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{{ t('common.resourceType') }}</TableHead>
                  <TableHead class="text-right">{{ t('resources.current') }}</TableHead>
                  <TableHead class="text-right">{{ t('resources.max') }}</TableHead>
                  <TableHead class="text-right">{{ t('resources.production') }}{{ t('resources.perHour') }}</TableHead>
                  <TableHead class="text-right">{{ t('resources.consumption') }}{{ t('resources.perHour') }}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="resourceType in resourceTypes" :key="resourceType.key">
                  <TableCell class="font-medium">
                    <div class="flex items-center gap-2">
                      <ResourceIcon :type="resourceType.key" size="sm" />
                      {{ t(`resources.${resourceType.key}`) }}
                    </div>
                  </TableCell>
                  <TableCell
                    class="text-right"
                    :class="getResourceColor(planet.resources[resourceType.key], capacity?.[resourceType.key] || Infinity)"
                  >
                    {{ formatNumber(planet.resources[resourceType.key]) }}
                  </TableCell>
                  <TableCell class="text-right text-muted-foreground">
                    {{ formatNumber(capacity?.[resourceType.key] || 0) }}
                  </TableCell>
                  <TableCell class="text-right text-green-600 dark:text-green-400">
                    +{{ formatNumber(production?.[resourceType.key] || 0) }}
                  </TableCell>
                  <TableCell class="text-right text-red-600 dark:text-red-400">
                    <template v-if="resourceType.key === 'energy'">-{{ formatNumber(energyConsumption) }}</template>
                    <template v-else>-</template>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </TabsContent>

          <!-- 产量详情标签页 -->
          <TabsContent value="production" class="mt-4">
            <div class="space-y-4">
              <div v-for="resourceType in resourceTypes" :key="resourceType.key" class="border-b last:border-b-0 pb-4 last:pb-0">
                <div class="flex items-center gap-2 mb-2">
                  <ResourceIcon :type="resourceType.key" size="sm" />
                  <span class="font-semibold">{{ t(`resources.${resourceType.key}`) }}</span>
                </div>

                <div v-if="productionBreakdown" class="ml-6 space-y-1 text-sm">
                  <!-- 电力有多个来源 -->
                  <template v-if="resourceType.key === 'energy' && productionBreakdown.energy.sources">
                    <div v-for="(source, idx) in productionBreakdown.energy.sources" :key="idx" class="flex justify-between">
                      <span class="text-muted-foreground">
                        {{ t(source.name) }}
                        <template v-if="source.name.startsWith('buildings.')">({{ t('common.level') }} {{ source.level }})</template>
                        <template v-else>({{ source.level }})</template>
                      </span>
                      <span class="text-green-600 dark:text-green-400">
                        +{{ formatNumber(Math.floor(source.production)) }}/{{ t('resources.hour') }}
                      </span>
                    </div>
                  </template>

                  <!-- 其他资源单一建筑产量 -->
                  <template v-else>
                    <div class="flex justify-between">
                      <span class="text-muted-foreground">
                        {{ t(productionBreakdown[resourceType.key].buildingName) }}
                        ({{ t('common.level') }} {{ productionBreakdown[resourceType.key].buildingLevel }})
                      </span>
                      <span class="text-green-600 dark:text-green-400">
                        +{{ formatNumber(Math.floor(productionBreakdown[resourceType.key].baseProduction)) }}/{{ t('resources.hour') }}
                      </span>
                    </div>
                  </template>

                  <!-- 加成列表 -->
                  <div v-for="(bonus, idx) in productionBreakdown[resourceType.key].bonuses" :key="idx" class="flex justify-between">
                    <span class="text-muted-foreground ml-4">
                      {{ t(bonus.name) }} ({{ bonus.percentage > 0 ? '+' : '' }}{{ bonus.percentage }}%)
                    </span>
                    <span :class="bonus.value > 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'">
                      {{ bonus.value > 0 ? '+' : '' }}{{ formatNumber(Math.floor(bonus.value)) }}/{{ t('resources.hour') }}
                    </span>
                  </div>

                  <!-- 最终产量 -->
                  <div class="flex justify-between font-semibold pt-1 border-t mt-1">
                    <span>{{ t('overview.totalProduction') }}</span>
                    <span class="text-green-600 dark:text-green-400">
                      +{{ formatNumber(Math.floor(productionBreakdown[resourceType.key].finalProduction)) }}/{{ t('resources.hour') }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </TabsContent>

          <!-- 消耗详情标签页 -->
          <TabsContent value="consumption" class="mt-4">
            <div class="space-y-2">
              <!-- 各建筑消耗 -->
              <div
                v-for="consumptionType in consumptionTypes"
                :key="consumptionType.key"
                v-show="consumptionBreakdown && consumptionBreakdown[consumptionType.key].buildingLevel > 0"
                class="flex justify-between text-sm"
              >
                <span v-if="consumptionBreakdown" class="text-muted-foreground">
                  {{ t(consumptionBreakdown[consumptionType.key].buildingName) }}
                  ({{ t('common.level') }} {{ consumptionBreakdown[consumptionType.key].buildingLevel }})
                </span>
                <span v-if="consumptionBreakdown" class="text-red-600 dark:text-red-400">
                  -{{ formatNumber(Math.floor(consumptionBreakdown[consumptionType.key].consumption)) }}/{{ t('resources.hour') }}
                </span>
              </div>

              <!-- 总消耗 -->
              <div v-if="consumptionBreakdown" class="flex justify-between font-semibold pt-2 border-t">
                <span>{{ t('overview.totalConsumption') }}</span>
                <span class="text-red-600 dark:text-red-400">
                  -{{ formatNumber(Math.floor(consumptionBreakdown.total)) }}/{{ t('resources.hour') }}
                </span>
              </div>

              <!-- 无消耗提示 -->
              <div v-if="consumptionBreakdown && consumptionBreakdown.total === 0" class="text-sm text-muted-foreground text-center py-2">
                {{ t('overview.noConsumption') }}
              </div>
            </div>
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>

    <!-- 舰队信息 -->
    <Card>
      <CardHeader>
        <CardTitle>{{ t('overview.fleetInfo') }}</CardTitle>
        <CardDescription>{{ t('overview.currentShips') }}</CardDescription>
      </CardHeader>
      <CardContent>
        <div class="grid grid-cols-3 sm:grid-cols-4 gap-3 sm:gap-4">
          <div v-for="(count, shipType) in planet.fleet" :key="shipType">
            <p class="text-xs sm:text-sm text-muted-foreground">{{ SHIPS[shipType].name }}</p>
            <p class="text-lg sm:text-xl font-bold">{{ count }}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
  import { useGameStore } from '@/stores/gameStore'
  import { useI18n } from '@/composables/useI18n'
  import { useGameConfig } from '@/composables/useGameConfig'
  import { computed, ref } from 'vue'
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
  import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
  import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
  import { Button } from '@/components/ui/button'
  import { Badge } from '@/components/ui/badge'
  import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle
  } from '@/components/ui/alert-dialog'
  import ResourceIcon from '@/components/common/ResourceIcon.vue'
  import { formatNumber, getResourceColor } from '@/utils/format'
  import { scaleNumber } from '@/utils/speed'
  import type { Planet } from '@/types/game'
  import { TechnologyType } from '@/types/game'
  import * as publicLogic from '@/logic/publicLogic'
  import * as resourceLogic from '@/logic/resourceLogic'
  import * as planetLogic from '@/logic/planetLogic'

  const gameStore = useGameStore()
  const { t } = useI18n()
  const { SHIPS } = useGameConfig()
  const planet = computed(() => gameStore.currentPlanet)

  // 获取科技加成
  const techBonuses = computed(() => ({
    mineralResearchLevel: gameStore.player.technologies[TechnologyType.MineralResearch] || 0,
    crystalResearchLevel: gameStore.player.technologies[TechnologyType.CrystalResearch] || 0,
    fuelResearchLevel: gameStore.player.technologies[TechnologyType.FuelResearch] || 0
  }))

  const production = computed(() =>
    planet.value ? publicLogic.getResourceProduction(planet.value, gameStore.player.officers, gameStore.gameSpeed, techBonuses.value) : null
  )
  const capacity = computed(() => (planet.value ? publicLogic.getResourceCapacity(planet.value, gameStore.player.officers) : null))

  // 能量消耗
  const energyConsumption = computed(() => {
    if (!planet.value) return 0
    return scaleNumber(resourceLogic.calculateEnergyConsumption(planet.value), gameStore.gameSpeed)
  })

  // 资源产量详细breakdown
  const productionBreakdown = computed(() => {
    if (!planet.value) return null
    return resourceLogic.calculateProductionBreakdown(planet.value, gameStore.player.officers, Date.now(), gameStore.gameSpeed, techBonuses.value)
  })

  // 资源消耗详细breakdown
  const consumptionBreakdown = computed(() => {
    if (!planet.value) return null
    return resourceLogic.calculateConsumptionBreakdown(planet.value, gameStore.gameSpeed)
  })

  // 资源类型配置
  const resourceTypes = [
    { key: 'metal' as const },
    { key: 'crystal' as const },
    { key: 'deuterium' as const },
    { key: 'energy' as const },
    { key: 'darkMatter' as const }
  ]

  // 消耗类型配置
  const consumptionTypes = [
    // 资源建筑
    { key: 'metalMine' as const },
    { key: 'crystalMine' as const },
    { key: 'deuteriumSynthesizer' as const },
    // 设施建筑
    { key: 'roboticsFactory' as const },
    { key: 'naniteFactory' as const },
    { key: 'shipyard' as const },
    { key: 'researchLab' as const },
    { key: 'missileSilo' as const },
    { key: 'terraformer' as const },
    { key: 'darkMatterCollector' as const },
    // 月球建筑
    { key: 'sensorPhalanx' as const },
    { key: 'jumpGate' as const }
  ]

  // 月球相关
  const moon = computed(() => {
    if (!planet.value || planet.value.isMoon) return null
    return getMoonForPlanet(planet.value.id)
  })

  const getMoonForPlanet = (planetId: string): Planet | null => {
    return gameStore.player.planets.find(p => p.isMoon && p.parentPlanetId === planetId) || null
  }

  // 切换到月球
  const switchToMoon = () => {
    if (moon.value) {
      gameStore.currentPlanetId = moon.value.id
    }
  }

  // 切换回母星
  const switchToParentPlanet = () => {
    if (planet.value?.parentPlanetId) {
      gameStore.currentPlanetId = planet.value.parentPlanetId
    }
  }

  // 放弃殖民地相关
  const showAbandonDialog = ref(false)
  const showErrorDialog = ref(false)
  const errorMessage = ref('')

  // 是否显示放弃按钮（非主星才显示）
  const canShowAbandonButton = computed(() => {
    if (!planet.value) return false
    // 找到主星（第一个非月球星球）
    const mainPlanet = gameStore.player.planets.find(p => !p.isMoon)
    // 当前星球不是主星时才显示放弃按钮
    return mainPlanet && mainPlanet.id !== planet.value.id
  })

  // 处理放弃殖民地
  const handleAbandonColony = () => {
    if (!planet.value) return

    const check = planetLogic.canAbandonColony(gameStore.player.planets, planet.value.id)
    if (!check.canAbandon) {
      showAbandonDialog.value = false
      errorMessage.value = check.reason ? t(check.reason) : t('planet.abandonFailed')
      showErrorDialog.value = true
      return
    }

    // 记录当前星球ID用于后续切换
    const abandonedPlanetId = planet.value.id

    // 执行放弃
    gameStore.player.planets = planetLogic.abandonColony(gameStore.player.planets, abandonedPlanetId)

    // 切换到主星
    const mainPlanet = gameStore.player.planets.find(p => !p.isMoon)
    if (mainPlanet) {
      gameStore.currentPlanetId = mainPlanet.id
    }

    showAbandonDialog.value = false
  }
</script>
