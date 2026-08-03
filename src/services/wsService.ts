/**
 * WebSocket 服务 — 与服务端实时通信
 */

import { useAuthStore } from '../stores/authStore'

export interface WSMessage {
  type: string
  data?: any
  timestamp: number
}

type MessageHandler = (data: any) => void

class WebSocketService {
  private ws: WebSocket | null = null
  private handlers: Map<string, Set<MessageHandler>> = new Map()
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private reconnectDelay = 3000
  private maxReconnectDelay = 30000
  private currentDelay = this.reconnectDelay
  private intentionalClose = false
  private _connected = false

  get connected(): boolean {
    return this._connected
  }

  /**
   * Connect to the WebSocket server.
   */
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) return

    const authStore = useAuthStore()
    const token = authStore.accessToken
    if (!token) return

    this.intentionalClose = false

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const url = `${protocol}//${host}/api/ws`

    this.ws = new WebSocket(url)

    this.ws.onopen = () => {
      this._connected = true
      this.currentDelay = this.reconnectDelay
      console.log('[WS] Connected')
      this.emit('connected', {})
    }

    this.ws.onmessage = (event: MessageEvent) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        this.emit(msg.type, msg.data)
        this.emit('*', msg) // wildcard handler
      } catch (e) {
        console.warn('[WS] Failed to parse message:', event.data)
      }
    }

    this.ws.onclose = (event: CloseEvent) => {
      this._connected = false
      console.log('[WS] Disconnected:', event.code, event.reason)
      this.emit('disconnected', { code: event.code, reason: event.reason })

      if (!this.intentionalClose) {
        this.scheduleReconnect()
      }
    }

    this.ws.onerror = (error: Event) => {
      console.error('[WS] Error:', error)
    }
  }

  /**
   * Disconnect from the WebSocket server.
   */
  disconnect(): void {
    this.intentionalClose = true
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this._connected = false
  }

  /**
   * Subscribe to a message type. Use '*' to listen to all messages.
   */
  on(type: string, handler: MessageHandler): () => void {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set())
    }
    this.handlers.get(type)!.add(handler)

    // Return unsubscribe function
    return () => {
      this.handlers.get(type)?.delete(handler)
    }
  }

  /**
   * Subscribe to a message type, auto-unsubscribe after first call.
   */
  once(type: string, handler: MessageHandler): () => void {
    const wrapper: MessageHandler = (data) => {
      handler(data)
      unsub()
    }
    const unsub = this.on(type, wrapper)
    return unsub
  }

  private emit(type: string, data: any): void {
    // Type-specific handlers
    this.handlers.get(type)?.forEach(h => {
      try { h(data) } catch (e) { console.error('[WS] Handler error:', e) }
    })
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) return

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      this.currentDelay = Math.min(this.currentDelay * 1.5, this.maxReconnectDelay)
      console.log(`[WS] Reconnecting (delay: ${this.currentDelay}ms)...`)
      this.connect()
    }, this.currentDelay)
  }
}

// Singleton
export const wsService = new WebSocketService()
