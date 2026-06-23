import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as spacesApi from '~/api/spaces'
import type { Space } from '~/api/spaces'

export const useSpacesStore = defineStore('spaces', () => {
  const spaces = ref<Space[]>([])
  const loading = ref(false)

  async function fetchSpaces(): Promise<void> {
    loading.value = true
    try {
      spaces.value = await spacesApi.listSpaces()
    } finally {
      loading.value = false
    }
  }

  async function createSpace(name: string, slug: string): Promise<Space> {
    const sp = await spacesApi.createSpace(name, slug)
    spaces.value.push(sp)
    return sp
  }

  async function renameSpace(slug: string, name: string): Promise<void> {
    await spacesApi.renameSpace(slug, name)
    const s = spaces.value.find((s) => s.slug === slug)
    if (s) s.name = name
  }

  async function deleteSpace(slug: string, confirmName: string): Promise<void> {
    await spacesApi.deleteSpace(slug, confirmName)
    spaces.value = spaces.value.filter((s) => s.slug !== slug)
  }

  return { spaces, loading, fetchSpaces, createSpace, renameSpace, deleteSpace }
})
