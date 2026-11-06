<template>
  <div class="bn-admin-page">
    <header class="bn-page-header">
      <h2>민감 작업 2인 이중 승인 대기열 (Dual-Control Approval Queue)</h2>
      <p class="bn-text-muted">대규모 개발자 정산, 앱 영구 차단, 관리자 권한 부여 등 고위험 작업에 대해 Four-Eyes Principle을 강제합니다.</p>
    </header>

    <div class="bn-form-card">
      <h3>새 승인 요청 발의</h3>
      <div class="bn-create-row">
        <select v-model="newOpType" class="bn-input">
          <option value="SETTLEMENT_PAYOUT_OVER_LIMIT">SETTLEMENT_PAYOUT_OVER_LIMIT</option>
          <option value="PERMANENT_APP_BAN">PERMANENT_APP_BAN</option>
          <option value="GRANT_ADMIN_ROLE">GRANT_ADMIN_ROLE</option>
          <option value="FORCED_REFUND">FORCED_REFUND</option>
        </select>
        <input v-model="newTarget" type="text" placeholder="대상 리소스 ID (선택)" class="bn-input" />
        <input v-model="newReason" type="text" placeholder="발의 사유 (필수)" class="bn-input bn-input-reason" />
        <button class="bn-btn bn-btn-primary" :disabled="createBusy" @click="createRequest">
          {{ createBusy ? '등록 중…' : '요청 발의' }}
        </button>
      </div>
      <div v-if="createError" class="bn-error">{{ createError }}</div>
    </div>

    <div v-if="isLoading" class="bn-loading">불러오는 중…</div>
    <div v-if="loadError" class="bn-error">{{ loadError }}</div>
    <div v-if="actionError" class="bn-error">{{ actionError }}</div>
    <div v-if="!isLoading && !loadError && pendingRequests.length === 0" class="bn-empty">대기 중인 이중 승인 요청이 없습니다.</div>

    <div v-if="!isLoading && !loadError" class="bn-form-card">
      <table class="bn-table">
        <thead>
          <tr>
            <th>요청 ID</th>
            <th>작업 유형</th>
            <th>발의 관리자</th>
            <th>대상 리소스</th>
            <th>사유</th>
            <th>등록일시</th>
            <th>이중 승인 조작</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="req in pendingRequests" :key="req.id">
            <td><code>{{ req.id.substring(0, 8) }}...</code></td>
            <td><span class="bn-op-tag">{{ req.operation_type }}</span></td>
            <td>
              <strong>{{ req.initiator_id.substring(0, 8) }}…</strong>
              <span v-if="req.initiator_id === currentAdminId" class="bn-self-tag">(본인)</span>
            </td>
            <td class="bn-text-muted">{{ req.target_resource_id || '-' }}</td>
            <td>{{ req.reason }}</td>
            <td class="bn-text-muted">{{ req.created_at }}</td>
            <td>
              <!-- SELF-APPROVAL IS DISABLED AND BLOCKED (TC-SEC-006) -->
              <div v-if="req.initiator_id === currentAdminId" class="bn-self-approval-blocked">
                <button class="bn-btn bn-btn-disabled" disabled title="본인이 발의한 요청은 승인할 수 없습니다 (제11조 직무분리 규정)">
                  승인 불가 (본인 발의)
                </button>
              </div>

              <div v-else class="bn-approval-actions">
                <button class="bn-btn bn-btn-approve" :disabled="actionBusy" @click="openApproveModal(req)">승인</button>
                <button class="bn-btn bn-btn-reject" :disabled="actionBusy" @click="rejectRequest(req.id)">반려</button>
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
        <p>작업: <strong>{{ activeModalRequest.operation_type }}</strong></p>
        <p class="bn-text-muted">보안 규정에 따라 TOTP 2단계 인증 코드를 입력해야 최종 승인됩니다.</p>
        <div class="bn-form-group">
          <label>TOTP 인증 코드 (6자리)</label>
          <input v-model="totpCode" type="text" maxlength="6" placeholder="000000" class="bn-input" />
        </div>
        <div v-if="actionError" class="bn-error">{{ actionError }}</div>
        <div class="bn-modal-actions">
          <button class="bn-btn bn-btn-primary" :disabled="totpCode.length !== 6 || actionBusy" @click="confirmApproval">
            {{ actionBusy ? '승인 중…' : '최종 승인 실행' }}
          </button>
          <button class="bn-btn bn-btn-secondary" @click="activeModalRequest = null">취소</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getSharedClient, storeSession, ApiClientError, type DualApprovalDTO } from '@bookingnetwork/api-client'

