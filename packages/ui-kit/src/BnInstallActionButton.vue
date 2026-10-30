<template>
  <div class="bn-install-action">
    <!-- DIRECT HOSTED OR VERIFIED BINARY DOWNLOAD -->
    <a
      v-if="isDirectDownload"
      :href="downloadUrl"
      class="bn-btn bn-btn-direct"
      :class="{ 'bn-btn-disabled': disabled }"
      @click="handleClick"
    >
      <svg class="bn-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
        <polyline points="7 10 12 15 17 10" />
        <line x1="12" y1="15" x2="12" y2="3" />
      </svg>
      <span>{{ label || '직접 APK 다운로드' }}</span>
      <span v-if="sizeBytes" class="bn-size-label">({{ formatSize(sizeBytes) }})</span>
    </a>

    <!-- EXTERNAL OFFICIAL STORE OUTLINK (Never fake download button!) -->
    <a
      v-else-if="externalUrl"
      :href="externalUrl"
      target="_blank"
      rel="noopener noreferrer"
      class="bn-btn bn-btn-external"
    >
      <span>{{ externalStoreName || '공식 페이지로 이동' }}</span>
      <svg class="bn-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
        <polyline points="15 3 21 3 21 9" />
        <line x1="10" y1="14" x2="21" y2="3" />
      </svg>
    </a>

    <!-- DISABLED OR RESTRICTED -->
    <button v-else class="bn-btn bn-btn-disabled" disabled>
      {{ disabledReason || '다운로드 불가' }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  distributionMode?: 'DIRECT_HOSTED' | 'EXTERNAL_LINK_ONLY' | 'FEDERATED_REPO' | string
  rightsLevel?: string
  downloadUrl?: string
  externalUrl?: string
  externalStoreName?: string
  sizeBytes?: number
  label?: string
  disabled?: boolean
  disabledReason?: string
}>()

const emit = defineEmits<{
  (e: 'click'): void
}>()

const isDirectDownload = computed(() => {
  return props.distributionMode === 'DIRECT_HOSTED' || props.distributionMode === 'FEDERATED_REPO'
})

function handleClick() {
  if (!props.disabled) {
    emit('click')
  }
}

function formatSize(bytes?: number): string {
  if (!bytes || bytes <= 0) return ''
  const mb = bytes / (1024 * 1024)
  if (mb >= 1) return `${mb.toFixed(1)} MB`
  return `${(bytes / 1024).toFixed(0)} KB`
}
</script>

<style scoped>
.bn-install-action {
  display: inline-flex;
}
.bn-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 600;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
}
.bn-btn-direct {
  background-color: var(--bn-trust-verified, #059669);
  color: #ffffff;
}
.bn-btn-direct:hover {
  filter: brightness(1.1);
}
.bn-btn-external {
  background-color: var(--bn-external-link, #4F46E5);
  color: #ffffff;
}
.bn-btn-external:hover {
  filter: brightness(1.1);
}
.bn-btn-disabled {
  background-color: var(--bn-border-subtle, #E2E8F0);
  color: var(--bn-text-muted, #64748B);
  cursor: not-allowed;
}
.bn-icon {
  width: 18px;
  height: 18px;
}
.bn-size-label {
  font-size: 13px;
  opacity: 0.85;
}
</style>

