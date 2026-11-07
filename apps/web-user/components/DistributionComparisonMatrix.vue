<template>
  <div class="distribution-comparison-matrix">
    <div class="matrix-header">
      <h3>Multi-Store Distribution & Transparency Comparison (USR-14, 15)</h3>
      <p class="subtitle">Real-time cross-store price, version, and cryptographic signature comparison</p>
    </div>

    <div class="table-responsive">
      <table class="comparison-table">
        <thead>
          <tr>
            <th>Store / Distribution Source</th>
            <th>Rights Provenance</th>
            <th>Version</th>
            <th>Signing Certificate SHA-256</th>
            <th>Price</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="source in sources" :key="source.sourceId" :class="{ 'is-primary': source.isPrimary }">
            <td class="store-cell">
              <div class="store-info">
                <strong>{{ source.storeName }}</strong>
                <span v-if="source.isPrimary" class="badge badge-primary">Recommended</span>
              </div>
            </td>
            <td>
              <span class="rights-pill" :class="'tier-' + source.rightsLevel.toLowerCase()">
                {{ formatRightsLevel(source.rightsLevel) }}
              </span>
              <div class="sync-text">Synced: {{ source.lastSyncedAt }}</div>
            </td>
            <td class="font-mono text-sm">
              v{{ source.versionName }} ({{ source.versionCode }})
            </td>
            <td>
              <div class="cert-box font-mono" :title="source.certSha256">
                {{ formatCertFingerprint(source.certSha256) }}
              </div>
            </td>
            <td class="price-cell font-mono">
              <span :class="source.price === 0 ? 'text-success font-bold' : ''">
                {{ source.price === 0 ? 'Free' : '₩ ' + source.price.toLocaleString() }}
              </span>
            </td>
            <td class="action-cell">
              <button 
                v-if="source.canDirectInstall" 
                @click="onDirectDownload(source)"
                class="btn btn-install btn-sm"
              >
                Direct Download
              </button>
              <a 
                v-else 
                :href="source.officialUrl" 
                target="_blank" 
                rel="noopener noreferrer"
                class="btn btn-outlink btn-sm"
              >
                {{ source.storeName }} ↗
              </a>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
interface DistributionSource {
  sourceId: string;
  storeName: string;
  rightsLevel: string;
  versionName: string;
  versionCode: number;
  certSha256: string;
  price: number;
  canDirectInstall: boolean;
  officialUrl: string;
  lastSyncedAt: string;
  isPrimary?: boolean;
}

const props = defineProps<{
  sources: DistributionSource[];
}>();

const emit = defineEmits<{
  (e: 'download', source: DistributionSource): void;
}>();

function formatRightsLevel(lvl: string): string {
  if (lvl.includes('LEVEL_1')) return 'Level 1: Direct Hosted';
  if (lvl.includes('LEVEL_2')) return 'Level 2: Open Source Mirror';
  if (lvl.includes('LEVEL_3')) return 'Level 3: Info Only';
  if (lvl.includes('LEVEL_4')) return 'Level 4: Official Link';
  return lvl;
}

function formatCertFingerprint(fp: string): string {
  if (!fp || fp.length < 16) return fp || 'N/A';
  return fp.substring(0, 8) + '...' + fp.substring(fp.length - 8);
}

function onDirectDownload(source: DistributionSource) {
  emit('download', source);
}
</script>

<style scoped>
.distribution-comparison-matrix {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 24px;
  margin: 24px 0;
}
.matrix-header {
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--bn-border-subtle);
}
.matrix-header h3 {
  font-family: var(--bn-font-display);
  font-size: 15px;
  font-weight: 800;
  margin: 0 0 4px 0;
}
.subtitle {
  color: var(--bn-text-secondary);
  font-size: 13px;
  margin: 0;
}
.table-responsive {
  overflow-x: auto;
}
.comparison-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.comparison-table th {
  text-align: left;
  padding: 10px 12px;
  border-bottom: 2px solid var(--bn-ink);
  color: var(--bn-text-muted);
  font-family: var(--bn-font-mono);
  font-weight: 400;
  font-size: 10px;
  letter-spacing: 0.08em;
  white-space: nowrap;
}
.comparison-table td {
  padding: 14px 12px;
  border-bottom: 1px solid var(--bn-border-subtle);
  vertical-align: middle;
}
.comparison-table tr:last-child td {
  border-bottom: none;
}
.comparison-table tr.is-primary td:first-child {
  box-shadow: inset 3px 0 0 var(--bn-seal);
}
.store-info {
  display: flex;
  align-items: center;
  gap: 8px;
}
.rights-pill {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.04em;
  padding: 3px 9px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-strong);
  color: var(--bn-text-secondary);
  white-space: nowrap;
}
.tier-level_1_direct_hosted { border-color: var(--bn-trust-verified); color: var(--bn-trust-verified); }
.tier-level_2_licensed_federated_binary { border-color: var(--bn-border-strong); color: var(--bn-text-primary); }
.tier-level_3_search_and_info_only { border-color: var(--bn-warning-amber); color: var(--bn-warning-amber); }
.tier-level_4_official_link_only { border-color: var(--bn-border-strong); color: var(--bn-text-muted); }
.sync-text {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  color: var(--bn-text-muted);
  margin-top: 4px;
}
.cert-box {
  background: var(--bn-bg-well);
  border: 1px solid var(--bn-border-subtle);
  padding: 4px 8px;
  border-radius: var(--bn-radius-sm);
  display: inline-block;
  font-size: 11px;
}
.price-cell {
  font-weight: 600;
}
.badge {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.06em;
  padding: 3px 9px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-seal);
  color: var(--bn-seal);
  white-space: nowrap;
}
.badge-primary { border-color: var(--bn-seal); color: var(--bn-seal); }
.btn {
  padding: 7px 14px;
  border-radius: var(--bn-radius-sm);
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  white-space: nowrap;
}
.btn-install {
  background: var(--bn-action-bg);
  color: var(--bn-action-text);
  border: none;
}
.btn-install:hover {
  background: var(--bn-action-bg-hover);
}
.btn-outlink {
  background: transparent;
  color: var(--bn-text-primary);
  border: 1px solid var(--bn-border-strong);
}
.btn-outlink:hover {
  border-color: var(--bn-text-primary);
}
.text-success { color: var(--bn-trust-verified); }
.text-sm { font-size: 12px; }
.font-bold { font-weight: 700; }
.font-mono { font-family: var(--bn-font-mono); }
</style>

