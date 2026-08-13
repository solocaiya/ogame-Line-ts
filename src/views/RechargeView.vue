<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useGameStore } from '@/stores/gameStore'
import { useI18n } from '@/composables/useI18n'
import { apiService } from '@/services/apiService'
import { formatNumber } from '@/utils/format'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import {
  Gem, CreditCard, Gift, Clock, Zap, Crown, CheckCircle,
  Loader2, Sparkles, TrendingUp, Shield, Swords, Rocket, Timer
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'

interface RechargeProduct {
  id: string; amountRMB: number; darkMatter: number; bonus: number; firstBonus: number
}
interface MonthlyCardInfo {
  id: string; name: string; amountRMB: number; dailyDM: number; durationDays: number
  perks: string[]; vipLevel: number; active: boolean; expiresAt: string
}
interface GiftPackInfo {
  id: string; name: string; costDM: number; onceOnly: boolean; cooldownHours: number
  contents: { darkMatter?: number; metal?: number; crystal?: number; deuterium?: number }
  available: boolean; purchased?: boolean; cooldownRemaining?: number
}

const gameStore = useGameStore()
const { t } = useI18n()

// --- State ---
const loading = ref(true)
const products = ref<RechargeProduct[]>([])
const monthlyCards = ref<MonthlyCardInfo[]>([])
const giftPacks = ref<GiftPackInfo[]>([])
const firstRecharge = ref(false)
const dailyReady = ref(false)
const processingId = ref<string | null>(null)
const claimingDaily = ref(false)

// --- Data Loading ---
async function loadData() {
  loading.value = true
  try {
    const [prodRes, cardsRes, packsRes] = await Promise.all([
      apiService.getRechargeProducts(),
      apiService.getMonthlyCards(),
      apiService.getGiftPacks(),
    ])
    products.value = prodRes.products
    firstRecharge.value = prodRes.firstRecharge
    monthlyCards.value = cardsRes.cards
    dailyReady.value = cardsRes.dailyReady
    giftPacks.value = packsRes.packs
  } catch (e) {
    toast.error(t('common.error'))
  } finally {
    loading.value = false
  }
}

// --- Purchase Handlers ---
async function handlePurchase(product: RechargeProduct) {
  if (processingId.value) return
  processingId.value = product.id
  try {
    const { orderId } = await apiService.createRechargeOrder(product.id)
    await apiService.confirmPayment(orderId)
    await gameStore.refreshWalletBalance()
    toast.success(t('recharge.purchaseSuccess', { amount: formatNumber(product.darkMatter + (firstRecharge.value ? product.firstBonus : 0)) }))
  } catch (e: any) {
    toast.error(e.message || t('common.error'))
  } finally {
    processingId.value = null
  }
}

async function handleBuyMonthlyCard(cardId: string) {
  if (processingId.value) return
  processingId.value = cardId
  try {
    await apiService.buyMonthlyCard(cardId)
    await Promise.all([
      gameStore.refreshWalletBalance(),
      loadData(),
    ])
    toast.success(t('recharge.cardPurchaseSuccess'))
  } catch (e: any) {
    toast.error(e.message || t('common.error'))
  } finally {
    processingId.value = null
  }
}

async function handleClaimDaily() {
  if (claimingDaily.value) return
  claimingDaily.value = true
  try {
    const res = await apiService.claimDailyDM()
    await gameStore.refreshWalletBalance()
    toast.success(t('recharge.dailyClaimSuccess', { amount: formatNumber(res.dmClaimed) }))
    dailyReady.value = false
  } catch (e: any) {
    toast.error(e.message || t('common.error'))
  } finally {
    claimingDaily.value = false
  }
}

async function handleBuyGiftPack(pack: GiftPackInfo) {
  if (processingId.value) return
  processingId.value = pack.id
  try {
    await apiService.buyGiftPack(pack.id)
    await Promise.all([
      gameStore.refreshWalletBalance(),
      loadData(),
    ])
    toast.success(t('recharge.giftPurchaseSuccess'))
  } catch (e: any) {
    toast.error(e.message || t('common.error'))
  } finally {
    processingId.value = null
  }
}

// --- Helpers ---
function getPerkIcon(perk: string) {
  if (perk.includes('build')) return Clock
  if (perk.includes('attack') || perk.includes('defense')) return Shield
  if (perk.includes('fleet')) return Rocket
  if (perk.includes('research')) return TrendingUp
  return Zap
}

function formatCooldown(hours: number): string {
  if (hours >= 168) return `${Math.floor(hours / 24)}天`
  if (hours >= 24) return `${Math.floor(hours / 24)}天${hours % 24}小时`
  return `${hours}小时`
}

function formatCountdown(hours: number): string {
  const h = Math.floor(hours)
  const m = Math.floor((hours - h) * 60)
  if (h >= 24) return `${Math.floor(h / 24)}天 ${h % 24}时`
  return `${h}时 ${m}分`
}

function getExpiryDays(expiresAt: string): number {
  if (!expiresAt) return 0
  const diff = new Date(expiresAt).getTime() - Date.now()
  return Math.max(0, Math.ceil(diff / 86400000))
}

function getTotalContents(pack: GiftPackInfo): string {
  const parts: string[] = []
  if (pack.contents.darkMatter) parts.push(`${formatNumber(pack.contents.darkMatter)} DM`)
  if (pack.contents.metal) parts.push(`${formatNumber(pack.contents.metal)} 金属`)
  if (pack.contents.crystal) parts.push(`${formatNumber(pack.contents.crystal)} 晶体`)
  if (pack.contents.deuterium) parts.push(`${formatNumber(pack.contents.deuterium)} 重氢`)
  return parts.join(' + ')
}

// --- Computed ---
const hasActiveCard = computed(() => monthlyCards.value.some(c => c.active))

onMounted(loadData)
</script>

<template>
  <div class="container mx-auto p-4 space-y-6 max-w-5xl">
    <!-- Header: Balance -->
    <div class="flex items-center justify-between">
      <h2 class="text-2xl font-bold flex items-center gap-2">
        <Gem class="w-7 h-7 text-purple-400" />
        {{ t('recharge.title') }}
      </h2>
      <Card class="px-4 py-2 bg-purple-950/30 border-purple-500/30">
        <div class="flex items-center gap-2">
          <Gem class="w-4 h-4 text-purple-400" />
          <span class="text-lg font-bold text-purple-300">{{ formatNumber(gameStore.darkMatterBalance) }}</span>
        </div>
      </Card>
    </div>

    <!-- Mock payment notice -->
    <div class="text-xs text-muted-foreground bg-yellow-500/10 border border-yellow-500/20 rounded px-3 py-1.5">
      ⚠️ {{ t('recharge.mockNotice') }}
    </div>

    <Tabs default-value="recharge" class="space-y-4">
      <TabsList class="grid w-full grid-cols-3">
        <TabsTrigger value="recharge">{{ t('recharge.tabRecharge') }}</TabsTrigger>
        <TabsTrigger value="monthly">{{ t('recharge.tabMonthly') }}</TabsTrigger>
        <TabsTrigger value="gifts">
          {{ t('recharge.tabGifts') }}
          <Badge v-if="giftPacks.some(p => p.available)" variant="default" class="ml-1 h-4 text-[10px] px-1">
            {{ giftPacks.filter(p => p.available).length }}
          </Badge>
        </TabsTrigger>
      </TabsList>

      <!-- ==================== Tab 1: Recharge Products ==================== -->
      <TabsContent value="recharge" class="space-y-4">
        <div v-if="loading" class="flex justify-center py-12">
          <Loader2 class="w-6 h-6 animate-spin text-muted-foreground" />
        </div>
        <div v-else class="grid grid-cols-2 md:grid-cols-3 gap-4">
          <Card
            v-for="product in products"
            :key="product.id"
            class="relative overflow-hidden cursor-pointer transition-all hover:ring-2 hover:ring-purple-400/50 hover:shadow-lg hover:shadow-purple-500/10"
            :class="{ 'opacity-60 pointer-events-none': processingId && processingId !== product.id }"
            @click="handlePurchase(product)"
          >
            <!-- First-recharge bonus banner -->
            <div
              v-if="firstRecharge && product.firstBonus > 0"
              class="absolute top-0 right-0 bg-gradient-to-l from-amber-500 to-orange-500 text-white text-[10px] font-bold px-2 py-0.5 rounded-bl"
            >
              <Sparkles class="w-3 h-3 inline mr-0.5" />
              {{ t('recharge.firstBonus') }} +{{ formatNumber(product.firstBonus) }}
            </div>

            <CardContent class="pt-6 pb-4 text-center space-y-2">
              <!-- DM amount -->
              <div class="flex items-center justify-center gap-1.5">
                <Gem class="w-6 h-6 text-purple-400" />
                <span class="text-2xl font-bold">{{ formatNumber(product.darkMatter) }}</span>
              </div>
              <!-- Bonus (non-first) -->
              <div v-if="product.bonus > 0 && !firstRecharge" class="text-xs text-green-400">
                +{{ formatNumber(product.bonus) }} {{ t('recharge.bonus') }}
              </div>
              <!-- Unit price -->
              <div class="text-[11px] text-muted-foreground">
                ≈ {{ (product.amountRMB / product.darkMatter).toFixed(2) }} ¥/DM
              </div>
              <!-- Price -->
              <div class="text-lg font-bold text-amber-400 pt-1">
                ¥{{ product.amountRMB }}
              </div>
            </CardContent>

            <div class="px-3 pb-3">
              <Button
                class="w-full"
                size="sm"
                :disabled="processingId !== null"
              >
                <Loader2 v-if="processingId === product.id" class="w-4 h-4 mr-1 animate-spin" />
                <CreditCard v-else class="w-4 h-4 mr-1" />
                {{ t('recharge.buy') }}
              </Button>
            </div>
          </Card>
        </div>
      </TabsContent>

      <!-- ==================== Tab 2: Monthly Cards ==================== -->
      <TabsContent value="monthly" class="space-y-4">
        <div v-if="loading" class="flex justify-center py-12">
          <Loader2 class="w-6 h-6 animate-spin text-muted-foreground" />
        </div>
        <template v-else>
          <!-- Daily claim banner -->
          <Card v-if="hasActiveCard" class="border-green-500/30 bg-green-950/20">
            <CardContent class="pt-4 pb-4 flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-full bg-green-500/20 flex items-center justify-center">
                  <Sparkles class="w-5 h-5 text-green-400" />
                </div>
                <div>
                  <p class="font-semibold text-green-300">{{ t('recharge.dailyClaimTitle') }}</p>
                  <p class="text-xs text-muted-foreground">{{ t('recharge.dailyClaimDesc') }}</p>
                </div>
              </div>
              <Button
                size="sm"
                :disabled="!dailyReady || claimingDaily"
                @click="handleClaimDaily"
                class="bg-green-600 hover:bg-green-700"
              >
                <Loader2 v-if="claimingDaily" class="w-4 h-4 mr-1 animate-spin" />
                <CheckCircle v-else class="w-4 h-4 mr-1" />
                {{ dailyReady ? t('recharge.claim') : t('recharge.claimed') }}
              </Button>
            </CardContent>
          </Card>

          <!-- Monthly card grid -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Card
              v-for="card in monthlyCards"
              :key="card.id"
              class="relative overflow-hidden"
              :class="card.active ? 'border-amber-500/40 bg-amber-950/10' : ''"
            >
              <!-- Active badge -->
              <div v-if="card.active" class="absolute top-2 right-2">
                <Badge variant="default" class="bg-amber-500 text-white">
                  <Crown class="w-3 h-3 mr-0.5" />
                  VIP {{ card.vipLevel }}
                </Badge>
              </div>

              <CardHeader class="pb-2">
                <CardTitle class="text-lg flex items-center gap-2">
                  <Crown class="w-5 h-5" :class="card.id === 'large_monthly' ? 'text-amber-400' : 'text-blue-400'" />
                  {{ card.name }}
                </CardTitle>
              </CardHeader>

              <CardContent class="flex flex-col space-y-3">
                <!-- Price + daily DM -->
                <div class="flex items-end justify-between">
                  <div>
                    <span class="text-2xl font-bold text-amber-400">¥{{ card.amountRMB }}</span>
                    <span class="text-xs text-muted-foreground ml-1">/ {{ card.durationDays }}天</span>
                  </div>
                  <div class="text-right">
                    <div class="flex items-center gap-1 text-purple-400 font-semibold">
                      <Gem class="w-3.5 h-3.5" />
                      {{ formatNumber(card.dailyDM) }} / {{ t('recharge.day') }}
                    </div>
                  </div>
                </div>

                <!-- Perks -->
                <div class="space-y-1.5 pt-1">
                  <div
                    v-for="perk in card.perks"
                    :key="perk"
                    class="flex items-center gap-2 text-xs text-muted-foreground"
                  >
                    <component :is="getPerkIcon(perk)" class="w-3.5 h-3.5 text-primary" />
                    <span>{{ t(`recharge.perks.${perk}`, perk) }}</span>
                  </div>
                </div>

                <!-- Expiry countdown (if active) -->
                <div v-if="card.active && card.expiresAt" class="text-xs text-muted-foreground flex items-center gap-1 pt-1">
                  <Timer class="w-3 h-3" />
                  {{ t('recharge.expiresIn', { days: getExpiryDays(card.expiresAt) }) }}
                </div>

                <!-- Buy button (pushed to bottom) -->
                <Button
                  v-if="!card.active"
                  class="w-full mt-auto"
                  size="sm"
                  :disabled="processingId !== null"
                  @click="handleBuyMonthlyCard(card.id)"
                >
                  <Loader2 v-if="processingId === card.id" class="w-4 h-4 mr-1 animate-spin" />
                  <CreditCard v-else class="w-4 h-4 mr-1" />
                  {{ t('recharge.buyCard') }}
                </Button>
                <div v-else class="text-center text-xs text-green-400 py-1.5 flex items-center justify-center gap-1 mt-auto">
                  <CheckCircle class="w-3.5 h-3.5" />
                  {{ t('recharge.active') }}
                </div>
              </CardContent>
            </Card>
          </div>
        </template>
      </TabsContent>

      <!-- ==================== Tab 3: Gift Packs ==================== -->
      <TabsContent value="gifts" class="space-y-4">
        <div v-if="loading" class="flex justify-center py-12">
          <Loader2 class="w-6 h-6 animate-spin text-muted-foreground" />
        </div>
        <div v-else-if="giftPacks.length === 0" class="text-center py-12 text-muted-foreground">
          <Gift class="w-10 h-10 mx-auto mb-2 opacity-40" />
          {{ t('recharge.noGifts') }}
        </div>
        <div v-else class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card
            v-for="pack in giftPacks"
            :key="pack.id"
            class="relative overflow-hidden"
            :class="{
              'border-green-500/30 bg-green-950/10': pack.available,
              'opacity-60': !pack.available,
            }"
          >
            <!-- Available badge -->
            <div v-if="pack.available" class="absolute top-2 right-2">
              <Badge variant="default" class="bg-green-500 text-white">
                <Sparkles class="w-3 h-3 mr-0.5" />
                {{ t('recharge.available') }}
              </Badge>
            </div>

            <CardHeader class="pb-2">
              <CardTitle class="text-base flex items-center gap-2">
                <Gift class="w-5 h-5 text-pink-400" />
                {{ pack.name }}
              </CardTitle>
            </CardHeader>

            <CardContent class="flex flex-col space-y-3">
              <!-- Contents -->
              <div class="text-sm text-muted-foreground">
                {{ getTotalContents(pack) }}
              </div>

              <!-- Cost -->
              <div class="flex items-center gap-1 text-purple-400 font-semibold">
                <Gem class="w-4 h-4" />
                {{ formatNumber(pack.costDM) }}
              </div>

              <!-- Status -->
              <div v-if="pack.onceOnly && pack.purchased" class="text-xs text-muted-foreground flex items-center gap-1">
                <CheckCircle class="w-3 h-3 text-green-400" />
                {{ t('recharge.purchased') }}
              </div>
              <div v-else-if="!pack.onceOnly && pack.cooldownRemaining && pack.cooldownRemaining > 0" class="text-xs text-muted-foreground flex items-center gap-1">
                <Clock class="w-3 h-3" />
                {{ t('recharge.cooldown') }}: {{ formatCountdown(pack.cooldownRemaining) }}
              </div>
              <div v-else-if="!pack.available" class="text-xs text-muted-foreground">
                {{ t('recharge.unavailable') }}
              </div>

              <!-- Buy button (pushed to bottom) -->
              <Button
                v-if="pack.available"
                class="w-full mt-auto"
                size="sm"
                :disabled="processingId !== null"
                @click="handleBuyGiftPack(pack)"
              >
                <Loader2 v-if="processingId === pack.id" class="w-4 h-4 mr-1 animate-spin" />
                <Gift v-else class="w-4 h-4 mr-1" />
                {{ t('recharge.buyGift') }}
              </Button>
              <div v-else class="mt-auto text-center text-xs text-muted-foreground py-2">
                {{ pack.purchased ? t('recharge.purchased') : t('recharge.unavailable') }}
              </div>
            </CardContent>
          </Card>
        </div>
      </TabsContent>
    </Tabs>
  </div>
</template>
