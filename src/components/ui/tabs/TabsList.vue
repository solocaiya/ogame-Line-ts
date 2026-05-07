<template>
  <TabsListRoot
    data-slot="tabs-list"
    v-bind="delegatedProps"
    :class="
      cn(
        'bg-muted text-muted-foreground inline-flex w-fit items-center justify-center rounded-lg p-[3px]',
        tabCount && tabCount > 3 ? (tabCount > 6 ? 'h-[85px] sm:h-9' : 'h-[65px] sm:h-9') : 'h-9',
        props.class
      )
    "
  >
    <slot />
  </TabsListRoot>
</template>

<script setup lang="ts">
  import type { TabsListProps } from 'reka-ui'
  import type { HTMLAttributes } from 'vue'
  import { reactiveOmit } from '@vueuse/core'
  import { TabsList as TabsListRoot } from 'reka-ui'
  import { cn } from '@/lib/utils'

  const props = defineProps<TabsListProps & { class?: HTMLAttributes['class']; tabCount?: number }>()

  const delegatedProps = reactiveOmit(props, 'class', 'tabCount')
</script>
