import { apiClient } from '~/utils/apiClient'

export interface SpaceList {
  id: number
  slug: string
  name: string
  table_slug: string
  field_filter: string | null
  published: boolean
  created_at: string
  updated_at: string
}

export interface CreateListPayload {
  name: string
  table_slug: string
  field_filter?: string | null
  published?: boolean
}

export interface UpdateListPayload {
  name?: string
  table_slug?: string
  field_filter?: string | null
  published?: boolean
}

export const listLists = (spaceSlug: string): Promise<SpaceList[]> =>
  apiClient.get(`/spaces/${spaceSlug}/lists`)

export const createList = (spaceSlug: string, data: CreateListPayload): Promise<SpaceList> =>
  apiClient.post(`/spaces/${spaceSlug}/lists`, data)

export const updateList = (
  spaceSlug: string,
  listSlug: string,
  data: UpdateListPayload
): Promise<SpaceList> =>
  apiClient.put(`/spaces/${spaceSlug}/lists/${listSlug}`, data)
