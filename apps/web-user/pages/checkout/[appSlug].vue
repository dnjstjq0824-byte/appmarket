<template>
  <div class="co-page">
    <div class="co-wrap">
      <!-- Breadcrumb & Title -->
      <div class="co-head">
        <a href="/" class="co-back">← 마켓 둘러보기로 돌아가기</a>
        <h1 class="co-title">유료 앱 주문 및 안전 결제 (USR-13)</h1>
        <p class="co-sub">구매 즉시 다운로드 권한이 부여되며, 영구 라이선스로 관리됩니다.</p>
      </div>

      <div v-if="isLoading" class="co-status">주문 정보를 불러오는 중…</div>
      <div v-if="loadError" class="co-load-error">{{ loadError }}</div>

      <!-- App Card Summary -->
      <div class="co-card co-app">
        <div class="co-icon">{{ appInfo.iconLetter }}</div>
        <div class="co-app-info">
          <div class="co-dev">{{ appInfo.developerName }}</div>
          <h2 class="co-app-title">{{ appInfo.title }}</h2>
          <div class="co-pkg">패키지: {{ appInfo.packageName }} • v{{ appInfo.version }}</div>
        </div>
      </div>

      <!-- Order Quote & Price Breakdown -->
      <div class="co-card">
        <h3 class="co-card-title">결제 금액 산출서 (Order Quote)</h3>

        <div class="co-quote">
          <div class="co-row">
            <span>앱 정가 (Gross List Price)</span>
            <span>₩{{ basePriceWon.toLocaleString() }}</span>
          </div>

          <div v-if="discountWon > 0" class="co-row co-discount">
            <span>프로모션 할인 적용 ({{ appliedPromoCode }})</span>
            <span>- ₩{{ discountWon.toLocaleString() }}</span>
          </div>

          <div class="co-row co-dim co-divide">
            <span>공급가액 (Net Amount)</span>
            <span>₩{{ netAmountWon.toLocaleString() }}</span>
          </div>

          <div class="co-row co-dim">
            <span>부가가치세 10% (VAT)</span>
            <span>₩{{ vatAmountWon.toLocaleString() }}</span>
          </div>

          <div class="co-row co-total">
            <span>최종 결제 금액 (Payable)</span>
            <span class="co-total-amount">₩{{ finalPayableWon.toLocaleString() }}</span>
          </div>
        </div>

        <!-- Promo Code Input Component -->
        <div class="co-promo">
          <PromoRedeemInput :app-id="appId" @applied="onPromoApplied" @removed="onPromoRemoved" />
        </div>
        <div v-if="orderError" class="co-order-error">{{ orderError }}</div>
      </div>

      <!-- Payment Method Selection -->
      <div class="co-card">
        <h3 class="co-card-title">결제 진행 방식</h3>
        <p class="co-note">
          주문 생성 후 PG사 결제창으로 이동합니다. PG사 웹훅 HMAC-SHA256 서명 검증으로 무결성이 보장되며 멱등 처리됩니다.
        </p>
        <div v-if="checkoutProvider" class="co-provider">
          결제 대행사: <strong>{{ checkoutProvider }}</strong>
        </div>
      </div>

      <!-- Checkout CTA -->
      <div class="co-cta">
        <button
          @click="executePayment"
          :disabled="isProcessing || isLoading || !appId || basePriceWon <= 0 || !releaseId"
          class="co-pay-btn"
        >
          <span v-if="isProcessing">결제 승인 처리 중...</span>
          <span v-else>
            {{ finalPayableWon === 0 ? '무료 리딤 완료 및 다운로드' : `₩${finalPayableWon.toLocaleString()} 안전 결제하기` }}
          </span>
        </button>

        <p class="co-fineprint">
          구매 후 14일 이내 미다운로드 시 전액 환불(청약철회) 가능합니다.
        </p>
      </div>

      <!-- Purchase Success Dialog -->
      <div v-if="purchaseSuccess" class="co-dialog-backdrop">
        <div class="co-dialog">
          <div class="co-dialog-check">✓</div>
          <h3 class="co-dialog-title">결제 완료 및 소유 권한 발급</h3>
          <p class="co-dialog-desc">
            [{{ appInfo.title }}]의 정품 라이선스가 발급되었습니다. 지금 바로 고속 CDN을 통해 안전하게 다운로드하세요.
          </p>

          <div class="co-receipt">
            <div>주문 번호: <span class="co-receipt-hl">{{ orderId }}</span></div>
            <div>결제 금액: ₩{{ finalPayableWon.toLocaleString() }}</div>
            <div>권한 상태: <span class="co-receipt-active">ACTIVE (인증됨)</span></div>
          </div>

          <div class="co-dialog-actions">
            <a :href="downloadUrl" class="co-download-btn">
              APK 다운로드 시작 (15분 유효 서명 토큰)
            </a>
            <a href="/account/library" class="co-library-btn">
              내 보관함으로 이동
            </a>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { getSharedClient, ApiClientError } from '@bookingnetwork/api-client';
