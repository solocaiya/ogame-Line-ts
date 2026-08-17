// 音效管理器

// 音效类型
export const SoundType = {
  BuildingComplete: 'building_complete',
  ResearchComplete: 'research_complete',
  FleetDispatch: 'fleet_dispatch',
  FleetReturn: 'fleet_return',
  BattleVictory: 'battle_victory',
  BattleDefeat: 'battle_defeat',
  CheckIn: 'checkin',
  AchievementUnlock: 'achievement_unlock',
  TradeComplete: 'trade_complete',
  Notification: 'notification',
  Click: 'click',
  Error: 'error'
} as const

export type SoundType = (typeof SoundType)[keyof typeof SoundType]

// 音频上下文（懒加载）
let audioContext: AudioContext | null = null
let audioUnlocked = false

const getAudioContext = (): AudioContext | null => {
  if (typeof window === 'undefined') return null
  if (!audioContext) {
    try {
      audioContext = new (window.AudioContext || (window as any).webkitAudioContext)()
    } catch {
      return null
    }
  }
  return audioContext
}

/**
 * 解锁音频上下文 — 必须在用户手势事件中调用
 * 浏览器要求 AudioContext 在用户交互后才能播放
 */
export const unlockAudio = () => {
  const ctx = getAudioContext()
  if (!ctx) return
  if (ctx.state === 'suspended') {
    ctx.resume().catch(() => {})
  }
  audioUnlocked = true

  // 如果 BGM 已启用但尚未播放，现在尝试恢复当前场景 BGM
  if (bgmEnabled && !bgmPlaying) {
    resumeCurrentBgm()
  }
}

/**
 * 初始化音频系统 — 在 App 挂载时调用
 * 注册全局用户交互监听器以解锁 AudioContext
 * 注册路由监听器以自动切换 BGM 场景
 */
export const initAudio = (router?: { afterEach: (cb: (to: { name?: string | symbol }) => void) => void }) => {
  if (typeof window === 'undefined') return

  // 监听首次用户交互以解锁音频
  const unlock = () => {
    unlockAudio()
    // 解锁后移除监听器
    window.removeEventListener('click', unlock)
    window.removeEventListener('keydown', unlock)
    window.removeEventListener('touchstart', unlock)
  }
  window.addEventListener('click', unlock, { once: false })
  window.addEventListener('keydown', unlock, { once: false })
  window.addEventListener('touchstart', unlock, { once: false })

  // 注册路由监听 — 自动切换 BGM 场景
  if (router) {
    router.afterEach((to) => {
      if (to.name) {
        switchBgm(to.name)
      }
    })
  }
}

// 生成简单的合成音效
const playTone = (
  frequency: number,
  duration: number,
  type: OscillatorType = 'sine',
  volume: number = 0.3
) => {
  const ctx = getAudioContext()
  if (!ctx) return
  // 确保上下文已恢复
  if (ctx.state === 'suspended') {
    ctx.resume().catch(() => {})
  }

  const oscillator = ctx.createOscillator()
  const gainNode = ctx.createGain()

  oscillator.connect(gainNode)
  gainNode.connect(ctx.destination)

  oscillator.type = type
  oscillator.frequency.setValueAtTime(frequency, ctx.currentTime)

  gainNode.gain.setValueAtTime(volume, ctx.currentTime)
  gainNode.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + duration)

  oscillator.start(ctx.currentTime)
  oscillator.stop(ctx.currentTime + duration)
}

