<template>
  <div class="ad-campaign-review-view p-6 bg-slate-900 text-slate-100 min-h-screen">
    <header class="mb-8">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">스폰서 광고 캠페인 심사 및 부정클릭 관제 (ADM-12)</h1>
          <p class="text-sm text-slate-400 mt-1">CPC 입찰 광고 소재 심사, 일일 예산 통제 및 3계층 클릭 사기 방지 실시간 현황</p>
        </div>
        <div class="flex gap-2">
          <span class="px-3 py-1 bg-emerald-950/80 border border-emerald-500/40 text-emerald-400 text-xs rounded-full font-mono">
            봇 필터: 100% 능동 차단 중
          </span>
        </div>
      </div>
    </header>

    <!-- Fraud Analytics Header Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">총 활성 캠페인</span>
        <div class="text-2xl font-bold mt-1 text-slate-200">{{ activeCampaignsCount }}</div>
      </div>
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">24시간 총 클릭수</span>
        <div class="text-2xl font-bold mt-1 text-sky-400">{{ totalClicks24h.toLocaleString() }}</div>
      </div>
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">차단된 봇/중복 클릭</span>
        <div class="text-2xl font-bold mt-1 text-rose-400">{{ blockedClicks24h.toLocaleString() }}</div>
        <div class="text-xs text-rose-300/80 mt-1">부정클릭률 {{ fraudRatePercent }}%</div>
      </div>
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">보호된 광고주 예산</span>
        <div class="text-2xl font-bold mt-1 text-emerald-400">₩{{ savedBudgetWon.toLocaleString() }}</div>
      </div>
    </div>

    <!-- Campaigns Review Table -->
    <div class="bg-slate-800/90 border border-slate-700 rounded-lg overflow-hidden">
      <div class="p-4 border-b border-slate-700 flex justify-between items-center">
        <h2 class="text-base font-semibold text-slate-200">심사 대기 및 집행 중인 광고 캠페인</h2>
        <div class="flex gap-2">
          <button 
            v-for="f in ['ALL', 'PENDING_REVIEW', 'ACTIVE', 'PAUSED_BUDGET']" 
            :key="f"
            @click="filterState = f"
            :class="['px-3 py-1 text-xs rounded transition', filterState === f ? 'bg-sky-600 text-white' : 'bg-slate-700 text-slate-300 hover:bg-slate-600']"
          >
            {{ f }}
          </button>
        </div>
      </div>

      <table class="w-full text-left text-sm text-slate-300">
        <thead class="bg-slate-900/60 text-xs uppercase tracking-wider text-slate-400 border-b border-slate-700">
          <tr>
            <th class="p-3">캠페인 ID / 대상 앱</th>
            <th class="p-3">개발사</th>
            <th class="p-3">소재 (타이틀/헤드라인)</th>
            <th class="p-3">타겟 키워드</th>
            <th class="p-3">CPC 입찰가 / 일일 예산</th>
            <th class="p-3">소진율</th>
            <th class="p-3">상태</th>
            <th class="p-3 text-right">조치</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-700/60">
          <tr v-for="c in filteredCampaigns" :key="c.id" class="hover:bg-slate-750 transition">
            <td class="p-3 font-mono text-xs">
              <div class="font-bold text-sky-400">{{ c.appTitle }}</div>
              <div class="text-slate-400">{{ c.id }}</div>
            </td>
            <td class="p-3 text-xs">{{ c.developerOrg }}</td>
            <td class="p-3 text-xs">
              <div class="font-medium text-slate-200">{{ c.headline }}</div>
              <div class="text-slate-400 line-clamp-1">{{ c.description }}</div>
            </td>
            <td class="p-3 text-xs">
              <span v-for="k in c.keywords" :key="k" class="inline-block bg-slate-700 px-1.5 py-0.5 rounded text-[11px] mr-1">
                {{ k }}
              </span>
            </td>
            <td class="p-3 text-xs font-mono">
              <div>입찰: ₩{{ c.cpcBidWon.toLocaleString() }}</div>
              <div class="text-slate-400">한도: ₩{{ c.dailyBudgetWon.toLocaleString() }}</div>
            </td>
            <td class="p-3 text-xs font-mono">
              <div class="w-24 bg-slate-700 h-2 rounded overflow-hidden mb-1">
                <div 
                  class="h-full bg-sky-500 rounded" 
                  :style="{ width: Math.min(100, (c.spentTodayWon / c.dailyBudgetWon) * 100) + '%' }"
                ></div>
              </div>
              <div class="text-[11px] text-slate-400">₩{{ c.spentTodayWon.toLocaleString() }} ({{ Math.round((c.spentTodayWon / c.dailyBudgetWon) * 100) }}%)</div>
            </td>
            <td class="p-3 text-xs">
              <span :class="getStatusBadgeClass(c.status)" class="px-2 py-0.5 rounded text-[11px] font-semibold">
                {{ c.status }}
              </span>
            </td>
            <td class="p-3 text-right text-xs">
              <div class="flex justify-end gap-1">
                <button 
                  v-if="c.status === 'PENDING_REVIEW'"
                  @click="approveCampaign(c.id)"
                  class="px-2.5 py-1 bg-emerald-600 hover:bg-emerald-500 text-white rounded font-medium text-xs transition"
                >
                  승인
                </button>
                <button 
                  v-if="c.status === 'PENDING_REVIEW'"
                  @click="rejectCampaign(c.id)"
                  class="px-2.5 py-1 bg-rose-600 hover:bg-rose-500 text-white rounded font-medium text-xs transition"
                >
                  반려
                </button>
                <button 
                  v-if="c.status === 'ACTIVE'"
                  @click="pauseCampaign(c.id)"
                  class="px-2.5 py-1 bg-amber-600 hover:bg-amber-500 text-white rounded font-medium text-xs transition"
                >
                  긴급 일시중지
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

