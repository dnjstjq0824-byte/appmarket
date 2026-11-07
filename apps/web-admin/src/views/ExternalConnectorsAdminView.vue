<template>
  <div class="external-connectors-view">
    <div class="view-header">
      <h2>External Connectors & Rights Provenance Control (ADM-10)</h2>
      <p class="subtitle">F-Droid Index v2 GPG synchronization, partner feeds, and 5-tier rights provenance monitoring</p>
    </div>

    <div class="toolbar">
      <button @click="showRegister = !showRegister" class="btn btn-secondary btn-sm">
        {{ showRegister ? 'Cancel' : '+ Register Connector' }}
      </button>
    </div>

    <form v-if="showRegister" class="register-card" @submit.prevent="registerConnector">
      <div class="form-grid">
        <label>ID <input v-model="regForm.id" required placeholder="partner-x" /></label>
        <label>Name <input v-model="regForm.name" required placeholder="Partner X Feed" /></label>
        <label>Kind <input v-model="regForm.kind" required placeholder="FDROID | PARTNER | OFFICIAL_LINK" /></label>
        <label>Feed URL <input v-model="regForm.feedUrl" placeholder="https://..." /></label>
        <label>Fingerprint <input v-model="regForm.fingerprint" placeholder="GPG fingerprint" /></label>
        <label>Packages (comma separated) <input v-model="regForm.packages" placeholder="com.x.app" /></label>
      </div>
      <button type="submit" class="btn btn-primary btn-sm" :disabled="isRegistering">
        {{ isRegistering ? 'Registering…' : 'Register' }}
      </button>
    </form>

    <div v-if="isLoading" class="loading-row">Loading connectors…</div>
    <div v-if="loadError" class="load-error">{{ loadError }}</div>
    <div v-if="actionError" class="load-error">{{ actionError }}</div>

    <!-- Connector Status Grid -->
    <div v-if="!isLoading && !loadError" class="connectors-grid">
      <div v-for="connector in connectors" :key="connector.id" class="connector-card">
        <div class="connector-header">
          <div>
            <h3>{{ connector.name }}</h3>
            <span class="font-mono text-xs text-muted">{{ connector.feedUrl || connector.kind }}</span>
          </div>
          <span class="badge" :class="connector.syncable ? 'badge-success' : 'badge-muted'">
            {{ connector.syncable ? 'SYNCABLE' : 'LINK ONLY' }}
          </span>
        </div>

        <div class="stats-row">
          <div class="stat-pill">
            <label>Kind</label>
            <span class="font-bold font-mono">{{ connector.kind }}</span>
          </div>
          <div class="stat-pill">
            <label>Last Item Count</label>
            <span class="font-mono font-bold">{{ (connector.lastItemCount ?? 0).toLocaleString() }}</span>
          </div>
          <div class="stat-pill">
            <label>Last Sync</label>
            <span class="font-bold text-xs">{{ connector.lastSyncAt || 'Never' }}</span>
          </div>
        </div>

        <div v-if="lastResult[connector.id]" class="sync-result">
          Last sync: {{ lastResult[connector.id] }}
        </div>

        <div class="card-footer">
          <button
            @click="triggerSync(connector.id)"
            class="btn btn-primary btn-sm"
            :disabled="syncing[connector.id] || !connector.syncable"
          >
            {{ syncing[connector.id] ? 'Syncing Index...' : 'Trigger Sync Now' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Rights Provenance Breakdown Panel -->
    <div class="provenance-panel">
      <h3>5-Tier Data Provenance Enforcement</h3>
      <p class="panel-desc">Strict legal separation of external catalog content per BN Master Directive</p>

      <div class="tier-list">
        <div class="tier-item">
          <span class="tier-badge tier-1">Level 1: Direct Hosted</span>
          <p>Native developer upload with explicit distribution grant. Full CDN hosting, in-app updates allowed.</p>
        </div>
        <div class="tier-item">
          <span class="tier-badge tier-2">Level 2: Open Source Mirror</span>
          <p>F-Droid / GPL / Apache open-source licensed binaries mirrored with cryptographic hash verification.</p>
        </div>
        <div class="tier-item">
          <span class="tier-badge tier-3">Level 3: Metadata Cached</span>
          <p>Public metadata cached for search discovery. Binary download redirects directly to upstream official site.</p>
        </div>
        <div class="tier-item">
          <span class="tier-badge tier-4">Level 4: Partner Outlink</span>
          <p>Verified commercial partner catalog. All downloads and commerce handled on partner landing page.</p>
        </div>
        <div class="tier-item">
          <span class="tier-badge tier-5">Level 5: Discovery Directory</span>
          <p>External index reference only. Strictly forbidden from initiating background or direct downloads.</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getSharedClient, ApiClientError } from '@bookingnetwork/api-client';

const api = getSharedClient();

interface Connector {
  id: string;
  name: string;
  kind: string;
  feedUrl?: string;
  syncable: boolean;
  lastSyncAt?: string;
  lastItemCount?: number;
}

const connectors = ref<Connector[]>([]);
const syncing = ref<Record<string, boolean>>({});
const lastResult = ref<Record<string, string>>({});
const isLoading = ref(true);
const loadError = ref('');
const actionError = ref('');

const regForm = ref({ id: '', name: '', kind: 'FDROID', feedUrl: '', fingerprint: '', packages: '' });
const showRegister = ref(false);
const isRegistering = ref(false);

onMounted(async () => {
  await loadConnectors();
  isLoading.value = false;
});

async function loadConnectors() {
  loadError.value = '';
  try {
    connectors.value = await api.listConnectors() as Connector[];
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load connectors';
  }
}

async function triggerSync(id: string) {
  syncing.value[id] = true;
  actionError.value = '';
  try {
    const res = await api.triggerConnectorSync(id) as { syncedAt: string; itemCount: number };
    lastResult.value[id] = `${res.itemCount} items at ${res.syncedAt}`;
    await loadConnectors();
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : `Sync failed for ${id}`;
  } finally {
    syncing.value[id] = false;
  }
}

async function registerConnector() {
  isRegistering.value = true;
  actionError.value = '';
  try {
    await api.registerConnector({
      id: regForm.value.id.trim(),
      name: regForm.value.name.trim(),
      kind: regForm.value.kind.trim(),
      feedUrl: regForm.value.feedUrl.trim() || undefined,
      fingerprint: regForm.value.fingerprint.trim() || undefined,
      packages: regForm.value.packages.split(',').map(p => p.trim()).filter(Boolean)
    });
    regForm.value = { id: '', name: '', kind: 'FDROID', feedUrl: '', fingerprint: '', packages: '' };
    showRegister.value = false;
    await loadConnectors();
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to register connector';
  } finally {
    isRegistering.value = false;
  }
}
</script>

<style scoped>
.external-connectors-view {
  max-width: 1020px;
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
.connectors-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 24px;
}
.connector-card, .provenance-panel {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 20px;
}
.connector-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
}
.connector-header h3 {
  font-family: var(--bn-font-display);
  margin: 0 0 4px 0;
  font-size: 15px;
  font-weight: 800;
}
.stats-row {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}
.stat-pill {
  flex: 1;
  background: var(--bn-bg-well);
  border: 1px solid var(--bn-border-subtle);
  padding: 10px 12px;
  border-radius: var(--bn-radius-sm);
}
.stat-pill label {
  font-family: var(--bn-font-mono);
  font-size: 9px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  display: block;
  margin-bottom: 4px;
}
.font-bold { font-weight: 700; }
.sync-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--bn-text-muted);
  margin-bottom: 16px;
  border-top: 1px solid var(--bn-border-subtle);
  padding-top: 10px;
}
.card-footer {
  display: flex;
  justify-content: flex-end;
}
.provenance-panel h3 {
  font-family: var(--bn-font-display);
  font-size: 15px;
  font-weight: 800;
  margin: 0 0 6px 0;
}
.panel-desc {
  font-size: 13px;
  color: var(--bn-text-secondary);
  margin: 0 0 16px;
}
.tier-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.tier-item {
  display: flex;
  gap: 16px;
  align-items: center;
  padding: 12px 14px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
}
.tier-badge {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.04em;
  padding: 4px 10px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-strong);
  color: var(--bn-text-secondary);
  min-width: 150px;
  text-align: center;
  flex-shrink: 0;
}
.tier-1 { border-color: var(--bn-trust-verified); color: var(--bn-trust-verified); }
.tier-2 { border-color: var(--bn-seal); color: var(--bn-seal); }
.tier-3 { border-color: var(--bn-warning-amber); color: var(--bn-warning-amber); }
.tier-4 { border-color: var(--bn-border-strong); color: var(--bn-text-secondary); }
.tier-5 { border-color: var(--bn-border-subtle); color: var(--bn-text-muted); }
.tier-item p {
  margin: 0;
  font-size: 12px;
  line-height: 1.6;
  color: var(--bn-text-secondary);
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
.badge-success { border-color: var(--bn-trust-verified); color: var(--bn-trust-verified); }
.badge-danger { border-color: var(--bn-danger-red); color: var(--bn-danger-red); }
.badge-muted { border-color: var(--bn-border-subtle); color: var(--bn-text-muted); }
.toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}
.register-card {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 20px;
  margin-bottom: 20px;
}
.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-bottom: 16px;
}
.form-grid label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.06em;
  color: var(--bn-text-muted);
}
.form-grid input {
  padding: 9px 12px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  font-family: var(--bn-font-sans);
  font-size: 13px;
}
.form-grid input:focus {
  outline: none;
  border-color: var(--bn-ink);
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
.sync-result {
  font-family: var(--bn-font-mono);
  font-size: 12px;
  color: var(--bn-trust-verified);
  background: var(--bn-trust-wash);
  border: 1px solid var(--bn-trust-verified);
  border-radius: var(--bn-radius-sm);
  padding: 8px 12px;
  margin-bottom: 12px;
}
.btn {
  padding: 8px 16px;
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}
.btn-primary {
  background: var(--bn-action-bg);
  color: var(--bn-action-text);
  border: none;
}
.btn-primary:hover:not(:disabled) {
  background: var(--bn-action-bg-hover);
}
.btn-primary:disabled { opacity: 0.45; cursor: not-allowed; }
.btn-secondary {
  background: transparent;
  border: 1px solid var(--bn-border-subtle);
  color: var(--bn-text-secondary);
}
.btn-secondary:hover:not(:disabled) {
  border-color: var(--bn-border-strong);
  color: var(--bn-text-primary);
}
.btn-secondary:disabled { opacity: 0.45; cursor: not-allowed; }
.btn-sm { font-size: 12px; padding: 6px 14px; }
.font-mono { font-family: var(--bn-font-mono); }
.text-success { color: var(--bn-trust-verified); }
.text-danger { color: var(--bn-danger-red); }
.text-muted { color: var(--bn-text-muted); }
.text-xs { font-size: 11px; }
@media (max-width: 860px) {
  .connectors-grid, .form-grid {
    grid-template-columns: 1fr;
  }
  .tier-item {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>

