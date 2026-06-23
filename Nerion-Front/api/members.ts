import { apiClient } from '~/utils/apiClient'

export interface SpaceMember {
  id: number
  user_id: number
  space_id: number
  role: 'admin' | 'member'
  name: string
  email: string
  joined_at: string
}

export const listMembers = (slug: string): Promise<SpaceMember[]> =>
  apiClient.get(`/spaces/${slug}/members`)

export const inviteMember = (slug: string, email: string): Promise<SpaceMember> =>
  apiClient.post(`/spaces/${slug}/members/invite`, { email })

export const changeMemberRole = (
  slug: string,
  userId: number,
  role: 'admin' | 'member'
): Promise<SpaceMember> =>
  apiClient.put(`/spaces/${slug}/members/${userId}/role`, { role })

export const removeMember = (slug: string, userId: number): Promise<null> =>
  apiClient.del(`/spaces/${slug}/members/${userId}`)
