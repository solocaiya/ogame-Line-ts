import { defineStore } from 'pinia'
import type { NPC } from '@/types/game'
import pkg from '../../package.json'
import { encryptData, decryptData } from '@/utils/crypto'

/**
 * NPC Store
 * 存储和管理所有NPC数据
 */
export const useNPCStore = defineStore('npc', {
  state: () => ({
    npcs: [] as NPC[],
    lastGrowthCheck: {} as Record<string, number> // npcId -> timestamp
  }),
  persist: {
    key: `${pkg.name}-npcs`,
    storage: localStorage,
    serializer: {
      serialize: state => encryptData(state),
      deserialize: value => decryptData(value)
    }
  }
})
