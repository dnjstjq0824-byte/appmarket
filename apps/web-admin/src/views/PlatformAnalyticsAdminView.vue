<template>
  <div class="platform-analytics-admin-view p-6 bg-slate-900 text-slate-100 min-h-screen">
    <header class="mb-8">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">플랫폼 전환율 퍼널 및 매출 통합 통계 센터 (ADM-15)</h1>
          <p class="text-sm text-slate-400 mt-1">검색→상세→다운로드→설치→결제 6단계 전환율 분석 및 비즈니스 KPI 모니터링</p>
        </div>
        <div class="flex gap-2">
          <button class="px-3 py-1 bg-slate-800 border border-slate-700 text-slate-300 text-xs rounded hover:bg-slate-700">
            데이터 새로고침
          </button>
        </div>
      </div>
    </header>

    <!-- Operational KPI Alert Banners -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <!-- Search to Detail CTR -->
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <div class="flex justify-between items-center text-xs text-slate-400">
          <span>Search-to-Detail CTR</span>
          <span class="text-emerald-400 font-mono">목표 >= 28%</span>
        </div>
        <div class="text-2xl font-bold mt-1 text-emerald-400">{{ searchToDetailCtr }}%</div>
        <div class="text-[11px] text-slate-400 mt-1">검색 결과 클릭 전환율 양호</div>
      </div>

      <!-- Detail to Download CVR -->
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <div class="flex justify-between items-center text-xs text-slate-400">
          <span>Detail-to-Download CVR</span>
          <span class="text-emerald-400 font-mono">목표 >= 18%</span>
        </div>
        <div class="text-2xl font-bold mt-1 text-emerald-400">{{ detailToDownloadCvr }}%</div>
        <div class="text-[11px] text-slate-400 mt-1">상세 페이지 전환율 기준 상회</div>
      </div>

      <!-- Download to Install Success Rate -->
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <div class="flex justify-between items-center text-xs text-slate-400">
          <span>Install Success Rate</span>
          <span class="text-emerald-400 font-mono">목표 >= 94%</span>
        </div>
        <div class="text-2xl font-bold mt-1 text-sky-400">{{ installSuccessRate }}%</div>
        <div class="text-[11px] text-slate-400 mt-1">호환성 오류 및 중단율 정상</div>
      </div>

      <!-- Total Platform GMV -->
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <div class="flex justify-between items-center text-xs text-slate-400">
          <span>당월 총 거래액 (GMV)</span>
          <span class="text-sky-400 font-mono">KRW</span>
        </div>
        <div class="text-2xl font-bold mt-1 text-sky-400">₩{{ totalGmvWon.toLocaleString() }}</div>
        <div class="text-[11px] text-slate-400 mt-1">수수료 수익: ₩{{ (totalGmvWon * 0.15).toLocaleString() }}</div>
      </div>
    </div>

    <!-- 6-Stage Conversion Funnel Visualizer -->
    <div class="bg-slate-800/90 border border-slate-700 rounded-lg p-6 mb-8">
      <h2 class="text-base font-semibold text-slate-200 mb-6">6단계 전 도메인 전환율 퍼널 (Funnel Telemetry)</h2>

      <div class="grid grid-cols-1 md:grid-cols-6 gap-2 text-center">
        <div v-for="(stage, idx) in funnelStages" :key="stage.name" class="relative bg-slate-900/80 p-4 rounded border border-slate-750">
          <div class="text-xs text-slate-400 font-medium mb-1">{{ stage.name }}</div>
          <div class="text-lg font-bold text-sky-400 font-mono">{{ stage.count.toLocaleString() }}</div>
          <div v-if="idx > 0" class="text-[11px] text-slate-500 mt-2">
            전 단계 대비 <span class="text-emerald-400 font-semibold">{{ ((stage.count / funnelStages[idx-1].count) * 100).toFixed(1) }}%</span>
          </div>
          <div v-else class="text-[11px] text-slate-500 mt-2">
            기준 모수 100%
          </div>
        </div>
      </div>
    </div>

    <!-- Platform Breakdown Tables -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="bg-slate-800/90 border border-slate-700 rounded-lg p-5">
        <h3 class="text-sm font-semibold text-slate-200 mb-4">플랫폼별 유입 비중</h3>
        <div class="space-y-3">
          <div v-for="p in platformStats" :key="p.platform" class="space-y-1">
            <div class="flex justify-between text-xs">
              <span class="text-slate-300 font-mono">{{ p.platform }}</span>
              <span class="text-slate-400">{{ p.sharePercent }}% ({{ p.users.toLocaleString() }}명)</span>
            </div>
            <div class="w-full bg-slate-700 h-2 rounded overflow-hidden">
              <div class="bg-sky-500 h-full rounded" :style="{ width: p.sharePercent + '%' }"></div>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-slate-800/90 border border-slate-700 rounded-lg p-5">
        <h3 class="text-sm font-semibold text-slate-200 mb-4">국가별 매출 순위</h3>
        <div class="space-y-3">
          <div v-for="c in countryStats" :key="c.code" class="space-y-1">
            <div class="flex justify-between text-xs">
              <span class="text-slate-300 font-mono">{{ c.code }} ({{ c.name }})</span>
              <span class="text-slate-400">₩{{ c.revenueWon.toLocaleString() }} ({{ c.sharePercent }}%)</span>
            </div>
            <div class="w-full bg-slate-700 h-2 rounded overflow-hidden">
              <div class="bg-indigo-500 h-full rounded" :style="{ width: c.sharePercent + '%' }"></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const searchToDetailCtr = ref(32.4);
const detailToDownloadCvr = ref(21.8);
const installSuccessRate = ref(96.2);
const totalGmvWon = ref(38500000);

const funnelStages = ref([
  { name: '1. 검색 노출', count: 125000 },
  { name: '2. 앱 상세 조회', count: 40500 },
  { name: '3. 다운로드 시작', count: 8830 },
  { name: '4. 다운로드 완료', count: 8610 },
  { name: '5. 설치 완료', count: 8280 },
  { name: '6. 유료 결제', count: 1140 }
]);

const platformStats = ref([
  { platform: 'ANDROID_CLIENT', sharePercent: 68, users: 85000 },
  { platform: 'WEB_USER_DESKTOP', sharePercent: 22, users: 27500 },
  { platform: 'WEB_USER_MOBILE', sharePercent: 10, users: 12500 }
]);

const countryStats = ref([
  { code: 'KR', name: '대한민국', revenueWon: 28875000, sharePercent: 75 },
  { code: 'US', name: '미국', revenueWon: 5775000, sharePercent: 15 },
  { code: 'JP', name: '일본', revenueWon: 2310000, sharePercent: 6 },
  { code: 'DE', name: '독일', revenueWon: 1540000, sharePercent: 4 }
]);
</script>

