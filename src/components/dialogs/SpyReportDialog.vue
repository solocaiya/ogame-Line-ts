<template>
  <Dialog v-model:open="isOpen">
    <ScrollableDialogContent container-class="sm:max-w-2xl max-h-[90vh]">
      <template #header>
        <DialogHeader>
          <DialogTitle class="flex items-center gap-2">
            <Eye class="h-5 w-5" />
            {{ t('messagesView.spyReport') }}
          </DialogTitle>
          <DialogDescription v-if="report">
            {{ formatDate(report.timestamp) }}
          </DialogDescription>
        </DialogHeader>
      </template>

      <div v-if="report" class="space-y-4">
        <!-- 目标星球信息 -->
        <div class="p-4 bg-muted/50 rounded-lg border">
          <div class="flex items-center gap-3">
            <div class="p-2 bg-background rounded-full border">
              <Globe class="h-5 w-5" />
            </div>
            <div>
              <p class="font-semibold">{{ report.targetPlanetName }}</p>
              <p class="text-sm text-muted-foreground">
                [{{ report.targetPosition.galaxy }}:{{ report.targetPosition.system }}:{{ report.targetPosition.position }}]
              </p>
            </div>
          </div>
        </div>

        <!-- 资源 -->
        <div class="space-y-3">
          <div class="flex items-center gap-2">
            <Coins class="h-4 w-4" />
            <h4 class="font-semibold text-sm">{{ t('messagesView.resources') }}</h4>
          </div>
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <!-- 金属 -->
            <div class="p-3 bg-muted/50 rounded-lg border">
              <div class="flex items-center gap-2 mb-1">
                <ResourceIcon type="metal" size="sm" />
                <span class="text-xs text-muted-foreground">{{ t('resources.metal') }}</span>
              </div>
              <p class="text-lg font-bold">
                {{ formatNumber(report.resources.metal) }}
              </p>
            </div>
            <!-- 晶体 -->
            <div class="p-3 bg-muted/50 rounded-lg border">
              <div class="flex items-center gap-2 mb-1">
                <ResourceIcon type="crystal" size="sm" />
                <span class="text-xs text-muted-foreground">{{ t('resources.crystal') }}</span>
              </div>
              <p class="text-lg font-bold">
                {{ formatNumber(report.resources.crystal) }}
              </p>
            </div>
            <!-- 氘 -->
            <div class="p-3 bg-muted/50 rounded-lg border">
              <div class="flex items-center gap-2 mb-1">
                <ResourceIcon type="deuterium" size="sm" />
                <span class="text-xs text-muted-foreground">{{ t('resources.deuterium') }}</span>
              </div>
              <p class="text-lg font-bold">
                {{ formatNumber(report.resources.deuterium) }}
              </p>
            </div>
            <!-- 暗物质 -->
            <div class="p-3 bg-muted/50 rounded-lg border">
              <div class="flex items-center gap-2 mb-1">
                <ResourceIcon type="darkMatter" size="sm" />
                <span class="text-xs text-muted-foreground">{{ t('resources.darkMatter') }}</span>
              </div>
              <p class="text-lg font-bold">
                {{ formatNumber(report.resources.darkMatter) }}
              </p>
            </div>
          </div>
        </div>

        <!-- 舰队（如果有） -->
        <div v-if="report.fleet && Object.keys(report.fleet).length > 0" class="space-y-3">
          <div class="flex items-center gap-2">
            <Rocket class="h-4 w-4" />
            <h4 class="font-semibold text-sm">{{ t('messagesView.fleet') }}</h4>
            <Badge variant="secondary" class="text-xs">{{ getTotalFleetCount(report.fleet) }}</Badge>
          </div>
          <div class="p-3 bg-muted/50 rounded-lg border">
            <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
              <div
                v-for="(count, shipType) in report.fleet"
                :key="shipType"
                class="flex items-center justify-between p-2 bg-white/50 dark:bg-black/20 rounded"
              >
                <span class="text-sm text-muted-foreground truncate">{{ SHIPS[shipType].name }}</span>
                <span class="font-bold">{{ count }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 防御设施（如果有） -->
        <div v-if="report.defense && hasDefense(report.defense)" class="space-y-3">
          <div class="flex items-center gap-2">
            <Shield class="h-4 w-4" />
            <h4 class="font-semibold text-sm">{{ t('messagesView.defense') }}</h4>
            <Badge variant="secondary" class="text-xs">{{ getTotalDefenseCount(report.defense) }}</Badge>
          </div>
          <div class="p-3 bg-muted/50 rounded-lg border">
            <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
              <div
                v-for="(count, defenseType) in report.defense"
                :key="defenseType"
                class="flex items-center justify-between p-2 bg-white/50 dark:bg-black/20 rounded"
              >
                <template v-if="count && count > 0">
                  <span class="text-sm text-muted-foreground truncate">{{ DEFENSES[defenseType].name }}</span>
                  <span class="font-bold">{{ count }}</span>
                </template>
              </div>
            </div>
          </div>
        </div>

        <!-- 建筑（如果有） -->
        <div v-if="report.buildings && Object.keys(report.buildings).length > 0" class="space-y-3">
          <div class="flex items-center gap-2">
            <Building class="h-4 w-4" />
            <h4 class="font-semibold text-sm">{{ t('messagesView.buildings') }}</h4>
          </div>
          <div class="p-3 bg-muted/50 rounded-lg border">
            <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
              <div
                v-for="(level, buildingType) in report.buildings"
                :key="buildingType"
                class="flex items-center justify-between p-2 bg-white/50 dark:bg-black/20 rounded"
              >
                <span class="text-sm text-muted-foreground truncate">{{ BUILDINGS[buildingType].name }}</span>
                <Badge variant="outline" class="font-bold">Lv.{{ level }}</Badge>
              </div>
            </div>
          </div>
        </div>
      </div>
    </ScrollableDialogContent>
  </Dialog>
</template>

<script setup lang="ts">
  import { ref, watch } from 'vue'
  import { useI18n } from '@/composables/useI18n'
  import { useGameConfig } from '@/composables/useGameConfig'
  import { Dialog, ScrollableDialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
  import { Badge } from '@/components/ui/badge'
  import ResourceIcon from '@/components/common/ResourceIcon.vue'
  import { formatNumber, formatDate } from '@/utils/format'
  import { Eye, Globe, Coins, Rocket, Shield, Building } from 'lucide-vue-next'
  import type { SpyReport } from '@/types/game'

  const props = defineProps<{
    report: SpyReport | null
    open: boolean
  }>()

  const emit = defineEmits<{
    (e: 'update:open', value: boolean): void
  }>()

  const { t } = useI18n()
  const { SHIPS, DEFENSES, BUILDINGS } = useGameConfig()

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

  // 检查是否有防御设施
  const hasDefense = (defense: any): boolean => {
    if (!defense) return false
    return Object.values(defense).some((count: any) => count > 0)
  }

  // 获取舰队总数
  const getTotalFleetCount = (fleet: Record<string, number>): number => {
    return Object.values(fleet).reduce((sum, count) => sum + count, 0)
  }

  // 获取防御总数
  const getTotalDefenseCount = (defense: Record<string, number>): number => {
    return Object.values(defense).reduce((sum, count) => sum + (count || 0), 0)
  }
</script>
