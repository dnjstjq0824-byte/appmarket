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
  padding: 24px;
  max-width: 1000px;
  margin: 0 auto;
}
.view-header {
  margin-bottom: 24px;
}
.subtitle {
  color: #64748b;
  font-size: 14px;
}
.connectors-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 24px;
}
.connector-card, .provenance-panel {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 20px;
}
.connector-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}
.connector-header h3 {
  margin: 0 0 4px 0;
  font-size: 16px;
}
.stats-row {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}
.stat-pill {
  flex: 1;
  background: #f8fafc;
  padding: 8px 12px;
  border-radius: 6px;
}
.stat-pill label {
  font-size: 10px;
  text-transform: uppercase;
  color: #64748b;
  display: block;
  margin-bottom: 2px;
}
.sync-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #64748b;
  margin-bottom: 16px;
  border-top: 1px solid #f1f5f9;
  padding-top: 10px;
}
.card-footer {
  display: flex;
  justify-content: flex-end;
}
.provenance-panel h3 {
  margin: 0 0 6px 0;
}
.panel-desc {
  font-size: 13px;
  color: #64748b;
  margin-bottom: 16px;
}
.tier-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.tier-item {
  display: flex;
  gap: 16px;
  align-items: center;
  padding: 10px;
  border: 1px solid #f1f5f9;
  border-radius: 6px;
}
.tier-badge {
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 700;
  min-width: 140px;
  text-align: center;
}
.tier-1 { background: #dcfce7; color: #166534; }
.tier-2 { background: #eff6ff; color: #1d4ed8; }
.tier-3 { background: #fef3c7; color: #92400e; }
.tier-4 { background: #f3e8ff; color: #6b21a8; }
.tier-5 { background: #f1f5f9; color: #475569; }
.tier-item p {
  margin: 0;
  font-size: 12px;
  color: #334155;
}
.badge {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
}
.badge-success { background: #dcfce7; color: #166534; }
.badge-danger { background: #fee2e2; color: #991b1b; }
.badge-muted { background: #f1f5f9; color: #64748b; }
.toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}
.register-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
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
  gap: 4px;
  font-size: 12px;
  font-weight: 600;
  color: #475569;
}
.form-grid input {
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 400;
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
.sync-result {
  font-size: 12px;
  color: #166534;
  background: #f0fdf4;
  border-radius: 6px;
  padding: 8px 12px;
  margin-bottom: 12px;
}
.btn {
  padding: 6px 14px;
  border-radius: 6px;
  cursor: pointer;
  border: none;
  font-weight: 600;
}
.btn-primary { background: #2563eb; color: white; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: #f1f5f9; color: #334155; }
.btn-sm { font-size: 12px; }
.font-mono { font-family: monospace; }
.text-success { color: #16a34a; }
.text-danger { color: #dc2626; }
.text-muted { color: #94a3b8; }
.text-xs { font-size: 11px; }
</style>

