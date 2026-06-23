export function fmtDate(dt: string | Date | null | undefined): string {
  if (!dt) return '—'
  return new Date(dt).toLocaleDateString('ru')
}

export function initials(name: string | null | undefined): string {
  return String(name ?? '')
    .split(' ')
    .map((p) => p[0])
    .join('')
    .slice(0, 2)
    .toUpperCase()
}
