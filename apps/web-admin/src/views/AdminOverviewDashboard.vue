<template>
  <div class="bn-admin-page">
    <header class="bn-page-header">
      <h2>플랫폼 종합 관제 대시보드 (Admin Overview Dashboard)</h2>
      <p class="bn-text-muted">실시간 활성 사용자, 다운로드 통계, 보안 위협 및 정산 현황을 모니터링합니다.</p>
    </header>

    <!-- Key Performance Metrics Grid -->
    <div class="bn-metrics-grid">
      <div class="bn-metric-card">
        <span class="bn-metric-title">실시간 활성 기기 (Active DAU) [시뮬레이션]</span>
        <strong class="bn-metric-value">1,482,900</strong>
        <span class="bn-metric-delta positive">+12.4% vs 지난주</span>
      </div>

      <div class="bn-metric-card">
        <span class="bn-metric-title">일일 앱 설치 / 업데이트 [시뮬레이션]</span>
        <strong class="bn-metric-value">389,120</strong>
        <span class="bn-metric-delta positive">+8.1%</span>
      </div>

      <div class="bn-metric-card">
        <span class="bn-metric-title">보안 검역 격리된 위협 (24h)</span>
        <strong class="bn-metric-value warning">14건</strong>
        <span class="bn-metric-delta">전수 차단 완료</span>
      </div>

      <div class="bn-metric-card">
        <span class="bn-metric-title">이중 승인 대기열 (Dual-Control)</span>
        <strong class="bn-metric-value">2건</strong>
        <span class="bn-metric-delta">승인 대기 중 (RBAC 큐 연동)</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
// Admin Overview Component
</script>

<style scoped>
.bn-admin-page {
  padding: 24px;
}
.bn-metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 20px;
  margin-top: 24px;
}
.bn-metric-card {
  background-color: var(--bn-bg-surface, #FFFFFF);
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  border-radius: 12px;
  padding: 20px;
  display: flex;
  flex-direction: column;
}
.bn-metric-title {
  font-size: 13px;
  color: var(--bn-text-muted, #64748B);
  font-weight: 500;
}
.bn-metric-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--bn-text-primary, #0F172A);
  margin: 10px 0 6px 0;
}
.bn-metric-value.warning {
  color: var(--bn-danger-red, #DC2626);
}
.bn-metric-delta {
  font-size: 12px;
  color: var(--bn-text-muted, #64748B);
}
.bn-metric-delta.positive {
  color: var(--bn-trust-verified, #059669);
  font-weight: 600;
}
</style>