// 各音效的播放函数
const soundPlayers: Record<SoundType, (volume: number) => void> = {
  [SoundType.BuildingComplete]: (v) => {
    playTone(523, 0.15, 'sine', v) // C5
    setTimeout(() => playTone(659, 0.15, 'sine', v), 100) // E5
    setTimeout(() => playTone(784, 0.2, 'sine', v), 200) // G5
  },
  [SoundType.ResearchComplete]: (v) => {
    playTone(440, 0.1, 'triangle', v)
    setTimeout(() => playTone(554, 0.1, 'triangle', v), 80)
    setTimeout(() => playTone(659, 0.1, 'triangle', v), 160)
    setTimeout(() => playTone(880, 0.3, 'triangle', v), 240)
  },
  [SoundType.FleetDispatch]: (v) => {
    playTone(220, 0.2, 'sawtooth', v * 0.5)
    setTimeout(() => playTone(330, 0.3, 'sawtooth', v * 0.5), 150)
  },
  [SoundType.FleetReturn]: (v) => {
    playTone(440, 0.15, 'sine', v)
    setTimeout(() => playTone(330, 0.2, 'sine', v), 120)
  },
  [SoundType.BattleVictory]: (v) => {
    playTone(523, 0.1, 'square', v * 0.4)
    setTimeout(() => playTone(659, 0.1, 'square', v * 0.4), 100)
    setTimeout(() => playTone(784, 0.1, 'square', v * 0.4), 200)
    setTimeout(() => playTone(1047, 0.3, 'square', v * 0.4), 300)
  },
  [SoundType.BattleDefeat]: (v) => {
    playTone(440, 0.2, 'sawtooth', v * 0.4)
    setTimeout(() => playTone(370, 0.2, 'sawtooth', v * 0.4), 200)
    setTimeout(() => playTone(294, 0.4, 'sawtooth', v * 0.4), 400)
  },
  [SoundType.CheckIn]: (v) => {
    playTone(659, 0.1, 'sine', v)
    setTimeout(() => playTone(784, 0.1, 'sine', v), 80)
    setTimeout(() => playTone(1047, 0.2, 'sine', v), 160)
  },
  [SoundType.AchievementUnlock]: (v) => {
    playTone(523, 0.1, 'triangle', v)
    setTimeout(() => playTone(659, 0.1, 'triangle', v), 100)
    setTimeout(() => playTone(784, 0.1, 'triangle', v), 200)
    setTimeout(() => playTone(1047, 0.1, 'triangle', v), 300)
    setTimeout(() => playTone(1319, 0.3, 'triangle', v), 400)
  },
  [SoundType.TradeComplete]: (v) => {
    playTone(440, 0.1, 'sine', v)
    setTimeout(() => playTone(554, 0.15, 'sine', v), 100)
  },
  [SoundType.Notification]: (v) => {
    playTone(880, 0.1, 'sine', v * 0.6)
    setTimeout(() => playTone(880, 0.1, 'sine', v * 0.6), 150)
  },
  [SoundType.Click]: (v) => {
    playTone(1000, 0.05, 'sine', v * 0.3)
  },
  [SoundType.Error]: (v) => {
    playTone(200, 0.3, 'square', v * 0.4)
  }
}

// ============ 多场景背景音乐（BGM）— HTMLAudioElement 场景切换 ============

/** BGM 循环模式 */
export type BgmLoopMode = 'seamless' | 'gap'

/** BGM 场景配置 */
export interface BgmSceneConfig {
  /** 场景标识 */
  id: string
  /** 音频文件路径（相对于 public/） */
  src: string
  /** 循环模式：seamless = 无缝循环，gap = 播放后间隔静音再重播 */
  loopMode: BgmLoopMode
  /** gap 模式下的静音间隔秒数（仅 gap 模式有效） */
  gapSeconds?: number
}

/**
 * 7 个 BGM 场景配置
 * - 4 个无缝循环（seamless）：主菜单、星系图、舰队/战斗、商店/充值
 * - 3 个间隔循环（gap）：建筑/资源、科研、联盟
 */
export const BGM_SCENES: Record<string, BgmSceneConfig> = {
  main: {
    id: 'main',
    src: '/audio/bgm/main.mp3',
    loopMode: 'seamless'
  },
  galaxy: {
    id: 'galaxy',
    src: '/audio/bgm/galaxy.mp3',
    loopMode: 'seamless'
  },
  buildings: {
    id: 'buildings',
    src: '/audio/bgm/buildings.mp3',
    loopMode: 'gap',
    gapSeconds: 9 // 8-10s 间隔
  },
  research: {
    id: 'research',
    src: '/audio/bgm/research.mp3',
    loopMode: 'gap',
    gapSeconds: 12 // 10-15s 间隔
  },
  fleet: {
    id: 'fleet',
    src: '/audio/bgm/fleet.mp3',
    loopMode: 'seamless'
  },
  alliance: {
    id: 'alliance',
    src: '/audio/bgm/alliance.mp3',
    loopMode: 'gap',
    gapSeconds: 10 // 8-12s 间隔
  },
  shop: {
    id: 'shop',
    src: '/audio/bgm/shop.mp3',
    loopMode: 'seamless'
  }
}

/**
 * 路由名称 → BGM 场景映射
 * 未映射的路由不会触发 BGM 切换（保持当前场景）
 */
const ROUTE_TO_SCENE: Record<string, string> = {
  // 主菜单/登录
  login: 'main',
  home: 'main',
  // 星系图
  galaxy: 'galaxy',
  // 建筑/资源（含星球队列）
  buildings: 'buildings',
  overview: 'buildings',
  shipyard: 'buildings',
  defense: 'buildings',
  planet: 'buildings',
  'planet-queue': 'buildings',
  // 科研
  research: 'research',
  // 舰队/战斗
  fleet: 'fleet',
  'battle-simulator': 'fleet',
  'battle-reports': 'fleet',
  // 联盟/外交
  alliance: 'alliance',
  diplomacy: 'alliance',
  // 商店/充值
  trader: 'shop',
  recharge: 'shop',
  'growth-fund': 'shop'
}

