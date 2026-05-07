<template>
  <Transition
    enter-active-class="transition-all duration-300 ease-out"
    enter-from-class="translate-y-4 opacity-0"
    enter-to-class="translate-y-0 opacity-100"
    leave-active-class="transition-all duration-200 ease-in"
    leave-from-class="translate-y-0 opacity-100"
    leave-to-class="translate-y-4 opacity-0"
  >
    <Button v-if="isVisible" variant="outline" size="icon" @click="scrollToTop">
      <ChevronUp class="h-4 w-4" />
    </Button>
  </Transition>
</template>

<script setup lang="ts">
  import { ref, onMounted, onUnmounted } from 'vue'
  import { Button } from '@/components/ui/button'
  import { ChevronUp } from 'lucide-vue-next'

  // 显示阈值（滚动超过这个距离才显示按钮）
  const SCROLL_THRESHOLD = 300

  const isVisible = ref(false)

  // 监听滚动事件
  const handleScroll = () => {
    isVisible.value = window.scrollY > SCROLL_THRESHOLD
  }

  // 丝滑返回顶部
  const scrollToTop = () => {
    window.scrollTo({
      top: 0,
      behavior: 'smooth'
    })
  }

  onMounted(() => {
    window.addEventListener('scroll', handleScroll, { passive: true })
    // 初始检查
    handleScroll()
  })

  onUnmounted(() => {
    window.removeEventListener('scroll', handleScroll)
  })
</script>
