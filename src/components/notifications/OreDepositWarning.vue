<template>
  <div v-if="showWarning" class="bg-amber-500/10 border-b border-amber-500/20">
    <div class="px-4 sm:px-6 py-2 flex items-center justify-between gap-3">
      <!-- 警告图标和信息 -->
      <div class="flex items-center gap-2 flex-1 min-w-0">
        <Mountain class="h-5 w-5 text-amber-500 shrink-0 animate-pulse" />
        <div class="flex-1 min-w-0">
          <p class="text-sm font-semibold text-amber-500">
            {{ warningTitle }}
          </p>
          <p class="text-xs text-muted-foreground truncate">
            {{ detailMessage }}
          </p>
        </div>
      </div>

      <!-- 查看详情按钮 -->
      <Button @click="goToBuildings" variant="outline" size="sm" class="shrink-0">
        {{ t('common.viewDetails') }}
      </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import { useRouter } from 'vue-router'
  import { useGameStore } from '@/stores/gameStore'
  import { Button } from '@/components/ui/button'
  import { Mountain } from 'lucide-vue-next'
  import { useI18n } from '@/composables/useI18n'
  import * as oreDepositLogic from '@/logic/oreDepositLogic'

  const gameStore = useGameStore()
  const router = useRouter()
  const { t } = useI18n()

  // 获取当前星球
  const planet = computed(() => gameStore.currentPlanet)

  // 检查各资源的矿脉状态
  const depositStatus = computed(() => {
    if (!planet.value || planet.value.isMoon) return null

    const deposits = planet.value.oreDeposits
    if (!deposits) return null

    const resources = ['metal', 'crystal', 'deuterium'] as const
    const warnings: { type: string; depleted: boolean; percentage: number }[] = []

    for (const resource of resources) {
      const isDepleted = oreDepositLogic.isDepositDepleted(deposits, resource)
      const isWarning = oreDepositLogic.isDepositWarning(deposits, resource)
      const percentage = oreDepositLogic.getDepositPercentage(deposits, resource)

      if (isDepleted || isWarning) {
        warnings.push({
          type: resource,
          depleted: isDepleted,
          percentage: Math.round(percentage)
        })
      }
    }

    return warnings.length > 0 ? warnings : null
  })

  // 是否显示警告
  const showWarning = computed(() => {
    return depositStatus.value !== null && depositStatus.value.length > 0
  })

  // 获取资源名称翻译
  const getResourceName = (type: string): string => {
    const resourceNames: Record<string, string> = {
      metal: t('resources.metal'),
      crystal: t('resources.crystal'),
      deuterium: t('resources.deuterium')
    }
    return resourceNames[type] || type
  }

  // 警告标题
  const warningTitle = computed(() => {
    if (!depositStatus.value) return ''

    const hasDepleted = depositStatus.value.some(s => s.depleted)
    if (hasDepleted) {
      return t('oreDeposit.depletedWarning')
    }
    return t('oreDeposit.lowWarning')
  })

  // 详细消息
  const detailMessage = computed(() => {
    if (!depositStatus.value) return ''

    const depletedResources = depositStatus.value.filter(s => s.depleted).map(s => getResourceName(s.type))

    const warningResources = depositStatus.value.filter(s => !s.depleted).map(s => `${getResourceName(s.type)} (${s.percentage}%)`)

    const parts: string[] = []

    if (depletedResources.length > 0) {
      parts.push(t('oreDeposit.depletedResources', { resources: depletedResources.join(', ') }))
    }

    if (warningResources.length > 0) {
      parts.push(t('oreDeposit.lowResources', { resources: warningResources.join(', ') }))
    }

    return parts.join(' | ')
  })

  // 跳转到建筑页面查看详情
  const goToBuildings = () => {
    router.push('/galaxy')
  }
</script>
