<template>
  <div class="apk-security-review-view">
    <div class="view-header">
      <h2>Binary Security & Vulnerability Review (ADM-05)</h2>
      <p class="subtitle">Multi-engine malware analysis, permission risk delta, and certificate integrity verification</p>
    </div>

    <!-- Security Summary Card -->
    <div class="security-card" :class="overallVerdictClass">
      <div class="card-icon">{{ verdictIcon }}</div>
      <div class="card-info">
        <h3>Security Verdict: {{ scanReport.verdict }}</h3>
        <p>{{ verdictDescription }}</p>
      </div>
      <div class="risk-score-gauge">
        <span class="score-label">Permission Risk Score</span>
        <span class="score-number font-mono" :class="riskScoreClass">{{ scanReport.permissionRiskScore }}/100</span>
      </div>
    </div>

    <!-- Scan Results Grid -->
    <div class="results-grid">
      <!-- Antivirus & YARA -->
      <div class="result-box">
        <h4>1. Static Malware & YARA Signatures</h4>
        <div v-if="scanReport.malwareDetections.length === 0" class="status-clean">
          ✓ No malware or trojan signatures detected (ClamAV & YARA clean)
        </div>
        <div v-else class="status-threat">
          <div v-for="threat in scanReport.malwareDetections" :key="threat.ruleName" class="threat-item">
            <strong>⚠️ {{ threat.ruleName }}</strong>
            <span>Severity: {{ threat.severity }} | Category: {{ threat.category }}</span>
          </div>
        </div>
      </div>

      <!-- Certificate Blacklist -->
      <div class="result-box">
        <h4>2. Certificate Reputation & Lineage</h4>
        <div class="cert-item">
          <label>Signing Scheme</label>
          <div class="val font-mono">v2 + v3 Verified (No v1 fallback vulnerability)</div>
        </div>
        <div class="cert-item">
          <label>Debug Certificate</label>
          <div class="val text-success">✓ False (Production signed)</div>
        </div>
        <div class="cert-item">
          <label>Known Blacklist Check</label>
          <div class="val" :class="scanReport.isCertBlacklisted ? 'text-danger' : 'text-success'">
            {{ scanReport.isCertBlacklisted ? '🚨 Blacklisted Malicious Certificate' : '✓ Clean (Not in blacklists)' }}
          </div>
        </div>
      </div>
    </div>

    <!-- Dangerous Permission Diff -->
    <div class="perm-diff-box">
      <h4>3. Dangerous Permission Delta (vs Previous Release)</h4>
      <p class="perm-desc">Highlights permissions newly requested in this build that access sensitive user data.</p>
      
      <table class="perm-table">
        <thead>
          <tr>
            <th>Permission</th>
            <th>Type</th>
            <th>Delta</th>
            <th>Risk Level</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="perm in scanReport.permissions" :key="perm.name">
            <td class="font-mono">{{ perm.name }}</td>
            <td>{{ perm.group }}</td>
            <td>
              <span class="badge" :class="perm.isNew ? 'badge-new' : 'badge-existing'">
                {{ perm.isNew ? '+ Added' : 'Unchanged' }}
              </span>
            </td>
            <td>
              <span class="risk-tag" :class="'risk-' + perm.risk.toLowerCase()">{{ perm.risk }}</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Decision Controls -->
    <div class="decision-actions">
      <button @click="quarantineArtifact" class="btn btn-danger">
        Quarantine & Reject Binary
      </button>
      <button 
        @click="approveSecurity" 
        class="btn btn-primary"
        :disabled="scanReport.malwareDetections.length > 0 || scanReport.isCertBlacklisted"
      >
        Approve Security Clearance
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

const scanReport = ref({
  verdict: 'MANUAL_REVIEW_REQUIRED',
  permissionRiskScore: 45,
  isCertBlacklisted: false,
  malwareDetections: [] as Array<{ ruleName: string; severity: string; category: string }>,
  permissions: [
    { name: 'android.permission.INTERNET', group: 'Network', isNew: false, risk: 'LOW' },
    { name: 'android.permission.ACCESS_FINE_LOCATION', group: 'Location', isNew: true, risk: 'HIGH' },
    { name: 'android.permission.POST_NOTIFICATIONS', group: 'Notifications', isNew: false, risk: 'LOW' },
    { name: 'android.permission.CAMERA', group: 'Hardware', isNew: true, risk: 'HIGH' }
  ]
});

