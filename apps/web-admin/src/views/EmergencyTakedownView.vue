<template>
  <div class="emergency-takedown-view">
    <div class="view-header">
      <h2>Emergency Takedown & Global Kill Switch (ADM-07)</h2>
      <p class="subtitle">Immediate revocation, Edge CDN purge, and download suspension for compromised applications</p>
    </div>

    <!-- Active Kill Switch Warning Banner -->
    <div class="alert alert-danger-banner">
      <div class="banner-text">
        <strong>HIGH-IMPACT ADMINISTRATIVE ACTION</strong>
        <span>
          Activating the Kill Switch immediately blocks new download tokens and marks every update response for this package as DISTRIBUTION_SUSPENDED with an uninstall advisory.
        </span>
      </div>
    </div>

    <!-- Action Form -->
    <div class="action-card">
      <h3>Execute Emergency Revocation</h3>

      <div class="form-grid">
        <div class="form-group">
          <label>Target Application ID / Package Name</label>
          <input 
            type="text" 
            v-model="targetPackageName" 
            placeholder="e.g. com.suspicious.app or app-uuid"
            class="input-control" 
          />
        </div>

        <div class="form-group">
          <label>Scope of Revocation</label>
          <select v-model="revocationScope" class="input-control">
            <option value="ALL_RELEASES">All Releases (Entire App Takedown)</option>
            <option value="SPECIFIC_RELEASE">Specific Release Version Only</option>
          </select>
        </div>

        <div class="form-group col-span-2">
          <label>Mandatory Revocation Reason Code</label>
          <select v-model="reasonCode" class="input-control">
            <option value="">-- Select Valid Justification --</option>
            <option value="MALWARE_CONFIRMED">Malware / Trojan Activity Confirmed (ClamAV/YARA/Sandbox)</option>
            <option value="ZERO_DAY_EXPLOIT">Zero-Day Remote Code Execution (RCE) Vulnerability</option>
            <option value="DMCA_COPYRIGHT_TAKEDOWN">Legal DMCA / Trademark Takedown Order</option>
            <option value="UNAUTHORIZED_BILLING_FRAUD">Billing Fraud / Rogue In-App Subscription Loop</option>
            <option value="LEGAL_COURT_ORDER">Court Order / Government Regulatory Injunction</option>
          </select>
        </div>

        <div class="form-group col-span-2">
          <label>Detailed Evidence & Incident Report Reference</label>
          <textarea 
            v-model="evidenceNotes" 
            rows="3" 
            class="input-control"
            placeholder="Specify security incident ID, ticket URL, or CVE identifier..."
          ></textarea>
        </div>
      </div>

      <div v-if="actionError" class="load-error">{{ actionError }}</div>
      <div v-if="actionDone" class="action-done">{{ actionDone }}</div>
      <div class="form-actions">
        <button
          @click="executeKillSwitch"
          class="btn btn-danger-action"
          :disabled="!isValidForm || isExecuting"
          id="btnTriggerKillSwitch"
        >
          {{ isExecuting ? 'Executing Kill Switch...' : 'EXECUTE IMMEDIATE KILL SWITCH' }}
        </button>
      </div>
    </div>

    <!-- Active Revocation Ledger -->
    <div class="ledger-card">
      <h3>Active Revoked Applications & Releases</h3>
      <div v-if="isLoading" class="loading-row">Loading kill-switch ledger…</div>
      <div v-if="loadError" class="load-error">{{ loadError }}</div>
      <div v-if="!isLoading && !loadError && activeRevocations.length === 0" class="empty-row">No active kill-switches. All packages are servable.</div>
      <table v-if="activeRevocations.length > 0" class="takedown-table">
        <thead>
          <tr>
            <th>Package / App</th>
            <th>Reason</th>
            <th>Revoked By</th>
            <th>Timestamp</th>
            <th>Distribution</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in activeRevocations" :key="item.app_id">
            <td class="font-mono text-sm">{{ item.app_id }}</td>
            <td>{{ item.reason }}</td>
            <td>{{ item.triggered_by }}</td>
            <td class="font-mono text-xs">{{ item.triggered_at }}</td>
            <td><span class="badge badge-danger">SUSPENDED</span></td>
            <td>
              <button @click="restoreApp(item.app_id)" class="btn btn-outline btn-xs" :disabled="isExecuting">Restore</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { getSharedClient, ApiClientError } from '@bookingnetwork/api-client';

const api = getSharedClient();

interface KillSwitchState {
  app_id: string;
  is_triggered: boolean;
  reason: string;
  triggered_by: string;
  triggered_at: string;
}

const targetPackageName = ref('');
const revocationScope = ref('ALL_RELEASES');
const reasonCode = ref('');
const evidenceNotes = ref('');
const isExecuting = ref(false);
const isLoading = ref(true);
const loadError = ref('');
const actionError = ref('');
const actionDone = ref('');

const activeRevocations = ref<KillSwitchState[]>([]);

const isValidForm = computed(() => {
  return targetPackageName.value.trim().length > 3 &&
         reasonCode.value.length > 0 &&
         evidenceNotes.value.trim().length >= 10;
});

