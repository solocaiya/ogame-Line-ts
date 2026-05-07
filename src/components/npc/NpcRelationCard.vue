<template>
  <div class="rounded-lg transition-shadow duration-300">
    <Card>
      <CardHeader>
        <div class="flex items-start justify-between">
          <div class="flex-1">
            <CardTitle class="flex items-center gap-2 flex-wrap">
              {{ npc.name }}
              <span v-if="npc.note" class="text-muted-foreground font-normal">({{ npc.note }})</span>
              <Badge :variant="statusBadgeVariant">
                {{ statusText }}
              </Badge>
              <!-- NPC难度等级徽章 -->
              <Badge
                v-if="npc.difficultyLevel"
                :variant="difficultyBadgeVariant"
                :class="difficultyLevelColor"
              >
                Lv.{{ npc.difficultyLevel }}
              </Badge>
            </CardTitle>
            <CardDescription class="mt-1">
              {{ npc.planets.length }} {{ t('diplomacy.planets') }}
              <span v-if="npc.allies && npc.allies.length > 0" class="ml-2">· {{ npc.allies.length }} {{ t('diplomacy.allies') }}</span>
            </CardDescription>
          </div>
          <!-- 编辑备注按钮 -->
          <Button
            variant="ghost"
            size="sm"
            class="h-8 w-8 p-0"
            @click="openNoteDialog"
            :title="npc.note ? t('diplomacy.actions.editNote') : t('diplomacy.actions.addNote')"
          >
            <Pencil class="h-4 w-4" />
          </Button>
        </div>
      </CardHeader>
      <CardContent class="space-y-4">
        <!-- 好感度进度条 -->
        <div class="space-y-2">
          <div class="flex items-center justify-between text-sm">
            <span class="text-muted-foreground">{{ t('diplomacy.reputation') }}</span>
            <span class="font-semibold" :class="reputationColor">{{ reputation > 0 ? '+' : '' }}{{ reputation }}</span>
          </div>
          <div class="relative">
            <!-- 背景进度条 -->
            <div class="h-2 bg-muted rounded-full overflow-hidden">
              <!-- 负值部分（左侧，红色） -->
              <div
                v-if="reputation < 0"
                class="h-full bg-red-500 dark:bg-red-600 absolute right-1/2"
                :style="{ width: `${Math.abs(reputation) / 2}%` }"
              />
              <!-- 正值部分（右侧，绿色） -->
              <div
                v-if="reputation > 0"
                class="h-full bg-green-500 dark:bg-green-600 absolute left-1/2"
                :style="{ width: `${reputation / 2}%` }"
              />
            </div>
            <!-- 中心线 -->
            <div class="absolute left-1/2 top-0 bottom-0 w-px bg-border" />
          </div>
          <div class="flex justify-between text-xs text-muted-foreground">
            <span>-100</span>
            <span>0</span>
            <span>+100</span>
          </div>
        </div>

        <!-- 盟友信息 -->
        <div v-if="npc.allies && npc.allies.length > 0" class="pt-2 border-t">
          <p class="text-sm text-muted-foreground mb-2">{{ t('diplomacy.alliedWith') }}:</p>
          <div class="flex flex-wrap gap-1">
            <Badge
              v-for="allyId in npc.allies.slice(0, 3)"
              :key="allyId"
              variant="outline"
              class="text-xs cursor-pointer hover:bg-accent transition-colors"
              :class="getAllyBorderClass(allyId)"
              @click="scrollToAlly(allyId)"
            >
              {{ getAllyName(allyId) }}
            </Badge>
            <Badge v-if="npc.allies.length > 3" variant="outline" class="text-xs">
              +{{ npc.allies.length - 3 }} {{ t('diplomacy.more') }}
            </Badge>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="flex gap-2 pt-2">
          <Button size="sm" variant="outline" class="flex-1" @click="handleGiftResources">
            <Gift class="h-4 w-4 mr-2" />
            {{ t('diplomacy.actions.gift') }}
          </Button>
          <Button size="sm" variant="outline" class="flex-1" @click="handleViewPlanets">
            <Globe class="h-4 w-4 mr-2" />
            {{ t('diplomacy.actions.viewPlanets') }}
          </Button>
        </div>

        <!-- 最近活动 -->
        <div v-if="recentEvent" class="pt-2 border-t">
          <p class="text-xs text-muted-foreground mb-1">{{ t('diplomacy.lastEvent') }}:</p>
          <div class="flex items-center gap-2 text-xs">
            <component :is="getEventIcon(recentEvent.reason)" class="h-3 w-3" />
            <span>{{ getEventText(recentEvent.reason) }}</span>
            <span class="text-muted-foreground">
              {{ formatRelativeTime((Date.now() - recentEvent.timestamp) / 1000, t) }}{{ t('diplomacy.ago') }}
            </span>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- 备注编辑对话框 -->
    <Dialog v-model:open="noteDialogOpen">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{{ npc.note ? t('diplomacy.actions.editNote') : t('diplomacy.actions.addNote') }}</DialogTitle>
          <DialogDescription class="sr-only">{{ t('diplomacy.note') }}</DialogDescription>
        </DialogHeader>
        <div class="py-4">
          <Input v-model="noteInput" :placeholder="t('diplomacy.notePlaceholder')" @keyup.enter="saveNote" />
        </div>
        <DialogFooter>
          <Button variant="outline" @click="noteDialogOpen = false">{{ t('common.cancel') }}</Button>
          <Button @click="saveNote">{{ t('common.save') }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { useNPCStore } from '@/stores/npcStore'
  import { useGameStore } from '@/stores/gameStore'
  import { useI18n } from '@/composables/useI18n'
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
  import { Badge } from '@/components/ui/badge'
  import { Button } from '@/components/ui/button'
  import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
  import { Input } from '@/components/ui/input'
  import { Gift, Globe, Sword, Eye, Trash2, Pencil } from 'lucide-vue-next'
  import { RelationStatus, DiplomaticEventType } from '@/types/game'
  import type { DiplomaticRelation, NPC } from '@/types/game'
  import { formatRelativeTime } from '@/utils/format'

  const props = defineProps<{
    npc: NPC
    relation?: DiplomaticRelation
  }>()

  const router = useRouter()
  const npcStore = useNPCStore()
  const gameStore = useGameStore()
  const { t } = useI18n()

  // 备注对话框状态
  const noteDialogOpen = ref(false)
  const noteInput = ref('')

  // 打开备注对话框
  const openNoteDialog = () => {
    noteInput.value = props.npc.note || ''
    noteDialogOpen.value = true
  }

  // 保存备注
  const saveNote = () => {
    const npc = npcStore.npcs.find(n => n.id === props.npc.id)
    if (npc) {
      npc.note = noteInput.value.trim() || undefined
    }
    noteDialogOpen.value = false
  }

  // 好感度值
  const reputation = computed(() => props.relation?.reputation || 0)

  // 关系状态
  const status = computed(() => props.relation?.status || RelationStatus.Neutral)

  // 关系状态文本
  const statusText = computed(() => {
    switch (status.value) {
      case RelationStatus.Friendly:
        return t('diplomacy.status.friendly')
      case RelationStatus.Hostile:
        return t('diplomacy.status.hostile')
      default:
        return t('diplomacy.status.neutral')
    }
  })

  // 关系状态Badge样式
  const statusBadgeVariant = computed(() => {
    switch (status.value) {
      case RelationStatus.Friendly:
        return 'default'
      case RelationStatus.Hostile:
        return 'destructive'
      default:
        return 'secondary'
    }
  })

  // 好感度颜色
  const reputationColor = computed(() => {
    if (reputation.value >= 20) return 'text-green-600 dark:text-green-400'
    if (reputation.value <= -20) return 'text-red-600 dark:text-red-400'
    return 'text-muted-foreground'
  })

  // NPC难度等级颜色
  const difficultyLevelColor = computed(() => {
    const level = props.npc.difficultyLevel
    if (!level) return 'text-muted-foreground'
    if (level <= 1) return 'text-green-600 dark:text-green-400' // 新手
    if (level <= 2) return 'text-lime-600 dark:text-lime-400' // 简单
    if (level <= 3) return 'text-yellow-600 dark:text-yellow-400' // 普通
    if (level <= 4) return 'text-orange-600 dark:text-orange-400' // 困难
    if (level <= 5) return 'text-red-600 dark:text-red-400' // 专家
    if (level <= 6) return 'text-purple-600 dark:text-purple-400' // 大师
    return 'text-pink-600 dark:text-pink-400' // 传奇及以上
  })

  // NPC难度等级Badge样式
  const difficultyBadgeVariant = computed((): 'default' | 'secondary' | 'destructive' | 'outline' => {
    const level = props.npc.difficultyLevel
    if (!level) return 'outline'
    if (level <= 2) return 'secondary'
    if (level <= 4) return 'default'
    return 'destructive'
  })

  // 最近的外交事件
  const recentEvent = computed(() => {
    if (!props.relation?.history || props.relation.history.length === 0) return null
    return props.relation.history[props.relation.history.length - 1]
  })

  // 获取盟友名称
  const getAllyName = (allyId: string) => {
    const ally = npcStore.npcs.find(n => n.id === allyId)
    if (!ally) return allyId.substring(0, 8)
    return ally.note ? `${ally.name}(${ally.note})` : ally.name
  }

  // 获取盟友与玩家的外交关系状态对应的边框样式
  const getAllyBorderClass = (allyId: string) => {
    const ally = npcStore.npcs.find(n => n.id === allyId)
    if (!ally) return ''

    const allyRelation = ally.relations?.[gameStore.player.id]
    if (!allyRelation) return '' // 无关系，使用默认边框

    switch (allyRelation.status) {
      case RelationStatus.Friendly:
        return 'border-green-500 dark:border-green-400'
      case RelationStatus.Hostile:
        return 'border-red-500 dark:border-red-400'
      default:
        return '' // 中立，使用默认边框
    }
  }

  // 获取事件图标
  const getEventIcon = (eventType: string) => {
    switch (eventType) {
      case DiplomaticEventType.GiftResources:
        return Gift
      case DiplomaticEventType.Attack:
      case DiplomaticEventType.AllyAttacked:
        return Sword
      case DiplomaticEventType.Spy:
        return Eye
      case DiplomaticEventType.StealDebris:
        return Trash2
      default:
        return Gift
    }
  }

  // 获取事件文本
  const getEventText = (eventType: string) => {
    switch (eventType) {
      case DiplomaticEventType.GiftResources:
        return t('diplomacy.events.gift')
      case DiplomaticEventType.Attack:
        return t('diplomacy.events.attack')
      case DiplomaticEventType.AllyAttacked:
        return t('diplomacy.events.allyAttacked')
      case DiplomaticEventType.Spy:
        return t('diplomacy.events.spy')
      case DiplomaticEventType.StealDebris:
        return t('diplomacy.events.stealDebris')
      default:
        return eventType
    }
  }

  // 赠送资源
  const handleGiftResources = () => {
    // 跳转到舰队页面，自动选择第一个NPC星球
    if (props.npc.planets.length > 0) {
      const targetPlanet = props.npc.planets[0]
      if (!targetPlanet) return

      router.push({
        path: '/fleet',
        query: {
          galaxy: targetPlanet.position.galaxy,
          system: targetPlanet.position.system,
          position: targetPlanet.position.position,
          gift: '1'
        }
      })
    }
  }

  // 查看星球
  const handleViewPlanets = () => {
    // 跳转到星系视图，定位到第一个NPC星球，并传递NPC ID用于高亮
    if (props.npc.planets.length > 0) {
      const targetPlanet = props.npc.planets[0]
      if (!targetPlanet) return

      router.push({
        path: '/galaxy',
        query: {
          galaxy: targetPlanet.position.galaxy,
          system: targetPlanet.position.system,
          highlightNpc: props.npc.id
        }
      })
    }
  }

  // 滚动到盟友卡片
  const scrollToAlly = (allyId: string) => {
    // 触发父组件的滚动事件
    // 通过emit通知父组件滚动到指定的NPC卡片
    const event = new CustomEvent('scrollToNpc', { detail: { npcId: allyId }, bubbles: true })
    document.dispatchEvent(event)
  }
</script>
