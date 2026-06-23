import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as schemaApi from '~/api/schema'
import type { TableMeta, UpdateFieldsPayload } from '~/api/schema'

export const useSchemaStore = defineStore('schema', () => {
  const tables = ref<TableMeta[]>([])
  const activeTable = ref<TableMeta | null>(null)
  const templates = ref<Awaited<ReturnType<typeof schemaApi.listTemplates>>>([])
  const loading = ref(false)

  async function fetchTables(spaceSlug: string): Promise<void> {
    loading.value = true
    try {
      tables.value = (await schemaApi.listTables(spaceSlug)) ?? []
    } finally {
      loading.value = false
    }
  }

  async function fetchTable(spaceSlug: string, tableSlug: string): Promise<TableMeta> {
    activeTable.value = await schemaApi.getTable(spaceSlug, tableSlug)
    return activeTable.value
  }

  async function createTable(
    spaceSlug: string,
    name: string,
    tableSlug: string,
    templateId?: string | null
  ): Promise<TableMeta> {
    const t = await schemaApi.createTable(spaceSlug, name, tableSlug, templateId)
    tables.value.push(t)
    return t
  }

  async function deleteTable(spaceSlug: string, tableSlug: string): Promise<void> {
    await schemaApi.deleteTable(spaceSlug, tableSlug)
    tables.value = tables.value.filter((t) => t.slug !== tableSlug)
  }

  async function updateFields(
    spaceSlug: string,
    tableSlug: string,
    fields: UpdateFieldsPayload[]
  ): Promise<void> {
    await schemaApi.updateFields(spaceSlug, tableSlug, fields)
    await fetchTable(spaceSlug, tableSlug)
  }

  async function fetchTemplates(): Promise<void> {
    if (templates.value.length) return
    templates.value = await schemaApi.listTemplates()
  }

  return {
    tables,
    activeTable,
    templates,
    loading,
    fetchTables,
    fetchTable,
    createTable,
    deleteTable,
    updateFields,
    fetchTemplates,
  }
})
