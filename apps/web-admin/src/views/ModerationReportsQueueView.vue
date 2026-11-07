<template>
  <div class="moderation-reports-view">
    <div class="view-header">
      <h2>Community Moderation & Developer Appeals (ADM-08)</h2>
      <p class="subtitle">Abuse report triage and developer appeal decisions</p>
    </div>

    <div class="tabs-container">
      <button
        v-for="tab in ['REPORTS', 'APPEALS']"
        :key="tab"
        class="tab-btn"
        :class="{ active: currentSection === tab }"
        @click="currentSection = tab"
      >
        {{ tab }}
      </button>
    </div>

    <div v-if="isLoading" class="loading-row">불러오는 중…</div>
    <div v-if="loadError" class="load-error">{{ loadError }}</div>
    <div v-if="actionError" class="load-error">{{ actionError }}</div>

    <template v-if="!isLoading && !loadError">
      <template v-if="currentSection === 'REPORTS'">
        <div class="tabs-container">
          <button
            v-for="tab in ['OPEN', 'IN_INVESTIGATION', 'ACTION_TAKEN', 'DISMISSED']"
            :key="tab"
            class="tab-btn tab-sm"
            :class="{ active: currentTab === tab }"
            @click="currentTab = tab; loadReports()"
          >
            {{ tab }} ({{ tabCounts[tab] || 0 }})
          </button>
        </div>

        <div class="table-card">
          <div v-if="reports.length === 0" class="empty-row">해당 상태의 신고가 없습니다.</div>
          <table v-else class="reports-table">
            <thead>
              <tr>
                <th>Report ID</th>
                <th>Category</th>
                <th>Target</th>
                <th>Reporter</th>
                <th>Priority</th>
                <th>Description</th>
                <th>Assignee</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="report in reports" :key="report.id">
                <td class="font-mono text-sm">{{ report.id }}</td>
                <td><span class="badge badge-cat">{{ report.report_category }}</span></td>
                <td>
                  <strong>{{ report.target_type }}</strong>
                  <div class="font-mono text-xs text-muted">{{ report.target_id }}</div>
                </td>
                <td>{{ report.reporter_email }}</td>
                <td><span class="badge" :class="report.priority === 'P0_URGENT' ? 'badge-danger' : 'badge-muted'">{{ report.priority }}</span></td>
                <td class="summary-cell">{{ report.description }}</td>
                <td>{{ report.assigned_moderator_id || 'Unclaimed' }}</td>
                <td>
                  <div class="action-cell">
                    <button
                      v-if="report.status === 'OPEN'"
                      @click="claimReport(report.id)"
                      class="btn btn-secondary btn-xs"
                      :disabled="isActing"
                    >
                      Claim
                    </button>
                    <template v-if="report.status === 'OPEN' || report.status === 'IN_INVESTIGATION'">
                      <input v-model="resolveNotes[report.id]" placeholder="조치 메모" class="notes-input" />
                      <button @click="resolveReport(report.id, true)" class="btn btn-danger btn-xs" :disabled="isActing">조치</button>
                      <button @click="resolveReport(report.id, false)" class="btn btn-outline btn-xs" :disabled="isActing">기각</button>
                    </template>
                    <span v-else class="text-muted text-xs">{{ report.resolution_notes }}</span>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>

      <template v-else>
        <div class="table-card">
          <div v-if="appeals.length === 0" class="empty-row">접수된 이의제기가 없습니다.</div>
          <table v-else class="reports-table">
            <thead>
              <tr>
                <th>Appeal ID</th>
                <th>Org / Release</th>
                <th>Reason</th>
                <th>Status</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="appeal in appeals" :key="appeal.id">
                <td class="font-mono text-sm">{{ appeal.id }}</td>
                <td>
                  <strong>{{ appeal.org_id }}</strong>
                  <div class="font-mono text-xs text-muted">{{ appeal.release_id }}</div>
                </td>
                <td class="summary-cell">{{ appeal.reason }}</td>
                <td><span class="badge" :class="appeal.status === 'PENDING' ? 'badge-pending' : appeal.status === 'APPROVED' ? 'badge-success' : 'badge-danger'">{{ appeal.status }}</span></td>
                <td>
                  <div v-if="appeal.status === 'PENDING'" class="action-cell">
                    <input v-model="appealNotes[appeal.id]" placeholder="결정 메모" class="notes-input" />
                    <button @click="decideAppeal(appeal.id, 'APPROVED')" class="btn btn-success btn-xs" :disabled="isActing">인용</button>
                    <button @click="decideAppeal(appeal.id, 'REJECTED')" class="btn btn-danger btn-xs" :disabled="isActing">기각</button>
                  </div>
                  <span v-else class="text-muted text-xs">{{ appeal.decision_notes }}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getSharedClient, ApiClientError } from '@bookingnetwork/api-client';

const api = getSharedClient();

interface Report {
  id: string;
  reporter_email: string;
  target_type: string;
  target_id: string;
  report_category: string;
  description: string;
  priority: string;
  status: string;
  assigned_moderator_id?: string;
  resolution_notes?: string;
}

interface Appeal {
  id: string;
  org_id: string;
  release_id: string;
  reason: string;
  status: string;
  decision_notes?: string;
}

