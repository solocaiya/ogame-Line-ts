<template>
  <span :class="[colors[type], sizes[size], 'inline-block rounded shrink-0']" />
</template>

<script setup lang="ts">
  interface Props {
    type: 'metal' | 'crystal' | 'deuterium' | 'darkMatter' | 'energy'
    size?: 'sm' | 'md' | 'lg'
  }

  withDefaults(defineProps<Props>(), {
    size: 'md'
  })

  // 使用纯色背景，在 Android WebView 中更可靠
  const colors = {
    metal: 'bg-slate-500',
    crystal: 'bg-cyan-500',
    deuterium: 'bg-green-500',
    darkMatter: 'bg-purple-700',
    energy: 'bg-yellow-500'
  }

  const sizes = {
    sm: 'w-3 h-3 min-w-3 min-h-3',
    md: 'w-4 h-4 min-w-4 min-h-4',
    lg: 'w-5 h-5 min-w-5 min-h-5'
  }
</script>
