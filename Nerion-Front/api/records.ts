import { apiClient } from '~/utils/apiClient'

export interface RecordData {
  id: number
  [key: string]: unknown
}

export interface RecordsResponse {
  data: RecordData[]
  meta: {
    total: number
    limit: number
    offset: number
  }
}

export interface ListRecordsParams {
  limit?: number
  offset?: number
  search?: string
  sortBy?: string
  sortDir?: 'asc' | 'desc'
}

export const listRecords = (
  slug: string,
  table: string,
  params: ListRecordsParams = {}
): Promise<RecordsResponse> => {
  const q = new URLSearchParams()
  if (params.limit !== undefined) q.set('limit', String(params.limit))
  if (params.offset !== undefined) q.set('offset', String(params.offset))
  if (params.search) q.set('search', params.search)
  if (params.sortBy) q.set('sort_by', params.sortBy)
  if (params.sortDir) q.set('sort_dir', params.sortDir)
  const qs = q.toString()
  return apiClient.get(`/spaces/${slug}/tables/${table}/records${qs ? '?' + qs : ''}`)
}

export const createRecord = (
  slug: string,
  table: string,
  data: Record<string, unknown>
): Promise<RecordData> =>
  apiClient.post(`/spaces/${slug}/tables/${table}/records`, data)

export const updateRecord = (
  slug: string,
  table: string,
  id: number,
  data: Record<string, unknown>
): Promise<RecordData> =>
  apiClient.put(`/spaces/${slug}/tables/${table}/records/${id}`, data)

export const deleteRecord = (slug: string, table: string, id: number): Promise<null> =>
  apiClient.del(`/spaces/${slug}/tables/${table}/records/${id}`)
