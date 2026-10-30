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
                📥 Direct Download
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
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 24px;
  margin: 24px 0;
}
.matrix-header {
  margin-bottom: 20px;
}
.matrix-header h3 {
  margin: 0 0 4px 0;
}
.subtitle {
  color: #64748b;
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
  border-bottom: 2px solid #e2e8f0;
  color: #475569;
  font-weight: 600;
}
.comparison-table td {
  padding: 14px 12px;
  border-bottom: 1px solid #f1f5f9;
  vertical-align: middle;
}
.comparison-table tr.is-primary {
  background: #f8fafc;
}
.store-info {
  display: flex;
  align-items: center;
  gap: 8px;
}
.rights-pill {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}
.tier-level_1_direct_hosted { background: #dcfce7; color: #166534; }
.tier-level_2_licensed_federated_binary { background: #eff6ff; color: #1d4ed8; }
.tier-level_3_search_and_info_only { background: #fef3c7; color: #92400e; }
.tier-level_4_official_link_only { background: #f3e8ff; color: #6b21a8; }
.sync-text {
  font-size: 11px;
  color: #94a3b8;
  margin-top: 2px;
}
.cert-box {
  background: #f1f5f9;
  padding: 4px 8px;
  border-radius: 4px;
  display: inline-block;
  font-size: 11px;
}
.price-cell {
  font-weight: 600;
}
.badge {
  padding: 2px 6px;
  border-radius: 10px;
  font-size: 10px;
  font-weight: 700;
}
.badge-primary { background: #dbeafe; color: #1e40af; }
.btn {
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.btn-install {
  background: #2563eb;
  color: white;
  border: none;
}
.btn-outlink {
  background: #f1f5f9;
  color: #1e293b;
  border: 1px solid #cbd5e1;
}
.btn-outlink:hover {
  background: #e2e8f0;
}
.text-success { color: #16a34a; }
.font-mono { font-family: monospace; }
</style>
