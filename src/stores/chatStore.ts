import { defineStore } from 'pinia'
import { apiService } from '@/services/apiService'
import { wsService } from '@/services/wsService'
import { useAllianceStore } from './allianceStore'
import type { ChatMessage, ChatChannelType, ChatDNDMode } from '@/types/game'

const ChatChannelTypeValues = { WORLD: 'world', ALLIANCE: 'alliance' } as const
const ChatDNDModeValues = { NONE: 'none', MUTE_WORLD: 'mute_world', MUTE_ALLIANCE: 'mute_alliance', MUTE_ALL: 'mute_all' } as const

export const useChatStore = defineStore('chat', {
  state: () => ({
    worldMessages: [] as ChatMessage[],
    allianceMessages: [] as ChatMessage[],
    unreadWorld: 0,
    unreadAlliance: 0,
    isOpen: false,
    activeChannel: 'world' as ChatChannelType,
    dndMode: 'none' as ChatDNDMode,
    lastSenderName: '' as string,
    _unsub: null as (() => void) | null,
    _initialized: false,
    _loadingMore: false,
    _hasMoreWorld: true,
    _hasMoreAlliance: true,
  }),

  getters: {
    totalUnread: (state) => state.unreadWorld + state.unreadAlliance,
    currentMessages: (state) => {
      return state.activeChannel === 'world' ? state.worldMessages : state.allianceMessages
    },
    isWorldMuted: (state) => state.dndMode === 'mute_world' || state.dndMode === 'mute_all',
    isAllianceMuted: (state) => state.dndMode === 'mute_alliance' || state.dndMode === 'mute_all',
    canUseAllianceChannel(): boolean {
      const allianceStore = useAllianceStore()
      return allianceStore.isInAlliance
    },
    hasMore: (state) => {
      return state.activeChannel === 'world' ? state._hasMoreWorld : state._hasMoreAlliance
    },
  },

  actions: {
    init() {
      if (this._initialized) return
      this._initialized = true
      this._subscribeWS()
      // Load initial world chat history
      this.loadHistory('world')
    },

    _subscribeWS() {
      if (this._unsub) this._unsub()

      this._unsub = wsService.on('chat:message', (data: ChatMessage) => {
        if (!data) return

        // Check DND
        if (data.channel === 'world' && this.isWorldMuted) {
          // Still add to messages but don't show notification
          this._addMessage(data)
          return
        }
        if (data.channel === 'alliance' && this.isAllianceMuted) {
          this._addMessage(data)
          return
        }

        this._addMessage(data)

        // Update unread if panel is closed or different channel
        if (!this.isOpen || this.activeChannel !== data.channel) {
          if (data.channel === 'world') {
            this.unreadWorld++
          } else if (data.channel === 'alliance') {
            this.unreadAlliance++
          }
          // Update last sender for bubble notification
          this.lastSenderName = data.senderName
        }
      })
    },

    _addMessage(msg: ChatMessage) {
      if (msg.channel === 'world') {
        // Avoid duplicates
        if (!this.worldMessages.some(m => m.id === msg.id)) {
          this.worldMessages.push(msg)
        }
      } else if (msg.channel === 'alliance') {
        if (!this.allianceMessages.some(m => m.id === msg.id)) {
          this.allianceMessages.push(msg)
        }
      }
    },

    open() {
      this.isOpen = true
      // Mark current channel as read
      this.markAsRead(this.activeChannel)
      // Load alliance history if switching to alliance and not loaded
      if (this.activeChannel === 'alliance' && this.allianceMessages.length === 0) {
        this.loadHistory('alliance')
      }
    },

    close() {
      this.isOpen = false
    },

    toggle() {
      if (this.isOpen) {
        this.close()
      } else {
        this.open()
      }
    },

    setActiveChannel(channel: ChatChannelType) {
      this.activeChannel = channel
      this.markAsRead(channel)
      // Load history if empty
      const msgs = channel === 'world' ? this.worldMessages : this.allianceMessages
      if (msgs.length === 0) {
        this.loadHistory(channel)
      }
    },

    markAsRead(channel: ChatChannelType) {
      if (channel === 'world') {
        this.unreadWorld = 0
      } else if (channel === 'alliance') {
        this.unreadAlliance = 0
      }
    },

    async sendMessage(content: string) {
      if (!content.trim()) return
      try {
        const res = await apiService.sendChatMessage(this.activeChannel, content.trim())
        // The WS event will add it to the list, but also add locally for instant feedback
        // if it's from us (the WS broadcast may arrive slightly later)
        if (res && !this.currentMessages.some(m => m.id === res.id)) {
          this._addMessage(res)
        }
      } catch (e: any) {
        console.error('[ChatStore] Failed to send message:', e)
        throw e
      }
    },

    async loadHistory(channel?: ChatChannelType) {
      const ch = channel || this.activeChannel
      if (this._loadingMore) return

      const msgs = ch === 'world' ? this.worldMessages : this.allianceMessages
      const before = msgs.length > 0 ? msgs[0].timestamp : undefined

      this._loadingMore = true
      try {
        const res = await apiService.getChatHistory(ch, 50, before)
        const newMsgs: ChatMessage[] = res.messages || []

        if (newMsgs.length < 50) {
          if (ch === 'world') this._hasMoreWorld = false
          else this._hasMoreAlliance = false
        }

        // Prepend older messages
        if (ch === 'world') {
          this.worldMessages = [...newMsgs.reverse(), ...this.worldMessages]
        } else {
          this.allianceMessages = [...newMsgs.reverse(), ...this.allianceMessages]
        }
      } catch (e) {
        console.error('[ChatStore] Failed to load history:', e)
      } finally {
        this._loadingMore = false
      }
    },

    async loadMore() {
      await this.loadHistory(this.activeChannel)
    },

    async setDNDMode(mode: ChatDNDMode) {
      try {
        await apiService.updateChatDND(mode)
        this.dndMode = mode
      } catch (e: any) {
        console.error('[ChatStore] Failed to update DND:', e)
        throw e
      }
    },

    destroy() {
      if (this._unsub) {
        this._unsub()
        this._unsub = null
      }
      this._initialized = false
    },
  },
})
