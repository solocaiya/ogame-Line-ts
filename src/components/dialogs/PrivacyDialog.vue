<template>
  <Dialog v-model:open="open">
    <DialogContent class="max-w-2xl max-h-[80vh] overflow-hidden flex flex-col">
      <DialogHeader>
        <DialogTitle>{{ t('privacy.title') }}</DialogTitle>
        <DialogDescription class="sr-only">{{ t('privacy.title') }}</DialogDescription>
      </DialogHeader>
      <div class="flex-1 overflow-y-auto pr-2 space-y-4 text-sm">
        <!-- 简介 -->
        <section>
          <h3 class="font-semibold mb-1">{{ t('privacy.sections.introduction.title') }}</h3>
          <p class="text-muted-foreground">{{ t('privacy.sections.introduction.content') }}</p>
        </section>

        <!-- 数据收集 -->
        <section>
          <h3 class="font-semibold mb-1">{{ t('privacy.sections.dataCollection.title') }}</h3>
          <p class="text-muted-foreground mb-1">{{ t('privacy.sections.dataCollection.content') }}</p>
          <ul class="list-disc list-inside text-muted-foreground ml-2 space-y-0.5">
            <li>{{ t('privacy.sections.dataCollection.items.gameProgress') }}</li>
            <li>{{ t('privacy.sections.dataCollection.items.settings') }}</li>
            <li>{{ t('privacy.sections.dataCollection.items.language') }}</li>
          </ul>
        </section>

        <!-- 数据存储 -->
        <section>
          <h3 class="font-semibold mb-1">{{ t('privacy.sections.dataStorage.title') }}</h3>
          <p class="text-muted-foreground">{{ t('privacy.sections.dataStorage.content') }}</p>
        </section>

        <!-- 无服务器通信 -->
        <section>
          <h3 class="font-semibold mb-1">{{ t('privacy.sections.noServer.title') }}</h3>
          <p class="text-muted-foreground">{{ t('privacy.sections.noServer.content') }}</p>
        </section>

        <!-- 第三方服务 -->
        <section>
          <h3 class="font-semibold mb-1">{{ t('privacy.sections.thirdParty.title') }}</h3>
          <p class="text-muted-foreground">{{ t('privacy.sections.thirdParty.content') }}</p>
        </section>

        <!-- 数据控制 -->
        <section>
          <h3 class="font-semibold mb-1">{{ t('privacy.sections.dataControl.title') }}</h3>
          <p class="text-muted-foreground mb-1">{{ t('privacy.sections.dataControl.content') }}</p>
          <ul class="list-disc list-inside text-muted-foreground ml-2 space-y-0.5">
            <li>{{ t('privacy.sections.dataControl.items.export') }}</li>
            <li>{{ t('privacy.sections.dataControl.items.import') }}</li>
            <li>{{ t('privacy.sections.dataControl.items.delete') }}</li>
          </ul>
        </section>

        <!-- 联系我们 -->
        <section>
          <h3 class="font-semibold mb-1">{{ t('privacy.sections.contact.title') }}</h3>
          <p class="text-muted-foreground">
            {{ t('privacy.sections.contact.content') }}
            <a
              :href="`https://github.com/${pkg.author.name}/${pkg.name}/issues`"
              target="_blank"
              rel="noopener noreferrer"
              class="text-primary hover:underline"
            >
              GitHub Issues
            </a>
          </p>
        </section>
      </div>
      <DialogFooter class="mt-4">
        <Button variant="outline" @click="open = false">
          {{ t('common.close') }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
  import { useI18n } from '@/composables/useI18n'
  import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
  import { Button } from '@/components/ui/button'
  import pkg from '../../../package.json'

  // 双向绑定 open 状态
  const open = defineModel<boolean>('open', { default: false })

  const { t } = useI18n()
</script>
