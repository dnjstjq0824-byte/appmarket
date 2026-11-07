<template>
  <div v-if="isOpen" class="refund-backdrop">
    <div class="refund-modal">
      <div class="refund-head">
        <div>
          <h3 class="refund-title">주문 환불 및 청약철회 신청 (USR-15)</h3>
          <p class="refund-sub">전자상거래법 제17조에 따른 14일 이내 청약철회 절차</p>
        </div>
        <button @click="close" class="refund-x" aria-label="닫기"><svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"><path d="M1 1l10 10M11 1L1 11" /></svg></button>
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
          이용 권한 즉시 회수 안내
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
  background: rgba(20, 17, 12, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  z-index: 50;
}
.refund-modal {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-top: 3px solid var(--bn-danger-red);
  border-radius: var(--bn-radius-md);
  padding: 24px;
  max-width: 460px;
  width: 100%;
  box-shadow: var(--bn-shadow-pop);
}
.refund-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
}
.refund-title {
  font-family: var(--bn-font-display);
  font-size: 16px;
  font-weight: 800;
  margin: 0;
  line-height: 1.4;
}
.refund-sub {
  font-size: 12px;
  color: var(--bn-text-secondary);
  margin-top: 4px;
}
.refund-x {
  background: none;
  border: none;
  color: var(--bn-text-muted);
  cursor: pointer;
  display: flex;
  padding: 4px;
  flex-shrink: 0;
}
.refund-x:hover {
  color: var(--bn-text-primary);
}
.refund-info {
  background: var(--bn-bg-well);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  padding: 14px;
  margin-bottom: 16px;
  font-family: var(--bn-font-mono);
  font-size: 12px;
  color: var(--bn-text-secondary);
}
.refund-info > div + div {
  margin-top: 6px;
}
.refund-app {
  font-family: var(--bn-font-sans);
  font-size: 14px;
  color: var(--bn-text-primary);
  font-weight: 700;
}
.refund-dim {
  color: var(--bn-text-muted);
}
.refund-amount {
  color: var(--bn-text-primary);
  font-weight: 700;
}
.refund-date {
  font-size: 11px;
  color: var(--bn-trust-verified);
}
.refund-warning {
  background: var(--bn-danger-wash);
  border: 1px solid var(--bn-danger-red);
  padding: 12px 14px;
  border-radius: var(--bn-radius-sm);
  margin-bottom: 16px;
  font-size: 12px;
  line-height: 1.65;
  color: var(--bn-text-secondary);
}
.refund-warning-title {
  font-weight: 800;
  font-size: 13px;
  color: var(--bn-danger-red);
  margin-bottom: 4px;
}
.refund-reason {
  margin-bottom: 20px;
}
.refund-label {
  display: block;
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  margin-bottom: 8px;
}
.refund-select {
  width: 100%;
  box-sizing: border-box;
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  padding: 10px 12px;
  font-family: var(--bn-font-sans);
  font-size: 13px;
  color: var(--bn-text-primary);
}
.refund-select:focus {
  outline: none;
  border-color: var(--bn-ink);
}
.refund-error {
  font-size: 12px;
  color: var(--bn-danger-red);
  margin-bottom: 12px;
}
.refund-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
.refund-cancel {
  padding: 9px 18px;
  background: transparent;
  border: 1px solid var(--bn-border-subtle);
  color: var(--bn-text-secondary);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}
.refund-cancel:hover {
  border-color: var(--bn-border-strong);
  color: var(--bn-text-primary);
}
.refund-submit {
  padding: 9px 18px;
  background: var(--bn-danger-red);
  color: #fff;
  border: none;
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}
.refund-submit:hover:not(:disabled) {
  filter: brightness(0.92);
}
.refund-submit:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
</style>
