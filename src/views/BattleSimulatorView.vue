<template>
  <div class="container mx-auto p-4 sm:p-6 space-y-6">
    <h1 class="text-2xl sm:text-3xl font-bold">{{ t('simulatorView.title') }}</h1>

    <!-- 标签切换 -->
    <Tabs v-model="activeTab" class="w-full">
      <TabsList class="grid w-full grid-cols-2">
        <TabsTrigger value="attacker">
          <Sword class="h-4 w-4 mr-2" />
          {{ t('simulatorView.attacker') }}
        </TabsTrigger>
        <TabsTrigger value="defender">
          <Shield class="h-4 w-4 mr-2" />
          {{ t('simulatorView.defender') }}
        </TabsTrigger>
      </TabsList>

      <!-- 攻击方配置 -->
      <TabsContent value="attacker" class="mt-4">
        <Card>
          <CardHeader>
            <CardTitle>{{ t('simulatorView.attackerConfig') }}</CardTitle>
            <CardDescription>{{ t('simulatorView.attackerConfigDesc') }}</CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <!-- 舰队配置 -->
            <div>
              <h3 class="text-sm font-medium mb-3">{{ t('simulatorView.fleet') }}</h3>
              <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
                <div v-for="shipType in Object.values(ShipType)" :key="shipType" class="space-y-1">
                  <Label :for="`attacker-${shipType}`" class="text-xs">{{ SHIPS[shipType].name }}</Label>
                  <Input
                    :id="`attacker-${shipType}`"
                    :model-value="attackerFleet[shipType] ?? 0"
                    @update:model-value="val => (attackerFleet[shipType] = typeof val === 'number' ? val : 0)"
                    type="number"
                    min="0"
                    class="h-8"
                  />
                </div>
              </div>
            </div>
            <!-- 科技等级 -->
            <div>
              <h3 class="text-sm font-medium mb-3">{{ t('simulatorView.techLevels') }}</h3>
              <div class="grid grid-cols-3 gap-3">
                <div v-for="techType in techTypes" :key="techType" class="space-y-1">
                  <Label :for="`attacker-${techType}`" class="text-xs">{{ t(`simulatorView.${techType}`) }}</Label>
                  <Input :id="`attacker-${techType}`" v-model.number="attackerTech[techType]" type="number" min="0" class="h-8" />
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <!-- 防守方配置 -->
      <TabsContent value="defender" class="mt-4">
        <Card>
          <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
            <div class="space-y-1">
              <CardTitle>{{ t('simulatorView.defenderConfig') }}</CardTitle>
              <CardDescription>{{ t('simulatorView.defenderConfigDesc') }}</CardDescription>
            </div>
            <Button variant="outline" size="sm" @click="showSpyReportSelector = true" :disabled="!gameStore.player?.spyReports?.length">
              <FileDown class="h-4 w-4 mr-2" />
              {{ t('simulatorView.importFromSpyReport') }}
            </Button>
          </CardHeader>
          <CardContent class="space-y-4">
            <!-- 舰队配置 -->
            <div>
              <h3 class="text-sm font-medium mb-3">{{ t('simulatorView.fleet') }}</h3>
              <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
                <div v-for="shipType in Object.values(ShipType)" :key="shipType" class="space-y-1">
                  <Label :for="`defender-${shipType}`" class="text-xs">{{ SHIPS[shipType].name }}</Label>
                  <Input
                    :id="`defender-${shipType}`"
                    :model-value="defenderFleet[shipType] ?? 0"
                    @update:model-value="val => (defenderFleet[shipType] = typeof val === 'number' ? val : 0)"
                    type="number"
                    min="0"
                    class="h-8"
                  />
                </div>
              </div>
            </div>

            <!-- 防御设施 -->
            <div>
              <h3 class="text-sm font-medium mb-3">{{ t('simulatorView.defenseStructures') }}</h3>
              <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
                <div v-for="defenseType in Object.values(DefenseType)" :key="defenseType" class="space-y-1">
                  <Label :for="`defense-${defenseType}`" class="text-xs">{{ DEFENSES[defenseType].name }}</Label>
                  <Input
                    :id="`defense-${defenseType}`"
                    :model-value="defenderDefense[defenseType] ?? 0"
                    @update:model-value="val => (defenderDefense[defenseType] = typeof val === 'number' ? val : 0)"
                    type="number"
                    min="0"
                    class="h-8"
                  />
                </div>
              </div>
            </div>

            <!-- 科技等级 -->
            <div>
              <h3 class="text-sm font-medium mb-3">{{ t('simulatorView.techLevels') }}</h3>
              <div class="grid grid-cols-3 gap-3">
                <div v-for="techType in techTypes" :key="techType" class="space-y-1">
                  <Label :for="`defender-${techType}`" class="text-xs">{{ t(`simulatorView.${techType}`) }}</Label>
                  <Input :id="`defender-${techType}`" v-model.number="defenderTech[techType]" type="number" min="0" class="h-8" />
                </div>
              </div>
            </div>

            <!-- 防守方资源 -->
            <div>
              <h3 class="text-sm font-medium mb-3">{{ t('simulatorView.defenderResources') }}</h3>
              <div class="grid grid-cols-3 gap-3">
                <div v-for="resourceType in resourceTypes" :key="resourceType.key" class="space-y-1">
                  <Label :for="`defender-${resourceType.key}`" class="text-xs flex items-center gap-1">
                    <ResourceIcon :type="resourceType.key" size="sm" />
                    {{ t(`resources.${resourceType.key}`) }}
                  </Label>
                  <Input
                    :id="`defender-${resourceType.key}`"
                    v-model.number="defenderResources[resourceType.key]"
                    type="number"
                    min="0"
                    class="h-8"
                  />
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>

    <!-- 操作按钮 -->
    <div class="flex gap-2">
      <Button @click="runSimulation" class="flex-1" size="lg">
        <Zap class="h-4 w-4 mr-2" />
        {{ t('simulatorView.startSimulation') }}
      </Button>
      <Button @click="resetSimulation" variant="outline" size="lg">
        <RotateCcw class="h-4 w-4 mr-2" />
        {{ t('simulatorView.reset') }}
      </Button>
    </div>

    <!-- 战斗结果对话框 -->
    <BattleReportDialog v-model:open="showResultDialog" :report="simulationResult" />

    <!-- 侦查报告选择对话框 -->
    <Dialog v-model:open="showSpyReportSelector">
      <DialogContent class="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{{ t('simulatorView.selectSpyReport') }}</DialogTitle>
        </DialogHeader>
        <div class="space-y-2">
          <div v-if="!sortedSpyReports.length" class="text-center py-8 text-muted-foreground">
            {{ t('simulatorView.noSpyReports') }}
          </div>
          <div
            v-for="report in sortedSpyReports"
            :key="report.id"
            @click="importFromSpyReport(report)"
            class="p-3 border rounded-lg cursor-pointer hover:bg-accent transition-colors"
          >
            <div class="flex justify-between items-start">
              <div>
                <div class="font-medium">{{ report.targetPlanetName }}</div>
                <div class="text-sm text-muted-foreground">
                  [{{ report.targetPosition.galaxy }}:{{ report.targetPosition.system }}:{{ report.targetPosition.position }}]
                </div>
              </div>
              <div class="text-sm text-muted-foreground">
                {{ formatTime(report.timestamp) }}
              </div>
            </div>
            <div class="mt-2 flex gap-4 text-xs">
              <span class="flex items-center gap-1">
                <ResourceIcon type="metal" size="sm" />
                {{ formatNumber(report.resources.metal) }}
              </span>
              <span class="flex items-center gap-1">
                <ResourceIcon type="crystal" size="sm" />
                {{ formatNumber(report.resources.crystal) }}
              </span>
              <span class="flex items-center gap-1">
                <ResourceIcon type="deuterium" size="sm" />
                {{ formatNumber(report.resources.deuterium) }}
              </span>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
  import { ref, toRaw, computed } from 'vue'
  import { useI18n } from '@/composables/useI18n'
  import { useGameConfig } from '@/composables/useGameConfig'
  import { useGameStore } from '@/stores/gameStore'
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
  import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
  import { Button } from '@/components/ui/button'
  import { Input } from '@/components/ui/input'
  import { Label } from '@/components/ui/label'
  import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
  import { ShipType, DefenseType, TechnologyType } from '@/types/game'
  import type { Fleet, BattleResult, SpyReport } from '@/types/game'
  import { workerManager } from '@/workers/workerManager'
  import ResourceIcon from '@/components/common/ResourceIcon.vue'
  import BattleReportDialog from '@/components/dialogs/BattleReportDialog.vue'
  import { Sword, Shield, Zap, RotateCcw, FileDown } from 'lucide-vue-next'
  import * as planetLogic from '@/logic/planetLogic'

  const { t } = useI18n()
  const { SHIPS, DEFENSES } = useGameConfig()
  const gameStore = useGameStore()

  // 侦查报告选择对话框状态
  const showSpyReportSelector = ref(false)

  // 科技类型配置
  const techTypes = ['weapon', 'shield', 'armor'] as const

  // 资源类型配置（用于防守方资源输入）
  const resourceTypes = [{ key: 'metal' as const }, { key: 'crystal' as const }, { key: 'deuterium' as const }]

  // 动态初始化所有舰船类型为0
  const initializeFleet = (): Partial<Fleet> => {
    const fleet: Partial<Fleet> = {}
    Object.values(ShipType).forEach(shipType => {
      fleet[shipType] = 0
    })
    return fleet
  }

  // 攻击方配置
  const attackerFleet = ref<Partial<Fleet>>(initializeFleet())

  const activeTab = ref('attacker')

  const attackerTech = ref({
    weapon: 0,
    shield: 0,
    armor: 0
  })

  // 防守方配置
  const defenderFleet = ref<Partial<Fleet>>(initializeFleet())

  // 动态初始化所有防御类型为0
  const initializeDefense = (): Partial<Record<DefenseType, number>> => {
    const defense: Partial<Record<DefenseType, number>> = {}
    Object.values(DefenseType).forEach(defenseType => {
      defense[defenseType] = 0
    })
    return defense
  }

  const defenderDefense = ref<Partial<Record<DefenseType, number>>>(initializeDefense())

  const defenderTech = ref({
    weapon: 0,
    shield: 0,
    armor: 0
  })

  const defenderResources = ref({
    metal: 100000,
    crystal: 50000,
    deuterium: 25000,
    darkMatter: 100,
    energy: 0
  })

  // 模拟结果
  const simulationResult = ref<BattleResult | null>(null)
  const showResultDialog = ref<boolean>(false)

  // 运行模拟（使用 Web Worker 进行计算）
  const runSimulation = async () => {
    // 使用 toRaw 将 Vue 响应式对象转换为普通对象，以便传递给 Worker
    const attackerSide = {
      ships: toRaw(attackerFleet.value),
      weaponTech: attackerTech.value.weapon,
      shieldTech: attackerTech.value.shield,
      armorTech: attackerTech.value.armor
    }

    const defenderSide = {
      ships: toRaw(defenderFleet.value),
      defense: toRaw(defenderDefense.value),
      weaponTech: defenderTech.value.weapon,
      shieldTech: defenderTech.value.shield,
      armorTech: defenderTech.value.armor
    }

    // 使用 Worker 执行战斗模拟
    const result = await workerManager.simulateBattle({
      attacker: attackerSide,
      defender: defenderSide,
      maxRounds: gameStore.battleToFinish ? 100 : 6
    })

    // 计算掠夺和残骸场
    const plunder =
      result.winner === 'attacker'
        ? await workerManager.calculatePlunder({
            defenderResources: toRaw(defenderResources.value),
            attackerFleet: result.attackerRemaining
          })
        : { metal: 0, crystal: 0, deuterium: 0, darkMatter: 0, energy: 0 }
    const debrisField = await workerManager.calculateDebris({
      attackerLosses: result.attackerLosses,
      defenderLosses: result.defenderLosses
    })
    const moonChance = planetLogic.calculateMoonChance(debrisField) / 100 // 转换为 0-1 范围

    simulationResult.value = {
      id: `sim_${Date.now()}`,
      timestamp: Date.now(),
      attackerId: 'simulator_attacker',
      defenderId: 'simulator_defender',
      attackerPlanetId: 'sim_attacker',
      defenderPlanetId: 'sim_defender',
      attackerFleet: attackerFleet.value,
      defenderFleet: defenderFleet.value,
      defenderDefense: defenderDefense.value,
      attackerLosses: result.attackerLosses,
      defenderLosses: result.defenderLosses,
      winner: result.winner,
      plunder,
      debrisField,
      rounds: result.rounds,
      attackerRemaining: result.attackerRemaining,
      defenderRemaining: result.defenderRemaining,
      roundDetails: result.roundDetails,
      moonChance
    }

    // 显示结果对话框
    showResultDialog.value = true
  }

  // 重置模拟
  const resetSimulation = () => {
    attackerFleet.value = initializeFleet()
    defenderFleet.value = initializeFleet()
    defenderDefense.value = initializeDefense()
    attackerTech.value = { weapon: 0, shield: 0, armor: 0 }
    defenderTech.value = { weapon: 0, shield: 0, armor: 0 }
    simulationResult.value = null
    showResultDialog.value = false
  }

  // 按时间排序的侦查报告
  const sortedSpyReports = computed(() => {
    return [...(gameStore.player?.spyReports || [])].sort((a, b) => b.timestamp - a.timestamp)
  })

  // 格式化时间
  const formatTime = (timestamp: number) => {
    return new Date(timestamp).toLocaleString()
  }

  // 格式化数字
  const formatNumber = (num: number) => {
    if (num >= 1e9) return (num / 1e9).toFixed(1) + 'B'
    if (num >= 1e6) return (num / 1e6).toFixed(1) + 'M'
    if (num >= 1e3) return (num / 1e3).toFixed(1) + 'K'
    return num.toString()
  }

  // 从侦查报告导入数据
  const importFromSpyReport = (report: SpyReport) => {
    // 先重置防守方数据
    defenderFleet.value = initializeFleet()
    defenderDefense.value = initializeDefense()

    // 填入资源
    if (report.resources) {
      defenderResources.value = {
        metal: report.resources.metal || 0,
        crystal: report.resources.crystal || 0,
        deuterium: report.resources.deuterium || 0,
        darkMatter: report.resources.darkMatter || 0,
        energy: 0
      }
    }

    // 填入舰队
    if (report.fleet) {
      Object.entries(report.fleet).forEach(([key, value]) => {
        defenderFleet.value[key as keyof Fleet] = value || 0
      })
    }

    // 填入防御
    if (report.defense) {
      Object.entries(report.defense).forEach(([key, value]) => {
        defenderDefense.value[key as DefenseType] = value || 0
      })
    }

    // 填入科技
    if (report.technologies) {
      defenderTech.value.weapon = report.technologies[TechnologyType.WeaponsTechnology] || 0
      defenderTech.value.shield = report.technologies[TechnologyType.ShieldingTechnology] || 0
      defenderTech.value.armor = report.technologies[TechnologyType.ArmourTechnology] || 0
    }

    // 关闭对话框并切换到防守方标签
    showSpyReportSelector.value = false
    activeTab.value = 'defender'
  }
</script>
