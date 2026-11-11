<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { mdiClose, mdiEmailOutline, mdiLogout, mdiMagnify, mdiMenu } from '@mdi/js'
import MdiIcon from '../components/MdiIcon.vue'
import { api } from '../api'
import { useAuth } from '../stores/auth'
import { addrString, avatarColor, fullDate, initial } from '../format'
import Sidebar from '../components/Sidebar.vue'
import MessageList from '../components/MessageList.vue'
import MessageDetail from '../components/MessageDetail.vue'
import ComposeModal from '../components/ComposeModal.vue'

const router = useRouter()
const auth = useAuth()

const mailboxes = ref([])
const activeMailbox = ref('INBOX')
const messages = ref([])
const total = ref(0)
const page = ref(0)
const limit = 50
const loadingList = ref(false)
const selected = ref(null)
const loadingDetail = ref(false)
const compose = ref(null)
const toast = ref('')
let toastTimer = null
const searchQuery = ref('')
const sidebarCollapsed = ref(false)

const filteredMessages = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return messages.value
  return messages.value.filter(
    (m) =>
      (m.subject || '').toLowerCase().includes(q) ||
      (m.from || '').toLowerCase().includes(q) ||
      (m.fromEmail || '').toLowerCase().includes(q) ||
      (m.snippet || '').toLowerCase().includes(q)
  )
})

function showToast(msg) {
  toast.value = msg
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => (toast.value = ''), 3000)
}

onMounted(async () => {
  await loadMailboxes()
  await loadMessages()
})

async function loadMailboxes() {
  try {
    const data = await api.mailboxes()
    mailboxes.value = data.mailboxes || []
  } catch (e) {
    showToast('Lỗi: ' + e.message)
  }
}

async function loadMessages() {
  loadingList.value = true
  selected.value = null
  try {
    const data = await api.messages(activeMailbox.value, page.value, limit)
    messages.value = data.messages || []
    total.value = data.total || 0
  } catch (e) {
    messages.value = []
    total.value = 0
    showToast('Lỗi: ' + e.message)
  }
  loadingList.value = false
}

function selectMailbox(name) {
  if (name === activeMailbox.value) return
  activeMailbox.value = name
  page.value = 0
  loadMessages()
}

function changePage(p) {
  page.value = p
  loadMessages()
}

function refresh() {
  loadMessages()
  loadMailboxes()
}

async function openMessage(m) {
  if (loadingDetail.value) return
  loadingDetail.value = true
  try {
    const data = await api.message(m.uid, activeMailbox.value)
    selected.value = data
    m.seen = true
  } catch (e) {
    showToast('Lỗi: ' + e.message)
  }
  loadingDetail.value = false
}

function closeMessage() {
  selected.value = null
  loadMailboxes()
}

async function toggleStar(m) {
  try {
    await api.flags(m.uid, activeMailbox.value, { flagged: !m.flagged })
    m.flagged = !m.flagged
    if (selected.value && selected.value.uid === m.uid) {
      const i = (selected.value.flags || []).indexOf('\\Flagged')
      if (m.flagged && i < 0) selected.value.flags.push('\\Flagged')
      if (!m.flagged && i >= 0) selected.value.flags.splice(i, 1)
    }
  } catch (e) {
    showToast('Lỗi: ' + e.message)
  }
}

async function deleteSelected() {
  const m = selected.value
  if (!m) return
  try {
    await api.flags(m.uid, activeMailbox.value, { deleted: true })
    showToast('Đã xóa thư')
    selected.value = null
    await loadMessages()
    loadMailboxes()
  } catch (e) {
    showToast('Lỗi: ' + e.message)
  }
}

async function markUnread() {
  const m = selected.value
  if (!m) return
  try {
    await api.flags(m.uid, activeMailbox.value, { seen: false })
    selected.value = null
    await loadMessages()
    loadMailboxes()
  } catch (e) {
    showToast('Lỗi: ' + e.message)
  }
}

async function runBatch({ action, uids }) {
  if (!uids.length) return
  try {
    for (const uid of uids) {
      if (action === 'star') await api.flags(uid, activeMailbox.value, { flagged: true })
      else if (action === 'unread') await api.flags(uid, activeMailbox.value, { seen: false })
      else if (action === 'read') await api.flags(uid, activeMailbox.value, { seen: true })
      else if (action === 'delete') await api.flags(uid, activeMailbox.value, { deleted: true })
    }
    if (action === 'delete') showToast(`Đã xóa ${uids.length} thư`)
    else showToast('Đã cập nhật')
    await loadMessages()
    loadMailboxes()
  } catch (e) {
    showToast('Lỗi: ' + e.message)
  }
}

function openCompose() {
  compose.value = {
    to: [],
    cc: [],
    bcc: [],
    subject: '',
    text: '',
    inReplyTo: '',
    references: '',
  }
}

function openReply(m, all = false) {
  const to = all
    ? [...(m.to || []).map(addrString), ...(m.cc || []).map(addrString)].filter(
        (x) => x && x !== auth.username
      )
    : [m.from && m.from.email].filter(Boolean)
  compose.value = {
    to,
    cc: [],
    bcc: [],
    subject: /^re:/i.test(m.subject || '') ? m.subject : 'Re: ' + (m.subject || ''),
    text: buildQuote(m),
    inReplyTo: m.messageId || '',
    references: m.references || '',
  }
}

