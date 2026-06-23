import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as recordsApi from '~/api/records'
import type { RecordData, ListRecordsParams } from '~/api/records'

export const useRecordsStore = defineStore('records', () => {
  const records = ref<RecordData[]>([])
  const total = ref(0)
  const loading = ref(false)

  async function fetchRecords(
    spaceSlug: string,
    tableSlug: string,
    params: ListRecordsParams = {}
  ): Promise<void> {
    loading.value = true
    try {
      const res = await recordsApi.listRecords(spaceSlug, tableSlug, params)
      records.value = res.data
      total.value = res.meta.total
    } finally {
      loading.value = false
    }
  }

  async function createRecord(
    spaceSlug: string,
    tableSlug: string,
    data: Record<string, unknown>
  ): Promise<RecordData> {
    const rec = await recordsApi.createRecord(spaceSlug, tableSlug, data)
    records.value.unshift(rec)
    total.value++
    return rec
  }

  async function updateRecord(
    spaceSlug: string,
    tableSlug: string,
    id: number,
    data: Record<string, unknown>
  ): Promise<RecordData> {
    const rec = await recordsApi.updateRecord(spaceSlug, tableSlug, id, data)
    const idx = records.value.findIndex((r) => r.id === id)
    if (idx !== -1) records.value[idx] = rec
    return rec
  }

  async function deleteRecord(spaceSlug: string, tableSlug: string, id: number): Promise<void> {
    await recordsApi.deleteRecord(spaceSlug, tableSlug, id)
    records.value = records.value.filter((r) => r.id !== id)
    total.value--
  }

  async function bulkDelete(spaceSlug: string, tableSlug: string, ids: number[]): Promise<void> {
    await Promise.all(ids.map((id) => recordsApi.deleteRecord(spaceSlug, tableSlug, id)))
    records.value = records.value.filter((r) => !ids.includes(r.id as number))
    total.value -= ids.length
  }

  return { records, total, loading, fetchRecords, createRecord, updateRecord, deleteRecord, bulkDelete }
})
