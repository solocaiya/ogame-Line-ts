<template>
  <component :is="iconComponent" :class="[colors[type], sizes[size], 'shrink-0']" :stroke-width="2" />
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import { Cog, Gem, Droplets, Zap, Atom } from 'lucide-vue-next'

  interface Props {
    type: 'metal' | 'crystal' | 'deuterium' | 'darkMatter' | 'energy'
    size?: 'sm' | 'md' | 'lg'
  }

  const props = withDefaults(defineProps<Props>(), {
    size: 'md'
  })

  const iconComponent = computed(() => {
    const map = {
      metal: Cog,
      crystal: Gem,
      deuterium: Droplets,
      energy: Zap,
      darkMatter: Atom
    }
    return map[props.type]
  })

  const colors = {
    metal: 'text-slate-400',
    crystal: 'text-cyan-400',
    deuterium: 'text-green-400',
    darkMatter: 'text-purple-400',
    energy: 'text-yellow-400'
  }

  const sizes = {
    sm: 'w-3.5 h-3.5',
    md: 'w-4 h-4',
    lg: 'w-5 h-5'
  }
</script>
