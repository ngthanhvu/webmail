<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { mdiClose, mdiLogout, mdiMagnify } from '@mdi/js'
import MdiIcon from '../components/MdiIcon.vue'
import { api } from '../api'
import { useAuth } from '../stores/auth'
import { addrString, fullDate, initial, mailboxLabel } from '../format'
import Sidebar from '../components/Sidebar.vue'
import MessageList from '../components/MessageList.vue'
import MessageDetail from '../components/MessageDetail.vue'
import ComposeModal from '../components/ComposeModal.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'

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
const confirmState = ref(null) // { type: 'message' | 'batch', uids: [] }
const deleting = ref(false)

// cache phía client: hiện dữ liệu cũ ngay, tải mới ở nền
const listCache = {}
let mailboxesCache = { t: 0, data: null }

function invalidateLists() {
  for (const k of Object.keys(listCache)) delete listCache[k]
}

const confirmTitle = computed(() => {
  const cs = confirmState.value
  if (!cs) return ''
  if (cs.type === 'message') return 'Xóa thư này?'
  return `Xóa ${cs.uids.length} thư?`
})

const confirmMessage = computed(() => {
  const cs = confirmState.value
  if (!cs) return ''
  if (cs.type === 'message') return 'Thư sẽ được chuyển vào Thùng rác. Bạn chắc chắn chứ?'
  return 'Các thư đã chọn sẽ được chuyển vào Thùng rác. Bạn chắc chắn chứ?'
})

const pageTitle = computed(() => {
  if (selected.value) {
    return selected.value.subject || '(không tiêu đề)'
  }
  const label = mailboxLabel(activeMailbox.value)
  return total.value > 0 ? `${label} (${total.value})` : label
})

const activeUnseen = computed(
  () => mailboxes.value.find((mb) => mb.name === activeMailbox.value)?.unseen || 0
)

watch(
  pageTitle,
  (t) => {
    document.title = t || 'Webmail'
  },
  { immediate: true }
)

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

async function loadMailboxes(force = false) {
  if (!force && mailboxesCache.data && Date.now() - mailboxesCache.t < 15000) {
    mailboxes.value = mailboxesCache.data
    return
  }
  try {
    const data = await api.mailboxes()
    mailboxes.value = data.mailboxes || []
    mailboxesCache = { t: Date.now(), data: mailboxes.value }
  } catch (e) {
    showToast('Lỗi: ' + e.message)
  }
}

