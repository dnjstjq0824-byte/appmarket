<template>
  <div class="ad-campaign-review-view">
    <header class="view-header">
      <div>
        <h1>스폰서 광고 캠페인 심사 (ADM-12)</h1>
        <p class="subtitle">CPC 입찰 광고 심사 및 일일 예산 집행 현황</p>
      </div>
      <div class="filter-row">
        <button
          v-for="f in ['ALL', 'PENDING_REVIEW', 'ACTIVE', 'PAUSED_MANUAL', 'PAUSED_DAILY_BUDGET_CAP', 'REJECTED']"
          :key="f"
          @click="filterState = f; loadCampaigns()"
          :class="['filter-btn', filterState === f ? 'active' : '']"
        >
          {{ f }}
        </button>
      </div>
    </header>

    <div class="summary-grid">
      <div class="summary-card">
        <span class="summary-label">심사 대기</span>
        <div class="summary-value">{{ pendingCount }}</div>
      </div>
      <div class="summary-card">
        <span class="summary-label">활성 캠페인</span>
        <div class="summary-value text-success">{{ activeCount }}</div>
      </div>
      <div class="summary-card">
        <span class="summary-label">금일 집행 합계</span>
        <div class="summary-value">₩{{ totalSpentToday.toLocaleString() }}</div>
      </div>
    </div>

    <div v-if="isLoading" class="loading-row">불러오는 중…</div>
    <div v-if="loadError" class="load-error">{{ loadError }}</div>
    <div v-if="actionError" class="load-error">{{ actionError }}</div>

    <div v-if="!isLoading && !loadError" class="list-card">
      <div v-if="campaigns.length === 0" class="empty-row">해당 상태의 캠페인이 없습니다.</div>
      <table v-else class="campaign-table">
        <thead>
          <tr>
            <th>캠페인 ID / 대상 앱</th>
            <th>조직</th>
            <th>타겟 키워드</th>
            <th>CPC 입찰가 / 일일 예산</th>
            <th>소진율</th>
            <th>품질 점수</th>
            <th>상태</th>
            <th>조치</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in campaigns" :key="c.id">
            <td class="mono">
              <div class="app-name">{{ c.app_name }}</div>
              <div class="muted">{{ c.id }}</div>
            </td>
            <td class="mono muted">{{ c.org_id }}</td>
            <td>
              <span v-for="k in (c.target_keywords || [])" :key="k" class="keyword-tag">{{ k }}</span>
            </td>
            <td class="mono">
              <div>입찰: ₩{{ Number(c.max_bid_cpc).toLocaleString() }}</div>
              <div class="muted">한도: ₩{{ Number(c.daily_budget).toLocaleString() }}</div>
            </td>
            <td class="mono">
              <div class="progress-track">
                <div class="progress-bar" :style="{ width: spendPercent(c) + '%' }"></div>
              </div>
              <div class="muted text-xs">₩{{ Number(c.spent_today).toLocaleString() }} ({{ spendPercent(c) }}%)</div>
            </td>
            <td class="mono">{{ Number(c.quality_score).toFixed(1) }}</td>
            <td>
              <span class="badge" :class="statusClass(c.status)">{{ c.status }}</span>
            </td>
            <td>
              <div v-if="c.status === 'PENDING_REVIEW'" class="action-cell">
                <button @click="review(c.id, true)" class="btn btn-success btn-xs" :disabled="isActing">승인</button>
                <button @click="review(c.id, false)" class="btn btn-danger btn-xs" :disabled="isActing">반려</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { getSharedClient, ApiClientError } from '@bookingnetwork/api-client';

const api = getSharedClient();

interface Campaign {
  id: string;
  org_id: string;
  app_id: string;
  app_name: string;
  target_keywords: string[];
  max_bid_cpc: number;
  daily_budget: number;
  spent_today: number;
  quality_score: number;
  status: string;
  created_at: string;
}

const filterState = ref('ALL');
const campaigns = ref<Campaign[]>([]);
const isLoading = ref(true);
const loadError = ref('');
const actionError = ref('');
const isActing = ref(false);

const pendingCount = computed(() => campaigns.value.filter(c => c.status === 'PENDING_REVIEW').length);
const activeCount = computed(() => campaigns.value.filter(c => c.status === 'ACTIVE').length);
const totalSpentToday = computed(() => campaigns.value.reduce((n, c) => n + Number(c.spent_today || 0), 0));

onMounted(async () => {
  await loadCampaigns();
  isLoading.value = false;
});

async function loadCampaigns() {
  loadError.value = '';
  try {
    campaigns.value = await api.listAdminAdCampaigns(filterState.value === 'ALL' ? undefined : filterState.value) as Campaign[];
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load campaigns';
  }
}

async function review(id: string, approve: boolean) {
  isActing.value = true;
  actionError.value = '';
  try {
    await api.reviewAdCampaign(id, approve);
    await loadCampaigns();
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to review campaign';
  } finally {
    isActing.value = false;
  }
}

function spendPercent(c: Campaign): number {
  if (!c.daily_budget) return 0;
  return Math.min(100, Math.round((Number(c.spent_today) / Number(c.daily_budget)) * 100));
}

