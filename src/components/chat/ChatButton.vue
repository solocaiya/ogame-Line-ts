<template>
  <div class="relative">
    <Popover v-model:open="popoverOpen" @update:open="onPopoverUpdate">
      <PopoverTrigger as-child>
        <Button
          variant="outline"
          size="icon"
          class="relative h-10 w-10 rounded-full shadow-lg"
          :class="{ 'animate-pulse': chatStore.lastSenderName && !popoverOpen }"
        >
          <MessageCircle :class="['h-5 w-5', hasUnread ? 'text-blue-400' : '']" />
          <Badge
            v-if="hasUnread"
            class="absolute -top-1 -right-1 h-5 w-5 flex items-center justify-center p-0 text-[10px]"
          >
            {{ displayUnread }}
          </Badge>
        </Button>
      </PopoverTrigger>
      <PopoverContent
        class="w-96 p-0"
        align="end"
        :side-offset="8"
      >
        <!-- Header with tabs -->
        <div class="border-b p-3 pb-0">
          <div class="flex items-center justify-between mb-2">
            <h3 class="font-semibold text-sm">{{ t('chat.title') }}</h3>
            <Button
              variant="ghost"
              size="icon"
              class="h-7 w-7"
              @click="showDND = !showDND"
            >
              <BellOff :class="['h-4 w-4', isAnyMuted ? 'text-yellow-500' : '']" />
            </Button>
          </div>

          <!-- DND selector -->
          <div v-if="showDND" class="mb-2 p-2 bg-muted/50 rounded-md">
            <p class="text-xs text-muted-foreground mb-1">{{ t('chat.dnd.title') }}</p>
            <Select :model-value="chatStore.dndMode" @update:model-value="(v: string) => chatStore.setDNDMode(v as any)">
              <SelectTrigger class="h-7 text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="none">{{ t('chat.dnd.none') }}</SelectItem>
                <SelectItem value="mute_world">{{ t('chat.dnd.mute_world') }}</SelectItem>
                <SelectItem value="mute_alliance">{{ t('chat.dnd.mute_alliance') }}</SelectItem>
                <SelectItem value="mute_all">{{ t('chat.dnd.mute_all') }}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <Tabs :model-value="chatStore.activeChannel" @update:model-value="onChannelChange">
            <TabsList class="grid w-full grid-cols-2">
              <TabsTrigger value="world" class="text-xs">
                {{ t('chat.world') }}
                <Badge v-if="chatStore.unreadWorld > 0" variant="destructive" class="ml-1 h-4 min-w-4 px-1 text-[9px]">
                  {{ chatStore.unreadWorld }}
                </Badge>
              </TabsTrigger>
              <TabsTrigger value="alliance" class="text-xs" :disabled="!chatStore.canUseAllianceChannel">
                {{ t('chat.alliance') }}
                <Badge v-if="chatStore.unreadAlliance > 0" variant="destructive" class="ml-1 h-4 min-w-4 px-1 text-[9px]">
                  {{ chatStore.unreadAlliance }}
                </Badge>
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </div>

        <!-- Messages area -->
        <ScrollArea class="h-80">
          <div class="p-3">
            <!-- Load more button -->
            <div v-if="chatStore.hasMore" class="mb-2">
              <Button
                variant="ghost"
                size="sm"
                class="w-full text-xs h-7"
                :disabled="loadingMore"
                @click="loadMore"
              >
                <Loader2 v-if="loadingMore" class="h-3 w-3 mr-1 animate-spin" />
                {{ t('chat.loadMore') }}
              </Button>
            </div>

            <!-- No messages -->
            <div v-if="chatStore.currentMessages.length === 0" class="flex flex-col items-center justify-center py-8 text-muted-foreground">
              <MessageCircle class="h-8 w-8 mb-2 opacity-50" />
              <p class="text-xs">{{ t('chat.noMessages') }}</p>
            </div>

            <!-- Messages -->
            <div v-else class="space-y-2">
              <div
                v-for="msg in chatStore.currentMessages"
                :key="msg.id"
                class="text-xs"
              >
                <div class="flex items-baseline gap-1">
                  <span class="font-semibold text-primary">
                    <span v-if="msg.senderTag" class="text-muted-foreground">[{{ msg.senderTag }}]</span>
                    {{ msg.senderName }}
                  </span>
                  <span class="text-muted-foreground text-[10px]">
                    {{ formatTime(msg.timestamp) }}
                  </span>
                </div>
                <p class="text-foreground/90 break-words">{{ msg.content }}</p>
              </div>
            </div>
          </div>
        </ScrollArea>

        <!-- Input area -->
        <div class="border-t p-3">
          <div class="flex gap-2">
            <Input
              v-model="messageInput"
              :placeholder="t('chat.inputPlaceholder')"
              class="h-8 text-xs flex-1"
              maxlength="500"
              @keydown.enter="sendMessage"
            />
            <Button
              size="sm"
              class="h-8 px-3"
              :disabled="!messageInput.trim()"
              @click="sendMessage"
            >
              <Send class="h-3 w-3" />
            </Button>
          </div>
        </div>
      </PopoverContent>
    </Popover>

    <!-- Bubble notification -->
    <Transition name="fade">
      <div
        v-if="showBubble && !popoverOpen"
        class="absolute bottom-12 right-0 bg-popover border rounded-lg shadow-lg px-3 py-2 max-w-48 z-50"
      >
        <p class="text-xs">
          <span class="font-semibold">{{ chatStore.lastSenderName }}</span>
          <span class="text-muted-foreground ml-1">{{ t('chat.newMessage') }}</span>
        </p>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { MessageCircle, Send, BellOff, Loader2 } from 'lucide-vue-next'
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useChatStore } from '@/stores/chatStore'
import type { ChatChannelType } from '@/types/game'

