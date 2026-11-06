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
  padding: 24px;
  max-width: 1200px;
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
.tab-sm {
  font-size: 12px;
  padding: 6px 12px;
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
.table-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px;
  overflow-x: auto;
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
  white-space: nowrap;
}
.reports-table td {
  padding: 12px 8px;
  border-bottom: 1px solid #f1f5f9;
  vertical-align: top;
}
.summary-cell {
  max-width: 280px;
  line-height: 1.4;
}
.notes-input {
  padding: 4px 8px;
  border: 1px solid #cbd5e1;
  border-radius: 4px;
  font-size: 12px;
  min-width: 120px;
}
.badge {
  padding: 3px 8px;
  border-radius: 12px;
  font-size: 10px;
  font-weight: 700;
  white-space: nowrap;
}
.badge-cat { background: #eff6ff; color: #1d4ed8; }
.badge-danger { background: #fee2e2; color: #991b1b; }
.badge-success { background: #dcfce7; color: #166534; }
.badge-pending { background: #e0e7ff; color: #3730a3; }
.badge-muted { background: #f1f5f9; color: #64748b; }
.action-cell {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
.btn {
  padding: 4px 10px;
  border-radius: 4px;
  cursor: pointer;
  border: none;
  font-weight: 600;
  white-space: nowrap;
}
.btn:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: #f1f5f9; color: #334155; }
.btn-danger { background: #dc2626; color: white; }
.btn-success { background: #16a34a; color: white; }
.btn-outline { border: 1px solid #cbd5e1; background: transparent; color: #64748b; }
.btn-xs { font-size: 11px; }
.text-muted { color: #94a3b8; }
.text-xs { font-size: 11px; }
.font-mono { font-family: monospace; }
</style>