// BGM 状态
let bgmEnabled = false
let bgmVolume = 0.3
let bgmCurrentScene: string | null = null
let bgmAudioElement: HTMLAudioElement | null = null
let bgmGapTimer: ReturnType<typeof setTimeout> | null = null
let bgmPlaying = false

/** 停止当前 BGM 播放（内部） */
const stopBgmPlayback = () => {
  if (bgmGapTimer) {
    clearTimeout(bgmGapTimer)
    bgmGapTimer = null
  }
  if (bgmAudioElement) {
    bgmAudioElement.pause()
    bgmAudioElement.removeAttribute('src')
    bgmAudioElement.load() // 释放资源
    bgmAudioElement = null
  }
  bgmPlaying = false
}

/** 启动音频元素播放（内部） */
const startAudioElement = (audio: HTMLAudioElement, scene: BgmSceneConfig) => {
  audio.volume = bgmVolume

  if (scene.loopMode === 'seamless') {
    // 无缝循环：HTMLAudioElement 原生 loop
    audio.loop = true
    audio.play().catch(() => {})
    bgmPlaying = true
  } else {
    // 间隔循环：播放一次，结束后静音 N 秒再重播
    audio.loop = false
    const scheduleGapReplay = () => {
      const gapMs = (scene.gapSeconds ?? 10) * 1000
      // 添加 ±20% 随机抖动，避免机械感
      const jitter = gapMs * 0.2 * (Math.random() * 2 - 1)
      bgmGapTimer = setTimeout(() => {
        if (bgmEnabled && bgmCurrentScene === scene.id) {
          audio.currentTime = 0
          audio.play().catch(() => {})
        }
      }, gapMs + jitter)
    }
    audio.addEventListener('ended', scheduleGapReplay)
    audio.play().catch(() => {})
    bgmPlaying = true
  }
}

/**
 * 根据路由名称切换 BGM 场景
 * 在 router.afterEach 中调用，实现自动场景切换
 */
export const switchBgm = (routeName: string | symbol | undefined) => {
  if (!bgmEnabled || !routeName || typeof routeName === 'symbol') return

  const sceneId = ROUTE_TO_SCENE[routeName]
  if (!sceneId || sceneId === bgmCurrentScene) return

  const scene = BGM_SCENES[sceneId]
  if (!scene) return

  // 停止当前 BGM
  stopBgmPlayback()

  // 创建新的音频元素
  bgmCurrentScene = sceneId
  bgmAudioElement = new Audio(scene.src)
  bgmAudioElement.preload = 'auto'

  // 音频加载失败时静默处理
  bgmAudioElement.addEventListener('error', () => {
    bgmPlaying = false
  })

  startAudioElement(bgmAudioElement, scene)
}

/** 恢复当前场景 BGM（音频解锁后调用） */
const resumeCurrentBgm = () => {
  if (!bgmCurrentScene || bgmPlaying) return
  const scene = BGM_SCENES[bgmCurrentScene]
  if (!scene) return

  if (!bgmAudioElement) {
    bgmAudioElement = new Audio(scene.src)
    bgmAudioElement.preload = 'auto'
    bgmAudioElement.addEventListener('error', () => {
      bgmPlaying = false
    })
  }
  startAudioElement(bgmAudioElement, scene)
}

/**
 * 播放背景音乐（从当前路由开始）
 */
export const playBgm = () => {
  bgmEnabled = true
  // 从当前路由推断场景
  let currentRoute = ''
  if (typeof window !== 'undefined') {
    const hash = window.location.hash.replace('#/', '').replace('#', '')
    currentRoute = hash || 'home'
  }
  switchBgm(currentRoute)
}

/**
 * 暂停背景音乐
 */
export const pauseBgm = () => {
  bgmEnabled = false
  stopBgmPlayback()
  bgmCurrentScene = null
}

/**
 * 切换背景音乐播放/暂停
 */
export const toggleBgm = () => {
  if (bgmEnabled) {
    pauseBgm()
  } else {
    playBgm()
  }
}

/**
 * 设置背景音乐音量 (0-1)
 */
export const setBgmVolume = (vol: number) => {
  bgmVolume = Math.max(0, Math.min(1, vol))
  if (bgmAudioElement) {
    bgmAudioElement.volume = bgmVolume
  }
}

/**
 * 获取背景音乐状态
 */
export const getBgmState = () => ({
  playing: bgmPlaying,
  volume: bgmVolume,
  enabled: bgmEnabled,
  currentScene: bgmCurrentScene
})

/**
 * 播放音效
 */
export const playSound = (
  type: SoundType,
  options?: { enabled?: boolean; volume?: number }
) => {
  const enabled = options?.enabled ?? true
  if (!enabled) return

  const volume = options?.volume ?? 0.7
  const player = soundPlayers[type]
  if (player) {
    try {
      player(volume)
    } catch {
      // 静默失败
    }
  }
}
