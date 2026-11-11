<script setup>
defineProps({
  path: { type: String, required: true },
  size: { type: [Number, String], default: 20 },
})
</script>

<template>
  <svg
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="currentColor"
    class="shrink-0"
    aria-hidden="true"
  >
    <path :d="path" />
  </svg>
</template>
