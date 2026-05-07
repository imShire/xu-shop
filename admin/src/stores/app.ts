import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface TagItem {
  path: string
  title: string
  name?: string
}

export const useAppStore = defineStore('app', () => {
  const sidebarCollapsed = ref(false)
  const tags = ref<TagItem[]>([])
  const breadcrumbs = ref<{ title: string; path?: string }[]>([])

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function addTag(tag: TagItem) {
    const exists = tags.value.find((t) => t.path === tag.path)
    if (!exists) {
      tags.value.push(tag)
    }
  }

  function removeTag(path: string) {
    const idx = tags.value.findIndex((t) => t.path === path)
    if (idx > -1) {
      tags.value.splice(idx, 1)
    }
  }

  function removeOtherTags(path: string) {
    tags.value = tags.value.filter((t) => t.path === path)
  }

  function removeAllTags() {
    tags.value = []
  }

  function setBreadcrumbs(items: { title: string; path?: string }[]) {
    breadcrumbs.value = items
  }

  return {
    sidebarCollapsed,
    tags,
    breadcrumbs,
    toggleSidebar,
    addTag,
    removeTag,
    removeOtherTags,
    removeAllTags,
    setBreadcrumbs,
  }
})
