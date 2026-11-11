<script setup>
import { ref } from 'vue'
import { mdiClose, mdiPaperclip, mdiSend, mdiTrashCan } from '@mdi/js'
import MdiIcon from './MdiIcon.vue'
import { api } from '../api'
import { formatBytes, splitEmails } from '../format'

const props = defineProps({
  draft: { type: Object, required: true },
})

const emit = defineEmits(['close', 'sent'])

const to = ref((props.draft.to || []).join(', '))
const cc = ref((props.draft.cc || []).join(', '))
const bcc = ref((props.draft.bcc || []).join(', '))
const showCc = ref(!!props.draft.cc?.length)
const showBcc = ref(!!props.draft.bcc?.length)
const subject = ref(props.draft.subject || '')
const body = ref(props.draft.text || '')
const files = ref([])
const sending = ref(false)
const error = ref('')
const fileInput = ref(null)

function addFiles(e) {
  const list = Array.from(e.target.files || [])
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
  e.target.value = ''
}

function removeFile(i) {
  files.value.splice(i, 1)
}

async function send() {
  error.value = ''
  const payload = {
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
    error.value = e.message
  }
  sending.value = false
}
</script>

<template>
  <div class="fixed bottom-0 right-4 z-50 flex h-[520px] max-h-[calc(100vh-2rem)] w-[560px] max-w-[calc(100vw-1rem)] flex-col overflow-hidden rounded-t-lg border border-gmail-border bg-white shadow-2xl">
    <div class="flex h-10 shrink-0 items-center justify-between bg-gmail-hover px-4">
      <span class="text-xs font-medium text-gmail-text">
        {{ draft.inReplyTo ? 'Trả lời' : 'Thư mới' }}
      </span>
      <button
        class="rounded-full p-1.5 text-gmail-sub transition hover:bg-white/70"
        title="Đóng"
        @click="emit('close')"
      >
        <MdiIcon :path="mdiClose" :size="18" />
      </button>
    </div>

    <div class="flex min-h-0 flex-1 flex-col overflow-y-auto">
      <div class="flex shrink-0 items-center border-b border-gmail-hover px-4">
        <input
          v-model="to"
          type="text"
          placeholder="Người nhận"
          class="h-10 min-w-0 flex-1 bg-transparent text-sm outline-none placeholder:text-gmail-light"
        />
        <button v-if="!showCc" class="px-1.5 text-xs text-gmail-sub hover:text-gmail-blue" @click="showCc = true">Cc</button>
        <button v-if="!showBcc" class="px-1.5 text-xs text-gmail-sub hover:text-gmail-blue" @click="showBcc = true">Bcc</button>
      </div>
      <div v-if="showCc" class="flex shrink-0 items-center border-b border-gmail-hover px-4">
        <span class="w-8 text-xs text-gmail-sub">Cc</span>
        <input v-model="cc" type="text" class="h-10 min-w-0 flex-1 bg-transparent text-sm outline-none" />
      </div>
      <div v-if="showBcc" class="flex shrink-0 items-center border-b border-gmail-hover px-4">
        <span class="w-8 text-xs text-gmail-sub">Bcc</span>
        <input v-model="bcc" type="text" class="h-10 min-w-0 flex-1 bg-transparent text-sm outline-none" />
      </div>
      <div class="shrink-0 border-b border-gmail-hover px-4">
        <input
          v-model="subject"
          type="text"
          placeholder="Tiêu đề"
          class="h-10 w-full bg-transparent text-sm font-medium outline-none placeholder:font-normal placeholder:text-gmail-light"
        />
      </div>

      <textarea
        v-model="body"
        placeholder="Nội dung thư..."
        class="min-h-[180px] w-full flex-1 resize-none px-4 py-3 text-sm leading-6 outline-none placeholder:text-gmail-light"
      ></textarea>

      <div v-if="files.length" class="flex flex-wrap gap-2 px-4 pb-3">
        <div
          v-for="(f, i) in files"
          :key="i"
          class="flex items-center gap-2 rounded border border-gmail-border py-1.5 pl-2 pr-2 text-xs text-gmail-text"
        >
          <span class="max-w-40 truncate">{{ f.filename }}</span>
          <span class="text-gmail-light">{{ formatBytes(f.size) }}</span>
          <button class="text-gmail-light hover:text-gmail-text" title="Gỡ đính kèm" @click="removeFile(i)">✕</button>
        </div>
      </div>

      <p v-if="error" class="mx-4 mb-3 rounded bg-red-50 px-3 py-2 text-sm text-red-600">{{ error }}</p>
    </div>

    <div class="flex h-14 shrink-0 items-center gap-2 border-t border-gmail-hover px-4">
      <button
        class="mr-auto flex h-10 w-10 items-center justify-center rounded-full text-gmail-sub transition hover:bg-gmail-hover"
        title="Vứt bỏ thư"
        @click="emit('close')"
      >
        <MdiIcon :path="mdiTrashCan" :size="21" />
      </button>
      <button
        :disabled="sending"
        class="flex items-center gap-2 rounded-full bg-gmail-blue px-6 py-2.5 text-sm font-medium text-white shadow-sm transition hover:bg-gmail-blueDark disabled:cursor-not-allowed disabled:opacity-60"
        @click="send"
      >
        <span v-if="sending">Đang gửi...</span>
        <span v-else>Gửi</span>
        <MdiIcon v-if="!sending" :path="mdiSend" :size="18" />
      </button>
      <button
        class="flex h-10 w-10 items-center justify-center rounded-full text-gmail-sub transition hover:bg-gmail-hover"
        title="Đính kèm tệp"
        @click="fileInput && fileInput.click()"
      >
        <MdiIcon :path="mdiPaperclip" :size="21" />
      </button>
      <input ref="fileInput" type="file" multiple class="hidden" @change="addFiles" />
    </div>
  </div>
</template>
