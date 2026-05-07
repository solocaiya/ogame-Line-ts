<template>
  <Dialog v-model:open="isOpen">
    <DialogContent class="max-w-2xl">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2">
          <component :is="getMissionIcon(report?.missionType)" class="h-5 w-5" />
          {{ t('messagesView.missionReportDetails') }}
        </DialogTitle>
        <DialogDescription>
          {{ t('messagesView.missionDetails') }}
        </DialogDescription>
      </DialogHeader>

      <div v-if="report" class="space-y-4">
        <!-- 任务状态 -->
        <div class="p-4 bg-muted/50 rounded-lg border">
          <div class="flex items-center gap-2 mb-2">
            <h3 class="font-semibold text-lg">{{ getMissionTypeName(report.missionType) }}</h3>
            <Badge :variant="report.success ? 'default' : 'destructive'">
              {{ report.success ? t('messagesView.missionSuccess') : t('messagesView.missionFailed') }}
            </Badge>
          </div>
          <p class="text-sm text-muted-foreground mb-2">
            {{ formatDate(report.timestamp) }}
          </p>
          <p class="text-sm">{{ report.message }}</p>
        </div>

        <!-- 起点和终点 -->
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-2">
            <h4 class="font-semibold text-sm">{{ t('messagesView.origin') }}</h4>
            <div class="p-3 bg-muted/30 rounded-md">
              <p class="font-medium">{{ report.originPlanetName }}</p>
            </div>
          </div>
          <div class="space-y-2">
            <h4 class="font-semibold text-sm">{{ t('messagesView.destination') }}</h4>
            <div class="p-3 bg-muted/30 rounded-md">
              <p class="font-medium" v-if="report.targetPlanetName">{{ report.targetPlanetName }}</p>
              <p class="text-sm text-muted-foreground" v-else>
                [{{ report.targetPosition.galaxy }}:{{ report.targetPosition.system }}:{{ report.targetPosition.position }}]
              </p>
            </div>
          </div>
        </div>

        <!-- 任务详情 -->
        <div class="space-y-4">
          <!-- 运输任务详情 -->
          <div v-if="report.details?.transportedResources" class="space-y-3">
            <div class="flex items-center gap-2">
              <Package class="h-4 w-4" />
              <h4 class="font-semibold text-sm">{{ t('messagesView.transportedResources') }}</h4>
            </div>
            <div class="grid grid-cols-3 gap-3">
              <div v-for="res in basicResourceFields" :key="res.key" class="p-3 bg-muted/50 rounded-lg border">
                <div class="flex items-center gap-2 mb-1">
                  <ResourceIcon :type="res.key" size="sm" />
                  <span class="text-xs text-muted-foreground">{{ t(`resources.${res.key}`) }}</span>
                </div>
                <p class="text-lg font-bold">
                  {{ report.details.transportedResources[res.key].toLocaleString() }}
                </p>
              </div>
            </div>
          </div>

          <!-- 回收任务详情 -->
          <div v-if="report.details?.recycledResources" class="space-y-3">
            <div class="flex items-center gap-2">
              <Recycle class="h-4 w-4" />
              <h4 class="font-semibold text-sm">{{ t('messagesView.recycledResources') }}</h4>
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div v-for="res in debrisResourceFields" :key="res.key" class="p-3 bg-muted/50 rounded-lg border">
                <div class="flex items-center gap-2 mb-1">
                  <ResourceIcon :type="res.key" size="sm" />
                  <span class="text-xs text-muted-foreground">{{ t(`resources.${res.key}`) }}</span>
                </div>
                <p class="text-lg font-bold">+{{ report.details.recycledResources[res.key].toLocaleString() }}</p>
              </div>
            </div>
            <!-- 剩余残骸 -->
            <div v-if="report.details.remainingDebris" class="mt-3">
              <div class="flex items-center gap-2 mb-2">
                <AlertTriangle class="h-4 w-4" />
                <span class="text-sm font-medium text-muted-foreground">{{ t('messagesView.remainingDebris') }}</span>
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div v-for="res in debrisResourceFields" :key="res.key" class="p-3 bg-muted/50 rounded-lg border">
                  <div class="flex items-center gap-2 mb-1">
                    <ResourceIcon :type="res.key" size="sm" />
                    <span class="text-xs text-muted-foreground">{{ t(`resources.${res.key}`) }}</span>
                  </div>
                  <p class="text-lg font-bold">{{ report.details.remainingDebris[res.key].toLocaleString() }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- 殖民任务详情 -->
          <div v-if="report.details?.newPlanetName">
            <p class="text-xs font-semibold text-muted-foreground">{{ t('messagesView.newPlanet') }}:</p>
            <div class="flex items-center gap-2 mt-1">
              <Globe class="h-4 w-4" />
              <span class="font-medium">{{ report.details.newPlanetName }}</span>
            </div>
          </div>

          <!-- 导弹攻击详情 -->
          <div v-if="report.details?.missileCount !== undefined" class="space-y-4">
            <!-- 导弹统计卡片 -->
            <div class="grid grid-cols-3 gap-3">
              <!-- 发射数量 -->
              <div class="p-3 bg-muted/50 rounded-lg border">
                <div class="flex items-center gap-2 mb-1">
                  <Rocket class="h-4 w-4" />
                  <span class="text-xs text-muted-foreground">{{ t('galaxyView.missileCount') }}</span>
                </div>
                <p class="text-xl font-bold">{{ report.details.missileCount }}</p>
              </div>
              <!-- 命中数量 -->
              <div class="p-3 bg-muted/50 rounded-lg border">
                <div class="flex items-center gap-2 mb-1">
                  <Target class="h-4 w-4" />
                  <span class="text-xs text-muted-foreground">{{ t('missionReports.hits') }}</span>
                </div>
                <p class="text-xl font-bold">{{ report.details.missileHits }}</p>
              </div>
              <!-- 被拦截数量 -->
              <div class="p-3 bg-muted/50 rounded-lg border">
                <div class="flex items-center gap-2 mb-1">
                  <ShieldAlert class="h-4 w-4" />
                  <span class="text-xs text-muted-foreground">{{ t('galaxyView.intercepted') }}</span>
                </div>
                <p class="text-xl font-bold">{{ report.details.missileIntercepted }}</p>
              </div>
            </div>

            <!-- 防御损失 -->
            <div v-if="Object.keys(report.details.defenseLosses || {}).length > 0">
              <div class="flex items-center gap-2 mb-2">
                <Flame class="h-4 w-4" />
                <h4 class="font-semibold text-sm">{{ t('galaxyView.defenseLosses') }}</h4>
              </div>
              <div class="p-3 bg-muted/50 rounded-lg border">
                <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
                  <div
                    v-for="(count, defenseType) in report.details.defenseLosses"
                    :key="defenseType"
                    class="flex items-center justify-between p-2 bg-white/50 dark:bg-black/20 rounded"
                  >
                    <span class="text-sm text-muted-foreground">{{ t('defenses.' + defenseType) }}</span>
                    <span class="font-bold text-destructive">-{{ count }}</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- 无损失提示 -->
            <div
              v-else-if="report.details.missileHits === 0"
              class="p-3 bg-muted/50 rounded-lg border flex items-center gap-2"
            >
              <ShieldCheck class="h-5 w-5" />
              <span class="text-sm">{{ t('messagesView.noLosses') }}</span>
            </div>
          </div>

          <!-- 探险任务详情 - 探险区域 -->
          <div v-if="report.missionType === MissionType.Expedition && report.details?.expeditionZone" class="space-y-1">
            <p class="text-xs font-semibold text-muted-foreground">{{ t('fleetView.expeditionZone') }}:</p>
            <div class="p-2 bg-muted/50 rounded flex items-center gap-2">
              <MapPin class="h-4 w-4 text-primary" />
              <span class="font-medium">{{ t(`fleetView.zones.${report.details.expeditionZone}.name`) }}</span>
            </div>
          </div>

          <!-- 探险任务详情 - 发现资源 -->
          <div v-if="report.details?.foundResources" class="space-y-1">
            <p class="text-xs font-semibold text-muted-foreground">{{ t('messagesView.resources') }}:</p>
            <div class="grid grid-cols-2 gap-2 text-sm p-2 bg-muted/50 rounded">
              <div v-for="res in allResourceFields" :key="res.key">
                <template v-if="(report.details?.foundResources?.[res.key] ?? 0) > 0">
                  <span class="text-muted-foreground">{{ t(`resources.${res.key}`) }}:</span>
                  <span class="ml-1 font-medium">
                    +{{ (report.details?.foundResources?.[res.key] ?? 0).toLocaleString() }}
                  </span>
                </template>
              </div>
            </div>
          </div>

          <!-- 探险任务详情 - 发现舰船 -->
          <div v-if="report.details?.foundFleet" class="space-y-1">
            <p class="text-xs font-semibold text-muted-foreground">{{ t('messagesView.fleet') }}:</p>
            <div class="grid grid-cols-2 gap-2 text-sm p-2 bg-muted/50 rounded">
              <div v-for="(count, shipType) in report.details.foundFleet" :key="shipType">
                <span class="text-muted-foreground">{{ t('ships.' + shipType) }}:</span>
                <span class="ml-1 font-medium">+{{ count }}</span>
              </div>
            </div>
          </div>

          <!-- 探险任务详情 - 损失舰船 -->
          <div v-if="report.details?.fleetLost" class="space-y-1">
            <p class="text-xs font-semibold text-muted-foreground">{{ t('messagesView.attackerLosses') }}:</p>
            <div class="grid grid-cols-2 gap-2 text-sm p-2 bg-muted/50 rounded">
              <div v-for="(count, shipType) in report.details.fleetLost" :key="shipType">
                <span class="text-muted-foreground">{{ t('ships.' + shipType) }}:</span>
                <span class="ml-1 font-medium text-destructive">-{{ count }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <DialogFooter>
        <Button variant="outline" @click="isOpen = false">{{ t('common.close') }}</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
  import { ref, watch } from 'vue'
  import { useI18n } from '@/composables/useI18n'
  import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
  import { Button } from '@/components/ui/button'
  import { Badge } from '@/components/ui/badge'
  import ResourceIcon from '@/components/common/ResourceIcon.vue'
  import { formatDate } from '@/utils/format'
  import {
    Package,
    Recycle,
    AlertTriangle,
    Globe,
    Rocket,
    Target,
    ShieldAlert,
    Flame,
    ShieldCheck,
    Truck,
    Eye,
    Sword,
    Compass,
    Skull,
    MapPin
  } from 'lucide-vue-next'
  import { MissionType } from '@/types/game'
  import type { MissionReport } from '@/types/game'

  const props = defineProps<{
    report: MissionReport | null
    open: boolean
  }>()

  const emit = defineEmits<{
    (e: 'update:open', value: boolean): void
  }>()

  const { t } = useI18n()

  const isOpen = ref(props.open)

  watch(
    () => props.open,
    newValue => {
      isOpen.value = newValue
    }
  )

  watch(isOpen, newValue => {
    emit('update:open', newValue)
  })

  // 资源字段配置
  type BasicResourceKey = 'metal' | 'crystal' | 'deuterium'
  const basicResourceFields: { key: BasicResourceKey }[] = [{ key: 'metal' }, { key: 'crystal' }, { key: 'deuterium' }]

  type DebrisResourceKey = 'metal' | 'crystal'
  const debrisResourceFields: { key: DebrisResourceKey }[] = [{ key: 'metal' }, { key: 'crystal' }]

  type AllResourceKey = 'metal' | 'crystal' | 'deuterium' | 'darkMatter'
  const allResourceFields: { key: AllResourceKey }[] = [
    { key: 'metal' },
    { key: 'crystal' },
    { key: 'deuterium' },
    { key: 'darkMatter' }
  ]

  // 获取任务类型名称
  const getMissionTypeName = (missionType?: MissionType): string => {
    if (missionType === undefined) return ''
    switch (missionType) {
      case MissionType.Transport:
        return t('fleetView.transport')
      case MissionType.Deploy:
        return t('fleetView.deploy')
      case MissionType.Attack:
        return t('fleetView.attackMission')
      case MissionType.Spy:
        return t('fleetView.spy')
      case MissionType.Colonize:
        return t('fleetView.colonize')
      case MissionType.Recycle:
        return t('fleetView.recycle')
      case MissionType.Expedition:
        return t('fleetView.expedition')
      case MissionType.Destroy:
        return t('fleetView.destroy')
      case MissionType.MissileAttack:
        return t('galaxyView.missileAttack')
      default:
        return t('common.unknown')
    }
  }

  // 获取任务图标
  const getMissionIcon = (missionType?: MissionType) => {
    if (missionType === undefined) return Package
    switch (missionType) {
      case MissionType.Transport:
        return Truck
      case MissionType.Deploy:
        return Package
      case MissionType.Attack:
        return Sword
      case MissionType.Spy:
        return Eye
      case MissionType.Colonize:
        return Globe
      case MissionType.Recycle:
        return Recycle
      case MissionType.MissileAttack:
        return Rocket
      case MissionType.Expedition:
        return Compass
      case MissionType.Destroy:
        return Skull
      default:
        return Package
    }
  }
</script>
