<template>
  <div v-if="!isUnlocked" class="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm flex items-center justify-center p-4">
    <Card class="max-w-md w-full">
      <CardHeader class="text-center">
        <div class="flex justify-center mb-4">
          <div class="rounded-full bg-muted p-4">
            <Lock :size="48" class="text-muted-foreground" />
          </div>
        </div>
        <CardTitle class="text-xl sm:text-2xl">{{ t('common.featureLocked') }}</CardTitle>
        <CardDescription class="text-sm sm:text-base">{{ t('common.unlockRequired') }}</CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="p-4 bg-muted rounded-lg space-y-2">
          <p class="text-sm font-medium text-center">{{ t('common.requiredBuilding') }}:</p>
          <div class="flex items-center justify-center gap-2">
            <span class="text-base sm:text-lg font-bold">{{ buildingName }}</span>
            <Badge variant="default">Lv {{ requiredLevel }}</Badge>
          </div>
          <p v-if="currentLevel !== undefined" class="text-xs text-center text-muted-foreground">
            {{ t('common.currentLevel') }}: Lv {{ currentLevel }}
          </p>
        </div>

        <div class="flex gap-2">
          <Button @click="goToBuildings" class="flex-1">
            <Building2 :size="16" class="mr-2" />
            {{ t('common.goToBuildings') }}
          </Button>
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import { useRouter } from 'vue-router'
  import { useGameStore } from '@/stores/gameStore'
  import { useI18n } from '@/composables/useI18n'
  import { useGameConfig } from '@/composables/useGameConfig'
  import { BuildingType } from '@/types/game'
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
  import { Button } from '@/components/ui/button'
  import { Badge } from '@/components/ui/badge'
  import { Lock, Building2 } from 'lucide-vue-next'

  interface Props {
    requiredBuilding: BuildingType
    requiredLevel: number
  }

  const props = defineProps<Props>()
  const router = useRouter()
  const gameStore = useGameStore()
  const { t } = useI18n()
  const { BUILDINGS } = useGameConfig()

  const buildingName = computed(() => BUILDINGS.value[props.requiredBuilding]?.name || props.requiredBuilding)

  const currentLevel = computed(() => {
    if (!gameStore.currentPlanet) return 0
    return gameStore.currentPlanet.buildings[props.requiredBuilding] || 0
  })

  const isUnlocked = computed(() => {
    return currentLevel.value >= props.requiredLevel
  })

  const goToBuildings = () => {
    router.push('/buildings')
  }
</script>
