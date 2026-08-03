// 新手引导配置

import type { TutorialStep } from '@/types/game'

export const TUTORIAL_STEPS: TutorialStep[] = [
  {
    id: 'welcome',
    title: 'tutorial.steps.welcome.title',
    content: 'tutorial.steps.welcome.content',
    placement: 'center',
    action: 'none',
    canSkip: true
  },
  {
    id: 'overview',
    title: 'tutorial.steps.overview.title',
    content: 'tutorial.steps.overview.content',
    route: '/overview',
    placement: 'right',
    action: 'none',
    canSkip: true
  },
  {
    id: 'build_solar',
    title: 'tutorial.steps.buildSolar.title',
    content: 'tutorial.steps.buildSolar.content',
    route: '/buildings',
    target: 'solarPlant',
    placement: 'bottom',
    action: 'build',
    actionTarget: 'solarPlant',
    canSkip: true
  },
  {
    id: 'build_metal',
    title: 'tutorial.steps.buildMetal.title',
    content: 'tutorial.steps.buildMetal.content',
    route: '/buildings',
    target: 'metalMine',
    placement: 'bottom',
    action: 'build',
    actionTarget: 'metalMine',
    canSkip: true
  },
  {
    id: 'build_crystal',
    title: 'tutorial.steps.buildCrystal.title',
    content: 'tutorial.steps.buildCrystal.content',
    route: '/buildings',
    target: 'crystalMine',
    placement: 'bottom',
    action: 'build',
    actionTarget: 'crystalMine',
    canSkip: true
  },
  {
    id: 'research_lab',
    title: 'tutorial.steps.researchLab.title',
    content: 'tutorial.steps.researchLab.content',
    route: '/buildings',
    target: 'researchLab',
    placement: 'bottom',
    action: 'build',
    actionTarget: 'researchLab',
    canSkip: true
  },
  {
    id: 'research_tech',
    title: 'tutorial.steps.researchTech.title',
    content: 'tutorial.steps.researchTech.content',
    route: '/research',
    placement: 'right',
    action: 'none',
    canSkip: true
  },
  {
    id: 'shipyard',
    title: 'tutorial.steps.shipyard.title',
    content: 'tutorial.steps.shipyard.content',
    route: '/shipyard',
    placement: 'right',
    action: 'none',
    canSkip: true
  },
  {
    id: 'galaxy',
    title: 'tutorial.steps.galaxy.title',
    content: 'tutorial.steps.galaxy.content',
    route: '/galaxy',
    placement: 'right',
    action: 'none',
    canSkip: true
  },
  {
    id: 'checkin',
    title: 'tutorial.steps.checkin.title',
    content: 'tutorial.steps.checkin.content',
    route: '/checkin',
    placement: 'right',
    action: 'none',
    canSkip: true
  },
  {
    id: 'complete',
    title: 'tutorial.steps.complete.title',
    content: 'tutorial.steps.complete.content',
    placement: 'center',
    action: 'none',
    canSkip: false
  }
]
