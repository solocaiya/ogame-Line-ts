/**
 * 队列添加动画 - 贝塞尔曲线抛物线效果
 */

import { ref } from 'vue'

interface AnimatingItem {
  id: string
  x: number
  y: number
  targetX: number
  targetY: number
  progress: number
  type: 'building' | 'technology' | 'ship' | 'defense'
}

// 全局动画状态
const animatingItems = ref<AnimatingItem[]>([])

// 贝塞尔曲线计算
const cubicBezier = (t: number, p0: number, p1: number, p2: number, p3: number): number => {
  const t2 = t * t
  const t3 = t2 * t
  const mt = 1 - t
  const mt2 = mt * mt
  const mt3 = mt2 * mt
  return mt3 * p0 + 3 * mt2 * t * p1 + 3 * mt * t2 * p2 + t3 * p3
}

// 获取队列按钮的位置
const getQueueButtonPosition = (): { x: number; y: number } => {
  const queueButton = document.querySelector('[data-tutorial="queue-button"]')
  if (queueButton) {
    const rect = queueButton.getBoundingClientRect()
    return {
      x: rect.left + rect.width / 2,
      y: rect.top + rect.height / 2
    }
  }
  // 默认右上角位置
  return { x: window.innerWidth - 50, y: 50 }
}

// 创建动画元素
const createAnimationElement = (type: string): HTMLElement => {
  const el = document.createElement('div')
  el.className = 'queue-fly-animation'

  // 根据类型设置不同的颜色
  const colors: Record<string, string> = {
    building: '#22c55e', // green
    technology: '#3b82f6', // blue
    ship: '#a855f7', // purple
    defense: '#f59e0b' // amber
  }

  el.style.cssText = `
    position: fixed;
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: ${colors[type] || '#22c55e'};
    box-shadow: 0 0 10px ${colors[type] || '#22c55e'}, 0 0 20px ${colors[type] || '#22c55e'}80;
    pointer-events: none;
    z-index: 9999;
    transform: translate(-50%, -50%);
  `

  document.body.appendChild(el)
  return el
}

// 执行抛物线动画
export const triggerQueueAnimation = (event: MouseEvent, type: 'building' | 'technology' | 'ship' | 'defense' = 'building'): void => {
  const startX = event.clientX
  const startY = event.clientY
  const target = getQueueButtonPosition()

  const el = createAnimationElement(type)

  // 动画参数
  const duration = 500 // ms
  const startTime = performance.now()

  // 计算控制点（向上拱起的抛物线）
  const midX = (startX + target.x) / 2

  // 控制点：创建一个向上凸起的曲线
  const cp1x = startX + (midX - startX) * 0.5
  const cp1y = startY - 80
  const cp2x = target.x - (target.x - midX) * 0.5
  const cp2y = target.y - 80

  const animate = (currentTime: number) => {
    const elapsed = currentTime - startTime
    const progress = Math.min(elapsed / duration, 1)

    // 使用缓动函数让动画更自然
    const easeOutCubic = 1 - Math.pow(1 - progress, 3)

    // 计算当前位置（三次贝塞尔曲线）
    const x = cubicBezier(easeOutCubic, startX, cp1x, cp2x, target.x)
    const y = cubicBezier(easeOutCubic, startY, cp1y, cp2y, target.y)

    // 缩放效果：开始时放大，结束时缩小
    const scale = 1 + Math.sin(progress * Math.PI) * 0.5

    // 透明度：结束时淡出
    const opacity = progress < 0.8 ? 1 : 1 - (progress - 0.8) / 0.2

    el.style.left = `${x}px`
    el.style.top = `${y}px`
    el.style.transform = `translate3d(-50%, -50%, 0) scale3d(${scale}, ${scale}, 1)`
    el.style.opacity = `${opacity}`

    if (progress < 1) {
      requestAnimationFrame(animate)
    } else {
      // 动画结束，移除元素
      el.remove()

      // 触发队列按钮的脉冲效果
      const queueButton = document.querySelector('[data-tutorial="queue-button"]')
      if (queueButton) {
        queueButton.classList.add('queue-pulse')
        setTimeout(() => {
          queueButton.classList.remove('queue-pulse')
        }, 300)
      }
    }
  }

  requestAnimationFrame(animate)
}

// 批量动画（用于一次添加多个项目）
export const triggerMultipleQueueAnimations = (
  event: MouseEvent,
  count: number,
  type: 'building' | 'technology' | 'ship' | 'defense' = 'building'
): void => {
  const delay = 50 // 每个动画之间的延迟
  const maxAnimations = Math.min(count, 5) // 最多显示5个动画

  for (let i = 0; i < maxAnimations; i++) {
    setTimeout(() => {
      // 添加一些随机偏移让动画更自然
      const offsetEvent = {
        ...event,
        clientX: event.clientX + (Math.random() - 0.5) * 20,
        clientY: event.clientY + (Math.random() - 0.5) * 20
      } as MouseEvent

      triggerQueueAnimation(offsetEvent, type)
    }, i * delay)
  }
}

export const useQueueAnimation = () => {
  return {
    animatingItems,
    triggerQueueAnimation,
    triggerMultipleQueueAnimations
  }
}
