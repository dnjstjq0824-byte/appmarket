<template>
  <div class="honest-install-action-box">
    <!-- Direct Install / Download Mode (Level 1 & Level 2) -->
    <div v-if="canDirectInstall" class="direct-download-container">
      <div class="action-top">
        <button 
          @click="initiateDirectDownload" 
          class="btn-install-primary" 
          :disabled="isRequestingToken"
          id="btnDirectDownload"
        >
          <svg class="btn-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3v12" /><path d="M7 10l5 5 5-5" /><path d="M4 21h16" /></svg>
          <span class="btn-label">
            {{ isRequestingToken ? 'Generating Secure Token...' : 'Download APK Directly' }}
          </span>
          <span class="btn-size font-mono">({{ formatBytes(fileSizeBytes) }})</span>
        </button>
      </div>

      <div class="download-metadata">
        <div class="hash-row">
          <span class="label">SHA-256 Checksum:</span>
          <code class="font-mono text-xs">{{ sha256Checksum }}</code>
          <button @click="copyHash" class="copy-btn btn-xs">Copy</button>
        </div>
        <div class="token-info">
          <span>Download link is cryptographically signed and expires in 15 minutes.</span>
        </div>
      </div>
    </div>

    <!-- External Official Link Mode (Level 3 & Level 4) -->
    <div v-else class="external-link-container">
      <div class="action-top">
        <a 
          :href="officialStoreUrl" 
          target="_blank" 
          rel="noopener noreferrer" 
          class="btn-external-link"
          id="btnExternalStoreLink"
        >
          <svg class="btn-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" /><path d="M15 3h6v6" /><path d="M10 14L21 3" /></svg>
          <span class="btn-label">View on {{ provenanceLabel }} ↗</span>
        </a>
      </div>

      <div class="external-provenance-notice">
        <div class="notice-icon" aria-hidden="true">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="9" /><path d="M12 11v5" /><path d="M12 8h.01" /></svg>
        </div>
        <div class="notice-text">
          <strong>External Distribution Only</strong>
          <p>
            This application is hosted externally by {{ provenanceLabel }}. In accordance with our honest distribution policy, we do not mirror or repackage binaries without authorization. Clicking above will securely redirect you to the official provider.
          </p>
          <span class="sync-date">Last metadata sync: {{ lastSyncedAt }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const props = defineProps<{
  appId: string;
  releaseId?: string;
  canDirectInstall: boolean;
  fileSizeBytes: number;
  sha256Checksum: string;
  provenanceLabel: string;
  officialStoreUrl: string;
  lastSyncedAt: string;
}>();

const isRequestingToken = ref(false);

async function initiateDirectDownload() {
  if (!props.releaseId) return;
  isRequestingToken.value = true;
  try {
    const res = await fetch(`/api/v1/catalog/apps/${props.appId}/releases/${props.releaseId}/download-token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' }
    });

    if (res.ok) {
      const { url } = await res.json();
      window.location.href = url;
    } else {
      alert('Unable to generate download token. Please try again.');
    }
  } catch (err) {
    console.error('Download token error', err);
  } finally {
    isRequestingToken.value = false;
  }
}

function copyHash() {
  navigator.clipboard.writeText(props.sha256Checksum);
  alert('SHA-256 fingerprint copied to clipboard!');
}

function formatBytes(bytes: number): string {
  if (!bytes) return '0 B';
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
</script>

<style scoped>
.honest-install-action-box {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 20px;
  margin: 20px 0;
}
.btn-install-primary {
  width: 100%;
  background: var(--bn-action-bg);
  color: var(--bn-action-text);
  padding: 14px 24px;
  border-radius: var(--bn-radius-sm);
  border: none;
  font-size: 15px;
  font-weight: 700;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}
.btn-install-primary:hover:not(:disabled) {
  background: var(--bn-action-bg-hover);
}
.btn-install-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.btn-external-link {
  width: 100%;
  background: transparent;
  color: var(--bn-text-primary);
  padding: 14px 24px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-strong);
  font-size: 15px;
  font-weight: 700;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  text-decoration: none;
  box-sizing: border-box;
}
.btn-external-link:hover {
  border-color: var(--bn-text-primary);
}
.btn-icon {
  flex-shrink: 0;
}
.btn-size {
  font-family: var(--bn-font-mono);
  font-size: 12px;
  font-weight: 400;
  opacity: 0.7;
}
.download-metadata {
  margin-top: 14px;
  font-size: 12px;
  color: var(--bn-text-muted);
}
.hash-row {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--bn-bg-well);
  border: 1px solid var(--bn-border-subtle);
  padding: 8px 10px;
  border-radius: var(--bn-radius-sm);
  margin-bottom: 8px;
}
.hash-row .label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  white-space: nowrap;
}
.hash-row code {
  flex: 1;
  min-width: 0;
  color: var(--bn-text-primary);
  word-break: break-all;
}
.copy-btn {
  background: transparent;
  border: 1px solid var(--bn-border-subtle);
  color: var(--bn-text-secondary);
  padding: 3px 10px;
  border-radius: var(--bn-radius-sm);
  cursor: pointer;
  font-size: 11px;
  font-weight: 600;
  white-space: nowrap;
}
.copy-btn:hover {
  border-color: var(--bn-border-strong);
  color: var(--bn-text-primary);
}
.token-info {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  color: var(--bn-trust-verified);
}
.external-provenance-notice {
  margin-top: 14px;
  background: var(--bn-bg-well);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  padding: 14px;
  display: flex;
  gap: 12px;
}
.notice-icon {
  color: var(--bn-text-secondary);
  flex-shrink: 0;
  margin-top: 1px;
}
.notice-text strong {
  display: block;
  font-size: 13px;
  margin-bottom: 4px;
}
.notice-text p {
  margin: 0 0 8px 0;
  font-size: 12px;
  color: var(--bn-text-secondary);
  line-height: 1.6;
}
.sync-date {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  color: var(--bn-text-muted);
}
.font-mono { font-family: var(--bn-font-mono); }
.text-xs { font-size: 11px; }
</style>

