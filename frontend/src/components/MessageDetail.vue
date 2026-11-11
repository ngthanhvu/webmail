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
import { avatarColor, addrString, formatBytes, fullDate, initial } from '../format'

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
  <section class="flex min-w-0 flex-1 flex-col bg-white">
    <!-- header: back + subject -->
    <div class="flex h-14 shrink-0 items-center gap-3 px-2 sm:px-4">
      <button
        class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full text-gmail-sub transition hover:bg-gmail-hover"
        title="Quay lại danh sách"
        @click="emit('back')"
      >
        <MdiIcon :path="mdiArrowLeft" :size="22" />
      </button>
      <h1 class="min-w-0 flex-1 truncate text-xl font-normal text-gmail-text">
        {{ m.subject || '(không tiêu đề)' }}
      </h1>
    </div>

    <div class="flex-1 overflow-y-auto px-2 pb-8 sm:px-4">
      <div class="group rounded-lg border border-gmail-border bg-white shadow-sm">
        <!-- message header -->
        <div class="flex items-start gap-3 px-4 py-3 sm:px-6">
          <div
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full text-sm font-bold text-white"
            :style="{ backgroundColor: avatarColor(m.from?.email || 'x') }"
          >
            {{ initial(m.from?.name || m.from?.email) }}
          </div>

          <div class="min-w-0 flex-1 pt-1">
            <div class="flex flex-wrap items-baseline gap-x-1.5">
              <span class="truncate text-sm font-bold text-gmail-text">
                {{ m.from?.name || m.from?.email || '(không rõ)' }}
              </span>
              <span class="hidden truncate text-xs text-gmail-light sm:inline">
                &lt;{{ m.from?.email }}&gt;
              </span>
            </div>
            <div class="mt-1 truncate text-xs text-gmail-sub">
              Tới: {{ toText || '(không rõ)' }}
              <span v-if="ccText" class="ml-2">Cc: {{ ccText }}</span>
            </div>
          </div>

          <div class="flex shrink-0 items-center gap-1 pt-1">
            <span class="whitespace-nowrap text-xs text-gmail-sub" :title="fullDate(m.date)">
              {{ fullDate(m.date) }}
            </span>
            <div class="flex items-center gap-0.5 opacity-0 transition group-hover:opacity-100">
              <button
                class="flex h-8 w-8 items-center justify-center rounded-full transition hover:bg-gmail-hover"
                :class="isFlagged ? 'text-gmail-star' : 'text-gmail-sub'"
                title="Đánh dấu sao"
                @click="emit('star', m)"
              >
                <MdiIcon :path="isFlagged ? mdiStar : mdiStarOutline" :size="20" />
              </button>
              <button
                class="flex h-8 w-8 items-center justify-center rounded-full text-gmail-sub transition hover:bg-gmail-hover"
                title="Đánh dấu chưa đọc"
                @click="emit('unread')"
              >
                <MdiIcon :path="mdiEmail" :size="20" />
              </button>
              <button
                class="flex h-8 w-8 items-center justify-center rounded-full text-gmail-sub transition hover:bg-gmail-hover"
                title="Xóa thư"
                @click="emit('delete')"
              >
                <MdiIcon :path="mdiTrashCan" :size="20" />
              </button>
            </div>
          </div>
        </div>

        <!-- body -->
        <div class="px-4 pb-4 pt-1 sm:px-6">
          <iframe
            v-if="m.html"
            :srcdoc="m.html"
            sandbox=""
            class="h-[70vh] w-full border-0"
            title="Nội dung thư"
          ></iframe>
          <pre v-else-if="m.text" class="whitespace-pre-wrap break-words font-sans text-sm leading-6 text-gmail-text">{{ m.text }}</pre>
          <p v-else class="text-sm text-gmail-light">(Thư không có nội dung hiển thị)</p>
        </div>

        <!-- attachments -->
        <div v-if="(m.attachments || []).length" class="border-t border-gmail-hover px-4 py-3 sm:px-6">
          <p class="mb-2 text-xs font-medium text-gmail-sub">
            {{ m.attachments.length }} tệp đính kèm
          </p>
          <div class="flex flex-wrap gap-2">
            <a
              v-for="a in m.attachments"
              :key="a.section"
              :href="attachmentUrl(a)"
              class="flex items-center gap-2 rounded border border-gmail-border py-1.5 pl-2 pr-3 text-xs text-gmail-text transition hover:bg-gmail-hover"
            >
              <span
                class="flex h-6 w-6 items-center justify-center rounded-full text-[10px] font-bold uppercase text-white"
                style="background-color: #1a73e8"
              >
                {{ (a.filename || 'f').split('.').pop()?.slice(0, 3) || 'f' }}
              </span>
              <span class="max-w-48 truncate">{{ a.filename }}</span>
              <span class="text-gmail-light">{{ formatBytes(a.size) }}</span>
            </a>
          </div>
        </div>

        <!-- reply buttons -->
        <div class="flex gap-2 border-t border-gmail-hover px-4 py-3 sm:px-6">
          <button
            class="flex items-center gap-2 rounded-full border border-gmail-border px-5 py-1.5 text-sm text-gmail-sub transition hover:bg-gmail-hover"
            @click="emit('reply', m)"
          >
            <MdiIcon :path="mdiReply" :size="18" />
            Trả lời
          </button>
          <button
            v-if="(m.to || []).length + (m.cc || []).length > 1"
            class="flex items-center gap-2 rounded-full border border-gmail-border px-5 py-1.5 text-sm text-gmail-sub transition hover:bg-gmail-hover"
            @click="emit('reply-all', m)"
          >
            <MdiIcon :path="mdiReplyAll" :size="18" />
            Trả lời tất cả
          </button>
          <button
            class="flex items-center gap-2 rounded-full border border-gmail-border px-5 py-1.5 text-sm text-gmail-sub transition hover:bg-gmail-hover"
            @click="emit('forward', m)"
          >
            <MdiIcon :path="mdiShare" :size="18" />
            Chuyển tiếp
          </button>
        </div>
      </div>
    </div>
  </section>
</template>
