<template>
  <div
    class="bn-skeleton"
    :style="{
      width: width || '100%',
      height: height || '20px',
      borderRadius: radius || '6px'
    }"
  ></div>
</template>

<script setup lang="ts">
defineProps<{
  width?: string
  height?: string
  radius?: string
}>()
</script>

<style scoped>
.bn-skeleton {
  background-color: var(--bn-bg-well);
  border: 1px solid var(--bn-border-subtle);
  animation: bn-skeleton-pulse 1.6s infinite ease-in-out;
}

@keyframes bn-skeleton-pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.45;
  }
}
</style>