import PromoRedeemInput from '../../components/PromoRedeemInput.vue';

const api = getSharedClient();
const route = useRoute();
const router = useRouter();
const slug = route.params.appSlug as string;

const appInfo = ref({
  title: '',
  developerName: '',
  packageName: '',
  version: '',
  iconLetter: '?'
});
const appId = ref('');
const releaseId = ref('');
const isLoading = ref(true);
const loadError = ref('');

const basePriceWon = ref(0);
const discountWon = ref(0);
const appliedPromoCode = ref('');
const isProcessing = ref(false);
const purchaseSuccess = ref(false);
const orderId = ref('');
const downloadUrl = ref('');
const orderError = ref('');
const checkoutProvider = ref('');
const idempotencyKey = ref('');

const finalPayableWon = computed(() => Math.max(0, basePriceWon.value - discountWon.value));
const netAmountWon = computed(() => Math.floor((finalPayableWon.value * 10000) / 11000));
const vatAmountWon = computed(() => finalPayableWon.value - netAmountWon.value);

function onPromoApplied(promo: { code: string; type: 'FREE' | 'PERCENT' | 'FIXED'; percent?: number; fixedMinor?: number }) {
  appliedPromoCode.value = promo.code;
  if (promo.type === 'FREE') {
    discountWon.value = basePriceWon.value;
  } else if (promo.type === 'PERCENT' && promo.percent) {
    discountWon.value = basePriceWon.value - Math.floor((basePriceWon.value * (100 - promo.percent)) / 100);
  } else if (promo.type === 'FIXED' && promo.fixedMinor) {
    discountWon.value = Math.min(basePriceWon.value, promo.fixedMinor);
  }
}

function onPromoRemoved() {
  appliedPromoCode.value = '';
  discountWon.value = 0;
}

function deviceId(): string {
  if (typeof window === 'undefined') return 'ssr-device';
  let id = localStorage.getItem('bn_device_id');
  if (!id) {
    id = 'dev-' + Math.random().toString(36).slice(2) + Date.now().toString(36);
    localStorage.setItem('bn_device_id', id);
  }
  return id;
}

async function executePayment() {
  if (!appId.value || basePriceWon.value <= 0) return;
  isProcessing.value = true;
  orderError.value = '';
  try {
    if (!idempotencyKey.value) {
      idempotencyKey.value = 'co-' + Date.now().toString(36) + '-' + Math.random().toString(36).slice(2, 10);
    }
    const order = await api.createOrder({
      app_id: appId.value,
      gross_amount: finalPayableWon.value,
      currency: 'KRW',
      idempotency_key: idempotencyKey.value,
      promo_code: appliedPromoCode.value || undefined
    });
    orderId.value = order.id;
    if (order.status === 'PAID') {
      const signed = await api.createDownloadToken(order.app_id, releaseId.value, deviceId());
      downloadUrl.value = signed.url;
      purchaseSuccess.value = true;
      return;
    }
    const session = await api.checkoutOrder(order.id);
    checkoutProvider.value = session.provider;
    if (typeof window !== 'undefined') {
      window.location.href = session.checkout_url;
    }
  } catch (err) {
    if (err instanceof ApiClientError && err.status === 401) {
      await router.push({ path: '/account/login', query: { redirect: route.fullPath } });
      return;
    }
    orderError.value = err instanceof ApiClientError ? err.message : '주문 생성에 실패했습니다.';
  } finally {
    isProcessing.value = false;
  }
}

onMounted(async () => {
  try {
    const detail = await api.getAppBySlug(slug);
    appId.value = detail.app.id;
    const dist = detail.distributions[0];
    releaseId.value = dist?.latest_release?.id || '';
    appInfo.value = {
      title: detail.localization?.title || detail.app.canonical_slug,
      developerName: detail.app.developer_org_id || detail.app.primary_package_name,
      packageName: detail.app.primary_package_name,
      version: dist?.latest_release?.version_name || '',
      iconLetter: (detail.localization?.title || '?').trim().slice(0, 1).toUpperCase()
    };
    const tiers = await api.getAppPrices(detail.app.id);
    const krw = tiers.find(t => t.currency === 'KRW') || tiers[0];
    if (!krw) {
      loadError.value = '이 앱에는 등록된 가격이 없습니다.';
      return;
    }
    basePriceWon.value = krw.list_price_minor;
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : '주문 정보를 불러오지 못했습니다.';
  } finally {
    isLoading.value = false;
  }
});
</script>

