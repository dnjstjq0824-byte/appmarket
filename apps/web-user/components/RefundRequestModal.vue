<template>
  <div v-if="isOpen" class="fixed inset-0 bg-black/75 flex items-center justify-center p-4 z-50">
    <div class="bg-slate-900 border border-slate-700 rounded-xl p-6 max-w-md w-full shadow-2xl text-slate-100">
      <div class="flex justify-between items-start mb-4">
        <div>
          <h3 class="text-base font-bold text-slate-100">주문 환불 및 청약철회 신청 (USR-15)</h3>
          <p class="text-xs text-slate-400 mt-0.5">전자상거래법 제17조에 따른 14일 이내 청약철회 절차</p>
        </div>
        <button @click="close" class="text-slate-400 hover:text-slate-200 text-sm">✕</button>
      </div>

      <!-- App Info Card -->
      <div class="bg-slate-950 border border-slate-800 rounded-lg p-3 mb-4 space-y-1 font-mono text-xs">
        <div class="text-slate-300 font-bold">{{ appTitle }}</div>
        <div class="text-slate-400">주문 ID: {{ orderId }}</div>
        <div class="text-sky-400">결제 금액: ₩{{ amountWon.toLocaleString() }}</div>
        <div class="text-[11px] text-emerald-400">구매일: {{ purchasedAt }} (14일 청약철회 가능 기간 내)</div>
      </div>

      <!-- Revocation Warning -->
      <div class="bg-rose-950/50 border border-rose-500/40 p-3 rounded-lg mb-4 text-xs text-rose-300">
        <div class="font-bold flex items-center gap-1 mb-1">
          <span>⚠️</span> 이용 권한 즉시 회수 안내
        </div>
        환불이 승인되면 계정의 앱 다운로드 및 실행 권한(Entitlement)이 즉시 회수되며, 기설치된 앱의 업데이트 및 인앱 연동이 중단됩니다.
      </div>

      <!-- Reason Selection -->
      <div class="mb-5">
        <label class="block text-xs font-semibold text-slate-300 mb-1.5">환불 사유 선택</label>
        <select v-model="selectedReason" class="w-full bg-slate-950 border border-slate-700 rounded-lg p-2.5 text-xs text-slate-200 focus:outline-none focus:border-sky-500">
          <option value="SIMPLE_CHANGE_OF_MIND">단순 변심 또는 중복 구매</option>
          <option value="DEVICE_INCOMPATIBLE">보유 기기와의 호환성 결함 및 실행 불가</option>
          <option value="MISLEADING_DESCRIPTION">앱 설명 및 스크린샷과 실제 기능 불일치</option>
          <option value="UNAUTHORIZED_CHILD">미성년자 또는 제3자 무단 결제</option>
        </select>
      </div>

      <!-- Action Buttons -->
      <div class="flex justify-end gap-2">
        <button @click="close" class="px-3.5 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-lg text-xs font-medium transition">
          취소
        </button>
        <button 
          @click="submitRefund" 
          :disabled="isSubmitting"
          class="px-3.5 py-2 bg-rose-600 hover:bg-rose-500 disabled:bg-slate-800 text-white rounded-lg text-xs font-semibold transition"
        >
          {{ isSubmitting ? '환불 처리 중...' : '환불 신청 완료' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const props = defineProps<{
  isOpen: boolean;
  orderId: string;
  appTitle: string;
  amountWon: number;
  purchasedAt: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'refund-completed', orderId: string): void;
}>();

const selectedReason = ref('SIMPLE_CHANGE_OF_MIND');
const isSubmitting = ref(false);

function close() {
  emit('close');
}

function submitRefund() {
  isSubmitting.value = true;
  setTimeout(() => {
    isSubmitting.value = false;
    emit('refund-completed', props.orderId);
  }, 400);
}
</script>

