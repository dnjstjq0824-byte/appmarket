<template>
  <div class="checkout-page min-h-screen bg-slate-950 text-slate-100 p-6 md:p-12">
    <div class="max-w-2xl mx-auto">
      <!-- Breadcrumb & Title -->
      <div class="mb-6">
        <a href="/" class="text-xs text-sky-400 hover:underline">← 마켓 둘러보기로 돌아가기</a>
        <h1 class="text-2xl font-bold tracking-tight text-slate-100 mt-2">유료 앱 주문 및 안전 결제 (USR-13)</h1>
        <p class="text-xs text-slate-400 mt-1">구매 즉시 다운로드 권한이 부여되며, 영구 라이선스로 관리됩니다.</p>
      </div>

      <!-- App Card Summary -->
      <div class="bg-slate-900 border border-slate-800 rounded-xl p-5 mb-6 flex gap-4 items-center">
        <div class="w-16 h-16 bg-gradient-to-br from-indigo-500 to-sky-600 rounded-xl flex items-center justify-center text-2xl font-bold shadow-lg">
          {{ appInfo.iconLetter }}
        </div>
        <div class="flex-1">
          <div class="text-xs text-sky-400 font-medium">{{ appInfo.developerName }}</div>
          <h2 class="text-lg font-bold text-slate-100">{{ appInfo.title }}</h2>
          <div class="text-xs text-slate-400 mt-0.5 font-mono">패키지: {{ appInfo.packageName }} • v{{ appInfo.version }}</div>
        </div>
      </div>

      <!-- Order Quote & Price Breakdown -->
      <div class="bg-slate-900 border border-slate-800 rounded-xl p-6 mb-6">
        <h3 class="text-sm font-semibold text-slate-200 mb-4">결제 금액 산출서 (Order Quote)</h3>

        <div class="space-y-3 font-mono text-xs">
          <div class="flex justify-between text-slate-300">
            <span>앱 정가 (Gross List Price)</span>
            <span>₩{{ basePriceWon.toLocaleString() }}</span>
          </div>

          <div v-if="discountWon > 0" class="flex justify-between text-emerald-400">
            <span>프로모션 할인 적용 ({{ appliedPromoCode }})</span>
            <span>- ₩{{ discountWon.toLocaleString() }}</span>
          </div>

          <div class="flex justify-between text-slate-400 pt-2 border-t border-slate-800">
            <span>공급가액 (Net Amount)</span>
            <span>₩{{ netAmountWon.toLocaleString() }}</span>
          </div>

          <div class="flex justify-between text-slate-400">
            <span>부가가치세 10% (VAT)</span>
            <span>₩{{ vatAmountWon.toLocaleString() }}</span>
          </div>

          <div class="flex justify-between text-base font-bold text-slate-100 pt-3 border-t border-slate-700">
            <span>최종 결제 금액 (Payable)</span>
            <span class="text-sky-400">₩{{ finalPayableWon.toLocaleString() }}</span>
          </div>
        </div>

        <!-- Promo Code Input Component -->
        <div class="mt-6">
          <PromoRedeemInput @applied="onPromoApplied" @removed="onPromoRemoved" />
        </div>
      </div>

      <!-- Payment Method Selection -->
      <div class="bg-slate-900 border border-slate-800 rounded-xl p-6 mb-6">
        <h3 class="text-sm font-semibold text-slate-200 mb-4">결제 수단 선택</h3>
        
        <div class="grid grid-cols-2 md:grid-cols-4 gap-3 text-xs">
          <button 
            v-for="pm in paymentMethods" 
            :key="pm.id"
            @click="selectedMethod = pm.id"
            :class="['p-3 rounded-lg border text-center transition font-medium', selectedMethod === pm.id ? 'bg-sky-950 border-sky-500 text-sky-300' : 'bg-slate-950 border-slate-800 text-slate-400 hover:border-slate-700']"
          >
            {{ pm.name }}
          </button>
        </div>

        <div class="mt-4 text-[11px] text-slate-500">
          안내: PG사 웹훅 HMAC-SHA256 서명 검증을 통해 100% 무결성이 보장되며 멱등 처리됩니다.
        </div>
      </div>

      <!-- Checkout CTA -->
      <div class="space-y-3">
        <button 
          @click="executePayment"
          :disabled="isProcessing"
          class="w-full py-3.5 bg-gradient-to-r from-sky-600 to-indigo-600 hover:from-sky-500 hover:to-indigo-500 disabled:from-slate-800 disabled:to-slate-800 text-white font-bold rounded-xl shadow-lg transition text-sm flex items-center justify-center gap-2"
        >
          <span v-if="isProcessing">결제 승인 처리 중...</span>
          <span v-else>
            {{ finalPayableWon === 0 ? '무료 리딤 완료 및 다운로드' : `₩${finalPayableWon.toLocaleString()} 안전 결제하기` }}
          </span>
        </button>

        <p class="text-center text-[11px] text-slate-500">
          구매 후 14일 이내 미다운로드 시 전액 환불(청약철회) 가능합니다.
        </p>
      </div>

      <!-- Purchase Success Dialog -->
      <div v-if="purchaseSuccess" class="fixed inset-0 bg-black/80 flex items-center justify-center p-4 z-50">
        <div class="bg-slate-900 border border-slate-700 rounded-xl p-6 max-w-md w-full text-center">
          <div class="w-12 h-12 bg-emerald-950 text-emerald-400 rounded-full flex items-center justify-center mx-auto mb-4 text-2xl border border-emerald-500/40">
            ✓
          </div>
          <h3 class="text-lg font-bold text-slate-100 mb-1">결제 완료 및 소유 권한 발급</h3>
          <p class="text-xs text-slate-400 mb-6">
            [{{ appInfo.title }}]의 정품 라이선스가 발급되었습니다. 지금 바로 고속 CDN을 통해 안전하게 다운로드하세요.
          </p>

          <div class="bg-slate-950 p-3 rounded-lg text-xs font-mono text-left mb-6 space-y-1">
            <div class="text-slate-400">주문 번호: <span class="text-sky-400">{{ orderId }}</span></div>
            <div class="text-slate-400">결제 금액: ₩{{ finalPayableWon.toLocaleString() }}</div>
            <div class="text-slate-400">권한 상태: <span class="text-emerald-400 font-bold">ACTIVE (인증됨)</span></div>
          </div>

          <div class="space-y-2">
            <a 
              :href="downloadUrl" 
              class="block w-full py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white font-bold text-xs rounded-lg transition"
            >
              APK 다운로드 시작 (15분 유효 서명 토큰)
            </a>
            <a 
              href="/account/library" 
              class="block w-full py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs rounded-lg transition"
            >
              내 보관함으로 이동
            </a>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import PromoRedeemInput from '../../components/PromoRedeemInput.vue';

