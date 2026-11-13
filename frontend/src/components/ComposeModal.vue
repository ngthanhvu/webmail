<script setup lang="ts">
import { ref, computed } from 'vue'
import { mdiClose, mdiPaperclip, mdiSend, mdiTrashCan } from '@mdi/js'
import MdiIcon from './MdiIcon.vue'
import ConfirmDialog from './ConfirmDialog.vue'
import { api } from '@/api'
import { formatBytes, splitEmails } from '@/utils/format'
import type { Draft, LocalAttachment, SendPayload } from '@/types'

const props = defineProps<{
  draft: Draft
}>()

const emit = defineEmits<{ close: []; sent: [] }>()

const to = ref((props.draft.to ?? []).join(', '))
const cc = ref((props.draft.cc ?? []).join(', '))
const bcc = ref((props.draft.bcc ?? []).join(', '))
const showCc = ref(!!props.draft.cc?.length)
const showBcc = ref(!!props.draft.bcc?.length)
const subject = ref(props.draft.subject || '')
const body = ref(props.draft.text || '')
const files = ref<LocalAttachment[]>([])
const sending = ref(false)
const error = ref('')
const fileInput = ref<HTMLInputElement | null>(null)
const confirmDiscard = ref(false)

const hasContent = computed(
  () =>
    to.value.trim() !== '' ||
    cc.value.trim() !== '' ||
    bcc.value.trim() !== '' ||
    subject.value.trim() !== '' ||
    body.value.trim() !== '' ||
    files.value.length > 0
)

function requestClose(): void {
  if (hasContent.value) {
    confirmDiscard.value = true
  } else {
    emit('close')
  }
}

function addFiles(e: Event): void {
  const input = e.target as HTMLInputElement
  const list = Array.from(input.files ?? [])
  list.forEach((f) => {
    const reader = new FileReader()
    reader.onload = () => {
      files.value.push({
        filename: f.name,
        contentType: f.type || 'application/octet-stream',
        contentBase64: String(reader.result).split(',')[1] || '',
        size: f.size,
      })
    }
    reader.readAsDataURL(f)
  })
  input.value = ''
}

function removeFile(i: number): void {
  files.value.splice(i, 1)
}

