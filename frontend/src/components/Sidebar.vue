<script setup lang="ts">
import {
  mdiArchive,
  mdiClose,
  mdiFolder,
  mdiInbox,
  mdiMenu,
  mdiPencil,
  mdiSend,
  mdiShieldAlert,
  mdiTrashCan,
} from '@mdi/js'
import MdiIcon from './MdiIcon.vue'
import { mailboxLabel } from '@/utils/format'
import type { Mailbox } from '@/types'

withDefaults(
  defineProps<{
    mailboxes?: Mailbox[]
    active?: string
    collapsed?: boolean
    mobile?: boolean
  }>(),
  {
    mailboxes: () => [],
    active: 'INBOX',
    collapsed: false,
    mobile: false,
  }
)

const emit = defineEmits<{
  select: [name: string]
  compose: []
  toggle: []
}>()

function iconFor(name: string): string {
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
    class="shrink-0 flex-col border-r border-oc-border bg-oc-panel"
    :class="
      mobile
        ? 'fixed inset-y-0 left-0 z-60 flex w-64 shadow-dialog'
        : ['hidden sm:flex', collapsed ? 'sm:w-12' : 'sm:w-56']
    "
  >
    <!-- brand -->
    <div
      class="flex h-12 shrink-0 items-center gap-2.5 border-b border-oc-border px-2.5"
      :class="collapsed && !mobile ? 'justify-center' : ''"
    >
      <button
        class="flex h-7 w-7 items-center justify-center text-oc-faint transition hover:text-oc-text"
        :title="mobile ? 'Đóng menu' : collapsed ? 'Mở rộng menu' : 'Thu gọn menu'"
        @click="emit('toggle')"
      >
        <MdiIcon :path="mobile ? mdiClose : mdiMenu" :size="18" />
      </button>
      <template v-if="!collapsed || mobile">
        <div
          class="flex h-6 w-6 items-center justify-center bg-oc-accent text-[11px] font-extrabold leading-none text-oc-bg"
        >
          &gt;_
        </div>
        <span class="text-sm font-bold tracking-tight text-oc-text">webmail</span>
      </template>
    </div>

    <!-- compose -->
    <div class="flex shrink-0 p-2.5" :class="collapsed ? 'justify-center' : ''">
      <button
        v-if="!collapsed"
        class="flex w-full items-center justify-center gap-2 bg-oc-accent px-3 py-2 text-xs font-bold text-oc-bg transition hover:bg-oc-accent-2"
        @click="emit('compose')"
      >
        <MdiIcon :path="mdiPencil" :size="15" />
        Soạn thư
      </button>
      <button
        v-else
        class="flex h-8 w-8 items-center justify-center bg-oc-accent text-oc-bg transition hover:bg-oc-accent-2"
        title="Soạn thư"
        @click="emit('compose')"
      >
        <MdiIcon :path="mdiPencil" :size="15" />
      </button>
    </div>

    <!-- mailboxes -->
    <nav class="min-h-0 flex-1 overflow-y-auto py-1" :class="collapsed ? 'px-1.5' : 'px-2'">
      <p
        v-if="!collapsed"
        class="px-2.5 pb-1 pt-2 text-[10px] font-bold uppercase tracking-[0.18em] text-oc-faint"
      >
        Hộp thư
      </p>
      <button
        v-for="mb in mailboxes"
        :key="mb.name"
        class="group relative flex w-full items-center transition"
        :class="[
          mb.name === active
            ? 'bg-oc-hover text-oc-text'
            : 'text-oc-sub hover:bg-oc-hover/60 hover:text-oc-text',
          collapsed ? 'h-8 justify-center' : 'h-8 gap-2.5 rounded-[3px] px-2.5',
        ]"
        :title="mailboxLabel(mb.name)"
        @click="emit('select', mb.name)"
      >
        <span
          v-if="mb.name === active"
          class="absolute inset-y-0 left-0 w-0.5 bg-oc-accent"
          :class="collapsed ? 'my-1' : ''"
        ></span>
        <MdiIcon
          :path="iconFor(mb.name)"
          :size="16"
          :class="mb.name === active ? 'text-oc-accent' : 'text-oc-faint group-hover:text-oc-sub'"
        />
        <template v-if="!collapsed">
          <span class="min-w-0 flex-1 truncate text-left text-[13px]">{{ mailboxLabel(mb.name) }}</span>
          <span
            v-if="mb.unseen > 0"
            class="shrink-0 text-[11px] font-bold"
            :class="mb.name === active ? 'text-oc-accent' : 'text-oc-sub'"
          >
            {{ mb.unseen > 99 ? '99+' : mb.unseen }}
          </span>
        </template>
      </button>
    </nav>

    <!-- footer -->
    <div
      class="flex h-8 shrink-0 items-center border-t border-oc-border px-3 text-[10px] text-oc-faint"
      :class="collapsed ? 'justify-center' : 'justify-between'"
    >
      <template v-if="!collapsed">
        <span>webmail v0.1</span>
        <span class="flex items-center gap-1.5"><span class="text-[8px] text-oc-accent">●</span>imap</span>
      </template>
      <template v-else>
        <span class="text-[8px] text-oc-accent">●</span>
      </template>
    </div>
  </aside>
</template>
