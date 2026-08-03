<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogScrollContent class="max-w-md flex flex-col">
      <DialogHeader class="shrink-0 text-center">
        <DialogTitle class="text-xl flex items-center justify-center gap-2">
          <Moon class="h-6 w-6 text-yellow-400" />
          {{ t('moon.firstLoginTitle') }}
        </DialogTitle>
        <DialogDescription>
          {{ t('moon.firstLoginDesc') }}
        </DialogDescription>
      </DialogHeader>

      <div class="flex flex-col items-center gap-4 py-6">
        <div class="relative">
          <div class="h-32 w-32 rounded-full bg-gradient-to-br from-gray-200 to-gray-400 dark:from-gray-600 dark:to-gray-800 shadow-lg flex items-center justify-center">
            <Moon class="h-16 w-16 text-yellow-300 drop-shadow-lg" />
          </div>
          <div class="absolute -top-2 -right-2">
            <Gift class="h-8 w-8 text-primary animate-bounce" />
          </div>
        </div>

        <p class="text-sm text-muted-foreground text-center px-4">
          {{ t('moon.firstLoginHint') }}
        </p>
      </div>

      <DialogFooter class="shrink-0">
        <Button class="w-full" @click="handleClaim">
          <Gift class="mr-2 h-4 w-4" />
          {{ t('moon.claimGift') }}
        </Button>
      </DialogFooter>
    </DialogScrollContent>
  </Dialog>
</template>

<script setup lang="ts">
  import { useI18n } from '@/composables/useI18n'
  import { Dialog, DialogScrollContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
  import { Button } from '@/components/ui/button'
  import { Moon, Gift } from 'lucide-vue-next'

  defineProps<{
    open: boolean
  }>()

  const emit = defineEmits<{
    'update:open': [value: boolean]
  }>()

  const { t } = useI18n()

  const handleClaim = () => {
    emit('update:open', false)
  }
</script>
