import { apiClient } from '~/utils/apiClient'

export interface ApiKey {
  id: number
  name: string
  key?: string
  scope: 'read' | 'write'
  last_used_at: string | null
  created_at: string
}

export const listKeys = (slug: string): Promise<ApiKey[]> =>
  apiClient.get(`/spaces/${slug}/api-keys`)

export const createKey = (slug: string, name: string, scope: 'read' | 'write'): Promise<ApiKey> =>
  apiClient.post(`/spaces/${slug}/api-keys`, { name, scope })

export const revokeKey = (slug: string, id: number): Promise<null> =>
  apiClient.del(`/spaces/${slug}/api-keys/${id}`)