onMounted(async () => {
  await loadLedger();
  isLoading.value = false;
});

async function loadLedger() {
  loadError.value = '';
  try {
    const list = await api.listKillSwitches() as KillSwitchState[];
    activeRevocations.value = list.filter(s => s.is_triggered);
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load kill-switch ledger';
  }
}

async function executeKillSwitch() {
  if (!confirm(`CRITICAL WARNING: Are you completely certain you want to activate the Kill Switch on ${targetPackageName.value}?`)) {
    return;
  }

  isExecuting.value = true;
  actionError.value = '';
  actionDone.value = '';
  try {
    await api.triggerKillSwitch(targetPackageName.value.trim(), {
      scope: revocationScope.value,
      reasonCode: reasonCode.value,
      evidence: evidenceNotes.value.trim()
    });
    targetPackageName.value = '';
    reasonCode.value = '';
    evidenceNotes.value = '';
    actionDone.value = 'Kill Switch executed. Download tokens and updates for this package are now blocked.';
    await loadLedger();
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to execute kill switch';
  } finally {
    isExecuting.value = false;
  }
}

async function restoreApp(packageName: string) {
  if (!confirm(`Reactivate ${packageName}? Downloads and updates will resume.`)) {
    return;
  }
  isExecuting.value = true;
  actionError.value = '';
  actionDone.value = '';
  try {
    await api.clearKillSwitch(packageName);
    actionDone.value = `${packageName} reactivated.`;
    await loadLedger();
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to lift kill switch';
  } finally {
    isExecuting.value = false;
  }
}
</script>

<style scoped>
.emergency-takedown-view {
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
.alert-danger-banner {
  background: var(--bn-danger-wash);
  border: 1px solid var(--bn-danger-red);
  border-radius: var(--bn-radius-md);
  padding: 16px 20px;
  margin-bottom: 24px;
}
.banner-text strong {
  display: block;
  font-family: var(--bn-font-mono);
  font-size: 11px;
  letter-spacing: 0.1em;
  color: var(--bn-danger-red);
  margin-bottom: 6px;
}
.banner-text span {
  font-size: 13px;
  line-height: 1.65;
  color: var(--bn-text-secondary);
}
.action-card, .ledger-card {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 24px;
  margin-bottom: 24px;
}
.action-card {
  border-top: 3px solid var(--bn-danger-red);
}
.action-card h3, .ledger-card h3 {
  font-family: var(--bn-font-display);
  margin: 0 0 16px 0;
  font-size: 15px;
  font-weight: 800;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--bn-border-subtle);
}
.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 20px;
}
.col-span-2 { grid-column: span 2; }
.form-group label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  display: block;
  margin-bottom: 8px;
}
.input-control {
  width: 100%;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  padding: 10px 12px;
  font-family: var(--bn-font-sans);
  font-size: 13px;
  box-sizing: border-box;
}
.input-control:focus {
  outline: none;
  border-color: var(--bn-ink);
}
textarea.input-control {
  resize: vertical;
  line-height: 1.6;
}
.form-actions {
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid var(--bn-border-subtle);
  padding-top: 16px;
}
.btn-danger-action {
  background: var(--bn-danger-red);
  color: #fff;
  padding: 12px 28px;
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  font-weight: 800;
  letter-spacing: 0.04em;
  border: none;
  cursor: pointer;
}
.btn-danger-action:hover:not(:disabled) {
  filter: brightness(0.92);
}
.btn-danger-action:disabled { opacity: 0.4; cursor: not-allowed; }
.takedown-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.takedown-table th {
  text-align: left;
  padding: 10px 12px;
  border-bottom: 2px solid var(--bn-ink);
  font-family: var(--bn-font-mono);
  font-weight: 400;
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  white-space: nowrap;
}
.takedown-table td {
  padding: 12px;
  border-bottom: 1px solid var(--bn-border-subtle);
  vertical-align: middle;
}
.takedown-table tr:last-child td {
  border-bottom: none;
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
.badge-danger { border-color: var(--bn-danger-red); color: var(--bn-danger-red); }
.badge-success { border-color: var(--bn-trust-verified); color: var(--bn-trust-verified); }
.btn {
  font-family: var(--bn-font-sans);
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}
.btn-outline {
  border: 1px solid var(--bn-border-strong);
  background: transparent;
  color: var(--bn-text-primary);
  padding: 5px 12px;
  border-radius: var(--bn-radius-sm);
}
.btn-outline:hover:not(:disabled) {
  border-color: var(--bn-text-primary);
}
.btn-outline:disabled { opacity: 0.45; cursor: not-allowed; }
.btn-xs { font-size: 11px; }
.font-mono { font-family: var(--bn-font-mono); }
.text-sm { font-size: 12px; }
.text-xs { font-size: 11px; }
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
.action-done {
  padding: 12px 16px;
  background: var(--bn-trust-wash);
  border: 1px solid var(--bn-trust-verified);
  color: var(--bn-trust-verified);
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
@media (max-width: 680px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
  .col-span-2 {
    grid-column: span 1;
  }
}
</style>

