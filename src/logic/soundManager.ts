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

// 生成简单的合成音效
const playTone = (
  frequency: number,
  duration: number,
  type: OscillatorType = 'sine',
  volume: number = 0.3
) => {
  const ctx = getAudioContext()
  if (!ctx) return

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
