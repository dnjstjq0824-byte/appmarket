<template>
  <div class="kyb-review-queue">
    <header class="view-header">
      <div>
        <h2>조직 KYB 심사 큐</h2>
        <p class="subtitle">법인 신원 검증 제출 건 승인·반려 (admin:kyb 권한 필요)</p>
      </div>
      <select v-model="statusFilter" class="status-select" @change="loadQueue">
        <option value="PENDING_REVIEW">심사 대기</option>
        <option value="APPROVED">승인됨</option>
        <option value="REJECTED">반려됨</option>
      </select>
    </header>

    <div v-if="isLoading" class="loading-row">불러오는 중…</div>
    <div v-if="loadError" class="load-error">{{ loadError }}</div>
    <div v-if="actionError" class="load-error">{{ actionError }}</div>

    <template v-if="!isLoading && !loadError">
      <div v-if="queue.length === 0" class="empty-row">해당 상태의 심사 건이 없습니다.</div>
      <div v-else class="queue-list">
        <article v-for="item in queue" :key="item.id" class="queue-card">
          <div class="queue-main">
            <div class="queue-id mono">{{ item.org_id }}</div>
            <div class="queue-meta">
              <span><strong>문서:</strong> {{ item.doc_type }}</span>
              <span class="mono">{{ item.doc_url }}</span>
              <span><strong>도메인:</strong> {{ item.domain }}</span>
              <span class="muted">제출 {{ item.submitted_at }}</span>
            </div>
            <div v-if="item.reviewer_id" class="queue-meta muted">
              <span>심사자 {{ item.reviewer_id }}</span>
              <span>{{ item.review_notes }}</span>
              <span>{{ item.reviewed_at }}</span>
            </div>
          </div>
          <div v-if="item.status === 'PENDING_REVIEW'" class="queue-actions">
            <input v-model="notes[item.id]" placeholder="심사 메모 (반려 시 권장)" class="notes-input" />
            <div class="btn-row">
              <button class="btn btn-success" :disabled="isActing" @click="decide(item, true)">승인</button>
              <button class="btn btn-danger" :disabled="isActing" @click="decide(item, false)">반려</button>
            </div>
          </div>
          <span v-else class="badge" :class="item.status === 'APPROVED' ? 'badge-success' : 'badge-danger'">{{ item.status }}</span>
        </article>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getSharedClient, ApiClientError } from '@bookingnetwork/api-client';

const api = getSharedClient();

interface VerificationItem {
  id: string;
  org_id: string;
  doc_type: string;
  doc_url: string;
  domain: string;
  status: string;
  submitted_at: string;
  reviewer_id?: string;
  review_notes?: string;
  reviewed_at?: string;
}

const queue = ref<VerificationItem[]>([]);
const notes = ref<Record<string, string>>({});
const statusFilter = ref('PENDING_REVIEW');
const isLoading = ref(true);
const loadError = ref('');
const actionError = ref('');
const isActing = ref(false);

onMounted(async () => {
  await loadQueue();
  isLoading.value = false;
});

async function loadQueue() {
  isLoading.value = true;
  loadError.value = '';
  try {
    queue.value = await api.listVerifications(statusFilter.value) as VerificationItem[];
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load verification queue';
  } finally {
    isLoading.value = false;
  }
}

async function decide(item: VerificationItem, approved: boolean) {
  const note = (notes.value[item.id] || '').trim();
  if (!approved && !note) {
    actionError.value = '반려 시 심사 메모가 필요합니다.';
    return;
  }
  isActing.value = true;
  actionError.value = '';
  try {
    await api.decideVerification(item.org_id, approved, note);
    await loadQueue();
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to record decision';
  } finally {
    isActing.value = false;
  }
}
</script>

<style scoped>
.kyb-review-queue {
  padding: 24px;
  max-width: 1000px;
  margin: 0 auto;
}
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 24px;
}
.view-header h2 {
  margin: 0 0 4px;
}
.subtitle {
  color: #64748b;
  font-size: 14px;
  margin: 0;
}
.status-select {
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
  background: white;
}
.loading-row {
  padding: 16px;
  color: #64748b;
  font-size: 14px;
}
.load-error {
  padding: 12px 16px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #991b1b;
  border-radius: 6px;
  font-size: 13px;
  margin-bottom: 16px;
}
.empty-row {
  padding: 24px;
  color: #64748b;
  font-size: 14px;
  text-align: center;
}
.queue-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.queue-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 20px 24px;
  display: flex;
  justify-content: space-between;
  gap: 16px;
}
.queue-main {
  flex: 1;
}
.queue-id {
  font-size: 12px;
  font-weight: 700;
  color: #1d4ed8;
  margin-bottom: 8px;
}
.queue-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 13px;
  margin-top: 4px;
}
.mono {
  font-family: ui-monospace, monospace;
  font-size: 12px;
  word-break: break-all;
}
.muted {
  color: #64748b;
}
.queue-actions {
  min-width: 260px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.notes-input {
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
}
.btn-row {
  display: flex;
  gap: 8px;
}
.btn {
  flex: 1;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  border: none;
}
.btn-success { background: #16a34a; color: white; }
.btn-danger { background: #dc2626; color: white; }
.btn:disabled { opacity: 0.5; cursor: not-allowed; }
.badge {
  align-self: flex-start;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  white-space: nowrap;
}
.badge-success { background: #dcfce7; color: #166534; }
.badge-danger { background: #fee2e2; color: #991b1b; }
</style>
