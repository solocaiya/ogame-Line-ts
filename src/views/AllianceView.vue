<template>
  <div class="container mx-auto p-4 max-w-4xl">
    <h1 class="text-2xl font-bold mb-6">{{ t('alliance.title') }}</h1>

    <!-- Error display -->
    <div v-if="allianceStore.error" class="mb-4 p-3 bg-destructive/10 text-destructive rounded-md text-sm">
      {{ translatedError }}
    </div>

    <!-- Loading state -->
    <div v-if="allianceStore.loading && !allianceStore.alliance" class="flex items-center justify-center py-12">
      <Loader2 class="h-8 w-8 animate-spin text-muted-foreground" />
    </div>

    <!-- Not in alliance — show search + create -->
    <template v-else-if="!allianceStore.alliance">
      <!-- Incoming invites -->
      <Card v-if="allianceStore.incomingInvites.length > 0" class="mb-6">
        <CardHeader>
          <CardTitle class="text-base">{{ t('alliance.incomingInvites') }}</CardTitle>
        </CardHeader>
        <CardContent class="space-y-2">
          <div
            v-for="invite in allianceStore.incomingInvites"
            :key="invite.id"
            class="flex items-center justify-between p-3 border rounded-md"
          >
            <div>
              <p class="font-medium text-sm">{{ invite.allianceName }}</p>
              <p class="text-xs text-muted-foreground">{{ invite.inviterName }}</p>
            </div>
            <div class="flex gap-2">
              <Button size="sm" class="h-7 text-xs" @click="acceptInvite(invite.id)">
                {{ t('alliance.accept') }}
              </Button>
              <Button size="sm" variant="outline" class="h-7 text-xs" @click="rejectInvite(invite.id)">
                {{ t('alliance.reject') }}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- Create alliance -->
      <Card class="mb-6">
        <CardHeader>
          <CardTitle class="text-base">{{ t('alliance.create') }}</CardTitle>
        </CardHeader>
        <CardContent>
          <div class="grid gap-3">
            <div>
              <Label class="text-xs">{{ t('alliance.name') }}</Label>
              <Input v-model="createName" :placeholder="t('alliance.namePlaceholder')" class="mt-1" maxlength="32" />
            </div>
            <div>
              <Label class="text-xs">{{ t('alliance.tag') }}</Label>
              <Input v-model="createTag" :placeholder="t('alliance.tagPlaceholder')" class="mt-1" maxlength="8" />
              <p class="text-xs text-muted-foreground mt-1">{{ t('alliance.tagHint') }}</p>
            </div>
            <Button
              :disabled="!canCreate || allianceStore.loading"
              @click="handleCreate"
            >
              <Loader2 v-if="allianceStore.loading" class="h-4 w-4 mr-2 animate-spin" />
              {{ t('alliance.create') }}
            </Button>
          </div>
        </CardContent>
      </Card>

      <!-- Search alliances -->
      <Card>
        <CardHeader>
          <CardTitle class="text-base">{{ t('alliance.search') }}</CardTitle>
        </CardHeader>
        <CardContent>
          <div class="flex gap-2 mb-4">
            <Input
              v-model="searchQuery"
              :placeholder="t('alliance.searchPlaceholder')"
              class="flex-1"
              @keydown.enter="handleSearch"
            />
            <Button variant="outline" :disabled="!searchQuery.trim()" @click="handleSearch">
              <Search class="h-4 w-4" />
            </Button>
          </div>

          <!-- Search results -->
          <div v-if="searchResults.length > 0" class="space-y-2">
            <div
              v-for="a in searchResults"
              :key="a.id"
              class="flex items-center justify-between p-3 border rounded-md"
            >
              <div>
                <p class="font-medium text-sm">{{ a.name }} <span class="text-muted-foreground">[{{ a.tag }}]</span></p>
                <p class="text-xs text-muted-foreground">{{ a.members?.length || 0 }} / {{ a.maxMembers || 30 }} {{ t('alliance.members') }}</p>
              </div>
              <Button size="sm" variant="outline" class="h-7 text-xs" @click="requestJoin(a.id)">
                {{ t('alliance.requestToJoin') }}
              </Button>
            </div>
          </div>
          <div v-else-if="searched && !searching" class="text-center py-4 text-muted-foreground text-sm">
            {{ t('alliance.noResults') }}
          </div>
        </CardContent>
      </Card>
    </template>

    <!-- In alliance — show management -->
    <template v-else>
      <!-- Alliance header -->
      <Card class="mb-6">
        <CardContent class="pt-6">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-xl font-bold">{{ allianceStore.alliance.name }}</h2>
              <p class="text-sm text-muted-foreground">
                [{{ allianceStore.alliance.tag }}] · {{ allianceStore.memberCount }} / {{ allianceStore.alliance.maxMembers }} {{ t('alliance.members') }}
              </p>
              <p v-if="allianceStore.alliance.description" class="text-sm mt-2">{{ allianceStore.alliance.description }}</p>
            </div>
            <Button variant="destructive" size="sm" @click="showLeaveConfirm = true">
              {{ t('alliance.leave') }}
            </Button>
          </div>
        </CardContent>
      </Card>

      <!-- Tabs: Members | Requests | Settings -->
      <Tabs v-model="activeTab">
        <TabsList class="grid w-full grid-cols-3">
          <TabsTrigger value="members">{{ t('alliance.members') }}</TabsTrigger>
          <TabsTrigger value="requests" :disabled="!allianceStore.isOfficerOrLeader">
            {{ t('alliance.requests') }}
            <Badge v-if="allianceStore.pendingRequestCount > 0" variant="destructive" class="ml-1 h-4 min-w-4 px-1 text-[9px]">
              {{ allianceStore.pendingRequestCount }}
            </Badge>
          </TabsTrigger>
          <TabsTrigger value="settings" :disabled="!allianceStore.isOfficerOrLeader">
            {{ t('alliance.settings') }}
          </TabsTrigger>
        </TabsList>

        <!-- Members tab -->
        <TabsContent value="members" class="mt-4">
          <Card>
            <CardContent class="pt-6">
              <div class="space-y-2">
                <div
                  v-for="member in sortedMembers"
                  :key="member.playerId"
                  class="flex items-center justify-between p-3 border rounded-md"
                >
                  <div class="flex items-center gap-3">
                    <Badge :variant="roleVariant(member.role)" class="text-[10px]">
                      {{ t(`alliance.role.${member.role}`) }}
                    </Badge>
                    <span class="text-muted-foreground text-xs">[{{ allianceStore.alliance.tag }}]</span>
                    <span class="font-medium text-sm">{{ member.username }}</span>
                  </div>
                  <!-- Actions for officers/leader (not on self) -->
                  <div v-if="allianceStore.isOfficerOrLeader && member.playerId !== currentPlayerId" class="flex gap-1">
                    <template v-if="allianceStore.isLeader && member.role === 'member'">
                      <Button size="sm" variant="ghost" class="h-7 text-xs" @click="promote(member.playerId)">
                        {{ t('alliance.promote') }}
                      </Button>
                    </template>
                    <template v-if="allianceStore.isLeader && member.role === 'officer'">
                      <Button size="sm" variant="ghost" class="h-7 text-xs" @click="demote(member.playerId)">
                        {{ t('alliance.demote') }}
                      </Button>
                    </template>
                    <Button size="sm" variant="ghost" class="h-7 text-xs text-destructive" @click="confirmKick(member)">
                      {{ t('alliance.kick') }}
                    </Button>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <!-- Requests tab -->
        <TabsContent value="requests" class="mt-4">
          <Card>
            <CardContent class="pt-6">
              <div v-if="allianceStore.alliance.pendingRequests.length === 0" class="text-center py-8 text-muted-foreground text-sm">
                {{ t('alliance.noRequests') }}
              </div>
              <div v-else class="space-y-2">
                <div
                  v-for="req in allianceStore.alliance.pendingRequests"
                  :key="req.id"
                  class="flex items-center justify-between p-3 border rounded-md"
                >
                  <div>
                    <p class="font-medium text-sm">{{ req.playerName }}</p>
                    <p v-if="req.message" class="text-xs text-muted-foreground mt-1">"{{ req.message }}"</p>
                  </div>
                  <div class="flex gap-2">
                    <Button size="sm" class="h-7 text-xs" @click="acceptRequest(req.id)">
                      {{ t('alliance.accept') }}
                    </Button>
                    <Button size="sm" variant="outline" class="h-7 text-xs" @click="rejectRequest(req.id)">
                      {{ t('alliance.reject') }}
                    </Button>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <!-- Settings tab -->
        <TabsContent value="settings" class="mt-4">
          <Card>
            <CardContent class="pt-6 space-y-4">
              <div>
                <Label class="text-xs">{{ t('alliance.description') }}</Label>
                <Input v-model="settingsForm.description" class="mt-1" maxlength="500" />
              </div>
              <div>
                <Label class="text-xs">{{ t('alliance.maxMembers') }}</Label>
                <Input v-model.number="settingsForm.maxMembers" type="number" min="5" max="100" class="mt-1" />
              </div>
              <div class="flex items-center justify-between">
                <div>
                  <Label class="text-xs">{{ t('alliance.autoAccept') }}</Label>
                  <p class="text-xs text-muted-foreground">{{ t('alliance.autoAcceptHint') }}</p>
                </div>
                <Switch v-model="settingsForm.autoAccept" />
              </div>
              <div class="flex items-center justify-between">
                <div>
                  <Label class="text-xs">{{ t('alliance.requireApproval') }}</Label>
                  <p class="text-xs text-muted-foreground">{{ t('alliance.requireApprovalHint') }}</p>
                </div>
                <Switch v-model="settingsForm.requireApproval" />
              </div>
              <Button @click="saveSettings" :disabled="savingSettings">
                <Loader2 v-if="savingSettings" class="h-4 w-4 mr-2 animate-spin" />
                {{ t('alliance.saveSettings') }}
              </Button>

              <!-- Disband (leader only) -->
              <div v-if="allianceStore.isLeader" class="border-t pt-4 mt-4">
                <Button variant="destructive" @click="showDisbandConfirm = true">
                  {{ t('alliance.disband') }}
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </template>

    <!-- Leave confirmation dialog -->
    <Dialog v-model:open="showLeaveConfirm">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('alliance.leave') }}</DialogTitle>
          <DialogDescription>{{ t('alliance.leaveConfirm') }}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" @click="showLeaveConfirm = false">{{ t('common.cancel') }}</Button>
          <Button variant="destructive" @click="handleLeave">{{ t('alliance.leave') }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Disband confirmation dialog -->
    <Dialog v-model:open="showDisbandConfirm">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('alliance.disband') }}</DialogTitle>
          <DialogDescription>{{ t('alliance.disbandConfirm') }}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" @click="showDisbandConfirm = false">{{ t('common.cancel') }}</Button>
          <Button variant="destructive" @click="handleDisband">{{ t('alliance.disband') }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Kick confirmation dialog -->
    <Dialog v-model:open="showKickConfirm">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('alliance.kick') }}</DialogTitle>
          <DialogDescription>{{ t('alliance.kickConfirm', { name: kickTarget?.username }) }}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" @click="showKickConfirm = false">{{ t('common.cancel') }}</Button>
          <Button variant="destructive" @click="handleKick">{{ t('alliance.kick') }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { Loader2, Search } from 'lucide-vue-next'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Switch } from '@/components/ui/switch'
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter,
} from '@/components/ui/dialog'
import { useAllianceStore } from '@/stores/allianceStore'
import { useAuthStore } from '@/stores/authStore'
import type { AllianceMember } from '@/types/game'

