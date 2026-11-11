<script setup>
import {
  mdiArchive,
  mdiEmailOutline,
  mdiFolder,
  mdiInbox,
  mdiMenu,
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

const emit = defineEmits(['select', 'compose', 'toggle'])

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
    class="hidden shrink-0 flex-col bg-white sm:flex"
    :class="collapsed ? 'w-16' : 'w-60 pl-2'"
  >
    <!-- logo + nút thu gọn -->
    <div
      class="flex h-16 shrink-0 items-center"
      :class="collapsed ? 'justify-center' : 'gap-1'"
    >
      <button
        class="flex h-10 w-10 items-center justify-center rounded-full text-gmail-sub transition hover:bg-gmail-hover"
        :title="collapsed ? 'Mở rộng menu' : 'Thu gọn menu'"
        @click="emit('toggle')"
      >
        <MdiIcon :path="mdiMenu" :size="23" />
      </button>
      <template v-if="!collapsed">
        <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-gmail-blue text-white shadow-sm">
          <MdiIcon :path="mdiEmailOutline" :size="20" />
        </div>
        <span class="text-[20px] font-normal text-gmail-sub">Mail</span>
      </template>
    </div>

    <!-- compose -->
    <div class="flex py-1" :class="collapsed ? 'justify-center' : ''">
      <button
        v-if="!collapsed"
        class="flex items-center gap-2.5 rounded-full bg-white py-2.5 pl-4 pr-6 text-sm font-medium text-gmail-text border transition hover:shadow-lg"
        @click="emit('compose')"
      >
        <MdiIcon :path="mdiPencil" :size="20" class="text-gmail-blue" />
        Soạn thư
      </button>
      <button
        v-else
        class="flex h-10 w-10 items-center justify-center rounded-full bg-white text-gmail-blue border transition hover:shadow-lg"
        title="Soạn thư"
        @click="emit('compose')"
      >
        <MdiIcon :path="mdiPencil" :size="20" />
      </button>
    </div>

    <nav class="flex-1 overflow-y-auto border-r" :class="collapsed ? 'px-1' : 'pr-2'">
      <button
        v-for="mb in mailboxes"
        :key="mb.name"
        class="flex w-full items-center text-sm transition"
        :class="[
          mb.name === active
            ? 'bg-gmail-active font-bold text-[#041e49]'
            : 'text-gmail-text hover:bg-gmail-hover',
          collapsed
            ? 'mx-auto h-10 w-10 justify-center rounded-full'
            : 'gap-4 rounded-r-full px-4 py-2.5',
        ]"
        :title="mailboxLabel(mb.name)"
        @click="emit('select', mb.name)"
      >
        <MdiIcon
          :path="iconFor(mb.name)"
          :size="collapsed ? 20 : 21"
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
