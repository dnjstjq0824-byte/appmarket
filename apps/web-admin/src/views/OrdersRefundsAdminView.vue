<template>
  <div class="orders-refunds-admin-view">
    <header class="view-header">
      <div>
        <h1>결제·환불 분쟁 및 복식부기 원장 대조 (ADM-13)</h1>
        <p class="subtitle">유료 앱 주문 내역, 고객 분쟁 강제 환불 집행 및 원장 대차평형 무결성 감사</p>
      </div>
      <span class="ledger-pill" :class="isLedgerBalanced ? 'ok' : 'bad'">
        {{ isLedgerBalanced ? '원장 대차평형 합치: 정상' : '원장 불일치 경보 발령' }}
      </span>
    </header>

    <div class="summary-grid">
      <div class="summary-card">
        <span class="summary-label">총 결제액 (Gross)</span>
        <div class="summary-value">₩{{ totalGross.toLocaleString() }}</div>
      </div>
      <div class="summary-card">
        <span class="summary-label">세금 + 수수료 + 개발사분 합계</span>
        <div class="summary-value">₩{{ totalParts.toLocaleString() }}</div>
      </div>
      <div class="summary-card">
        <span class="summary-label">대차 차액 (Discrepancy)</span>
        <div class="summary-value mono">₩{{ (totalGross - totalParts).toLocaleString() }}</div>
        <div class="summary-note">오차 0원 필수</div>
      </div>
      <div class="summary-card">
        <span class="summary-label">환불 주문</span>
        <div class="summary-value text-amber">{{ refundedCount }}건</div>
      </div>
    </div>

    <div class="list-card">
      <div class="list-header">
        <h2>주문 내역 및 환불 분쟁 심사 목록</h2>
        <div class="filters">
          <input v-model="filterUser" placeholder="사용자 ID 필터" class="filter-input" @change="loadOrders" />
          <input v-model="filterOrg" placeholder="조직 ID 필터" class="filter-input" @change="loadOrders" />
          <input v-model="searchKeyword" placeholder="주문 ID 검색..." class="filter-input" />
        </div>
      </div>

      <div v-if="isLoading" class="loading-row">불러오는 중…</div>
      <div v-if="loadError" class="load-error">{{ loadError }}</div>
      <div v-if="actionError" class="load-error">{{ actionError }}</div>

      <div v-if="!isLoading && !loadError && filteredOrders.length === 0" class="empty-row">주문이 없습니다.</div>
      <table v-if="filteredOrders.length > 0" class="orders-table">
        <thead>
          <tr>
            <th>주문 ID / 일시</th>
            <th>고객</th>
            <th>앱 / 금액</th>
            <th>세금 / 수수료 / 개발사분</th>
            <th>결제 상태</th>
            <th>분쟁 처리</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="order in filteredOrders" :key="order.id">
            <td class="mono">
              <div class="order-id">{{ order.id }}</div>
              <div class="muted">{{ order.created_at }}</div>
            </td>
            <td class="mono muted">{{ order.user_id }}</td>
            <td>
              <div class="mono muted">{{ order.app_id }}</div>
              <div class="mono">₩{{ Number(order.gross_amount).toLocaleString() }} {{ order.currency }}</div>
            </td>
            <td class="mono text-xs">
              <div>세금: ₩{{ Number(order.tax_amount).toLocaleString() }}</div>
              <div>수수료: ₩{{ Number(order.commission_amount).toLocaleString() }}</div>
              <div>개발사: ₩{{ Number(order.dev_payout_amount).toLocaleString() }}</div>
            </td>
            <td>
              <span class="badge" :class="order.status === 'PAID' ? 'badge-success' : order.status === 'REFUNDED' ? 'badge-danger' : 'badge-muted'">{{ order.status }}</span>
            </td>
            <td>
              <button
                v-if="order.status !== 'REFUNDED'"
                @click="openRefundModal(order)"
                class="btn btn-danger btn-xs"
              >
                강제 전액 환불
              </button>
              <span v-else class="muted text-xs">-</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="selectedOrder" class="modal-backdrop">
      <div class="modal">
        <h3>분쟁 강제 환불 집행</h3>
        <p class="modal-desc">환불을 승인하면 고객의 앱 다운로드 권한이 즉시 회수되고 복식부기 원장에 역분개 전표가 기록됩니다.</p>
        <div class="modal-facts mono">
          <div>주문 ID: {{ selectedOrder.id }}</div>
          <div>환불 금액: ₩{{ Number(selectedOrder.gross_amount).toLocaleString() }}</div>
          <div>대상 고객: {{ selectedOrder.user_id }}</div>
        </div>
        <label class="field">환불 사유 (감사 로그 필수 기록)
          <select v-model="refundReason">
            <option value="CUSTOMER_REQUEST_COOLING_OFF">14일 이내 청약철회 (미사용/오구매)</option>
            <option value="APP_DEFECT_CRITICAL">앱 치명적 결함 및 실행 불가</option>
            <option value="MALWARE_SECURITY_TAKEDOWN">악성코드 보안 삭제에 따른 직권 전액 환불</option>
            <option value="UNAUTHORIZED_PAYMENT">명의 도용 및 비정상 부정 결제</option>
          </select>
        </label>
        <div class="modal-actions">
          <button @click="selectedOrder = null" class="btn btn-secondary btn-sm">취소</button>
          <button @click="executeRefund" class="btn btn-danger btn-sm" :disabled="isActing">
            {{ isActing ? '집행 중…' : '환불 승인 및 권한 즉시 회수' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { getSharedClient, ApiClientError, type OrderDTO } from '@bookingnetwork/api-client';

const api = getSharedClient();

const orders = ref<OrderDTO[]>([]);
const searchKeyword = ref('');
const filterUser = ref('');
const filterOrg = ref('');
const selectedOrder = ref<OrderDTO | null>(null);
const refundReason = ref('CUSTOMER_REQUEST_COOLING_OFF');
const isLoading = ref(true);
const loadError = ref('');
const actionError = ref('');
const isActing = ref(false);

const totalGross = computed(() => orders.value.reduce((n, o) => n + Number(o.gross_amount || 0), 0));
const totalParts = computed(() => orders.value.reduce((n, o) => n + Number(o.tax_amount || 0) + Number(o.commission_amount || 0) + Number(o.dev_payout_amount || 0), 0));
const isLedgerBalanced = computed(() => totalGross.value === totalParts.value);
const refundedCount = computed(() => orders.value.filter(o => o.status === 'REFUNDED').length);

const filteredOrders = computed(() => {
  if (!searchKeyword.value) return orders.value;
  const kw = searchKeyword.value.toLowerCase();
  return orders.value.filter(o => o.id.toLowerCase().includes(kw) || String(o.user_id).toLowerCase().includes(kw));
});

onMounted(async () => {
  await loadOrders();
  isLoading.value = false;
});

async function loadOrders() {
  loadError.value = '';
  try {
    const params: Record<string, string> = {};
    if (filterUser.value.trim()) params.user = filterUser.value.trim();
    if (filterOrg.value.trim()) params.org = filterOrg.value.trim();
    orders.value = await api.listAdminOrders(params);
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load orders';
  }
}

function openRefundModal(order: OrderDTO) {
  selectedOrder.value = order;
}

async function executeRefund() {
  if (!selectedOrder.value) return;
  isActing.value = true;
  actionError.value = '';
  try {
    await api.adminRefundOrder(selectedOrder.value.id, refundReason.value);
    selectedOrder.value = null;
    await loadOrders();
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to execute refund';
  } finally {
    isActing.value = false;
  }
}
</script>

<style scoped>
.orders-refunds-admin-view {
  padding: 24px;
  max-width: 1200px;
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
.ledger-pill {
  padding: 6px 14px;
  border-radius: 9999px;
  font-size: 12px;
  font-weight: 700;
  white-space: nowrap;
}
.ledger-pill.ok { background: #dcfce7; color: #166534; }
.ledger-pill.bad { background: #fee2e2; color: #991b1b; }
.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
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
.summary-note {
  font-size: 11px;
  color: #64748b;
}
.text-amber { color: #d97706; }
.mono { font-family: ui-monospace, monospace; font-size: 12px; }
.muted { color: #64748b; }
.text-xs { font-size: 11px; }
.order-id { font-weight: 700; color: #1d4ed8; }
.list-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px;
  overflow-x: auto;
}
.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.list-header h2 {
  margin: 0;
  font-size: 16px;
}
.filters {
  display: flex;
  gap: 8px;
}
.filter-input {
  padding: 6px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 12px;
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
.orders-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.orders-table th {
  text-align: left;
  padding: 10px 8px;
  border-bottom: 2px solid #e2e8f0;
  color: #475569;
  white-space: nowrap;
}
.orders-table td {
  padding: 12px 8px;
  border-bottom: 1px solid #f1f5f9;
  vertical-align: top;
}
.badge {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  white-space: nowrap;
}
.badge-success { background: #dcfce7; color: #166534; }
.badge-danger { background: #fee2e2; color: #991b1b; }
.badge-muted { background: #f1f5f9; color: #64748b; }
.btn {
  padding: 6px 14px;
  border-radius: 6px;
  cursor: pointer;
  border: none;
  font-weight: 600;
}
.btn:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-danger { background: #dc2626; color: white; }
.btn-secondary { background: #f1f5f9; color: #334155; }
.btn-xs { font-size: 11px; padding: 4px 10px; }
.btn-sm { font-size: 12px; }
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  z-index: 50;
}
.modal {
  background: white;
  border-radius: 12px;
  padding: 24px;
  max-width: 440px;
  width: 100%;
}
.modal h3 { margin: 0 0 8px; }
.modal-desc { font-size: 12px; color: #64748b; margin: 0 0 12px; }
.modal-facts {
  background: #f8fafc;
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  font-weight: 600;
  color: #475569;
  margin-bottom: 16px;
}
.field select {
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 400;
}
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