const appInfo = ref({
  title: 'Vector Pro Studio for Android',
  developerName: 'Creative Design Labs',
  packageName: 'com.creativedesign.vectorpro',
  version: '2.4.0',
  iconLetter: 'V'
});

const basePriceWon = ref(15000);
const discountWon = ref(0);
const appliedPromoCode = ref('');
const selectedMethod = ref('TOSS_PAY');
const isProcessing = ref(false);
const purchaseSuccess = ref(false);
const orderId = ref('');
const downloadUrl = ref('');

const paymentMethods = [
  { id: 'TOSS_PAY', name: '토스페이' },
  { id: 'CREDIT_CARD', name: '신용/체크카드' },
  { id: 'NAVER_PAY', name: '네이버페이' },
  { id: 'KAKAO_PAY', name: '카카오페이' }
];

const finalPayableWon = computed(() => Math.max(0, basePriceWon.value - discountWon.value));
const netAmountWon = computed(() => Math.floor((finalPayableWon.value * 10000) / 11000));
const vatAmountWon = computed(() => finalPayableWon.value - netAmountWon.value);

function onPromoApplied(promo: { code: string; type: 'FREE' | 'PERCENT'; percent?: number }) {
  appliedPromoCode.value = promo.code;
  if (promo.type === 'FREE') {
    discountWon.value = basePriceWon.value;
  } else if (promo.type === 'PERCENT' && promo.percent) {
    discountWon.value = Math.floor((basePriceWon.value * promo.percent) / 100);
  }
}

function onPromoRemoved() {
  appliedPromoCode.value = '';
  discountWon.value = 0;
}

function executePayment() {
  isProcessing.value = true;
  setTimeout(() => {
    isProcessing.value = false;
    orderId.value = `ord-${Date.now()}`;
    downloadUrl.value = `/api/v1/distribution/download/tokenized?pkg=${appInfo.value.packageName}&token=hmac_mock_signed_token_15m`;
    purchaseSuccess.value = true;
  }, 500);
}
</script>

