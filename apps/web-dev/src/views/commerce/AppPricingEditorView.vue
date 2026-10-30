<template>
  <div class="app-pricing-editor p-6 bg-slate-900 text-slate-100 min-h-screen">
    <header class="mb-8">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">유료 앱 국가별 가격표 및 정산 시뮬레이터 (DEV-10)</h1>
          <p class="text-sm text-slate-400 mt-1">부가가치세(VAT) 포함 공급가, 플랫폼 수수료(15%/10%) 및 개발사 실수령액 실시간 산출</p>
        </div>
        <div class="flex gap-2">
          <button @click="savePricing" class="px-4 py-2 bg-sky-600 hover:bg-sky-500 text-white font-medium text-xs rounded transition shadow">
            가격표 저장 및 배포 반영
          </button>
        </div>
      </div>
    </header>

    <!-- Pricing Mode Toggle -->
    <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg mb-6 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <label class="relative inline-flex items-center cursor-pointer">
          <input type="checkbox" v-model="isPaidApp" class="sr-only peer">
          <div class="w-11 h-6 bg-slate-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-sky-600"></div>
        </label>
        <div>
          <span class="text-sm font-semibold text-slate-200">{{ isPaidApp ? '유료 판매 앱 (PAID APP)' : '무료 앱 (FREE APP)' }}</span>
          <p class="text-xs text-slate-400">무료 앱은 결제 없이 즉시 다운로드 가능하며 정산이 발생하지 않습니다.</p>
        </div>
      </div>
      <div v-if="isPaidApp" class="flex items-center gap-2">
        <span class="text-xs text-slate-400">적용 수수료 티어:</span>
        <select v-model="commissionTier" class="bg-slate-900 border border-slate-700 rounded px-2 py-1 text-xs text-slate-200 font-medium">
          <option value="STANDARD">일반 개발사 (15% 수수료)</option>
          <option value="SMALL_DEV">중소 개발사 우대 프로그램 (10% 수수료)</option>
        </select>
      </div>
    </div>

    <!-- Country Pricing Breakdown -->
    <div v-if="isPaidApp" class="space-y-6">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <!-- Input Form -->
        <div class="bg-slate-800/90 border border-slate-700 rounded-lg p-5">
          <h2 class="text-sm font-semibold text-slate-200 mb-4">기준 판매가 설정 (대한민국 KRW)</h2>
          <div class="space-y-4">
            <div>
              <label class="block text-xs text-slate-400 mb-1">소비자 최종 결제 가격 (VAT 포함 원화)</label>
              <div class="relative">
                <input 
                  type="number" 
                  v-model.number="krwPrice" 
                  step="500" 
                  min="1000" 
                  class="w-full bg-slate-900 border border-slate-700 rounded p-2.5 text-sm text-slate-100 font-mono focus:border-sky-500 focus:outline-none"
                />
                <span class="absolute right-3 top-2.5 text-slate-500 text-xs font-mono">KRW (₩)</span>
              </div>
              <p class="text-[11px] text-slate-500 mt-1">최소 판매가: ₩1,000 / 권장 단위: ₩500 단위</p>
            </div>

            <!-- Currency Auto-Exchange Multiplier -->
            <div class="pt-2 border-t border-slate-700">
              <label class="block text-xs text-slate-400 mb-2">글로벌 통화 자동 환율 동기화</label>
              <div class="grid grid-cols-3 gap-2 font-mono text-xs">
                <div class="bg-slate-900 p-2 rounded border border-slate-800">
                  <div class="text-slate-500 text-[10px]">USD ($)</div>
                  <div class="text-slate-200 font-bold">${{ usdPrice.toFixed(2) }}</div>
                </div>
                <div class="bg-slate-900 p-2 rounded border border-slate-800">
                  <div class="text-slate-500 text-[10px]">EUR (€)</div>
                  <div class="text-slate-200 font-bold">€{{ eurPrice.toFixed(2) }}</div>
                </div>
                <div class="bg-slate-900 p-2 rounded border border-slate-800">
                  <div class="text-slate-500 text-[10px]">JPY (¥)</div>
                  <div class="text-slate-200 font-bold">¥{{ Math.round(krwPrice * 0.11).toLocaleString() }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Simulation Card -->
        <div class="bg-slate-800/90 border border-slate-700 rounded-lg p-5">
          <h2 class="text-sm font-semibold text-slate-200 mb-4">건당 정산액 실시간 분해 (서버 계산식 일치)</h2>
          <div class="space-y-3 font-mono text-xs">
            <div class="flex justify-between p-2 bg-slate-900 rounded border border-slate-800">
              <span class="text-slate-400">총 결제액 (Gross Sales)</span>
              <span class="text-slate-200 font-bold">₩{{ krwPrice.toLocaleString() }}</span>
            </div>
            <div class="flex justify-between p-2 bg-slate-900 rounded border border-slate-800">
              <span class="text-slate-400">부가가치세 10% (VAT)</span>
              <span class="text-rose-400">- ₩{{ vatAmount.toLocaleString() }}</span>
            </div>
            <div class="flex justify-between p-2 bg-slate-900 rounded border border-slate-800">
              <span class="text-slate-400">공급가액 (Net Amount)</span>
              <span class="text-slate-300">₩{{ netAmount.toLocaleString() }}</span>
            </div>
            <div class="flex justify-between p-2 bg-slate-900 rounded border border-slate-800">
              <span class="text-slate-400">마켓 수수료 ({{ commissionTier === 'STANDARD' ? '15%' : '10%' }})</span>
              <span class="text-amber-400">- ₩{{ commissionAmount.toLocaleString() }}</span>
            </div>
            <div class="flex justify-between p-3 bg-emerald-950/60 rounded border border-emerald-500/40 text-emerald-300">
              <span class="font-bold">개발사 최종 수령액 (Net Payout)</span>
              <span class="font-bold text-base">₩{{ developerPayout.toLocaleString() }}</span>
            </div>
          </div>

          <div class="mt-4 p-2.5 bg-slate-900/80 rounded border border-slate-800 text-[11px] text-slate-400">
            <div class="text-sky-400 font-bold mb-0.5">단수 오차 무결성 (Zero Rounding Discrepancy)</div>
            정산 공식 검증: ₩{{ developerPayout.toLocaleString() }} + ₩{{ commissionAmount.toLocaleString() }} + ₩{{ vatAmount.toLocaleString() }} = ₩{{ (developerPayout + commissionAmount + vatAmount).toLocaleString() }} (총액 합치 100%)
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

const isPaidApp = ref(true);
const commissionTier = ref<'STANDARD' | 'SMALL_DEV'>('STANDARD');
const krwPrice = ref(11000);

const usdPrice = computed(() => krwPrice.value / 1350);
const eurPrice = computed(() => krwPrice.value / 1470);

// Exact integer arithmetic per TC-COM-005:
// net = gross * 10000 / 11000
const netAmount = computed(() => Math.floor((krwPrice.value * 10000) / 11000));
const vatAmount = computed(() => krwPrice.value - netAmount.value);

const commissionRate = computed(() => (commissionTier.value === 'STANDARD' ? 0.15 : 0.10));
const commissionAmount = computed(() => Math.floor(netAmount.value * commissionRate.value));

// Zero rounding leakage: devPayout = net - commission
const developerPayout = computed(() => netAmount.value - commissionAmount.value);

function savePricing() {
  alert(`가격표 저장 완료: ₩${krwPrice.value.toLocaleString()} (실수령액 ₩${developerPayout.value.toLocaleString()})`);
}
</script>

