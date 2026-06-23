import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as authApi from '~/api/auth'
import type { User } from '~/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const accessToken = ref<string | null>(
    import.meta.client ? localStorage.getItem('nerion_access_token') : null
  )
  const refreshToken = ref<string | null>(
    import.meta.client ? localStorage.getItem('nerion_refresh_token') : null
  )

  async function login(email: string, password: string): Promise<void> {
    const data = await authApi.login(email, password)
    accessToken.value = data.access_token
    refreshToken.value = data.refresh_token
    if (import.meta.client) {
      localStorage.setItem('nerion_access_token', data.access_token)
      localStorage.setItem('nerion_refresh_token', data.refresh_token)
    }
    await fetchMe()
  }

  async function logout(): Promise<void> {
    try {
      if (refreshToken.value) await authApi.logout(refreshToken.value)
    } catch {
      // ignore errors on logout
    }
    accessToken.value = null
    refreshToken.value = null
    user.value = null
    if (import.meta.client) {
      localStorage.removeItem('nerion_access_token')
      localStorage.removeItem('nerion_refresh_token')
    }
  }

  async function fetchMe(): Promise<void> {
    try {
      user.value = await authApi.getMe()
    } catch {
      user.value = null
    }
  }

  function initials(): string {
    if (!user.value?.name) return '?'
    return user.value.name
      .split(' ')
      .map((p) => p[0])
      .join('')
      .slice(0, 2)
      .toUpperCase()
  }

  return { user, accessToken, refreshToken, login, logout, fetchMe, initials }
})
