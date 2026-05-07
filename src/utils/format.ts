/**
 * 格式化数字为英文单位（K, M, B, T, Q）
 * @param num 数字
 * @param decimals 小数位数，默认2
 * @returns 格式化后的字符串
 */
export const formatNumber = (num: number, decimals: number = 2): string => {
  if (num >= 1_000_000_000_000_000) {
    return (num / 1_000_000_000_000_000).toFixed(decimals) + 'Q'
  } else if (num >= 1_000_000_000_000) {
    return (num / 1_000_000_000_000).toFixed(decimals) + 'T'
  } else if (num >= 1_000_000_000) {
    return (num / 1_000_000_000).toFixed(decimals) + 'B'
  } else if (num >= 1_000_000) {
    return (num / 1_000_000).toFixed(decimals) + 'M'
  } else if (num >= 1_000) {
    return (num / 1_000).toFixed(decimals) + 'K'
  }
  return Math.floor(num).toString()
}

/**
 * 获取资源颜色类（根据占用率）
 * @param current 当前数量
 * @param max 最大容量
 * @returns Tailwind CSS 类名
 */
export const getResourceColor = (current: number, max: number): string => {
  const ratio = current / max
  if (ratio >= 1) return 'text-red-600 dark:text-red-400'
  if (ratio >= 0.7) return 'text-yellow-600 dark:text-yellow-400'
  return ''
}
/**
 * 格式化相对时间（用于显示"多久之前"）
 * @param seconds 秒数
 * @param t 翻译函数
 * @returns 格式化后的相对时间字符串
 */
export const formatRelativeTime = (seconds: number, t: (key: string) => string): string => {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = Math.floor(seconds % 60)

  if (days > 0) {
    return `${days}${t('time.days')}${hours}${t('time.hours')}`
  }
  if (hours > 0) {
    return `${hours}${t('time.hours')}${minutes}${t('time.minutes')}`
  }
  if (minutes > 0) {
    return `${minutes}${t('time.minutes')}`
  }
  return `${secs}${t('time.seconds')}`
}

/**
 * 格式化时间（秒转为 年:天:时:分:秒）
 * @param seconds 秒数
 * @returns 格式化后的时间字符串
 * 例如：
 * 1:02:03:04:05
 * 02:03:04:05
 * 03:04:05
 */
export const formatTime = (seconds: number): string => {
  const YEAR = 365 * 86400
  const years = Math.floor(seconds / YEAR)
  seconds %= YEAR
  const days = Math.floor(seconds / 86400)
  seconds %= 86400
  const hours = Math.floor(seconds / 3600)
  seconds %= 3600
  const minutes = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  const h = hours.toString().padStart(2, '0')
  const m = minutes.toString().padStart(2, '0')
  const s = secs.toString().padStart(2, '0')
  if (years > 0) {
    return `${years}:${days}:${h}:${m}:${s}`
  }
  if (days > 0) {
    return `${days}:${h}:${m}:${s}`
  }
  return `${h}:${m}:${s}`
}

/**
 * 格式化日期
 * @param timestamp 时间戳
 * @returns 格式化后的日期字符串
 */
export const formatDate = (timestamp: number): string => {
  const date = new Date(timestamp)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

/**
 * 检查资源是否充足并返回对应的颜色类
 * @param available 当前可用资源数量
 * @param required 所需资源数量
 * @returns Tailwind CSS 类名（资源不足时返回红色）
 */
export const getResourceCostColor = (available: number, required: number): string => {
  if (available < required) {
    return 'text-red-600 dark:text-red-400'
  }
  return ''
}
