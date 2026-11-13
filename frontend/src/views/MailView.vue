<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { mdiClose, mdiLogout, mdiMagnify, mdiMenu, mdiWeatherNight, mdiWeatherSunny } from '@mdi/js'
import MdiIcon from '@/components/MdiIcon.vue'
import { api } from '@/api'
import { useAuth } from '@/stores/auth'
import { useTheme } from '@/stores/theme'
import { addrString, fullDate, initial, mailboxLabel } from '@/utils/format'
import type {
  BatchPayload,
  ConfirmState,
  Draft,
  Mailbox,
  Message,
  MessageSummary,
  StarTarget,
} from '@/types'
import Sidebar from '@/components/Sidebar.vue'
import MessageList from '@/components/MessageList.vue'
import MessageDetail from '@/components/MessageDetail.vue'
import ComposeModal from '@/components/ComposeModal.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'

const router = useRouter()
const auth = useAuth()
const theme = useTheme()

const mailboxes = ref<Mailbox[]>([])
const activeMailbox = ref('INBOX')
const messages = ref<MessageSummary[]>([])
const total = ref(0)
const page = ref(0)
const limit = 50
const loadingList = ref(false)
const selected = ref<Message | null>(null)
const loadingDetail = ref(false)
const compose = ref<Draft | null>(null)
const toast = ref('')
let toastTimer: ReturnType<typeof setTimeout> | null = null
const searchQuery = ref('')
const sidebarCollapsed = ref(false)
const mobileNav = ref(false)
const confirmState = ref<ConfirmState | null>(null)
const deleting = ref(false)

// cache phía client: hiện dữ liệu cũ ngay, tải mới ở nền
const listCache = new Map<string, { messages: MessageSummary[]; total: number }>()
let mailboxesCache: { t: number; data: Mailbox[] | null } = { t: 0, data: null }

function invalidateLists(): void {
  listCache.clear()
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

function showToast(msg: string): void {
  toast.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => (toast.value = ''), 3000)
}

function errorMessage(e: unknown): string {
  return e instanceof Error ? e.message : String(e)
}

onMounted(async () => {
  await loadMailboxes()
  await loadMessages()
})

async function loadMailboxes(force = false): Promise<void> {
  if (!force && mailboxesCache.data && Date.now() - mailboxesCache.t < 15000) {
    mailboxes.value = mailboxesCache.data
    return
  }
  try {
    const data = await api.mailboxes()
    mailboxes.value = data.mailboxes || []
    mailboxesCache = { t: Date.now(), data: mailboxes.value }
  } catch (e) {
    showToast('Lỗi: ' + errorMessage(e))
  }
}

async function loadMessages(): Promise<void> {
  const key = `${activeMailbox.value}:${page.value}`
  const cached = listCache.get(key)
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
    listCache.set(key, { messages: messages.value, total: total.value })
  } catch (e) {
    if (!cached) {
      messages.value = []
      total.value = 0
    }
    showToast('Lỗi: ' + errorMessage(e))
  }
  loadingList.value = false
}

function selectMailbox(name: string): void {
  if (name === activeMailbox.value) return
  activeMailbox.value = name
  page.value = 0
  loadMessages()
}

function selectMailboxMobile(name: string): void {
  mobileNav.value = false
  selectMailbox(name)
}

function openComposeMobile(): void {
  mobileNav.value = false
  openCompose()
}

function changePage(p: number): void {
  page.value = p
  loadMessages()
}

function refresh(): void {
  loadMessages()
  loadMailboxes()
}

async function openMessage(m: MessageSummary): Promise<void> {
  if (loadingDetail.value) return
  loadingDetail.value = true
  try {
    const data = await api.message(m.uid, activeMailbox.value)
    selected.value = data
    m.seen = true
  } catch (e) {
    showToast('Lỗi: ' + errorMessage(e))
  }
  loadingDetail.value = false
}

function closeMessage(): void {
  selected.value = null
  loadMailboxes()
}

async function toggleStar(m: StarTarget): Promise<void> {
  try {
    await api.flags(m.uid, activeMailbox.value, { flagged: !m.flagged })
    m.flagged = !m.flagged
    if (selected.value && selected.value.uid === m.uid) {
      const flags = selected.value.flags ?? []
      const i = flags.indexOf('\\Flagged')
      if (m.flagged && i < 0) selected.value.flags = [...flags, '\\Flagged']
      if (!m.flagged && i >= 0) selected.value.flags = flags.filter((f) => f !== '\\Flagged')
    }
  } catch (e) {
    showToast('Lỗi: ' + errorMessage(e))
  }
}

function deleteSelected(): void {
  if (!selected.value) return
  confirmState.value = { type: 'message' }
}

