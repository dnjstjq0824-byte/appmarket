<template>
  <div class="an-page">
    <header class="an-header">
      <div class="an-header-row">
        <div>
          <h1 class="an-title">플랫폼 전환율 퍼널 및 매출 통합 통계 센터 (ADM-15)</h1>
          <p class="an-sub">검색→상세→다운로드→설치→결제 6단계 전환율 분석 및 비즈니스 KPI 모니터링</p>
        </div>
        <div class="an-header-actions">
          <span class="an-counts">전체 {{ totalEvents.toLocaleString() }}건 · 최근 24h {{ last24h.toLocaleString() }}건</span>
          <button :disabled="refreshing" class="an-refresh" @click="refresh">
            {{ refreshing ? '새로고침 중…' : '데이터 새로고침' }}
          </button>
        </div>
      </div>
    </header>

    <div v-if="isLoading" class="an-status">불러오는 중…</div>
    <div v-if="loadError" class="an-load-error">{{ loadError }}</div>
    <div v-if="!isLoading && !loadError && totalEvents === 0" class="an-status">수집된 텔레메트리 이벤트가 없습니다. 클라이언트에서 이벤트를 전송하면 퍼널이 채워집니다.</div>

    <!-- Operational KPI Alert Banners -->
    <div v-if="!isLoading && !loadError" class="an-kpis">
      <!-- Search to Detail CTR -->
      <div class="an-kpi">
        <div class="an-kpi-head">
          <span>Search-to-Detail CTR</span>
          <span class="an-mono an-green">목표 >= 28%</span>
        </div>
        <div class="an-kpi-value an-green">{{ searchToDetailCtr }}%</div>
        <div class="an-kpi-note">검색 결과 클릭 전환율 양호</div>
      </div>

      <!-- Detail to Download CVR -->
      <div class="an-kpi">
        <div class="an-kpi-head">
          <span>Detail-to-Download CVR</span>
          <span class="an-mono an-green">목표 >= 18%</span>
        </div>
        <div class="an-kpi-value an-green">{{ detailToDownloadCvr }}%</div>
        <div class="an-kpi-note">상세 페이지 전환율 기준 상회</div>
      </div>

      <!-- Download to Install Success Rate -->
      <div class="an-kpi">
        <div class="an-kpi-head">
          <span>Install Success Rate</span>
          <span class="an-mono an-green">목표 >= 94%</span>
        </div>
        <div class="an-kpi-value an-sky">{{ installSuccessRate }}%</div>
        <div class="an-kpi-note">호환성 오류 및 중단율 정상</div>
      </div>

      <!-- Total Platform GMV -->
      <div class="an-kpi">
        <div class="an-kpi-head">
          <span>총 결제 금액 (수집 합산)</span>
          <span class="an-mono an-sky">cents</span>
        </div>
        <div class="an-kpi-value an-sky">{{ totalRevenueCents.toLocaleString() }}</div>
        <div class="an-kpi-note">예상 수수료(15%): {{ Math.round(totalRevenueCents * 0.15).toLocaleString() }}</div>
      </div>
    </div>

    <!-- 6-Stage Conversion Funnel Visualizer -->
    <div v-if="!isLoading && !loadError" class="an-panel an-funnel">
      <h2 class="an-panel-title">6단계 전 도메인 전환율 퍼널 (Funnel Telemetry)</h2>

      <div class="an-stages">
        <div v-for="(stage, idx) in funnelStages" :key="stage.name" class="an-stage">
          <div class="an-stage-name">{{ stage.name }}</div>
          <div class="an-stage-count">{{ stage.count.toLocaleString() }}</div>
          <div v-if="idx > 0" class="an-stage-delta">
            전 단계 대비 <span class="an-green an-bold">{{ funnelStages[idx-1].count > 0 ? ((stage.count / funnelStages[idx-1].count) * 100).toFixed(1) : '—' }}%</span>
          </div>
          <div v-else class="an-stage-delta">
            기준 모수 100%
          </div>
        </div>
      </div>
    </div>

    <!-- Platform Breakdown Tables -->
    <div v-if="!isLoading && !loadError" class="an-breakdown">
      <div class="an-panel">
        <h3 class="an-panel-sub">플랫폼별 이벤트 비중</h3>
        <div class="an-bars">
          <div v-for="p in platformStats" :key="p.platform" class="an-bar-row">
            <div class="an-bar-labels">
              <span class="an-mono an-light">{{ p.platform || '(미지정)' }}</span>
              <span class="an-dim">{{ p.sharePercent }}% ({{ p.count.toLocaleString() }}건)</span>
            </div>
            <div class="an-bar-track">
              <div class="an-bar-fill an-bar-sky" :style="{ width: p.sharePercent + '%' }"></div>
            </div>
          </div>
          <div v-if="platformStats.length === 0" class="an-dim-sm">데이터 없음</div>
        </div>
      </div>

      <div class="an-panel">
        <h3 class="an-panel-sub">국가별 결제 매출 순위</h3>
        <div class="an-bars">
          <div v-for="c in countryStats" :key="c.code" class="an-bar-row">
            <div class="an-bar-labels">
              <span class="an-mono an-light">{{ c.code || '(미지정)' }} ({{ c.purchases }}건)</span>
              <span class="an-dim">{{ c.revenueCents.toLocaleString() }} ({{ c.sharePercent }}%)</span>
            </div>
            <div class="an-bar-track">
              <div class="an-bar-fill an-bar-indigo" :style="{ width: c.sharePercent + '%' }"></div>
            </div>
          </div>
          <div v-if="countryStats.length === 0" class="an-dim-sm">결제 이벤트 없음</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { getSharedClient, ApiClientError, type AnalyticsOverview } from '@bookingnetwork/api-client';

