<template>
  <div class="signing-lineage-view">
    <div class="view-header">
      <h2>Cryptographic Signing & Key Rotation Lineage</h2>
      <p class="subtitle">Android v2/v3 Signature Block analysis and Proof-of-Rotation lineage tree</p>
    </div>

    <div v-if="orgs.length > 0" class="org-row">
      <label for="orgSelect">Organization</label>
      <select id="orgSelect" v-model="activeOrgId" @change="loadLineage">
        <option v-for="org in orgs" :key="org.id" :value="org.id">{{ org.org_name }}</option>
      </select>
    </div>
    <div v-if="isLoading" class="loading-row">Loading signing lineage…</div>
    <div v-if="loadError" class="load-error">{{ loadError }}</div>

    <div v-if="!isLoading && !loadError" class="status-summary-card" :class="lineageStatusClass">
      <div class="status-info">
        <h3>{{ lineageStatusTitle }}</h3>
        <p>{{ lineageStatusDescription }}</p>
      </div>
    </div>

    <!-- Releases -->
    <div v-if="!isLoading && !loadError" class="lineage-tree-section">
      <h3>Releases ({{ releases.length }})</h3>
      <div v-if="releases.length === 0" class="empty-row">No releases recorded for this app.</div>
      <div v-else class="release-list">
        <div v-for="rel in releases" :key="rel.id" class="release-row">
          <strong>v{{ rel.versionName }}</strong>
          <span class="font-mono">build {{ rel.versionCode }}</span>
          <span class="badge badge-secondary">{{ rel.track }}</span>
          <span class="badge badge-primary">{{ rel.status }}</span>
        </div>
      </div>
    </div>

    <!-- Signature nodes -->
    <div v-if="!isLoading && !loadError" class="lineage-tree-section">
      <h3>Recorded Signing Keys ({{ signatures.length }})</h3>
      <div v-if="signatures.length === 0" class="empty-row">No signatures recorded yet.</div>
      <div v-else class="lineage-nodes">
        <div v-for="node in signatures" :key="node.artifactId" class="lineage-node-container">
          <div class="lineage-node" :class="{ 'is-current': node.isPrimary }">
            <div class="node-badge">
              <span v-if="node.isPrimary" class="badge badge-success">Primary Lineage Key</span>
              <span v-else class="badge badge-secondary">Recorded Key</span>
              <span class="badge badge-primary">Scheme v{{ node.schemeVersion }}</span>
            </div>

            <div class="node-field">
              <label>SHA-256 Certificate Fingerprint</label>
              <div class="font-mono hash-box">{{ node.certSha256 }}</div>
            </div>

            <div class="node-meta">
              <span>Distribution: <code>{{ node.distributionId }}</code></span>
              <span>Artifact: <code>{{ node.artifactId }}</code></span>
            </div>
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
import { ref, computed, onMounted } from 'vue';
import { getSharedClient, ApiClientError, type OrgSummary } from '@bookingnetwork/api-client';

const api = getSharedClient();

interface SignatureNode {
  distributionId: string;
  artifactId: string;
  certSha256: string;
  schemeVersion: number;
  isPrimary: boolean;
}

interface ReleaseNode {
  id: string;
  versionName: string;
  versionCode: number;
  track: string;
  status: string;
}

const orgs = ref<OrgSummary[]>([]);
const activeOrgId = ref('');
const signatures = ref<SignatureNode[]>([]);
const releases = ref<ReleaseNode[]>([]);
const isLoading = ref(true);
const loadError = ref('');

onMounted(async () => {
  try {
    orgs.value = await api.listOrgs();
    if (orgs.value.length > 0) {
      activeOrgId.value = orgs.value[0].id;
      await loadLineage();
    }
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load organizations';
  } finally {
    isLoading.value = false;
  }
});

async function loadLineage() {
  if (!activeOrgId.value) return;
  isLoading.value = true;
  loadError.value = '';
  try {
    const lineage = await api.getLineage('current', activeOrgId.value) as { releases: ReleaseNode[]; signatures: SignatureNode[] };
    releases.value = lineage.releases || [];
    signatures.value = lineage.signatures || [];
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load lineage';
  } finally {
    isLoading.value = false;
  }
}

const hasV2Plus = computed(() => signatures.value.some(s => s.schemeVersion >= 2));
const lineageStatusClass = computed(() => (hasV2Plus.value ? 'status-valid' : 'status-warn'));
const lineageStatusTitle = computed(() =>
  signatures.value.length === 0
    ? 'No Signatures Recorded'
    : hasV2Plus.value
      ? 'Verified v2/v3 Signing Lineage'
      : 'Legacy v1 Signatures Only'
);
const lineageStatusDescription = computed(() =>
  signatures.value.length === 0
    ? 'Upload a signed binary to record its certificate lineage.'
    : hasV2Plus.value
      ? `${signatures.value.length} signature record(s) on file. All current keys meet the v2+ scheme requirement.`
      : 'Current signatures use scheme v1. Re-sign with v2/v3 to meet the release requirement.'
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
.status-warn {
  background: #fffbeb;
  border: 1px solid #fde68a;
  color: #92400e;
}
.org-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  font-size: 13px;
}
.org-row select {
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
  background: white;
}
.loading-row {
  padding: 16px;
  color: #64748b;
  font-size: 14px;
}
.load-error {
  padding: 12px 16px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #991b1b;
  border-radius: 6px;
  font-size: 13px;
  margin-bottom: 16px;
}
.empty-row {
  padding: 16px;
  color: #64748b;
  font-size: 14px;
  text-align: center;
}
.release-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 12px;
}
.release-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  background: #f8fafc;
  border-radius: 6px;
  font-size: 13px;
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

