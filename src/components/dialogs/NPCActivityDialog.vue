<template>
  <Dialog v-model:open="isOpen">
    <DialogContent class="max-w-2xl">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2">
          <Recycle class="h-5 w-5" />
          {{ t('messagesView.npcActivityDetails') }}
        </DialogTitle>
        <DialogDescription>
          {{ t('messagesView.activityDescription') }}
        </DialogDescription>
      </DialogHeader>

      <div v-if="notification" class="space-y-4">
        <!-- NPC信息 -->
        <div class="p-4 bg-muted/50 rounded-lg border">
          <div class="flex items-center gap-2 mb-2">
            <h3 class="font-semibold text-lg">{{ npcName }}</h3>
            <Badge variant="secondary">{{ t('messagesView.activityType.' + notification.activityType) }}</Badge>
          </div>
          <p class="text-sm text-muted-foreground">
            {{ formatDate(notification.timestamp) }}
          </p>
        </div>

        <!-- 活动位置 -->
        <div class="space-y-2">
          <h4 class="font-semibold text-sm">{{ t('messagesView.activityLocation') }}</h4>
          <div class="p-3 bg-muted/30 rounded-md">
            <div class="flex items-center gap-2 mb-2">
              <Globe class="h-4 w-4" />
              <span class="font-medium">
                {{ t('messagesView.position') }}: [{{ notification.targetPosition.galaxy }}:{{
                  notification.targetPosition.system
                }}:{{ notification.targetPosition.position }}]
              </span>
            </div>
            <p v-if="notification.targetPlanetName" class="text-sm text-muted-foreground">
              {{ t('messagesView.nearPlanet') }}: {{ notification.targetPlanetName }}
            </p>
          </div>
        </div>

        <!-- 活动描述 -->
        <div class="space-y-2">
          <h4 class="font-semibold text-sm">{{ t('messagesView.activityDescription') }}</h4>
          <div class="p-3 bg-muted/30 rounded-md">
            <p class="text-sm">
              {{
                t('messagesView.npcActivityMessage', {
                  npc: npcName,
                  activity: t('messagesView.activityType.' + notification.activityType),
                  position: `[${notification.targetPosition.galaxy}:${notification.targetPosition.system}:${notification.targetPosition.position}]`
                })
              }}
            </p>
          </div>
        </div>

        <!-- 到达时间 -->
        <div class="space-y-2">
          <h4 class="font-semibold text-sm">{{ t('messagesView.arrivalTime') }}</h4>
          <div class="p-3 bg-muted/30 rounded-md">
            <p class="font-medium">{{ formatDate(notification.arrivalTime) }}</p>
          </div>
        </div>

        <!-- 提示信息 -->
        <div class="p-3 bg-muted/50 rounded-md border">
          <p class="text-sm text-muted-foreground">
            {{ t('messagesView.npcActivityTip') }}
          </p>
        </div>
      </div>

      <DialogFooter>
        <Button variant="outline" @click="isOpen = false">{{ t('common.close') }}</Button>
        <Button @click="viewLocationInGalaxy">{{ t('messagesView.viewInGalaxy') }}</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
  import { ref, watch, computed } from 'vue'
  import { useRouter } from 'vue-router'
  import { useNPCStore } from '@/stores/npcStore'
  import { useI18n } from '@/composables/useI18n'
  import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
  import { Button } from '@/components/ui/button'
  import { Badge } from '@/components/ui/badge'
  import { formatDate } from '@/utils/format'
  import { Recycle, Globe } from 'lucide-vue-next'
  import type { NPCActivityNotification } from '@/types/game'

  const props = defineProps<{
    notification: NPCActivityNotification | null
    open: boolean
  }>()

  const emit = defineEmits<{
    (e: 'update:open', value: boolean): void
  }>()

  const router = useRouter()
  const npcStore = useNPCStore()
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

  // 获取NPC名称
  const npcName = computed(() => {
    if (!props.notification) return ''
    if (!npcStore.npcs?.length) return props.notification.npcName

    // 通过 npcId 查找
    if (props.notification.npcId) {
      const npc = npcStore.npcs.find(n => n.id === props.notification!.npcId)
      if (npc) return npc.name
    }

    // 尝试从旧名称中提取ID并查找
    const idMatch = props.notification.npcName.match(/npc_\d+/)
    if (idMatch) {
      const extractedId = idMatch[0]
      const npc = npcStore.npcs.find(n => n.id === extractedId)
      if (npc) return npc.name
    }

    return props.notification.npcName
  })

  // 在银河系中查看位置
  const viewLocationInGalaxy = () => {
    if (!props.notification?.targetPosition) return
    isOpen.value = false
    router.push(
      `/galaxy?galaxy=${props.notification.targetPosition.galaxy}&system=${props.notification.targetPosition.system}`
    )
  }
</script>
