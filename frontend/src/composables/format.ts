export function relativeTime(iso?: string | null): string {
  if (!iso) return 'nunca'
  const diff = Math.round((Date.now() - new Date(iso).getTime()) / 1000)
  if (!Number.isFinite(diff)) return 'data indisponível'
  if (diff < 45) return 'agora há pouco'
  const rtf = new Intl.RelativeTimeFormat('pt-BR', { numeric: 'auto' })
  if (diff < 3600) return rtf.format(-Math.round(diff / 60), 'minute')
  if (diff < 86400) return rtf.format(-Math.round(diff / 3600), 'hour')
  return rtf.format(-Math.round(diff / 86400), 'day')
}

export function dateTime(iso?: string | null): string {
  if (!iso) return '—'
  if (!Number.isFinite(new Date(iso).getTime())) return '—'
  return new Date(iso).toLocaleString('pt-BR', { dateStyle: 'short', timeStyle: 'medium' })
}

export type StatusKey = 'online' | 'warning' | 'offline' | 'never_seen' | string

export const statusMeta: Record<string, { label: string; cls: string }> = {
  online: { label: 'Online', cls: 'badge-success' },
  warning: { label: 'Atenção', cls: 'badge-warning' },
  offline: { label: 'Offline', cls: 'badge-error' },
  never_seen: { label: 'Aguardando', cls: 'badge-ghost' },
  valid: { label: 'Válida', cls: 'badge-success' },
  invalid: { label: 'Inválida', cls: 'badge-error' },
}

export function statusInfo(status: string) {
  return statusMeta[status] ?? { label: status, cls: 'badge-ghost' }
}
