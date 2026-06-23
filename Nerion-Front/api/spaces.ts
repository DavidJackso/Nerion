import { apiClient } from '~/utils/apiClient'

export interface Space {
  id: number
  name: string
  slug: string
  owner_id: number
  created_at: string
  updated_at: string
}

export const listSpaces = (): Promise<Space[]> =>
  apiClient.get('/spaces')

export const createSpace = (name: string, slug: string): Promise<Space> =>
  apiClient.post('/spaces', { name, slug })

export const renameSpace = (slug: string, name: string): Promise<Space> =>
  apiClient.put(`/spaces/${slug}`, { name })

export const deleteSpace = (slug: string, confirmName: string): Promise<null> =>
  apiClient.del(`/spaces/${slug}`, { confirm_name: confirmName })
