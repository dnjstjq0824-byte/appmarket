<template>
  <main class="bn-container">
    <section class="bn-library-header">
      <div>
        <h1>내 앱 보관함 및 구독 관리</h1>
        <p class="bn-sub">구매한 유료 앱의 영구 이용 권한과 다운로드 이력을 관리합니다.</p>
      </div>
      <div class="bn-header-stats">
        <div class="bn-stat-box">
          <span class="bn-stat-label">보유 앱</span>
          <strong class="bn-stat-value">{{ entitlements.length }}개</strong>
        </div>
        <div class="bn-stat-box">
          <span class="bn-stat-label">주문 내역</span>
          <strong class="bn-stat-value">{{ orders.length }}건</strong>
        </div>
      </div>
    </section>

    <div v-if="isLoading" class="bn-status">보관함을 불러오는 중…</div>
    <div v-if="loadError" class="bn-status bn-error-text">
      {{ loadError }}
      <NuxtLink v-if="needsLogin" to="/account/login">로그인하기</NuxtLink>
    </div>

    <section v-if="!isLoading && !loadError">
      <h2>보유 앱 (Entitlements)</h2>
      <div v-if="entitlements.length === 0" class="bn-status">보유한 앱이 없습니다. 스토어에서 앱을 구매하면 여기에 표시됩니다.</div>
      <div v-else class="bn-table-wrapper">
        <table class="bn-lib-table">
          <thead>
            <tr>
              <th>앱</th>
              <th>버전</th>
              <th>권한 상태</th>
              <th>획득일</th>
              <th>다운로드</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in enriched" :key="item.entitlement.id">
              <td>
                <NuxtLink :to="`/apps/${item.slug}`" class="bn-app-link">
                  <strong>{{ item.title }}</strong>
                </NuxtLink>
                <div class="bn-mono-sm">{{ item.packageName }}</div>
              </td>
              <td>{{ item.version }}</td>
              <td>
                <span class="bn-status-tag" :class="item.entitlement.status === 'ACTIVE' ? 'active' : 'revoked'">
                  {{ item.entitlement.status }}
                </span>
              </td>
              <td class="bn-mono-sm">{{ item.entitlement.granted_at.slice(0, 10) }}</td>
              <td>
                <button
                  class="bn-btn-secondary"
                  :disabled="downloadingId === item.entitlement.id || !item.releaseId"
                  @click="downloadApp(item)"
                >
                  {{ downloadingId === item.entitlement.id ? '발급 중…' : 'APK 다운로드' }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <h2>주문 내역</h2>
      <div v-if="orders.length === 0" class="bn-status">주문 내역이 없습니다.</div>
      <div v-else class="bn-table-wrapper">
        <table class="bn-lib-table">
          <thead>
            <tr>
              <th>주문 ID</th>
              <th>앱</th>
              <th>결제 금액</th>
              <th>상태</th>
              <th>주문일</th>
              <th>환불</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="order in orders" :key="order.id">
              <td class="bn-mono-sm">{{ order.id }}</td>
              <td>{{ orderTitles[order.app_id] || order.app_id }}</td>
              <td>{{ order.gross_amount.toLocaleString() }} {{ order.currency }}</td>
              <td>
                <span class="bn-status-tag" :class="order.status === 'PAID' ? 'active' : order.status === 'REFUNDED' ? 'revoked' : ''">
                  {{ order.status }}
                </span>
              </td>
              <td class="bn-mono-sm">{{ order.created_at.slice(0, 10) }}</td>
              <td>
                <button
                  v-if="order.status === 'PAID'"
                  class="bn-btn-secondary"
                  @click="openRefund(order)"
                >
                  환불 신청
                </button>
                <span v-else>—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <RefundRequestModal
      :is-open="refundModalOpen"
      :order-id="refundTarget?.id || ''"
      :app-title="refundTarget ? (orderTitles[refundTarget.app_id] || refundTarget.app_id) : ''"
      :amount-won="refundTarget?.gross_amount || 0"
      :purchased-at="refundTarget?.created_at.slice(0, 10) || ''"
      @close="refundModalOpen = false"
      @refund-completed="onRefundCompleted"
    />
  </main>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getSharedClient, ApiClientError, type OrderDTO, type UserEntitlement } from '@bookingnetwork/api-client';
import RefundRequestModal from '~/components/RefundRequestModal.vue';

const api = getSharedClient();

interface EnrichedItem {
  entitlement: UserEntitlement;
  title: string;
  slug: string;
  packageName: string;
  version: string;
  releaseId: string;
}

const isLoading = ref(true);
const loadError = ref('');
const needsLogin = ref(false);
const entitlements = ref<UserEntitlement[]>([]);
const orders = ref<OrderDTO[]>([]);
const enriched = ref<EnrichedItem[]>([]);
const orderTitles = ref<Record<string, string>>({});
const downloadingId = ref('');

const refundModalOpen = ref(false);
const refundTarget = ref<OrderDTO | null>(null);

