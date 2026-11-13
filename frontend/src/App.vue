<script setup lang="ts">
import { onBeforeUnmount, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '@/stores/auth'
import { useTheme } from '@/stores/theme'

const router = useRouter()
const auth = useAuth()
const theme = useTheme()

function onAuthRequired(): void {
  auth.username = null
  auth.checked = true
  if (router.currentRoute.value.path !== '/login') {
    router.push('/login')
  }
}

onMounted(() => {
  theme.apply()
  window.addEventListener('auth:required', onAuthRequired)
})
onBeforeUnmount(() => window.removeEventListener('auth:required', onAuthRequired))
</script>

<template>
  <router-view />
</template>
