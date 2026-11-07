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
  max-width: 920px;
  margin: 0 auto;
}
.view-header {
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 2px solid var(--bn-ink);
}
.view-header h2 {
  font-family: var(--bn-font-display);
  font-size: 24px;
  font-weight: 900;
  letter-spacing: -0.02em;
  margin: 0 0 4px;
}
.subtitle {
  color: var(--bn-text-secondary);
  font-size: 13px;
  margin: 0;
}
.status-summary-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  border-radius: var(--bn-radius-md);
  margin-bottom: 24px;
  border: 1px solid var(--bn-border-subtle);
  background: var(--bn-bg-surface);
}
.status-valid {
  background: var(--bn-trust-wash);
  border-color: var(--bn-trust-verified);
}
.status-valid .status-info h3 {
  color: var(--bn-trust-verified);
}
.status-warn {
  background: var(--bn-warning-wash);
  border-color: var(--bn-warning-amber);
}
.status-warn .status-info h3 {
  color: var(--bn-warning-amber);
}
.org-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  font-size: 13px;
  color: var(--bn-text-secondary);
}
.org-row label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
}
.org-row select {
  padding: 8px 12px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  font-family: var(--bn-font-sans);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
}
.loading-row {
  padding: 24px;
  color: var(--bn-text-muted);
  font-size: 14px;
  text-align: center;
}
.load-error {
  padding: 12px 16px;
  background: var(--bn-danger-wash);
  border: 1px solid var(--bn-danger-red);
  color: var(--bn-danger-red);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  margin-bottom: 16px;
}
.empty-row {
  padding: 20px 16px;
  color: var(--bn-text-muted);
  font-size: 14px;
  text-align: center;
}
.release-list {
  display: flex;
  flex-direction: column;
  gap: 0;
  margin-top: 12px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  overflow: hidden;
}
.release-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  border-bottom: 1px solid var(--bn-border-subtle);
  font-size: 13px;
}
.release-row:last-child {
  border-bottom: none;
}
.font-mono {
  font-family: var(--bn-font-mono);
  font-size: 12px;
  color: var(--bn-text-secondary);
}
.status-info h3 {
  margin: 0 0 4px 0;
  font-family: var(--bn-font-display);
  font-size: 16px;
  font-weight: 800;
}
.status-info p {
  margin: 0;
  font-size: 13px;
  color: var(--bn-text-secondary);
}
.lineage-tree-section {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 24px;
  margin-bottom: 24px;
}
.lineage-tree-section > h3 {
  font-family: var(--bn-font-display);
  font-size: 15px;
  font-weight: 800;
  margin: 0;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--bn-border-subtle);
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
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 20px;
  background: var(--bn-bg-well);
}
.lineage-node.is-current {
  border-color: var(--bn-trust-verified);
  background: var(--bn-bg-surface);
}
.node-badge {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 14px;
}
.badge {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.04em;
  padding: 3px 9px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-strong);
  color: var(--bn-text-secondary);
  white-space: nowrap;
}
.badge-primary {
  border-color: var(--bn-seal);
  color: var(--bn-seal);
}
.badge-success {
  border-color: var(--bn-trust-verified);
  color: var(--bn-trust-verified);
}
.badge-secondary {
  border-color: var(--bn-border-strong);
  color: var(--bn-text-secondary);
}
.node-field {
  margin-bottom: 14px;
}
.node-field label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  margin-bottom: 6px;
  display: block;
}
.hash-box {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  padding: 8px 12px;
  border-radius: var(--bn-radius-sm);
  word-break: break-all;
  font-size: 12px;
}
.node-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  font-size: 12px;
  color: var(--bn-text-muted);
  border-top: 1px solid var(--bn-border-subtle);
  padding-top: 12px;
}
.node-meta code {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  color: var(--bn-text-secondary);
}
.guidance-card {
  background: var(--bn-warning-wash);
  border: 1px solid var(--bn-warning-amber);
  border-radius: var(--bn-radius-md);
  padding: 20px;
}
.guidance-card h4 {
  font-family: var(--bn-font-display);
  font-size: 14px;
  font-weight: 800;
  margin: 0 0 10px 0;
  color: var(--bn-warning-amber);
}
.guidance-card ul {
  margin: 0;
  padding-left: 20px;
  font-size: 13px;
  line-height: 1.7;
  color: var(--bn-text-secondary);
}
.guidance-card code {
  font-family: var(--bn-font-mono);
  font-size: 12px;
  color: var(--bn-text-primary);
}
</style>

