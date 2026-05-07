<template>
  <Dialog v-model:open="isOpen">
    <DialogContent class="max-w-md">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2">
          <Eye class="h-5 w-5" />
          {{ t('messagesView.spiedNotificationDetails') }}
        </DialogTitle>
        <DialogDescription class="sr-only">
          {{ t('messagesView.spyDetected') }}
        </DialogDescription>
      </DialogHeader>

      <div v-if="notification" class="space-y-4">
        <!-- 侦查者信息 -->
        <div class="p-4 bg-muted/50 rounded-lg border">
          <div class="flex items-center gap-3">
            <div class="p-2 bg-background rounded-full border">
              <AlertTriangle class="h-5 w-5" />
            </div>
            <div class="flex-1">
              <div class="flex items-center gap-2 mb-1">
                <h3 class="font-semibold">{{ npcName }}</h3>
                <Badge :variant="notification.detectionSuccess ? 'destructive' : 'secondary'" class="text-xs">
                  {{ notification.detectionSuccess ? t('messagesView.detected') : t('messagesView.undetected') }}
                </Badge>
              </div>
              <p class="text-sm text-muted-foreground">
                {{ formatDate(notification.timestamp) }}
              </p>
            </div>
          </div>
        </div>

        <!-- 被侦查星球 -->
        <div class="space-y-2">
          <h4 class="font-semibold text-sm">{{ t('messagesView.targetPlanet') }}</h4>
          <div class="p-3 bg-muted/30 rounded-md border flex items-center gap-2">
            <Globe class="h-4 w-4" />
            <span class="font-medium">{{ notification.targetPlanetName }}</span>
          </div>
        </div>

        <!-- 消息内容 -->
        <div class="space-y-2">
          <h4 class="font-semibold text-sm">{{ t('messagesView.detectionResult') }}</h4>
          <div class="p-3 bg-muted/30 rounded-md border">
            <p class="text-sm">
              {{
                t('messagesView.spiedNotificationMessage', {
                  npc: npcName,
                  planet: notification.targetPlanetName
                })
              }}
            </p>
          </div>
        </div>

        <!-- 建议 -->
        <div class="p-3 bg-muted/30 rounded-md border">
          <p class="text-sm text-muted-foreground">
            {{ t('messagesView.spiedNotificationTip') }}
          </p>
        </div>
      </div>

      <DialogFooter>
        <Button variant="outline" @click="isOpen = false">{{ t('common.close') }}</Button>
        <Button @click="viewNPCInGalaxy">{{ t('messagesView.viewInGalaxy') }}</Button>
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
  import { Eye, AlertTriangle, Globe } from 'lucide-vue-next'
  import type { SpiedNotification } from '@/types/game'

  const props = defineProps<{
    notification: SpiedNotification | null
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

  // 在银河系中查看NPC
  const viewNPCInGalaxy = () => {
    if (!props.notification?.npcId) return
    const npc = npcStore.npcs.find(n => n.id === props.notification!.npcId)
    if (npc && npc.planets && npc.planets.length > 0) {
      isOpen.value = false
      const homePlanet = npc.planets[0]?.position
      if (homePlanet) {
        router.push(`/galaxy?galaxy=${homePlanet.galaxy}&system=${homePlanet.system}`)
      }
    }
  }
</script>
