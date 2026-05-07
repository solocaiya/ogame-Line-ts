import { ref, onMounted, watch } from 'vue'
import { useGameStore } from '@/stores/gameStore'

type Theme = 'light' | 'dark'

const isDark = ref<boolean>(false)

export const useTheme = () => {
  const gameStore = useGameStore()

  // 初始化主题
  onMounted(() => {
    if (!gameStore.isDark) {
      // 首次访问，使用系统主题偏好
      isDark.value = window.matchMedia('(prefers-color-scheme: dark)').matches
      gameStore.isDark = isDark.value ? 'dark' : 'light'
    } else {
      // 从 gameStore 读取保存的主题（Pinia会自动从localStorage恢复）
      const savedTheme = gameStore.isDark as Theme
      isDark.value = savedTheme === 'dark'
    }
    applyTheme()
  })

  // 监听主题变化
  watch(isDark, () => {
    applyTheme()
    gameStore.isDark = isDark.value ? 'dark' : 'light'
  })

  // 应用主题
  const applyTheme = () => {
    if (isDark.value) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }

  // 切换主题
  const toggleTheme = () => {
    isDark.value = !isDark.value
  }

  return {
    isDark,
    toggleTheme
  }
}