function deviceId(): string {
  if (typeof window === 'undefined') return 'ssr-device';
  let id = localStorage.getItem('bn_device_id');
  if (!id) {
    id = 'dev-' + Math.random().toString(36).slice(2) + Date.now().toString(36);
    localStorage.setItem('bn_device_id', id);
  }
  return id;
}

async function downloadApp(item: EnrichedItem) {
  downloadingId.value = item.entitlement.id;
  try {
    const signed = await api.createDownloadToken(item.entitlement.app_id, item.releaseId, deviceId());
    if (typeof window !== 'undefined') {
      const a = document.createElement('a');
      a.href = signed.url;
      a.download = `${item.packageName}.apk`;
      a.rel = 'noopener';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
    }
  } catch {
    loadError.value = '다운로드 URL 발급에 실패했습니다.';
  } finally {
    downloadingId.value = '';
  }
}

function openRefund(order: OrderDTO) {
  refundTarget.value = order;
  refundModalOpen.value = true;
}

function onRefundCompleted(orderId: string) {
  refundModalOpen.value = false;
  orders.value = orders.value.map(o => o.id === orderId ? { ...o, status: 'REFUNDED' } : o);
}

onMounted(async () => {
  try {
    const [ents, ords] = await Promise.all([api.listEntitlements(), api.listMyOrders()]);
    entitlements.value = ents;
    orders.value = ords;
    const appIds = [...new Set([...ents.map(e => e.app_id), ...ords.map(o => o.app_id)])];
    const titles: Record<string, string> = {};
    const details = await Promise.all(appIds.map(async (id) => {
      try {
        const d = await api.getAppBySlug(id);
        titles[id] = d.localization?.title || d.app.canonical_slug;
        return d;
      } catch {
        titles[id] = id;
        return null;
      }
    }));
    orderTitles.value = titles;
    const byId: Record<string, (typeof details)[number]> = {};
    appIds.forEach((id, i) => { byId[id] = details[i]; });
    enriched.value = ents.map((entitlement) => {
      const d = byId[entitlement.app_id];
      const dist = d?.distributions[0];
      return {
        entitlement,
        title: titles[entitlement.app_id] || entitlement.app_id,
        slug: d?.app.canonical_slug || entitlement.app_id,
        packageName: d?.app.primary_package_name || '',
        version: dist?.latest_release?.version_name || '—',
        releaseId: dist?.latest_release?.id || ''
      };
    });
  } catch (err) {
    if (err instanceof ApiClientError && err.status === 401) {
      needsLogin.value = true;
      loadError.value = '보관함을 보려면 로그인이 필요합니다.';
    } else {
      loadError.value = err instanceof ApiClientError ? err.message : '보관함을 불러오지 못했습니다.';
    }
  } finally {
    isLoading.value = false;
  }
});
</script>

<style scoped>
.bn-container {
  max-width: 1000px;
  margin: 0 auto;
  padding: 40px 16px;
}
.bn-library-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 32px;
  flex-wrap: wrap;
  gap: 16px;
}
.bn-library-header h1 {
  margin: 0;
  font-size: 24px;
}
.bn-sub {
  color: var(--bn-text-muted, #64748B);
  font-size: 14px;
  margin: 6px 0 0;
}
.bn-header-stats {
  display: flex;
  gap: 16px;
}
.bn-stat-box {
  display: flex;
  flex-direction: column;
}
.bn-stat-label {
  font-size: 11px;
  color: var(--bn-text-muted, #64748B);
}
.bn-stat-value {
  font-size: 16px;
}
.bn-status {
  padding: 24px;
  text-align: center;
  color: var(--bn-text-muted, #64748B);
  font-size: 14px;
}
.bn-error-text {
  color: #b91c1c;
}
.bn-table-wrapper {
  overflow-x: auto;
  margin-bottom: 32px;
}
.bn-lib-table {
  width: 100%;
  border-collapse: collapse;
  background-color: var(--bn-bg-surface, #fff);
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  border-radius: 8px;
  font-size: 14px;
}
.bn-lib-table th, .bn-lib-table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid var(--bn-border-subtle, #E2E8F0);
}
.bn-lib-table th {
  background-color: var(--bn-bg-elevated, #F8FAFC);
  font-weight: 600;
  color: var(--bn-text-secondary, #475569);
  font-size: 12px;
  text-transform: uppercase;
}
.bn-app-link {
  color: inherit;
  text-decoration: none;
}
.bn-app-link:hover {
  text-decoration: underline;
}
.bn-mono-sm {
  font-family: monospace;
  font-size: 11px;
  color: var(--bn-text-muted, #64748B);
  margin-top: 4px;
}
.bn-status-tag {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 4px;
  background-color: var(--bn-bg-elevated, #F1F5F9);
  color: var(--bn-text-muted, #64748B);
}
.bn-status-tag.active {
  background-color: #ECFDF5;
  color: #059669;
}
.bn-status-tag.revoked {
  background-color: #FEF2F2;
  color: #B91C1C;
}
.bn-btn-secondary {
  background-color: var(--bn-bg-elevated, #F1F5F9);
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  border-radius: 6px;
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}
.bn-btn-secondary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
