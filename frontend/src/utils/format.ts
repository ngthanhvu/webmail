import type { Address } from '@/types'

export function initial(s: string | null | undefined): string {
  const c = (s || '?').trim().charAt(0)
  return c ? c.toUpperCase() : '?'
}

export function formatBytes(n: number | null | undefined): string {
  if (n == null) return ''
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(0)} KB`
  return `${(n / (1024 * 1024)).toFixed(1)} MB`
}

export function shortDate(d: string | null | undefined): string {
  if (!d) return ''
  const date = new Date(d)
  if (Number.isNaN(date.getTime())) return ''
  const now = new Date()
  const sameDay =
    date.getDate() === now.getDate() &&
    date.getMonth() === now.getMonth() &&
    date.getFullYear() === now.getFullYear()
  if (sameDay) {
    return date.toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' })
  }
  if (date.getFullYear() === now.getFullYear()) {
    return `${date.getDate()} thg ${date.getMonth() + 1}`
  }
  return String(date.getFullYear())
}

export function fullDate(d: string | null | undefined): string {
  if (!d) return ''
  const date = new Date(d)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleString('vi-VN', {
    hour: '2-digit',
    minute: '2-digit',
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })
}

export function addrString(a: Address | null | undefined): string {
  if (!a) return ''
  if (a.name) return `${a.name} <${a.email}>`
  return a.email || ''
}

export function splitEmails(s: string): string[] {
  return String(s || '')
    .split(/[,;]/)
    .map((x) => x.trim())
    .filter(Boolean)
}

export function mailboxLabel(name: string): string {
  const map: Record<string, string> = {
    INBOX: 'Hộp thư đến',
    'INBOX/Drafts': 'Nháp',
    'INBOX/Sent': 'Đã gửi',
    'INBOX/Trash': 'Thùng rác',
    'INBOX/Junk': 'Thư rác',
  }
  if (map[name]) return map[name]
  const lower = name.toLowerCase()
  if (lower === 'sent' || lower.endsWith('/sent') || lower === 'sent messages' || lower === 'sent items' || lower.endsWith('/sent mail')) return 'Đã gửi'
  if (lower === 'drafts' || lower.endsWith('/drafts') || lower === 'draft') return 'Nháp'
  if (lower === 'trash' || lower.endsWith('/trash') || lower === 'deleted messages' || lower === 'deleted items' || lower === 'bin') return 'Thùng rác'
  if (lower === 'junk' || lower === 'spam' || lower.endsWith('/junk') || lower.endsWith('/spam') || lower === 'bulk mail') return 'Thư rác'
  if (lower === 'archive' || lower.endsWith('/archive') || lower === 'all mail') return 'Lưu trữ'
  if (lower.endsWith('important')) return 'Quan trọng'
  if (lower.endsWith('starred')) return 'Đánh dấu sao'
  return name
}
