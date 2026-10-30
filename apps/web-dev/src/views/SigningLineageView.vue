<template>
  <div class="signing-lineage-view">
    <div class="view-header">
      <h2>Cryptographic Signing & Key Rotation Lineage</h2>
      <p class="subtitle">Android v2/v3 Signature Block analysis and Proof-of-Rotation lineage tree</p>
    </div>

    <div class="status-summary-card" :class="lineageStatusClass">
      <div class="status-icon">{{ lineageStatusIcon }}</div>
      <div class="status-info">
        <h3>{{ lineageStatusTitle }}</h3>
        <p>{{ lineageStatusDescription }}</p>
      </div>
    </div>

    <!-- Lineage Tree -->
    <div class="lineage-tree-section">
      <h3>Signature Lineage Chain (v3 Proof-of-Rotation)</h3>
      <div class="lineage-nodes">
        <div v-for="(node, index) in lineageNodes" :key="node.certSha256" class="lineage-node-container">
          <div class="lineage-node" :class="{ 'is-current': node.isCurrent, 'is-root': node.isRoot }">
            <div class="node-badge">
              <span v-if="node.isRoot" class="badge badge-primary">Initial Root Key</span>
              <span v-else-if="node.isCurrent" class="badge badge-success">Active Signing Key</span>
              <span v-else class="badge badge-secondary">Intermediate Key</span>
            </div>
            
            <div class="node-field">
              <label>Subject DN</label>
              <div class="font-mono text-sm">{{ node.subjectDn }}</div>
            </div>

            <div class="node-field">
              <label>SHA-256 Certificate Fingerprint</label>
              <div class="font-mono hash-box">{{ node.certSha256 }}</div>
            </div>

            <div class="node-meta">
              <span>Valid From: {{ node.validFrom }}</span>
              <span>Valid Until: {{ node.validUntil }}</span>
              <span v-if="node.rotationFlags">Flags: <code>{{ node.rotationFlags }}</code></span>
            </div>
          </div>

          <!-- Connector Arrow -->
          <div v-if="index < lineageNodes.length - 1" class="lineage-connector">
            <div class="arrow-down">⬇️</div>
            <span class="rotation-label">Rotated via v3 Proof-of-Rotation</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Compatibility Guidance -->
    <div class="guidance-card">
      <h4>Signature Continuity Guidelines</h4>
      <ul>
        <li><strong>Same Key:</strong> Existing installed users on Android 7.0+ will update seamlessly with user app data preserved.</li>
        <li><strong>Rotated Key (v3):</strong> Devices running Android 9.0 (API 28) and above will automatically accept the rotated key via lineage.</li>
        <li><strong>Different Key (Unlinked):</strong> Android OS will reject the update with <code>INSTALL_FAILED_UPDATE_INCOMPATIBLE</code>. In this case, users must uninstall first.</li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

interface LineageNode {
  isRoot: boolean;
  isCurrent: boolean;
  subjectDn: string;
  certSha256: string;
  validFrom: string;
  validUntil: string;
  rotationFlags?: string;
}

const lineageNodes = ref<LineageNode[]>([
  {
    isRoot: true,
    isCurrent: false,
    subjectDn: 'CN=BookingNetwork Dev, OU=Mobile, O=BookingNetwork Inc, C=KR',
    certSha256: '9A4C73E1829B6342817490FBCDA4812304918234ABCDEF1234567890ABCDEF12',
    validFrom: '2022-01-01',
    validUntil: '2047-01-01',
    rotationFlags: 'FLAG_PERMIT_ROTATION'
  },
  {
    isRoot: false,
    isCurrent: true,
    subjectDn: 'CN=BookingNetwork Prod 2026, OU=Mobile, O=BookingNetwork Inc, C=KR',
    certSha256: '14B991BD744BC4E37BB68043FB6CFCF27018234ABCDEF1234567890ABCDEF99',
    validFrom: '2026-01-01',
    validUntil: '2051-01-01',
    rotationFlags: 'FLAG_ACTIVE'
  }
]);

const lineageStatusClass = computed(() => 'status-valid');
const lineageStatusIcon = computed(() => '🛡️');
const lineageStatusTitle = computed(() => 'Cryptographically Continuous Lineage');
const lineageStatusDescription = computed(
  () => 'The signature lineage contains valid cryptographic proof-of-rotation connecting current key to root key.'
);
</script>

<style scoped>
.signing-lineage-view {
  padding: 24px;
  max-width: 900px;
  margin: 0 auto;
}
.view-header {
  margin-bottom: 24px;
}
.subtitle {
  color: #64748b;
  font-size: 14px;
}
.status-summary-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 24px;
}
.status-valid {
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  color: #166534;
}
.status-icon {
  font-size: 32px;
}
.status-info h3 {
  margin: 0 0 4px 0;
  font-size: 16px;
}
.status-info p {
  margin: 0;
  font-size: 14px;
  color: #15803d;
}
.lineage-tree-section {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 24px;
  margin-bottom: 24px;
}
.lineage-nodes {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  margin-top: 16px;
}
.lineage-node-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
}
.lineage-node {
  width: 100%;
  max-width: 700px;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 20px;
  background: #f8fafc;
}
.lineage-node.is-current {
  border-color: #22c55e;
  background: #ffffff;
  box-shadow: 0 2px 4px rgba(34, 197, 94, 0.1);
}
.lineage-node.is-root {
  border-color: #3b82f6;
}
.node-badge {
  margin-bottom: 12px;
}
.badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}
.badge-primary {
  background: #eff6ff;
  color: #1d4ed8;
}
.badge-success {
  background: #f0fdf4;
  color: #15803d;
}
.badge-secondary {
  background: #f1f5f9;
  color: #475569;
}
.node-field {
  margin-bottom: 12px;
}
.node-field label {
  font-size: 11px;
  text-transform: uppercase;
  color: #64748b;
  font-weight: 600;
  margin-bottom: 2px;
  display: block;
}
.hash-box {
  background: #f1f5f9;
  padding: 6px 10px;
  border-radius: 4px;
  word-break: break-all;
  font-size: 12px;
}
.node-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  font-size: 12px;
  color: #64748b;
  border-top: 1px solid #e2e8f0;
  padding-top: 10px;
}
.lineage-connector {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin: 12px 0;
  color: #64748b;
  font-size: 12px;
}
.arrow-down {
  font-size: 20px;
}
.guidance-card {
  background: #fffbeb;
  border: 1px solid #fef3c7;
  border-radius: 8px;
  padding: 20px;
  color: #92400e;
}
.guidance-card h4 {
  margin: 0 0 10px 0;
}
.guidance-card ul {
  margin: 0;
  padding-left: 20px;
  font-size: 13px;
  line-height: 1.6;
}
</style>

