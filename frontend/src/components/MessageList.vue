<script setup>
import { computed, ref, watch } from 'vue'
import {
  mdiChevronLeft,
  mdiChevronRight,
  mdiClose,
  mdiEmail,
  mdiEmailOpen,
  mdiEmailOutline,
  mdiRefresh,
  mdiReply,
  mdiStar,
  mdiStarOutline,
  mdiTrashCan,
} from '@mdi/js'
import MdiIcon from './MdiIcon.vue'
import { avatarColor, initial, shortDate, fullDate } from '../format'

const props = defineProps({
  messages: { type: Array, default: () => [] },
  total: { type: Number, default: 0 },
  page: { type: Number, default: 0 },
  limit: { type: Number, default: 50 },
  mailbox: { type: String, default: 'INBOX' },
  loading: { type: Boolean, default: false },
})

const emit = defineEmits(['select', 'star', 'page', 'reply', 'refresh', 'batch'])

const selected = ref([])

const allSelected = computed(
  () => props.messages.length > 0 && selected.value.length === props.messages.length
)

const rangeText = computed(() => {
  if (props.total === 0) return 'Không có thư'
  const from = props.page * props.limit + 1
  const to = Math.min(props.total, (props.page + 1) * props.limit)
  return `${from}–${to} / ${props.total}`
})

const pages = computed(() => Math.max(1, Math.ceil(props.total / props.limit)))

function toggleAll() {
  selected.value = allSelected.value ? [] : props.messages.map((m) => m.uid)
}

function toggleOne(uid) {
  const i = selected.value.indexOf(uid)
  if (i >= 0) selected.value.splice(i, 1)
  else selected.value.push(uid)
}

function isSelected(uid) {
  return selected.value.includes(uid)
}

watch(
  () => props.mailbox,
  () => (selected.value = [])
)
watch(
  () => props.messages,
  () => {
    selected.value = selected.value.filter((uid) => props.messages.some((m) => m.uid === uid))
  }
)

function batch(action) {
  if (!selected.value.length) return
  emit('batch', { action, uids: [...selected.value] })
}

function rowClass(m) {
  if (isSelected(m.uid)) return 'bg-gmail-selected'
  if (!m.seen) return 'bg-white hover:bg-gmail-rowhover'
  return 'hover:bg-gmail-rowhover'
}

function senderDisplay(m) {
  return m.from || m.fromEmail || '(không rõ)'
}
</script>

