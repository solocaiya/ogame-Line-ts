<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useGameStore } from '@/stores/gameStore'
import { useI18n } from '@/composables/useI18n'
import { apiService } from '@/services/apiService'
import { formatNumber } from '@/utils/format'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import {
  TrendingUp, Gift, CheckCircle, Lock, Loader2, Sparkles, Crown, Gem
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'

interface GrowthFundStageStatus {
  id: string
  pointsReq: number
  rewardDM: number
  description: string
  claimed: boolean
  claimable: boolean
}

interface GrowthFundStatus {
  purchased: boolean
  totalClaimed: number
  stages: GrowthFundStageStatus[]
  playerPoints: number
}

const gameStore = useGameStore()
const { t } = useI18n()

// --- State ---
const loading = ref(true)
const fundStatus = ref<GrowthFundStatus | null>(null)
const processing = ref(false)
const claimingStageId = ref<string | null>(null)

// --- Data Loading ---
async function loadData() {
  loading.value = true
  try {
    fundStatus.value = await apiService.getGrowthFund()
  } catch {
    toast.error(t('growthFund.loadFailed'))
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

// --- Computed ---
const totalRewardDM = computed(() => {
  if (!fundStatus.value) return 0
  return fundStatus.value.stages.reduce((sum, s) => sum + s.rewardDM, 0)
})

const claimedDM = computed(() => {
  if (!fundStatus.value) return 0
  return fundStatus.value.totalClaimed
})

const remainingDM = computed(() => {
  return totalRewardDM.value - claimedDM.value
})

const nextStage = computed<GrowthFundStageStatus | null>(() => {
  if (!fundStatus.value) return null
  return fundStatus.value.stages.find(s => !s.claimed && !s.claimable) ?? null
})

const prevStagePoints = computed(() => {
  if (!fundStatus.value || !nextStage.value) return 0
  const idx = fundStatus.value.stages.findIndex(s => s.id === nextStage.value!.id)
  if (idx <= 0) return 0
  return fundStatus.value.stages[idx - 1].pointsReq
})

const nextStageProgress = computed(() => {
  if (!fundStatus.value || !nextStage.value) return 100
  const current = fundStatus.value.playerPoints
  const from = prevStagePoints.value
  const to = nextStage.value.pointsReq
  if (to <= from) return 100
  const pct = ((current - from) / (to - from)) * 100
  return Math.max(0, Math.min(100, pct))
})

const claimableCount = computed(() => {
  if (!fundStatus.value) return 0
  return fundStatus.value.stages.filter(s => s.claimable).length
})

// --- Handlers ---
async function handleBuy() {
  processing.value = true
  try {
    await apiService.buyGrowthFund()
    toast.success(t('growthFund.buySuccess'))
    await loadData()
    gameStore.refreshWalletBalance?.()
  } catch {
    toast.error(t('growthFund.buyFailed'))
  } finally {
    processing.value = false
  }
}

async function handleClaim(stage: GrowthFundStageStatus) {
  claimingStageId.value = stage.id
  try {
    await apiService.claimGrowthFund(stage.id)
    toast.success(t('growthFund.claimSuccess', { amount: formatNumber(stage.rewardDM) }))
    await loadData()
    gameStore.refreshWalletBalance?.()
  } catch {
    toast.error(t('growthFund.claimFailed'))
  } finally {
    claimingStageId.value = null
  }
}

function getStageIcon(stage: GrowthFundStageStatus) {
  if (stage.claimed) return CheckCircle
  if (stage.claimable) return Gift
  return Lock
}

function getStageIconClass(stage: GrowthFundStageStatus) {
  if (stage.claimed) return 'text-green-500'
  if (stage.claimable) return 'text-amber-500 animate-pulse'
  return 'text-muted-foreground'
}

function getStageVariant(stage: GrowthFundStageStatus) {
  if (stage.claimed) return 'default' as const
  if (stage.claimable) return 'default' as const
  return 'secondary' as const
}
</script>

<template>
  <div class="p-4 space-y-4">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <h2 class="text-2xl font-bold flex items-center gap-2">
        <TrendingUp class="h-6 w-6 text-amber-500" />
        {{ t('growthFund.title') }}
      </h2>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-12">
      <Loader2 class="h-8 w-8 animate-spin text-muted-foreground" />
    </div>

    <template v-else-if="fundStatus">
      <!-- Not purchased: Hero card -->
      <Card v-if="!fundStatus.purchased" class="border-amber-500/30 bg-gradient-to-br from-amber-500/5 to-purple-500/5">
        <CardHeader>
          <CardTitle class="flex items-center gap-2 text-xl">
            <Sparkles class="h-5 w-5 text-amber-500" />
            {{ t('growthFund.heroTitle') }}
          </CardTitle>
        </CardHeader>
        <CardContent class="space-y-4">
          <p class="text-muted-foreground">
            {{ t('growthFund.heroDesc') }}
          </p>
          <div class="grid grid-cols-2 gap-4 text-center">
            <div class="p-3 rounded-lg bg-muted/50">
              <div class="text-2xl font-bold text-amber-500">{{ formatNumber(totalRewardDM) }}</div>
              <div class="text-xs text-muted-foreground">{{ t('growthFund.totalReturn') }}</div>
            </div>
            <div class="p-3 rounded-lg bg-muted/50">
              <div class="text-2xl font-bold">980</div>
              <div class="text-xs text-muted-foreground">{{ t('growthFund.cost') }} (DM)</div>
            </div>
          </div>
          <div class="p-3 rounded-lg border border-dashed border-amber-500/30 text-sm text-muted-foreground">
            {{ t('growthFund.howItWorks') }}
          </div>
          <!-- Stages preview (visible before purchase) -->
          <div class="space-y-2">
            <div class="text-sm font-semibold text-amber-400">{{ t('growthFund.stagesPreview') }}</div>
            <div
              v-for="stage in fundStatus.stages"
              :key="stage.id"
              class="flex items-center justify-between p-2.5 rounded-lg bg-muted/30 border border-border/50"
            >
              <div class="flex items-center gap-2">
                <div
                  class="w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold"
                  :class="stage.claimed ? 'bg-green-500/20 text-green-400' : 'bg-muted text-muted-foreground'"
                >
                  {{ stage.stageNumber }}
                </div>
                <div>
                  <div class="text-sm font-medium">{{ stage.name }}</div>
                  <div class="text-xs text-muted-foreground">{{ stage.conditionText }}</div>
                </div>
              </div>
              <div class="flex items-center gap-1 text-sm font-bold text-amber-500">
                <Gem class="w-3.5 h-3.5" />
                +{{ formatNumber(stage.rewardDM) }}
              </div>
            </div>
          </div>
          <Button
            class="w-full text-base py-6"
            size="lg"
            :disabled="processing"
            @click="handleBuy"
          >
            <Loader2 v-if="processing" class="mr-2 h-4 w-4 animate-spin" />
            <Gift v-else class="mr-2 h-5 w-5" />
            {{ t('growthFund.buyButton') }} — 980 DM
          </Button>
        </CardContent>
      </Card>

      <!-- Purchased: Status overview -->
      <template v-else>
        <!-- Summary cards -->
        <div class="grid grid-cols-3 gap-3">
          <Card>
            <CardContent class="pt-4 text-center">
              <div class="text-lg font-bold text-green-500">{{ formatNumber(claimedDM) }}</div>
              <div class="text-xs text-muted-foreground">{{ t('growthFund.claimed') }}</div>
            </CardContent>
          </Card>
          <Card>
            <CardContent class="pt-4 text-center">
              <div class="text-lg font-bold text-amber-500">{{ formatNumber(remainingDM) }}</div>
              <div class="text-xs text-muted-foreground">{{ t('growthFund.remaining') }}</div>
            </CardContent>
          </Card>
          <Card>
            <CardContent class="pt-4 text-center">
              <div class="text-lg font-bold">{{ formatNumber(fundStatus.playerPoints) }}</div>
              <div class="text-xs text-muted-foreground">{{ t('growthFund.myPoints') }}</div>
            </CardContent>
          </Card>
        </div>

        <!-- Progress to next stage -->
        <Card v-if="nextStage">
          <CardHeader class="pb-2">
            <CardTitle class="text-sm flex items-center justify-between">
              <span>{{ t('growthFund.nextStage') }}</span>
              <span class="text-muted-foreground">
                {{ formatNumber(fundStatus.playerPoints) }} / {{ formatNumber(nextStage.pointsReq) }}
              </span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <Progress :model-value="nextStageProgress" class="h-3" />
            <p class="text-xs text-muted-foreground mt-2">
              {{ t('growthFund.nextStageReward', { amount: formatNumber(nextStage.rewardDM) }) }}
            </p>
          </CardContent>
        </Card>

        <!-- All stages complete -->
        <Card v-else class="border-green-500/30 bg-green-500/5">
          <CardContent class="pt-4 text-center space-y-2">
            <Crown class="h-10 w-10 text-amber-500 mx-auto" />
            <p class="font-bold text-lg">{{ t('growthFund.allClaimed') }}</p>
            <p class="text-sm text-muted-foreground">
              {{ t('growthFund.totalClaimedAmount', { amount: formatNumber(claimedDM) }) }}
            </p>
          </CardContent>
        </Card>

        <!-- Stages list -->
        <div class="space-y-2">
          <h3 class="text-sm font-semibold text-muted-foreground px-1">
            {{ t('growthFund.stages') }}
          </h3>
          <Card
            v-for="stage in fundStatus.stages"
            :key="stage.id"
            :class="[
              stage.claimable ? 'border-amber-500/40 bg-amber-500/5' : '',
              stage.claimed ? 'opacity-70' : ''
            ]"
          >
            <CardContent class="py-3 flex items-center gap-3">
              <!-- Icon -->
              <component
                :is="getStageIcon(stage)"
                class="h-5 w-5 shrink-0"
                :class="getStageIconClass(stage)"
              />

              <!-- Info -->
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium truncate">{{ stage.description }}</span>
                  <Badge :variant="getStageVariant(stage)" class="shrink-0 text-[10px] h-4 px-1.5">
                    {{ stage.claimed ? t('growthFund.statusClaimed') : stage.claimable ? t('growthFund.statusClaimable') : t('growthFund.statusLocked') }}
                  </Badge>
                </div>
                <div class="text-xs text-muted-foreground mt-0.5">
                  {{ t('growthFund.pointsReq', { points: formatNumber(stage.pointsReq) }) }}
                  ·
                  {{ t('growthFund.reward', { amount: formatNumber(stage.rewardDM) }) }}
                </div>
              </div>

              <!-- Action -->
              <Button
                v-if="stage.claimable"
                size="sm"
                :disabled="claimingStageId === stage.id"
                @click="handleClaim(stage)"
                class="shrink-0"
              >
                <Loader2 v-if="claimingStageId === stage.id" class="h-3 w-3 animate-spin" />
                <Gift v-else class="h-3 w-3" />
                <span class="ml-1">{{ t('growthFund.claim') }}</span>
              </Button>
            </CardContent>
          </Card>
        </div>
      </template>
    </template>
  </div>
</template>
