const JSON_HEADERS = { 'Content-Type': 'application/json' }

async function request(path, opts = {}) {
  const res = await fetch(path, { credentials: 'same-origin', ...opts })
  let data = null
  try {
    data = await res.json()
  } catch {
    /* non-JSON response */
  }
  if (!res.ok) {
    if (res.status === 401 && !path.endsWith('/login')) {
      window.dispatchEvent(new CustomEvent('auth:required'))
    }
    throw new Error((data && data.error) || `Lỗi HTTP ${res.status}`)
  }
  return data
}

export const api = {
  login: (username, password) =>
    request('/api/login', {
      method: 'POST',
      headers: JSON_HEADERS,
      body: JSON.stringify({ username, password }),
    }),
  logout: () => request('/api/logout', { method: 'POST' }),
  me: () => request('/api/me'),
  mailboxes: () => request('/api/mailboxes'),
  messages: (mailbox, page = 0, limit = 50) =>
    request(`/api/messages?mailbox=${encodeURIComponent(mailbox)}&page=${page}&limit=${limit}`),
  message: (uid, mailbox) =>
    request(`/api/messages/${uid}?mailbox=${encodeURIComponent(mailbox)}`),
  flags: (uid, mailbox, body) =>
    request(`/api/messages/${uid}/flags?mailbox=${encodeURIComponent(mailbox)}`, {
      method: 'POST',
      headers: JSON_HEADERS,
      body: JSON.stringify(body),
    }),
  attachmentUrl: (uid, mailbox, section, filename, contentType) => {
    let url = `/api/messages/${uid}/attachment?mailbox=${encodeURIComponent(mailbox)}&section=${encodeURIComponent(section)}`
    if (filename) url += `&filename=${encodeURIComponent(filename)}`
    if (contentType) url += `&contentType=${encodeURIComponent(contentType)}`
    return url
  },
  send: (payload) =>
    request('/api/send', {
      method: 'POST',
      headers: JSON_HEADERS,
      body: JSON.stringify(payload),
    }),
}
