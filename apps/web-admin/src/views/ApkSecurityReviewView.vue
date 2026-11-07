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
.queue-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  font-size: 13px;
  color: var(--bn-text-secondary);
}
.queue-bar label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  flex-shrink: 0;
}
.queue-bar select {
  padding: 8px 12px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  font-family: var(--bn-font-mono);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  min-width: 320px;
  flex: 1;
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
  padding: 28px 24px;
  color: var(--bn-text-muted);
  font-size: 14px;
  text-align: center;
  border: 1px dashed var(--bn-border-strong);
  border-radius: var(--bn-radius-md);
}
.security-card {
  display: flex;
  align-items: center;
  gap: 24px;
  padding: 20px 24px;
  border-radius: var(--bn-radius-md);
  margin-bottom: 24px;
  border: 1px solid var(--bn-border-subtle);
  background: var(--bn-bg-surface);
}
.verdict-clean { background: var(--bn-trust-wash); border-color: var(--bn-trust-verified); }
.verdict-clean .card-info h3 { color: var(--bn-trust-verified); }
.verdict-warning { background: var(--bn-warning-wash); border-color: var(--bn-warning-amber); }
.verdict-warning .card-info h3 { color: var(--bn-warning-amber); }
.verdict-danger { background: var(--bn-danger-wash); border-color: var(--bn-danger-red); }
.verdict-danger .card-info h3 { color: var(--bn-danger-red); }
.card-info { flex: 1; }
.card-info h3 {
  font-family: var(--bn-font-display);
  font-size: 17px;
  font-weight: 800;
  margin: 0 0 4px 0;
}
.card-info p {
  margin: 0;
  font-size: 13px;
  color: var(--bn-text-secondary);
  font-family: var(--bn-font-mono);
}
.risk-score-gauge {
  text-align: right;
  flex-shrink: 0;
  padding-left: 24px;
  border-left: 1px solid var(--bn-border-subtle);
}
.score-label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  display: block;
  margin-bottom: 4px;
}
.score-number {
  font-family: var(--bn-font-display);
  font-size: 34px;
  font-weight: 900;
  letter-spacing: -0.02em;
}
.font-mono { font-family: var(--bn-font-mono); }
.text-danger { color: var(--bn-danger-red); }
.text-warning { color: var(--bn-warning-amber); }
.text-success { color: var(--bn-trust-verified); }
.results-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 24px;
}
.result-box, .perm-diff-box {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 20px;
}
.perm-diff-box {
  margin-bottom: 24px;
}
.result-box h4, .perm-diff-box h4 {
  font-family: var(--bn-font-display);
  margin: 0 0 14px 0;
  font-size: 14px;
  font-weight: 800;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--bn-border-subtle);
}
.status-clean {
  color: var(--bn-trust-verified);
  font-size: 13px;
  font-weight: 600;
}
.status-threat {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.threat-item {
  background: var(--bn-danger-wash);
  border: 1px solid var(--bn-danger-red);
  padding: 8px 12px;
  border-radius: var(--bn-radius-sm);
  font-size: 12px;
  font-family: var(--bn-font-mono);
}
.threat-item strong {
  display: block;
  color: var(--bn-danger-red);
}
.cert-item { margin-bottom: 12px; }
.cert-item label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  display: block;
  margin-bottom: 4px;
}
.cert-item .val { font-size: 14px; font-weight: 600; }
.cert-item .val.font-mono { font-size: 13px; }
.perm-list {
  margin: 6px 0 0;
  padding-left: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.perm-list li {
  font-size: 12px;
  font-family: var(--bn-font-mono);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  padding: 4px 10px;
  width: fit-content;
}
.raw-report {
  background: var(--bn-bg-well);
  border: 1px solid var(--bn-border-subtle);
  color: var(--bn-text-secondary);
  font-family: var(--bn-font-mono);
  font-size: 12px;
  line-height: 1.7;
  padding: 14px;
  border-radius: var(--bn-radius-sm);
  overflow-x: auto;
  white-space: pre-wrap;
  margin: 0;
}
.decision-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  flex-wrap: wrap;
}
.btn {
  padding: 10px 20px;
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  text-decoration: none;
  display: inline-block;
}
.btn-primary {
  background: var(--bn-action-bg);
  color: var(--bn-action-text);
  border: none;
}
.btn-primary:hover {
  background: var(--bn-action-bg-hover);
}
.btn-secondary {
  background: transparent;
  border: 1px solid var(--bn-border-subtle);
  color: var(--bn-text-secondary);
}
.btn-secondary:hover {
  border-color: var(--bn-border-strong);
  color: var(--bn-text-primary);
}
@media (max-width: 760px) {
  .results-grid {
    grid-template-columns: 1fr;
  }
  .security-card {
    flex-direction: column;
    align-items: flex-start;
  }
  .risk-score-gauge {
    border-left: none;
    padding-left: 0;
    text-align: left;
  }
}
</style>
