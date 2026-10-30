<template>
  <div class="promo-code-manager p-6 bg-slate-900 text-slate-100 min-h-screen">
    <header class="mb-8">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">프로모션 및 리딤 코드 관리 (DEV-11)</h1>
          <p class="text-sm text-slate-400 mt-1">앱 할인 쿠폰 및 100% 무료 리딤 프로모션 코드 대량 발급 및 소진 현황</p>
        </div>
        <div class="flex gap-2">
          <button @click="openCreateModal" class="px-4 py-2 bg-sky-600 hover:bg-sky-500 text-white font-medium text-xs rounded transition shadow">
            + 프로모션 코드 일괄 생성
          </button>
        </div>
      </div>
    </header>

    <!-- Promo Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">활성 캠페인</span>
        <div class="text-2xl font-bold mt-1 text-slate-200">{{ promoBatches.length }}개</div>
      </div>
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">발급된 총 코드 수</span>
        <div class="text-2xl font-bold mt-1 text-sky-400">{{ totalCodesIssued }}개</div>
      </div>
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">리딤 사용 완료</span>
        <div class="text-2xl font-bold mt-1 text-emerald-400">{{ totalCodesRedeemed }}개</div>
      </div>
      <div class="bg-slate-800/80 border border-slate-700 p-4 rounded-lg">
        <span class="text-xs text-slate-400">소진율</span>
        <div class="text-2xl font-bold mt-1 text-indigo-400">{{ ((totalCodesRedeemed / totalCodesIssued) * 100).toFixed(1) }}%</div>
      </div>
    </div>

    <!-- Promo Batches Table -->
    <div class="bg-slate-800/90 border border-slate-700 rounded-lg overflow-hidden">
      <div class="p-4 border-b border-slate-700 flex justify-between items-center">
        <h2 class="text-base font-semibold text-slate-200">생성된 프로모션 코드 배치 목록</h2>
      </div>

      <table class="w-full text-left text-sm text-slate-300">
        <thead class="bg-slate-900/60 text-xs uppercase tracking-wider text-slate-400 border-b border-slate-700">
          <tr>
            <th class="p-3">배치 이름 / 생성일</th>
            <th class="p-3">혜택 구분</th>
            <th class="p-3">발급 수 / 사용 수</th>
            <th class="p-3">유효 기간</th>
            <th class="p-3">상태</th>
            <th class="p-3 text-right">관리</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-700/60">
          <tr v-for="b in promoBatches" :key="b.id" class="hover:bg-slate-750 transition">
            <td class="p-3 font-mono text-xs">
              <div class="font-bold text-sky-400">{{ b.name }}</div>
              <div class="text-slate-400">{{ b.createdAt }}</div>
            </td>
            <td class="p-3 text-xs">
              <span class="px-2 py-0.5 rounded text-[11px] font-mono font-bold bg-indigo-950 text-indigo-300 border border-indigo-500/30">
                {{ b.discountType === 'FREE' ? '100% 무료 리딤' : `${b.discountPercent}% 할인` }}
              </span>
            </td>
            <td class="p-3 text-xs font-mono">
              <div>{{ b.redeemedCount }} / {{ b.totalCount }}</div>
              <div class="w-24 bg-slate-700 h-1.5 rounded overflow-hidden mt-1">
                <div class="bg-emerald-500 h-full" :style="{ width: ((b.redeemedCount / b.totalCount) * 100) + '%' }"></div>
              </div>
            </td>
            <td class="p-3 text-xs font-mono text-slate-400">
              {{ b.expiresAt }}
            </td>
            <td class="p-3 text-xs">
              <span :class="b.isActive ? 'bg-emerald-950 text-emerald-400 border-emerald-500/30' : 'bg-slate-700 text-slate-400 border-slate-600'" class="px-2 py-0.5 rounded text-[11px] font-semibold border">
                {{ b.isActive ? '진행 중' : '만료/정지' }}
              </span>
            </td>
            <td class="p-3 text-right text-xs">
              <button @click="exportCsv(b)" class="px-2 py-1 bg-slate-700 hover:bg-slate-600 text-slate-200 rounded text-xs transition mr-1">
                CSV 다운로드
              </button>
              <button @click="toggleActive(b)" class="px-2 py-1 bg-amber-600 hover:bg-amber-500 text-white rounded text-xs transition">
                {{ b.isActive ? '비활성화' : '재활성화' }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create Promo Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-black/70 flex items-center justify-center p-4 z-50">
      <div class="bg-slate-800 border border-slate-700 rounded-lg p-6 max-w-md w-full">
        <h3 class="text-lg font-bold text-slate-100 mb-4">새 프로모션 코드 배치 생성</h3>
        <div class="space-y-4 text-xs">
          <div>
            <label class="block text-slate-400 mb-1">배치 캠페인 이름</label>
            <input v-model="newBatch.name" class="w-full bg-slate-900 border border-slate-700 rounded p-2 text-slate-200" placeholder="예: 2026 가을 출시 기념 100% 무료" />
          </div>
          <div>
            <label class="block text-slate-400 mb-1">혜택 구분</label>
            <select v-model="newBatch.discountType" class="w-full bg-slate-900 border border-slate-700 rounded p-2 text-slate-200">
              <option value="FREE">100% 무료 리딤 (소비자 결제액 0원)</option>
              <option value="PERCENT">정률 할인</option>
            </select>
          </div>
          <div v-if="newBatch.discountType === 'PERCENT'">
            <label class="block text-slate-400 mb-1">할인율 (%)</label>
            <input type="number" v-model.number="newBatch.discountPercent" min="1" max="99" class="w-full bg-slate-900 border border-slate-700 rounded p-2 text-slate-200 font-mono" />
          </div>
          <div>
            <label class="block text-slate-400 mb-1">생성할 코드 수량</label>
            <input type="number" v-model.number="newBatch.totalCount" min="1" max="1000" class="w-full bg-slate-900 border border-slate-700 rounded p-2 text-slate-200 font-mono" />
          </div>
        </div>
        <div class="flex justify-end gap-2 mt-6">
          <button @click="showModal = false" class="px-3 py-1.5 bg-slate-700 hover:bg-slate-600 text-slate-300 rounded text-xs">취소</button>
          <button @click="createBatch" class="px-3 py-1.5 bg-sky-600 hover:bg-sky-500 text-white rounded text-xs font-semibold">코드 일괄 생성</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

interface PromoBatch {
  id: string;
  name: string;
  createdAt: string;
  discountType: 'FREE' | 'PERCENT';
  discountPercent?: number;
  totalCount: number;
  redeemedCount: number;
  expiresAt: string;
  isActive: boolean;
}

const showModal = ref(false);

const promoBatches = ref<PromoBatch[]>([
  {
    id: 'batch-001',
    name: '베타 테스터 전용 100% 무료 리딤',
    createdAt: '2026-10-20',
    discountType: 'FREE',
    totalCount: 500,
    redeemedCount: 382,
    expiresAt: '2026-12-31',
    isActive: true
  },
  {
    id: 'batch-002',
    name: '블랙프라이데이 50% 할인 쿠폰',
    createdAt: '2026-10-25',
    discountType: 'PERCENT',
    discountPercent: 50,
    totalCount: 1000,
    redeemedCount: 215,
    expiresAt: '2026-11-30',
    isActive: true
  }
]);

const newBatch = ref({
  name: '',
  discountType: 'FREE' as 'FREE' | 'PERCENT',
  discountPercent: 30,
  totalCount: 100
});

const totalCodesIssued = computed(() => promoBatches.value.reduce((sum, b) => sum + b.totalCount, 0));
const totalCodesRedeemed = computed(() => promoBatches.value.reduce((sum, b) => sum + b.redeemedCount, 0));

function openCreateModal() {
  showModal.value = true;
}

function createBatch() {
  promoBatches.value.push({
    id: `batch-${Date.now()}`,
    name: newBatch.value.name || '새 프로모션 배치',
    createdAt: new Date().toISOString().split('T')[0],
    discountType: newBatch.value.discountType,
    discountPercent: newBatch.value.discountType === 'PERCENT' ? newBatch.value.discountPercent : undefined,
    totalCount: newBatch.value.totalCount,
    redeemedCount: 0,
    expiresAt: '2026-12-31',
    isActive: true
  });
  showModal.value = false;
}

function toggleActive(b: PromoBatch) {
  b.isActive = !b.isActive;
}

function exportCsv(b: PromoBatch) {
  alert(`배치 [${b.name}] 프로모션 코드 CSV 내보내기 완료 (${b.totalCount}건)`);
}
</script>

