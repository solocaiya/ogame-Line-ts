<template>
  <button
    type="button"
    role="switch"
    :aria-checked="checked"
    :data-state="checked ? 'checked' : 'unchecked'"
    :disabled="disabled"
    :class="
      cn(
        'peer inline-flex h-6 w-11 shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:cursor-not-allowed disabled:opacity-50',
        checked ? 'bg-primary' : 'bg-input',
        props.class
      )
    "
    @click="toggle"
  >
    <span
      :data-state="checked ? 'checked' : 'unchecked'"
      :class="
        cn(
          'pointer-events-none block h-5 w-5 rounded-full bg-background shadow-lg ring-0 transition-transform',
          checked ? 'translate-x-5' : 'translate-x-0'
        )
      "
    />
  </button>
</template>

<script setup lang="ts">
  import { cn } from '@/lib/utils'
  import { type HTMLAttributes } from 'vue'

  const props = defineProps<{
    checked?: boolean
    disabled?: boolean
    class?: HTMLAttributes['class']
  }>()

  const emit = defineEmits<{
    (e: 'update:checked', value: boolean): void
  }>()

  const toggle = () => {
    if (props.disabled) return
    emit('update:checked', !props.checked)
  }
</script>

