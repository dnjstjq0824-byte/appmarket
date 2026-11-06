<template>
  <div class="app-pricing-editor">
    <header class="view-header">
      <div>
        <h2>유료 앱 국가별 가격표 및 정산 시뮬레이터</h2>
        <p class="subtitle">부가가치세(VAT) 포함 공급가, 플랫폼 수수료(15%) 및 개발사 실수령액 실시간 산출</p>
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
        <select id="appSelect" v-model="activeAppId" @change="loadTiers">
          <option v-for="app in apps" :key="app.id" :value="app.id">{{ app.slug }}</option>
        </select>
      </div>
    </div>
    <div v-if="isLoading" class="loading-row">불러오는 중…</div>
    <div v-if="loadError" class="load-error">{{ loadError }}</div>
    <div v-if="actionError" class="load-error">{{ actionError }}</div>

    <template v-if="!isLoading && !loadError">
      <div class="grid-2">
        <div class="card">
          <h3>기준 판매가 설정</h3>
          <label class="field">국가 코드
            <input v-model="form.countryCode" maxlength="2" placeholder="KR" />
          </label>
          <label class="field">통화 (ISO)
            <input v-model="form.currency" maxlength="3" placeholder="KRW" />
          </label>
          <label class="field">소비자 최종 결제 가격 (minor units)
            <input v-model.number="form.listPriceMinor" type="number" min="0" step="100" />
          </label>
          <label class="check-row">
            <input v-model="form.taxInclusive" type="checkbox" />
            세금 포함가 (Tax Inclusive)
          </label>
          <button class="btn btn-primary" :disabled="isActing || !activeAppId" @click="saveTier">
            {{ isActing ? '저장 중…' : '가격표 저장 및 배포 반영' }}
          </button>
        </div>

        <div class="card">
          <h3>건당 정산액 실시간 분해</h3>
          <div class="breakdown">
            <div class="break-row"><span>총 결제액 (Gross Sales)</span><strong>₩{{ gross.toLocaleString() }}</strong></div>
            <div class="break-row"><span>부가가치세 10% (VAT)</span><span class="text-danger">- ₩{{ vatAmount.toLocaleString() }}</span></div>
            <div class="break-row"><span>공급가액 (Net Amount)</span><span>₩{{ netAmount.toLocaleString() }}</span></div>
            <div class="break-row"><span>마켓 수수료 (15%)</span><span class="text-amber">- ₩{{ commissionAmount.toLocaleString() }}</span></div>
            <div class="break-row total"><span>개발사 최종 수령액 (Net Payout)</span><strong>₩{{ developerPayout.toLocaleString() }}</strong></div>
          </div>
          <p class="verify-note">정산 공식 검증: ₩{{ developerPayout.toLocaleString() }} + ₩{{ commissionAmount.toLocaleString() }} + ₩{{ vatAmount.toLocaleString() }} = ₩{{ (developerPayout + commissionAmount + vatAmount).toLocaleString() }}</p>
        </div>
      </div>

      <div class="card">
        <h3>등록된 국가별 가격표</h3>
        <div v-if="tiers.length === 0" class="empty-row">등록된 가격표가 없습니다. 무료 앱은 결제 없이 다운로드됩니다.</div>
        <table v-else class="tier-table">
          <thead>
            <tr>
              <th>국가</th>
              <th>통화</th>
              <th>판매가 (minor)</th>
              <th>세금 포함</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="t in tiers" :key="`${t.countryCode}-${t.currency}`">
              <td class="mono">{{ t.countryCode }}</td>
              <td class="mono">{{ t.currency }}</td>
              <td class="mono">{{ Number(t.listPriceMinor).toLocaleString() }}</td>
              <td>{{ t.taxInclusive ? '포함' : '별도' }}</td>
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

interface PriceTier {
  countryCode: string;
  currency: string;
  listPriceMinor: number;
  taxInclusive: boolean;
}

interface OrgApp {
  id: string;
  slug: string;
}

const orgs = ref<OrgSummary[]>([]);
const apps = ref<OrgApp[]>([]);
const activeOrgId = ref('');
const activeAppId = ref('');
const tiers = ref<PriceTier[]>([]);
const form = ref({ countryCode: 'KR', currency: 'KRW', listPriceMinor: 11000, taxInclusive: true });
const isLoading = ref(true);
const loadError = ref('');
const actionError = ref('');
const isActing = ref(false);

// Exact integer arithmetic mirroring the backend PricingCalculator:
// net = floor(gross * 10000 / 11000), vat = gross - net,
// commission = floor(net * 1500 / 10000), payout = gross - vat - commission.
const gross = computed(() => Math.max(0, Math.floor(form.value.listPriceMinor || 0)));
const netAmount = computed(() => Math.floor((gross.value * 10000) / 11000));
const vatAmount = computed(() => gross.value - netAmount.value);
const commissionAmount = computed(() => Math.floor((netAmount.value * 1500) / 10000));
const developerPayout = computed(() => gross.value - vatAmount.value - commissionAmount.value);

onMounted(async () => {
  try {
    orgs.value = await api.listOrgs();
    if (orgs.value.length > 0) {
      activeOrgId.value = orgs.value[0].id;
      await loadApps();
      if (apps.value.length > 0) {
        activeAppId.value = apps.value[0].id;
      }
      await loadTiers();
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
  await loadTiers();
}

async function loadApps() {
  apps.value = await api.listOrgApps(activeOrgId.value) as OrgApp[];
}

async function loadTiers() {
  if (!activeAppId.value) {
    tiers.value = [];
    return;
  }
  try {
    tiers.value = await api.listPriceTiers(activeAppId.value, activeOrgId.value) as PriceTier[];
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load price tiers';
  }
}

async function saveTier() {
  if (!activeAppId.value) return;
  isActing.value = true;
  actionError.value = '';
  try {
    await api.setPriceTier(activeAppId.value, {
      orgId: activeOrgId.value,
      countryCode: form.value.countryCode.trim().toUpperCase(),
      currency: form.value.currency.trim().toUpperCase(),
      listPriceMinor: Math.floor(form.value.listPriceMinor || 0),
      taxInclusive: form.value.taxInclusive
    });
    await loadTiers();
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to save price tier';
  } finally {
    isActing.value = false;
  }
}
</script>

<style scoped>
.app-pricing-editor {
  padding: 24px;
  max-width: 1000px;
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
.grid-2 {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 24px;
  margin-bottom: 24px;
}
.card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 24px;
  margin-bottom: 24px;
}
.grid-2 .card {
  margin-bottom: 0;
}
.card h3 {
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
.field input {
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 400;
}
.check-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  margin-bottom: 16px;
}
.breakdown {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 13px;
}
.break-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 12px;
  background: #f8fafc;
  border-radius: 6px;
}
.break-row.total {
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  color: #166534;
  font-size: 14px;
}
.text-danger { color: #dc2626; }
.text-amber { color: #d97706; }
.verify-note {
  font-size: 11px;
  color: #64748b;
  margin: 12px 0 0;
}
.tier-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.tier-table th {
  text-align: left;
  padding: 10px 12px;
  border-bottom: 2px solid #e2e8f0;
  color: #475569;
}
.tier-table td {
  padding: 10px 12px;
  border-bottom: 1px solid #f1f5f9;
}
.mono {
  font-family: ui-monospace, monospace;
  font-size: 12px;
}
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
</style>
