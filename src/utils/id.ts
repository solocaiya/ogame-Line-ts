/**
 * 统一生成带前缀的业务ID
 * 便于后续集中调整ID规则
 */
export const generateId = (prefix: string): string => {
  const timestamp = Date.now()
  return `${prefix}_${timestamp}_${Math.random().toString(36).slice(2, 9)}`
}
