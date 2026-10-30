<template>
  <div class="ad-campaign-manager p-6 bg-slate-900 text-slate-100 min-h-screen">
    <header class="mb-8">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">스폰서 검색광고 캠페인 매니저 (DEV-12)</h1>
          <p class="text-sm text-slate-400 mt-1">검색 결과 최상단 스폰서 배치, CPC 입찰가 관리, 일일 예산 캡 및 광고 수익률(ROAS) 분석</p>
        </div>
        <div class="flex gap-2">
          <button @click="openCreateCampaignModal" class="px-4 py-2 bg-sky-600 hover:bg-sky-500 text-white font-medium text-xs rounded transition shadow">
            + 새 스폰서 캠페인 생성
          </button>
        </div>
      </div>
    </header>

    <!-- Key Metrics Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">당월 총 광고 집행액</span>
        <div class="text-2xl font-bold mt-1 text-slate-200">₩{{ totalSpendWon.toLocaleString() }}</div>
      </div>
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">광고 유입 다운로드</span>
        <div class="text-2xl font-bold mt-1 text-sky-400">{{ totalAdConversions.toLocaleString() }}건</div>
        <div class="text-[11px] text-slate-400 mt-1">전환 단가(CPA): ₩{{ Math.round(totalSpendWon / totalAdConversions).toLocaleString() }}</div>
      </div>
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">평균 광고 수익률 (ROAS)</span>
        <div class="text-2xl font-bold mt-1 text-emerald-400">{{ averageRoas.toFixed(1) }}x</div>
        <div class="text-[11px] text-emerald-300 mt-1">광고비 대비 {{ (averageRoas * 100).toFixed(0) }}% 매출 달성</div>
      </div>
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">부정클릭 환불 보호</span>
        <div class="text-2xl font-bold mt-1 text-indigo-400">₩184,000</div>
        <div class="text-[11px] text-slate-400 mt-1">사기 클릭 0원 과금 정책 적용</div>
      </div>
    </div>

    <!-- Campaigns List -->
    <div class="bg-slate-800/90 border border-slate-700 rounded-lg overflow-hidden">
      <div class="p-4 border-b border-slate-700 flex justify-between items-center">
        <h2 class="text-base font-semibold text-slate-200">운영 중인 스폰서 검색 광고 캠페인</h2>
      </div>

      <table class="w-full text-left text-sm text-slate-300">
        <thead class="bg-slate-900/60 text-xs uppercase tracking-wider text-slate-400 border-b border-slate-700">
          <tr>
            <th class="p-3">캠페인 / 대상 앱</th>
            <th class="p-3">타겟 키워드</th>
            <th class="p-3">CPC 입찰가</th>
            <th class="p-3">일일 예산 한도 (Budget Cap)</th>
            <th class="p-3">오늘 소진율</th>
            <th class="p-3">ROAS</th>
            <th class="p-3">상태</th>
            <th class="p-3 text-right">제어</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-700/60">
          <tr v-for="c in campaigns" :key="c.id" class="hover:bg-slate-750 transition">
            <td class="p-3 font-mono text-xs">
              <div class="font-bold text-sky-400">{{ c.name }}</div>
              <div class="text-slate-400">{{ c.appName }}</div>
            </td>
            <td class="p-3 text-xs">
              <span v-for="k in c.keywords" :key="k" class="inline-block bg-slate-700 px-1.5 py-0.5 rounded text-[11px] mr-1">
                {{ k }}
              </span>
            </td>
            <td class="p-3 text-xs font-mono">
              ₩{{ c.cpcBidWon.toLocaleString() }}
            </td>
            <td class="p-3 text-xs font-mono">
              ₩{{ c.dailyBudgetWon.toLocaleString() }}
            </td>
            <td class="p-3 text-xs font-mono">
              <div class="w-24 bg-slate-700 h-2 rounded overflow-hidden mb-1">
                <div 
                  class="h-full bg-sky-500 rounded" 
                  :style="{ width: Math.min(100, (c.todaySpendWon / c.dailyBudgetWon) * 100) + '%' }"
                ></div>
              </div>
              <div class="text-[11px] text-slate-400">
                ₩{{ c.todaySpendWon.toLocaleString() }} ({{ Math.round((c.todaySpendWon / c.dailyBudgetWon) * 100) }}%)
              </div>
            </td>
            <td class="p-3 text-xs font-mono font-bold text-emerald-400">
              {{ c.roas.toFixed(1) }}x
            </td>
            <td class="p-3 text-xs">
              <span :class="getStatusBadgeClass(c.status)" class="px-2 py-0.5 rounded text-[11px] font-semibold border">
                {{ c.status }}
              </span>
            </td>
            <td class="p-3 text-right text-xs">
              <button 
                @click="toggleStatus(c)"
                :class="c.status === 'RUNNING' ? 'bg-amber-600 hover:bg-amber-500' : 'bg-emerald-600 hover:bg-emerald-500'"
                class="px-2.5 py-1 text-white rounded font-medium text-xs transition"
              >
                {{ c.status === 'RUNNING' ? '일시중지' : '재개' }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

interface CampaignItem {
  id: string;
  name: string;
  appName: string;
  keywords: string[];
  cpcBidWon: number;
  dailyBudgetWon: number;
  todaySpendWon: number;
  roas: number;
  status: 'RUNNING' | 'PAUSED' | 'BUDGET_EXHAUSTED';
}

const totalSpendWon = ref(1840000);
const totalAdConversions = ref(2450);
const averageRoas = ref(3.8);

const campaigns = ref<CampaignItem[]>([
  {
    id: 'cmp-01',
    name: '생산성 카테고리 1위 선점',
    appName: 'Super Office Suite',
    keywords: ['office', 'document', 'pdf', 'excel'],
    cpcBidWon: 450,
    dailyBudgetWon: 150000,
    todaySpendWon: 112000,
    roas: 4.2,
    status: 'RUNNING'
  },
  {
    id: 'cmp-02',
    name: '레트로 카메라 키워드 공격',
    appName: 'Pixel Retro Cam',
    keywords: ['camera', 'photo', 'vintage'],
    cpcBidWon: 280,
    dailyBudgetWon: 50000,
    todaySpendWon: 50000,
    roas: 3.1,
    status: 'BUDGET_EXHAUSTED'
  }
]);

function getStatusBadgeClass(status: string) {
  switch (status) {
    case 'RUNNING': return 'bg-emerald-950 text-emerald-400 border-emerald-500/30';
    case 'PAUSED': return 'bg-amber-950 text-amber-400 border-amber-500/30';
    case 'BUDGET_EXHAUSTED': return 'bg-slate-700 text-slate-300 border-slate-600';
    default: return 'bg-slate-800 text-slate-400';
  }
}

function toggleStatus(c: CampaignItem) {
  c.status = c.status === 'RUNNING' ? 'PAUSED' : 'RUNNING';
}

function openCreateCampaignModal() {
  alert('새 캠페인 생성 마법사를 엽니다.');
}
</script>

