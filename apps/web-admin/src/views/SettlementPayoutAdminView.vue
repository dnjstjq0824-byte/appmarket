<template>
  <div class="settlement-payout-admin-view p-6 bg-slate-900 text-slate-100 min-h-screen">
    <header class="mb-8">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">월간 개발자 정산 및 이중승인(Dual-Control) 지급 집행 (ADM-14)</h1>
          <p class="text-sm text-slate-400 mt-1">14일 에스크로 성숙 완료 정산 집행, 15%/10% 플랫폼 수수료 및 2인 관리자 교차 승인</p>
        </div>
        <div class="flex items-center gap-2">
          <span class="px-3 py-1 bg-sky-950/80 border border-sky-500/40 text-sky-400 text-xs rounded-full font-mono">
            현재 로그인 관리자: {{ currentAdminId }}
          </span>
        </div>
      </div>
    </header>

    <!-- Settlement Overview Metrics -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">이번 달 정산 대상 총액</span>
        <div class="text-2xl font-bold mt-1 text-slate-200">₩{{ totalGrossSales.toLocaleString() }}</div>
      </div>
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">플랫폼 수수료 매출 (15%/10%)</span>
        <div class="text-2xl font-bold mt-1 text-emerald-400">₩{{ totalCommission.toLocaleString() }}</div>
      </div>
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">개발자 지급 예정액</span>
        <div class="text-2xl font-bold mt-1 text-sky-400">₩{{ totalDevPayout.toLocaleString() }}</div>
      </div>
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">2차 승인 대기 건</span>
        <div class="text-2xl font-bold mt-1 text-amber-400">{{ pendingSecondApprovalCount }}건</div>
      </div>
    </div>

    <!-- Settlements Table -->
    <div class="bg-slate-800/90 border border-slate-700 rounded-lg overflow-hidden">
      <div class="p-4 border-b border-slate-700 flex justify-between items-center">
        <h2 class="text-base font-semibold text-slate-200">개발자 조직별 월간 정산 내역 및 2인 이중승인 현황</h2>
      </div>

      <table class="w-full text-left text-sm text-slate-300">
        <thead class="bg-slate-900/60 text-xs uppercase tracking-wider text-slate-400 border-b border-slate-700">
          <tr>
            <th class="p-3">정산 ID / 정산월</th>
            <th class="p-3">개발사 / 사업자번호</th>
            <th class="p-3">총 결제액 (Gross)</th>
            <th class="p-3">수수료 / 부가세</th>
            <th class="p-3">실 지급액 (Net Payout)</th>
            <th class="p-3">14일 에스크로 상태</th>
            <th class="p-3">승인 진행 단계</th>
            <th class="p-3 text-right">승인/집행</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-700/60">
          <tr v-for="s in settlements" :key="s.id" class="hover:bg-slate-750 transition">
            <td class="p-3 font-mono text-xs">
              <div class="font-bold text-sky-400">{{ s.id }}</div>
              <div class="text-slate-400">{{ s.billingMonth }}</div>
            </td>
            <td class="p-3 text-xs">
              <div class="text-slate-200 font-semibold">{{ s.orgName }}</div>
              <div class="font-mono text-slate-400 text-[11px]">{{ s.businessNumber }}</div>
            </td>
            <td class="p-3 text-xs font-mono">
              ₩{{ s.grossSalesWon.toLocaleString() }}
            </td>
            <td class="p-3 text-xs font-mono">
              <div>수수료: ₩{{ s.commissionWon.toLocaleString() }}</div>
              <div class="text-slate-400">VAT: ₩{{ s.vatWon.toLocaleString() }}</div>
            </td>
            <td class="p-3 text-xs font-mono font-bold text-emerald-400">
              ₩{{ s.netPayoutWon.toLocaleString() }}
            </td>
            <td class="p-3 text-xs font-mono">
              <span :class="s.escrowMatured ? 'text-emerald-400' : 'text-amber-400'">
                {{ s.escrowMatured ? '성숙 완료 (출금 가능)' : '에스크로 보호 중' }}
              </span>
            </td>
            <td class="p-3 text-xs">
              <div class="space-y-1">
                <div class="text-[11px] flex items-center gap-1">
                  <span :class="s.firstApprover ? 'text-emerald-400' : 'text-slate-500'">●</span>
                  1차: {{ s.firstApprover || '대기 중' }}
                </div>
                <div class="text-[11px] flex items-center gap-1">
                  <span :class="s.secondApprover ? 'text-emerald-400' : 'text-slate-500'">●</span>
                  2차: {{ s.secondApprover || '대기 중' }}
                </div>
              </div>
            </td>
            <td class="p-3 text-right text-xs">
              <!-- Step 1 Approval -->
              <button 
                v-if="!s.firstApprover"
                @click="approveFirst(s)"
                :disabled="!s.escrowMatured || s.complianceHold"
                class="px-2.5 py-1 bg-sky-600 hover:bg-sky-500 disabled:bg-slate-700 disabled:text-slate-500 text-white rounded font-medium text-xs transition"
              >
                1차 승인
              </button>

              <!-- Step 2 Final Approval (Dual Control: Must NOT be the same as 1st approver!) -->
              <button 
                v-else-if="s.firstApprover && !s.secondApprover"
                @click="approveSecond(s)"
                :disabled="s.firstApprover === currentAdminId"
                class="px-2.5 py-1 bg-emerald-600 hover:bg-emerald-500 disabled:bg-slate-700 disabled:text-slate-500 text-white rounded font-medium text-xs transition"
                :title="s.firstApprover === currentAdminId ? '동일인 2차 승인 불가 (Self-Approval Prohibited)' : '최종 지급 집행'"
              >
                2차 최종 지급
              </button>

              <span v-else class="text-emerald-400 font-mono font-bold text-xs">
                지급 완료
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Error Banner for Self-Approval Attempt -->
    <div v-if="selfApprovalError" class="fixed bottom-6 right-6 bg-rose-900 border border-rose-600 text-rose-100 p-4 rounded-lg shadow-xl text-xs max-w-md">
      <div class="font-bold text-sm mb-1">이중 승인 위반 방지</div>
      <div>1차 승인자와 동일한 관리자는 2차 최종 지급을 승인할 수 없습니다. (TC-COM-006 준수)</div>
      <button @click="selfApprovalError = false" class="mt-2 px-2 py-1 bg-rose-700 hover:bg-rose-600 rounded text-white text-[11px]">확인</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

