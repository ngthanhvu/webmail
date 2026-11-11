<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { mdiEmailOutline } from '@mdi/js'
import MdiIcon from '../components/MdiIcon.vue'
import { useAuth } from '../stores/auth'

const router = useRouter()
const auth = useAuth()

const username = ref('')
const password = ref('')
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
  <div class="flex min-h-screen flex-col items-center justify-center bg-white px-4">
    <div class="w-full max-w-[384px]">
      <div class="mb-8 flex flex-col items-center">
        <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gmail-blue text-white shadow-md">
          <MdiIcon :path="mdiEmailOutline" :size="28" />
        </div>
        <h1 class="mt-5 text-2xl font-normal text-gmail-text">Webmail</h1>
        <p class="mt-1.5 text-sm text-gmail-sub">Đăng nhập bằng tài khoản email của bạn</p>
      </div>

      <form class="space-y-5" @submit.prevent="submit">
        <div>
          <input
            v-model="username"
            type="text"
            autocomplete="username"
            placeholder="Tài khoản email"
            class="w-full rounded-lg border border-gmail-border px-4 py-3.5 text-sm text-gmail-text outline-none transition focus:border-gmail-blue focus:ring-1 focus:ring-gmail-blue"
          />
        </div>
        <div>
          <input
            v-model="password"
            type="password"
            autocomplete="current-password"
            placeholder="Mật khẩu"
            class="w-full rounded-lg border border-gmail-border px-4 py-3.5 text-sm text-gmail-text outline-none transition focus:border-gmail-blue focus:ring-1 focus:ring-gmail-blue"
          />
        </div>

        <p v-if="error" class="text-sm text-red-600">{{ error }}</p>

        <button
          type="submit"
          :disabled="loading"
          class="w-full rounded-full bg-gmail-blue py-3 text-sm font-medium text-white shadow-sm transition hover:bg-gmail-blueDark disabled:cursor-not-allowed disabled:opacity-60"
        >
          <span v-if="loading">Đang đăng nhập...</span>
          <span v-else>Đăng nhập</span>
        </button>
      </form>

      <p class="mt-10 text-center text-xs text-gmail-light">
        Hộp thư của bạn được đọc trực tiếp qua IMAP/SMTP
      </p>
    </div>
  </div>
</template>