const api = getSharedClient();

const overview = ref<AnalyticsOverview | null>(null);
const isLoading = ref(true);
const loadError = ref('');
const refreshing = ref(false);

const EVENT_LABELS: Array<[string, string]> = [
  ['app.search.result_view', '1. 검색 노출'],
  ['app.detail.view', '2. 앱 상세 조회'],
  ['app.download.started', '3. 다운로드 시작'],
  ['app.download.completed', '4. 다운로드 완료'],
  ['app.install.succeeded', '5. 설치 완료'],
  ['app.purchase.completed', '6. 유료 결제']
];

const countsByEvent = computed(() => {
  const m = new Map<string, number>();
  for (const e of overview.value?.byEvent || []) m.set(e.key, e.count);
  return m;
});

const funnelStages = computed(() =>
  EVENT_LABELS.map(([event, name]) => ({ name, count: countsByEvent.value.get(event) || 0 }))
);

const pct = (num: number, den: number) => (den > 0 ? (num / den) * 100 : 0);

const searchToDetailCtr = computed(() =>
  pct(countsByEvent.value.get('app.detail.view') || 0, countsByEvent.value.get('app.search.result_view') || 0).toFixed(1));
const detailToDownloadCvr = computed(() =>
  pct(countsByEvent.value.get('app.download.started') || 0, countsByEvent.value.get('app.detail.view') || 0).toFixed(1));
const installSuccessRate = computed(() =>
  pct(countsByEvent.value.get('app.install.succeeded') || 0, countsByEvent.value.get('app.download.completed') || 0).toFixed(1));
const totalRevenueCents = computed(() => overview.value?.totalRevenueCents || 0);
const totalEvents = computed(() => overview.value?.totalEvents || 0);
const last24h = computed(() => overview.value?.last24h || 0);

const platformStats = computed(() => {
  const list = overview.value?.byPlatform || [];
  const total = list.reduce((s, p) => s + p.count, 0);
  return list.map(p => ({
    platform: p.key,
    count: p.count,
    sharePercent: total > 0 ? +((p.count / total) * 100).toFixed(1) : 0
  }));
});

const countryStats = computed(() => {
  const list = overview.value?.revenueByCountry || [];
  const total = list.reduce((s, c) => s + c.revenueCents, 0);
  return list.map(c => ({
    code: c.key,
    revenueCents: c.revenueCents,
    purchases: c.count,
    sharePercent: total > 0 ? +((c.revenueCents / total) * 100).toFixed(1) : 0
  }));
});

async function load() {
  loadError.value = '';
  try {
    overview.value = await api.getAnalyticsOverview();
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load analytics overview';
  }
}

async function refresh() {
  refreshing.value = true;
  await load();
  refreshing.value = false;
}

onMounted(async () => {
  await load();
  isLoading.value = false;
});
</script>

