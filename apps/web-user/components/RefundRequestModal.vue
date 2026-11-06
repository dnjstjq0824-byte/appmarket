<template>
  <div v-if="isOpen" class="refund-backdrop">
    <div class="refund-modal">
      <div class="refund-head">
        <div>
          <h3 class="refund-title">주문 환불 및 청약철회 신청 (USR-15)</h3>
          <p class="refund-sub">전자상거래법 제17조에 따른 14일 이내 청약철회 절차</p>
        </div>
        <button @click="close" class="refund-x">✕</button>
      </div>

      <!-- App Info Card -->
      <div class="refund-info">
        <div class="refund-app">{{ appTitle }}</div>
        <div class="refund-dim">주문 ID: {{ orderId }}</div>
        <div class="refund-amount">결제 금액: ₩{{ amountWon.toLocaleString() }}</div>
        <div class="refund-date">구매일: {{ purchasedAt }} (14일 청약철회 가능 기간 내)</div>
      </div>

      <!-- Revocation Warning -->
      <div class="refund-warning">
        <div class="refund-warning-title">
          <span>⚠️</span> 이용 권한 즉시 회수 안내
        </div>
        환불이 승인되면 계정의 앱 다운로드 및 실행 권한(Entitlement)이 즉시 회수되며, 기설치된 앱의 업데이트 및 인앱 연동이 중단됩니다.
      </div>

      <!-- Reason Selection -->
      <div class="refund-reason">
        <label class="refund-label">환불 사유 선택</label>
        <select v-model="selectedReason" class="refund-select">
          <option value="SIMPLE_CHANGE_OF_MIND">단순 변심 또는 중복 구매</option>
          <option value="DEVICE_INCOMPATIBLE">보유 기기와의 호환성 결함 및 실행 불가</option>
          <option value="MISLEADING_DESCRIPTION">앱 설명 및 스크린샷과 실제 기능 불일치</option>
          <option value="UNAUTHORIZED_CHILD">미성년자 또는 제3자 무단 결제</option>
        </select>
      </div>

      <div v-if="submitError" class="refund-error">{{ submitError }}</div>

      <!-- Action Buttons -->
      <div class="refund-actions">
        <button @click="close" class="refund-cancel">
          취소
        </button>
        <button
          @click="submitRefund"
          :disabled="isSubmitting"
          class="refund-submit"
        >
          {{ isSubmitting ? '환불 처리 중...' : '환불 신청 완료' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { getSharedClient, ApiClientError } from '@bookingnetwork/api-client';

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

const api = getSharedClient();
const selectedReason = ref('SIMPLE_CHANGE_OF_MIND');
const isSubmitting = ref(false);
const submitError = ref('');

function close() {
  emit('close');
}

async function submitRefund() {
  isSubmitting.value = true;
  submitError.value = '';
  try {
    await api.refundOrder(props.orderId, selectedReason.value);
    emit('refund-completed', props.orderId);
  } catch (err) {
    submitError.value = err instanceof ApiClientError ? err.message : '환불 신청에 실패했습니다.';
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<style scoped>
.refund-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  z-index: 50;
}
.refund-modal {
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 12px;
  padding: 24px;
  max-width: 448px;
  width: 100%;
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.5);
  color: #f1f5f9;
}
.refund-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}
.refund-title {
  font-size: 16px;
  font-weight: 700;
  color: #f1f5f9;
  margin: 0;
}
.refund-sub {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 2px;
}
.refund-x {
  background: none;
  border: none;
  color: #94a3b8;
  font-size: 14px;
  cursor: pointer;
}
.refund-x:hover {
  color: #e2e8f0;
}
.refund-info {
  background: #020617;
  border: 1px solid #1e293b;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 16px;
  font-family: monospace;
  font-size: 12px;
}
.refund-info > div + div {
  margin-top: 4px;
}
.refund-app {
  color: #e2e8f0;
  font-weight: 700;
}
.refund-dim {
  color: #94a3b8;
}
.refund-amount {
  color: #38bdf8;
}
.refund-date {
  font-size: 11px;
  color: #34d399;
}
.refund-warning {
  background: rgba(76, 5, 25, 0.5);
  border: 1px solid rgba(244, 63, 94, 0.4);
  padding: 12px;
  border-radius: 8px;
  margin-bottom: 16px;
  font-size: 12px;
  color: #fda4af;
}
.refund-warning-title {
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 4px;
}
.refund-reason {
  margin-bottom: 20px;
}
.refund-label {
  display: block;
  font-size: 12px;
  font-weight: 600;
  color: #cbd5e1;
  margin-bottom: 6px;
}
.refund-select {
  width: 100%;
  background: #020617;
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 10px;
  font-size: 12px;
  color: #e2e8f0;
}
.refund-select:focus {
  outline: none;
  border-color: #0ea5e9;
}
.refund-error {
  font-size: 11px;
  color: #fb7185;
  margin-bottom: 12px;
}
.refund-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
.refund-cancel {
  padding: 8px 14px;
  background: #1e293b;
  color: #cbd5e1;
  border: none;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s;
}
.refund-cancel:hover {
  background: #334155;
}
.refund-submit {
  padding: 8px 14px;
  background: #e11d48;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}
.refund-submit:hover {
  background: #f43f5e;
}
.refund-submit:disabled {
  background: #1e293b;
  cursor: not-allowed;
}
</style>
