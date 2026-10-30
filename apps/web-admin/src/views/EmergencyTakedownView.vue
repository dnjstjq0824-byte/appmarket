<template>
  <div class="emergency-takedown-view">
    <div class="view-header">
      <h2>Emergency Takedown & Global Kill Switch (ADM-07)</h2>
      <p class="subtitle">Immediate revocation, Edge CDN purge, and download suspension for compromised applications</p>
    </div>

    <!-- Active Kill Switch Warning Banner -->
    <div class="alert alert-danger-banner">
      <div class="banner-icon">🚨</div>
      <div class="banner-text">
        <strong>HIGH-IMPACT ADMINISTRATIVE ACTION</strong>
        <span>
          Activating the Kill Switch immediately revokes distribution, invalidates all active download tokens within 100ms, purges CDN Edge caches globally, and triggers client-side uninstallation advisories.
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

      <div class="form-actions">
        <button 
          @click="executeKillSwitch" 
          class="btn btn-danger-action"
          :disabled="!isValidForm || isExecuting"
          id="btnTriggerKillSwitch"
        >
          {{ isExecuting ? 'Executing Kill Switch...' : '🚨 EXECUTE IMMEDIATE KILL SWITCH' }}
        </button>
      </div>
    </div>

    <!-- Active Revocation Ledger -->
    <div class="ledger-card">
      <h3>Active Revoked Applications & Releases</h3>
      <table class="takedown-table">
        <thead>
          <tr>
            <th>Application</th>
            <th>Package</th>
            <th>Reason Code</th>
            <th>Revoked By</th>
            <th>Timestamp</th>
            <th>Edge Purge</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in activeRevocations" :key="item.id">
            <td class="font-bold">{{ item.appName }}</td>
            <td class="font-mono text-sm">{{ item.packageName }}</td>
            <td><span class="badge badge-danger">{{ item.reasonCode }}</span></td>
            <td>{{ item.revokedBy }}</td>
            <td class="font-mono text-xs">{{ item.timestamp }}</td>
            <td><span class="badge badge-success">✓ Purged</span></td>
            <td>
              <button @click="restoreApp(item.id)" class="btn btn-outline btn-xs">Restore</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

const targetPackageName = ref('');
const revocationScope = ref('ALL_RELEASES');
const reasonCode = ref('');
const evidenceNotes = ref('');
const isExecuting = ref(false);

const activeRevocations = ref([
  {
    id: 'takedown-001',
    appName: 'Trojan Fake Updater',
    packageName: 'com.fraud.fakestore.update',
    reasonCode: 'MALWARE_CONFIRMED',
    revokedBy: 'admin_sec_ops@bookingnetwork.com',
    timestamp: '2026-10-30 18:22:10 UTC'
  }
]);

const isValidForm = computed(() => {
  return targetPackageName.value.trim().length > 3 &&
         reasonCode.value.length > 0 &&
         evidenceNotes.value.trim().length >= 10;
});

async function executeKillSwitch() {
  if (!confirm(`CRITICAL WARNING: Are you completely certain you want to activate the Kill Switch on ${targetPackageName.value}?`)) {
    return;
  }

  isExecuting.value = true;
  try {
    const res = await fetch(`/api/v1/admin/security/kill-switch/${targetPackageName.value}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        scope: revocationScope.value,
        reasonCode: reasonCode.value,
        evidence: evidenceNotes.value
      })
    });

    if (res.ok) {
      activeRevocations.value.unshift({
        id: `takedown-${Date.now()}`,
        appName: targetPackageName.value,
        packageName: targetPackageName.value,
        reasonCode: reasonCode.value,
        revokedBy: 'admin_sec_ops@bookingnetwork.com',
        timestamp: new Date().toISOString()
      });
      targetPackageName.value = '';
      reasonCode.value = '';
      evidenceNotes.value = '';
      alert('Kill Switch executed successfully. Downloads blocked and CDN purged.');
    }
  } catch (err) {
    console.error(err);
  } finally {
    isExecuting.value = false;
  }
}

async function restoreApp(id: string) {
  if (confirm('Reactivate this application?')) {
    activeRevocations.value = activeRevocations.value.filter(item => item.id !== id);
  }
}
</script>

<style scoped>
.emergency-takedown-view {
  padding: 24px;
  max-width: 1000px;
  margin: 0 auto;
}
.view-header {
  margin-bottom: 20px;
}
.subtitle {
  color: #64748b;
  font-size: 14px;
}
.alert-danger-banner {
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-left: 4px solid #dc2626;
  border-radius: 6px;
  padding: 16px;
  display: flex;
  gap: 16px;
  align-items: center;
  margin-bottom: 24px;
}
.banner-icon { font-size: 28px; }
.banner-text strong { display: block; color: #991b1b; margin-bottom: 4px; }
.banner-text span { font-size: 13px; color: #7f1d1d; line-height: 1.4; }
.action-card, .ledger-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 24px;
  margin-bottom: 24px;
}
.action-card h3, .ledger-card h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
}
.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 20px;
}
.col-span-2 { grid-column: span 2; }
.form-group label {
  font-size: 12px;
  text-transform: uppercase;
  color: #64748b;
  font-weight: 600;
  display: block;
  margin-bottom: 6px;
}
.input-control {
  width: 100%;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  padding: 10px;
  font-family: inherit;
  font-size: 13px;
  box-sizing: border-box;
}
.form-actions {
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid #f1f5f9;
  padding-top: 16px;
}
.btn-danger-action {
  background: #dc2626;
  color: white;
  padding: 12px 24px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 700;
  border: none;
  cursor: pointer;
}
.btn-danger-action:disabled { opacity: 0.4; cursor: not-allowed; }
.takedown-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.takedown-table th {
  text-align: left;
  padding: 10px 8px;
  border-bottom: 2px solid #e2e8f0;
  color: #475569;
}
.takedown-table td {
  padding: 12px 8px;
  border-bottom: 1px solid #f1f5f9;
}
.badge {
  padding: 3px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
}
.badge-danger { background: #fee2e2; color: #991b1b; }
.badge-success { background: #dcfce7; color: #166534; }
.btn-outline { border: 1px solid #cbd5e1; background: transparent; padding: 4px 8px; border-radius: 4px; cursor: pointer; }
.btn-xs { font-size: 11px; }
.font-mono { font-family: monospace; }
</style>

