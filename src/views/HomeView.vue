<template>
  <div class="min-h-screen flex flex-col items-center justify-center bg-gradient-to-b from-background to-muted/50 p-4">
    <!-- Logo 和标题 -->
    <div class="text-center mb-8 animate-fade-in">
      <img src="@/assets/logo.png" alt="OGame Logo" class="w-24 h-24 mx-auto mb-4" />
      <h1 class="text-4xl sm:text-5xl font-bold bg-gradient-to-r from-primary to-primary/60 bg-clip-text text-transparent">
        {{ pkg.title }}
      </h1>
      <p class="text-muted-foreground mt-2 text-sm sm:text-base">{{ t('home.subtitle') }}</p>
    </div>

    <!-- 登录/注册按钮 -->
    <div class="flex flex-col gap-4 w-full max-w-xs mb-8">
      <Button size="lg" class="w-full text-lg h-14" @click="goToLogin">
        <LogIn class="mr-2 h-5 w-5" />
        {{ t('home.loginRegister') }}
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
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { useRouter } from 'vue-router'
  import { useGameStore } from '@/stores/gameStore'
  import { useAuthStore } from '@/stores/authStore'
  import { useI18n } from '@/composables/useI18n'
  import { localeNames, type Locale } from '@/locales'
  import { Button } from '@/components/ui/button'
  import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover'
  import { LogIn, Languages, Shield } from 'lucide-vue-next'
  import PrivacyDialog from '@/components/dialogs/PrivacyDialog.vue'
  import pkg from '../../package.json'

  const router = useRouter()
  const gameStore = useGameStore()
  const authStore = useAuthStore()
  const { t } = useI18n()

  const showPrivacyDialog = ref(false)

  const availableLocales: Locale[] = ['zh-CN', 'zh-TW', 'en', 'de', 'ru', 'ko', 'ja']

  // 已登录且同意隐私协议 → 自动跳转游戏
  onMounted(() => {
    if (authStore.accessToken) {
      if (gameStore.player.privacyAgreed) {
        router.replace('/overview')
      }
      // 未同意隐私协议则留在首页，等用户从登录页回来后处理
    }
  })

  const goToLogin = () => {
    router.push('/login')
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