const { t } = useI18n()
const allianceStore = useAllianceStore()
const authStore = useAuthStore()

const currentPlayerId = computed(() => authStore.user?.id || '')

// Create form
const createName = ref('')
const createTag = ref('')
const canCreate = computed(() => createName.value.trim().length >= 3 && /^[A-Za-z0-9]{3,8}$/.test(createTag.value.trim()))

// Search
const searchQuery = ref('')
const searchResults = ref<any[]>([])
const searched = ref(false)
const searching = ref(false)

// Tabs
const activeTab = ref('members')

// Settings form
const settingsForm = ref({
  description: '',
  maxMembers: 30,
  autoAccept: false,
  requireApproval: true,
})
const savingSettings = ref(false)

// Dialogs
const showLeaveConfirm = ref(false)
const showDisbandConfirm = ref(false)
const showKickConfirm = ref(false)
const kickTarget = ref<AllianceMember | null>(null)

// Sorted members — leader first, then officers, then members
const sortedMembers = computed(() => {
  if (!allianceStore.alliance) return []
  const roleOrder: Record<string, number> = { leader: 0, officer: 1, member: 2 }
  return [...allianceStore.alliance.members].sort((a, b) => {
    return (roleOrder[a.role] ?? 2) - (roleOrder[b.role] ?? 2)
  })
})

