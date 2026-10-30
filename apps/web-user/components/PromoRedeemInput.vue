<template>
  <div class="promo-redeem-input bg-slate-900 border border-slate-800 rounded-lg p-3">
    <div class="text-xs font-semibold text-slate-300 mb-2">프로모션 / 리딤 코드</div>
    
    <div v-if="!appliedCode" class="flex gap-2">
      <input 
        v-model="inputCode" 
        placeholder="쿠폰 코드 입력 (예: FREE2026)" 
        class="flex-1 bg-slate-950 border border-slate-700 rounded px-3 py-1.5 text-xs text-slate-100 font-mono focus:border-sky-500 focus:outline-none uppercase"
        @keyup.enter="applyPromo"
      />
      <button 
        @click="applyPromo" 
        :disabled="!inputCode.trim() || isChecking"
        class="px-3 py-1.5 bg-sky-600 hover:bg-sky-500 disabled:bg-slate-800 disabled:text-slate-600 text-white font-medium text-xs rounded transition"
      >
        {{ isChecking ? '확인 중...' : '적용' }}
      </button>
    </div>

    <!-- Applied Promo Banner -->
    <div v-else class="flex items-center justify-between bg-emerald-950/60 border border-emerald-500/40 p-2.5 rounded text-xs">
      <div class="flex items-center gap-2">
        <span class="text-emerald-400 font-bold">✓</span>
        <div>
          <span class="font-mono font-bold text-emerald-300">{{ appliedCode }}</span>
          <span class="text-emerald-400 ml-1">({{ discountLabel }})</span>
        </div>
      </div>
      <button @click="removePromo" class="text-slate-400 hover:text-rose-400 text-xs transition">
        삭제
      </button>
    </div>

    <div v-if="errorMessage" class="text-[11px] text-rose-400 mt-1.5">
      {{ errorMessage }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const emit = defineEmits<{
  (e: 'applied', promo: { code: string; type: 'FREE' | 'PERCENT'; percent?: number }): void;
  (e: 'removed'): void;
}>();

const inputCode = ref('');
const appliedCode = ref('');
const discountLabel = ref('');
const isChecking = ref(false);
const errorMessage = ref('');

function applyPromo() {
  const code = inputCode.value.trim().toUpperCase();
  if (!code) return;

  isChecking.value = true;
  errorMessage.value = '';

  setTimeout(() => {
    isChecking.value = false;
    if (code === 'FREE2026' || code === 'LAUNCH100') {
      appliedCode.value = code;
      discountLabel.value = '100% 무료 리딤 적용';
      emit('applied', { code, type: 'FREE', percent: 100 });
    } else if (code === 'HALF50') {
      appliedCode.value = code;
      discountLabel.value = '50% 할인 쿠폰';
      emit('applied', { code, type: 'PERCENT', percent: 50 });
    } else {
      errorMessage.value = '유효하지 않거나 이미 만료된 프로모션 코드입니다.';
    }
  }, 200);
}

function removePromo() {
  appliedCode.value = '';
  discountLabel.value = '';
  inputCode.value = '';
  errorMessage.value = '';
  emit('removed');
}
</script>

