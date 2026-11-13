import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

export type ThemeName = 'light' | 'dark'

const KEY = 'webmail-theme'

function initialTheme(): ThemeName {
  try {
    const saved = localStorage.getItem(KEY)
    if (saved === 'dark' || saved === 'light') return saved
  } catch {
    /* ignore */
  }
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export const useTheme = defineStore('theme', () => {
  const current = ref<ThemeName>(initialTheme())
  const isDark = computed(() => current.value === 'dark')

  function apply(): void {
    document.documentElement.classList.toggle('dark', isDark.value)
    document
      .querySelector('meta[name="theme-color"]')
      ?.setAttribute('content', isDark.value ? '#0a0a0b' : '#f4f4f6')
  }

  function toggle(): void {
    current.value = isDark.value ? 'light' : 'dark'
    try {
      localStorage.setItem(KEY, current.value)
    } catch {
      /* ignore */
    }
    apply()
  }

  return { current, isDark, apply, toggle }
})
