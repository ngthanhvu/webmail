<script setup>
import { onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from './stores/auth'

const router = useRouter()
const auth = useAuth()

function onAuthRequired() {
  auth.username = null
  auth.checked = true
  if (router.currentRoute.value.path !== '/login') {
    router.push('/login')
  }
}

onMounted(() => window.addEventListener('auth:required', onAuthRequired))
onBeforeUnmount(() => window.removeEventListener('auth:required', onAuthRequired))
</script>

<template>
  <router-view />
</template>
