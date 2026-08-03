<template>
  <Teleport to="body">
    <div v-if="isActive" class="fixed inset-0 z-50">
      <!-- 遮罩层 -->
      <div class="absolute inset-0 bg-black/60" @click="handleOverlayClick" />

      <!-- 高亮目标元素（通过 CSS 实现） -->
      <div
        v-if="currentStep.target && targetRect"
        class="absolute pointer-events-none z-10 ring-2 ring-primary ring-offset-2 rounded-lg transition-all duration-300"
        :style="highlightStyle"
      />

      <!-- 引导卡片 -->
      <div
        class="absolute z-20 bg-card border border-border rounded-xl shadow-2xl p-6 max-w-sm animate-in fade-in"
        :style="cardStyle"
      >
        <!-- 步骤指示器 -->
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center gap-1">
            <div
              v-for="(_, i) in totalSteps"
              :key="i"
              class="h-1.5 w-4 rounded-full transition-colors"
              :class="i < currentStepIndex ? 'bg-primary' : i === currentStepIndex ? 'bg-primary/70' : 'bg-muted'"
            />
          </div>
          <span class="text-xs text-muted-foreground">
            {{ currentStepIndex + 1 }} / {{ totalSteps }}
          </span>
        </div>

        <h3 class="text-lg font-bold mb-2">{{ t(currentStep.title) }}</h3>
        <p class="text-sm text-muted-foreground mb-4">{{ t(currentStep.content) }}</p>

        <div class="flex items-center justify-between">
          <Button
            v-if="currentStep.canSkip"
            variant="ghost"
            size="sm"
            @click="skipTutorial"
          >
            {{ t('tutorial.skip') }}
          </Button>
          <div v-else />
          <Button @click="nextStep">
            {{ isLastStep ? t('tutorial.finish') : t('tutorial.next') }}
          </Button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
  import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
  import { useGameStore } from '@/stores/gameStore'
  import { useI18n } from '@/composables/useI18n'
  import { Button } from '@/components/ui/button'
  import { TUTORIAL_STEPS } from '@/config/tutorialConfig'
  import { playSound, SoundType } from '@/logic/soundManager'

  const gameStore = useGameStore()
  const { t } = useI18n()

  const isActive = ref(false)
  const currentStepIndex = ref(0)
  const targetRect = ref<DOMRect | null>(null)

  const totalSteps = TUTORIAL_STEPS.length
  const currentStep = computed(() => TUTORIAL_STEPS[currentStepIndex.value])
  const isLastStep = computed(() => currentStepIndex.value === totalSteps - 1)

  const highlightStyle = computed(() => {
    if (!targetRect.value) return {}
    return {
      top: `${targetRect.value.top - 4}px`,
      left: `${targetRect.value.left - 4}px`,
      width: `${targetRect.value.width + 8}px`,
      height: `${targetRect.value.height + 8}px`
    }
  })

  const cardStyle = computed(() => {
    if (!targetRect.value || currentStep.value.placement === 'center') {
      return { top: '50%', left: '50%', transform: 'translate(-50%, -50%)' }
    }

    const placement = currentStep.value.placement || 'bottom'
    const rect = targetRect.value
    const style: Record<string, string> = {}

    switch (placement) {
      case 'bottom':
        style.top = `${rect.bottom + 16}px`
        style.left = `${Math.max(16, rect.left + rect.width / 2 - 160)}px`
        break
      case 'top':
        style.bottom = `${window.innerHeight - rect.top + 16}px`
        style.left = `${Math.max(16, rect.left + rect.width / 2 - 160)}px`
        break
      case 'right':
        style.top = `${rect.top}px`
        style.left = `${rect.right + 16}px`
        break
      case 'left':
        style.top = `${rect.top}px`
        style.right = `${window.innerWidth - rect.left + 16}px`
        break
    }

    return style
  })

  const updateTargetRect = () => {
    if (!currentStep.value.target) {
      targetRect.value = null
      return
    }

    const el = document.querySelector(`[data-tutorial="${currentStep.value.target}"]`)
    if (el) {
      targetRect.value = el.getBoundingClientRect()
      el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    } else {
      targetRect.value = null
    }
  }

  const nextStep = () => {
    playSound(SoundType.Click, { enabled: gameStore.player.soundEnabled !== false, volume: (gameStore.player.soundVolume ?? 0.7) * 0.3 })

    if (isLastStep.value) {
      completeTutorial()
      return
    }

    currentStepIndex.value++
    const nextStepData = TUTORIAL_STEPS[currentStepIndex.value]

    if (nextStepData.route) {
      window.location.hash = `#${nextStepData.route}`
    }

    nextTick(() => {
      setTimeout(updateTargetRect, 300)
    })
  }

  const skipTutorial = () => {
    completeTutorial()
  }

  const completeTutorial = () => {
    isActive.value = false
    gameStore.player.tutorialCompleted = true
    playSound(SoundType.AchievementUnlock, { enabled: gameStore.player.soundEnabled !== false, volume: gameStore.player.soundVolume ?? 0.7 })
  }

  const handleOverlayClick = () => {
    // 不响应遮罩点击，必须点按钮
  }

  const startTutorial = () => {
    currentStepIndex.value = 0
    isActive.value = true
    const firstStep = TUTORIAL_STEPS[0]
    if (firstStep.route) {
      window.location.hash = `#${firstStep.route}`
    }
    nextTick(() => {
      setTimeout(updateTargetRect, 300)
    })
  }

  // 监听窗口大小变化
  const handleResize = () => {
    if (isActive.value) {
      updateTargetRect()
    }
  }

  // 监听步骤变化
  watch(currentStepIndex, () => {
    nextTick(() => {
      setTimeout(updateTargetRect, 300)
    })
  })

  onMounted(() => {
    window.addEventListener('resize', handleResize)

    // 如果未完成新手引导，自动开始
    if (!gameStore.player.tutorialCompleted) {
      startTutorial()
    }
  })

  onUnmounted(() => {
    window.removeEventListener('resize', handleResize)
  })

  defineExpose({ startTutorial, isActive })
</script>
