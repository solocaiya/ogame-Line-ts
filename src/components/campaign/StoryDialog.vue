<template>
  <Dialog :open="true" @update:open="handleClose">
    <DialogContent class="max-w-2xl p-0 overflow-hidden bg-gradient-to-b from-background to-background/95">
      <!-- 可访问性标题（隐藏） -->
      <VisuallyHidden>
        <DialogTitle>{{ t('campaign.dialogue.title') }}</DialogTitle>
        <DialogDescription>{{ t('campaign.dialogue.description') }}</DialogDescription>
      </VisuallyHidden>

      <!-- 对话框头部 - 星空效果 -->
      <div class="absolute inset-0 pointer-events-none overflow-hidden">
        <div class="stars-bg" />
      </div>

      <!-- 对话内容区 -->
      <div class="relative p-6 min-h-[300px] flex flex-col">
        <!-- 说话者信息 -->
        <div v-if="currentDialogue" class="mb-4">
          <div class="flex items-center gap-3">
            <!-- 头像 -->
            <div :class="['w-12 h-12 rounded-full flex items-center justify-center', speakerStyles.bg]">
              <component :is="speakerIcon" :class="['w-6 h-6', speakerStyles.text]" />
            </div>

            <!-- 说话者名称 -->
            <div>
              <span :class="['font-semibold', speakerStyles.text]">
                {{ getSpeakerName(currentDialogue) }}
              </span>
              <div v-if="currentDialogue.speaker === 'mysterious'" class="text-xs text-muted-foreground">
                {{ t('campaign.dialogue.unknownSource') }}
              </div>
            </div>
          </div>
        </div>

        <!-- 对话文本 - 打字机效果 -->
        <div class="flex-1 min-h-[120px]">
          <div class="text-base leading-relaxed">
            <span>{{ displayedText }}</span>
            <span v-if="isTyping" class="animate-pulse">▊</span>
          </div>
        </div>

        <!-- 选项按钮 -->
        <div v-if="showChoices && currentDialogue?.choices" class="space-y-2 mt-4">
          <Button
            v-for="(choice, index) in currentDialogue.choices"
            :key="index"
            variant="outline"
            class="w-full justify-start text-left h-auto py-3 px-4"
            @click="handleChoice(choice)"
          >
            <ChevronRight class="w-4 h-4 mr-2 shrink-0" />
            <span>{{ t(choice.textKey) }}</span>
          </Button>
        </div>

        <!-- 继续按钮 -->
        <div v-if="!showChoices" class="mt-4 flex justify-end gap-2">
          <Button v-if="isTyping" variant="ghost" size="sm" @click="skipTyping">
            {{ t('campaign.dialogue.skip') }}
          </Button>
          <Button v-else @click="handleContinue" :disabled="isTyping">
            {{ isLastDialogue ? t('campaign.dialogue.finish') : t('campaign.dialogue.continue') }}
            <ChevronRight class="w-4 h-4 ml-1" />
          </Button>
        </div>

        <!-- 进度指示器 -->
        <div class="mt-4 flex justify-center gap-1">
          <div
            v-for="(_, index) in dialogues"
            :key="index"
            :class="[
              'w-2 h-2 rounded-full transition-all',
              index < currentIndex ? 'bg-primary' : index === currentIndex ? 'bg-primary/50' : 'bg-muted'
            ]"
          />
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
  import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
  import { useI18n } from '@/composables/useI18n'
  import { Dialog, DialogContent, DialogTitle, DialogDescription } from '@/components/ui/dialog'
  import { VisuallyHidden } from 'reka-ui'
  import { Button } from '@/components/ui/button'
  import type { StoryDialogue, DialogueChoice } from '@/types/game'
  import { User, Bot, HelpCircle, MessageCircle, ChevronRight } from 'lucide-vue-next'

  const props = defineProps<{
    dialogues: StoryDialogue[]
  }>()

  const emit = defineEmits<{
    close: []
    choice: [choice: DialogueChoice]
  }>()

  const { t } = useI18n()

  // 当前对话索引
  const currentIndex = ref(0)

  // 打字机效果状态
  const displayedText = ref('')
  const isTyping = ref(false)
  const typewriterInterval = ref<ReturnType<typeof setInterval> | null>(null)

  // 当前对话
  const currentDialogue = computed(() => props.dialogues[currentIndex.value])

  // 是否是最后一个对话
  const isLastDialogue = computed(() => currentIndex.value >= props.dialogues.length - 1)

  // 是否显示选项
  const showChoices = computed(() => {
    return !isTyping.value && currentDialogue.value?.choices && currentDialogue.value.choices.length > 0
  })

  // 说话者图标
  const speakerIcon = computed(() => {
    switch (currentDialogue.value?.speaker) {
      case 'player':
        return User
      case 'npc':
        return Bot
      case 'mysterious':
        return HelpCircle
      default:
        return MessageCircle
    }
  })

  // 说话者样式
  const speakerStyles = computed(() => {
    switch (currentDialogue.value?.speaker) {
      case 'player':
        return { bg: 'bg-blue-500/20', text: 'text-blue-400' }
      case 'npc':
        return { bg: 'bg-green-500/20', text: 'text-green-400' }
      case 'mysterious':
        return { bg: 'bg-purple-500/20', text: 'text-purple-400' }
      default:
        return { bg: 'bg-muted', text: 'text-muted-foreground' }
    }
  })

  // 获取说话者名称
  const getSpeakerName = (dialogue: StoryDialogue): string => {
    if (dialogue.speakerNameKey) {
      return t(dialogue.speakerNameKey)
    }
    switch (dialogue.speaker) {
      case 'player':
        return t('campaign.dialogue.player')
      case 'npc':
        return t('campaign.dialogue.npc')
      case 'mysterious':
        return t('campaign.dialogue.mysterious')
      default:
        return t('campaign.dialogue.narrator')
    }
  }

  // 开始打字机效果
  const startTypewriter = () => {
    const text = t(currentDialogue.value?.textKey || '')
    if (!text) return

    displayedText.value = ''
    isTyping.value = true
    let charIndex = 0

    typewriterInterval.value = setInterval(() => {
      if (charIndex < text.length) {
        displayedText.value += text[charIndex]
        charIndex++
      } else {
        stopTypewriter()
      }
    }, 30) // 每30ms显示一个字符
  }

  // 停止打字机效果
  const stopTypewriter = () => {
    if (typewriterInterval.value) {
      clearInterval(typewriterInterval.value)
      typewriterInterval.value = null
    }
    isTyping.value = false
  }

  // 跳过打字机效果
  const skipTyping = () => {
    stopTypewriter()
    displayedText.value = t(currentDialogue.value?.textKey || '')
  }

  // 处理继续
  const handleContinue = () => {
    if (isLastDialogue.value) {
      emit('close')
    } else {
      currentIndex.value++
    }
  }

  // 处理选项选择
  const handleChoice = (choice: DialogueChoice) => {
    emit('choice', choice)

    // 如果有下一个对话ID，跳转到对应对话
    if (choice.nextDialogueId) {
      const nextIndex = props.dialogues.findIndex(d => d.id === choice.nextDialogueId)
      if (nextIndex !== -1) {
        currentIndex.value = nextIndex
        return
      }
    }

    // 否则继续到下一个
    handleContinue()
  }

  // 处理关闭
  const handleClose = (open: boolean) => {
    if (!open) {
      emit('close')
    }
  }

  // 监听对话变化，启动打字机
  watch(
    currentIndex,
    () => {
      stopTypewriter()
      startTypewriter()
    },
    { immediate: false }
  )

  // 初始化
  onMounted(() => {
    startTypewriter()
  })

  // 清理
  onUnmounted(() => {
    stopTypewriter()
  })
</script>

<style scoped>
  .stars-bg {
    position: absolute;
    inset: 0;
    background-image: radial-gradient(circle at 10% 20%, rgba(255, 255, 255, 0.05) 1px, transparent 1px),
      radial-gradient(circle at 90% 80%, rgba(255, 255, 255, 0.03) 1px, transparent 1px),
      radial-gradient(circle at 50% 50%, rgba(255, 255, 255, 0.02) 2px, transparent 2px);
    background-size: 50px 50px, 80px 80px, 120px 120px;
    animation: twinkle 3s ease-in-out infinite;
  }

  @keyframes twinkle {
    0%,
    100% {
      opacity: 0.5;
    }
    50% {
      opacity: 1;
    }
  }
</style>
