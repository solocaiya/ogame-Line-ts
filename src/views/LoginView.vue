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

    <!-- 隐私协议同意弹窗（登录成功后） -->
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
          <AlertDialogCancel @click="privacyAgreed = false; showAgreementDialog = false">{{ t('common.cancel') }}</AlertDialogCancel>
          <AlertDialogAction :disabled="!privacyAgreed" @click="confirmPrivacy">
            {{ t('home.agreeAndStart') }}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

    <!-- 隐私协议全文弹窗 -->
    <PrivacyDialog v-model:open="showPrivacyDialog" />
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
import { Rocket, AlertCircle, Loader2 } from 'lucide-vue-next'
import PrivacyDialog from '@/components/dialogs/PrivacyDialog.vue'

const { t } = useI18n()
const router = useRouter()
const authStore = useAuthStore()
const gameStore = useGameStore()

const mode = ref<'login' | 'register'>('login')
const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const showAgreementDialog = ref(false)
const showPrivacyDialog = ref(false)
const privacyAgreed = ref(false)

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
    // 登录成功，清除游客标记
    localStorage.removeItem('guest_mode')
    // 检查隐私协议
    if (gameStore.player.privacyAgreed) {
      router.push('/overview')
    } else {
      showAgreementDialog.value = true
    }
  } catch {
    // 错误已在 authStore.error 中
  }
}

function confirmPrivacy() {
  if (privacyAgreed.value) {
    gameStore.player.privacyAgreed = true
    showAgreementDialog.value = false
    router.push('/overview')
  }
}

function skipLogin() {
  // 游客模式
  localStorage.setItem('guest_mode', 'true')
  if (gameStore.player.privacyAgreed) {
    router.push('/overview')
  } else {
    showAgreementDialog.value = true
  }
}

// 监听错误变化以清除
watch(() => authStore.error, () => {
  if (authStore.error) {
    setTimeout(() => authStore.clearError(), 5000)
  }
})
</script>