function roleVariant(role: string) {
  if (role === 'leader') return 'default'
  if (role === 'officer') return 'secondary'
  return 'outline'
}

// Init settings form when alliance loads
watch(() => allianceStore.alliance, (a) => {
  if (a) {
    settingsForm.value = {
      description: a.description || '',
      maxMembers: a.maxMembers || 30,
      autoAccept: a.autoAccept || false,
      requireApproval: a.requireApproval !== false,
    }
  }
}, { immediate: true })

async function handleCreate() {
  try {
    await allianceStore.createAlliance(createName.value.trim(), createTag.value.trim())
  } catch (e) {
    console.error('Failed to create alliance:', e)
  }
}

async function handleSearch() {
  if (!searchQuery.value.trim()) return
  searching.value = true
  searched.value = true
  try {
    searchResults.value = await allianceStore.searchAlliances(searchQuery.value.trim())
  } finally {
    searching.value = false
  }
}

async function requestJoin(allianceId: string) {
  try {
    await allianceStore.requestJoin(allianceId)
  } catch (e) {
    console.error('Failed to request join:', e)
  }
}

async function acceptInvite(inviteId: string) {
  try {
    await allianceStore.acceptInvite(inviteId)
  } catch (e) {
    console.error('Failed to accept invite:', e)
  }
}

