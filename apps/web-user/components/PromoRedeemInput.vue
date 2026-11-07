<template>
  <div class="promo-redeem-input">
    <div class="promo-title">프로모션 / 리딤 코드</div>

    <div v-if="!appliedCode" class="promo-row">
      <input
        v-model="inputCode"
        placeholder="쿠폰 코드 입력 (예: FREE2026)"
        class="promo-input"
        @keyup.enter="applyPromo"
      />
      <button
        @click="applyPromo"
        :disabled="!inputCode.trim() || isChecking"
        class="promo-apply-btn"
      >
        {{ isChecking ? '확인 중...' : '적용' }}
      </button>
    </div>

    <!-- Applied Promo Banner -->
    <div v-else class="promo-applied">
      <div class="promo-applied-left">
        <span class="promo-check">✓</span>
        <div>
          <span class="promo-code">{{ appliedCode }}</span>
          <span class="promo-label">({{ discountLabel }})</span>
        </div>
      </div>
      <button @click="removePromo" class="promo-remove">
        삭제
      </button>
    </div>

    <div v-if="errorMessage" class="promo-error">
      {{ errorMessage }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { getSharedClient, ApiClientError } from '@bookingnetwork/api-client';

const props = defineProps<{ appId: string }>();

const emit = defineEmits<{
  (e: 'applied', promo: { code: string; type: 'FREE' | 'PERCENT' | 'FIXED'; percent?: number; fixedMinor?: number }): void;
  (e: 'removed'): void;
}>();

const api = getSharedClient();
const inputCode = ref('');
const appliedCode = ref('');
const discountLabel = ref('');
const isChecking = ref(false);
const errorMessage = ref('');

async function applyPromo() {
  const code = inputCode.value.trim().toUpperCase();
  if (!code || !props.appId) return;

  isChecking.value = true;
  errorMessage.value = '';

  try {
    const res = await api.validatePromo(code, props.appId);
    appliedCode.value = res.code;
    const promo = res.promotion;
    if (promo.discount_type === 'FREE_REDEEM') {
      discountLabel.value = '100% 무료 리딤 적용';
      emit('applied', { code: res.code, type: 'FREE', percent: 100 });
    } else if (promo.discount_type === 'PERCENTAGE') {
      discountLabel.value = `${promo.discount_value}% 할인 쿠폰`;
      emit('applied', { code: res.code, type: 'PERCENT', percent: promo.discount_value });
    } else {
      discountLabel.value = `${promo.discount_value.toLocaleString()}원 정액 할인`;
      emit('applied', { code: res.code, type: 'FIXED', fixedMinor: promo.discount_value });
    }
  } catch (err) {
    errorMessage.value = err instanceof ApiClientError ? err.message : '유효하지 않거나 이미 만료된 프로모션 코드입니다.';
  } finally {
    isChecking.value = false;
  }
}

function removePromo() {
  appliedCode.value = '';
  discountLabel.value = '';
  inputCode.value = '';
  errorMessage.value = '';
  emit('removed');
}
</script>

<style scoped>
.promo-redeem-input {
  background: var(--bn-bg-well);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  padding: 14px;
}
.promo-title {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  margin-bottom: 10px;
}
.promo-row {
  display: flex;
  gap: 8px;
}
.promo-input {
  flex: 1;
  min-width: 0;
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  padding: 8px 12px;
  font-size: 13px;
  color: var(--bn-text-primary);
  font-family: var(--bn-font-mono);
  text-transform: uppercase;
}
.promo-input:focus {
  border-color: var(--bn-ink);
  outline: none;
}
.promo-apply-btn {
  padding: 8px 18px;
  background: var(--bn-action-bg);
  color: var(--bn-action-text);
  font-weight: 700;
  font-size: 13px;
  border: none;
  border-radius: var(--bn-radius-sm);
  cursor: pointer;
  white-space: nowrap;
}
.promo-apply-btn:hover:not(:disabled) {
  background: var(--bn-action-bg-hover);
}
.promo-apply-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.promo-applied {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  background: var(--bn-trust-wash);
  border: 1px solid var(--bn-trust-verified);
  padding: 10px 12px;
  border-radius: var(--bn-radius-sm);
  font-size: 12px;
}
.promo-applied-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.promo-check {
  color: var(--bn-trust-verified);
  font-weight: 800;
}
.promo-code {
  font-family: var(--bn-font-mono);
  font-weight: 700;
  color: var(--bn-text-primary);
}
.promo-label {
  color: var(--bn-trust-verified);
  margin-left: 4px;
}
.promo-remove {
  background: none;
  border: none;
  color: var(--bn-text-muted);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}
.promo-remove:hover {
  color: var(--bn-danger-red);
}
.promo-error {
  font-size: 12px;
  color: var(--bn-danger-red);
  margin-top: 8px;
}
</style>
