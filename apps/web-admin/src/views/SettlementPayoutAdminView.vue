<template>
  <div class="settlement-payout-admin-view">
    <header class="view-header">
      <div>
        <h1>월간 개발자 정산 및 이중승인 지급 집행 (ADM-14)</h1>
        <p class="subtitle">정산 요청 생성과 2인 관리자 교차 승인 (Four-Eyes Principle)</p>
      </div>
      <span class="admin-pill">현재 로그인 관리자: {{ currentAdminId }}</span>
    </header>

    <div class="summary-grid">
      <div class="summary-card">
        <span class="summary-label">지급 요청 총액</span>
        <div class="summary-value">₩{{ totalRequested.toLocaleString() }}</div>
      </div>
      <div class="summary-card">
        <span class="summary-label">지급 완료</span>
        <div class="summary-value text-success">₩{{ totalPaid.toLocaleString() }}</div>
      </div>
      <div class="summary-card">
        <span class="summary-label">승인 대기 건</span>
        <div class="summary-value text-amber">{{ pendingCount }}건</div>
      </div>
    </div>

    <form class="create-card" @submit.prevent="createPayout">
      <h3>새 정산 지급 요청</h3>
      <div class="form-grid">
        <label>조직 ID <input v-model="form.orgId" required placeholder="org-..." /></label>
        <label>금액 (minor) <input v-model.number="form.amount" type="number" min="1" required /></label>
        <label>통화 <input v-model="form.currency" maxlength="3" placeholder="KRW" /></label>
      </div>
      <button type="submit" class="btn btn-primary btn-sm" :disabled="isActing">
        {{ isActing ? '생성 중…' : '지급 요청 생성' }}
      </button>
    </form>

    <div v-if="isLoading" class="loading-row">불러오는 중…</div>
    <div v-if="loadError" class="load-error">{{ loadError }}</div>
    <div v-if="actionError" class="load-error">{{ actionError }}</div>

    <div v-if="!isLoading && !loadError" class="list-card">
      <div v-if="payouts.length === 0" class="empty-row">지급 요청이 없습니다.</div>
      <table v-else class="payout-table">
        <thead>
          <tr>
            <th>정산 ID</th>
            <th>조직</th>
            <th>금액</th>
            <th>상태</th>
            <th>요청자 → 승인자</th>
            <th>승인/집행</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in payouts" :key="p.id">
            <td class="mono"><div class="payout-id">{{ p.id }}</div><div class="muted">{{ p.created_at }}</div></td>
            <td class="mono muted">{{ p.org_id }}</td>
            <td class="mono">₩{{ Number(p.amount).toLocaleString() }} {{ p.currency }}</td>
            <td><span class="badge" :class="p.status === 'PAID_COMPLETED' ? 'badge-success' : 'badge-pending'">{{ p.status }}</span></td>
            <td class="mono text-xs">{{ p.created_by }} → {{ p.approved_by || '대기 중' }}</td>
            <td>
              <button
                v-if="p.status === 'PENDING_APPROVAL'"
                @click="approve(p)"
                :disabled="isActing || p.created_by === currentAdminId"
                class="btn btn-success btn-xs"
                :title="p.created_by === currentAdminId ? '동일인 2차 승인 불가 (Self-Approval Prohibited)' : '최종 지급 집행'"
              >
                최종 지급 승인
              </button>
              <span v-else class="text-success mono text-xs">지급 완료</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="selfApprovalError" class="error-toast">
      <div class="error-title">이중 승인 위반 방지</div>
      <div>요청자와 동일한 관리자는 최종 지급을 승인할 수 없습니다.</div>
      <button @click="selfApprovalError = false" class="btn btn-danger btn-xs toast-btn">확인</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { getSharedClient, ApiClientError } from '@bookingnetwork/api-client';

const api = getSharedClient();

interface Payout {
  id: string;
  org_id: string;
  amount: number;
  currency: string;
  status: string;
  created_by: string;
  approved_by?: string;
  created_at: string;
  approved_at?: string;
}

const currentAdminId = ref('');
const payouts = ref<Payout[]>([]);
const form = ref({ orgId: '', amount: 0, currency: 'KRW' });
const selfApprovalError = ref(false);
const isLoading = ref(true);
const loadError = ref('');
const actionError = ref('');
const isActing = ref(false);

