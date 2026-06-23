import { apiClient } from '~/utils/apiClient'

export interface AuthTokens {
  access_token: string
  refresh_token: string
}

export interface User {
  id: number
  name: string
  email: string
  role: string
  email_verified: boolean
  created_at: string
}

export const login = (email: string, password: string): Promise<AuthTokens> =>
  apiClient.post('/auth/login', { email, password })

export const register = (name: string, email: string, password: string): Promise<AuthTokens> =>
  apiClient.post('/auth/register', { name, email, password })

export const logout = (refreshToken: string): Promise<null> =>
  apiClient.post('/auth/logout', { refresh_token: refreshToken })

export const resetRequest = (email: string): Promise<null> =>
  apiClient.post('/auth/password/reset-request', { email })

export const resetPassword = (token: string, password: string): Promise<null> =>
  apiClient.post('/auth/password/reset', { token, password })

export const getMe = (): Promise<User> =>
  apiClient.get('/me')
