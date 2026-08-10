import { defineStore } from 'pinia'
import { apiService } from '@/services/apiService'
import { wsService } from '@/services/wsService'
import type { Alliance, AllianceMember, AllianceJoinRequest } from '@/types/game'

interface AllianceData {
  id: string
  name: string
  tag: string
  description: string
  leaderId: string
  members: AllianceMember[]
  pendingRequests: AllianceJoinRequest[]
  maxMembers: number
  autoAccept: boolean
  requireApproval: boolean
  createdAt: number
}

interface IncomingInvite {
  id: string
  allianceId: string
  allianceName: string
  inviterName: string
  createdAt: number
}

export const useAllianceStore = defineStore('alliance', {
  state: () => ({
    alliance: null as AllianceData | null,
    incomingInvites: [] as IncomingInvite[],
    loading: false,
    error: null as string | null,
    _unsubs: [] as (() => void)[],
  }),

  getters: {
    isInAlliance: (state) => !!state.alliance,
    isLeader: (state) => !!state.alliance && !!state.alliance.leaderId,
    isOfficerOrLeader: (state) => {
      if (!state.alliance) return false
      // Check if current user is leader or officer
      const members = state.alliance.members
      const leaderId = state.alliance.leaderId
      // We need the current player ID — check via member role
      return members.some(m =>
        (m.playerId === leaderId && m.role === 'leader') ||
        m.role === 'officer'
      )
    },
    memberCount: (state) => state.alliance?.members.length ?? 0,
    pendingRequestCount: (state) => state.alliance?.pendingRequests.length ?? 0,
    inviteCount: (state) => state.incomingInvites.length,
  },

  actions: {
    init(currentPlayerId: string) {
      this._currentPlayerId = currentPlayerId
      this.fetchMyAlliance()
      this._subscribeWS()
    },

    _currentPlayerId: '' as string,

    _subscribeWS() {
      // Clean up previous subscriptions
      this._unsubs.forEach(fn => fn())
      this._unsubs = []

      this._unsubs.push(
        wsService.on('alliance:join_request', (data) => {
          if (this.alliance && data) {
            // Add to pending requests if not already there
            const exists = this.alliance.pendingRequests.some(r => r.id === data.id)
            if (!exists) {
              this.alliance.pendingRequests.push(data)
            }
          }
        }),

        wsService.on('alliance:invite_received', (data) => {
          if (data) {
            const exists = this.incomingInvites.some(i => i.id === data.id || i.allianceId === data.allianceId)
            if (!exists) {
              this.incomingInvites.push(data)
            }
          }
        }),

        wsService.on('alliance:request_accepted', (data) => {
          if (data) {
            // Remove from invites
            this.incomingInvites = this.incomingInvites.filter(i => i.allianceId !== data.allianceId)
            // Refresh alliance data
            this.fetchMyAlliance()
          }
        }),

        wsService.on('alliance:member_joined', (data) => {
          if (this.alliance && data) {
            const exists = this.alliance.members.some(m => m.playerId === data.playerId)
            if (!exists) {
              this.alliance.members.push(data)
            }
            // Remove from pending requests
            this.alliance.pendingRequests = this.alliance.pendingRequests.filter(
              r => r.playerId !== data.playerId
            )
          }
        }),

        wsService.on('alliance:member_left', (data) => {
          if (this.alliance && data) {
            this.alliance.members = this.alliance.members.filter(m => m.playerId !== data.playerId)
            // If the player who left was the leader and it's us, clear alliance
            if (data.playerId === this._currentPlayerId) {
              this.alliance = null
            }
          }
        }),

        wsService.on('alliance:disbanded', () => {
          this.alliance = null
        }),

        wsService.on('alliance:settings_updated', (data) => {
          if (this.alliance && data) {
            if (data.description !== undefined) this.alliance.description = data.description
            if (data.maxMembers !== undefined) this.alliance.maxMembers = data.maxMembers
            if (data.autoAccept !== undefined) this.alliance.autoAccept = data.autoAccept
            if (data.requireApproval !== undefined) this.alliance.requireApproval = data.requireApproval
          }
        }),

        wsService.on('alliance:role_changed', (data) => {
          if (this.alliance && data) {
            const member = this.alliance.members.find(m => m.playerId === data.playerId)
            if (member) {
              member.role = data.newRole
            }
          }
        }),
      )
    },

    async fetchMyAlliance() {
      this.loading = true
      this.error = null
      try {
        const res = await apiService.getMyAlliance()
        this.alliance = res.alliance || null
      } catch (e: any) {
        // 404 or no alliance is fine
        if (e?.message?.includes('404') || e?.message?.includes('not in')) {
          this.alliance = null
        } else {
          this.error = e.message || 'Failed to load alliance'
        }
      } finally {
        this.loading = false
      }
    },

    async createAlliance(name: string, tag: string) {
      this.loading = true
      this.error = null
      try {
        const res = await apiService.createAlliance(name, tag)
        this.alliance = res.alliance || null
        return res
      } catch (e: any) {
        this.error = e.message || 'Failed to create alliance'
        throw e
      } finally {
        this.loading = false
      }
    },

    async leaveAlliance() {
      this.loading = true
      this.error = null
      try {
        await apiService.leaveAlliance()
        this.alliance = null
      } catch (e: any) {
        this.error = e.message || 'Failed to leave alliance'
        throw e
      } finally {
        this.loading = false
      }
    },

    async requestJoin(allianceId: string, message?: string) {
      this.error = null
      try {
        await apiService.requestJoinAlliance(allianceId, message)
      } catch (e: any) {
        this.error = e.message || 'Failed to request join'
        throw e
      }
    },

    async acceptRequest(requestId: string) {
      this.error = null
      try {
        await apiService.acceptAllianceRequest(requestId)
        // Remove from pending
        if (this.alliance) {
          this.alliance.pendingRequests = this.alliance.pendingRequests.filter(r => r.id !== requestId)
        }
      } catch (e: any) {
        this.error = e.message || 'Failed to accept request'
        throw e
      }
    },

    async rejectRequest(requestId: string) {
      this.error = null
      try {
        await apiService.rejectAllianceRequest(requestId)
        if (this.alliance) {
          this.alliance.pendingRequests = this.alliance.pendingRequests.filter(r => r.id !== requestId)
        }
      } catch (e: any) {
        this.error = e.message || 'Failed to reject request'
        throw e
      }
    },

    async acceptInvite(inviteId: string) {
      this.error = null
      try {
        await apiService.acceptAllianceInvite(inviteId)
        this.incomingInvites = this.incomingInvites.filter(i => i.id !== inviteId)
        await this.fetchMyAlliance()
      } catch (e: any) {
        this.error = e.message || 'Failed to accept invite'
        throw e
      }
    },

    async rejectInvite(inviteId: string) {
      this.error = null
      try {
        await apiService.rejectAllianceInvite(inviteId)
        this.incomingInvites = this.incomingInvites.filter(i => i.id !== inviteId)
      } catch (e: any) {
        this.error = e.message || 'Failed to reject invite'
        throw e
      }
    },

    async updateSettings(settings: { description?: string; maxMembers?: number; autoAccept?: boolean; requireApproval?: boolean }) {
      this.error = null
      try {
        await apiService.updateAllianceSettings(settings)
        // Update local state
        if (this.alliance) {
          Object.assign(this.alliance, settings)
        }
      } catch (e: any) {
        this.error = e.message || 'Failed to update settings'
        throw e
      }
    },

    async updateMemberRole(playerId: string, role: string) {
      this.error = null
      try {
        await apiService.updateAllianceMemberRole(playerId, role)
        if (this.alliance) {
          const member = this.alliance.members.find(m => m.playerId === playerId)
          if (member) member.role = role
        }
      } catch (e: any) {
        this.error = e.message || 'Failed to update role'
        throw e
      }
    },

    async removeMember(playerId: string) {
      this.error = null
      try {
        await apiService.removeMember(playerId)
        if (this.alliance) {
          this.alliance.members = this.alliance.members.filter(m => m.playerId !== playerId)
        }
      } catch (e: any) {
        this.error = e.message || 'Failed to remove member'
        throw e
      }
    },

    async searchAlliances(query: string) {
      try {
        const res = await apiService.searchAlliances(query)
        return res.alliances || []
      } catch (e: any) {
        this.error = e.message || 'Failed to search alliances'
        return []
      }
    },

    clearError() {
      this.error = null
    },

    destroy() {
      this._unsubs.forEach(fn => fn())
      this._unsubs = []
    },
  },
})
