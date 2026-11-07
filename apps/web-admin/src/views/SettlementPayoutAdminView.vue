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
  max-width: 1100px;
  margin: 0 auto;
}
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 2px solid var(--bn-ink);
}
.view-header h1 {
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
.admin-pill {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  padding: 6px 14px;
  border: 1px solid var(--bn-seal);
  color: var(--bn-seal);
  border-radius: var(--bn-radius-pill);
  white-space: nowrap;
  flex-shrink: 0;
}
.summary-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 24px;
}
.summary-card {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 16px 20px;
}
.summary-label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.06em;
  color: var(--bn-text-muted);
  display: block;
  margin-bottom: 8px;
}
.summary-value {
  font-family: var(--bn-font-display);
  font-size: 26px;
  font-weight: 900;
  letter-spacing: -0.02em;
}
.text-success { color: var(--bn-trust-verified); }
.text-amber { color: var(--bn-warning-amber); }
.create-card {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 20px;
  margin-bottom: 24px;
}
.create-card h3 {
  font-family: var(--bn-font-display);
  font-size: 15px;
  font-weight: 800;
  margin: 0 0 14px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--bn-border-subtle);
}
.form-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 16px;
}
.form-grid label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.06em;
  color: var(--bn-text-muted);
}
.form-grid input {
  padding: 9px 12px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  font-family: var(--bn-font-sans);
  font-size: 13px;
}
.form-grid input:focus {
  outline: none;
  border-color: var(--bn-ink);
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
.list-card {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 20px;
  overflow-x: auto;
}
.payout-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  min-width: 860px;
}
.payout-table th {
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
.payout-table td {
  padding: 12px;
  border-bottom: 1px solid var(--bn-border-subtle);
  vertical-align: middle;
}
.payout-table tr:last-child td {
  border-bottom: none;
}
.mono { font-family: var(--bn-font-mono); font-size: 12px; }
.muted { color: var(--bn-text-muted); }
.text-xs { font-size: 11px; }
.payout-id { font-family: var(--bn-font-sans); font-weight: 700; }
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
.badge-success { border-color: var(--bn-trust-verified); color: var(--bn-trust-verified); }
.badge-pending { border-color: var(--bn-warning-amber); color: var(--bn-warning-amber); }
.btn {
  padding: 8px 16px;
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
}
.btn:disabled { opacity: 0.45; cursor: not-allowed; }
.btn-primary {
  background: var(--bn-action-bg);
  color: var(--bn-action-text);
  border: none;
}
.btn-primary:hover:not(:disabled) {
  background: var(--bn-action-bg-hover);
}
.btn-success {
  background: transparent;
  border: 1px solid var(--bn-trust-verified);
  color: var(--bn-trust-verified);
}
.btn-success:hover:not(:disabled) {
  background: var(--bn-trust-wash);
}
.btn-danger {
  background: var(--bn-danger-red);
  border: none;
  color: #fff;
}
.btn-danger:hover:not(:disabled) {
  filter: brightness(0.92);
}
.btn-xs { font-size: 11px; padding: 5px 12px; }
.btn-sm { font-size: 12px; }
.error-toast {
  position: fixed;
  bottom: 24px;
  right: 24px;
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-danger-red);
  border-top: 3px solid var(--bn-danger-red);
  box-shadow: var(--bn-shadow-pop);
  padding: 16px 18px;
  border-radius: var(--bn-radius-md);
  font-size: 12px;
  color: var(--bn-text-secondary);
  max-width: 360px;
  z-index: 70;
}
.error-title {
  font-weight: 800;
  font-size: 13px;
  color: var(--bn-danger-red);
  margin-bottom: 4px;
}
.toast-btn { margin-top: 10px; }
@media (max-width: 860px) {
  .summary-grid, .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
