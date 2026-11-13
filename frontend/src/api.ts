import type {
  FlagsPayload,
  Mailbox,
  Message,
  MessageSummary,
  SendPayload,
} from '@/types'

const JSON_HEADERS = { 'Content-Type': 'application/json' }

async function request<T>(path: string, opts: RequestInit = {}): Promise<T> {
  const res = await fetch(path, { credentials: 'same-origin', ...opts })
  let data: unknown = null
  try {
    data = await res.json()
  } catch {
    /* non-JSON response */
  }
  if (!res.ok) {
    if (res.status === 401 && !path.endsWith('/login')) {
      window.dispatchEvent(new CustomEvent('auth:required'))
    }
    const message = (data as { error?: string } | null)?.error ?? `Lỗi HTTP ${res.status}`
    throw new Error(message)
  }
  return data as T
}

export const api = {
  login: (username: string, password: string) =>
    request<{ ok: boolean; username: string }>('/api/login', {
      method: 'POST',
      headers: JSON_HEADERS,
      body: JSON.stringify({ username, password }),
    }),
  logout: () => request<{ ok: boolean }>('/api/logout', { method: 'POST' }),
  me: () => request<{ username: string }>('/api/me'),
  mailboxes: () => request<{ mailboxes: Mailbox[] }>('/api/mailboxes'),
  messages: (mailbox: string, page = 0, limit = 50) =>
    request<{ messages: MessageSummary[]; total: number }>(
      `/api/messages?mailbox=${encodeURIComponent(mailbox)}&page=${page}&limit=${limit}`
    ),
  message: (uid: number, mailbox: string) =>
    request<Message>(`/api/messages/${uid}?mailbox=${encodeURIComponent(mailbox)}`),
  flags: (uid: number, mailbox: string, body: FlagsPayload) =>
    request<{ ok: boolean }>(
      `/api/messages/${uid}/flags?mailbox=${encodeURIComponent(mailbox)}`,
      {
        method: 'POST',
        headers: JSON_HEADERS,
        body: JSON.stringify(body),
      }
    ),
  attachmentUrl: (
    uid: number,
    mailbox: string,
    section: string,
    filename?: string,
    contentType?: string
  ) => {
    let url = `/api/messages/${uid}/attachment?mailbox=${encodeURIComponent(mailbox)}&section=${encodeURIComponent(section)}`
    if (filename) url += `&filename=${encodeURIComponent(filename)}`
    if (contentType) url += `&contentType=${encodeURIComponent(contentType)}`
    return url
  },
  send: (payload: SendPayload) =>
    request<{ ok: boolean }>('/api/send', {
      method: 'POST',
      headers: JSON_HEADERS,
      body: JSON.stringify(payload),
    }),
}
