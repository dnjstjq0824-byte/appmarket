<template>
  <div class="orders-refunds-admin-view p-6 bg-slate-900 text-slate-100 min-h-screen">
    <header class="mb-8">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">결제·환불 분쟁 및 복식부기 원장 대조 (ADM-13)</h1>
          <p class="text-sm text-slate-400 mt-1">유료 앱 주문 내역, 고객 분쟁 강제 환불 집행 및 복식부기 대차평형(Zero-Sum) 무결성 감사</p>
        </div>
        <div class="flex items-center gap-3">
          <span :class="isLedgerBalanced ? 'bg-emerald-950/80 border-emerald-500/40 text-emerald-400' : 'bg-rose-950/80 border-rose-500/40 text-rose-400'" class="px-3 py-1 border text-xs rounded-full font-mono font-semibold">
            {{ isLedgerBalanced ? '원장 대차평형 합치: 100% 정상' : '원장 불일치 경보 발령' }}
          </span>
        </div>
      </div>
    </header>

    <!-- Ledger Balance Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">총 차변 (Total Debits)</span>
        <div class="text-2xl font-bold mt-1 text-sky-400">₩{{ totalDebits.toLocaleString() }}</div>
      </div>
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">총 대변 (Total Credits)</span>
        <div class="text-2xl font-bold mt-1 text-indigo-400">₩{{ totalCredits.toLocaleString() }}</div>
      </div>
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">대차 차액 (Discrepancy)</span>
        <div class="text-2xl font-bold mt-1 font-mono text-emerald-400">₩{{ (totalDebits - totalCredits).toLocaleString() }}</div>
        <div class="text-[11px] text-slate-400 mt-1">오차 0원 필수 (Zero Rounding Leakage)</div>
      </div>
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">분쟁 환불 대기</span>
        <div class="text-2xl font-bold mt-1 text-amber-400">{{ disputedOrders.length }}건</div>
      </div>
    </div>

    <!-- Orders & Disputes Table -->
    <div class="bg-slate-800/90 border border-slate-700 rounded-lg overflow-hidden mb-8">
      <div class="p-4 border-b border-slate-700 flex justify-between items-center">
        <h2 class="text-base font-semibold text-slate-200">주문 내역 및 환불 분쟁 심사 목록</h2>
        <div class="flex gap-2">
          <input 
            v-model="searchKeyword"
            placeholder="주문 ID / 사용자 이메일 검색..." 
            class="px-3 py-1 bg-slate-900 border border-slate-700 rounded text-xs text-slate-200 placeholder-slate-500 w-64 focus:outline-none focus:border-sky-500"
          />
        </div>
      </div>

      <table class="w-full text-left text-sm text-slate-300">
        <thead class="bg-slate-900/60 text-xs uppercase tracking-wider text-slate-400 border-b border-slate-700">
          <tr>
            <th class="p-3">주문 ID / 일시</th>
            <th class="p-3">고객 (사용자)</th>
            <th class="p-3">앱 / 금액</th>
            <th class="p-3">PG 거래번호</th>
            <th class="p-3">결제 상태</th>
            <th class="p-3">소유 권한 (Entitlement)</th>
            <th class="p-3 text-right">분쟁 처리</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-700/60">
          <tr v-for="order in filteredOrders" :key="order.id" class="hover:bg-slate-750 transition">
            <td class="p-3 font-mono text-xs">
              <div class="text-sky-400 font-bold">{{ order.id }}</div>
              <div class="text-slate-400">{{ order.createdAt }}</div>
            </td>
            <td class="p-3 text-xs">
              <div class="text-slate-200">{{ order.userEmail }}</div>
              <div class="text-slate-500 text-[11px]">{{ order.userId }}</div>
            </td>
            <td class="p-3 text-xs">
              <div class="text-slate-200 font-medium">{{ order.appName }}</div>
              <div class="font-mono text-slate-300">₩{{ order.amountWon.toLocaleString() }}</div>
            </td>
            <td class="p-3 text-xs font-mono text-slate-400">{{ order.pgTid }}</td>
            <td class="p-3 text-xs">
              <span :class="getOrderStatusBadge(order.status)" class="px-2 py-0.5 rounded text-[11px] font-semibold">
                {{ order.status }}
              </span>
            </td>
            <td class="p-3 text-xs font-mono">
              <span :class="order.entitlementActive ? 'text-emerald-400' : 'text-slate-500'" class="font-bold">
                {{ order.entitlementActive ? 'ACTIVE (이용가능)' : 'REVOKED (회수됨)' }}
              </span>
            </td>
            <td class="p-3 text-right text-xs">
              <button 
                v-if="order.status === 'PAID' || order.status === 'DISPUTED'"
                @click="openRefundModal(order)"
                class="px-2.5 py-1 bg-rose-600 hover:bg-rose-500 text-white rounded font-medium text-xs transition"
              >
                강제 전액 환불
              </button>
              <span v-else class="text-slate-500 text-xs">-</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Refund Action Modal -->
    <div v-if="selectedOrder" class="fixed inset-0 bg-black/70 flex items-center justify-center p-4 z-50">
      <div class="bg-slate-800 border border-slate-700 rounded-lg p-6 max-w-md w-full">
        <h3 class="text-lg font-bold text-slate-100 mb-2">분쟁 강제 환불 집행</h3>
        <p class="text-xs text-slate-400 mb-4">
          환불을 승인하면 PG사에 취소 요청이 전송되고, 고객의 앱 다운로드 권한(Entitlement)이 즉시 회수되며, 복식부기 원장에 역분개 전표가 기록됩니다.
        </p>
        <div class="bg-slate-900 p-3 rounded text-xs space-y-1 mb-4 font-mono">
          <div>주문 ID: {{ selectedOrder.id }}</div>
          <div>환불 금액: ₩{{ selectedOrder.amountWon.toLocaleString() }}</div>
          <div>대상 고객: {{ selectedOrder.userEmail }}</div>
        </div>
        <div class="mb-4">
          <label class="block text-xs font-semibold text-slate-300 mb-1">환불 사유 (감사 로그 필수 기록)</label>
          <select v-model="refundReason" class="w-full bg-slate-900 border border-slate-700 rounded p-2 text-xs text-slate-200">
            <option value="CUSTOMER_REQUEST_COOLING_OFF">14일 이내 청약철회 (미사용/오구매)</option>
            <option value="APP_DEFECT_CRITICAL">앱 치명적 결함 및 실행 불가</option>
            <option value="MALWARE_SECURITY_TAKEDOWN">악성코드 보안 삭제에 따른 직권 전액 환불</option>
            <option value="UNAUTHORIZED_PAYMENT">명의 도용 및 비정상 부정 결제</option>
          </select>
        </div>
        <div class="flex justify-end gap-2">
          <button @click="selectedOrder = null" class="px-3 py-1.5 bg-slate-700 hover:bg-slate-600 text-slate-300 rounded text-xs">
            취소
          </button>
          <button @click="executeRefund" class="px-3 py-1.5 bg-rose-600 hover:bg-rose-500 text-white font-medium rounded text-xs">
            환불 승인 및 권한 즉시 회수
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

