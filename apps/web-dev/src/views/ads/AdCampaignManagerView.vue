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
  padding: 24px;
  max-width: 1100px;
  margin: 0 auto;
}
.view-header {
  margin-bottom: 24px;
}
.view-header h2 {
  margin: 0 0 4px;
}
.subtitle {
  color: #64748b;
  font-size: 14px;
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
}
.select-group select {
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
  background: white;
  min-width: 200px;
}
.loading-row {
  padding: 16px;
  color: #64748b;
  font-size: 14px;
}
.load-error {
  padding: 12px 16px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #991b1b;
  border-radius: 6px;
  font-size: 13px;
  margin-bottom: 16px;
}
.empty-row {
  padding: 24px;
  color: #64748b;
  font-size: 14px;
  text-align: center;
}
.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}
.summary-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px 20px;
}
.summary-card label {
  font-size: 12px;
  color: #64748b;
  display: block;
  margin-bottom: 6px;
}
.summary-value {
  font-size: 24px;
  font-weight: 700;
  color: #0f172a;
}
.text-success { color: #16a34a; }
.text-amber { color: #d97706; }
.create-card, .list-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 24px;
  margin-bottom: 24px;
}
.create-card h3, .list-card h3 {
  margin: 0 0 16px;
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
  gap: 4px;
  font-size: 12px;
  font-weight: 600;
  color: #475569;
}
.form-grid input {
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 400;
}
.campaign-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.campaign-table th {
  text-align: left;
  padding: 10px 12px;
  border-bottom: 2px solid #e2e8f0;
  color: #475569;
  font-weight: 600;
}
.campaign-table td {
  padding: 12px;
  border-bottom: 1px solid #f1f5f9;
  vertical-align: top;
}
.campaign-id {
  font-family: ui-monospace, monospace;
  font-size: 12px;
  font-weight: 700;
  color: #1d4ed8;
}
.campaign-app {
  font-size: 12px;
  color: #64748b;
}
.keyword-tag {
  display: inline-block;
  background: #f1f5f9;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  margin: 0 4px 4px 0;
}
.mono {
  font-family: ui-monospace, monospace;
  font-size: 12px;
}
.badge {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  white-space: nowrap;
}
.badge-success { background: #dcfce7; color: #166534; }
.badge-warning { background: #fef3c7; color: #92400e; }
.badge-danger { background: #fee2e2; color: #991b1b; }
.badge-pending { background: #e0e7ff; color: #3730a3; }
.btn {
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  border: none;
}
.btn-primary { background: #2563eb; color: white; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-warning { background: #f59e0b; color: white; }
.btn-success { background: #16a34a; color: white; }
.btn-sm { font-size: 12px; padding: 4px 10px; }
.btn:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
