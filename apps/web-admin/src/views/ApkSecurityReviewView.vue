<template>
  <div class="apk-security-review-view">
    <div class="view-header">
      <h2>Binary Security & Vulnerability Review (ADM-05)</h2>
      <p class="subtitle">Multi-engine malware analysis, permission risk delta, and certificate integrity verification</p>
    </div>

    <div class="queue-bar">
      <label for="scanSelect">Stored scan reports ({{ scans.length }})</label>
      <select id="scanSelect" v-model="activeArtifactId" @change="loadReport">
        <option v-for="s in scans" :key="s.id" :value="s.artifact_id">
          {{ s.artifact_id }} — {{ s.scan_verdict }} ({{ s.scanned_at }})
        </option>
      </select>
    </div>
    <div v-if="isLoading" class="loading-row">Loading scan reports…</div>
    <div v-if="loadError" class="load-error">{{ loadError }}</div>

    <template v-if="!isLoading && !loadError">
      <div v-if="scans.length === 0" class="empty-row">No scan reports stored yet. Reports appear after binaries complete the upload pipeline.</div>

      <template v-else-if="report">
        <div class="security-card" :class="overallVerdictClass">
          <div class="card-info">
            <h3>Security Verdict: {{ report.scan_verdict }}</h3>
            <p>{{ verdictDescription }} Engine {{ report.scanner_engine_version }} · Scanned {{ report.scanned_at }}</p>
          </div>
          <div class="risk-score-gauge">
            <span class="score-label">Permission Risk Score</span>
            <span class="score-number font-mono" :class="riskScoreClass">{{ report.risk_score }}/100</span>
          </div>
        </div>

        <div class="results-grid">
          <div class="result-box">
            <h4>1. Static Malware & YARA Signatures</h4>
            <div class="cert-item">
              <label>ClamAV Status</label>
              <div class="val" :class="report.clamav_status === 'CLEAN' ? 'text-success' : 'text-danger'">{{ report.clamav_status }}</div>
            </div>
            <div v-if="(report.yara_matched_rules || []).length === 0" class="status-clean">
              No YARA rules matched.
            </div>
            <div v-else class="status-threat">
              <div v-for="rule in (report.yara_matched_rules || [])" :key="rule" class="threat-item">
                <strong>{{ rule }}</strong>
              </div>
            </div>
          </div>

          <div class="result-box">
            <h4>2. Brand Impersonation & Permissions</h4>
            <div class="cert-item">
              <label>Impersonation Similarity Score</label>
              <div class="val font-mono">{{ report.impersonation_similarity_score }}</div>
            </div>
            <div class="cert-item">
              <label>Dangerous Permissions ({{ (report.dangerous_permissions || []).length }})</label>
              <div v-if="(report.dangerous_permissions || []).length === 0" class="val text-success">None declared</div>
              <ul v-else class="perm-list">
                <li v-for="perm in (report.dangerous_permissions || [])" :key="perm" class="font-mono">{{ perm }}</li>
              </ul>
            </div>
          </div>
        </div>

        <div v-if="report.raw_report" class="perm-diff-box">
          <h4>3. Raw Scanner Report</h4>
          <pre class="raw-report">{{ report.raw_report }}</pre>
        </div>

        <div class="decision-actions">
          <router-link to="/security/kill-switch" class="btn btn-secondary">
            Open Kill-Switch Console
          </router-link>
          <router-link to="/apps/metadata" class="btn btn-primary">
            Continue to Metadata Review
          </router-link>
        </div>
      </template>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { getSharedClient, ApiClientError } from '@bookingnetwork/api-client';

const api = getSharedClient();

interface ScanReport {
  id: string;
  artifact_id: string;
  scanner_engine_version: string;
  clamav_status: string;
  yara_matched_rules: string[];
  dangerous_permissions: string[];
  impersonation_similarity_score: number;
  risk_score: number;
  scan_verdict: string;
  raw_report: string;
  scanned_at: string;
}

const scans = ref<ScanReport[]>([]);
const activeArtifactId = ref('');
const report = ref<ScanReport | null>(null);
const isLoading = ref(true);
const loadError = ref('');

const overallVerdictClass = computed(() => {
  if (!report.value) return 'verdict-clean';
  if (report.value.scan_verdict === 'INFECTED_QUARANTINED' || report.value.clamav_status === 'INFECTED') return 'verdict-danger';
  if (report.value.scan_verdict === 'CLEAN_HIGH_RISK_MANUAL_REVIEW' || report.value.risk_score >= 70) return 'verdict-warning';
  return 'verdict-clean';
});

const verdictDescription = computed(() => {
  if (!report.value) return '';
  if (report.value.scan_verdict === 'INFECTED_QUARANTINED') return 'Critical threat detected. Artifact is quarantined.';
  if (report.value.scan_verdict === 'CLEAN_HIGH_RISK_MANUAL_REVIEW') return 'Clean signatures but high risk. Manual review required.';
  if (report.value.scan_verdict === 'SCAN_ERROR') return 'Scanner error. Re-run the pipeline or inspect manually.';
  return 'Static signatures clean. Permission delta acceptable.';
});

const riskScoreClass = computed(() => {
  const score = report.value?.risk_score ?? 0;
  if (score >= 70) return 'text-danger';
  if (score >= 40) return 'text-warning';
  return 'text-success';
});

onMounted(async () => {
  try {
    scans.value = await api.listScans() as ScanReport[];
    if (scans.value.length > 0) {
      activeArtifactId.value = scans.value[scans.value.length - 1].artifact_id;
      await loadReport();
    }
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load scan reports';
  } finally {
    isLoading.value = false;
  }
});

async function loadReport() {
  if (!activeArtifactId.value) return;
  try {
    report.value = await api.getScanReport(activeArtifactId.value) as ScanReport;
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load scan report';
  }
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
.queue-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  font-size: 13px;
}
.queue-bar select {
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
  background: white;
  min-width: 320px;
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
  padding: 24px;
  color: #64748b;
  font-size: 14px;
  text-align: center;
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
.perm-list { margin: 6px 0 0; padding-left: 18px; font-size: 12px; }
.raw-report {
  background: #0f172a;
  color: #e2e8f0;
  font-size: 12px;
  padding: 12px;
  border-radius: 6px;
  overflow-x: auto;
  white-space: pre-wrap;
}
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
  text-decoration: none;
  display: inline-block;
}
.btn-primary { background: #2563eb; color: white; }
.btn-secondary { background: #f1f5f9; color: #334155; }
</style>