<template>
  <section class="flex min-w-0 flex-1 flex-col bg-gmail-list">
    <!-- toolbar -->
    <div class="flex h-12 shrink-0 items-center gap-1 px-2">
      <template v-if="selected.length === 0">
        <label class="flex h-10 w-10 cursor-pointer items-center justify-center rounded-full transition hover:bg-gmail-rowhover">
          <input type="checkbox" class="h-4 w-4 accent-gmail-blue" :checked="allSelected" @change="toggleAll" />
        </label>
        <button
          class="flex h-10 w-10 items-center justify-center rounded-full text-gmail-sub transition hover:bg-gmail-rowhover"
          title="Làm mới"
          @click="emit('refresh')"
        >
          <MdiIcon :path="mdiRefresh" :size="21" />
        </button>
      </template>
      <template v-else>
        <button
          class="flex h-10 w-10 items-center justify-center rounded-full text-gmail-sub transition hover:bg-gmail-rowhover"
          title="Bỏ chọn"
          @click="selected = []"
        >
          <MdiIcon :path="mdiClose" :size="21" />
        </button>
        <span class="ml-1 text-sm text-gmail-sub">Đã chọn {{ selected.length }}</span>
        <button
          class="ml-2 flex h-10 w-10 items-center justify-center rounded-full text-gmail-sub transition hover:bg-gmail-rowhover"
          title="Đánh dấu sao"
          @click="batch('star')"
        >
          <MdiIcon :path="mdiStarOutline" :size="21" />
        </button>
        <button
          class="flex h-10 w-10 items-center justify-center rounded-full text-gmail-sub transition hover:bg-gmail-rowhover"
          title="Đánh dấu chưa đọc"
          @click="batch('unread')"
        >
          <MdiIcon :path="mdiEmail" :size="21" />
        </button>
        <button
          class="flex h-10 w-10 items-center justify-center rounded-full text-gmail-sub transition hover:bg-gmail-rowhover"
          title="Đánh dấu đã đọc"
          @click="batch('read')"
        >
          <MdiIcon :path="mdiEmailOpen" :size="21" />
        </button>
        <button
          class="flex h-10 w-10 items-center justify-center rounded-full text-gmail-sub transition hover:bg-gmail-rowhover"
          title="Xóa"
          @click="batch('delete')"
        >
          <MdiIcon :path="mdiTrashCan" :size="21" />
        </button>
      </template>

      <div v-if="pages > 1" class="ml-auto flex items-center gap-1 pr-2">
        <button
          class="rounded-full p-1.5 text-gmail-sub transition hover:bg-gmail-rowhover disabled:cursor-default disabled:opacity-30"
          :disabled="page === 0"
          title="Trang trước"
          @click="emit('page', page - 1)"
        >
          <MdiIcon :path="mdiChevronLeft" :size="22" />
        </button>
        <span class="whitespace-nowrap text-xs text-gmail-sub">{{ rangeText }}</span>
        <button
          class="rounded-full p-1.5 text-gmail-sub transition hover:bg-gmail-rowhover disabled:cursor-default disabled:opacity-30"
          :disabled="(page + 1) * limit >= total"
          title="Trang sau"
          @click="emit('page', page + 1)"
        >
          <MdiIcon :path="mdiChevronRight" :size="22" />
        </button>
      </div>
      <span v-else class="ml-auto pr-4 text-xs text-gmail-sub">{{ rangeText }}</span>
    </div>

    <!-- message rows -->
    <div class="flex-1 overflow-y-auto" :class="loading ? '' : 'border-t border-gmail-border'">
      <div v-if="loading" class="flex flex-col items-center gap-2 p-10 text-gmail-light">
        <svg class="h-6 w-6 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z" />
        </svg>
        <span class="text-sm">Đang tải...</span>
      </div>

      <template v-else>
        <div
          v-for="m in messages"
          :key="m.uid"
          class="group flex h-12 cursor-pointer items-center gap-2 border-b border-gmail-rowhover pl-2 pr-4 text-sm sm:gap-3"
          :class="rowClass(m)"
          @click="emit('select', m)"
        >
          <input
            type="checkbox"
            class="h-4 w-4 shrink-0 accent-gmail-blue opacity-0 transition group-hover:opacity-100"
            :class="{ '!opacity-100': isSelected(m.uid) }"
            :checked="isSelected(m.uid)"
            @click.stop="toggleOne(m.uid)"
          />

          <button
            class="shrink-0 rounded-full p-1.5 transition hover:scale-110"
            :title="m.flagged ? 'Bỏ đánh dấu sao' : 'Đánh dấu sao'"
            @click.stop="emit('star', m)"
          >
            <MdiIcon
              :path="m.flagged ? mdiStar : mdiStarOutline"
              :size="21"
              :class="m.flagged ? 'text-gmail-star' : 'text-gmail-light hover:text-gmail-star'"
            />
          </button>

          <div
            class="flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-full text-[13px] font-bold text-white"
            :style="{ backgroundColor: avatarColor(m.fromEmail || m.from || 'x') }"
          >
            {{ initial(senderDisplay(m)) }}
          </div>

          <div class="w-40 shrink-0 truncate sm:w-52" :class="!m.seen ? 'font-bold text-gmail-text' : 'font-medium text-gmail-sub'">
            {{ senderDisplay(m) }}
          </div>

          <div class="min-w-0 flex-1 truncate" :class="!m.seen ? 'font-bold text-gmail-text' : 'text-gmail-sub'">
            <span class="font-medium">{{ m.subject || '(không tiêu đề)' }}</span>
            <span class="ml-2 font-normal text-gmail-light" :class="!m.seen ? 'font-bold' : ''">- {{ m.snippet }}</span>
          </div>

          <div class="hidden shrink-0 items-center group-hover:flex">
            <button
              class="flex h-8 w-8 items-center justify-center rounded-full text-gmail-sub transition hover:bg-white/60"
              title="Trả lời"
              @click.stop="emit('reply', m)"
            >
              <MdiIcon :path="mdiReply" :size="20" />
            </button>
          </div>

          <span
            class="w-16 shrink-0 text-right text-xs text-gmail-sub group-hover:hidden"
            :title="fullDate(m.date)"
          >
            {{ shortDate(m.date) }}
          </span>
        </div>

        <div v-if="messages.length === 0" class="flex flex-col items-center gap-2 p-12 text-gmail-light">
          <MdiIcon :path="mdiEmailOutline" :size="40" />
          <span class="text-sm">Không có thư trong hộp thư này</span>
        </div>
      </template>
    </div>
  </section>
</template>