async function loadMessages() {
  const key = `${activeMailbox.value}:${page.value}`
  const cached = listCache[key]
  if (cached) {
    // hiện ngay dữ liệu cũ, tải mới ở nền (stale-while-revalidate)
    messages.value = cached.messages
    total.value = cached.total
    loadingList.value = false
  } else {
    loadingList.value = true
    selected.value = null
  }
  try {
    const data = await api.messages(activeMailbox.value, page.value, limit)
    if (`${activeMailbox.value}:${page.value}` !== key) return // đã chuyển hộp thư khác
    messages.value = data.messages || []
    total.value = data.total || 0
    listCache[key] = { messages: messages.value, total: total.value }
  } catch (e) {
    if (!cached) {
      messages.value = []
      total.value = 0
    }
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

function deleteSelected() {
  if (!selected.value) return
  confirmState.value = { type: 'message' }
}

async function doDelete() {
  const cs = confirmState.value
  if (!cs || deleting.value) return
  deleting.value = true
  try {
    if (cs.type === 'message') {
      await api.flags(selected.value.uid, activeMailbox.value, { deleted: true })
      showToast('Đã xóa thư')
      selected.value = null
      invalidateLists()
      await loadMessages()
      loadMailboxes(true)
    } else {
      for (const uid of cs.uids) {
        await api.flags(uid, activeMailbox.value, { deleted: true })
      }
      showToast(`Đã xóa ${cs.uids.length} thư`)
      invalidateLists()
      await loadMessages()
      loadMailboxes(true)
    }
    confirmState.value = null
  } catch (e) {
    showToast('Lỗi: ' + e.message)
  }
  deleting.value = false
}

async function markUnread() {
  const m = selected.value
  if (!m) return
  try {
    await api.flags(m.uid, activeMailbox.value, { seen: false })
    selected.value = null
    invalidateLists()
    await loadMessages()
    loadMailboxes(true)
  } catch (e) {
    showToast('Lỗi: ' + e.message)
  }
}

async function runBatch({ action, uids }) {
  if (!uids.length) return
  if (action === 'delete') {
    confirmState.value = { type: 'batch', uids }
    return
  }
  try {
    for (const uid of uids) {
      if (action === 'star') await api.flags(uid, activeMailbox.value, { flagged: true })
      else if (action === 'unread') await api.flags(uid, activeMailbox.value, { seen: false })
      else if (action === 'read') await api.flags(uid, activeMailbox.value, { seen: true })
      else if (action === 'delete') await api.flags(uid, activeMailbox.value, { deleted: true })
    }
    if (action === 'delete') showToast(`Đã xóa ${uids.length} thư`)
    else showToast('Đã cập nhật')
    invalidateLists()
    await loadMessages()
    loadMailboxes(true)
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
    invalidateLists()
    loadMessages()
  }
  loadMailboxes(true)
}

async function logout() {
  await auth.logout()
  router.push('/login')
}
</script>

<template>
  <div class="flex h-screen flex-col overflow-hidden bg-oc-bg">
    <div class="flex min-h-0 flex-1">
      <Sidebar
        :mailboxes="mailboxes"
        :active="activeMailbox"
        :collapsed="sidebarCollapsed"
        @select="selectMailbox"
        @compose="openCompose"
        @toggle="sidebarCollapsed = !sidebarCollapsed"
      />

      <div class="flex min-w-0 flex-1 flex-col">
        <!-- header -->
        <header class="flex h-12 shrink-0 items-center gap-3 border-b border-oc-border bg-oc-panel px-3">
          <div
            class="flex h-8 max-w-md flex-1 items-center gap-2 border border-oc-border bg-oc-bg px-3 transition focus-within:border-oc-accent/60"
          >
            <MdiIcon :path="mdiMagnify" :size="15" class="shrink-0 text-oc-faint" />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Tìm trong hộp thư..."
              class="h-full min-w-0 flex-1 bg-transparent text-[13px] text-oc-text outline-none placeholder:text-oc-faint"
            />
            <button
              v-if="searchQuery"
              class="text-oc-faint transition hover:text-oc-text"
              title="Xóa tìm kiếm"
              @click="searchQuery = ''"
            >
              <MdiIcon :path="mdiClose" :size="15" />
            </button>
          </div>

          <div class="ml-auto flex shrink-0 items-center gap-2">
            <button
              class="flex h-8 w-8 items-center justify-center border border-oc-border text-oc-sub transition hover:border-oc-border2 hover:text-oc-text"
              title="Đăng xuất"
              @click="logout"
            >
              <MdiIcon :path="mdiLogout" :size="16" />
            </button>
            <div
              class="hidden h-8 items-center gap-2 border border-oc-border bg-oc-bg px-2 sm:flex"
              :title="auth.username"
            >
              <span
                class="flex h-5 w-5 items-center justify-center bg-oc-hover text-[10px] font-bold text-oc-sub"
                >{{ initial(auth.username) }}</span
              >
              <span class="max-w-40 truncate text-[11px] text-oc-sub">{{ auth.username }}</span>
            </div>
          </div>
        </header>

        <!-- body: list full-width, bấm thư -> mở trang đọc thay thế -->
        <div class="flex min-h-0 flex-1">
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
            <svg class="h-5 w-5 animate-spin text-oc-accent" fill="none" viewBox="0 0 24 24">
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
      </div>
    </div>

    <!-- status bar kiểu TUI -->
    <footer
      class="flex h-6 shrink-0 items-center gap-3 border-t border-oc-border bg-oc-panel px-3 text-[10px] text-oc-faint"
    >
      <span class="flex items-center gap-1.5">
        <span class="text-[8px] leading-none text-oc-accent">●</span>
        <span class="font-bold uppercase tracking-[0.15em] text-oc-sub">{{
          mailboxLabel(activeMailbox)
        }}</span>
      </span>
      <span class="hidden sm:inline">{{ total }} thư</span>
      <span v-if="activeUnseen > 0" class="text-oc-accent">{{ activeUnseen }} chưa đọc</span>
      <span class="ml-auto hidden truncate sm:inline">{{ auth.username }}</span>
    </footer>

    <ComposeModal v-if="compose" :draft="compose" @close="compose = null" @sent="onSent" />

    <ConfirmDialog
      :open="!!confirmState"
      :title="confirmTitle"
      :message="confirmMessage"
      confirm-text="Xóa"
      :loading="deleting"
      @confirm="doDelete"
      @cancel="confirmState = null"
    />

    <transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="translate-y-2 opacity-0"
      leave-active-class="transition duration-150 ease-in"
      leave-to-class="translate-y-2 opacity-0"
    >
      <div
        v-if="toast"
        class="fixed bottom-10 left-4 z-50 flex items-center gap-2.5 border border-oc-border2 border-l-2 border-l-oc-accent bg-oc-panel px-4 py-2.5 text-xs text-oc-text shadow-dialog"
      >
        {{ toast }}
      </div>
    </transition>
  </div>
</template>
