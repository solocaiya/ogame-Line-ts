<template>
  <Transition
    enter-active-class="transition-all duration-300 ease-out"
    enter-from-class="-translate-y-4 opacity-0"
    enter-to-class="translate-y-0 opacity-100"
    leave-active-class="transition-all duration-200 ease-in"
    leave-from-class="translate-y-0 opacity-100"
    leave-to-class="-translate-y-4 opacity-0"
  >
    <div v-if="isHintVisible && currentHint" class="fixed top-16 right-2 max-w-[280px] z-100 pointer-events-auto">
      <div class="bg-card border rounded-lg shadow-lg p-3" role="alert" aria-live="polite">
        <!-- 标题栏 -->
        <div class="flex items-center gap-2 mb-2">
          <component :is="getIcon(currentHint.icon)" class="h-4 w-4 text-primary shrink-0" />
          <h4 class="font-medium text-sm">{{ t(currentHint.titleKey) }}</h4>
        </div>

        <!-- 内容 -->
        <p class="text-sm text-muted-foreground mb-3 line-clamp-3">{{ t(currentHint.messageKey) }}</p>

        <!-- 操作按钮 -->
        <div class="flex items-center justify-end">
          <Button size="sm" class="text-xs h-7" @click="handleDismiss">
            {{ t('hints.dontShowAgain') }}
          </Button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
  import { useHints } from '@/composables/useHints'
  import { useI18n } from '@/composables/useI18n'
  import { Button } from '@/components/ui/button'
  import {
    Home,
    Building,
    FlaskConical,
    Rocket,
    Plane,
    Globe,
    Handshake,
    Mail,
    Shield,
    Lightbulb,
    Users,
    Swords,
    Settings,
    Wand2
  } from 'lucide-vue-next'
  import type { Component } from 'vue'

  const { t } = useI18n()
  const { currentHint, isHintVisible, dismissHint } = useHints()

  // 图标名称到组件的映射
  const iconMap: Record<string, Component> = {
    home: Home,
    building: Building,
    flask: FlaskConical,
    rocket: Rocket,
    plane: Plane,
    globe: Globe,
    handshake: Handshake,
    mail: Mail,
    shield: Shield,
    users: Users,
    swords: Swords,
    settings: Settings,
    wand: Wand2
  }

  const getIcon = (iconName?: string) => {
    if (!iconName) return Lightbulb
    return iconMap[iconName] || Lightbulb
  }

  // 不再显示 - 永久关闭
  const handleDismiss = () => {
    dismissHint(true)
  }
</script>