interface Campaign {
  id: string;
  appTitle: string;
  developerOrg: string;
  headline: string;
  description: string;
  keywords: string[];
  cpcBidWon: number;
  dailyBudgetWon: number;
  spentTodayWon: number;
  status: 'PENDING_REVIEW' | 'ACTIVE' | 'PAUSED_BUDGET' | 'REJECTED';
}

const filterState = ref<string>('ALL');

const campaigns = ref<Campaign[]>([
  {
    id: 'camp-0192ce-8801',
    appTitle: 'Super productivity Office',
    developerOrg: 'OmniWork Corp',
    headline: '팀 생산성을 300% 높이는 오피스 도구',
    description: '문서 편집 및 클라우드 실시간 동기화를 단 하나의 앱으로 해결하세요.',
    keywords: ['office', 'productivity', 'document'],
    cpcBidWon: 450,
    dailyBudgetWon: 200000,
    spentTodayWon: 142000,
    status: 'ACTIVE'
  },
  {
    id: 'camp-0192ce-8802',
    appTitle: 'Pixel Art Photo Studio',
    developerOrg: 'Creative Labs',
    headline: '원클릭 8비트 레트로 필터 카메라',
    description: '레트로 게임 스타일의 감성 필터로 일상을 특별하게 기록하세요.',
    keywords: ['photo', 'camera', 'retro'],
    cpcBidWon: 300,
    dailyBudgetWon: 50000,
    spentTodayWon: 50000,
    status: 'PAUSED_BUDGET'
  },
  {
    id: 'camp-0192ce-8803',
    appTitle: 'Secure Vault Authenticator',
    developerOrg: 'ZeroTrust Security',
    headline: 'E2E 암호화 2차 인증 및 패스키 저장소',
    description: '공식 상용급 TOTP 및 하드웨어 시큐리티 키 연동 보안 앱.',
    keywords: ['security', 'totp', '2fa', 'passkey'],
    cpcBidWon: 600,
    dailyBudgetWon: 100000,
    spentTodayWon: 0,
    status: 'PENDING_REVIEW'
  }
]);

const activeCampaignsCount = computed(() => campaigns.value.filter(c => c.status === 'ACTIVE').length);
const totalClicks24h = ref(18420);
const blockedClicks24h = ref(2140);
const fraudRatePercent = computed(() => ((blockedClicks24h.value / (totalClicks24h.value + blockedClicks24h.value)) * 100).toFixed(1));
const savedBudgetWon = ref(963000);

const filteredCampaigns = computed(() => {
  if (filterState.value === 'ALL') return campaigns.value;
  return campaigns.value.filter(c => c.status === filterState.value);
});

function getStatusBadgeClass(status: string) {
  switch (status) {
    case 'ACTIVE': return 'bg-emerald-950 text-emerald-400 border border-emerald-500/30';
    case 'PENDING_REVIEW': return 'bg-amber-950 text-amber-400 border border-amber-500/30';
    case 'PAUSED_BUDGET': return 'bg-slate-700 text-slate-300 border border-slate-600';
    case 'REJECTED': return 'bg-rose-950 text-rose-400 border border-rose-500/30';
    default: return 'bg-slate-800 text-slate-400';
  }
}

function approveCampaign(id: string) {
  const item = campaigns.value.find(c => c.id === id);
  if (item) item.status = 'ACTIVE';
}

function rejectCampaign(id: string) {
  const item = campaigns.value.find(c => c.id === id);
  if (item) item.status = 'REJECTED';
}

function pauseCampaign(id: string) {
  const item = campaigns.value.find(c => c.id === id);
  if (item) item.status = 'PAUSED_BUDGET';
}
</script>