interface SettlementItem {
  id: string;
  billingMonth: string;
  orgName: string;
  businessNumber: string;
  grossSalesWon: number;
  commissionWon: number;
  vatWon: number;
  netPayoutWon: number;
  escrowMatured: boolean;
  complianceHold: boolean;
  firstApprover?: string;
  secondApprover?: string;
}

const currentAdminId = ref('admin_finance_02');
const selfApprovalError = ref(false);

const settlements = ref<SettlementItem[]>([
  {
    id: 'stl-202610-001',
    billingMonth: '2026-10',
    orgName: 'OmniWork Corp',
    businessNumber: '120-81-12345',
    grossSalesWon: 11000000,
    commissionWon: 1500000,
    vatWon: 1000000,
    netPayoutWon: 8500000,
    escrowMatured: true,
    complianceHold: false,
    firstApprover: 'admin_finance_01',
    secondApprover: undefined
  },
  {
    id: 'stl-202610-002',
    billingMonth: '2026-10',
    orgName: 'Creative Labs',
    businessNumber: '214-82-67890',
    grossSalesWon: 5500000,
    commissionWon: 750000,
    vatWon: 500000,
    netPayoutWon: 4250000,
    escrowMatured: true,
    complianceHold: false,
    firstApprover: undefined,
    secondApprover: undefined
  },
  {
    id: 'stl-202610-003',
    billingMonth: '2026-10',
    orgName: 'ZeroTrust Security',
    businessNumber: '107-86-54321',
    grossSalesWon: 22000000,
    commissionWon: 3000000,
    vatWon: 2000000,
    netPayoutWon: 17000000,
    escrowMatured: false, // In 14-day escrow hold
    complianceHold: false,
    firstApprover: undefined,
    secondApprover: undefined
  }
]);

const totalGrossSales = computed(() => settlements.value.reduce((sum, s) => sum + s.grossSalesWon, 0));
const totalCommission = computed(() => settlements.value.reduce((sum, s) => sum + s.commissionWon, 0));
const totalDevPayout = computed(() => settlements.value.reduce((sum, s) => sum + s.netPayoutWon, 0));
const pendingSecondApprovalCount = computed(() => settlements.value.filter(s => s.firstApprover && !s.secondApprover).length);

function approveFirst(s: SettlementItem) {
  s.firstApprover = currentAdminId.value;
}

function approveSecond(s: SettlementItem) {
  if (s.firstApprover === currentAdminId.value) {
    selfApprovalError.value = true;
    return;
  }
  s.secondApprover = currentAdminId.value;
}
</script>