const { t } = useI18n()
const chatStore = useChatStore()

const popoverOpen = ref(false)
const messageInput = ref('')
const showDND = ref(false)
const showBubble = ref(false)
const loadingMore = ref(false)
let bubbleTimer: ReturnType<typeof setTimeout> | null = null

const hasUnread = computed(() => chatStore.totalUnread > 0)
const displayUnread = computed(() => chatStore.totalUnread > 99 ? '99+' : chatStore.totalUnread)
const isAnyMuted = computed(() => chatStore.dndMode !== 'none')

function onPopoverUpdate(open: boolean) {
  if (open) {
    chatStore.open()
    showBubble.value = false
    if (bubbleTimer) {
      clearTimeout(bubbleTimer)
      bubbleTimer = null
    }
  } else {
    chatStore.close()
  }
}

function onChannelChange(channel: string) {
  chatStore.setActiveChannel(channel as ChatChannelType)
}

async function sendMessage() {
  const content = messageInput.value.trim()
  if (!content) return
  messageInput.value = ''
  try {
    await chatStore.sendMessage(content)
  } catch (e) {
    console.error('Failed to send message:', e)
  }
}

async function loadMore() {
  loadingMore.value = true
  await chatStore.loadMore()
  loadingMore.value = false
}

function formatTime(timestamp: number): string {
  const date = new Date(timestamp)
  const now = new Date()
  const isToday = date.toDateString() === now.toDateString()
  if (isToday) {
    return date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
  }
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' }) +
    ' ' + date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
}

// Watch for new messages to show bubble notification
watch(() => chatStore.lastSenderName, (name) => {
  if (name && !popoverOpen.value) {
    showBubble.value = true
    if (bubbleTimer) clearTimeout(bubbleTimer)
    bubbleTimer = setTimeout(() => {
      showBubble.value = false
    }, 3000)
  }
})

onMounted(() => {
  chatStore.init()
})

onUnmounted(() => {
  if (bubbleTimer) clearTimeout(bubbleTimer)
})
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
