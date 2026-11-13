<script setup lang="ts">
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
import { initial, shortDate, fullDate, mailboxLabel } from '@/utils/format'
import type { BatchAction, BatchPayload, MessageSummary } from '@/types'

const props = withDefaults(
  defineProps<{
    messages?: MessageSummary[]
    total?: number
    page?: number
    limit?: number
    mailbox?: string
    loading?: boolean
  }>(),
  {
    messages: () => [],
    total: 0,
    page: 0,
    limit: 50,
    mailbox: 'INBOX',
    loading: false,
  }
)

const emit = defineEmits<{
  select: [m: MessageSummary]
  star: [m: MessageSummary]
  page: [p: number]
  reply: [m: MessageSummary]
  refresh: []
  batch: [payload: BatchPayload]
}>()

const selected = ref<number[]>([])

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

function toggleAll(): void {
  selected.value = allSelected.value ? [] : props.messages.map((m) => m.uid)
}

function toggleOne(uid: number): void {
  const i = selected.value.indexOf(uid)
  if (i >= 0) selected.value.splice(i, 1)
  else selected.value.push(uid)
}

function isSelected(uid: number): boolean {
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

function batch(action: BatchAction): void {
  if (!selected.value.length) return
  emit('batch', { action, uids: [...selected.value] })
}

function rowClass(m: MessageSummary): string {
  if (isSelected(m.uid)) return 'bg-oc-accent-dim'
  return 'hover:bg-oc-hover/60'
}

function senderDisplay(m: MessageSummary): string {
  return m.from || m.fromEmail || '(không rõ)'
}
</script>

<template>
  <section class="flex min-w-0 flex-1 flex-col bg-oc-bg">
    <!-- toolbar -->
    <div class="flex h-10 shrink-0 items-center gap-1 border-b border-oc-border bg-oc-panel px-2.5">
      <template v-if="selected.length === 0">
        <label
          class="flex h-7 w-7 cursor-pointer items-center justify-center transition hover:bg-oc-hover"
        >
          <input type="checkbox" class="h-3.5 w-3.5" :checked="allSelected" @change="toggleAll" />
        </label>
        <button
          class="flex h-7 w-7 items-center justify-center text-oc-faint transition hover:bg-oc-hover hover:text-oc-text"
          title="Làm mới"
          @click="emit('refresh')"
        >
          <MdiIcon :path="mdiRefresh" :size="16" />
        </button>
        <span class="ml-1.5 truncate text-[11px] font-bold uppercase tracking-[0.15em] text-oc-faint">
          {{ mailboxLabel(mailbox) }}
        </span>
      </template>
      <template v-else>
        <button
          class="flex h-7 w-7 items-center justify-center text-oc-faint transition hover:bg-oc-hover hover:text-oc-text"
          title="Bỏ chọn"
          @click="selected = []"
        >
          <MdiIcon :path="mdiClose" :size="16" />
        </button>
        <span class="ml-1 text-xs text-oc-sub">Đã chọn {{ selected.length }}</span>
        <button
          class="flex h-7 w-7 items-center justify-center text-oc-faint transition hover:bg-oc-hover hover:text-oc-text"
          title="Đánh dấu sao"
          @click="batch('star')"
        >
          <MdiIcon :path="mdiStarOutline" :size="16" />
        </button>
        <button
          class="flex h-7 w-7 items-center justify-center text-oc-faint transition hover:bg-oc-hover hover:text-oc-text"
          title="Đánh dấu chưa đọc"
          @click="batch('unread')"
        >
          <MdiIcon :path="mdiEmail" :size="16" />
        </button>
        <button
          class="flex h-7 w-7 items-center justify-center text-oc-faint transition hover:bg-oc-hover hover:text-oc-text"
          title="Đánh dấu đã đọc"
          @click="batch('read')"
        >
          <MdiIcon :path="mdiEmailOpen" :size="16" />
        </button>
        <button
          class="flex h-7 w-7 items-center justify-center text-oc-faint transition hover:bg-oc-hover hover:text-oc-error"
          title="Xóa"
          @click="batch('delete')"
        >
          <MdiIcon :path="mdiTrashCan" :size="16" />
        </button>
      </template>

      <div class="ml-auto flex items-center gap-1">
        <span class="mr-1 hidden whitespace-nowrap text-[11px] text-oc-faint sm:inline">{{
          rangeText
        }}</span>
        <template v-if="pages > 1">
          <button
            class="flex h-6 w-6 items-center justify-center text-oc-faint transition hover:text-oc-text disabled:cursor-default disabled:opacity-30"
            :disabled="page === 0"
            title="Trang trước"
            @click="emit('page', page - 1)"
          >
            <MdiIcon :path="mdiChevronLeft" :size="16" />
          </button>
          <button
            class="flex h-6 w-6 items-center justify-center text-oc-faint transition hover:text-oc-text disabled:cursor-default disabled:opacity-30"
            :disabled="(page + 1) * limit >= total"
            title="Trang sau"
            @click="emit('page', page + 1)"
          >
            <MdiIcon :path="mdiChevronRight" :size="16" />
          </button>
        </template>
      </div>
    </div>

    <!-- message rows -->
    <div class="flex-1 overflow-y-auto">
      <div v-if="loading" class="flex flex-col items-center gap-2.5 p-10 text-oc-faint">
        <svg class="h-5 w-5 animate-spin text-oc-accent" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z" />
        </svg>
        <span class="text-xs">Đang tải...</span>
      </div>

      <template v-else>
        <div
          v-for="m in messages"
          :key="m.uid"
          class="group flex h-10 cursor-pointer items-center gap-2.5 border-b border-oc-border/50 px-3 text-[13px] transition-colors"
          :class="rowClass(m)"
          @click="emit('select', m)"
        >
          <input
            type="checkbox"
            class="h-3.5 w-3.5 shrink-0 opacity-0 transition group-hover:opacity-100"
            :class="{ 'opacity-100!': isSelected(m.uid) }"
            :checked="isSelected(m.uid)"
            @click.stop="toggleOne(m.uid)"
          />

          <button
            class="hidden h-5 w-5 shrink-0 items-center justify-center transition sm:flex"
            :title="m.flagged ? 'Bỏ đánh dấu sao' : 'Đánh dấu sao'"
            @click.stop="emit('star', m)"
          >
            <MdiIcon
              :path="m.flagged ? mdiStar : mdiStarOutline"
              :size="15"
              :class="
                m.flagged
                  ? 'text-oc-star'
                  : 'text-oc-faint opacity-0 transition group-hover:opacity-100'
              "
            />
          </button>

          <span
            class="hidden h-6 w-6 shrink-0 items-center justify-center rounded-[3px] border border-oc-border bg-oc-surface text-[10px] font-bold text-oc-sub sm:flex"
          >
            {{ initial(senderDisplay(m)) }}
          </span>

          <div class="flex w-28 shrink-0 items-center gap-1.5 sm:w-48">
            <span
              class="w-1.5 shrink-0 text-center text-[8px] leading-none"
              :class="m.seen ? 'text-transparent' : 'text-oc-accent'"
              >●</span
            >
            <span
              class="min-w-0 flex-1 truncate"
              :class="m.seen ? 'text-oc-sub' : 'font-bold text-oc-text'"
            >
              {{ senderDisplay(m) }}
            </span>
          </div>

          <div class="min-w-0 flex-1 truncate text-oc-sub">
            <span :class="m.seen ? '' : 'text-oc-text'">{{ m.subject || '(không tiêu đề)' }}</span>
            <span class="ml-2 text-oc-faint">— {{ m.snippet }}</span>
          </div>

          <button
            class="hidden h-6 w-6 shrink-0 items-center justify-center text-oc-faint transition hover:text-oc-accent group-hover:flex"
            title="Trả lời"
            @click.stop="emit('reply', m)"
          >
            <MdiIcon :path="mdiReply" :size="15" />
          </button>

          <span
            class="w-20 shrink-0 text-right text-[11px] text-oc-faint group-hover:hidden"
            :title="fullDate(m.date)"
          >
            {{ shortDate(m.date) }}
          </span>
        </div>

        <div v-if="messages.length === 0" class="flex flex-col items-center gap-3 p-12 text-oc-faint">
          <MdiIcon :path="mdiEmailOutline" :size="32" />
          <span class="text-xs">Không có thư trong hộp thư này</span>
        </div>
      </template>
    </div>
  </section>
</template>