async function rejectInvite(inviteId: string) {
  try {
    await allianceStore.rejectInvite(inviteId)
  } catch (e) {
    console.error('Failed to reject invite:', e)
  }
}

async function acceptRequest(requestId: string) {
  try {
    await allianceStore.acceptRequest(requestId)
  } catch (e) {
    console.error('Failed to accept request:', e)
  }
}

async function rejectRequest(requestId: string) {
  try {
    await allianceStore.rejectRequest(requestId)
  } catch (e) {
    console.error('Failed to reject request:', e)
  }
}

async function promote(playerId: string) {
  try {
    await allianceStore.updateMemberRole(playerId, 'officer')
  } catch (e) {
    console.error('Failed to promote:', e)
  }
}

async function demote(playerId: string) {
  try {
    await allianceStore.updateMemberRole(playerId, 'member')
  } catch (e) {
    console.error('Failed to demote:', e)
  }
}

function confirmKick(member: AllianceMember) {
  kickTarget.value = member
  showKickConfirm.value = true
}

async function handleKick() {
  if (!kickTarget.value) return
  try {
    await allianceStore.removeMember(kickTarget.value.playerId)
  } catch (e) {
    console.error('Failed to kick:', e)
  }
  showKickConfirm.value = false
  kickTarget.value = null
}

async function handleLeave() {
  try {
    await allianceStore.leaveAlliance()
  } catch (e) {
    console.error('Failed to leave:', e)
  }
  showLeaveConfirm.value = false
}

async function handleDisband() {
  // Disbanding = leader leaving
  try {
    await allianceStore.leaveAlliance()
  } catch (e) {
    console.error('Failed to disband:', e)
  }
  showDisbandConfirm.value = false
}

async function saveSettings() {
  savingSettings.value = true
  try {
    await allianceStore.updateSettings({
      description: settingsForm.value.description,
      maxMembers: settingsForm.value.maxMembers,
      autoAccept: settingsForm.value.autoAccept,
      requireApproval: settingsForm.value.requireApproval,
    })
  } catch (e) {
    console.error('Failed to save settings:', e)
  } finally {
    savingSettings.value = false
  }
}

// Translate raw backend error strings
const translatedError = computed(() => {
  const err = allianceStore.error
  if (!err) return ''
  const map: Record<string, string> = {
    'already in an alliance': t('alliance.errors.alreadyInAlliance'),
    'alliance not found': t('alliance.errors.allianceNotFound'),
    'player not found': t('alliance.errors.playerNotFound'),
    'not authorized': t('alliance.errors.notAuthorized'),
    'name already taken': t('alliance.errors.nameTaken'),
    'tag already taken': t('alliance.errors.tagTaken'),
  }
  return map[err] || err
})

onMounted(() => {
  allianceStore.clearError()
  // Refresh alliance state in case user joined/left via another path
  allianceStore.fetchMyAlliance()
})
</script>