function openForward(m) {
  compose.value = {
    to: [],
    cc: [],
    bcc: [],
    subject: /^fwd:/i.test(m.subject || '') ? m.subject : 'Fwd: ' + (m.subject || ''),
    text: buildQuote(m, true),
    inReplyTo: '',
    references: '',
  }
}

function buildQuote(m, forward = false) {
  const from = m.from ? (m.from.name || m.from.email) : ''
  const header = forward
    ? `---------- Thư chuyển tiếp ----------\nTừ: ${from}\nNgày: ${fullDate(m.date)}\nTiêu đề: ${m.subject || ''}\n\n`
    : `Vào ${fullDate(m.date)}, ${from} đã viết:\n\n`
  const body = (m.text || '').split('\n').map((l) => (l ? '> ' + l : '>')).join('\n')
  return '\n\n' + header + body
}

function onSent() {
  compose.value = null
  showToast('Đã gửi thư')
  if (activeMailbox.value.toLowerCase().includes('sent')) {
    loadMessages()
  }
  loadMailboxes()
}

async function logout() {
  await auth.logout()
  router.push('/login')
}
</script>

<template>
  <div class="flex h-screen flex-col overflow-hidden bg-white">
    <!-- header -->
    <header class="flex h-16 shrink-0 items-center gap-2 bg-white px-2 sm:px-4">
      <button
        class="hidden h-10 w-10 shrink-0 items-center justify-center rounded-full text-gmail-sub transition hover:bg-gmail-hover sm:flex"
        :title="sidebarCollapsed ? 'Mở rộng menu' : 'Thu gọn menu'"
        @click="sidebarCollapsed = !sidebarCollapsed"
      >
        <MdiIcon :path="mdiMenu" :size="23" />
      </button>

      <div class="flex shrink-0 items-center gap-2 pr-1">
        <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-gmail-blue text-white shadow-sm">
          <MdiIcon :path="mdiEmailOutline" :size="22" />
        </div>
        <span class="hidden text-[22px] font-normal text-gmail-sub xl:block">Webmail</span>
      </div>

      <div class="flex h-12 max-w-3xl flex-1 items-center gap-3 rounded-full bg-gmail-hover px-5">
        <MdiIcon :path="mdiMagnify" :size="21" class="shrink-0 text-gmail-sub" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Tìm trong hộp thư"
          class="h-full min-w-0 flex-1 bg-transparent text-sm text-gmail-text outline-none placeholder:text-gmail-light"
        />
        <button v-if="searchQuery" class="text-gmail-sub hover:text-gmail-text" title="Xóa tìm kiếm" @click="searchQuery = ''">
          <MdiIcon :path="mdiClose" :size="20" />
        </button>
      </div>

      <div class="ml-auto flex shrink-0 items-center gap-2">
        <button
          class="flex h-10 w-10 items-center justify-center rounded-full text-gmail-sub transition hover:bg-gmail-hover"
          title="Đăng xuất"
          @click="logout"
        >
          <MdiIcon :path="mdiLogout" :size="21" />
        </button>
        <div
          class="flex h-8 w-8 items-center justify-center rounded-full text-sm font-bold text-white"
          :style="{ backgroundColor: avatarColor(auth.username || 'x') }"
          :title="auth.username"
        >
          {{ initial(auth.username) }}
        </div>
      </div>
    </header>

    <!-- body: list full-width, bấm thư -> mở trang đọc thay thế -->
    <div class="flex min-h-0 flex-1">
      <Sidebar
        :mailboxes="mailboxes"
        :active="activeMailbox"
        :collapsed="sidebarCollapsed"
        @select="selectMailbox"
        @compose="openCompose"
      />

      <MessageDetail
        v-if="selected"
        :message="selected"
        :mailbox="activeMailbox"
        @back="closeMessage"
        @reply="openReply"
        @reply-all="(m) => openReply(m, true)"
        @forward="openForward"
        @delete="deleteSelected"
        @unread="markUnread"
        @star="toggleStar"
      />

      <div v-else-if="loadingDetail" class="flex flex-1 items-center justify-center">
        <svg class="h-6 w-6 animate-spin text-gmail-light" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z" />
        </svg>
      </div>

      <MessageList
        v-else
        :messages="filteredMessages"
        :total="total"
        :page="page"
        :limit="limit"
        :mailbox="activeMailbox"
        :loading="loadingList"
        @select="openMessage"
        @star="toggleStar"
        @page="changePage"
        @reply="openReply"
        @refresh="refresh"
        @batch="runBatch"
      />
    </div>

    <ComposeModal v-if="compose" :draft="compose" @close="compose = null" @sent="onSent" />

    <transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="translate-y-2 opacity-0"
      leave-active-class="transition duration-150 ease-in"
      leave-to-class="translate-y-2 opacity-0"
    >
      <div v-if="toast" class="fixed bottom-6 left-6 z-50 rounded bg-[#323639] px-5 py-3 text-sm text-white shadow-lg">
        {{ toast }}
      </div>
    </transition>
  </div>
</template>
