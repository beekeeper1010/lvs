<template>
  <svg :width="size" :height="size" viewBox="0 0 24 24" fill="currentColor">
    <path d="M8 5.14v13.72c0 .81.89 1.3 1.57.87l10.6-6.86a1.04 1.04 0 0 0 0-1.74L9.57 4.27A1.03 1.03 0 0 0 8 5.14z" />
  </svg>
</template>

<script setup>
defineProps({
  size: { type: [Number, String], default: 14 },
})
</script>
