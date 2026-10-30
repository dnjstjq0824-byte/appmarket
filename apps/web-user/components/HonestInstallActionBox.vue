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
          <span class="btn-icon">📥</span>
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
          <span>🛡️ Download link is cryptographically signed and expires in 15 minutes.</span>
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
          <span class="btn-icon">🔗</span>
          <span class="btn-label">View on {{ provenanceLabel }} ↗</span>
        </a>
      </div>

      <div class="external-provenance-notice">
        <div class="notice-icon">ℹ️</div>
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
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 20px;
  margin: 20px 0;
}
.btn-install-primary {
  width: 100%;
  background: #2563eb;
  color: white;
  padding: 14px 24px;
  border-radius: 8px;
  border: none;
  font-size: 16px;
  font-weight: 700;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  transition: background 0.15s;
}
.btn-install-primary:hover {
  background: #1d4ed8;
}
.btn-install-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.btn-external-link {
  width: 100%;
  background: #0f172a;
  color: white;
  padding: 14px 24px;
  border-radius: 8px;
  border: none;
  font-size: 16px;
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
  background: #1e293b;
}
.btn-size {
  font-size: 13px;
  opacity: 0.85;
}
.download-metadata {
  margin-top: 14px;
  font-size: 12px;
  color: #64748b;
}
.hash-row {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #f8fafc;
  padding: 8px;
  border-radius: 4px;
  margin-bottom: 6px;
}
.hash-row code {
  color: #0f172a;
  word-break: break-all;
}
.copy-btn {
  background: #e2e8f0;
  border: none;
  padding: 2px 8px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 11px;
}
.token-info {
  font-size: 11px;
  color: #166534;
}
.external-provenance-notice {
  margin-top: 14px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-left: 4px solid #3b82f6;
  border-radius: 4px;
  padding: 12px;
  display: flex;
  gap: 12px;
}
.notice-icon {
  font-size: 20px;
}
.notice-text strong {
  display: block;
  font-size: 13px;
  color: #1e293b;
  margin-bottom: 2px;
}
.notice-text p {
  margin: 0 0 6px 0;
  font-size: 12px;
  color: #475569;
  line-height: 1.4;
}
.sync-date {
  font-size: 11px;
  color: #94a3b8;
}
.font-mono { font-family: monospace; }
</style>

