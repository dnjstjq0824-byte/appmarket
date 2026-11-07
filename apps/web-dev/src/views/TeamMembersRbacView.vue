<template>
  <div class="bn-dev-page">
    <header class="bn-page-header">
      <h2>조직 팀원 및 역할 관리 (Team Members RBAC)</h2>
      <p class="bn-text-muted">조직에 팀원을 초대하고 직무 분리 원칙에 따른 권한을 할당합니다.</p>
    </header>

    <div v-if="orgs.length > 0" class="bn-org-row">
      <label for="orgSelect">조직</label>
      <select id="orgSelect" v-model="activeOrgId" class="bn-select" @change="loadMembers">
        <option v-for="org in orgs" :key="org.id" :value="org.id">{{ org.org_name }}</option>
      </select>
    </div>
    <div v-if="isLoading" class="bn-loading">불러오는 중…</div>
    <div v-if="loadError" class="bn-error">{{ loadError }}</div>
    <div v-if="actionError" class="bn-error">{{ actionError }}</div>

    <div v-if="!isLoading && !loadError" class="bn-form-card">
      <div class="bn-invite-bar">
        <input v-model="newUserId" type="text" placeholder="팀원 사용자 ID..." class="bn-input" />
        <select v-model="newRole" class="bn-select">
          <option value="DEV_RELEASE_MGR">릴리스 관리자 (DEV_RELEASE_MGR)</option>
          <option value="DEV_EDITOR">콘텐츠 편집자 (DEV_EDITOR)</option>
          <option value="DEV_FINANCE">재무/정산 담당 (DEV_FINANCE)</option>
          <option value="DEV_SUPPORT">고객 지원 (DEV_SUPPORT)</option>
        </select>
        <button class="bn-btn bn-btn-primary" :disabled="isActing" @click="inviteMember">초대</button>
      </div>

      <div v-if="members.length === 0" class="bn-empty">등록된 팀원이 없습니다. 위에서 초대해 주십시오.</div>
      <table v-else class="bn-table">
        <thead>
          <tr>
            <th>팀원 사용자 ID</th>
            <th>역할</th>
            <th>권한 요약</th>
            <th>관리</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in members" :key="m.user_id">
            <td><strong>{{ m.user_id }}</strong></td>
            <td><span class="bn-role-tag">{{ m.role }}</span></td>
            <td class="bn-text-muted">{{ roleDescription(m.role) }}</td>
            <td>
              <button
                class="bn-btn-text-danger"
                :disabled="isActing"
                @click="removeMember(m.user_id)"
              >
                삭제
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getSharedClient, ApiClientError, type OrgSummary } from '@bookingnetwork/api-client'

const api = getSharedClient()

interface Member {
  org_id: string
  user_id: string
  role: string
}

const ROLE_DESCRIPTIONS: Record<string, string> = {
  DEV_OWNER: '조직 전권 및 소유권 관리',
  DEV_RELEASE_MGR: 'APK 업로드 및 출시 심사 제출',
  DEV_EDITOR: '스토어 리스팅 콘텐츠 편집',
  DEV_FINANCE: '매출 통계 및 정산 계좌 관리',
  DEV_SUPPORT: '리뷰 답글 및 고객 지원'
}

const orgs = ref<OrgSummary[]>([])
const activeOrgId = ref('')
const members = ref<Member[]>([])
const newUserId = ref('')
const newRole = ref('DEV_RELEASE_MGR')
const isLoading = ref(true)
const loadError = ref('')
const actionError = ref('')
const isActing = ref(false)

onMounted(async () => {
  try {
    orgs.value = await api.listOrgs()
    if (orgs.value.length > 0) {
      activeOrgId.value = orgs.value[0].id
      await loadMembers()
    }
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load organizations'
  } finally {
    isLoading.value = false
  }
})

async function loadMembers() {
  if (!activeOrgId.value) return
  isLoading.value = true
  loadError.value = ''
  try {
    members.value = await api.listMembers(activeOrgId.value) as Member[]
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load members'
  } finally {
    isLoading.value = false
  }
}

