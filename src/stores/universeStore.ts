import { defineStore } from 'pinia'
import type { Planet, DebrisField } from '@/types/game'
import pkg from '../../package.json'
import { encryptData, decryptData } from '@/utils/crypto'

/**
 * 宇宙地图 Store
 * 存储宇宙中的所有星球和残骸场
 * 使用普通 localStorage 存储，不加密（地图数据是静态/共享数据）
 */
export const useUniverseStore = defineStore('universe', {
  state: () => ({
    // 宇宙星球地图：key 格式为 "galaxy:system:position"
    planets: {} as Record<string, Planet>,
    // 残骸场：key 格式为 "galaxy:system:position"
    debrisFields: {} as Record<string, DebrisField>
  }),
  persist: {
    key: `${pkg.name}-universe`,
    storage: localStorage,
    serializer: {
      serialize: state => encryptData(state),
      deserialize: value => decryptData(value)
    }
  }
})
