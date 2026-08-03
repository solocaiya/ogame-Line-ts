// 备注书签系统逻辑

import type { Player } from '@/types/game'

export interface Bookmark {
  id: string
  timestamp: number
  // 坐标信息
  galaxy: number
  system: number
  position: number
  // 备注
  name: string // 自定义名称
  note: string // 备注内容
  // 分类
  category: 'planet' | 'moon' | 'ally' | 'enemy' | 'resource' | 'other'
  // 颜色标记
  color?: string // 十六进制颜色
  // 是否收藏
  starred: boolean
}

export const BOOKMARK_CATEGORIES = [
  'planet',
  'moon',
  'ally',
  'enemy',
  'resource',
  'other'
] as const

export type BookmarkCategory = (typeof BOOKMARK_CATEGORIES)[number]

export const BOOKMARK_COLORS = [
  '#ef4444', // red
  '#f97316', // orange
  '#eab308', // yellow
  '#22c55e', // green
  '#3b82f6', // blue
  '#8b5cf6', // purple
  '#ec4899', // pink
  '#6b7280' // gray
] as const

/**
 * 添加书签
 */
export const addBookmark = (
  player: Player,
  data: Omit<Bookmark, 'id' | 'timestamp' | 'starred'>
): Bookmark => {
  if (!player.bookmarks) {
    player.bookmarks = []
  }

  const bookmark: Bookmark = {
    ...data,
    id: `bm_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`,
    timestamp: Date.now(),
    starred: false
  }

  player.bookmarks.unshift(bookmark)
  return bookmark
}

/**
 * 删除书签
 */
export const removeBookmark = (player: Player, bookmarkId: string): boolean => {
  if (!player.bookmarks) return false
  const index = player.bookmarks.findIndex(b => b.id === bookmarkId)
  if (index === -1) return false
  player.bookmarks.splice(index, 1)
  return true
}

/**
 * 更新书签
 */
export const updateBookmark = (
  player: Player,
  bookmarkId: string,
  updates: Partial<Omit<Bookmark, 'id' | 'timestamp'>>
): boolean => {
  if (!player.bookmarks) return false
  const bookmark = player.bookmarks.find(b => b.id === bookmarkId)
  if (!bookmark) return false
  Object.assign(bookmark, updates)
  return true
}

/**
 * 切换收藏状态
 */
export const toggleBookmarkStar = (player: Player, bookmarkId: string): boolean => {
  if (!player.bookmarks) return false
  const bookmark = player.bookmarks.find(b => b.id === bookmarkId)
  if (!bookmark) return false
  bookmark.starred = !bookmark.starred
  return true
}

/**
 * 格式化坐标字符串
 */
export const formatCoordinate = (galaxy: number, system: number, position: number): string => {
  return `[${galaxy}:${system}:${position}]`
}
