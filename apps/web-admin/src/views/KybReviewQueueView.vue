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
  max-width: 1020px;
  margin: 0 auto;
}
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 2px solid var(--bn-ink);
}
.view-header h2 {
  font-family: var(--bn-font-display);
  font-size: 24px;
  font-weight: 900;
  letter-spacing: -0.02em;
  margin: 0 0 4px;
}
.subtitle {
  color: var(--bn-text-secondary);
  font-size: 13px;
  margin: 0;
}
.status-select {
  padding: 8px 12px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  font-family: var(--bn-font-sans);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  flex-shrink: 0;
}
.loading-row {
  padding: 24px;
  color: var(--bn-text-muted);
  font-size: 14px;
  text-align: center;
}
.load-error {
  padding: 12px 16px;
  background: var(--bn-danger-wash);
  border: 1px solid var(--bn-danger-red);
  color: var(--bn-danger-red);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  margin-bottom: 16px;
}
.empty-row {
  padding: 28px 24px;
  color: var(--bn-text-muted);
  font-size: 14px;
  text-align: center;
  border: 1px dashed var(--bn-border-strong);
  border-radius: var(--bn-radius-md);
}
.queue-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.queue-card {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 20px 24px;
  display: flex;
  justify-content: space-between;
  gap: 20px;
}
.queue-main {
  flex: 1;
  min-width: 0;
}
.queue-id {
  font-family: var(--bn-font-mono);
  font-size: 12px;
  font-weight: 700;
  margin-bottom: 8px;
}
.queue-meta {
  display: flex;
  flex-direction: column;
  gap: 5px;
  font-size: 13px;
  margin-top: 4px;
  color: var(--bn-text-secondary);
}
.mono {
  font-family: var(--bn-font-mono);
  font-size: 12px;
  word-break: break-all;
}
.muted {
  color: var(--bn-text-muted);
}
.queue-meta.muted {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid var(--bn-border-subtle);
  font-size: 12px;
}
.queue-actions {
  min-width: 260px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.notes-input {
  padding: 9px 12px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  font-family: var(--bn-font-sans);
  font-size: 13px;
}
.notes-input:focus {
  outline: none;
  border-color: var(--bn-ink);
}
.btn-row {
  display: flex;
  gap: 8px;
}
.btn {
  flex: 1;
  padding: 9px 12px;
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}
.btn-success {
  background: transparent;
  border: 1px solid var(--bn-trust-verified);
  color: var(--bn-trust-verified);
}
.btn-success:hover:not(:disabled) {
  background: var(--bn-trust-wash);
}
.btn-danger {
  background: var(--bn-danger-red);
  border: none;
  color: #fff;
}
.btn-danger:hover:not(:disabled) {
  filter: brightness(0.92);
}
.btn:disabled { opacity: 0.45; cursor: not-allowed; }
.badge {
  align-self: flex-start;
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.04em;
  padding: 3px 9px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-strong);
  color: var(--bn-text-secondary);
  white-space: nowrap;
}
.badge-success { border-color: var(--bn-trust-verified); color: var(--bn-trust-verified); }
.badge-danger { border-color: var(--bn-danger-red); color: var(--bn-danger-red); }
@media (max-width: 720px) {
  .queue-card {
    flex-direction: column;
  }
  .queue-actions {
    min-width: 0;
  }
}
</style>
