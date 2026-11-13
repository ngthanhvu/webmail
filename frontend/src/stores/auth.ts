import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { api } from '@/api'

export const useAuth = defineStore('auth', () => {
  const username = ref<string | null>(null)
  const checked = ref(false)

  const isLoggedIn = computed(() => username.value !== null)

  async function check(): Promise<void> {
    try {
      const data = await api.me()
      username.value = data.username
    } catch {
      username.value = null
    }
    checked.value = true
  }

  async function login(user: string, password: string): Promise<void> {
    const data = await api.login(user, password)
    username.value = data.username
  }

  async function logout(): Promise<void> {
    try {
      await api.logout()
    } catch {
      /* ignore */
    }
    username.value = null
  }

  return { username, checked, isLoggedIn, check, login, logout }
})
