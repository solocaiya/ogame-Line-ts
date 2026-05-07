<template>
  <Dialog :open="dialogStore.isOpen" @update:open="handleClose">
    <ScrollableDialogContent
      v-if="dialogStore.type && dialogStore.itemType"
      container-class="sm:max-w-[90vw] md:max-w-3xl lg:max-w-4xl max-h-[90vh]"
    >
      <template #header>
        <DialogHeader>
          <DialogTitle class="flex items-center gap-2">
            {{ itemTitle }}
            <Badge v-if="dialogStore.currentLevel !== undefined" variant="outline">
              {{ t('common.currentLevel') }} {{ dialogStore.currentLevel }}
            </Badge>
          </DialogTitle>
          <DialogDescription>
            {{ itemDescription }}
          </DialogDescription>
        </DialogHeader>
      </template>

      <ItemDetailView :type="dialogStore.type" :itemType="dialogStore.itemType" :currentLevel="dialogStore.currentLevel" />
    </ScrollableDialogContent>
  </Dialog>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import { Dialog, ScrollableDialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
  import { Badge } from '@/components/ui/badge'
  import { useDetailDialogStore } from '@/stores/detailDialogStore'
  import { useI18n } from '@/composables/useI18n'
  import ItemDetailView from '@/components/common/ItemDetailView.vue'

  const { t } = useI18n()
  const dialogStore = useDetailDialogStore()

  const itemTitle = computed(() => {
    if (!dialogStore.type || !dialogStore.itemType) return ''
    const typeMap = {
      building: 'buildings',
      technology: 'technologies',
      ship: 'ships',
      defense: 'defenses'
    }
    return t(`${typeMap[dialogStore.type]}.${dialogStore.itemType}`)
  })

  const itemDescription = computed(() => {
    if (!dialogStore.type || !dialogStore.itemType) return ''
    const typeMap = {
      building: 'buildingDescriptions',
      technology: 'technologyDescriptions',
      ship: 'shipDescriptions',
      defense: 'defenseDescriptions'
    }
    return t(`${typeMap[dialogStore.type]}.${dialogStore.itemType}`)
  })

  const handleClose = (open: boolean) => {
    if (!open) {
      dialogStore.close()
    }
  }
</script>
