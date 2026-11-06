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
  background: #0f172a;
  border: 1px solid #1e293b;
  border-radius: 8px;
  padding: 12px;
}
.promo-title {
  font-size: 12px;
  font-weight: 600;
  color: #cbd5e1;
  margin-bottom: 8px;
}
.promo-row {
  display: flex;
  gap: 8px;
}
.promo-input {
  flex: 1;
  background: #020617;
  border: 1px solid #334155;
  border-radius: 4px;
  padding: 6px 12px;
  font-size: 12px;
  color: #f1f5f9;
  font-family: monospace;
  text-transform: uppercase;
}
.promo-input:focus {
  border-color: #0ea5e9;
  outline: none;
}
.promo-apply-btn {
  padding: 6px 12px;
  background: #0284c7;
  color: #fff;
  font-weight: 500;
  font-size: 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.15s;
}
.promo-apply-btn:hover {
  background: #0ea5e9;
}
.promo-apply-btn:disabled {
  background: #1e293b;
  color: #475569;
  cursor: not-allowed;
}
.promo-applied {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(6, 78, 59, 0.6);
  border: 1px solid rgba(16, 185, 129, 0.4);
  padding: 10px;
  border-radius: 4px;
  font-size: 12px;
}
.promo-applied-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.promo-check {
  color: #34d399;
  font-weight: 700;
}
.promo-code {
  font-family: monospace;
  font-weight: 700;
  color: #6ee7b7;
}
.promo-label {
  color: #34d399;
  margin-left: 4px;
}
.promo-remove {
  background: none;
  border: none;
  color: #94a3b8;
  font-size: 12px;
  cursor: pointer;
  transition: color 0.15s;
}
.promo-remove:hover {
  color: #fb7185;
}
.promo-error {
  font-size: 11px;
  color: #fb7185;
  margin-top: 6px;
}
</style>
