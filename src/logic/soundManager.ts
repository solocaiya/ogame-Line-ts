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

  // 如果 BGM 已启用但尚未播放，现在尝试播放
  if (bgmEnabled && !bgmPlaying) {
    startBgmOscillators()
  }
}

/**
 * 初始化音频系统 — 在 App 挂载时调用
 * 注册全局用户交互监听器以解锁 AudioContext
 */
export const initAudio = () => {
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

// ============ 背景音乐（BGM）— Web Audio API 合成环境音乐 ============

let bgmEnabled = false
let bgmPlaying = false
let bgmVolume = 0.3
let bgmGainNode: GainNode | null = null
let bgmOscillators: OscillatorNode[] = []

/**
 * 启动 BGM 振荡器 — 生成柔和的太空环境音乐
 * 使用多个低频正弦波叠加产生氛围感
 */
const startBgmOscillators = () => {
  const ctx = getAudioContext()
  if (!ctx || bgmPlaying) return
  if (ctx.state === 'suspended') {
    ctx.resume().catch(() => {})
  }

  // 主增益节点
  bgmGainNode = ctx.createGain()
  bgmGainNode.gain.setValueAtTime(bgmVolume * 0.15, ctx.currentTime) // 降低基础音量
  bgmGainNode.connect(ctx.destination)

  // 环境和弦：C3 + E3 + G3 + B3 叠加，产生柔和的太空感
  const frequencies = [130.81, 164.81, 196.00, 246.94] // C3, E3, G3, B3
  const types: OscillatorType[] = ['sine', 'sine', 'sine', 'triangle']

  frequencies.forEach((freq, i) => {
    const osc = ctx.createOscillator()
    const oscGain = ctx.createGain()

    osc.type = types[i]
    osc.frequency.setValueAtTime(freq, ctx.currentTime)

    // 每个振荡器音量递减
    oscGain.gain.setValueAtTime(0.3 - i * 0.05, ctx.currentTime)

    // 添加缓慢的 LFO 调制，产生呼吸感
    const lfo = ctx.createOscillator()
    const lfoGain = ctx.createGain()
    lfo.type = 'sine'
    lfo.frequency.setValueAtTime(0.1 + i * 0.05, ctx.currentTime) // 非常缓慢
    lfoGain.gain.setValueAtTime(0.1, ctx.currentTime)
    lfo.connect(lfoGain)
    lfoGain.connect(oscGain.gain)
    lfo.start(ctx.currentTime)

    osc.connect(oscGain)
    oscGain.connect(bgmGainNode!)
    osc.start(ctx.currentTime)

    bgmOscillators.push(osc)
  })

  bgmPlaying = true
}

/**
 * 停止 BGM 振荡器
 */
const stopBgmOscillators = () => {
  bgmOscillators.forEach(osc => {
    try { osc.stop() } catch { /* already stopped */ }
  })
  bgmOscillators = []
  if (bgmGainNode) {
    bgmGainNode.disconnect()
    bgmGainNode = null
  }
  bgmPlaying = false
}

/**
 * 播放背景音乐
 */
export const playBgm = () => {
  bgmEnabled = true
  const ctx = getAudioContext()
  if (!ctx) return

  // 如果音频已解锁，直接启动
  if (ctx.state === 'running') {
    startBgmOscillators()
  }
  // 否则等 unlockAudio() 调用时再启动
}

/**
 * 暂停背景音乐
 */
export const pauseBgm = () => {
  bgmEnabled = false
  stopBgmOscillators()
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
  if (bgmGainNode) {
    const ctx = getAudioContext()
    if (ctx) {
      bgmGainNode.gain.setValueAtTime(bgmVolume * 0.15, ctx.currentTime)
    }
  }
}

/**
 * 获取背景音乐状态
 */
export const getBgmState = () => ({
  playing: bgmPlaying,
  volume: bgmVolume,
  enabled: bgmEnabled
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
