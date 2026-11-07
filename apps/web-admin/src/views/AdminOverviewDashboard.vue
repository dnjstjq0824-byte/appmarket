<template>
  <div class="bn-admin-page">
    <header class="bn-page-header">
      <h2>플랫폼 종합 관제 대시보드 (Admin Overview Dashboard)</h2>
      <p class="bn-text-muted">실시간 활성 사용자, 다운로드 통계, 보안 위협 및 정산 현황을 모니터링합니다.</p>
    </header>

    <div v-if="isLoading" class="bn-loading">불러오는 중…</div>
    <div v-if="loadError" class="bn-error">{{ loadError }}</div>

    <!-- Key Performance Metrics Grid -->
    <div v-if="!isLoading && !loadError" class="bn-metrics-grid">
      <div class="bn-metric-card">
        <span class="bn-metric-title">텔레메트리 이벤트 (최근 24h / 전체)</span>
        <strong class="bn-metric-value">{{ last24h.toLocaleString() }} / {{ totalEvents.toLocaleString() }}</strong>
        <span class="bn-metric-delta">수집된 클라이언트 이벤트 기준</span>
      </div>

      <div class="bn-metric-card">
        <span class="bn-metric-title">앱 심사 대기열</span>
        <strong class="bn-metric-value">{{ reviewQueueCount }}건</strong>
        <span class="bn-metric-delta">메타데이터 심사 대기 중</span>
      </div>

      <div class="bn-metric-card">
        <span class="bn-metric-title">활성 킬스위치</span>
        <strong class="bn-metric-value" :class="{ warning: killSwitchCount > 0 }">{{ killSwitchCount }}건</strong>
        <span class="bn-metric-delta">배포 차단 중인 패키지</span>
      </div>

      <div class="bn-metric-card">
        <span class="bn-metric-title">이중 승인 대기열 (Dual-Control)</span>
        <strong class="bn-metric-value">{{ pendingApprovals }}건</strong>
        <span class="bn-metric-delta">2인 승인 대기 중</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { getSharedClient, ApiClientError } from '@bookingnetwork/api-client';

const api = getSharedClient();

const isLoading = ref(true);
const loadError = ref('');
const totalEvents = ref(0);
const last24h = ref(0);
const reviewQueueCount = ref(0);
const killSwitchCount = ref(0);
const pendingApprovals = ref(0);

onMounted(async () => {
  try {
    const [overview, queue, switches, approvals] = await Promise.all([
      api.getAnalyticsOverview(),
      api.getReviewQueue(),
      api.listKillSwitches(),
      api.listApprovals()
    ]);
    totalEvents.value = overview.totalEvents;
    last24h.value = overview.last24h;
    reviewQueueCount.value = (queue as unknown[]).length;
    killSwitchCount.value = (switches as unknown[]).length;
    pendingApprovals.value = approvals.length;
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load overview metrics';
  } finally {
    isLoading.value = false;
  }
});
</script>

<style scoped>
.bn-admin-page {
  max-width: 1080px;
  margin: 0 auto;
}
.bn-page-header {
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 2px solid var(--bn-ink);
}
.bn-page-header h2 {
  font-family: var(--bn-font-display);
  font-size: 24px;
  font-weight: 900;
  letter-spacing: -0.02em;
  margin: 0 0 4px;
}
.bn-text-muted {
  color: var(--bn-text-secondary);
  font-size: 13px;
  margin: 0;
}
.bn-metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 12px;
}
.bn-metric-card {
  background-color: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 20px;
  display: flex;
  flex-direction: column;
}
.bn-metric-title {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
}
.bn-metric-value {
  font-family: var(--bn-font-display);
  font-size: 32px;
  font-weight: 900;
  letter-spacing: -0.02em;
  margin: 10px 0 6px 0;
}
.bn-metric-value.warning {
  color: var(--bn-danger-red);
}
.bn-metric-delta {
  font-size: 12px;
  color: var(--bn-text-secondary);
}
.bn-metric-delta.positive {
  color: var(--bn-trust-verified);
  font-weight: 600;
}
.bn-loading {
  padding: 24px;
  text-align: center;
  color: var(--bn-text-muted);
  font-size: 14px;
}
.bn-error {
  padding: 12px 16px;
  margin-top: 12px;
  background: var(--bn-danger-wash);
  border: 1px solid var(--bn-danger-red);
  border-radius: var(--bn-radius-sm);
  color: var(--bn-danger-red);
  font-size: 13px;
}
</style>

