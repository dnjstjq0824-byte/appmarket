<template>
  <div class="promo-code-manager">
    <header class="view-header">
      <div>
        <h2>프로모션 및 리딤 코드 관리</h2>
        <p class="subtitle">앱 할인 쿠폰 및 무료 리딤 프로모션 코드 발급 및 소진 현황</p>
      </div>
      <button class="btn btn-primary" @click="showModal = true">+ 프로모션 코드 일괄 생성</button>
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
        <select id="appSelect" v-model="activeAppId" @change="loadPromotions">
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
          <label>활성 캠페인</label>
          <div class="summary-value">{{ activePromos }}</div>
        </div>
        <div class="summary-card">
          <label>발급된 총 코드 수</label>
          <div class="summary-value text-primary">{{ totalCodesIssued }}</div>
        </div>
        <div class="summary-card">
          <label>리딤 사용 완료</label>
          <div class="summary-value text-success">{{ totalCodesRedeemed }}</div>
        </div>
        <div class="summary-card">
          <label>소진율</label>
          <div class="summary-value">{{ redemptionRate }}%</div>
        </div>
      </div>

      <div v-if="freshCodes.length > 0" class="fresh-codes">
        <h3>방금 발급된 코드 ({{ freshCodes.length }}건 — 지금만 표시됩니다)</h3>
        <div class="code-chips">
          <code v-for="code in freshCodes" :key="code">{{ code }}</code>
        </div>
        <button class="btn btn-secondary btn-sm" @click="downloadFreshCsv">CSV 다운로드</button>
      </div>

      <div class="list-card">
        <h3>생성된 프로모션 목록</h3>
        <div v-if="promotions.length === 0" class="empty-row">생성된 프로모션이 없습니다.</div>
        <table v-else class="promo-table">
          <thead>
            <tr>
              <th>캠페인 이름</th>
              <th>혜택 구분</th>
              <th>사용 수 / 발급 수</th>
              <th>유효 기간</th>
              <th>상태</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in promotions" :key="p.id">
              <td><strong>{{ p.campaignName }}</strong></td>
              <td><span class="benefit-tag">{{ benefitLabel(p) }}</span></td>
              <td class="mono">
                <div>{{ p.usedCount }} / {{ p.codeCount }}</div>
                <div class="progress-track">
                  <div class="progress-bar" :style="{ width: usagePercent(p) + '%' }"></div>
                </div>
              </td>
              <td class="mono muted">{{ formatDate(p.startsAt) }} ~ {{ formatDate(p.endsAt) }}</td>
              <td><span class="badge" :class="promoStatusClass(p)">{{ promoStatus(p) }}</span></td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>

    <div v-if="showModal" class="modal-backdrop">
      <div class="modal">
        <h3>새 프로모션 코드 배치 생성</h3>
        <label class="field">배치 캠페인 이름
          <input v-model="newBatch.name" placeholder="예: 2026 가을 출시 기념 무료 리딤" />
        </label>
        <label class="field">혜택 구분
          <select v-model="newBatch.discountType">
            <option value="FREE_REDEEM">100% 무료 리딤 (소비자 결제액 0원)</option>
            <option value="PERCENTAGE">정률 할인 (%)</option>
            <option value="FIXED_MINOR">정액 할인 (minor units)</option>
          </select>
        </label>
        <label v-if="newBatch.discountType !== 'FREE_REDEEM'" class="field">할인값
          <input v-model.number="newBatch.discountValue" type="number" min="1" />
        </label>
        <div class="field-row">
          <label class="field">시작일
            <input v-model="newBatch.startsAt" type="date" />
          </label>
          <label class="field">종료일
            <input v-model="newBatch.endsAt" type="date" />
          </label>
        </div>
        <div class="field-row">
          <label class="field">코드당 최대 사용
            <input v-model.number="newBatch.maxUses" type="number" min="1" />
          </label>
          <label class="field">생성할 코드 수량
            <input v-model.number="newBatch.codeCount" type="number" min="1" max="1000" />
          </label>
        </div>
        <label class="field">코드 접두사
          <input v-model="newBatch.codePrefix" placeholder="예: FALL26" />
        </label>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showModal = false">취소</button>
          <button class="btn btn-primary" :disabled="isActing" @click="createBatch">
            {{ isActing ? '생성 중…' : '코드 일괄 생성' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { getSharedClient, ApiClientError, type OrgSummary } from '@bookingnetwork/api-client';

const api = getSharedClient();

interface Promotion {
  id: string;
  appId: string;
  campaignName: string;
  discountType: string;
  discountValue: number;
  startsAt: string;
  endsAt: string;
  maxUses: number;
  usedCount: number;
  codeCount: number;
}

interface OrgApp {
  id: string;
  slug: string;
}

const orgs = ref<OrgSummary[]>([]);
const apps = ref<OrgApp[]>([]);
const activeOrgId = ref('');
const activeAppId = ref('');
const promotions = ref<Promotion[]>([]);
const freshCodes = ref<string[]>([]);
const showModal = ref(false);
const isLoading = ref(true);
const loadError = ref('');
const actionError = ref('');
const isActing = ref(false);

const newBatch = ref({
  name: '',
  discountType: 'FREE_REDEEM',
  discountValue: 30,
  startsAt: new Date().toISOString().slice(0, 10),
  endsAt: '',
  maxUses: 1,
  codeCount: 100,
  codePrefix: ''
});

const activePromos = computed(() => promotions.value.filter(p => promoStatus(p) === '진행 중').length);
const totalCodesIssued = computed(() => promotions.value.reduce((sum, p) => sum + p.codeCount, 0));
const totalCodesRedeemed = computed(() => promotions.value.reduce((sum, p) => sum + p.usedCount, 0));
const redemptionRate = computed(() => {
  if (totalCodesIssued.value === 0) return '0.0';
  return ((totalCodesRedeemed.value / totalCodesIssued.value) * 100).toFixed(1);
});

onMounted(async () => {
  try {
    orgs.value = await api.listOrgs();
    if (orgs.value.length > 0) {
      activeOrgId.value = orgs.value[0].id;
      await loadApps();
      if (apps.value.length > 0) {
        activeAppId.value = apps.value[0].id;
      }
      await loadPromotions();
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
  await loadPromotions();
}

async function loadApps() {
  apps.value = await api.listOrgApps(activeOrgId.value) as OrgApp[];
}

async function loadPromotions() {
  if (!activeAppId.value) {
    promotions.value = [];
    return;
  }
  try {
    promotions.value = await api.listPromotions(activeAppId.value, activeOrgId.value) as Promotion[];
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load promotions';
  }
}

async function createBatch() {
  if (!activeAppId.value) return;
  isActing.value = true;
  actionError.value = '';
  try {
    const res = await api.createPromotion(activeAppId.value, {
      orgId: activeOrgId.value,
      campaignName: newBatch.value.name.trim(),
      discountType: newBatch.value.discountType,
      discountValue: newBatch.value.discountType === 'FREE_REDEEM' ? 0 : newBatch.value.discountValue,
      startsAt: new Date(newBatch.value.startsAt + 'T00:00:00Z').toISOString(),
      endsAt: new Date((newBatch.value.endsAt || newBatch.value.startsAt) + 'T23:59:59Z').toISOString(),
      maxUses: newBatch.value.maxUses,
      codeCount: newBatch.value.codeCount,
      codePrefix: newBatch.value.codePrefix.trim()
    }) as { promotion: Promotion; codes: Array<{ code: string } | string> };
    freshCodes.value = (res.codes || []).map(c => (typeof c === 'string' ? c : c.code));
    showModal.value = false;
    await loadPromotions();
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to create promotion';
  } finally {
    isActing.value = false;
  }
}

function downloadFreshCsv() {
  const blob = new Blob([['code', ...freshCodes.value].join('\n')], { type: 'text/csv' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'promo-codes.csv';
  a.click();
  URL.revokeObjectURL(url);
}

function benefitLabel(p: Promotion): string {
  if (p.discountType === 'FREE_REDEEM') return '100% 무료 리딤';
  if (p.discountType === 'PERCENTAGE') return `${p.discountValue}% 할인`;
  return `${Number(p.discountValue).toLocaleString()} 정액 할인`;
}

function usagePercent(p: Promotion): number {
  if (p.codeCount === 0) return 0;
  return Math.min(100, Math.round((p.usedCount / p.codeCount) * 100));
}

function promoStatus(p: Promotion): string {
  const now = new Date();
  if (new Date(p.startsAt) > now) return '시작 전';
  if (new Date(p.endsAt) < now) return '만료';
  if (p.codeCount > 0 && p.usedCount >= p.codeCount * p.maxUses) return '소진';
  return '진행 중';
}

function promoStatusClass(p: Promotion): string {
  const s = promoStatus(p);
  if (s === '진행 중') return 'badge-success';
  if (s === '시작 전') return 'badge-pending';
  return 'badge-muted';
}

function formatDate(iso: string): string {
  return String(iso || '').slice(0, 10);
}
</script>

<style scoped>
.promo-code-manager {
  padding: 24px;
  max-width: 1100px;
  margin: 0 auto;
}
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
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
.text-primary { color: #1d4ed8; }
.text-success { color: #16a34a; }
.fresh-codes {
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 8px;
  padding: 16px 20px;
  margin-bottom: 24px;
}
.fresh-codes h3 {
  margin: 0 0 12px;
  font-size: 14px;
}
.code-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
  max-height: 160px;
  overflow-y: auto;
}
.code-chips code {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  padding: 2px 8px;
  font-size: 12px;
}
.list-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 24px;
  margin-bottom: 24px;
}
.list-card h3 {
  margin: 0 0 16px;
}
.promo-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.promo-table th {
  text-align: left;
  padding: 10px 12px;
  border-bottom: 2px solid #e2e8f0;
  color: #475569;
}
.promo-table td {
  padding: 12px;
  border-bottom: 1px solid #f1f5f9;
  vertical-align: top;
}
.benefit-tag {
  display: inline-block;
  background: #eef2ff;
  color: #4338ca;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}
.mono {
  font-family: ui-monospace, monospace;
  font-size: 12px;
}
.muted { color: #64748b; }
.progress-track {
  width: 96px;
  background: #e2e8f0;
  height: 6px;
  border-radius: 3px;
  overflow: hidden;
  margin-top: 4px;
}
.progress-bar {
  background: #16a34a;
  height: 100%;
}
.badge {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  white-space: nowrap;
}
.badge-success { background: #dcfce7; color: #166534; }
.badge-pending { background: #e0e7ff; color: #3730a3; }
.badge-muted { background: #f1f5f9; color: #64748b; }
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  z-index: 50;
}
.modal {
  background: white;
  border-radius: 12px;
  padding: 24px;
  max-width: 480px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
}
.modal h3 {
  margin: 0 0 16px;
}
.field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  font-weight: 600;
  color: #475569;
  margin-bottom: 12px;
}
.field input, .field select {
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 400;
}
.field-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}
.btn {
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  border: none;
  white-space: nowrap;
}
.btn-primary { background: #2563eb; color: white; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: #f1f5f9; color: #334155; }
.btn-sm { font-size: 12px; padding: 4px 10px; }
</style>
