<template>
  <div class="ad-campaign-manager">
    <header class="view-header">
      <div>
        <h2>스폰서 검색광고 캠페인 매니저</h2>
        <p class="subtitle">검색 결과 스폰서 배치, CPC 입찰가 관리, 일일 예산 캡</p>
      </div>
    </header>

    <div class="select-row">
      <div class="select-group">
        <label for="orgSelect">조직</label>
        <select id="orgSelect" v-model="activeOrgId" @change="onOrgChange">
          <option v-for="org in orgs" :key="org.id" :value="org.id">{{ org.org_name }}</option>
        </select>
      </div>
      <div class="select-group">
        <label for="appSelect">앱</label>
        <select id="appSelect" v-model="activeAppId" @change="loadCampaigns">
          <option v-for="app in apps" :key="app.id" :value="app.id">{{ app.slug }}</option>
        </select>
      </div>
    </div>
    <div v-if="isLoading" class="loading-row">불러오는 중…</div>
    <div v-if="loadError" class="load-error">{{ loadError }}</div>
    <div v-if="actionError" class="load-error">{{ actionError }}</div>

    <template v-if="!isLoading && !loadError">
      <div class="summary-grid">
        <div class="summary-card">
          <label>전체 캠페인</label>
          <div class="summary-value">{{ campaigns.length }}</div>
        </div>
        <div class="summary-card">
          <label>활성 (ACTIVE)</label>
          <div class="summary-value text-success">{{ activeCount }}</div>
        </div>
        <div class="summary-card">
          <label>일시중지</label>
          <div class="summary-value text-amber">{{ pausedCount }}</div>
        </div>
        <div class="summary-card">
          <label>심사 대기</label>
          <div class="summary-value">{{ pendingCount }}</div>
        </div>
      </div>

      <form class="create-card" @submit.prevent="createCampaign">
        <h3>새 스폰서 캠페인 생성</h3>
        <div class="form-grid">
          <label>대상 앱 표시명
            <input v-model="form.appName" required placeholder="Super Office Suite" />
          </label>
          <label>타겟 키워드 (쉼표 구분)
            <input v-model="form.keywords" required placeholder="office, document, pdf" />
          </label>
          <label>최대 CPC 입찰가 (원)
            <input v-model.number="form.maxBidCpc" type="number" min="1" required />
          </label>
          <label>일일 예산 (원)
            <input v-model.number="form.dailyBudget" type="number" min="1" required />
          </label>
        </div>
        <button type="submit" class="btn btn-primary" :disabled="isActing || !activeAppId">
          {{ isActing ? '생성 중…' : '캠페인 생성' }}
        </button>
      </form>

      <div class="list-card">
        <h3>캠페인 목록</h3>
        <div v-if="campaigns.length === 0" class="empty-row">운영 중인 캠페인이 없습니다.</div>
        <table v-else class="campaign-table">
          <thead>
            <tr>
              <th>캠페인 / 대상 앱</th>
              <th>타겟 키워드</th>
              <th>CPC 입찰가</th>
              <th>일일 예산</th>
              <th>상태</th>
              <th>제어</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="c in campaigns" :key="c.id">
              <td>
                <div class="campaign-id">{{ c.id }}</div>
                <div class="campaign-app">{{ c.appName }}</div>
              </td>
              <td>
                <span v-for="k in c.targetKeywords" :key="k" class="keyword-tag">{{ k }}</span>
              </td>
              <td class="mono">₩{{ Number(c.maxBidCpc).toLocaleString() }}</td>
              <td class="mono">₩{{ Number(c.dailyBudget).toLocaleString() }}</td>
              <td>
                <span class="badge" :class="statusClass(c.status)">{{ c.status }}</span>
              </td>
              <td>
                <button
                  v-if="c.status === 'ACTIVE' || c.status === 'PAUSED_MANUAL'"
                  class="btn btn-sm"
                  :class="c.status === 'ACTIVE' ? 'btn-warning' : 'btn-success'"
                  :disabled="isActing"
                  @click="togglePaused(c)"
                >
                  {{ c.status === 'ACTIVE' ? '일시중지' : '재개' }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { getSharedClient, ApiClientError, type OrgSummary } from '@bookingnetwork/api-client';

const api = getSharedClient();

interface Campaign {
  id: string;
  appId: string;
  appName: string;
  targetKeywords: string[];
  maxBidCpc: number;
  dailyBudget: number;
  status: string;
  createdAt: string;
}

interface OrgApp {
  id: string;
  slug: string;
}

const orgs = ref<OrgSummary[]>([]);
const apps = ref<OrgApp[]>([]);
const activeOrgId = ref('');
const activeAppId = ref('');
const campaigns = ref<Campaign[]>([]);
const form = ref({ appName: '', keywords: '', maxBidCpc: 500, dailyBudget: 100000 });
const isLoading = ref(true);
const loadError = ref('');
const actionError = ref('');
const isActing = ref(false);

const activeCount = computed(() => campaigns.value.filter(c => c.status === 'ACTIVE').length);
const pausedCount = computed(() => campaigns.value.filter(c => c.status === 'PAUSED_MANUAL' || c.status === 'PAUSED_DAILY_BUDGET_CAP').length);
const pendingCount = computed(() => campaigns.value.filter(c => c.status === 'PENDING_REVIEW').length);

onMounted(async () => {
  try {
    orgs.value = await api.listOrgs();
    if (orgs.value.length > 0) {
      activeOrgId.value = orgs.value[0].id;
      await loadApps();
      if (apps.value.length > 0) {
        activeAppId.value = apps.value[0].id;
      }
      await loadCampaigns();
    }
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load organizations';
  } finally {
    isLoading.value = false;
  }
});

async function onOrgChange() {
  await loadApps();
  activeAppId.value = apps.value.length > 0 ? apps.value[0].id : '';
  await loadCampaigns();
}

async function loadApps() {
  apps.value = await api.listOrgApps(activeOrgId.value) as OrgApp[];
}

async function loadCampaigns() {
  if (!activeOrgId.value) return;
  const appHandle = activeAppId.value || 'current';
  try {
    campaigns.value = await api.listCampaigns(appHandle, activeOrgId.value) as Campaign[];
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load campaigns';
  }
}

async function createCampaign() {
  if (!activeAppId.value) return;
  isActing.value = true;
  actionError.value = '';
  try {
    await api.createCampaign(activeAppId.value, {
      orgId: activeOrgId.value,
      appName: form.value.appName.trim(),
      targetKeywords: form.value.keywords.split(',').map(k => k.trim()).filter(Boolean),
      maxBidCpc: form.value.maxBidCpc,
      dailyBudget: form.value.dailyBudget
    });
    form.value = { appName: '', keywords: '', maxBidCpc: 500, dailyBudget: 100000 };
    await loadCampaigns();
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to create campaign';
  } finally {
    isActing.value = false;
  }
}

async function togglePaused(c: Campaign) {
  isActing.value = true;
  actionError.value = '';
  try {
    await api.setCampaignPaused(activeAppId.value || 'current', c.id, c.status === 'ACTIVE', activeOrgId.value);
    await loadCampaigns();
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to update campaign';
  } finally {
    isActing.value = false;
  }
}

function statusClass(status: string): string {
  switch (status) {
    case 'ACTIVE': return 'badge-success';
    case 'PAUSED_MANUAL':
    case 'PAUSED_DAILY_BUDGET_CAP': return 'badge-warning';
    case 'REJECTED': return 'badge-danger';
    default: return 'badge-pending';
  }
}
</script>

<style scoped>
.ad-campaign-manager {
  max-width: 1080px;
  margin: 0 auto;
}
.view-header {
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 2px solid var(--bn-ink);
}
.view-header h2 {
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
.select-row {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.select-group {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--bn-text-secondary);
}
.select-group label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
}
.select-group select {
  padding: 8px 12px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  font-family: var(--bn-font-sans);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  min-width: 200px;
}
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
.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 24px;
}
.summary-card {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 16px 18px;
}
.summary-card label {
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
.text-amber { color: var(--bn-warning-amber); }
.create-card, .list-card {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 24px;
  margin-bottom: 24px;
}
.create-card h3, .list-card h3 {
  font-family: var(--bn-font-display);
  font-size: 15px;
  font-weight: 800;
  margin: 0 0 16px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--bn-border-subtle);
}
.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-bottom: 16px;
}
.form-grid label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.06em;
  color: var(--bn-text-muted);
}
.form-grid input {
  padding: 9px 12px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  font-family: var(--bn-font-sans);
  font-size: 13px;
}
.form-grid input:focus {
  outline: none;
  border-color: var(--bn-ink);
}
.campaign-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
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
.campaign-id {
  font-family: var(--bn-font-mono);
  font-size: 12px;
  font-weight: 700;
}
.campaign-app {
  font-size: 12px;
  color: var(--bn-text-muted);
  margin-top: 2px;
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
.mono {
  font-family: var(--bn-font-mono);
  font-size: 12px;
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
.badge-warning { border-color: var(--bn-warning-amber); color: var(--bn-warning-amber); }
.badge-danger { border-color: var(--bn-danger-red); color: var(--bn-danger-red); }
.badge-pending { border-color: var(--bn-seal); color: var(--bn-seal); }
.btn {
  padding: 9px 18px;
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}
.btn-primary {
  background: var(--bn-action-bg);
  color: var(--bn-action-text);
  border: none;
}
.btn-primary:hover:not(:disabled) {
  background: var(--bn-action-bg-hover);
}
.btn-warning {
  background: transparent;
  border: 1px solid var(--bn-warning-amber);
  color: var(--bn-warning-amber);
}
.btn-warning:hover:not(:disabled) {
  background: var(--bn-warning-wash);
}
.btn-success {
  background: transparent;
  border: 1px solid var(--bn-trust-verified);
  color: var(--bn-trust-verified);
}
.btn-success:hover:not(:disabled) {
  background: var(--bn-trust-wash);
}
.btn-sm { font-size: 12px; padding: 5px 12px; }
.btn:disabled { opacity: 0.45; cursor: not-allowed; }
@media (max-width: 900px) {
  .summary-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (max-width: 560px) {
  .summary-grid, .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
