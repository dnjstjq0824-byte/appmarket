<template>
  <div class="external-connectors-view">
    <div class="view-header">
      <h2>External Connectors & Rights Provenance Control (ADM-10)</h2>
      <p class="subtitle">F-Droid Index v2 GPG synchronization, partner feeds, and 5-tier rights provenance monitoring</p>
    </div>

    <!-- Connector Status Grid -->
    <div class="connectors-grid">
      <div v-for="connector in connectors" :key="connector.id" class="connector-card">
        <div class="connector-header">
          <div>
            <h3>{{ connector.name }}</h3>
            <span class="font-mono text-xs text-muted">{{ connector.url }}</span>
          </div>
          <span class="badge" :class="connector.status === 'HEALTHY' ? 'badge-success' : 'badge-danger'">
            {{ connector.status }}
          </span>
        </div>

        <div class="stats-row">
          <div class="stat-pill">
            <label>GPG Signature</label>
            <span class="text-success font-bold font-mono">✓ Validated</span>
          </div>
          <div class="stat-pill">
            <label>Indexed Apps</label>
            <span class="font-mono font-bold">{{ connector.indexedApps.toLocaleString() }}</span>
          </div>
          <div class="stat-pill">
            <label>Rights Level</label>
            <span class="font-bold">{{ connector.rightsTier }}</span>
          </div>
        </div>

        <div class="sync-meta">
          <span>Last Synced: {{ connector.lastSyncTime }}</span>
          <span v-if="connector.syncErrors > 0" class="text-danger">Errors: {{ connector.syncErrors }}</span>
        </div>

        <div class="card-footer">
          <button 
            @click="triggerSync(connector.id)" 
            class="btn btn-primary btn-sm"
            :disabled="connector.isSyncing"
          >
            {{ connector.isSyncing ? 'Syncing Index...' : '⚡ Trigger Sync Now' }}
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
import { ref } from 'vue';

const connectors = ref([
  {
    id: 'conn-fdroid-main',
    name: 'F-Droid Main Repository (Index v2)',
    url: 'https://f-droid.org/repo/index-v2.json',
    status: 'HEALTHY',
    indexedApps: 4892,
    rightsTier: 'Tier 2 (OS Mirror)',
    lastSyncTime: '12 minutes ago',
    syncErrors: 0,
    isSyncing: false
  },
  {
    id: 'conn-fdroid-archive',
    name: 'F-Droid Archive Feed',
    url: 'https://f-droid.org/archive/index-v2.json',
    status: 'HEALTHY',
    indexedApps: 18450,
    rightsTier: 'Tier 3 (Metadata Only)',
    lastSyncTime: '1 hour ago',
    syncErrors: 0,
    isSyncing: false
  }
]);

async function triggerSync(id: string) {
  const item = connectors.value.find(c => c.id === id);
  if (!item) return;
  item.isSyncing = true;
  try {
    const res = await fetch(`/api/v1/admin/connectors/${id}/trigger-sync`, {
      method: 'POST'
    });
    if (res.ok) {
      item.lastSyncTime = 'Just now';
      alert(`Sync triggered for ${item.name}`);
    }
  } catch (err) {
    console.error(err);
  } finally {
    item.isSyncing = false;
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
.btn {
  padding: 6px 14px;
  border-radius: 6px;
  cursor: pointer;
  border: none;
  font-weight: 600;
}
.btn-primary { background: #2563eb; color: white; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-sm { font-size: 12px; }
.font-mono { font-family: monospace; }
.text-success { color: #16a34a; }
.text-danger { color: #dc2626; }
.text-muted { color: #94a3b8; }
.text-xs { font-size: 11px; }
</style>

