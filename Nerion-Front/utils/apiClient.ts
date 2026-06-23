export class ApiError extends Error {
  code: string
  fields: Record<string, string> | null

  constructor(code: string, message: string, fields: Record<string, string> | null = null) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.fields = fields
  }
}

function getBaseUrl(): string {
  // useRuntimeConfig is only available inside setup/composable context.
  // At module level we fall back to the env var injected by Nuxt at build time.
  try {
    const config = useRuntimeConfig()
    if (config.public.apiUrl) return String(config.public.apiUrl) + '/api/v1'
  } catch {
    // Outside of Nuxt context (e.g. Vitest unit tests)
  }
  return (
    (typeof process !== 'undefined' && process.env?.NUXT_PUBLIC_API_URL
      ? process.env.NUXT_PUBLIC_API_URL
      : '') + '/api/v1'
  )
}

let _refreshing: Promise<void> | null = null

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  retried = false
): Promise<T> {
  const token = import.meta.client ? localStorage.getItem('nerion_access_token') : null
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(getBaseUrl() + path, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  if (res.status === 401 && !retried && import.meta.client) {
    if (!_refreshing) {
      _refreshing = (async () => {
        const rt = localStorage.getItem('nerion_refresh_token')
        if (!rt) throw new ApiError('unauthorized', 'Требуется авторизация')
        const r = await fetch(getBaseUrl() + '/auth/refresh', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: rt }),
        })
        if (!r.ok) throw new ApiError('unauthorized', 'Сессия истекла')
        const d = await r.json()
        localStorage.setItem('nerion_access_token', d.access_token)
        localStorage.setItem('nerion_refresh_token', d.refresh_token)
      })()
    }
    try {
      await _refreshing
    } finally {
      _refreshing = null
    }
    return request<T>(method, path, body, true)
  }

  if (res.status === 204) return null as T

  const text = await res.text()
  if (!text) throw new ApiError('error', 'Ошибка сервера: пустой ответ')

  let data: unknown
  try {
    data = JSON.parse(text)
  } catch {
    throw new ApiError('error', `Ошибка сервера (${res.status})`)
  }

  if (!res.ok) {
    const err = (data as { error?: { code?: string; message?: string; fields?: Record<string, string> } })?.error ?? {}
    throw new ApiError(err.code ?? 'error', err.message ?? 'Ошибка сервера', err.fields ?? null)
  }

  return data as T
}

export const apiClient = {
  get: <T = unknown>(path: string) => request<T>('GET', path),
  post: <T = unknown>(path: string, body?: unknown) => request<T>('POST', path, body),
  put: <T = unknown>(path: string, body?: unknown) => request<T>('PUT', path, body),
  del: <T = unknown>(path: string, body?: unknown) => request<T>('DELETE', path, body),
}