function roleDescription(role: string): string {
  return ROLE_DESCRIPTIONS[role] || '할당된 조직 역할'
}

async function inviteMember() {
  if (!newUserId.value.trim() || !activeOrgId.value) return
  isActing.value = true
  actionError.value = ''
  try {
    await api.addMember(activeOrgId.value, newUserId.value.trim(), newRole.value)
    newUserId.value = ''
    await loadMembers()
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to invite member'
  } finally {
    isActing.value = false
  }
}

async function removeMember(userId: string) {
  if (!confirm(`Remove member ${userId} from this organization?`)) return
  isActing.value = true
  actionError.value = ''
  try {
    await api.removeMember(activeOrgId.value, userId)
    await loadMembers()
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to remove member'
  } finally {
    isActing.value = false
  }
}
</script>

<style scoped>
.bn-dev-page {
  max-width: 960px;
  margin: 0 auto;
}
.bn-page-header {
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 2px solid var(--bn-ink);
}
.bn-page-header h2 {
  font-family: var(--bn-font-display);
  font-size: 24px;
  font-weight: 900;
  letter-spacing: -0.02em;
  margin: 0 0 4px;
}
.bn-text-muted {
  color: var(--bn-text-secondary);
  font-size: 13px;
  margin: 0;
}
td.bn-text-muted {
  font-size: 13px;
}
.bn-org-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  font-size: 13px;
  color: var(--bn-text-secondary);
}
.bn-org-row label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  flex-shrink: 0;
}
.bn-loading {
  padding: 24px;
  color: var(--bn-text-muted);
  font-size: 14px;
  text-align: center;
}
.bn-error {
  padding: 12px 16px;
  background: var(--bn-danger-wash);
  border: 1px solid var(--bn-danger-red);
  color: var(--bn-danger-red);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  margin-bottom: 16px;
}
.bn-empty {
  padding: 28px 24px;
  text-align: center;
  color: var(--bn-text-muted);
  font-size: 14px;
  border: 1px dashed var(--bn-border-strong);
  border-radius: var(--bn-radius-md);
}
.bn-form-card {
  background-color: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 24px;
}
.bn-invite-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}
.bn-input {
  flex: 1;
  min-width: 240px;
  padding: 10px 12px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-subtle);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  font-family: var(--bn-font-sans);
  font-size: 13px;
  box-sizing: border-box;
}
.bn-input:focus {
  outline: none;
  border-color: var(--bn-ink);
}
.bn-select {
  padding: 10px 12px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-subtle);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  font-family: var(--bn-font-sans);
  font-size: 13px;
}
.bn-btn {
  font-family: var(--bn-font-sans);
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
}
.bn-btn-primary {
  background-color: var(--bn-action-bg);
  color: var(--bn-action-text);
  border: none;
  padding: 10px 20px;
  border-radius: var(--bn-radius-sm);
}
.bn-btn-primary:hover:not(:disabled) {
  background-color: var(--bn-action-bg-hover);
}
.bn-btn-primary:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.bn-table {
  width: 100%;
  border-collapse: collapse;
}
.bn-table th, .bn-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid var(--bn-border-subtle);
  font-size: 13px;
  vertical-align: middle;
}
.bn-table tr:last-child td {
  border-bottom: none;
}
.bn-table th {
  font-family: var(--bn-font-mono);
  font-weight: 400;
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
}
.bn-role-tag {
  display: inline-block;
  font-family: var(--bn-font-mono);
  font-size: 11px;
  padding: 3px 9px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-strong);
  color: var(--bn-text-primary);
  white-space: nowrap;
}
.bn-btn-text-danger {
  color: var(--bn-danger-red);
  background: none;
  border: none;
  cursor: pointer;
  font-size: 13px;
  font-weight: 700;
  padding: 4px 0;
}
.bn-btn-text-danger:hover:not(:disabled) {
  text-decoration: underline;
  text-underline-offset: 3px;
}
.bn-btn-text-danger:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
</style>

