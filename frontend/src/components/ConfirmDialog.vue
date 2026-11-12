<script setup>
defineProps({
  open: { type: Boolean, default: false },
  title: { type: String, default: 'Xác nhận' },
  message: { type: String, default: '' },
  confirmText: { type: String, default: 'Xóa' },
  cancelText: { type: String, default: 'Hủy' },
  loading: { type: Boolean, default: false },
})

const emit = defineEmits(['confirm', 'cancel'])
</script>

<template>
  <teleport to="body">
    <transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="opacity-0"
      leave-active-class="transition duration-100 ease-in"
      leave-to-class="opacity-0"
    >
      <div
        v-if="open"
        class="fixed inset-0 z-[70] flex items-center justify-center bg-black/60 px-4"
        @click.self="emit('cancel')"
      >
        <transition
          appear
          enter-active-class="transition duration-150 ease-out"
          enter-from-class="scale-95 opacity-0"
          leave-active-class="transition duration-100 ease-in"
          leave-to-class="scale-95 opacity-0"
        >
          <div
            v-if="open"
            class="w-full max-w-sm border border-oc-border2 bg-oc-panel p-5 shadow-dialog"
          >
            <h3 class="flex items-center gap-2 text-sm font-bold text-oc-text">
              <span class="text-oc-accent">!</span>
              {{ title }}
            </h3>
            <p class="mt-2 text-[13px] leading-5 text-oc-sub">{{ message }}</p>
            <div class="mt-5 flex justify-end gap-2">
              <button
                class="border border-oc-border px-3.5 py-1.5 text-xs text-oc-sub transition hover:border-oc-border2 hover:text-oc-text"
                @click="emit('cancel')"
              >
                {{ cancelText }}
              </button>
              <button
                :disabled="loading"
                class="flex items-center gap-2 bg-oc-accent px-4 py-1.5 text-xs font-bold text-oc-bg transition hover:bg-oc-accent2 disabled:cursor-not-allowed disabled:opacity-60"
                @click="emit('confirm')"
              >
                <svg v-if="loading" class="h-3.5 w-3.5 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z" />
                </svg>
                {{ loading ? 'Đang xử lý...' : confirmText }}
              </button>
            </div>
          </div>
        </transition>
      </div>
    </transition>
  </teleport>
</template>