async function doDelete(): Promise<void> {
  const cs = confirmState.value
  if (!cs || deleting.value) return
  deleting.value = true
  try {
    if (cs.type === 'message') {
      await api.flags(selected.value!.uid, activeMailbox.value, { deleted: true })
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
    showToast('Lỗi: ' + errorMessage(e))
  }
  deleting.value = false
}

async function markUnread(): Promise<void> {
  const m = selected.value
  if (!m) return
  try {
    await api.flags(m.uid, activeMailbox.value, { seen: false })
    selected.value = null
    invalidateLists()
    await loadMessages()
    loadMailboxes(true)
  } catch (e) {
    showToast('Lỗi: ' + errorMessage(e))
  }
}

async function runBatch({ action, uids }: BatchPayload): Promise<void> {
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
    }
    showToast('Đã cập nhật')
    invalidateLists()
    await loadMessages()
    loadMailboxes(true)
  } catch (e) {
    showToast('Lỗi: ' + errorMessage(e))
  }
}

function openCompose(): void {
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

function openReply(m: Message, all = false): void {
  const to = all
    ? [...(m.to ?? []).map(addrString), ...(m.cc ?? []).map(addrString)].filter(
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

// reply từ list row: chỉ có summary, cần tải thư đầy đủ trước
async function replyFromList(m: MessageSummary): Promise<void> {
  try {
    const full = await api.message(m.uid, activeMailbox.value)
    openReply(full)
  } catch (e) {
    showToast('Lỗi: ' + errorMessage(e))
  }
}

function openForward(m: Message): void {
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

function buildQuote(m: Message, forward = false): string {
  const from = m.from ? m.from.name || m.from.email : ''
  const header = forward
    ? `---------- Thư chuyển tiếp ----------\nTừ: ${from}\nNgày: ${fullDate(m.date)}\nTiêu đề: ${m.subject || ''}\n\n`
    : `Vào ${fullDate(m.date)}, ${from} đã viết:\n\n`
  const body = (m.text || '')
    .split('\n')
    .map((l) => (l ? '> ' + l : '>'))
    .join('\n')
  return '\n\n' + header + body
}

function onSent(): void {
  compose.value = null
  showToast('Đã gửi thư')
  if (activeMailbox.value.toLowerCase().includes('sent')) {
    invalidateLists()
    loadMessages()
  }
  loadMailboxes(true)
}

async function logout(): Promise<void> {
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
        <header class="flex h-12 shrink-0 items-center gap-2 border-b border-oc-border bg-oc-panel px-2.5 sm:px-3">
          <button
            class="flex h-8 w-8 shrink-0 items-center justify-center border border-oc-border text-oc-sub transition hover:border-oc-border-2 hover:text-oc-text sm:hidden"
            title="Menu"
            @click="mobileNav = true"
          >
            <MdiIcon :path="mdiMenu" :size="16" />
          </button>
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
              class="flex h-8 w-8 items-center justify-center border border-oc-border text-oc-sub transition hover:border-oc-border-2 hover:text-oc-text"
              :title="theme.isDark ? 'Chuyển sang chế độ sáng' : 'Chuyển sang chế độ tối'"
              @click="theme.toggle()"
            >
              <MdiIcon :path="theme.isDark ? mdiWeatherSunny : mdiWeatherNight" :size="16" />
            </button>
            <button
              class="flex h-8 w-8 items-center justify-center border border-oc-border text-oc-sub transition hover:border-oc-border-2 hover:text-oc-text"
              title="Đăng xuất"
              @click="logout"
            >
              <MdiIcon :path="mdiLogout" :size="16" />
            </button>
            <div
              class="hidden h-8 items-center gap-2 border border-oc-border bg-oc-bg px-2 sm:flex"
              :title="auth.username ?? undefined"
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
            @reply="replyFromList"
            @refresh="refresh"
            @batch="runBatch"
          />
        </div>
      </div>
    </div>

    <!-- drawer sidebar cho mobile -->
    <transition
      enter-active-class="transition-opacity duration-200"
      enter-from-class="opacity-0"
      leave-active-class="transition-opacity duration-150"
      leave-to-class="opacity-0"
    >
      <div
        v-if="mobileNav"
        class="fixed inset-0 z-55 bg-black/60 sm:hidden"
        @click="mobileNav = false"
      ></div>
    </transition>
    <transition
      enter-active-class="transition-transform duration-200 ease-out"
      enter-from-class="-translate-x-full"
      enter-to-class="translate-x-0"
      leave-active-class="transition-transform duration-150 ease-in"
      leave-from-class="translate-x-0"
      leave-to-class="-translate-x-full"
    >
      <Sidebar
        v-if="mobileNav"
        mobile
        class="sm:hidden"
        :mailboxes="mailboxes"
        :active="activeMailbox"
        @select="selectMailboxMobile"
        @compose="openComposeMobile"
        @toggle="mobileNav = false"
      />
    </transition>

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
        class="fixed bottom-10 left-4 z-50 flex items-center gap-2.5 border border-oc-border-2 border-l-2 border-l-oc-accent bg-oc-panel px-4 py-2.5 text-xs text-oc-text shadow-dialog"
      >
        {{ toast }}
      </div>
    </transition>
  </div>
</template>
