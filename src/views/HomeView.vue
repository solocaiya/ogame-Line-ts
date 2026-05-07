<template>
  <div class="min-h-screen flex flex-col items-center justify-center bg-gradient-to-b from-background to-muted/50 p-4">
    <!-- Logo 和标题 -->
    <div class="text-center mb-8 animate-fade-in">
      <img src="@/assets/logo.svg" alt="OGame Logo" class="w-24 h-24 mx-auto mb-4" />
      <h1 class="text-4xl sm:text-5xl font-bold bg-gradient-to-r from-primary to-primary/60 bg-clip-text text-transparent">
        {{ pkg.title }}
      </h1>
      <p class="text-muted-foreground mt-2 text-sm sm:text-base">{{ t('home.subtitle') }}</p>
    </div>

    <!-- 开始游戏按钮 -->
    <div class="flex flex-col gap-4 w-full max-w-xs mb-8">
      <Button size="lg" class="w-full text-lg h-14" @click="handleStartGame">
        <Rocket class="mr-2 h-5 w-5" />
        {{ t('home.startGame') }}
      </Button>
    </div>

    <!-- 底部操作按钮 -->
    <div class="flex flex-wrap items-center justify-center gap-4">
      <!-- 语言切换 -->
      <Popover>
        <PopoverTrigger as-child>
          <Button variant="outline" size="sm">
            <Languages class="mr-2 h-4 w-4" />
            {{ localeNames[gameStore.locale] }}
          </Button>
        </PopoverTrigger>
        <PopoverContent class="w-48 p-2" side="top">
          <div class="space-y-1">
            <Button
              v-for="locale in availableLocales"
              :key="locale"
              @click="gameStore.locale = locale"
              :variant="gameStore.locale === locale ? 'secondary' : 'ghost'"
              class="w-full justify-start"
              size="sm"
            >
              {{ localeNames[locale] }}
            </Button>
          </div>
        </PopoverContent>
      </Popover>

      <!-- 隐私协议按钮 -->
      <Button variant="ghost" size="sm" @click="showPrivacyDialog = true">
        <Shield class="mr-2 h-4 w-4" />
        {{ t('settings.privacyPolicy') }}
      </Button>
    </div>

    <!-- 隐私协议弹窗 -->
    <PrivacyDialog v-model:open="showPrivacyDialog" />

    <!-- 隐私协议同意确认弹窗（开始游戏时） -->
    <AlertDialog v-model:open="showAgreementDialog">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{{ t('home.privacyAgreement') }}</AlertDialogTitle>
          <AlertDialogDescription>
            {{ t('home.privacyAgreementDesc') }}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <div class="my-4 max-h-48 overflow-y-auto text-sm text-muted-foreground border rounded-lg p-3">
          <p>{{ t('privacy.sections.introduction.content') }}</p>
          <p class="mt-2">{{ t('privacy.sections.dataCollection.content') }}</p>
          <p class="mt-2">{{ t('privacy.sections.thirdParty.content') }}</p>
        </div>
        <div class="flex items-center gap-2 mb-4">
          <Checkbox id="privacy-agree" v-model="privacyAgreed" />
          <label for="privacy-agree" class="text-sm cursor-pointer">
            {{ t('home.agreeToPrivacy') }}
            <Button variant="link" class="p-0 h-auto text-sm" @click.prevent="showPrivacyDialog = true">
              {{ t('home.viewFullPolicy') }}
            </Button>
          </label>
        </div>
        <AlertDialogFooter>
          <AlertDialogCancel @click="privacyAgreed = false">{{ t('common.cancel') }}</AlertDialogCancel>
          <AlertDialogAction :disabled="!privacyAgreed" @click="confirmStartGame">
            {{ t('home.agreeAndStart') }}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>

<script setup lang="ts">
  import { ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { useGameStore } from '@/stores/gameStore'
  import { useI18n } from '@/composables/useI18n'
  import { localeNames, type Locale } from '@/locales'
  import { Button } from '@/components/ui/button'
  import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover'
  import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle
  } from '@/components/ui/alert-dialog'
  import { Checkbox } from '@/components/ui/checkbox'
  import { Rocket, Languages, Shield } from 'lucide-vue-next'
  import PrivacyDialog from '@/components/dialogs/PrivacyDialog.vue'
  import pkg from '../../package.json'

  const router = useRouter()
  const gameStore = useGameStore()
  const { t } = useI18n()

  const showPrivacyDialog = ref(false)
  const showAgreementDialog = ref(false)
  const privacyAgreed = ref(false)

  const availableLocales: Locale[] = ['zh-CN', 'zh-TW', 'en', 'de', 'ru', 'ko', 'ja']

  // 处理开始游戏
  const handleStartGame = () => {
    // 如果已经同意过，直接进入总览页面
    if (gameStore.player.privacyAgreed) {
      router.push('/overview')
    } else {
      // 显示同意弹窗
      showAgreementDialog.value = true
    }
  }

  // 确认开始游戏
  const confirmStartGame = () => {
    if (privacyAgreed.value) {
      gameStore.player.privacyAgreed = true
      showAgreementDialog.value = false
      router.push('/overview')
    }
  }
</script>

<style scoped>
  .animate-fade-in {
    animation: fadeIn 0.5s ease-out;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translate3d(0, -10px, 0);
    }
    to {
      opacity: 1;
      transform: translate3d(0, 0, 0);
    }
  }
</style>
