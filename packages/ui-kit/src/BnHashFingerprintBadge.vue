<template>
  <span
    class="bn-hash-badge"
    :title="copied ? '복사되었습니다!' : '클릭하여 전체 해시 복사'"
    @click="copyToClipboard"
  >
    <code class="bn-hash-text">{{ truncatedHash }}</code>
    <svg class="bn-copy-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
      <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
    </svg>
  </span>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

const props = defineProps<{
  hash: string
}>()

const copied = ref(false)

const truncatedHash = computed(() => {
  if (!props.hash || props.hash.length < 16) return props.hash || ''
  return `${props.hash.slice(0, 8)}...${props.hash.slice(-8)}`
})

async function copyToClipboard() {
  if (!props.hash) return
  try {
    await navigator.clipboard.writeText(props.hash)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch (err) {
    console.error('Failed to copy hash: ', err)
  }
}
</script>

<style scoped>
.bn-hash-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background-color: var(--bn-bg-well);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  padding: 3px 8px;
  cursor: pointer;
  user-select: none;
}
.bn-hash-badge:hover {
  border-color: var(--bn-border-strong);
}
.bn-hash-text {
  font-family: var(--bn-font-mono);
  font-size: 12px;
  color: var(--bn-text-secondary);
}
.bn-copy-icon {
  width: 13px;
  height: 13px;
  opacity: 0.6;
}
</style>

