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
  background: linear-gradient(
    90deg,
    var(--bn-bg-elevated, #F1F5F9) 25%,
    var(--bn-border-subtle, #E2E8F0) 50%,
    var(--bn-bg-elevated, #F1F5F9) 75%
  );
  background-size: 200% 100%;
  animation: bn-skeleton-shimmer 1.5s infinite linear;
}

@keyframes bn-skeleton-shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}
</style>

