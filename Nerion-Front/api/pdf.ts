import { apiClient } from '~/utils/apiClient'

export interface PdfTemplate {
  id: number
  name: string
  file_url: string
  space_id: number
  created_at: string
}

export interface PdfMapping {
  placeholder: string
  table: string
  field: string
  format?: string
}

export interface PdfJob {
  id: number
  template_id: number
  status: 'pending' | 'processing' | 'done' | 'failed'
  output_url: string | null
  created_at: string
  updated_at: string
}

export interface GeneratePdfPayload {
  template_id: number
  record_ids: number[]
}

function getApiBaseUrl(): string {
  try {
    const config = useRuntimeConfig()
    if (config.public.apiUrl) return String(config.public.apiUrl) + '/api/v1'
  } catch {
    // outside Nuxt context
  }
  return (
    (typeof process !== 'undefined' && process.env?.NUXT_PUBLIC_API_URL
      ? process.env.NUXT_PUBLIC_API_URL
      : '') + '/api/v1'
  )
}

export const listTemplates = (spaceSlug: string): Promise<PdfTemplate[]> =>
  apiClient.get(`/spaces/${spaceSlug}/pdf/templates`)

export async function uploadTemplate(
  spaceSlug: string,
  name: string,
  file: File
): Promise<PdfTemplate> {
  const token = import.meta.client ? localStorage.getItem('nerion_access_token') : null
  const fd = new FormData()
  fd.append('name', name)
  fd.append('file', file)
  const res = await fetch(`${getApiBaseUrl()}/spaces/${spaceSlug}/pdf/templates`, {
    method: 'POST',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
    body: fd,
  })
  if (!res.ok) {
    const d = await res.json().catch(() => ({}))
    throw new Error(
      (d as { error?: { message?: string } })?.error?.message ?? 'Ошибка загрузки'
    )
  }
  return res.json() as Promise<PdfTemplate>
}

export const saveMapping = (
  spaceSlug: string,
  templateId: number,
  mappings: PdfMapping[]
): Promise<null> =>
  apiClient.post(`/spaces/${spaceSlug}/pdf/templates/${templateId}/mapping`, { mappings })

export const generatePDF = (spaceSlug: string, data: GeneratePdfPayload): Promise<PdfJob> =>
  apiClient.post(`/spaces/${spaceSlug}/pdf/generate`, data)

export const getJob = (spaceSlug: string, jobId: number): Promise<PdfJob> =>
  apiClient.get(`/spaces/${spaceSlug}/pdf/jobs/${jobId}`)

export const listArchive = (spaceSlug: string): Promise<PdfJob[]> =>
  apiClient.get(`/spaces/${spaceSlug}/pdf/archive`)