interface OrderItem {
  id: string;
  createdAt: string;
  userId: string;
  userEmail: string;
  appName: string;
  amountWon: number;
  pgTid: string;
  status: 'PAID' | 'DISPUTED' | 'REFUNDED';
  entitlementActive: boolean;
}

const searchKeyword = ref('');
const selectedOrder = ref<OrderItem | null>(null);
const refundReason = ref('CUSTOMER_REQUEST_COOLING_OFF');

const orders = ref<OrderItem[]>([
  {
    id: 'ord-0192ce-7701',
    createdAt: '2026-10-27 10:20',
    userId: 'usr-4411',
    userEmail: 'devuser1@example.com',
    appName: 'Vector Pro Illustrator',
    amountWon: 15000,
    pgTid: 'toss_pay_20261027102001',
    status: 'PAID',
    entitlementActive: true
  },
  {
    id: 'ord-0192ce-7702',
    createdAt: '2026-10-27 11:45',
    userId: 'usr-4412',
    userEmail: 'buyer2@example.com',
    appName: 'Hyper Sonic DAW Audio',
    amountWon: 33000,
    pgTid: 'toss_pay_20261027114502',
    status: 'DISPUTED',
    entitlementActive: true
  },
  {
    id: 'ord-0192ce-7703',
    createdAt: '2026-10-26 18:30',
    userId: 'usr-4413',
    userEmail: 'gamer3@example.com',
    appName: 'Retro Dungeon RPG',
    amountWon: 5500,
    pgTid: 'toss_pay_20261026183003',
    status: 'REFUNDED',
    entitlementActive: false
  }
]);

const totalDebits = ref(53500);
const totalCredits = ref(53500);
const isLedgerBalanced = computed(() => totalDebits.value === totalCredits.value);

const disputedOrders = computed(() => orders.value.filter(o => o.status === 'DISPUTED'));

const filteredOrders = computed(() => {
  if (!searchKeyword.value) return orders.value;
  const kw = searchKeyword.value.toLowerCase();
  return orders.value.filter(o => o.id.toLowerCase().includes(kw) || o.userEmail.toLowerCase().includes(kw));
});

function getOrderStatusBadge(status: string) {
  switch (status) {
    case 'PAID': return 'bg-emerald-950 text-emerald-400 border border-emerald-500/30';
    case 'DISPUTED': return 'bg-amber-950 text-amber-400 border border-amber-500/30';
    case 'REFUNDED': return 'bg-rose-950 text-rose-400 border border-rose-500/30';
    default: return 'bg-slate-800 text-slate-400';
  }
}

function openRefundModal(order: OrderItem) {
  selectedOrder.value = order;
}

function executeRefund() {
  if (!selectedOrder.value) return;
  selectedOrder.value.status = 'REFUNDED';
  selectedOrder.value.entitlementActive = false;
  selectedOrder.value = null;
}
</script>