function statusClass(status: string): string {
  switch (status) {
    case 'ACTIVE': return 'badge-success';
    case 'PENDING_REVIEW': return 'badge-pending';
    case 'REJECTED': return 'badge-danger';
    default: return 'badge-muted';
  }
}
</script>

<style scoped>
.ad-campaign-review-view {
  max-width: 1200px;
  margin: 0 auto;
}
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 2px solid var(--bn-ink);
  flex-wrap: wrap;
}
.view-header h1 {
  font-family: var(--bn-font-display);
  font-size: 24px;
  font-weight: 900;
  letter-spacing: -0.02em;
  margin: 0 0 4px;
}
.subtitle {
  color: var(--bn-text-secondary);
  font-size: 13px;
  margin: 0;
}
.filter-row {
  display: flex;
  gap: 0;
  flex-wrap: wrap;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  overflow: hidden;
}
.filter-btn {
  padding: 7px 12px;
  font-family: var(--bn-font-mono);
  font-size: 11px;
  border: none;
  border-left: 1px solid var(--bn-border-subtle);
  background: transparent;
  color: var(--bn-text-secondary);
  cursor: pointer;
}
.filter-btn:first-child {
  border-left: none;
}
.filter-btn.active {
  background: var(--bn-ink);
  color: var(--bn-bg-canvas);
  font-weight: 700;
}
.summary-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 24px;
}
.summary-card {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 16px 20px;
}
.summary-label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  display: block;
  margin-bottom: 8px;
}
.summary-value {
  font-family: var(--bn-font-display);
  font-size: 28px;
  font-weight: 900;
  letter-spacing: -0.02em;
}
.text-success { color: var(--bn-trust-verified); }
.loading-row {
  padding: 24px;
  color: var(--bn-text-muted);
  font-size: 14px;
  text-align: center;
}
.load-error {
  padding: 12px 16px;
  background: var(--bn-danger-wash);
  border: 1px solid var(--bn-danger-red);
  color: var(--bn-danger-red);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  margin-bottom: 16px;
}
.empty-row {
  padding: 28px 24px;
  color: var(--bn-text-muted);
  font-size: 14px;
  text-align: center;
  border: 1px dashed var(--bn-border-strong);
  border-radius: var(--bn-radius-md);
}
.list-card {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 16px;
  overflow-x: auto;
}
.campaign-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  min-width: 900px;
}
.campaign-table th {
  text-align: left;
  padding: 10px 12px;
  border-bottom: 2px solid var(--bn-ink);
  font-family: var(--bn-font-mono);
  font-weight: 400;
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  white-space: nowrap;
}
.campaign-table td {
  padding: 12px;
  border-bottom: 1px solid var(--bn-border-subtle);
  vertical-align: middle;
}
.campaign-table tr:last-child td {
  border-bottom: none;
}
.mono {
  font-family: var(--bn-font-mono);
  font-size: 12px;
}
.muted { color: var(--bn-text-muted); }
.text-xs { font-size: 11px; }
.app-name {
  font-family: var(--bn-font-sans);
  font-weight: 700;
  font-size: 13px;
}
.keyword-tag {
  display: inline-block;
  font-family: var(--bn-font-mono);
  font-size: 11px;
  border: 1px solid var(--bn-border-subtle);
  color: var(--bn-text-secondary);
  padding: 2px 8px;
  border-radius: var(--bn-radius-sm);
  margin: 0 4px 4px 0;
}
.progress-track {
  width: 96px;
  background: var(--bn-bg-well);
  border: 1px solid var(--bn-border-subtle);
  height: 6px;
  border-radius: var(--bn-radius-sm);
  overflow: hidden;
  margin-bottom: 6px;
}
.progress-bar {
  background: var(--bn-ink);
  height: 100%;
}
.badge {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.04em;
  padding: 3px 9px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-strong);
  color: var(--bn-text-secondary);
  white-space: nowrap;
}
.badge-success { border-color: var(--bn-trust-verified); color: var(--bn-trust-verified); }
.badge-pending { border-color: var(--bn-seal); color: var(--bn-seal); }
.badge-danger { border-color: var(--bn-danger-red); color: var(--bn-danger-red); }
.badge-muted { border-color: var(--bn-border-subtle); color: var(--bn-text-muted); }
.action-cell {
  display: flex;
  gap: 6px;
}
.btn {
  padding: 5px 12px;
  border-radius: var(--bn-radius-sm);
  cursor: pointer;
  font-size: 12px;
  font-weight: 700;
  white-space: nowrap;
}
.btn:disabled { opacity: 0.45; cursor: not-allowed; }
.btn-success {
  background: transparent;
  border: 1px solid var(--bn-trust-verified);
  color: var(--bn-trust-verified);
}
.btn-success:hover:not(:disabled) {
  background: var(--bn-trust-wash);
}
.btn-danger {
  background: var(--bn-danger-red);
  border: none;
  color: #fff;
}
.btn-danger:hover:not(:disabled) {
  filter: brightness(0.92);
}
.btn-xs { font-size: 11px; }
@media (max-width: 760px) {
  .summary-grid {
    grid-template-columns: 1fr;
  }
}
</style>
