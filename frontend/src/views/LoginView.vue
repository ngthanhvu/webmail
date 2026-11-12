<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { mdiAlertCircle, mdiArrowRight, mdiEmailOutline, mdiEye, mdiEyeOff, mdiLock } from '@mdi/js'
import MdiIcon from '../components/MdiIcon.vue'
import { useAuth } from '../stores/auth'

onMounted(() => {
  document.title = 'Webmail'
})

const router = useRouter()
const auth = useAuth()

const username = ref('')
const password = ref('')
const showPassword = ref(false)
const error = ref('')
const loading = ref(false)

async function submit() {
  if (!username.value || !password.value) {
    error.value = 'Vui lòng nhập tài khoản và mật khẩu'
    return
  }
  loading.value = true
  error.value = ''
  try {
    await auth.login(username.value.trim(), password.value)
    router.push('/')
  } catch (e) {
    error.value = e.message
  }
  loading.value = false
}
</script>

<template>
  <div class="relative flex min-h-screen items-center justify-center overflow-hidden bg-oc-bg p-4">
    <!-- ánh sáng cam mờ ở nền -->
    <div class="pointer-events-none absolute -top-40 left-1/4 h-96 w-96 rounded-full bg-oc-accent/10 blur-[120px]"></div>
    <div class="pointer-events-none absolute -bottom-40 right-1/4 h-96 w-96 rounded-full bg-oc-accent/5 blur-[120px]"></div>

    <div class="relative w-full max-w-sm border border-oc-border bg-oc-panel shadow-dialog">
      <!-- title bar kiểu terminal -->
      <div class="flex h-9 items-center gap-1.5 border-b border-oc-border bg-oc-surface px-3">
        <span class="h-2.5 w-2.5 rounded-full bg-oc-error/70"></span>
        <span class="h-2.5 w-2.5 rounded-full bg-oc-star/70"></span>
        <span class="h-2.5 w-2.5 rounded-full bg-oc-ok/70"></span>
        <span class="ml-2 truncate text-[11px] text-oc-faint">webmail — đăng nhập</span>
      </div>

      <div class="p-6">
        <!-- brand -->
        <div class="flex items-center gap-2.5">
          <div
            class="flex h-9 w-9 items-center justify-center bg-oc-accent text-sm font-extrabold leading-none text-oc-bg"
          >
            &gt;_
          </div>
          <div>
            <p class="text-base font-bold leading-none text-oc-text">webmail</p>
            <p class="mt-1.5 text-[10px] font-bold uppercase tracking-[0.18em] text-oc-faint">
              imap · smtp
            </p>
          </div>
        </div>

        <form class="mt-7 space-y-4" @submit.prevent="submit">
          <div>
            <label for="username" class="mb-1.5 block text-[10px] font-bold uppercase tracking-[0.15em] text-oc-faint">
              Tài khoản email
            </label>
            <div class="relative">
              <MdiIcon
                :path="mdiEmailOutline"
                :size="15"
                class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-oc-faint"
              />
              <input
                id="username"
                v-model="username"
                type="text"
                autocomplete="username"
                placeholder="ten@thanhvu.net"
                class="h-10 w-full border border-oc-border bg-oc-bg pl-9 pr-3 text-[13px] text-oc-text outline-none transition placeholder:text-oc-faint hover:border-oc-border2 focus:border-oc-accent"
              />
            </div>
          </div>

          <div>
            <label for="password" class="mb-1.5 block text-[10px] font-bold uppercase tracking-[0.15em] text-oc-faint">
              Mật khẩu
            </label>
            <div class="relative">
              <MdiIcon
                :path="mdiLock"
                :size="15"
                class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-oc-faint"
              />
              <input
                id="password"
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                autocomplete="current-password"
                placeholder="••••••••"
                class="h-10 w-full border border-oc-border bg-oc-bg pl-9 pr-11 text-[13px] text-oc-text outline-none transition placeholder:text-oc-faint hover:border-oc-border2 focus:border-oc-accent"
              />
              <button
                type="button"
                class="absolute right-2 top-1/2 flex h-7 w-7 -translate-y-1/2 items-center justify-center text-oc-faint transition hover:text-oc-sub"
                :title="showPassword ? 'Ẩn mật khẩu' : 'Hiện mật khẩu'"
                @click="showPassword = !showPassword"
              >
                <MdiIcon :path="showPassword ? mdiEyeOff : mdiEye" :size="16" />
              </button>
            </div>
          </div>

          <p
            v-if="error"
            class="flex items-start gap-2 border-l-2 border-oc-error bg-oc-error/10 px-3 py-2 text-xs text-oc-error"
          >
            <MdiIcon :path="mdiAlertCircle" :size="14" class="mt-0.5 shrink-0" />
            <span>{{ error }}</span>
          </p>

          <button
            type="submit"
            :disabled="loading"
            class="flex h-10 w-full items-center justify-center gap-2 bg-oc-accent text-[13px] font-bold text-oc-bg transition hover:bg-oc-accent2 disabled:cursor-not-allowed disabled:opacity-60"
          >
            <svg v-if="loading" class="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z" />
            </svg>
            <span v-else>Đăng nhập</span>
            <MdiIcon v-if="!loading" :path="mdiArrowRight" :size="15" />
          </button>
        </form>
      </div>

      <div class="border-t border-oc-border px-6 py-3 text-center text-[10px] text-oc-faint">
        Mật khẩu chỉ lưu trong phiên đăng nhập, mã hoá phía máy chủ
      </div>
    </div>
  </div>
</template>
