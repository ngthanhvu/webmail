<script setup lang="ts">
defineProps<{
  path: string
  size?: number | string
}>()
</script>

<template>
  <svg
    :width="size ?? 20"
    :height="size ?? 20"
    viewBox="0 0 24 24"
    fill="currentColor"
    class="shrink-0"
    aria-hidden="true"
  >
    <path :d="path" />
  </svg>
</template>