const api = getSharedClient()

const currentAdminId = ref('')
const totpCode = ref('')
const activeModalRequest = ref<DualApprovalDTO | null>(null)

const pendingRequests = ref<DualApprovalDTO[]>([])
const isLoading = ref(true)
const loadError = ref('')
const actionError = ref('')
const actionBusy = ref(false)

const newOpType = ref('SETTLEMENT_PAYOUT_OVER_LIMIT')
const newTarget = ref('')
const newReason = ref('')
const createBusy = ref(false)
const createError = ref('')

onMounted(async () => {
  try {
    const me = await api.getMe()
    currentAdminId.value = me.sub
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load current admin'
  }
  await loadPending()
  isLoading.value = false
})

async function loadPending() {
  loadError.value = ''
  try {
    pendingRequests.value = await api.listApprovals()
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load pending approvals'
  }
}

async function createRequest() {
  if (!newReason.value.trim()) {
    createError.value = '사유를 입력하십시오.'
    return
  }
  createBusy.value = true
  createError.value = ''
  try {
    const created = await api.createApproval({
      operation_type: newOpType.value,
      target_resource_id: newTarget.value.trim(),
      reason: newReason.value.trim()
    })
    pendingRequests.value = [created, ...pendingRequests.value]
    newTarget.value = ''
    newReason.value = ''
  } catch (err) {
    createError.value = err instanceof ApiClientError ? err.message : 'Failed to create approval request'
  } finally {
    createBusy.value = false
  }
}

function openApproveModal(req: DualApprovalDTO) {
  activeModalRequest.value = req
  totpCode.value = ''
  actionError.value = ''
}

async function confirmApproval() {
  if (!activeModalRequest.value) return
  actionBusy.value = true
  actionError.value = ''
  try {
    const mfa = await api.verifyMfa(currentAdminId.value, totpCode.value.trim())
    if (!mfa.access_token) {
      actionError.value = 'MFA 검증에 실패했습니다.'
      return
    }
    storeSession(mfa.access_token, mfa.refresh_token)
    await api.approveDualControl(activeModalRequest.value.id)
    pendingRequests.value = pendingRequests.value.filter(r => r.id !== activeModalRequest.value!.id)
    activeModalRequest.value = null
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Approval failed'
  } finally {
    actionBusy.value = false
  }
}

async function rejectRequest(id: string) {
  const reason = window.prompt('반려 사유를 입력하십시오:')
  if (reason === null) return
  actionBusy.value = true
  actionError.value = ''
  try {
    await api.rejectDualControl(id, reason)
    pendingRequests.value = pendingRequests.value.filter(r => r.id !== id)
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Rejection failed'
  } finally {
    actionBusy.value = false
  }
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
.bn-create-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 12px;
}
.bn-input-reason {
  flex: 1;
  min-width: 200px;
}
.bn-loading, .bn-empty {
  padding: 24px;
  text-align: center;
  color: var(--bn-text-muted, #64748B);
}
.bn-error {
  padding: 12px 16px;
  margin-top: 12px;
  background: #FEF2F2;
  border: 1px solid #FECACA;
  border-radius: 8px;
  color: #B91C1C;
  font-size: 14px;
}
</style>

