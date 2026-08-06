<template>
  <div class="brand">
    <div class="brand-badge">
      <PlayIcon :size="14" />
    </div>
    <span class="brand-name">LVS</span>
  </div>
</template>

<script setup>
import PlayIcon from './PlayIcon.vue'
</script>

<style scoped>
.brand {
  display: flex;
  align-items: center;
  gap: 10px;
}
.brand-badge {
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 9px;
  background: var(--grad);
  color: #fff;
  box-shadow: 0 6px 18px rgba(124, 92, 255, 0.4);
}
.brand-name {
  font-size: 17px;
  font-weight: 800;
  letter-spacing: 4px;
  background: var(--grad);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
</style>