<style scoped>
.co-page {
  min-height: 100vh;
  background: #020617;
  color: #f1f5f9;
  padding: 24px;
}
@media (min-width: 768px) {
  .co-page { padding: 48px; }
}
.co-wrap {
  max-width: 672px;
  margin: 0 auto;
}
.co-head { margin-bottom: 24px; }
.co-back {
  font-size: 12px;
  color: #38bdf8;
  text-decoration: none;
}
.co-back:hover { text-decoration: underline; }
.co-title {
  font-size: 24px;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: #f1f5f9;
  margin: 8px 0 0;
}
.co-sub {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 4px;
}
.co-status {
  font-size: 14px;
  color: #94a3b8;
  margin-bottom: 24px;
}
.co-load-error {
  font-size: 14px;
  color: #fb7185;
  background: #4c0519;
  border: 1px solid #881337;
  border-radius: 4px;
  padding: 12px;
  margin-bottom: 24px;
}
.co-card {
  background: #0f172a;
  border: 1px solid #1e293b;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 24px;
}
.co-app {
  display: flex;
  gap: 16px;
  align-items: center;
  padding: 20px;
}
.co-icon {
  width: 64px;
  height: 64px;
  background: linear-gradient(135deg, #6366f1, #0284c7);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: 700;
  box-shadow: 0 10px 15px rgba(0, 0, 0, 0.4);
  flex-shrink: 0;
}
.co-app-info { flex: 1; }
.co-dev {
  font-size: 12px;
  color: #38bdf8;
  font-weight: 500;
}
.co-app-title {
  font-size: 18px;
  font-weight: 700;
  color: #f1f5f9;
  margin: 0;
}
.co-pkg {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 2px;
  font-family: monospace;
}
.co-card-title {
  font-size: 14px;
  font-weight: 600;
  color: #e2e8f0;
  margin: 0 0 16px;
}
.co-quote {
  font-family: monospace;
  font-size: 12px;
}
.co-quote > div + div { margin-top: 12px; }
.co-row {
  display: flex;
  justify-content: space-between;
  color: #cbd5e1;
}
.co-discount { color: #34d399; }
.co-dim { color: #94a3b8; }
.co-divide {
  padding-top: 8px;
  border-top: 1px solid #1e293b;
}
.co-total {
  font-size: 16px;
  font-weight: 700;
  color: #f1f5f9;
  padding-top: 12px;
  border-top: 1px solid #334155;
}
.co-total-amount { color: #38bdf8; }
.co-promo { margin-top: 24px; }
.co-order-error {
  margin-top: 16px;
  font-size: 12px;
  color: #fb7185;
}
.co-note {
  font-size: 12px;
  color: #94a3b8;
  margin: 0;
}
.co-provider {
  margin-top: 12px;
  font-size: 12px;
  color: #cbd5e1;
}
.co-cta > * + * { margin-top: 12px; }
.co-pay-btn {
  width: 100%;
  padding: 14px;
  background: linear-gradient(90deg, #0284c7, #4f46e5);
  color: #fff;
  font-weight: 700;
  border: none;
  border-radius: 12px;
  box-shadow: 0 10px 15px rgba(0, 0, 0, 0.4);
  font-size: 14px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  transition: filter 0.15s;
}
.co-pay-btn:hover { filter: brightness(1.1); }
.co-pay-btn:disabled {
  background: #1e293b;
  cursor: not-allowed;
  filter: none;
}
.co-fineprint {
  text-align: center;
  font-size: 11px;
  color: #64748b;
}
.co-dialog-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  z-index: 50;
}
.co-dialog {
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 12px;
  padding: 24px;
  max-width: 448px;
  width: 100%;
  text-align: center;
}
.co-dialog-check {
  width: 48px;
  height: 48px;
  background: #052e16;
  color: #34d399;
  border-radius: 9999px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
  font-size: 24px;
  border: 1px solid rgba(16, 185, 129, 0.4);
}
.co-dialog-title {
  font-size: 18px;
  font-weight: 700;
  color: #f1f5f9;
  margin: 0 0 4px;
}
.co-dialog-desc {
  font-size: 12px;
  color: #94a3b8;
  margin: 0 0 24px;
}
.co-receipt {
  background: #020617;
  padding: 12px;
  border-radius: 8px;
  font-size: 12px;
  font-family: monospace;
  text-align: left;
  margin-bottom: 24px;
  color: #94a3b8;
}
.co-receipt > div + div { margin-top: 4px; }
.co-receipt-hl { color: #38bdf8; }
.co-receipt-active {
  color: #34d399;
  font-weight: 700;
}
.co-dialog-actions > * + * { margin-top: 8px; }
.co-download-btn {
  display: block;
  width: 100%;
  padding: 10px;
  background: #059669;
  color: #fff;
  font-weight: 700;
  font-size: 12px;
  border-radius: 8px;
  text-decoration: none;
  transition: background 0.15s;
}
.co-download-btn:hover { background: #10b981; }
.co-library-btn {
  display: block;
  width: 100%;
  padding: 8px;
  background: #1e293b;
  color: #cbd5e1;
  font-size: 12px;
  border-radius: 8px;
  text-decoration: none;
  transition: background 0.15s;
}
.co-library-btn:hover { background: #334155; }
</style>