const currentSection = ref('REPORTS');
const currentTab = ref('OPEN');
const reports = ref<Report[]>([]);
const appeals = ref<Appeal[]>([]);
const tabCounts = ref<Record<string, number>>({});
const resolveNotes = ref<Record<string, string>>({});
const appealNotes = ref<Record<string, string>>({});
const isLoading = ref(true);
const loadError = ref('');
const actionError = ref('');
const isActing = ref(false);

onMounted(async () => {
  await Promise.all([loadReports(), loadAppeals(), loadCounts()]);
  isLoading.value = false;
});

async function loadReports() {
  try {
    reports.value = await api.listReports(currentTab.value) as Report[];
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load reports';
  }
}

async function loadAppeals() {
  try {
    appeals.value = await api.listAppeals() as Appeal[];
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load appeals';
  }
}

async function loadCounts() {
  for (const tab of ['OPEN', 'IN_INVESTIGATION', 'ACTION_TAKEN', 'DISMISSED']) {
    try {
      const list = await api.listReports(tab) as Report[];
      tabCounts.value[tab] = list.length;
    } catch {
      tabCounts.value[tab] = 0;
    }
  }
}

async function claimReport(id: string) {
  isActing.value = true;
  actionError.value = '';
  try {
    await api.claimReport(id);
    await Promise.all([loadReports(), loadCounts()]);
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to claim report';
  } finally {
    isActing.value = false;
  }
}

async function resolveReport(id: string, actionTaken: boolean) {
  isActing.value = true;
  actionError.value = '';
  try {
    await api.resolveReport(id, actionTaken, (resolveNotes.value[id] || '').trim());
    await Promise.all([loadReports(), loadCounts()]);
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to resolve report';
  } finally {
    isActing.value = false;
  }
}

async function decideAppeal(id: string, decision: string) {
  isActing.value = true;
  actionError.value = '';
  try {
    await api.decideAppeal(id, decision, (appealNotes.value[id] || '').trim());
    await loadAppeals();
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to decide appeal';
  } finally {
    isActing.value = false;
  }
}
</script>

<style scoped>
.moderation-reports-view {
  max-width: 1200px;
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
.tabs-container {
  display: flex;
  gap: 0;
  margin-bottom: 16px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  overflow: hidden;
  width: fit-content;
  max-width: 100%;
  flex-wrap: wrap;
}
.tab-btn {
  padding: 8px 16px;
  border: none;
  border-left: 1px solid var(--bn-border-subtle);
  background: transparent;
  color: var(--bn-text-secondary);
  cursor: pointer;
  font-family: var(--bn-font-mono);
  font-size: 12px;
}
.tab-btn:first-child {
  border-left: none;
}
.tab-btn.active {
  background: var(--bn-ink);
  color: var(--bn-bg-canvas);
  font-weight: 700;
}
.tab-sm {
  font-size: 11px;
  padding: 7px 12px;
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
.table-card {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 16px;
  overflow-x: auto;
}
.reports-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  min-width: 960px;
}
.reports-table th {
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
.reports-table td {
  padding: 12px;
  border-bottom: 1px solid var(--bn-border-subtle);
  vertical-align: middle;
}
.reports-table tr:last-child td {
  border-bottom: none;
}
.summary-cell {
  max-width: 280px;
  line-height: 1.55;
  color: var(--bn-text-secondary);
}
.notes-input {
  padding: 6px 10px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  font-family: var(--bn-font-sans);
  font-size: 12px;
  min-width: 120px;
}
.notes-input:focus {
  outline: none;
  border-color: var(--bn-ink);
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
.badge-cat { border-color: var(--bn-seal); color: var(--bn-seal); }
.badge-danger { border-color: var(--bn-danger-red); color: var(--bn-danger-red); }
.badge-success { border-color: var(--bn-trust-verified); color: var(--bn-trust-verified); }
.badge-pending { border-color: var(--bn-warning-amber); color: var(--bn-warning-amber); }
.badge-muted { border-color: var(--bn-border-subtle); color: var(--bn-text-muted); }
.action-cell {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  align-items: center;
}
.btn {
  padding: 5px 12px;
  border-radius: var(--bn-radius-sm);
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
}
.btn:disabled { opacity: 0.45; cursor: not-allowed; }
.btn-secondary {
  background: transparent;
  border: 1px solid var(--bn-border-subtle);
  color: var(--bn-text-secondary);
}
.btn-secondary:hover:not(:disabled) {
  border-color: var(--bn-border-strong);
  color: var(--bn-text-primary);
}
.btn-danger {
  background: var(--bn-danger-red);
  border: none;
  color: #fff;
}
.btn-danger:hover:not(:disabled) {
  filter: brightness(0.92);
}
.btn-success {
  background: transparent;
  border: 1px solid var(--bn-trust-verified);
  color: var(--bn-trust-verified);
}
.btn-success:hover:not(:disabled) {
  background: var(--bn-trust-wash);
}
.btn-outline {
  border: 1px solid var(--bn-border-strong);
  background: transparent;
  color: var(--bn-text-primary);
}
.btn-outline:hover:not(:disabled) {
  border-color: var(--bn-text-primary);
}
.btn-xs { font-size: 11px; }
.text-muted { color: var(--bn-text-muted); }
.text-xs { font-size: 11px; }
.text-sm { font-size: 12px; }
.font-mono { font-family: var(--bn-font-mono); }
</style>
