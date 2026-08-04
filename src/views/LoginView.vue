<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-b from-slate-950 via-slate-900 to-slate-950 p-4">
    <div class="absolute inset-0 overflow-hidden pointer-events-none">
      <BgStars />
    </div>

    <Card class="w-full max-w-md relative z-10">
      <CardHeader class="text-center space-y-2">
        <div class="mx-auto w-16 h-16 rounded-full bg-primary/20 flex items-center justify-center mb-2">
          <Rocket class="h-8 w-8 text-primary" />
        </div>
        <CardTitle class="text-2xl font-bold">{{ t('login.title') }}</CardTitle>
        <CardDescription>{{ t('login.subtitle') }}</CardDescription>
      </CardHeader>

      <CardContent>
        <!-- Tab 切换 -->
        <div class="flex rounded-lg bg-muted p-1 mb-6">
          <button
            class="flex-1 py-2 text-sm font-medium rounded-md transition-colors"
            :class="mode === 'login' ? 'bg-background text-foreground shadow' : 'text-muted-foreground hover:text-foreground'"
            @click="switchMode('login')"
          >
            {{ t('login.login') }}
          </button>
          <button
            class="flex-1 py-2 text-sm font-medium rounded-md transition-colors"
            :class="mode === 'register' ? 'bg-background text-foreground shadow' : 'text-muted-foreground hover:text-foreground'"
            @click="switchMode('register')"
          >
            {{ t('login.register') }}
          </button>
        </div>

        <!-- 错误提示 -->
        <div v-if="authStore.error" class="mb-4 p-3 rounded-lg bg-destructive/10 border border-destructive/20 text-destructive text-sm flex items-center gap-2">
          <AlertCircle class="h-4 w-4 shrink-0" />
          {{ authStore.error }}
        </div>

        <form @submit.prevent="handleSubmit" class="space-y-4">
          <div class="space-y-2">
            <Label for="username">{{ t('login.username') }}</Label>
            <Input
              id="username"
              v-model="username"
              :placeholder="t('login.usernamePlaceholder')"
              :disabled="authStore.loading"
              autocomplete="username"
              required
            />
          </div>

          <div class="space-y-2">
            <Label for="password">{{ t('login.password') }}</Label>
            <Input
              id="password"
              v-model="password"
              type="password"
              :placeholder="t('login.passwordPlaceholder')"
              :disabled="authStore.loading"
              autocomplete="current-password"
              required
            />
          </div>

          <div v-if="mode === 'register'" class="space-y-2">
            <Label for="confirmPassword">{{ t('login.confirmPassword') }}</Label>
            <Input
              id="confirmPassword"
              v-model="confirmPassword"
              type="password"
              :placeholder="t('login.confirmPasswordPlaceholder')"
              :disabled="authStore.loading"
              autocomplete="new-password"
              required
            />
          </div>

          <Button type="submit" class="w-full" :disabled="authStore.loading || !isValid">
            <Loader2 v-if="authStore.loading" class="mr-2 h-4 w-4 animate-spin" />
            {{ mode === 'login' ? t('login.loginButton') : t('login.registerButton') }}
          </Button>
        </form>

        <!-- 游客模式 -->
        <div class="mt-6 text-center">
          <button
            class="text-sm text-muted-foreground hover:text-foreground underline underline-offset-4 transition-colors"
            @click="skipLogin"
          >
            {{ t('login.skipAsGuest') }}
          </button>
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import { useAuthStore } from '@/stores/authStore'
import { useGameStore } from '@/stores/gameStore'
import { StarsBackground as BgStars } from '@/components/ui/bg-stars'
import Card from '@/components/ui/card/Card.vue'
import CardHeader from '@/components/ui/card/CardHeader.vue'
import CardTitle from '@/components/ui/card/CardTitle.vue'
import CardDescription from '@/components/ui/card/CardDescription.vue'
import CardContent from '@/components/ui/card/CardContent.vue'
import Button from '@/components/ui/button/Button.vue'
import Input from '@/components/ui/input/Input.vue'
import Label from '@/components/ui/label/Label.vue'
import { Rocket, AlertCircle, Loader2 } from 'lucide-vue-next'

const { t } = useI18n()
const router = useRouter()
const authStore = useAuthStore()
const gameStore = useGameStore()

const mode = ref<'login' | 'register'>('login')
const username = ref('')
const password = ref('')
const confirmPassword = ref('')

const isValid = computed(() => {
  if (username.value.length < 3) return false
  if (password.value.length < 6) return false
  if (mode.value === 'register' && password.value !== confirmPassword.value) return false
  return true
})

function switchMode(m: 'login' | 'register') {
  mode.value = m
  authStore.clearError()
}

async function handleSubmit() {
  if (!isValid.value) return

  try {
    if (mode.value === 'login') {
      await authStore.login(username.value, password.value)
    } else {
      await authStore.register(username.value, password.value)
    }
    // 登录成功，清除游客标记并跳转到总览
    localStorage.removeItem('guest_mode')
    router.push('/overview')
  } catch {
    // 错误已在 authStore.error 中
  }
}

function skipLogin() {
  // 游客模式：标记为游客，直接进入游戏（不登录服务器）
  localStorage.setItem('guest_mode', 'true')
  if (!gameStore.player.privacyAgreed) {
    router.push('/')
  } else {
    router.push('/overview')
  }
}

// 监听错误变化以清除
watch(() => authStore.error, () => {
  if (authStore.error) {
    setTimeout(() => authStore.clearError(), 5000)
  }
})
</script>
