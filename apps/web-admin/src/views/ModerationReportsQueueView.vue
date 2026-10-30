<template>
  <div class="moderation-reports-view">
    <div class="view-header">
      <h2>Community Moderation & DMCA Triage Queue (ADM-08)</h2>
      <p class="subtitle">30-minute reviewer lease locking to prevent dual-triage conflicts</p>
    </div>

    <!-- Filter Tabs -->
    <div class="tabs-container">
      <button 
        v-for="tab in ['PENDING', 'CLAIMED', 'RESOLVED', 'DISMISSED']" 
        :key="tab"
        class="tab-btn"
        :class="{ active: currentTab === tab }"
        @click="currentTab = tab"
      >
        {{ tab }} ({{ getReportCount(tab) }})
      </button>
    </div>

    <!-- Reports Table -->
    <div class="table-card">
      <table class="reports-table">
        <thead>
          <tr>
            <th>Report ID</th>
            <th>Type</th>
            <th>Target App</th>
            <th>Reporter</th>
            <th>Reason Summary</th>
            <th>Lease Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="report in filteredReports" :key="report.id">
            <td class="font-mono text-sm">{{ report.id }}</td>
            <td>
              <span class="badge" :class="'badge-' + report.type.toLowerCase()">{{ report.type }}</span>
            </td>
            <td>
              <strong>{{ report.appName }}</strong>
              <div class="font-mono text-xs text-muted">{{ report.packageName }}</div>
            </td>
            <td>{{ report.reporterEmail }}</td>
            <td class="summary-cell">{{ report.summary }}</td>
            <td>
              <div v-if="report.claimedBy" class="lease-box">
                <span class="claimed-name">👤 {{ report.claimedBy }}</span>
                <span class="lease-timer font-mono text-xs">Expires in {{ report.remainingLeaseMin }}m</span>
              </div>
              <span v-else class="text-muted text-xs">Unclaimed</span>
            </td>
            <td>
              <div class="action-cell">
                <button 
                  v-if="!report.claimedBy" 
                  @click="claimReport(report.id)" 
                  class="btn btn-secondary btn-xs"
                >
                  Claim (30m)
                </button>
                <button 
                  v-else 
                  @click="resolveReport(report.id, 'UPHELD')" 
                  class="btn btn-danger btn-xs"
                >
                  Takedown
                </button>
                <button 
                  v-if="report.claimedBy" 
                  @click="resolveReport(report.id, 'DISMISSED')" 
                  class="btn btn-outline btn-xs"
                >
                  Dismiss
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

const currentTab = ref('PENDING');

const reports = ref([
  {
    id: 'REP-2026-0912',
    type: 'DMCA_COPYRIGHT',
    appName: 'Brand Cloner Game',
    packageName: 'com.clone.brandgame',
    reporterEmail: 'legal@brandowner.com',
    summary: 'Unauthorized reproduction of copyrighted 3D character assets in app binary.',
    claimedBy: null as string | null,
    remainingLeaseMin: 0,
    status: 'PENDING'
  },
  {
    id: 'REP-2026-0913',
    type: 'MALWARE_PHISHING',
    appName: 'Free Wifi Connect Master',
    packageName: 'com.wifi.freemaster',
    reporterEmail: 'security-research@netsec.org',
    summary: 'Exfiltrates Wi-Fi network passwords and device contacts to an unencrypted C2 IP.',
    claimedBy: 'admin_triage@bookingnetwork.com',
    remainingLeaseMin: 22,
    status: 'CLAIMED'
  }
]);

const filteredReports = computed(() => {
  return reports.value.filter(r => r.status === currentTab.value);
});

function getReportCount(status: string): number {
  return reports.value.filter(r => r.status === status).length;
}

function claimReport(id: string) {
  const item = reports.value.find(r => r.id === id);
  if (item) {
    item.claimedBy = 'current_admin@bookingnetwork.com';
    item.remainingLeaseMin = 30;
    item.status = 'CLAIMED';
  }
}

function resolveReport(id: string, resolution: 'UPHELD' | 'DISMISSED') {
  const item = reports.value.find(r => r.id === id);
  if (item) {
    item.status = resolution === 'UPHELD' ? 'RESOLVED' : 'DISMISSED';
    alert(`Report ${id} marked as ${resolution}`);
  }
}
</script>

<style scoped>
.moderation-reports-view {
  padding: 24px;
  max-width: 1100px;
  margin: 0 auto;
}
.view-header {
  margin-bottom: 20px;
}
.subtitle {
  color: #64748b;
  font-size: 14px;
}
.tabs-container {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}
.tab-btn {
  padding: 8px 16px;
  border-radius: 6px;
  border: 1px solid #cbd5e1;
  background: white;
  cursor: pointer;
  font-size: 13px;
  font-weight: 600;
}
.tab-btn.active {
  background: #2563eb;
  color: white;
  border-color: #2563eb;
}
.table-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px;
}
.reports-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.reports-table th {
  text-align: left;
  padding: 10px 8px;
  border-bottom: 2px solid #e2e8f0;
  color: #475569;
}
.reports-table td {
  padding: 12px 8px;
  border-bottom: 1px solid #f1f5f9;
}
.summary-cell {
  max-width: 280px;
  line-height: 1.4;
}
.lease-box {
  display: flex;
  flex-direction: column;
}
.claimed-name { font-size: 11px; font-weight: 600; }
.lease-timer { color: #d97706; }
.badge {
  padding: 3px 8px;
  border-radius: 12px;
  font-size: 10px;
  font-weight: 700;
}
.badge-dmca_copyright { background: #eff6ff; color: #1d4ed8; }
.badge-malware_phishing { background: #fee2e2; color: #991b1b; }
.action-cell {
  display: flex;
  gap: 6px;
}
.btn {
  padding: 4px 10px;
  border-radius: 4px;
  cursor: pointer;
  border: none;
  font-weight: 600;
}
.btn-secondary { background: #f1f5f9; color: #334155; }
.btn-danger { background: #dc2626; color: white; }
.btn-outline { border: 1px solid #cbd5e1; background: transparent; color: #64748b; }
.btn-xs { font-size: 11px; }
.text-muted { color: #94a3b8; }
.text-xs { font-size: 11px; }
.font-mono { font-family: monospace; }
</style>
