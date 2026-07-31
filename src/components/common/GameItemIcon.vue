<template>
  <img
    v-if="src"
    :src="src"
    :alt="alt"
    :class="sizeClass"
    class="shrink-0 object-contain"
    loading="lazy"
  />
  <!-- 占位符：图标未加载时显示首字母 -->
  <div
    v-else
    :class="[sizeClass, 'shrink-0 rounded bg-muted flex items-center justify-center text-muted-foreground font-bold']"
  >
    {{ alt?.charAt(0) || '?' }}
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue'

  const props = withDefaults(
    defineProps<{
      src?: string
      alt?: string
      size?: 'sm' | 'md' | 'lg'
    }>(),
    {
      size: 'md'
    }
  )

  const sizeClass = computed(() => {
    switch (props.size) {
      case 'sm':
        return 'w-6 h-6'
      case 'md':
        return 'w-8 h-8'
      case 'lg':
        return 'w-10 h-10'
    }
  })
</script>
