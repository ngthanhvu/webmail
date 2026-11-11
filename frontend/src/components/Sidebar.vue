<script setup>
import {
  mdiArchive,
  mdiFolder,
  mdiInbox,
  mdiPencil,
  mdiSend,
  mdiShieldAlert,
  mdiTrashCan,
} from '@mdi/js'
import MdiIcon from './MdiIcon.vue'
import { mailboxLabel } from '../format'

defineProps({
  mailboxes: { type: Array, default: () => [] },
  active: { type: String, default: 'INBOX' },
  collapsed: { type: Boolean, default: false },
})

const emit = defineEmits(['select', 'compose'])

function iconFor(name) {
  switch (mailboxLabel(name)) {
    case 'Hộp thư đến':
      return mdiInbox
    case 'Đã gửi':
      return mdiSend
    case 'Nháp':
      return mdiPencil
    case 'Thùng rác':
      return mdiTrashCan
    case 'Thư rác':
      return mdiShieldAlert
    case 'Lưu trữ':
      return mdiArchive
    default:
      return mdiFolder
  }
}
</script>

<template>
  <aside
    class="hidden shrink-0 flex-col bg-white py-2 sm:flex"
    :class="collapsed ? 'w-[72px] pl-0.5' : 'w-60 pl-2'"
  >
    <!-- compose -->
    <div class="mb-2 flex justify-center px-2 py-2">
      <button
        v-if="!collapsed"
        class="flex w-full items-center gap-3 rounded-2xl bg-white py-3.5 pl-5 pr-7 text-sm font-medium text-gmail-text shadow-md transition hover:shadow-lg active:shadow-sm"
        @click="emit('compose')"
      >
        <MdiIcon :path="mdiPencil" :size="22" class="text-gmail-blue" />
        Soạn thư
      </button>
      <button
        v-else
        class="flex h-14 w-14 items-center justify-center rounded-2xl bg-white text-gmail-blue shadow-md transition hover:shadow-lg active:shadow-sm"
        title="Soạn thư"
        @click="emit('compose')"
      >
        <MdiIcon :path="mdiPencil" :size="24" />
      </button>
    </div>

    <nav class="flex-1 overflow-y-auto" :class="collapsed ? 'px-1' : 'pr-2'">
      <button
        v-for="mb in mailboxes"
        :key="mb.name"
        class="flex w-full items-center text-sm transition"
        :class="[
          mb.name === active
            ? 'bg-gmail-active font-bold text-[#041e49]'
            : 'text-gmail-text hover:bg-gmail-hover',
          collapsed
            ? 'mx-auto h-12 w-14 justify-center rounded-full'
            : 'gap-4 rounded-r-full px-4 py-2.5',
        ]"
        :title="mailboxLabel(mb.name)"
        @click="emit('select', mb.name)"
      >
        <MdiIcon
          :path="iconFor(mb.name)"
          :size="collapsed ? 24 : 21"
          :class="mb.name === active ? 'text-gmail-blue' : 'text-gmail-sub'"
        />
        <template v-if="!collapsed">
          <span class="min-w-0 flex-1 truncate text-left">{{ mailboxLabel(mb.name) }}</span>
          <span
            v-if="mb.unseen > 0"
            class="shrink-0 text-right font-bold"
            :class="mb.name === active ? 'text-[#041e49]' : 'text-gmail-text'"
          >
            {{ mb.unseen > 99 ? '99+' : mb.unseen }}
          </span>
        </template>
      </button>
    </nav>
  </aside>
</template>
