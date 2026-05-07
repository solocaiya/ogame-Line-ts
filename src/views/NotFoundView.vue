<template>
  <div class="container mx-auto p-4 sm:p-6 flex items-center justify-center min-h-[60vh]">
    <Empty class="border-0">
      <EmptyMedia>
        <div class="text-8xl sm:text-9xl font-bold text-muted-foreground/20">404</div>
      </EmptyMedia>
      <EmptyHeader>
        <EmptyTitle>{{ t('notFound.title') }}</EmptyTitle>
        <EmptyDescription>{{ t('notFound.description') }}</EmptyDescription>
      </EmptyHeader>
      <EmptyContent>
        <Button @click="goHome" size="lg">
          <Home class="mr-2 h-4 w-4" />
          {{ t('notFound.goHome') }}
        </Button>
      </EmptyContent>
    </Empty>
  </div>
</template>

<script setup lang="ts">
  import { useRouter } from 'vue-router'
  import { useI18n } from '@/composables/useI18n'
  import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
  import { Button } from '@/components/ui/button'
  import { Home } from 'lucide-vue-next'

  const router = useRouter()
  const { t } = useI18n()

  const goHome = () => {
    router.push('/')
  }
</script>
