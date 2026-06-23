import { apiClient } from '~/utils/apiClient'

export interface FieldMeta {
  id: number
  name: string
  slug: string
  type: string
  required: boolean
  options?: string[] | null
  target_table?: string | null
  multiple?: boolean
  accept?: string | null
}

export interface TableMeta {
  id: number
  name: string
  slug: string
  space_id: number
  fields: FieldMeta[]
  created_at: string
  updated_at: string
}

export interface TableTemplate {
  id: string
  name: string
  description: string
  fields: Omit<FieldMeta, 'id'>[]
}

export interface UpdateFieldsPayload {
  name: string
  slug: string
  type: string
  required?: boolean
  options?: string[] | null
  target_table?: string | null
  multiple?: boolean
  accept?: string | null
}

export const listTables = (slug: string): Promise<TableMeta[]> =>
  apiClient.get(`/spaces/${slug}/tables`)

export const createTable = (
  slug: string,
  name: string,
  tableSlug: string,
  templateId?: string | null
): Promise<TableMeta> =>
  apiClient.post(`/spaces/${slug}/tables`, {
    name,
    slug: tableSlug,
    ...(templateId ? { template_id: templateId } : {}),
  })

export const getTable = (slug: string, table: string): Promise<TableMeta> =>
  apiClient.get(`/spaces/${slug}/tables/${table}`)

export const updateFields = (
  slug: string,
  table: string,
  fields: UpdateFieldsPayload[]
): Promise<TableMeta> =>
  apiClient.put(`/spaces/${slug}/tables/${table}/fields`, { fields })

export const deleteTable = (slug: string, table: string): Promise<null> =>
  apiClient.del(`/spaces/${slug}/tables/${table}`)

export const listTemplates = (): Promise<TableTemplate[]> =>
  apiClient.get('/templates')
