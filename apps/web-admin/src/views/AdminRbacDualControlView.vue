<template>
  <div class="bn-admin-page">
    <header class="bn-page-header">
      <h2>민감 작업 2인 이중 승인 대기열 (Dual-Control Approval Queue)</h2>
      <p class="bn-text-muted">대규모 개발자 정산, 앱 영구 차단, 관리자 권한 부여 등 고위험 작업에 대해 Four-Eyes Principle을 강제합니다.</p>
    </header>

    <div class="bn-form-card">
      <table class="bn-table">
        <thead>
          <tr>
            <th>요청 ID</th>
            <th>작업 유형</th>
            <th>발의 관리자</th>
            <th>사유</th>
            <th>등록일시</th>
            <th>이중 승인 조작</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="req in pendingRequests" :key="req.id">
            <td><code>{{ req.id.substring(0, 8) }}...</code></td>
            <td><span class="bn-op-tag">{{ req.operationType }}</span></td>
            <td>
              <strong>{{ req.initiatorEmail }}</strong>
              <span v-if="req.initiatorId === currentAdminId" class="bn-self-tag">(본인)</span>
            </td>
            <td>{{ req.reason }}</td>
            <td class="bn-text-muted">{{ req.createdAt }}</td>
            <td>
              <!-- SELF-APPROVAL IS DISABLED AND BLOCKED (TC-SEC-006) -->
              <div v-if="req.initiatorId === currentAdminId" class="bn-self-approval-blocked">
                <button class="bn-btn bn-btn-disabled" disabled title="본인이 발의한 요청은 승인할 수 없습니다 (제11조 직무분리 규정)">
                  승인 불가 (본인 발의)
                </button>
              </div>

              <div v-else class="bn-approval-actions">
                <button class="bn-btn bn-btn-approve" @click="openApproveModal(req)">승인</button>
                <button class="bn-btn bn-btn-reject" @click="rejectRequest(req.id)">반려</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Step-up MFA Confirmation Modal -->
    <div v-if="activeModalRequest" class="bn-modal-backdrop">
      <div class="bn-modal">
        <h3>이중 승인 최종 확인 및 MFA 인증</h3>
        <p>작업: <strong>{{ activeModalRequest.operationType }}</strong></p>
        <p class="bn-text-muted">보안 규정에 따라 TOTP 2단계 인증 코드를 입력해야 최종 승인됩니다.</p>
        <div class="bn-form-group">
          <label>TOTP 인증 코드 (6자리)</label>
          <input v-model="totpCode" type="text" maxlength="6" placeholder="000000" class="bn-input" />
        </div>
        <div class="bn-modal-actions">
          <button class="bn-btn bn-btn-primary" :disabled="totpCode.length !== 6" @click="confirmApproval">
            최종 승인 실행
          </button>
          <button class="bn-btn bn-btn-secondary" @click="activeModalRequest = null">취소</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const currentAdminId = ref('admin-current-user-uuid')
const totpCode = ref('')
const activeModalRequest = ref<any>(null)

const pendingRequests = ref([
  {
    id: 'req-001-settle',
    operationType: 'SETTLEMENT_PAYOUT_OVER_LIMIT',
    initiatorId: 'admin-current-user-uuid', // Current user initiated!
    initiatorEmail: 'me@admin.bookingnetwork.internal',
    reason: '월간 대규모 개발자 수익 배분 지급 ($45,000)',
    createdAt: '2026-10-31 09:30'
  },
  {
    id: 'req-002-ban',
    operationType: 'PERMANENT_APP_BAN',
    initiatorId: 'admin-other-uuid', // Distinct admin initiated
    initiatorEmail: 'security-lead@admin.bookingnetwork.internal',
    reason: 'C2 원격 명령 실행 백도어 적발로 인한 영구 차단',
    createdAt: '2026-10-31 10:15'
  }
])

function openApproveModal(req: any) {
  activeModalRequest.value = req
  totpCode.value = ''
}

function confirmApproval() {
  if (activeModalRequest.value) {
    pendingRequests.value = pendingRequests.value.filter(r => r.id !== activeModalRequest.value.id)
    activeModalRequest.value = null
    alert('이중 승인이 정상적으로 집행되었습니다.')
  }
}

function rejectRequest(id: string) {
  pendingRequests.value = pendingRequests.value.filter(r => r.id !== id)
}
</script>

<style scoped>
.bn-admin-page {
  padding: 24px;
}
.bn-form-card {
  background-color: var(--bn-bg-surface, #FFFFFF);
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  border-radius: 12px;
  padding: 24px;
  margin-top: 20px;
}
.bn-table {
  width: 100%;
  border-collapse: collapse;
}
.bn-table th, .bn-table td {
  padding: 14px 16px;
  text-align: left;
  border-bottom: 1px solid var(--bn-border-subtle, #E2E8F0);
  font-size: 14px;
}
.bn-op-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 700;
  background-color: #FEF3C7;
  color: #92400E;
}
.bn-self-tag {
  font-size: 11px;
  color: var(--bn-text-muted, #64748B);
  margin-left: 4px;
}
.bn-btn-approve {
  background-color: var(--bn-trust-verified, #059669);
  color: white;
  border: none;
  padding: 6px 12px;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
}
.bn-btn-reject {
  background-color: var(--bn-danger-red, #DC2626);
  color: white;
  border: none;
  padding: 6px 12px;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
}
.bn-btn-disabled {
  background-color: var(--bn-border-subtle, #E2E8F0);
  color: var(--bn-text-muted, #64748B);
  border: none;
  padding: 6px 12px;
  border-radius: 6px;
  cursor: not-allowed;
  font-size: 12px;
}
.bn-approval-actions {
  display: flex;
  gap: 8px;
}
.bn-modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
}
.bn-modal {
  background: var(--bn-bg-surface, #FFFFFF);
  padding: 24px;
  border-radius: 12px;
  width: 440px;
}
.bn-modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}
</style>

