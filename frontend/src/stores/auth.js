import { defineStore } from 'pinia'
import { api } from '../api'

export const useAuth = defineStore('auth', {
  state: () => ({
    username: null,
    checked: false,
  }),
  getters: {
    isLoggedIn: (s) => !!s.username,
  },
  actions: {
    async check() {
      try {
        const data = await api.me()
        this.username = data.username
      } catch {
        this.username = null
      }
      this.checked = true
    },
    async login(username, password) {
      const data = await api.login(username, password)
      this.username = data.username
    },
    async logout() {
      try {
        await api.logout()
      } catch {
        /* ignore */
      }
      this.username = null
    },
  },
})
