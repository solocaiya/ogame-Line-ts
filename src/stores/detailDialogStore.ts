import { defineStore } from 'pinia'
import type { BuildingType, TechnologyType, ShipType, DefenseType } from '@/types/game'

export type DetailDialogType = 'building' | 'technology' | 'ship' | 'defense'

export interface DetailDialogState {
  isOpen: boolean
  type: DetailDialogType | null
  itemType: BuildingType | TechnologyType | ShipType | DefenseType | null
  currentLevel?: number // 用于建筑和科技
}

export const useDetailDialogStore = defineStore('detailDialog', {
  state: (): DetailDialogState => ({
    isOpen: false,
    type: null,
    itemType: null,
    currentLevel: undefined
  }),
  actions: {
    openBuilding(buildingType: BuildingType, currentLevel: number) {
      this.isOpen = true
      this.type = 'building'
      this.itemType = buildingType
      this.currentLevel = currentLevel
    },
    openTechnology(technologyType: TechnologyType, currentLevel: number) {
      this.isOpen = true
      this.type = 'technology'
      this.itemType = technologyType
      this.currentLevel = currentLevel
    },
    openShip(shipType: ShipType) {
      this.isOpen = true
      this.type = 'ship'
      this.itemType = shipType
      this.currentLevel = undefined
    },
    openDefense(defenseType: DefenseType) {
      this.isOpen = true
      this.type = 'defense'
      this.itemType = defenseType
      this.currentLevel = undefined
    },
    close() {
      this.isOpen = false
      this.type = null
      this.itemType = null
      this.currentLevel = undefined
    }
  }
})
