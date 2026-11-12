<script setup>
import { computed } from 'vue'
import {
  mdiArrowLeft,
  mdiEmail,
  mdiReply,
  mdiReplyAll,
  mdiShare,
  mdiStar,
  mdiStarOutline,
  mdiTrashCan,
} from '@mdi/js'
import MdiIcon from './MdiIcon.vue'
import { api } from '../api'
import { addrString, formatBytes, fullDate, initial } from '../format'

const props = defineProps({
  message: { type: Object, required: true },
  mailbox: { type: String, default: 'INBOX' },
})

const emit = defineEmits(['back', 'reply', 'reply-all', 'forward', 'delete', 'unread', 'star'])

const m = computed(() => props.message)
const isFlagged = computed(() => (m.value.flags || []).includes('\\Flagged'))
const toText = computed(() => (m.value.to || []).map(addrString).join(', '))
const ccText = computed(() => (m.value.cc || []).map(addrString).join(', '))

function attachmentUrl(a) {
  return api.attachmentUrl(m.value.uid, props.mailbox, a.section, a.filename, a.contentType)
}
</script>

<template>
  <section class="flex min-w-0 flex-1 flex-col bg-oc-bg">
    <!-- header: back + subject + actions -->
    <div class="flex h-12 shrink-0 items-center gap-2 border-b border-oc-border bg-oc-panel px-3">
      <button
        class="flex h-8 w-8 shrink-0 items-center justify-center border border-oc-border text-oc-sub transition hover:border-oc-border2 hover:text-oc-text"
        title="Quay lại danh sách"
        @click="emit('back')"
      >
        <MdiIcon :path="mdiArrowLeft" :size="16" />
      </button>
      <h1 class="min-w-0 flex-1 truncate text-sm font-bold text-oc-text">
        {{ m.subject || '(không tiêu đề)' }}
      </h1>
      <MdiIcon v-if="isFlagged" :path="mdiStar" :size="14" class="shrink-0 text-oc-star" />
      <div class="flex shrink-0 items-center gap-1">
        <button
          class="flex h-7 w-7 items-center justify-center transition hover:bg-oc-hover"
          :class="isFlagged ? 'text-oc-star' : 'text-oc-faint hover:text-oc-text'"
          title="Đánh dấu sao"
          @click="emit('star', m)"
        >
          <MdiIcon :path="isFlagged ? mdiStar : mdiStarOutline" :size="16" />
        </button>
        <button
          class="flex h-7 w-7 items-center justify-center text-oc-faint transition hover:bg-oc-hover hover:text-oc-text"
          title="Đánh dấu chưa đọc"
          @click="emit('unread')"
        >
          <MdiIcon :path="mdiEmail" :size="16" />
        </button>
        <button
          class="flex h-7 w-7 items-center justify-center text-oc-faint transition hover:bg-oc-hover hover:text-oc-error"
          title="Xóa thư"
          @click="emit('delete')"
        >
          <MdiIcon :path="mdiTrashCan" :size="16" />
        </button>
      </div>
    </div>

    <div class="min-h-0 flex-1 overflow-y-auto p-4 sm:p-6">
      <article class="mx-auto max-w-3xl border border-oc-border bg-oc-panel">
        <!-- message header -->
        <header class="flex items-start gap-3 border-b border-oc-border px-4 py-4 sm:px-5">
          <span
            class="flex h-8 w-8 shrink-0 items-center justify-center rounded-[3px] border border-oc-border bg-oc-surface text-xs font-bold text-oc-sub"
          >
            {{ initial(m.from?.name || m.from?.email) }}
          </span>

          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-baseline gap-x-2">
              <span class="text-sm font-bold text-oc-text">
                {{ m.from?.name || m.from?.email || '(không rõ)' }}
              </span>
              <span class="hidden truncate text-xs text-oc-faint sm:inline">
                &lt;{{ m.from?.email }}&gt;
              </span>
            </div>
            <div class="mt-1.5 space-y-0.5 text-[11px] leading-4 text-oc-faint">
              <p class="truncate">
                <span class="font-bold uppercase tracking-wider">tới</span>
                {{ toText || '(không rõ)' }}
              </p>
              <p v-if="ccText" class="truncate">
                <span class="font-bold uppercase tracking-wider">cc</span>
                {{ ccText }}
              </p>
            </div>
          </div>

          <span
            class="shrink-0 whitespace-nowrap pt-0.5 text-[11px] text-oc-faint"
            :title="fullDate(m.date)"
          >
            {{ fullDate(m.date) }}
          </span>
        </header>

        <!-- body -->
        <div class="px-4 py-4 sm:px-5">
          <iframe
            v-if="m.html"
            :srcdoc="m.html"
            sandbox=""
            class="h-[70vh] w-full border-0 bg-white"
            title="Nội dung thư"
          ></iframe>
          <pre
            v-else-if="m.text"
            class="whitespace-pre-wrap break-words text-[13px] leading-6 text-oc-text"
            >{{ m.text }}</pre
          >
          <p v-else class="text-[13px] text-oc-faint">(Thư không có nội dung hiển thị)</p>
        </div>

        <!-- attachments -->
        <div v-if="(m.attachments || []).length" class="border-t border-oc-border px-4 py-3 sm:px-5">
          <p class="mb-2 text-[10px] font-bold uppercase tracking-[0.15em] text-oc-faint">
            {{ m.attachments.length }} tệp đính kèm
          </p>
          <div class="flex flex-wrap gap-2">
            <a
              v-for="a in m.attachments"
              :key="a.section"
              :href="attachmentUrl(a)"
              class="flex items-center gap-2 rounded-[3px] border border-oc-border bg-oc-surface px-2.5 py-1.5 text-xs text-oc-sub transition hover:border-oc-border2 hover:text-oc-text"
            >
              <span class="text-[10px] font-bold uppercase text-oc-accent">{{
                (a.filename || 'f').split('.').pop()?.slice(0, 4) || 'f'
              }}</span>
              <span class="max-w-48 truncate">{{ a.filename }}</span>
              <span class="text-oc-faint">{{ formatBytes(a.size) }}</span>
            </a>
          </div>
        </div>

        <!-- reply buttons -->
        <div class="flex flex-wrap gap-2 border-t border-oc-border px-4 py-3 sm:px-5">
          <button
            class="flex items-center gap-2 border border-oc-accent/40 px-3.5 py-1.5 text-xs font-bold text-oc-accent transition hover:bg-oc-accent hover:text-oc-bg"
            @click="emit('reply', m)"
          >
            <MdiIcon :path="mdiReply" :size="15" />
            Trả lời
          </button>
          <button
            v-if="(m.to || []).length + (m.cc || []).length > 1"
            class="flex items-center gap-2 border border-oc-border px-3.5 py-1.5 text-xs text-oc-sub transition hover:border-oc-border2 hover:text-oc-text"
            @click="emit('reply-all', m)"
          >
            <MdiIcon :path="mdiReplyAll" :size="15" />
            Trả lời tất cả
          </button>
          <button
            class="flex items-center gap-2 border border-oc-border px-3.5 py-1.5 text-xs text-oc-sub transition hover:border-oc-border2 hover:text-oc-text"
            @click="emit('forward', m)"
          >
            <MdiIcon :path="mdiShare" :size="15" />
            Chuyển tiếp
          </button>
        </div>
      </article>
    </div>
  </section>
</template>