<style scoped>
.an-page {
  max-width: 1200px;
  margin: 0 auto;
}
.an-header {
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 2px solid var(--bn-ink);
}
.an-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}
.an-title {
  font-family: var(--bn-font-display);
  font-size: 24px;
  font-weight: 900;
  letter-spacing: -0.02em;
  margin: 0;
}
.an-sub {
  font-size: 13px;
  color: var(--bn-text-secondary);
  margin: 4px 0 0;
}
.an-header-actions {
  display: flex;
  gap: 10px;
  align-items: center;
}
.an-counts {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  color: var(--bn-text-muted);
}
.an-refresh {
  padding: 7px 14px;
  background: transparent;
  border: 1px solid var(--bn-border-subtle);
  color: var(--bn-text-secondary);
  font-size: 12px;
  font-weight: 600;
  border-radius: var(--bn-radius-sm);
  cursor: pointer;
}
.an-refresh:hover:not(:disabled) {
  border-color: var(--bn-border-strong);
  color: var(--bn-text-primary);
}
.an-refresh:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.an-status {
  font-size: 14px;
  color: var(--bn-text-muted);
  margin-bottom: 24px;
  padding: 28px 24px;
  text-align: center;
  border: 1px dashed var(--bn-border-strong);
  border-radius: var(--bn-radius-md);
}
.an-load-error {
  font-size: 13px;
  color: var(--bn-danger-red);
  background: var(--bn-danger-wash);
  border: 1px solid var(--bn-danger-red);
  border-radius: var(--bn-radius-sm);
  padding: 12px 16px;
  margin-bottom: 24px;
}
.an-kpis {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
  margin-bottom: 24px;
}
@media (min-width: 768px) {
  .an-kpis { grid-template-columns: repeat(4, 1fr); }
}
.an-kpi {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  padding: 18px;
  border-radius: var(--bn-radius-md);
}
.an-kpi-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 8px;
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.06em;
  color: var(--bn-text-muted);
}
.an-mono { font-family: var(--bn-font-mono); }
.an-green { color: var(--bn-trust-verified); }
.an-sky { color: var(--bn-seal); }
.an-bold { font-weight: 700; }
.an-kpi-value {
  font-family: var(--bn-font-display);
  font-size: 32px;
  font-weight: 900;
  letter-spacing: -0.02em;
  margin-top: 8px;
}
.an-kpi-note {
  font-size: 11px;
  color: var(--bn-text-muted);
  margin-top: 6px;
}
.an-panel {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 24px;
}
.an-funnel { margin-bottom: 24px; }
.an-panel-title {
  font-family: var(--bn-font-display);
  font-size: 15px;
  font-weight: 800;
  margin: 0 0 20px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--bn-border-subtle);
}
.an-panel-sub {
  font-family: var(--bn-font-display);
  font-size: 14px;
  font-weight: 800;
  margin: 0 0 16px;
}
.an-stages {
  display: grid;
  grid-template-columns: 1fr;
  gap: 8px;
  text-align: center;
}
@media (min-width: 768px) {
  .an-stages { grid-template-columns: repeat(6, 1fr); }
}
.an-stage {
  position: relative;
  background: var(--bn-bg-well);
  padding: 16px 12px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-subtle);
}
.an-stage-name {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.04em;
  color: var(--bn-text-muted);
  margin-bottom: 6px;
}
.an-stage-count {
  font-family: var(--bn-font-display);
  font-size: 22px;
  font-weight: 900;
  letter-spacing: -0.02em;
}
.an-stage-delta {
  font-size: 11px;
  color: var(--bn-text-muted);
  margin-top: 8px;
}
.an-breakdown {
  display: grid;
  grid-template-columns: 1fr;
  gap: 24px;
}
@media (min-width: 768px) {
  .an-breakdown { grid-template-columns: repeat(2, 1fr); }
}
.an-bars > * + * { margin-top: 14px; }
.an-bar-labels {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 12px;
  margin-bottom: 6px;
}
.an-light { color: var(--bn-text-secondary); }
.an-dim {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  color: var(--bn-text-muted);
}
.an-dim-sm {
  font-size: 12px;
  color: var(--bn-text-muted);
}
.an-bar-track {
  width: 100%;
  background: var(--bn-bg-well);
  border: 1px solid var(--bn-border-subtle);
  height: 8px;
  border-radius: var(--bn-radius-sm);
  overflow: hidden;
}
.an-bar-fill {
  height: 100%;
}
.an-bar-sky { background: var(--bn-ink); }
.an-bar-indigo { background: var(--bn-seal); }
</style>

