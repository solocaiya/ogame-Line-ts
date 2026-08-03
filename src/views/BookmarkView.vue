<template>
  <div class="container mx-auto p-4 sm:p-6 space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl sm:text-3xl font-bold">{{ t('bookmark.title') }}</h1>
      <Button @click="showAddDialog = true">
        <Plus class="mr-2 h-4 w-4" />
        {{ t('bookmark.add') }}
      </Button>
    </div>

    <!-- 筛选 -->
    <Card>
      <CardContent class="p-4 flex flex-wrap gap-2">
        <Button
          variant="outline"
          size="sm"
          :class="{ 'bg-primary text-primary-foreground': filterCategory === 'all' }"
          @click="filterCategory = 'all'"
        >
          {{ t('bookmark.all') }}
        </Button>
        <Button
          v-for="cat in BOOKMARK_CATEGORIES"
          :key="cat"
          variant="outline"
          size="sm"
          :class="{ 'bg-primary text-primary-foreground': filterCategory === cat }"
          @click="filterCategory = cat"
        >
          {{ t(`bookmark.categories.${cat}`) }}
        </Button>
        <Button
          variant="outline"
          size="sm"
          :class="{ 'bg-yellow-500 text-white': showStarredOnly }"
          @click="showStarredOnly = !showStarredOnly"
        >
          <Star class="mr-1 h-3 w-3" />
          {{ t('bookmark.starred') }}
        </Button>
      </CardContent>
    </Card>

    <!-- 书签列表 -->
    <div v-if="!filteredBookmarks.length" class="text-center py-12 text-muted-foreground">
      <BookmarkIcon class="mx-auto h-12 w-12 mb-4 opacity-50" />
      <p>{{ t('bookmark.empty') }}</p>
    </div>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <Card
        v-for="bm in filteredBookmarks"
        :key="bm.id"
        class="relative"
        :class="{ 'ring-2 ring-yellow-400': bm.starred }"
      >
        <CardHeader class="pb-2">
          <div class="flex items-start justify-between">
            <div class="flex items-center gap-2">
              <div
                class="h-3 w-3 rounded-full"
                :style="{ backgroundColor: bm.color || '#6b7280' }"
              />
              <CardTitle class="text-sm">{{ bm.name }}</CardTitle>
            </div>
            <div class="flex items-center gap-1">
              <Button variant="ghost" size="icon" class="h-7 w-7" @click="toggleStar(bm.id)">
                <Star class="h-4 w-4" :class="bm.starred ? 'fill-yellow-400 text-yellow-400' : ''" />
              </Button>
              <Button variant="ghost" size="icon" class="h-7 w-7" @click="openEdit(bm)">
                <Pencil class="h-3 w-3" />
              </Button>
              <Button variant="ghost" size="icon" class="h-7 w-7" @click="remove(bm.id)">
                <Trash2 class="h-3 w-3 text-red-500" />
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent class="space-y-2">
          <div class="text-xs text-muted-foreground font-mono">
            [{{ bm.galaxy }}:{{ bm.system }}:{{ bm.position }}]
          </div>
          <Badge variant="outline" class="text-xs">
            {{ t(`bookmark.categories.${bm.category}`) }}
          </Badge>
          <p v-if="bm.note" class="text-xs text-muted-foreground mt-1 line-clamp-2">{{ bm.note }}</p>
        </CardContent>
      </Card>
    </div>

    <!-- 添加/编辑对话框 -->
    <Dialog v-model:open="showAddDialog">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ editingBookmark ? t('bookmark.edit') : t('bookmark.add') }}</DialogTitle>
        </DialogHeader>
        <div class="space-y-4">
          <div class="space-y-2">
            <Label>{{ t('bookmark.name') }}</Label>
            <Input v-model="form.name" :placeholder="t('bookmark.namePlaceholder')" />
          </div>
          <div class="grid grid-cols-3 gap-2">
            <div class="space-y-2">
              <Label>{{ t('bookmark.galaxy') }}</Label>
              <Input v-model.number="form.galaxy" type="number" :min="1" />
            </div>
            <div class="space-y-2">
              <Label>{{ t('bookmark.system') }}</Label>
              <Input v-model.number="form.system" type="number" :min="1" />
            </div>
            <div class="space-y-2">
              <Label>{{ t('bookmark.position') }}</Label>
              <Input v-model.number="form.position" type="number" :min="1" :max="15" />
            </div>
          </div>
          <div class="space-y-2">
            <Label>{{ t('bookmark.category') }}</Label>
            <select
              v-model="form.category"
              class="w-full h-9 rounded-md border border-input bg-background px-3 text-sm"
            >
              <option v-for="cat in BOOKMARK_CATEGORIES" :key="cat" :value="cat">
                {{ t(`bookmark.categories.${cat}`) }}
              </option>
            </select>
          </div>
          <div class="space-y-2">
            <Label>{{ t('bookmark.note') }}</Label>
            <textarea
              v-model="form.note"
              :placeholder="t('bookmark.notePlaceholder')"
              :rows="3"
              class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />
          </div>
          <div class="space-y-2">
            <Label>{{ t('bookmark.color') }}</Label>
            <div class="flex gap-2">
              <button
                v-for="color in BOOKMARK_COLORS"
                :key="color"
                class="h-7 w-7 rounded-full border-2 transition-transform"
                :class="{ 'border-white scale-110': form.color === color }"
                :style="{ backgroundColor: color }"
                @click="form.color = color"
              />
            </div>
          </div>
          <div class="flex justify-end gap-2">
            <Button variant="outline" @click="closeDialog">{{ t('common.cancel') }}</Button>
            <Button @click="saveBookmark" :disabled="!isValidForm">{{ t('common.save') }}</Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed } from 'vue'
  import { useGameStore } from '@/stores/gameStore'
  import { useI18n } from '@/composables/useI18n'
  import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
  import { Button } from '@/components/ui/button'
  import { Input } from '@/components/ui/input'
  import { Label } from '@/components/ui/label'
  import { Badge } from '@/components/ui/badge'
  import { Dialog, DialogContent, DialogHeader,DialogTitle } from '@/components/ui/dialog'
  import { Plus, Star, Pencil, Trash2, Bookmark as BookmarkIcon } from 'lucide-vue-next'
  import {
    BOOKMARK_CATEGORIES,
    BOOKMARK_COLORS,
    addBookmark,
    removeBookmark,
    updateBookmark,
    toggleBookmarkStar,
    type BookmarkCategory,
    type Bookmark
  } from '@/logic/bookmarkLogic'
  import { toast } from 'vue-sonner'

  const gameStore = useGameStore()
  const { t } = useI18n()

  const filterCategory = ref<'all' | BookmarkCategory>('all')
  const showStarredOnly = ref(false)
  const showAddDialog = ref(false)
  const editingBookmark = ref<Bookmark | null>(null)

  const form = ref({
    name: '',
    galaxy: 1,
    system: 1,
    position: 1,
    category: 'planet' as BookmarkCategory,
    note: '',
    color: BOOKMARK_COLORS[4]
  })

  const bookmarks = computed(() => gameStore.player.bookmarks || [])

  const filteredBookmarks = computed(() => {
    let list = bookmarks.value
    if (filterCategory.value !== 'all') {
      list = list.filter(b => b.category === filterCategory.value)
    }
    if (showStarredOnly.value) {
      list = list.filter(b => b.starred)
    }
    return list.sort((a, b) => (b.starred ? 1 : 0) - (a.starred ? 1 : 0) || b.timestamp - a.timestamp)
  })

  const isValidForm = computed(() => {
    return form.value.name.trim() && form.value.galaxy > 0 && form.value.system > 0 && form.value.position > 0 && form.value.position <= 15
  })

  const toggleStar = (id: string) => {
    toggleBookmarkStar(gameStore.player, id)
  }

  const remove = (id: string) => {
    if (removeBookmark(gameStore.player, id)) {
      toast.success(t('bookmark.deleted'))
    }
  }

  const openEdit = (bm: Bookmark) => {
    editingBookmark.value = bm
    form.value = {
      name: bm.name,
      galaxy: bm.galaxy,
      system: bm.system,
      position: bm.position,
      category: bm.category,
      note: bm.note,
      color: bm.color || BOOKMARK_COLORS[4]
    }
    showAddDialog.value = true
  }

  const closeDialog = () => {
    showAddDialog.value = false
    editingBookmark.value = null
    form.value = { name: '', galaxy: 1, system: 1, position: 1, category: 'planet', note: '', color: BOOKMARK_COLORS[4] }
  }

  const saveBookmark = () => {
    if (!isValidForm.value) return

    if (editingBookmark.value) {
      updateBookmark(gameStore.player, editingBookmark.value.id, { ...form.value })
      toast.success(t('bookmark.updated'))
    } else {
      addBookmark(gameStore.player, { ...form.value })
      toast.success(t('bookmark.added'))
    }
    closeDialog()
  }
</script>