const totalRequested = computed(() => payouts.value.reduce((n, p) => n + Number(p.amount || 0), 0));
const totalPaid = computed(() => payouts.value.filter(p => p.status === 'PAID_COMPLETED').reduce((n, p) => n + Number(p.amount || 0), 0));
const pendingCount = computed(() => payouts.value.filter(p => p.status === 'PENDING_APPROVAL').length);

onMounted(async () => {
  try {
    const me = await api.getMe();
    currentAdminId.value = me.sub || me.email || '';
  } catch {
    currentAdminId.value = '';
  }
  await loadPayouts();
  isLoading.value = false;
});

async function loadPayouts() {
  loadError.value = '';
  try {
    payouts.value = await api.listSettlements() as Payout[];
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load payouts';
  }
}

async function createPayout() {
  isActing.value = true;
  actionError.value = '';
  try {
    await api.createSettlement({
      orgId: form.value.orgId.trim(),
      amount: form.value.amount,
      currency: form.value.currency.trim().toUpperCase() || 'KRW'
    });
    form.value = { orgId: '', amount: 0, currency: 'KRW' };
    await loadPayouts();
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to create payout';
  } finally {
    isActing.value = false;
  }
}

async function approve(p: Payout) {
  if (p.created_by === currentAdminId.value) {
    selfApprovalError.value = true;
    return;
  }
  isActing.value = true;
  actionError.value = '';
  try {
    await api.approveSettlement(p.id, currentAdminId.value, p.created_by);
    await loadPayouts();
  } catch (err) {
    if (err instanceof ApiClientError && (err.status === 403 || /self|dual|four-eyes/i.test(err.message))) {
      selfApprovalError.value = true;
    } else {
      actionError.value = err instanceof ApiClientError ? err.message : 'Failed to approve payout';
    }
  } finally {
    isActing.value = false;
  }
}
</script>

<style scoped>
.settlement-payout-admin-view {
  padding: 24px;
  max-width: 1100px;
  margin: 0 auto;
}
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 20px;
}
.view-header h1 {
  margin: 0 0 4px;
  font-size: 22px;
}
.subtitle {
  color: #64748b;
  font-size: 14px;
  margin: 0;
}
.admin-pill {
  padding: 6px 14px;
  background: #e0e7ff;
  color: #3730a3;
  border-radius: 9999px;
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}
.summary-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}
.summary-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px 20px;
}
.summary-label {
  font-size: 12px;
  color: #64748b;
  display: block;
  margin-bottom: 6px;
}
.summary-value {
  font-size: 22px;
  font-weight: 700;
  color: #0f172a;
}
.text-success { color: #16a34a; }
.text-amber { color: #d97706; }
.create-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 24px;
}
.create-card h3 { margin: 0 0 12px; }
.form-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 16px;
}
.form-grid label {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  font-weight: 600;
  color: #475569;
}
.form-grid input {
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 400;
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
.list-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px;
  overflow-x: auto;
}
.payout-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.payout-table th {
  text-align: left;
  padding: 10px 8px;
  border-bottom: 2px solid #e2e8f0;
  color: #475569;
  white-space: nowrap;
}
.payout-table td {
  padding: 12px 8px;
  border-bottom: 1px solid #f1f5f9;
  vertical-align: top;
}
.mono { font-family: ui-monospace, monospace; font-size: 12px; }
.muted { color: #64748b; }
.text-xs { font-size: 11px; }
.payout-id { font-weight: 700; color: #1d4ed8; }
.badge {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  white-space: nowrap;
}
.badge-success { background: #dcfce7; color: #166534; }
.badge-pending { background: #fef3c7; color: #92400e; }
.btn {
  padding: 6px 14px;
  border-radius: 6px;
  cursor: pointer;
  border: none;
  font-weight: 600;
}
.btn:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-primary { background: #2563eb; color: white; }
.btn-success { background: #16a34a; color: white; }
.btn-danger { background: #dc2626; color: white; }
.btn-xs { font-size: 11px; padding: 4px 10px; }
.btn-sm { font-size: 12px; }
.error-toast {
  position: fixed;
  bottom: 24px;
  right: 24px;
  background: #7f1d1d;
  border: 1px solid #dc2626;
  color: #fee2e2;
  padding: 16px;
  border-radius: 8px;
  font-size: 12px;
  max-width: 360px;
}
.error-title { font-weight: 700; margin-bottom: 4px; }
.toast-btn { margin-top: 8px; }
</style>