const overallVerdictClass = computed(() => {
  if (scanReport.value.malwareDetections.length > 0 || scanReport.value.isCertBlacklisted) return 'verdict-danger';
  if (scanReport.value.permissionRiskScore >= 70) return 'verdict-warning';
  return 'verdict-clean';
});

const verdictIcon = computed(() => {
  if (scanReport.value.malwareDetections.length > 0 || scanReport.value.isCertBlacklisted) return '🚨';
  if (scanReport.value.permissionRiskScore >= 70) return '⚠️';
  return '🛡️';
});

const verdictDescription = computed(() => {
  if (scanReport.value.malwareDetections.length > 0) return 'Critical threat detected. Quarantine required.';
  if (scanReport.value.permissionRiskScore >= 70) return 'High permission risk detected. Review justification.';
  return 'Static signatures clean. Permission delta acceptable.';
});

const riskScoreClass = computed(() => {
  if (scanReport.value.permissionRiskScore >= 70) return 'text-danger';
  if (scanReport.value.permissionRiskScore >= 40) return 'text-warning';
  return 'text-success';
});

async function quarantineArtifact() {
  const reason = prompt('Enter reason for security quarantine:');
  if (!reason) return;
  alert('Artifact quarantined and developer notified.');
}

async function approveSecurity() {
  alert('Security clearance approved.');
}
</script>

<style scoped>
.apk-security-review-view {
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
.security-card {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 24px;
}
.verdict-clean { background: #f0fdf4; border: 1px solid #bbf7d0; color: #166534; }
.verdict-warning { background: #fffbeb; border: 1px solid #fef3c7; color: #92400e; }
.verdict-danger { background: #fef2f2; border: 1px solid #fecaca; color: #991b1b; }
.card-icon { font-size: 36px; }
.card-info { flex: 1; }
.card-info h3 { margin: 0 0 4px 0; }
.card-info p { margin: 0; font-size: 13px; }
.risk-score-gauge { text-align: right; }
.score-label { font-size: 11px; text-transform: uppercase; display: block; margin-bottom: 2px; }
.score-number { font-size: 26px; font-weight: 700; }
.text-danger { color: #dc2626; }
.text-warning { color: #d97706; }
.text-success { color: #16a34a; }
.results-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 24px;
}
.result-box, .perm-diff-box {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 20px;
}
.result-box h4, .perm-diff-box h4 { margin: 0 0 12px 0; font-size: 15px; }
.status-clean { color: #166534; font-size: 13px; font-weight: 500; }
.status-threat { display: flex; flex-direction: column; gap: 8px; }
.threat-item { background: #fef2f2; border-left: 3px solid #dc2626; padding: 8px; border-radius: 4px; font-size: 12px; }
.threat-item strong { display: block; color: #991b1b; }
.cert-item { margin-bottom: 10px; }
.cert-item label { font-size: 11px; color: #64748b; text-transform: uppercase; display: block; }
.cert-item .val { font-size: 13px; font-weight: 500; }
.perm-desc { font-size: 12px; color: #64748b; margin-bottom: 16px; }
.perm-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.perm-table th { text-align: left; padding: 8px; border-bottom: 2px solid #e2e8f0; color: #475569; }
.perm-table td { padding: 10px 8px; border-bottom: 1px solid #f1f5f9; }
.badge { padding: 2px 6px; border-radius: 4px; font-size: 11px; font-weight: 600; }
.badge-new { background: #fef3c7; color: #b45309; }
.badge-existing { background: #f1f5f9; color: #64748b; }
.risk-tag { padding: 2px 8px; border-radius: 4px; font-size: 11px; font-weight: 600; }
.risk-high { background: #fee2e2; color: #991b1b; }
.risk-low { background: #f0fdf4; color: #166534; }
.decision-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}
.btn {
  padding: 10px 18px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  border: none;
}
.btn-primary { background: #2563eb; color: white; }
.btn-primary:disabled { opacity: 0.4; cursor: not-allowed; }
.btn-danger { background: #dc2626; color: white; }
</style>