async function send(): Promise<void> {
  error.value = ''
  const payload: SendPayload = {
    to: splitEmails(to.value),
    cc: showCc.value ? splitEmails(cc.value) : [],
    bcc: showBcc.value ? splitEmails(bcc.value) : [],
    subject: subject.value,
    text: body.value,
    html: '',
    inReplyTo: props.draft.inReplyTo || '',
    references: props.draft.references || '',
    attachments: files.value.map((f) => ({
      filename: f.filename,
      contentType: f.contentType,
      contentBase64: f.contentBase64,
    })),
  }
  if (!payload.to.length && !payload.cc.length && !payload.bcc.length) {
    error.value = 'Vui lòng nhập ít nhất một người nhận'
    return
  }
  sending.value = true
  try {
    await api.send(payload)
    emit('sent')
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
  sending.value = false
}
</script>

<template>
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
    @click.self="requestClose()"
  >
    <div
      class="flex h-[560px] max-h-full w-[560px] max-w-full flex-col border border-oc-border-2 bg-oc-panel shadow-dialog"
    >
      <!-- title bar -->
      <div class="flex h-10 shrink-0 items-center justify-between border-b border-oc-border bg-oc-surface px-3">
        <span class="flex items-center gap-2 text-xs font-bold uppercase tracking-[0.12em] text-oc-text">
          <span class="text-oc-accent">▸</span>
          {{ draft.inReplyTo ? 'Trả lời thư' : 'Thư mới' }}
        </span>
        <button
          class="flex h-6 w-6 items-center justify-center text-oc-faint transition hover:text-oc-error"
          title="Đóng"
          @click="requestClose()"
        >
          <MdiIcon :path="mdiClose" :size="16" />
        </button>
      </div>

      <div class="flex min-h-0 flex-1 flex-col overflow-y-auto">
        <div
          class="flex shrink-0 items-center border-b border-oc-border px-3 transition focus-within:border-oc-accent/60"
        >
          <span class="w-10 shrink-0 text-[10px] font-bold uppercase tracking-wider text-oc-faint">Tới</span>
          <input
            v-model="to"
            type="text"
            placeholder="nguoi_nhan@vi-du.com"
            class="h-10 min-w-0 flex-1 bg-transparent text-[13px] text-oc-text outline-none placeholder:text-oc-faint"
          />
          <button
            v-if="!showCc"
            class="px-1.5 text-[10px] font-bold uppercase tracking-wider text-oc-faint transition hover:text-oc-accent"
            @click="showCc = true"
          >
            Cc
          </button>
          <button
            v-if="!showBcc"
            class="px-1.5 text-[10px] font-bold uppercase tracking-wider text-oc-faint transition hover:text-oc-accent"
            @click="showBcc = true"
          >
            Bcc
          </button>
        </div>
        <div
          v-if="showCc"
          class="flex shrink-0 items-center border-b border-oc-border px-3 transition focus-within:border-oc-accent/60"
        >
          <span class="w-10 shrink-0 text-[10px] font-bold uppercase tracking-wider text-oc-faint">Cc</span>
          <input
            v-model="cc"
            type="text"
            class="h-10 min-w-0 flex-1 bg-transparent text-[13px] text-oc-text outline-none placeholder:text-oc-faint"
          />
        </div>
        <div
          v-if="showBcc"
          class="flex shrink-0 items-center border-b border-oc-border px-3 transition focus-within:border-oc-accent/60"
        >
          <span class="w-10 shrink-0 text-[10px] font-bold uppercase tracking-wider text-oc-faint">Bcc</span>
          <input
            v-model="bcc"
            type="text"
            class="h-10 min-w-0 flex-1 bg-transparent text-[13px] text-oc-text outline-none placeholder:text-oc-faint"
          />
        </div>
        <div
          class="shrink-0 border-b border-oc-border px-3 transition focus-within:border-oc-accent/60"
        >
          <input
            v-model="subject"
            type="text"
            placeholder="Tiêu đề"
            class="h-10 w-full bg-transparent text-[13px] font-bold text-oc-text outline-none placeholder:font-normal placeholder:text-oc-faint"
          />
        </div>

        <textarea
          v-model="body"
          placeholder="Nội dung thư..."
          class="min-h-[180px] w-full flex-1 resize-none bg-oc-bg px-3 py-3 text-[13px] leading-6 text-oc-text outline-none placeholder:text-oc-faint"
        ></textarea>

        <div v-if="files.length" class="flex flex-wrap gap-2 px-3 pb-3">
          <div
            v-for="(f, i) in files"
            :key="i"
            class="flex items-center gap-2 rounded-[3px] border border-oc-border bg-oc-surface px-2.5 py-1.5 text-xs text-oc-sub"
          >
            <span class="max-w-40 truncate">{{ f.filename }}</span>
            <span class="text-oc-faint">{{ formatBytes(f.size) }}</span>
            <button
              class="text-oc-faint transition hover:text-oc-error"
              title="Gỡ đính kèm"
              @click="removeFile(i)"
            >
              ✕
            </button>
          </div>
        </div>

        <p
          v-if="error"
          class="mx-3 mb-3 border-l-2 border-oc-error bg-oc-error/10 px-3 py-2 text-xs text-oc-error"
        >
          {{ error }}
        </p>
      </div>

      <!-- footer -->
      <div class="flex h-12 shrink-0 items-center gap-2 border-t border-oc-border px-3">
        <button
          class="flex h-8 w-8 items-center justify-center text-oc-faint transition hover:text-oc-error"
          title="Vứt bỏ thư"
          @click="requestClose()"
        >
          <MdiIcon :path="mdiTrashCan" :size="17" />
        </button>
        <button
          class="flex h-8 w-8 items-center justify-center text-oc-faint transition hover:text-oc-text"
          title="Đính kèm tệp"
          @click="fileInput && fileInput.click()"
        >
          <MdiIcon :path="mdiPaperclip" :size="17" />
        </button>
        <input ref="fileInput" type="file" multiple class="hidden" @change="addFiles" />
        <button
          :disabled="sending"
          class="ml-auto flex items-center gap-2 bg-oc-accent px-4 py-1.5 text-xs font-bold text-oc-bg transition hover:bg-oc-accent-2 disabled:cursor-not-allowed disabled:opacity-60"
          @click="send"
        >
          <span v-if="sending">Đang gửi...</span>
          <span v-else>Gửi</span>
          <MdiIcon v-if="!sending" :path="mdiSend" :size="14" />
        </button>
      </div>

      <ConfirmDialog
        :open="confirmDiscard"
        title="Vứt bỏ thư đang soạn?"
        message="Nội dung bạn đã nhập sẽ bị mất và không thể khôi phục."
        confirm-text="Vứt bỏ"
        @confirm="emit('close')"
        @cancel="confirmDiscard = false"
      />
    </div>
  </div>
</template>
